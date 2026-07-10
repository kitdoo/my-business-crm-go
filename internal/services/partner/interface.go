// Package partner defines the Partner service interface; the implementation
// lives in the partner subpackage.
package partner

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates partner business rules on top of storages.Storage.
type Service interface {
	Create(ctx context.Context, in *entities.PartnerCreate) (*entities.Partner, error)
	Get(ctx context.Context, id string) (*entities.Partner, error)
	List(ctx context.Context, in *entities.PartnersList) (*entities.List[entities.Partner], error)
	Update(ctx context.Context, in *entities.PartnerUpdate) (*entities.Partner, error)
	Delete(ctx context.Context, in *entities.PartnerDelete) error
}
