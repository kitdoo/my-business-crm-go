package inventorymovement

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/altessa-s/go-atlas/domain/converter"
	"github.com/altessa-s/go-atlas/domain/converter/codec/unixtime"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"

	inventorymovementpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/inventory_movement"

	inventorymovementsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/inventory_movement/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/reqctx"
	invsvc "github.com/kitdoo/my-business-crm-go/internal/services/inventorymovement"
)

// withUnixTimeCodec bridges the int64 Unix-second proto field (createdAt)
// with the time.Time entity field — see PROTO_DEVELOPMENT_STANDARD.md §
// Timestamps.
var withUnixTimeCodec = converter.WithCodecs(unixtime.New())

// Handler implements inventorymovementsvcpb.InventoryMovementsServiceServer.
type Handler struct {
	inventorymovementsvcpb.UnimplementedInventoryMovementsServiceServer
	svc invsvc.Service
}

// New builds a Handler.
func New(svc invsvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	inventorymovementsvcpb.RegisterInventoryMovementsServiceServer(gs, h)
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (h *Handler) Create(ctx context.Context, in *inventorymovementsvcpb.InventoryMovementCreateRequest) (*inventorymovementsvcpb.InventoryMovementCreateResponse, error) {
	create := converter.Convert(in, &entities.InventoryMovementCreate{}, withUnixTimeCodec)
	// CreatedBy comes from the authenticated caller, not the request body
	// (the proto Create request carries no createdBy field); see
	// internal/pkg/reqctx for why this is currently always empty.
	if userID, ok := reqctx.UserIDFromContext(ctx); ok {
		create.CreatedBy = userID
	}

	m, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, MapError(err)
	}
	return &inventorymovementsvcpb.InventoryMovementCreateResponse{
		Movement: converter.Convert(m, &inventorymovementpb.InventoryMovement{}, withUnixTimeCodec),
	}, nil
}

func (h *Handler) List(ctx context.Context, in *inventorymovementsvcpb.InventoryMovementsListRequest) (*inventorymovementsvcpb.InventoryMovementsListResponse, error) {
	listIn := &entities.InventoryMovementsList{
		WarehouseID: optionalString(in.GetWarehouseId()),
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		listIn.Types = coreslices.To(f.GetTypes(), func(t inventorymovementpb.MovementType) entities.MovementType {
			return entities.MovementType(t)
		})
		listIn.ProductIDs = f.GetProductIds()
		listIn.CreatedBy = f.GetCreatedBy()
		if p := f.GetCreatedAt(); p != nil {
			listIn.CreatedAt = converter.Convert(p, &entities.PeriodFilter{}, withUnixTimeCodec)
		}
	}

	result, err := h.svc.List(ctx, listIn)
	if err != nil {
		return nil, MapError(err)
	}

	out := &inventorymovementsvcpb.InventoryMovementsListResponse{
		Items: coreslices.To(result.Items, func(m *entities.InventoryMovement) *inventorymovementpb.InventoryMovement {
			return converter.Convert(m, &inventorymovementpb.InventoryMovement{}, withUnixTimeCodec)
		}),
	}
	if result.NextCursor != "" {
		out.NextCursor = &result.NextCursor
	}
	return out, nil
}

func (h *Handler) GetHistory(ctx context.Context, in *inventorymovementsvcpb.InventoryMovementGetHistoryRequest) (*inventorymovementsvcpb.InventoryMovementGetHistoryResponse, error) {
	historyIn := &entities.InventoryMovementGetHistory{
		ProductID:   in.GetProductId(),
		WarehouseID: in.GetWarehouseId(),
	}
	if pg := in.GetPagination(); pg != nil {
		historyIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		historyIn.Types = coreslices.To(f.GetTypes(), func(t inventorymovementpb.MovementType) entities.MovementType {
			return entities.MovementType(t)
		})
		if p := f.GetCreatedAt(); p != nil {
			historyIn.CreatedAt = converter.Convert(p, &entities.PeriodFilter{}, withUnixTimeCodec)
		}
	}

	result, err := h.svc.GetHistory(ctx, historyIn)
	if err != nil {
		return nil, MapError(err)
	}

	out := &inventorymovementsvcpb.InventoryMovementGetHistoryResponse{
		Items: coreslices.To(result.Items, func(m *entities.InventoryMovement) *inventorymovementpb.InventoryMovement {
			return converter.Convert(m, &inventorymovementpb.InventoryMovement{}, withUnixTimeCodec)
		}),
	}
	if result.NextCursor != "" {
		out.NextCursor = &result.NextCursor
	}
	return out, nil
}

// MapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message.
func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrProductNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrProductNotFound.Error())
	case errors.Is(err, errs.ErrWarehouseNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrWarehouseNotFound.Error())
	case errors.Is(err, errs.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, errs.ErrInsufficientStock.Error())
	case errors.Is(err, errs.ErrInvalidListCursor):
		return status.Error(codes.InvalidArgument, errs.ErrInvalidListCursor.Error())
	case errors.Is(err, errs.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, errs.ErrInvalidArgument.Error())
	case errors.Is(err, errs.ErrNotImplemented):
		return status.Error(codes.Unimplemented, errs.ErrNotImplemented.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
