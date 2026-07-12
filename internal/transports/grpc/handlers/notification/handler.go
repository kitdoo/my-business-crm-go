package notification

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notificationsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/notification/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	notificationsvc "github.com/kitdoo/my-business-crm-go/internal/services/notification"
)

// Handler implements notificationsvcpb.NotificationsServiceServer.
type Handler struct {
	notificationsvcpb.UnimplementedNotificationsServiceServer
	svc notificationsvc.Service
}

// New builds a Handler.
func New(svc notificationsvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	notificationsvcpb.RegisterNotificationsServiceServer(gs, h)
}

func (h *Handler) Send(ctx context.Context, in *notificationsvcpb.SendMessageRequest) (*notificationsvcpb.SendMessageResponse, error) {
	err := h.svc.Send(ctx, &entities.NotificationMessage{
		Subject: in.GetSubject(),
		Message: in.GetMessage(),
		ReplyTo: in.GetEmail(),
	})
	if err != nil {
		return nil, MapError(err)
	}
	return &notificationsvcpb.SendMessageResponse{}, nil
}

// MapError maps domain sentinels to gRPC status codes. Default is
// codes.Internal with an opaque message.
func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errs.ErrSMTPNotConfigured):
		return status.Error(codes.Unavailable, "email delivery is not configured")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
