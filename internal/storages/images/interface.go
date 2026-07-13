// Package images defines the Image metadata storage interface; the
// MongoDB implementation lives in the mongo subpackage.
package images

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists uploaded-image metadata (content type, size), keyed by
// the same id the file is stored under on disk. Write-once: there is no
// Update or Delete — an id's bytes and format never change once uploaded
// (see ProductImageUploader's "remove only drops the id" contract).
type Storage interface {
	Insert(ctx context.Context, img *entities.Image) error
	Get(ctx context.Context, id string) (*entities.Image, error)
}
