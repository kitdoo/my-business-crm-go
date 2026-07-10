package product

import (
	"context"
	"errors"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/altessa-s/go-atlas/domain/converter"
	convcodec "github.com/altessa-s/go-atlas/domain/converter/codec"
	"github.com/altessa-s/go-atlas/domain/converter/codec/unixtime"
	"github.com/altessa-s/go-atlas/domain/proto/fieldmask"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"

	commonpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/common"

	productpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/product"

	productsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/product/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	productsvc "github.com/kitdoo/my-business-crm-go/internal/services/product"
)

// localizedStringCodec bridges entities.LocalizedString (a plain map, kept
// map-shaped through the entity/storage layers) against the single-field
// wrapper message crm.types.common.LocalizedString. The converter cannot
// bridge a map directly onto a wrapper message on its own, so this codec -
// analogous to unixtime.New() - is the sanctioned extension point instead of
// a hand-written protoTo.../entityTo... function.
func localizedStringCodec() convcodec.Codec {
	lsType := reflect.TypeFor[entities.LocalizedString]()
	pbType := reflect.TypeFor[*commonpb.LocalizedString]()

	return func(fieldName string, src, dst reflect.Value, next convcodec.CodecHandler) {
		switch {
		case src.Type() == lsType && dst.Type() == pbType:
			if src.IsNil() {
				dst.Set(reflect.Zero(pbType))
				return
			}
			ls, _ := src.Interface().(entities.LocalizedString)
			dst.Set(reflect.ValueOf(&commonpb.LocalizedString{Values: map[string]string(ls)}))
		case src.Type() == pbType && dst.Type() == lsType:
			if src.IsNil() {
				dst.Set(reflect.Zero(lsType))
				return
			}
			pb, _ := src.Interface().(*commonpb.LocalizedString)
			dst.Set(reflect.ValueOf(entities.LocalizedString(pb.GetValues())))
		default:
			next(fieldName, src, dst)
		}
	}
}

// withProtoCodecs bridges timestamps (unixtime.New, see
// PROTO_DEVELOPMENT_STANDARD.md § Timestamps) and localized strings
// (including Details map values) between entities.Product and
// productpb.Product.
var withProtoCodecs = converter.WithCodecs(unixtime.New(), localizedStringCodec())

// withPeriodCodec bridges commonpb.PeriodFilter (*int64 bounds) against
// entities.PeriodFilter (*time.Time bounds).
var withPeriodCodec = converter.WithCodecs(unixtime.New())

// Handler implements productsvcpb.ProductsServiceServer.
type Handler struct {
	productsvcpb.UnimplementedProductsServiceServer
	svc productsvc.Service
}

// New builds a Handler.
func New(svc productsvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	productsvcpb.RegisterProductsServiceServer(gs, h)
}

func applyReadMask(mask *fieldmaskpb.FieldMask, msg *productpb.Product) {
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

func (h *Handler) Create(ctx context.Context, in *productsvcpb.ProductCreateRequest) (*productsvcpb.ProductCreateResponse, error) {
	create := converter.Convert(in, &entities.ProductCreate{}, withProtoCodecs)

	p, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(p, &productpb.Product{}, withProtoCodecs)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &productsvcpb.ProductCreateResponse{Product: out}, nil
}

func (h *Handler) Get(ctx context.Context, in *productsvcpb.ProductGetRequest) (*productsvcpb.ProductGetResponse, error) {
	p, err := h.svc.Get(ctx, in.GetId())
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(p, &productpb.Product{}, withProtoCodecs)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &productsvcpb.ProductGetResponse{Product: out}, nil
}

func (h *Handler) List(ctx context.Context, in *productsvcpb.ProductsListRequest) (*productsvcpb.ProductsListResponse, error) {
	listIn := &entities.ProductsList{
		BrandID:           optionalString(in.GetBrandId()),
		CategoryID:        optionalString(in.GetCategoryId()),
		IncludeTotalCount: in.GetOptions().GetIncludeTotalCount(),
	}
	if sort := in.GetSort(); sort != nil {
		listIn.Sort = entities.ProductsListSort{
			Field:     entities.ProductsListSortField(sort.GetField()),
			Direction: entities.SortDirection(sort.GetDirection()),
		}
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		listIn.Statuses = coreslices.To(f.GetStatuses(), func(s productpb.ProductStatus) entities.ProductStatus {
			return entities.ProductStatus(s)
		})
		listIn.BrandIDs = f.GetBrandIds()
		listIn.CategoryIDs = f.GetCategoryIds()
		listIn.SKUs = f.GetSkus()
		if p := f.GetCreatedAt(); p != nil {
			listIn.CreatedAt = converter.Convert(p, &entities.PeriodFilter{}, withPeriodCodec)
		}
	}

	result, err := h.svc.List(ctx, listIn)
	if err != nil {
		return nil, MapError(err)
	}

	readMask := in.GetOptions().GetReadMask()
	out := &productsvcpb.ProductsListResponse{
		Items: coreslices.To(result.Items, func(p *entities.Product) *productpb.Product {
			item := converter.Convert(p, &productpb.Product{}, withProtoCodecs)
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

func (h *Handler) Update(ctx context.Context, in *productsvcpb.ProductUpdateRequest) (*productsvcpb.ProductUpdateResponse, error) {
	update := &entities.ProductUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

	if pbMask := in.GetOptions().GetUpdateMask(); pbMask != nil {
		fm := fieldmask.FromProtoFieldMask(pbMask).Union(fieldmask.FromPaths("id"))
		if err := fm.ApplyUpdateMask(in); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	converter.Convert(in, update, withProtoCodecs, converter.WithIgnoreNilValues())

	p, err := h.svc.Update(ctx, update)
	if err != nil {
		return nil, MapError(err)
	}
	return &productsvcpb.ProductUpdateResponse{Product: converter.Convert(p, &productpb.Product{}, withProtoCodecs)}, nil
}

func (h *Handler) Delete(ctx context.Context, in *productsvcpb.ProductDeleteRequest) (*productsvcpb.ProductDeleteResponse, error) {
	del := &entities.ProductDelete{ID: in.GetId(), Etag: optionalString(in.GetOptions().GetEtag())}
	if err := h.svc.Delete(ctx, del); err != nil {
		return nil, MapError(err)
	}
	return &productsvcpb.ProductDeleteResponse{}, nil
}

// MapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message.
func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrProductNotFound):
		return status.Error(codes.NotFound, errs.ErrProductNotFound.Error())
	case errors.Is(err, errs.ErrProductSKUConflict):
		return status.Error(codes.AlreadyExists, errs.ErrProductSKUConflict.Error())
	case errors.Is(err, errs.ErrProductBrandNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrProductBrandNotFound.Error())
	case errors.Is(err, errs.ErrProductCategoryNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrProductCategoryNotFound.Error())
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
