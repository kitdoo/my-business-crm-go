package entities

import "time"

// Image is metadata for a file uploaded through the image handler
// (internal/transports/http/handlers/image). The file's bytes live on disk
// at <dir>/<ID> (no extension); this record is what lets serving pick the
// right Content-Type without re-sniffing magic bytes on every request, and
// gives any other consumer (e.g. Product.ImageIDs) a place to learn the
// format of an id it's holding.
type Image struct {
	ID          string
	ContentType string
	SizeBytes   int64
	CreatedAt   time.Time
}
