package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// LocalizedString is a language-code -> text map, mirroring
// crm.types.common.LocalizedString on the wire. It is a plain map so
// converter.Convert bridges it against the storage model directly; the
// proto boundary (map <-> *commonpb.LocalizedString) is bridged by a
// dedicated codec in the transport handler, not here.
type LocalizedString map[string]string

// BrandStatus is int32-aligned with crm.types.brand.BrandStatus so
// converter.Convert maps it both ways as a plain scalar.
type BrandStatus int32

const (
	BrandStatusUnspecified BrandStatus = iota
	BrandStatusActive
	BrandStatusInactive
)

// Brand is a manufacturer/label a product belongs to.
type Brand struct {
	ID          string
	Name        LocalizedString
	Description LocalizedString
	Status      BrandStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // nil = active
	Etag        string     // OCC token; rolled on every write
}

// BrandNew creates a Brand with a fresh ID, timestamps, and etag.
func BrandNew(init ...func(*Brand)) *Brand {
	b := &Brand{ID: uuid.NewString()}
	if len(init) > 0 {
		init[0](b)
	}
	b.UpdateTimestamps()
	b.UpdateEtag()
	return b
}

func (b *Brand) UpdateTimestamps() {
	now := time.Now().UTC()
	b.UpdatedAt = now
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
}

// UpdateEtag rolls a fresh OCC token. go-atlas has no ready-made etag
// generator (checked core/types/etag - it does not exist in this module
// version), so a random UUID is used, matching the ID generation strategy
// already used across the codebase.
func (b *Brand) UpdateEtag() { b.Etag = uuid.NewString() }

func (b *Brand) BeforeUpdate() {
	b.UpdateTimestamps()
	b.UpdateEtag()
}

// BrandCreate is the Create input. Merge applies it onto a freshly
// constructed Brand via the converter.
type BrandCreate struct {
	Name        LocalizedString
	Description LocalizedString
}

func (c *BrandCreate) Merge(dst *Brand) *Brand {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// BrandUpdate is the Update input. Nil fields mean "leave unchanged".
type BrandUpdate struct {
	ID          string `normalize:"trim"`
	Name        LocalizedString
	Description LocalizedString
	Status      *BrandStatus
	Etag        *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *BrandUpdate) Merge(dst *Brand) *Brand {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// BrandDelete is the Delete input.
type BrandDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

// BrandsListSortField is int32-aligned with BrandsListRequest.Sort.Field.
type BrandsListSortField int32

const (
	BrandsListSortFieldCreatedAt BrandsListSortField = iota
	BrandsListSortFieldName
)

type BrandsListSort struct {
	Field     BrandsListSortField
	Direction SortDirection
}

// BrandsList is the single List input; scope/filters/sort/pagination all
// live inside it, per the List(ctx, in *XxxList) convention.
type BrandsList struct {
	Statuses          []BrandStatus
	CreatedAt         *PeriodFilter
	Sort              BrandsListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
