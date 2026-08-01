// Package invoice implements the invoice-PDF-generation endpoint. This is
// plain HTTP, not gRPC — like internal/transports/http/handlers/image, it
// has no dedicated proto service; a Sale is rendered to a PDF and streamed
// back (optionally also emailed) via a single POST.
package invoice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	slogx "github.com/altessa-s/go-atlas/observability/slog"
	httpserver "github.com/altessa-s/go-atlas/transport/http/server"
	"github.com/altessa-s/go-atlas/transport/http/server/writer"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/rbac"
	invoicesvc "github.com/kitdoo/my-business-crm-go/internal/services/invoice"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
	grpcrbac "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/interceptors/rbac"
)

const pathPrefix = "/invoices"

// requestBody is the JSON body of POST /invoices/{saleId} — every field
// mirrors entities.InvoiceGenerate except SaleID, which comes off the URL.
type requestBody struct {
	// DocumentType selects "offer" (any Sale status — the default when
	// omitted) or "receipt" (requires a paid-or-later Sale status — see
	// entities.InvoiceDocumentType). Serbian "ponuda"/"račun" respectively
	// — that naming stays confined to the rendered PDF text and UI labels,
	// not the wire contract.
	DocumentType            string  `json:"documentType"`
	BuyerTaxID              string  `json:"buyerTaxId"`
	BuyerRegistrationNumber string  `json:"buyerRegistrationNumber"`
	BuyerCode               string  `json:"buyerCode"`
	BuyerAddress            *string `json:"buyerAddress"`
	IncludeVAT              bool    `json:"includeVat"`
	SendEmail               bool    `json:"sendEmail"`
	EmailOverride           *string `json:"emailOverride"`
}

// documentTypes maps the wire string to entities.InvoiceDocumentType — an
// explicit whitelist rather than a generic string-to-enum helper since
// there are only two values.
var documentTypes = map[string]entities.InvoiceDocumentType{
	"":        entities.InvoiceDocumentTypeOffer,
	"offer":   entities.InvoiceDocumentTypeOffer,
	"receipt": entities.InvoiceDocumentTypeReceipt,
}

// Handler implements httpserver.Handler for POST /invoices/{saleId}.
// Requires the caller's role to hold grpcrbac.InvoicesGeneratePermission,
// checked inline the same way image.Handler does — this is a plain-HTTP
// endpoint the gRPC RBAC interceptor never sees.
type Handler struct {
	invoices invoicesvc.Service
	users    usersvc.Service
	rbac     rbac.Table
	logger   *slog.Logger
}

var _ httpserver.Handler = (*Handler)(nil)

// New builds a Handler.
func New(invoices invoicesvc.Service, users usersvc.Service, table rbac.Table) *Handler {
	return &Handler{
		invoices: invoices,
		users:    users,
		rbac:     table,
		logger:   slog.Default().With(slogx.Module("http:invoice")),
	}
}

// Register attaches the generate route.
func (h *Handler) Register(r httpserver.RouteRegistrar, _ <-chan struct{}) {
	r.Handle(pathPrefix+"/{saleId}", h.generate).Methods(http.MethodPost)
}

func (h *Handler) generate(rw writer.ReadWriter) {
	ctx := rw.Request().Context()

	if err := h.requirePermission(ctx, rw.Request()); err != nil {
		writeAuthError(rw, err)
		return
	}

	saleID := mux.Vars(rw.Request())["saleId"]
	if saleID == "" {
		_ = rw.WriteError(errors.New("missing saleId"), http.StatusBadRequest)
		return
	}

	var body requestBody
	if rw.Request().ContentLength != 0 {
		if err := json.NewDecoder(rw.Request().Body).Decode(&body); err != nil {
			_ = rw.WriteError(errors.New("invalid request body"), http.StatusBadRequest)
			return
		}
	}

	docType, ok := documentTypes[body.DocumentType]
	if !ok {
		_ = rw.WriteError(fmt.Errorf("invalid documentType: %q", body.DocumentType), http.StatusBadRequest)
		return
	}

	pdf, err := h.invoices.Generate(ctx, &entities.InvoiceGenerate{
		SaleID:                  saleID,
		DocumentType:            docType,
		BuyerTaxID:              body.BuyerTaxID,
		BuyerRegistrationNumber: body.BuyerRegistrationNumber,
		BuyerCode:               body.BuyerCode,
		BuyerAddress:            body.BuyerAddress,
		IncludeVAT:              body.IncludeVAT,
		SendEmail:               body.SendEmail,
		EmailOverride:           body.EmailOverride,
	})
	if err != nil {
		writeGenerateError(rw, err)
		return
	}

	w := rw.ResponseWriter()
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="invoice.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

func (h *Handler) requirePermission(ctx context.Context, r *http.Request) error {
	token := bearerToken(r)
	if token == "" {
		return errs.ErrUnauthenticated
	}
	u, err := h.users.Authenticate(ctx, token)
	if err != nil {
		return err
	}
	if !h.rbac.Allowed(u.Role.String(), grpcrbac.InvoicesGeneratePermission) {
		return errs.ErrForbidden
	}
	return nil
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(auth, prefix))
}

func writeAuthError(rw writer.ReadWriter, err error) {
	if errors.Is(err, errs.ErrForbidden) {
		_ = rw.WriteError(err, http.StatusForbidden)
		return
	}
	_ = rw.WriteError(err, http.StatusUnauthorized)
}

// notFoundErrors are every "no such row" sentinel Generate's dependency
// chain (Sale -> Client/Partner -> SKU -> Variant -> Product) can surface.
var notFoundErrors = []error{
	errs.ErrSaleNotFound,
	errs.ErrClientNotFound,
	errs.ErrPartnerNotFound,
	errs.ErrProductSkuNotFound,
	errs.ErrProductVariantNotFound,
	errs.ErrProductNotFound,
}

func writeGenerateError(rw writer.ReadWriter, err error) {
	switch {
	case errors.Is(err, errs.ErrInvoiceNotConfigured):
		_ = rw.WriteError(err, http.StatusServiceUnavailable)
	case errors.Is(err, errs.ErrInvoiceSaleHasNoBuyer), errors.Is(err, errs.ErrInvoiceReceiptRequiresPayment):
		_ = rw.WriteError(err, http.StatusUnprocessableEntity)
	default:
		for _, nf := range notFoundErrors {
			if errors.Is(err, nf) {
				_ = rw.WriteError(err, http.StatusNotFound)
				return
			}
		}
		_ = rw.WriteError(fmt.Errorf("generate invoice: %w", err), http.StatusInternalServerError)
	}
}
