package partner

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

	partnerpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/partner"

	partnersvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/partner/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	partnersvc "github.com/kitdoo/my-business-crm-go/internal/services/partner"
)

// withUnixTimeCodec bridges int64 Unix-second proto fields (createdAt,
// updatedAt, …) with time.Time entity fields — see
// PROTO_DEVELOPMENT_STANDARD.md § Timestamps.
var withUnixTimeCodec = converter.WithCodecs(unixtime.New())

// Handler implements partnersvcpb.PartnersServiceServer.
type Handler struct {
	partnersvcpb.UnimplementedPartnersServiceServer
	svc partnersvc.Service
}

// New builds a Handler.
func New(svc partnersvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	partnersvcpb.RegisterPartnersServiceServer(gs, h)
}

func applyReadMask(mask *fieldmaskpb.FieldMask, msg *partnerpb.Partner) {
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

func (h *Handler) Create(ctx context.Context, in *partnersvcpb.PartnerCreateRequest) (*partnersvcpb.PartnerCreateResponse, error) {
	create := converter.Convert(in, &entities.PartnerCreate{}, withUnixTimeCodec)

	p, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(p, &partnerpb.Partner{}, withUnixTimeCodec)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &partnersvcpb.PartnerCreateResponse{Partner: out}, nil
}

func (h *Handler) Get(ctx context.Context, in *partnersvcpb.PartnerGetRequest) (*partnersvcpb.PartnerGetResponse, error) {
	p, err := h.svc.Get(ctx, in.GetId())
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(p, &partnerpb.Partner{}, withUnixTimeCodec)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &partnersvcpb.PartnerGetResponse{Partner: out}, nil
}

func (h *Handler) List(ctx context.Context, in *partnersvcpb.PartnersListRequest) (*partnersvcpb.PartnersListResponse, error) {
	listIn := &entities.PartnersList{
		IncludeTotalCount: in.GetOptions().GetIncludeTotalCount(),
	}
	if sort := in.GetSort(); sort != nil {
		listIn.Sort = entities.PartnersListSort{
			Field:     entities.PartnersListSortField(sort.GetField()),
			Direction: entities.SortDirection(sort.GetDirection()),
		}
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		listIn.Statuses = coreslices.To(f.GetStatuses(), func(s partnerpb.PartnerStatus) entities.PartnerStatus {
			return entities.PartnerStatus(s)
		})
		if p := f.GetCreatedAt(); p != nil {
			listIn.CreatedAt = converter.Convert(p, &entities.PeriodFilter{}, withUnixTimeCodec)
		}
	}

	result, err := h.svc.List(ctx, listIn)
	if err != nil {
		return nil, MapError(err)
	}

	readMask := in.GetOptions().GetReadMask()
	out := &partnersvcpb.PartnersListResponse{
		Items: coreslices.To(result.Items, func(p *entities.Partner) *partnerpb.Partner {
			item := converter.Convert(p, &partnerpb.Partner{}, withUnixTimeCodec)
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

// ListPublic backs the public website's dealer-network listing (see the
// comment on PartnersService.ListPublic in partner_service.proto) — it
// always forces status=ACTIVE and hand-builds PartnerPublic so
// commissionPercentage/note can never reach an anonymous caller, even if
// this method is later exempted from RBAC.
func (h *Handler) ListPublic(ctx context.Context, _ *partnersvcpb.PartnersListPublicRequest) (*partnersvcpb.PartnersListPublicResponse, error) {
	isPublic := true
	result, err := h.svc.List(ctx, &entities.PartnersList{
		Statuses:   []entities.PartnerStatus{entities.PartnerStatusActive},
		IsPublic:   &isPublic,
		Pagination: entities.ListPagination{Limit: 200},
	})
	if err != nil {
		return nil, MapError(err)
	}
	return &partnersvcpb.PartnersListPublicResponse{
		Items: coreslices.To(result.Items, func(p *entities.Partner) *partnersvcpb.PartnerPublic {
			return &partnersvcpb.PartnerPublic{
				Name:    p.Name,
				Phone:   p.Phone,
				Email:   p.Email,
				Address: p.Address,
			}
		}),
	}, nil
}

func (h *Handler) Update(ctx context.Context, in *partnersvcpb.PartnerUpdateRequest) (*partnersvcpb.PartnerUpdateResponse, error) {
	update := &entities.PartnerUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

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
	return &partnersvcpb.PartnerUpdateResponse{Partner: converter.Convert(p, &partnerpb.Partner{}, withUnixTimeCodec)}, nil
}

func (h *Handler) Delete(ctx context.Context, in *partnersvcpb.PartnerDeleteRequest) (*partnersvcpb.PartnerDeleteResponse, error) {
	del := &entities.PartnerDelete{ID: in.GetId(), Etag: optionalString(in.GetOptions().GetEtag())}
	if err := h.svc.Delete(ctx, del); err != nil {
		return nil, MapError(err)
	}
	return &partnersvcpb.PartnerDeleteResponse{}, nil
}

// MapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message.
func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrPartnerNotFound):
		return status.Error(codes.NotFound, errs.ErrPartnerNotFound.Error())
	case errors.Is(err, errs.ErrPartnerPhoneConflict):
		return status.Error(codes.AlreadyExists, errs.ErrPartnerPhoneConflict.Error())
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
