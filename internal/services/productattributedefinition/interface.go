// Package productattributedefinition defines the ProductAttributeDefinition
// service interface; the implementation lives in the
// productattributedefinition subpackage.
package productattributedefinition

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service reads the characteristics catalog. Read-only — no
// Create/Update/Delete, entries are seeded/managed directly in the
// database by a developer.
type Service interface {
	Get(ctx context.Context, id string) (*entities.ProductAttributeDefinition, error)
	List(ctx context.Context, in *entities.ProductAttributeDefinitionsList) (*entities.List[entities.ProductAttributeDefinition], error)
}
