package productattributedefinition

import (
	"context"
	"errors"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/altessa-s/go-atlas/domain/converter"
	convcodec "github.com/altessa-s/go-atlas/domain/converter/codec"
	"github.com/altessa-s/go-atlas/domain/converter/codec/unixtime"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"

	commonpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/common"
	productattributedefinitionpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/product_attribute_definition"

	productattributedefinitionsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/product_attribute_definition/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	productattributedefinitionsvc "github.com/kitdoo/my-business-crm-go/internal/services/productattributedefinition"
)

// localizedStringCodec bridges entities.LocalizedString against the
// wrapper message crm.types.common.LocalizedString — see
// handlers/product/handler.go for the canonical explanation of why this
// exists (duplicated per-package by convention, not shared).
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

var withProtoCodecs = converter.WithCodecs(unixtime.New(), localizedStringCodec())

// Handler implements productattributedefinitionsvcpb.ProductAttributeDefinitionsServiceServer.
type Handler struct {
	productattributedefinitionsvcpb.UnimplementedProductAttributeDefinitionsServiceServer
	svc productattributedefinitionsvc.Service
}

// New builds a Handler.
func New(svc productattributedefinitionsvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	productattributedefinitionsvcpb.RegisterProductAttributeDefinitionsServiceServer(gs, h)
}

func (h *Handler) Get(ctx context.Context, in *productattributedefinitionsvcpb.ProductAttributeDefinitionGetRequest) (*productattributedefinitionsvcpb.ProductAttributeDefinitionGetResponse, error) {
	d, err := h.svc.Get(ctx, in.GetId())
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(d, &productattributedefinitionpb.ProductAttributeDefinition{}, withProtoCodecs)
	return &productattributedefinitionsvcpb.ProductAttributeDefinitionGetResponse{ProductAttributeDefinition: out}, nil
}

func (h *Handler) List(ctx context.Context, in *productattributedefinitionsvcpb.ProductAttributeDefinitionsListRequest) (*productattributedefinitionsvcpb.ProductAttributeDefinitionsListResponse, error) {
	listIn := &entities.ProductAttributeDefinitionsList{}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		listIn.IsPublic = f.IsPublic
	}

	result, err := h.svc.List(ctx, listIn)
	if err != nil {
		return nil, MapError(err)
	}

	out := &productattributedefinitionsvcpb.ProductAttributeDefinitionsListResponse{
		Items: coreslices.To(result.Items, func(d *entities.ProductAttributeDefinition) *productattributedefinitionpb.ProductAttributeDefinition {
			return converter.Convert(d, &productattributedefinitionpb.ProductAttributeDefinition{}, withProtoCodecs)
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
	case errors.Is(err, errs.ErrProductAttributeDefinitionNotFound):
		return status.Error(codes.NotFound, errs.ErrProductAttributeDefinitionNotFound.Error())
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
