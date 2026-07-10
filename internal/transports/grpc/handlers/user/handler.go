package user

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

	userpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/user"

	usersvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/user/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
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
// entities.User and userpb.User.
var withProtoCodecs = converter.WithCodecs(unixtime.New(), localizedStringCodec())

// withPeriodCodec bridges commonpb.PeriodFilter (*int64 bounds) against
// entities.PeriodFilter (*time.Time bounds).
var withPeriodCodec = converter.WithCodecs(unixtime.New())

// Handler implements usersvcpb.UsersServiceServer.
type Handler struct {
	usersvcpb.UnimplementedUsersServiceServer
	svc usersvc.Service
}

// New builds a Handler.
func New(svc usersvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	usersvcpb.RegisterUsersServiceServer(gs, h)
}

func applyReadMask(mask *fieldmaskpb.FieldMask, msg *userpb.User) {
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

func (h *Handler) Create(ctx context.Context, in *usersvcpb.UserCreateRequest) (*usersvcpb.UserCreateResponse, error) {
	create := converter.Convert(in, &entities.UserCreate{}, withProtoCodecs)

	u, err := h.svc.Create(ctx, create)
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(u, &userpb.User{}, withProtoCodecs)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &usersvcpb.UserCreateResponse{User: out}, nil
}

func (h *Handler) Get(ctx context.Context, in *usersvcpb.UserGetRequest) (*usersvcpb.UserGetResponse, error) {
	u, err := h.svc.Get(ctx, in.GetId())
	if err != nil {
		return nil, MapError(err)
	}
	out := converter.Convert(u, &userpb.User{}, withProtoCodecs)
	applyReadMask(in.GetOptions().GetReadMask(), out)
	return &usersvcpb.UserGetResponse{User: out}, nil
}

func (h *Handler) List(ctx context.Context, in *usersvcpb.UsersListRequest) (*usersvcpb.UsersListResponse, error) {
	listIn := &entities.UsersList{
		IncludeTotalCount: in.GetOptions().GetIncludeTotalCount(),
	}
	if sort := in.GetSort(); sort != nil {
		listIn.Sort = entities.UsersListSort{
			Field:     entities.UsersListSortField(sort.GetField()),
			Direction: entities.SortDirection(sort.GetDirection()),
		}
	}
	if pg := in.GetPagination(); pg != nil {
		listIn.Pagination = entities.ListPagination{Limit: pg.GetLimit(), Cursor: pg.GetCursor()}
	}
	if f := in.GetFilter(); f != nil {
		listIn.Statuses = coreslices.To(f.GetStatuses(), func(s userpb.UserStatus) entities.UserStatus {
			return entities.UserStatus(s)
		})
		listIn.Roles = coreslices.To(f.GetRoles(), func(r userpb.UserRole) entities.UserRole {
			return entities.UserRole(r)
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
	out := &usersvcpb.UsersListResponse{
		Items: coreslices.To(result.Items, func(u *entities.User) *userpb.User {
			item := converter.Convert(u, &userpb.User{}, withProtoCodecs)
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

func (h *Handler) Update(ctx context.Context, in *usersvcpb.UserUpdateRequest) (*usersvcpb.UserUpdateResponse, error) {
	update := &entities.UserUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

	if pbMask := in.GetOptions().GetUpdateMask(); pbMask != nil {
		fm := fieldmask.FromProtoFieldMask(pbMask).Union(fieldmask.FromPaths("id"))
		if err := fm.ApplyUpdateMask(in); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	converter.Convert(in, update, withProtoCodecs, converter.WithIgnoreNilValues())

	u, err := h.svc.Update(ctx, update)
	if err != nil {
		return nil, MapError(err)
	}
	return &usersvcpb.UserUpdateResponse{User: converter.Convert(u, &userpb.User{}, withProtoCodecs)}, nil
}

func (h *Handler) Delete(ctx context.Context, in *usersvcpb.UserDeleteRequest) (*usersvcpb.UserDeleteResponse, error) {
	del := &entities.UserDelete{ID: in.GetId(), Etag: optionalString(in.GetOptions().GetEtag())}
	if err := h.svc.Delete(ctx, del); err != nil {
		return nil, MapError(err)
	}
	return &usersvcpb.UserDeleteResponse{}, nil
}

func (h *Handler) Login(ctx context.Context, in *usersvcpb.UserLoginRequest) (*usersvcpb.UserLoginResponse, error) {
	token, u, err := h.svc.Login(ctx, &entities.UserLogin{Login: in.GetLogin(), Password: in.GetPassword()})
	if err != nil {
		return nil, MapError(err)
	}
	return &usersvcpb.UserLoginResponse{
		Token: token,
		User:  converter.Convert(u, &userpb.User{}, withProtoCodecs),
	}, nil
}

func (h *Handler) ChangePassword(ctx context.Context, in *usersvcpb.UserChangePasswordRequest) (*usersvcpb.UserChangePasswordResponse, error) {
	change := &entities.UserChangePassword{
		ID:              in.GetId(),
		CurrentPassword: in.GetCurrentPassword(),
		NewPassword:     in.GetNewPassword(),
	}
	if err := h.svc.ChangePassword(ctx, change); err != nil {
		return nil, MapError(err)
	}
	return &usersvcpb.UserChangePasswordResponse{}, nil
}

// MapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message.
func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrUserNotFound):
		return status.Error(codes.NotFound, errs.ErrUserNotFound.Error())
	case errors.Is(err, errs.ErrUserPhoneConflict):
		return status.Error(codes.AlreadyExists, errs.ErrUserPhoneConflict.Error())
	case errors.Is(err, errs.ErrUserEmailConflict):
		return status.Error(codes.AlreadyExists, errs.ErrUserEmailConflict.Error())
	case errors.Is(err, errs.ErrUserInvalidCredentials):
		return status.Error(codes.Unauthenticated, errs.ErrUserInvalidCredentials.Error())
	case errors.Is(err, errs.ErrUserInactive):
		return status.Error(codes.FailedPrecondition, errs.ErrUserInactive.Error())
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
