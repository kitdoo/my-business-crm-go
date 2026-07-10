package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// ProductStatus is int32-aligned with crm.types.product.ProductStatus so
// converter.Convert maps it both ways as a plain scalar.
type ProductStatus int32

const (
	ProductStatusUnspecified ProductStatus = iota
	ProductStatusDraft
	ProductStatusActive
	ProductStatusInactive
)

// Product carries no price or stock — those live in ProductPrice and
// Inventory.
type Product struct {
	ID          string
	SKU         string // unique
	Name        LocalizedString
	Description LocalizedString
	BrandID     string
	CategoryIDs []string
	// Details is characteristic name -> localized value (material, size, …).
	Details map[string]LocalizedString
	// ImageIDs are display order, first is main; populated via Update only.
	ImageIDs  []string
	Status    ProductStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // nil = active
	Etag      string     // OCC token; rolled on every write
}

// ProductNew creates a Product with a fresh ID, timestamps, and etag.
func ProductNew(init ...func(*Product)) *Product {
	p := &Product{ID: uuid.NewString()}
	if len(init) > 0 {
		init[0](p)
	}
	p.UpdateTimestamps()
	p.UpdateEtag()
	return p
}

func (p *Product) UpdateTimestamps() {
	now := time.Now().UTC()
	p.UpdatedAt = now
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
}

func (p *Product) UpdateEtag() { p.Etag = uuid.NewString() }

func (p *Product) BeforeUpdate() {
	p.UpdateTimestamps()
	p.UpdateEtag()
}

// ProductCreate is the Create input. Merge applies it onto a freshly
// constructed Product via the converter.
type ProductCreate struct {
	SKU         string `normalize:"trim"`
	Name        LocalizedString
	Description LocalizedString
	BrandID     string `normalize:"trim"`
	CategoryIDs []string
	Details     map[string]LocalizedString
}

func (c *ProductCreate) Merge(dst *Product) *Product {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// ProductUpdate is the Update input. Nil fields mean "leave unchanged". SKU
// is immutable and has no field here. CategoryIDs/ImageIDs are full
// replacements applied only when present in the request's update mask (a
// repeated field has no wire-level presence), so the handler only sets them
// when requested.
type ProductUpdate struct {
	ID          string `normalize:"trim"`
	Name        LocalizedString
	Description LocalizedString
	BrandID     *string `normalize:"trim,nil_on_empty"`
	CategoryIDs []string
	Details     map[string]LocalizedString
	ImageIDs    []string
	Status      *ProductStatus
	Etag        *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *ProductUpdate) Merge(dst *Product) *Product {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// ProductDelete is the Delete input.
type ProductDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// ProductsListSortField is int32-aligned with ProductsListRequest.Sort.Field.
type ProductsListSortField int32

const (
	ProductsListSortFieldCreatedAt ProductsListSortField = iota
	ProductsListSortFieldName
)

type ProductsListSort struct {
	Field     ProductsListSortField
	Direction SortDirection
}

// ProductsList is the single List input; scope/filters/sort/pagination all
// live inside it, per the List(ctx, in *XxxList) convention.
type ProductsList struct {
	BrandID           *string `normalize:"trim,nil_on_empty"`
	CategoryID        *string `normalize:"trim,nil_on_empty"`
	Statuses          []ProductStatus
	BrandIDs          []string
	CategoryIDs       []string
	SKUs              []string
	CreatedAt         *PeriodFilter
	Sort              ProductsListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
