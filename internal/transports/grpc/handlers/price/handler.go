package price

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/altessa-s/go-atlas/domain/converter"
	"github.com/altessa-s/go-atlas/domain/converter/codec/unixtime"
	"github.com/altessa-s/go-atlas/domain/proto/fieldmask"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"

	pricepb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/price"

	pricesvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/price/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	pricesvc "github.com/kitdoo/my-business-crm-go/internal/services/price"
)

// withUnixTimeCodec bridges int64 Unix-second proto fields (createdAt,
// updatedAt, …) with time.Time entity fields — see
// PROTO_DEVELOPMENT_STANDARD.md § Timestamps.
var withUnixTimeCodec = converter.WithCodecs(unixtime.New())

// Handler implements pricesvcpb.PricesServiceServer.
type Handler struct {
	pricesvcpb.UnimplementedPricesServiceServer
	svc pricesvc.Service
}

// New builds a Handler.
func New(svc pricesvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	pricesvcpb.RegisterPricesServiceServer(gs, h)
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (h *Handler) Create(ctx context.Context, in *pricesvcpb.ProductPriceCreateRequest) (*pricesvcpb.ProductPriceCreateResponse, error) {
	create := converter.Convert(in, &entities.ProductPriceCreate{}, withUnixTimeCodec)

	p, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, MapError(err)
	}
	return &pricesvcpb.ProductPriceCreateResponse{Price: converter.Convert(p, &pricepb.ProductPrice{}, withUnixTimeCodec)}, nil
}

func (h *Handler) Get(ctx context.Context, in *pricesvcpb.ProductPriceGetRequest) (*pricesvcpb.ProductPriceGetResponse, error) {
	p, err := h.svc.Get(ctx, in.GetSkuId())
	if err != nil {
		return nil, MapError(err)
	}
	return &pricesvcpb.ProductPriceGetResponse{Price: converter.Convert(p, &pricepb.ProductPrice{}, withUnixTimeCodec)}, nil
}

func (h *Handler) Update(ctx context.Context, in *pricesvcpb.ProductPriceUpdateRequest) (*pricesvcpb.ProductPriceUpdateResponse, error) {
	update := &entities.ProductPriceUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

	if pbMask := in.GetOptions().GetUpdateMask(); pbMask != nil {
		fm := fieldmask.FromProtoFieldMask(pbMask).Union(fieldmask.FromPaths("id"))
		if err := fm.ApplyUpdateMask(in); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	converter.Convert(in, update, withUnixTimeCodec, converter.WithIgnoreNilValues())

	p, err := h.svc.Update(ctx, update)
	if err != nil {
		return nil, MapError(err)
	}
	return &pricesvcpb.ProductPriceUpdateResponse{Price: converter.Convert(p, &pricepb.ProductPrice{}, withUnixTimeCodec)}, nil
}

func (h *Handler) Delete(ctx context.Context, in *pricesvcpb.ProductPriceDeleteRequest) (*pricesvcpb.ProductPriceDeleteResponse, error) {
	del := &entities.ProductPriceDelete{ID: in.GetId(), Etag: optionalString(in.GetOptions().GetEtag())}
	if err := h.svc.Delete(ctx, del); err != nil {
		return nil, MapError(err)
	}
	return &pricesvcpb.ProductPriceDeleteResponse{}, nil
}

func (h *Handler) GetHistory(ctx context.Context, in *pricesvcpb.ProductPriceGetHistoryRequest) (*pricesvcpb.ProductPriceGetHistoryResponse, error) {
	historyIn := &entities.ProductPriceGetHistory{SKUID: in.GetSkuId()}
	if pg := in.GetPagination(); pg != nil {
		historyIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		if p := f.GetCreatedAt(); p != nil {
			historyIn.CreatedAt = converter.Convert(p, &entities.PeriodFilter{}, withUnixTimeCodec)
		}
	}

	result, err := h.svc.GetHistory(ctx, historyIn)
	if err != nil {
		return nil, MapError(err)
	}

	out := &pricesvcpb.ProductPriceGetHistoryResponse{
		Items: coreslices.To(result.Items, func(p *entities.ProductPrice) *pricepb.ProductPrice {
			return converter.Convert(p, &pricepb.ProductPrice{}, withUnixTimeCodec)
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
	case errors.Is(err, errs.ErrProductPriceNotFound):
		return status.Error(codes.NotFound, errs.ErrProductPriceNotFound.Error())
	case errors.Is(err, errs.ErrProductSkuNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrProductSkuNotFound.Error())
	case errors.Is(err, errs.ErrProductPriceExists):
		return status.Error(codes.AlreadyExists, errs.ErrProductPriceExists.Error())
	case errors.Is(err, errs.ErrStaleEntity):
		return status.Error(codes.Aborted, errs.ErrStaleEntity.Error())
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
