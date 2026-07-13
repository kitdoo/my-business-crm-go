package client

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/altessa-s/go-atlas/domain/converter"
	"github.com/altessa-s/go-atlas/domain/converter/codec/unixtime"
	"github.com/altessa-s/go-atlas/domain/proto/fieldmask"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"

	clientpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/client"

	clientsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/client/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	clientsvc "github.com/kitdoo/my-business-crm-go/internal/services/client"
)

// withUnixTimeCodec bridges int64 Unix-second proto fields (createdAt,
// updatedAt, …) with time.Time entity fields — see
// PROTO_DEVELOPMENT_STANDARD.md § Timestamps.
var withUnixTimeCodec = converter.WithCodecs(unixtime.New())

// Handler implements clientsvcpb.ClientsServiceServer.
type Handler struct {
	clientsvcpb.UnimplementedClientsServiceServer
	svc clientsvc.Service
}

// New builds a Handler.
func New(svc clientsvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	clientsvcpb.RegisterClientsServiceServer(gs, h)
}

func applyReadMask(mask *fieldmaskpb.FieldMask, msg *clientpb.Client) {
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

func (h *Handler) Create(ctx context.Context, in *clientsvcpb.ClientCreateRequest) (*clientsvcpb.ClientCreateResponse, error) {
	create := converter.Convert(in, &entities.ClientCreate{}, withUnixTimeCodec)

	c, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(c, &clientpb.Client{}, withUnixTimeCodec)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &clientsvcpb.ClientCreateResponse{Client: out}, nil
}

func (h *Handler) Get(ctx context.Context, in *clientsvcpb.ClientGetRequest) (*clientsvcpb.ClientGetResponse, error) {
	c, err := h.svc.Get(ctx, in.GetId())
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(c, &clientpb.Client{}, withUnixTimeCodec)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &clientsvcpb.ClientGetResponse{Client: out}, nil
}

func (h *Handler) List(ctx context.Context, in *clientsvcpb.ClientsListRequest) (*clientsvcpb.ClientsListResponse, error) {
	listIn := &entities.ClientsList{
		IncludeTotalCount: in.GetOptions().GetIncludeTotalCount(),
	}
	if sort := in.GetSort(); sort != nil {
		listIn.Sort = entities.ClientsListSort{
			Field:     entities.ClientsListSortField(sort.GetField()),
			Direction: entities.SortDirection(sort.GetDirection()),
		}
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		if p := f.GetCreatedAt(); p != nil {
			listIn.CreatedAt = converter.Convert(p, &entities.PeriodFilter{}, withUnixTimeCodec)
		}
		if f.Email != nil {
			listIn.Email = f.Email
		}
	}

	result, err := h.svc.List(ctx, listIn)
	if err != nil {
		return nil, MapError(err)
	}

	readMask := in.GetOptions().GetReadMask()
	out := &clientsvcpb.ClientsListResponse{
		Items: coreslices.To(result.Items, func(c *entities.Client) *clientpb.Client {
			item := converter.Convert(c, &clientpb.Client{}, withUnixTimeCodec)
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

func (h *Handler) Update(ctx context.Context, in *clientsvcpb.ClientUpdateRequest) (*clientsvcpb.ClientUpdateResponse, error) {
	update := &entities.ClientUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

	if pbMask := in.GetOptions().GetUpdateMask(); pbMask != nil {
		fm := fieldmask.FromProtoFieldMask(pbMask).Union(fieldmask.FromPaths("id"))
		if err := fm.ApplyUpdateMask(in); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	converter.Convert(in, update, withUnixTimeCodec, converter.WithIgnoreNilValues())

	c, err := h.svc.Update(ctx, update)
	if err != nil {
		return nil, MapError(err)
	}
	return &clientsvcpb.ClientUpdateResponse{Client: converter.Convert(c, &clientpb.Client{}, withUnixTimeCodec)}, nil
}

func (h *Handler) Delete(ctx context.Context, in *clientsvcpb.ClientDeleteRequest) (*clientsvcpb.ClientDeleteResponse, error) {
	del := &entities.ClientDelete{ID: in.GetId(), Etag: optionalString(in.GetOptions().GetEtag())}
	if err := h.svc.Delete(ctx, del); err != nil {
		return nil, MapError(err)
	}
	return &clientsvcpb.ClientDeleteResponse{}, nil
}

// MapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message.
func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrClientNotFound):
		return status.Error(codes.NotFound, errs.ErrClientNotFound.Error())
	case errors.Is(err, errs.ErrClientPhoneConflict):
		return status.Error(codes.AlreadyExists, errs.ErrClientPhoneConflict.Error())
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
