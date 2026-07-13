package category

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

	categorypb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/category"

	categorysvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/category/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	categorysvc "github.com/kitdoo/my-business-crm-go/internal/services/category"
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
// PROTO_DEVELOPMENT_STANDARD.md § Timestamps) and localized strings between
// entities.Category and categorypb.Category.
var withProtoCodecs = converter.WithCodecs(unixtime.New(), localizedStringCodec())

// Handler implements categorysvcpb.CategoriesServiceServer.
type Handler struct {
	categorysvcpb.UnimplementedCategoriesServiceServer
	svc categorysvc.Service
}

// New builds a Handler.
func New(svc categorysvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	categorysvcpb.RegisterCategoriesServiceServer(gs, h)
}

func applyReadMask(mask *fieldmaskpb.FieldMask, msg *categorypb.Category) {
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

func (h *Handler) Create(ctx context.Context, in *categorysvcpb.CategoryCreateRequest) (*categorysvcpb.CategoryCreateResponse, error) {
	create := converter.Convert(in, &entities.CategoryCreate{}, withProtoCodecs)

	c, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(c, &categorypb.Category{}, withProtoCodecs)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &categorysvcpb.CategoryCreateResponse{Category: out}, nil
}

func (h *Handler) Get(ctx context.Context, in *categorysvcpb.CategoryGetRequest) (*categorysvcpb.CategoryGetResponse, error) {
	c, err := h.svc.Get(ctx, in.GetId())
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(c, &categorypb.Category{}, withProtoCodecs)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &categorysvcpb.CategoryGetResponse{Category: out}, nil
}

func (h *Handler) List(ctx context.Context, in *categorysvcpb.CategoriesListRequest) (*categorysvcpb.CategoriesListResponse, error) {
	listIn := &entities.CategoriesList{
		IncludeTotalCount: in.GetOptions().GetIncludeTotalCount(),
	}
	if sort := in.GetSort(); sort != nil {
		listIn.Sort = entities.CategoriesListSort{
			Field:     entities.CategoriesListSortField(sort.GetField()),
			Direction: entities.SortDirection(sort.GetDirection()),
		}
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		listIn.Statuses = coreslices.To(f.GetStatuses(), func(s categorypb.CategoryStatus) entities.CategoryStatus {
			return entities.CategoryStatus(s)
		})
		if p := f.GetCreatedAt(); p != nil {
			listIn.CreatedAt = converter.Convert(p, &entities.PeriodFilter{}, withPeriodCodec)
		}
	}

	result, err := h.svc.List(ctx, listIn)
	if err != nil {
		return nil, MapError(err)
	}

	readMask := in.GetOptions().GetReadMask()
	out := &categorysvcpb.CategoriesListResponse{
		Items: coreslices.To(result.Items, func(c *entities.Category) *categorypb.Category {
			item := converter.Convert(c, &categorypb.Category{}, withProtoCodecs)
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

func (h *Handler) Update(ctx context.Context, in *categorysvcpb.CategoryUpdateRequest) (*categorysvcpb.CategoryUpdateResponse, error) {
	update := &entities.CategoryUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

	if pbMask := in.GetOptions().GetUpdateMask(); pbMask != nil {
		fm := fieldmask.FromProtoFieldMask(pbMask).Union(fieldmask.FromPaths("id"))
		if err := fm.ApplyUpdateMask(in); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	converter.Convert(in, update, withProtoCodecs, converter.WithIgnoreNilValues())

	c, err := h.svc.Update(ctx, update)
	if err != nil {
		return nil, MapError(err)
	}
	return &categorysvcpb.CategoryUpdateResponse{Category: converter.Convert(c, &categorypb.Category{}, withProtoCodecs)}, nil
}

func (h *Handler) Delete(ctx context.Context, in *categorysvcpb.CategoryDeleteRequest) (*categorysvcpb.CategoryDeleteResponse, error) {
	del := &entities.CategoryDelete{ID: in.GetId(), Etag: optionalString(in.GetOptions().GetEtag())}
	if err := h.svc.Delete(ctx, del); err != nil {
		return nil, MapError(err)
	}
	return &categorysvcpb.CategoryDeleteResponse{}, nil
}

// withPeriodCodec bridges commonpb.PeriodFilter (*int64 bounds) against
// entities.PeriodFilter (*time.Time bounds).
var withPeriodCodec = converter.WithCodecs(unixtime.New())

// MapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message.
func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrCategoryNotFound):
		return status.Error(codes.NotFound, errs.ErrCategoryNotFound.Error())
	case errors.Is(err, errs.ErrCategoryHasProducts):
		return status.Error(codes.FailedPrecondition, errs.ErrCategoryHasProducts.Error())
	case errors.Is(err, errs.ErrStaleEntity):
		return status.Error(codes.Aborted, errs.ErrStaleEntity.Error())
	case errors.Is(err, errs.ErrInvalidListCursor):
		return status.Error(codes.InvalidArgument, errs.ErrInvalidListCursor.Error())
	case errors.Is(err, errs.ErrLocalizedStringMissingRequiredLocale):
		// err, not the sentinel: the entities layer wraps this with the
		// actual field name ("name: ..."/"description: ..."), which the
		// BFF regex-extracts into a per-field UI error (TD §9.4) instead
		// of an unattributed toast.
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, errs.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, errs.ErrInvalidArgument.Error())
	case errors.Is(err, errs.ErrNotImplemented):
		return status.Error(codes.Unimplemented, errs.ErrNotImplemented.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
