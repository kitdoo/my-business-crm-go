package inventory

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/altessa-s/go-atlas/domain/converter"
	"github.com/altessa-s/go-atlas/domain/converter/codec/unixtime"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"

	inventorypb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/inventory"

	inventorysvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/inventory/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	inventorysvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
)

// withUnixTimeCodec bridges int64 Unix-second proto fields (updatedAt) with
// time.Time entity fields — see PROTO_DEVELOPMENT_STANDARD.md § Timestamps.
var withUnixTimeCodec = converter.WithCodecs(unixtime.New())

// Handler implements inventorysvcpb.InventoryServiceServer.
type Handler struct {
	inventorysvcpb.UnimplementedInventoryServiceServer
	svc inventorysvc.Service
}

// New builds a Handler.
func New(svc inventorysvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	inventorysvcpb.RegisterInventoryServiceServer(gs, h)
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (h *Handler) Get(ctx context.Context, in *inventorysvcpb.InventoryGetRequest) (*inventorysvcpb.InventoryGetResponse, error) {
	i, err := h.svc.Get(ctx, in.GetProductId(), in.GetWarehouseId())
	if err != nil {
		return nil, MapError(err)
	}
	return &inventorysvcpb.InventoryGetResponse{Inventory: converter.Convert(i, &inventorypb.Inventory{}, withUnixTimeCodec)}, nil
}

func (h *Handler) List(ctx context.Context, in *inventorysvcpb.InventoryListRequest) (*inventorysvcpb.InventoryListResponse, error) {
	listIn := &entities.InventoryList{
		ProductID:   optionalString(in.GetProductId()),
		WarehouseID: optionalString(in.GetWarehouseId()),
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		if f.MinQuantity != nil {
			listIn.MinQuantity = f.MinQuantity
		}
		if f.MaxQuantity != nil {
			listIn.MaxQuantity = f.MaxQuantity
		}
	}

	result, err := h.svc.List(ctx, listIn)
	if err != nil {
		return nil, MapError(err)
	}

	out := &inventorysvcpb.InventoryListResponse{
		Items: coreslices.To(result.Items, func(i *entities.Inventory) *inventorypb.Inventory {
			return converter.Convert(i, &inventorypb.Inventory{}, withUnixTimeCodec)
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
	case errors.Is(err, errs.ErrInventoryNotFound):
		return status.Error(codes.NotFound, errs.ErrInventoryNotFound.Error())
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
