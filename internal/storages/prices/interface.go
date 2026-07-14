// Package prices defines the ProductPrice storage interface; the MongoDB
// implementation lives in the mongo subpackage.
package prices

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists the current price per variant plus an append-only price
// history log. All reads against the current price filter out soft-deleted
// rows.
type Storage interface {
	// Insert returns errs.ErrProductPriceExists when variantId already has
	// an active price.
	Insert(ctx context.Context, p *entities.ProductPrice) error
	// Get looks up the current price by its own id (used by Update/Delete,
	// which address a price by id).
	Get(ctx context.Context, id string) (*entities.ProductPrice, error)
	// GetByVariantID looks up the current price for a variant.
	GetByVariantID(ctx context.Context, variantID string) (*entities.ProductPrice, error)
	// Update writes p, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match.
	Update(ctx context.Context, p *entities.ProductPrice, oldEtag string) error
	// SoftDelete hides a price by stamping deleted_at. The caller supplies
	// the new timestamp and etag via in, so this method generates nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
	// AppendHistory records snapshot into the append-only price history log.
	// The caller supplies a fully-populated snapshot (a fresh ID, the
	// pre-change values); this method generates nothing.
	AppendHistory(ctx context.Context, snapshot *entities.ProductPrice) error
	// GetHistory lists price history log entries for a variant.
	GetHistory(ctx context.Context, in *entities.ProductPriceGetHistory) (*entities.List[entities.ProductPrice], error)
}
