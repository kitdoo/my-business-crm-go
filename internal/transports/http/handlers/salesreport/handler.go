// Package salesreport implements the sales-report-generation endpoint.
// This is plain HTTP, not gRPC — same reasoning as
// internal/transports/http/handlers/invoice: rendering a period's worth of
// Sales to a file and streaming it back has no dedicated proto service.
package salesreport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	slogx "github.com/altessa-s/go-atlas/observability/slog"
	httpserver "github.com/altessa-s/go-atlas/transport/http/server"
	"github.com/altessa-s/go-atlas/transport/http/server/writer"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/rbac"
	salesreportsvc "github.com/kitdoo/my-business-crm-go/internal/services/salesreport"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
	grpcrbac "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/interceptors/rbac"
)

const path = "/sales-report"

// requestBody is the JSON body of POST /sales-report — Unix-second bounds,
// both inclusive, matching entities.PeriodFilter's own convention.
type requestBody struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

// Handler implements httpserver.Handler for POST /sales-report. Requires
// the caller's role to hold grpcrbac.SalesReportGeneratePermission, checked
// inline the same way invoice.Handler does — this is a plain-HTTP endpoint
// the gRPC RBAC interceptor never sees.
type Handler struct {
	salesReports salesreportsvc.Service
	users        usersvc.Service
	rbac         rbac.Table
	logger       *slog.Logger
}

var _ httpserver.Handler = (*Handler)(nil)

// New builds a Handler.
func New(salesReports salesreportsvc.Service, users usersvc.Service, table rbac.Table) *Handler {
	return &Handler{
		salesReports: salesReports,
		users:        users,
		rbac:         table,
		logger:       slog.Default().With(slogx.Module("http:salesreport")),
	}
}

// Register attaches the generate route.
func (h *Handler) Register(r httpserver.RouteRegistrar, _ <-chan struct{}) {
	r.Handle(path, h.generate).Methods(http.MethodPost)
}

func (h *Handler) generate(rw writer.ReadWriter) {
	ctx := rw.Request().Context()

	if err := h.requirePermission(ctx, rw.Request()); err != nil {
		writeAuthError(rw, err)
		return
	}

	var body requestBody
	if err := json.NewDecoder(rw.Request().Body).Decode(&body); err != nil {
		_ = rw.WriteError(errors.New("invalid request body"), http.StatusBadRequest)
		return
	}
	if body.From == 0 || body.To == 0 || body.To < body.From {
		_ = rw.WriteError(errors.New("from/to are required and to must not precede from"), http.StatusBadRequest)
		return
	}

	xlsx, err := h.salesReports.Generate(ctx, &entities.SalesReportGenerate{
		From: time.Unix(body.From, 0).UTC(),
		To:   time.Unix(body.To, 0).UTC(),
	})
	if err != nil {
		_ = rw.WriteError(fmt.Errorf("generate sales report: %w", err), http.StatusInternalServerError)
		return
	}

	w := rw.ResponseWriter()
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="sales-report.xlsx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(xlsx)
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
	if !h.rbac.Allowed(u.Role.String(), grpcrbac.SalesReportGeneratePermission) {
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
