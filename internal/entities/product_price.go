package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// ProductPrice is the current price for a product. Only one active row
// exists per VariantID at a time; prior values live in the price history
// log (see storages/prices.Storage.GetHistory), snapshotted on every
// Update.
type ProductPrice struct {
	ID        string
	VariantID string // unique among active rows
	// PriceAmount is in basis points (amount * 100).
	PriceAmount int64
	// Currency is the system-wide ISO 4217 code, stamped at creation from
	// config; never settable via Create/Update.
	Currency       string
	DiscountAmount *int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time // nil = active
	Etag           string     // OCC token; rolled on every write
}

// ProductPriceNew creates a ProductPrice with a fresh ID, timestamps, and etag.
func ProductPriceNew(init ...func(*ProductPrice)) *ProductPrice {
	p := &ProductPrice{ID: uuid.NewString()}
	if len(init) > 0 {
		init[0](p)
	}
	p.UpdateTimestamps()
	p.UpdateEtag()
	return p
}

func (p *ProductPrice) UpdateTimestamps() {
	now := time.Now().UTC()
	p.UpdatedAt = now
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
}

func (p *ProductPrice) UpdateEtag() { p.Etag = uuid.NewString() }

func (p *ProductPrice) BeforeUpdate() {
	p.UpdateTimestamps()
	p.UpdateEtag()
}

// ProductPriceCreate is the Create input. Currency is not part of it — the
// service stamps it from config.
type ProductPriceCreate struct {
	VariantID      string `normalize:"trim"`
	PriceAmount    int64
	DiscountAmount *int64
	Currency       string
}

func (c *ProductPriceCreate) Merge(dst *ProductPrice) *ProductPrice {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// ProductPriceUpdate is the Update input. Nil fields mean "leave unchanged".
// Currency is immutable and has no field here.
type ProductPriceUpdate struct {
	ID             string `normalize:"trim"`
	PriceAmount    *int64
	DiscountAmount *int64
	Etag           *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *ProductPriceUpdate) Merge(dst *ProductPrice) *ProductPrice {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// ProductPriceDelete is the Delete input.
type ProductPriceDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// ProductPriceGetHistory is the GetHistory input.
type ProductPriceGetHistory struct {
	VariantID  string `normalize:"trim"`
	CreatedAt  *PeriodFilter
	Pagination ListPagination
}
