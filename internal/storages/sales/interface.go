// Package sales defines the Sale storage interface; the MongoDB
// implementation lives in the mongo subpackage. There is no SoftDelete —
// the TD lists no Delete method for Sale; status moves only through
// Update (backing UpdateStatus/Cancel).
package sales

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists sales.
type Storage interface {
	Insert(ctx context.Context, s *entities.Sale) error
	Get(ctx context.Context, id string) (*entities.Sale, error)
	List(ctx context.Context, in *entities.SalesList) (*entities.List[entities.Sale], error)
	// Update writes s, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match.
	Update(ctx context.Context, s *entities.Sale, oldEtag string) error
}
