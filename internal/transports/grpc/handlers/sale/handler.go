package sale

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/altessa-s/go-atlas/domain/converter"
	"github.com/altessa-s/go-atlas/domain/converter/codec/unixtime"
	"github.com/altessa-s/go-atlas/domain/proto/fieldmask"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"

	salepb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/sale"

	salesvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/sale/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/reqctx"
	salesvc "github.com/kitdoo/my-business-crm-go/internal/services/sale"
)

// withProtoCodecs bridges timestamps (unixtime.New, see
// PROTO_DEVELOPMENT_STANDARD.md § Timestamps) between entities.Sale and
// salepb.Sale; nested Items convert by field-name matching via
// WithHandleEmbeddedStructs.
var withProtoCodecs = converter.WithCodecs(unixtime.New())

// Handler implements salesvcpb.SalesServiceServer.
type Handler struct {
	salesvcpb.UnimplementedSalesServiceServer
	svc    salesvc.Service
	logger *slog.Logger
}

// New builds a Handler.
func New(svc salesvc.Service) *Handler {
	return &Handler{svc: svc, logger: slog.Default().With(slogx.Module("handler:sale"))}
}

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	salesvcpb.RegisterSalesServiceServer(gs, h)
}

func applyReadMask(mask *fieldmaskpb.FieldMask, msg *salepb.Sale) {
	if msg == nil || mask == nil || len(mask.GetPaths()) == 0 {
		return
	}
	fieldmask.FromProtoFieldMask(mask).Filter(msg)
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toSalePB(sl *entities.Sale) *salepb.Sale {
	return converter.Convert(sl, &salepb.Sale{}, withProtoCodecs, converter.WithHandleEmbeddedStructs(true))
}

func (h *Handler) Create(ctx context.Context, in *salesvcpb.SaleCreateRequest) (*salesvcpb.SaleCreateResponse, error) {
	create := converter.Convert(in, &entities.SaleCreate{}, withProtoCodecs, converter.WithHandleEmbeddedStructs(true))
	// clientId/newClient is a oneof (crm.grpc.sale.v1.SaleCreateRequest.
	// clientRef) — the generic converter above can't map an interface
	// field by reflection, so it's resolved by hand here.
	switch ref := in.GetClientRef().(type) {
	case *salesvcpb.SaleCreateRequest_ClientId:
		create.ClientID = ref.ClientId
	case *salesvcpb.SaleCreateRequest_NewClient_:
		nc := ref.NewClient
		create.NewClient = &entities.ClientCreate{
			Name:    nc.GetName(),
			Phone:   nc.GetPhone(),
			Email:   nc.GetEmail(),
			Address: nc.GetAddress(),
		}
	}
	// CreatedBy comes from the authenticated caller, not the request body;
	// see internal/pkg/reqctx for why this is currently always empty.
	if userID, ok := reqctx.UserIDFromContext(ctx); ok {
		create.CreatedBy = userID
	}

	sl, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, h.mapError(ctx, err)
	}
	out := toSalePB(sl)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &salesvcpb.SaleCreateResponse{Sale: out}, nil
}

func (h *Handler) Get(ctx context.Context, in *salesvcpb.SaleGetRequest) (*salesvcpb.SaleGetResponse, error) {
	sl, err := h.svc.Get(ctx, in.GetId())
	if err != nil {
		return nil, h.mapError(ctx, err)
	}
	out := toSalePB(sl)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &salesvcpb.SaleGetResponse{Sale: out}, nil
}

func (h *Handler) List(ctx context.Context, in *salesvcpb.SalesListRequest) (*salesvcpb.SalesListResponse, error) {
	listIn := &entities.SalesList{
		ClientID:          optionalString(in.GetClientId()),
		WarehouseID:       optionalString(in.GetWarehouseId()),
		PartnerID:         optionalString(in.GetPartnerId()),
		IncludeTotalCount: in.GetOptions().GetIncludeTotalCount(),
	}
	if sort := in.GetSort(); sort != nil {
		listIn.Sort = entities.SalesListSort{
			Field:     entities.SalesListSortField(sort.GetField()),
			Direction: entities.SortDirection(sort.GetDirection()),
		}
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		listIn.Statuses = coreslices.To(f.GetStatuses(), func(st salepb.SaleStatus) entities.SaleStatus {
			return entities.SaleStatus(st)
		})
		listIn.CreatedBy = f.GetCreatedBy()
		listIn.SKUIDs = f.GetSkuIds()
		if p := f.GetCreatedAt(); p != nil {
			listIn.CreatedAt = converter.Convert(p, &entities.PeriodFilter{}, withProtoCodecs)
		}
	}

	result, err := h.svc.List(ctx, listIn)
	if err != nil {
		return nil, h.mapError(ctx, err)
	}

	readMask := in.GetOptions().GetReadMask()
	out := &salesvcpb.SalesListResponse{
		Items: coreslices.To(result.Items, func(sl *entities.Sale) *salepb.Sale {
			item := toSalePB(sl)
			applyReadMask(readMask, item)
			return item
		}),
	}
	if result.Total >= 0 {
		out.Total = &result.Total
	}
	if result.NextCursor != "" {
		out.NextCursor = &result.NextCursor
	}
	return out, nil
}

