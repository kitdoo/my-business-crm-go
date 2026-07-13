// Package productattributedefinitions defines the ProductAttributeDefinition
// storage interface; the MongoDB implementation lives in the mongo
// subpackage. Read-only — no Create/Update/Delete at all, entries are
// seeded/managed directly in the database by a developer.
package productattributedefinitions

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage reads the characteristics catalog.
type Storage interface {
	// Get returns errs.ErrProductAttributeDefinitionNotFound if id doesn't exist.
	Get(ctx context.Context, id string) (*entities.ProductAttributeDefinition, error)
	List(ctx context.Context, in *entities.ProductAttributeDefinitionsList) (*entities.List[entities.ProductAttributeDefinition], error)
}
