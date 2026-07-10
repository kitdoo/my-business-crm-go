// Package clients defines the Client storage interface; the MongoDB
// implementation lives in the mongo subpackage.
package clients

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists clients. All reads filter out soft-deleted rows.
type Storage interface {
	// Insert returns errs.ErrClientPhoneConflict when phone collides with
	// another active client.
	Insert(ctx context.Context, c *entities.Client) error
	Get(ctx context.Context, id string) (*entities.Client, error)
	List(ctx context.Context, in *entities.ClientsList) (*entities.List[entities.Client], error)
	// Update writes c, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match, or
	// errs.ErrClientPhoneConflict on a phone collision.
	Update(ctx context.Context, c *entities.Client, oldEtag string) error
	// SoftDelete hides a client by stamping deleted_at. The caller supplies
	// the new timestamp and etag via in, so this method generates nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
}
