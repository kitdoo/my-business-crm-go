// Package notification implements the notification.Service interface on
// top of mailer.Service.
package notification

import (
	"context"
	"log/slog"

	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/mailer"
	notificationsvc "github.com/kitdoo/my-business-crm-go/internal/services/notification"
)

var _ notificationsvc.Service = (*Service)(nil)

// defaultSubject is used when a caller sends an empty Subject.
const defaultSubject = "Website notification"

// Service is the notification.Service implementation.
type Service struct {
	mailer    mailer.Service
	recipient string
	logger    *slog.Logger
}

// New builds a Service that emails every message to recipient.
func New(mailerSvc mailer.Service, recipient string) *Service {
	return &Service{
		mailer:    mailerSvc,
		recipient: recipient,
		logger:    slog.Default().With(slogx.Module("service:notification")),
	}
}

func (s *Service) Send(ctx context.Context, in *entities.NotificationMessage) error {
	subject := in.Subject
	if subject == "" {
		subject = defaultSubject
	}

	err := s.mailer.Send(ctx, mailer.Message{
		To:      s.recipient,
		Subject: subject,
		Body:    in.Message,
		ReplyTo: in.ReplyTo,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "send notification email", slog.String("error", err.Error()))
		return err
	}
	return nil
}
