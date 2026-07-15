package productsku

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

	productskupb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/product_sku"

	skusvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/product_sku/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	skusvc "github.com/kitdoo/my-business-crm-go/internal/services/productsku"
)

// localizedStringCodec bridges entities.LocalizedString (a plain map, kept
// map-shaped through the entity/storage layers) against the single-field
// wrapper message crm.types.common.LocalizedString — see the identical
// codec in handlers/product/handler.go for the full rationale.
func localizedStringCodec() convcodec.Codec {
	lsType := reflect.TypeFor[entities.LocalizedString]()
	pbType := reflect.TypeFor[*commonpb.LocalizedString]()
	pbValueType := pbType.Elem()

	return func(fieldName string, src, dst reflect.Value, next convcodec.CodecHandler) {
		switch {
		case src.Type() == lsType && dst.Type() == pbType:
			if src.IsNil() {
				dst.Set(reflect.Zero(pbType))
				return
			}
			ls, _ := src.Interface().(entities.LocalizedString)
			dst.Set(reflect.ValueOf(&commonpb.LocalizedString{Values: map[string]string(ls)}))
		case src.Type() == lsType && dst.Type() == pbValueType:
			ls, _ := src.Interface().(entities.LocalizedString)
			dst.FieldByName("Values").Set(reflect.ValueOf(map[string]string(ls)))
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
// (including Attributes map values) between entities.ProductSKU and
// productskupb.ProductSKU.
var withProtoCodecs = converter.WithCodecs(unixtime.New(), localizedStringCodec())

// withPeriodCodec bridges commonpb.PeriodFilter (*int64 bounds) against
// entities.PeriodFilter (*time.Time bounds).
var withPeriodCodec = converter.WithCodecs(unixtime.New())

// Handler implements skusvcpb.ProductSKUsServiceServer.
type Handler struct {
	skusvcpb.UnimplementedProductSKUsServiceServer
	svc skusvc.Service
}

// New builds a Handler.
func New(svc skusvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	skusvcpb.RegisterProductSKUsServiceServer(gs, h)
}

func applyReadMask(mask *fieldmaskpb.FieldMask, msg *productskupb.ProductSKU) {
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

func (h *Handler) Create(ctx context.Context, in *skusvcpb.ProductSKUCreateRequest) (*skusvcpb.ProductSKUCreateResponse, error) {
	create := converter.Convert(in, &entities.ProductSKUCreate{}, withProtoCodecs)

	v, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(v, &productskupb.ProductSKU{}, withProtoCodecs)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &skusvcpb.ProductSKUCreateResponse{Sku: out}, nil
}

func (h *Handler) Get(ctx context.Context, in *skusvcpb.ProductSKUGetRequest) (*skusvcpb.ProductSKUGetResponse, error) {
	v, err := h.svc.Get(ctx, in.GetId())
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(v, &productskupb.ProductSKU{}, withProtoCodecs)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &skusvcpb.ProductSKUGetResponse{Sku: out}, nil
}

func (h *Handler) List(ctx context.Context, in *skusvcpb.ProductSKUsListRequest) (*skusvcpb.ProductSKUsListResponse, error) {
	listIn := &entities.ProductSKUsList{
		VariantID:         optionalString(in.GetVariantId()),
		IncludeTotalCount: in.GetOptions().GetIncludeTotalCount(),
	}
	if sort := in.GetSort(); sort != nil {
		listIn.Sort = entities.ProductSKUsListSort{
			Field:     entities.ProductSKUsListSortField(sort.GetField()),
			Direction: entities.SortDirection(sort.GetDirection()),
		}
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		listIn.Statuses = coreslices.To(f.GetStatuses(), func(s productskupb.ProductSkuStatus) entities.ProductSkuStatus {
			return entities.ProductSkuStatus(s)
		})
		listIn.VariantIDs = f.GetVariantIds()
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
	out := &skusvcpb.ProductSKUsListResponse{
		Items: coreslices.To(result.Items, func(v *entities.ProductSKU) *productskupb.ProductSKU {
			item := converter.Convert(v, &productskupb.ProductSKU{}, withProtoCodecs)
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

func (h *Handler) Update(ctx context.Context, in *skusvcpb.ProductSKUUpdateRequest) (*skusvcpb.ProductSKUUpdateResponse, error) {
	update := &entities.ProductSKUUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

	if pbMask := in.GetOptions().GetUpdateMask(); pbMask != nil {
		fm := fieldmask.FromProtoFieldMask(pbMask).Union(fieldmask.FromPaths("id"))
		if err := fm.ApplyUpdateMask(in); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	converter.Convert(in, update, withProtoCodecs, converter.WithIgnoreNilValues())

	v, err := h.svc.Update(ctx, update)
	if err != nil {
		return nil, MapError(err)
	}
	return &skusvcpb.ProductSKUUpdateResponse{Sku: converter.Convert(v, &productskupb.ProductSKU{}, withProtoCodecs)}, nil
}

func (h *Handler) Delete(ctx context.Context, in *skusvcpb.ProductSKUDeleteRequest) (*skusvcpb.ProductSKUDeleteResponse, error) {
	del := &entities.ProductSKUDelete{ID: in.GetId(), Etag: optionalString(in.GetOptions().GetEtag())}
	if err := h.svc.Delete(ctx, del); err != nil {
		return nil, MapError(err)
	}
	return &skusvcpb.ProductSKUDeleteResponse{}, nil
}

// MapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message.
func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrProductSkuNotFound):
		return status.Error(codes.NotFound, errs.ErrProductSkuNotFound.Error())
	case errors.Is(err, errs.ErrProductSkuSKUConflict):
		return status.Error(codes.AlreadyExists, errs.ErrProductSkuSKUConflict.Error())
	case errors.Is(err, errs.ErrProductSkuVariantNotFound):
		return status.Error(codes.InvalidArgument, errs.ErrProductSkuVariantNotFound.Error())
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
