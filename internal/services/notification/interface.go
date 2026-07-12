// Package notification defines the Service interface for the generic
// outbound-message sender; the implementation lives in the notification
// subpackage.
package notification

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service delivers an already-composed message (from any approved
// caller — see internal/transports/grpc/interceptors/clientkey) to the
// configured recipient (appconfig.SMTPConfig.To) by email.
type Service interface {
	Send(ctx context.Context, in *entities.NotificationMessage) error
}
