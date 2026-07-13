// Package client defines the Client service interface; the implementation
// lives in the client subpackage.
package client

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates client business rules on top of storages.Storage.
type Service interface {
	Create(ctx context.Context, in *entities.ClientCreate) (*entities.Client, error)
	Get(ctx context.Context, id string) (*entities.Client, error)
	List(ctx context.Context, in *entities.ClientsList) (*entities.List[entities.Client], error)
	Update(ctx context.Context, in *entities.ClientUpdate) (*entities.Client, error)
	Delete(ctx context.Context, in *entities.ClientDelete) error
	// FindOrCreateByEmail returns the active client whose Email matches
	// in.Email, ignoring the rest of in (an existing match is trusted
	// as-is, never overwritten); creates a new client from in when no
	// match exists. Used by SalesService.Create so a seller never has to
	// create a client as a separate step first (TD §12.3).
	FindOrCreateByEmail(ctx context.Context, in *entities.ClientCreate) (*entities.Client, error)
}