func (h *Handler) UpdateStatus(ctx context.Context, in *salesvcpb.SaleUpdateStatusRequest) (*salesvcpb.SaleUpdateStatusResponse, error) {
	update := &entities.SaleUpdateStatus{
		ID:     in.GetId(),
		Status: entities.SaleStatus(in.GetStatus()),
		Etag:   optionalString(in.GetEtag()),
	}
	sl, err := h.svc.UpdateStatus(ctx, update)
	if err != nil {
		return nil, h.mapError(ctx, err)
	}
	return &salesvcpb.SaleUpdateStatusResponse{Sale: toSalePB(sl)}, nil
}

func (h *Handler) Cancel(ctx context.Context, in *salesvcpb.SaleCancelRequest) (*salesvcpb.SaleCancelResponse, error) {
	cancel := &entities.SaleCancel{
		ID:     in.GetId(),
		Reason: optionalString(in.GetReason()),
		Etag:   optionalString(in.GetEtag()),
	}
	if userID, ok := reqctx.UserIDFromContext(ctx); ok {
		cancel.CreatedBy = userID
	}

	sl, err := h.svc.Cancel(ctx, cancel)
	if err != nil {
		return nil, h.mapError(ctx, err)
	}
	return &salesvcpb.SaleCancelResponse{Sale: toSalePB(sl)}, nil
}

// mapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message to the caller — but the real
// error is logged here first, since this is the only place an unclassified
// failure from any of Create's dependencies (partners/clients/warehouses/
// skus/prices/storage) still passes through before being discarded. Without
// this, such failures were invisible: the services below only log via
// DebugContext, which the default deployed log level (error, see
// configs/logger.yaml) filters out entirely.
func (h *Handler) mapError(ctx context.Context, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrSaleNotFound):
		return status.Error(codes.NotFound, errs.ErrSaleNotFound.Error())
	case errors.Is(err, errs.ErrClientNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrClientNotFound.Error())
	case errors.Is(err, errs.ErrWarehouseNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrWarehouseNotFound.Error())
	case errors.Is(err, errs.ErrPartnerNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrPartnerNotFound.Error())
	case errors.Is(err, errs.ErrProductSkuNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrProductSkuNotFound.Error())
	case errors.Is(err, errs.ErrProductPriceNotFound):
		return status.Error(codes.FailedPrecondition, errs.ErrProductPriceNotFound.Error())
	case errors.Is(err, errs.ErrSaleWarehouseInactive):
		return status.Error(codes.FailedPrecondition, errs.ErrSaleWarehouseInactive.Error())
	case errors.Is(err, errs.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, errs.ErrInsufficientStock.Error())
	case errors.Is(err, errs.ErrSaleTerminalStatus):
		return status.Error(codes.FailedPrecondition, errs.ErrSaleTerminalStatus.Error())
	case errors.Is(err, errs.ErrSaleMissingClientOrPartner):
		return status.Error(codes.InvalidArgument, errs.ErrSaleMissingClientOrPartner.Error())
	case errors.Is(err, errs.ErrStaleEntity):
		return status.Error(codes.Aborted, errs.ErrStaleEntity.Error())
	case errors.Is(err, errs.ErrInvalidListCursor):
		return status.Error(codes.InvalidArgument, errs.ErrInvalidListCursor.Error())
	case errors.Is(err, errs.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, errs.ErrInvalidArgument.Error())
	case errors.Is(err, errs.ErrNotImplemented):
		return status.Error(codes.Unimplemented, errs.ErrNotImplemented.Error())
	default:
		h.logger.ErrorContext(ctx, "unclassified sale error mapped to Internal", slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
