package entities

import (
	"fmt"
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

// PriceUnit is int32-aligned with crm.types.product.PriceUnit so
// converter.Convert maps it both ways as a plain scalar.
type PriceUnit int32

const (
	PriceUnitUnspecified PriceUnit = iota
	PriceUnitPiece
	PriceUnitSquareMeter
)

// Product is a catalog card grouping one or more ProductVariant — it
// carries no sku, images, price, or stock; those live on ProductVariant
// (SKU, ImageIDs) and, per variant, on ProductPrice and Inventory. A
// product is not itself purchasable — see ProductVariant.
type Product struct {
	ID          string
	Name        LocalizedString
	Description LocalizedString
	BrandID     string
	CategoryIDs []string
	// Details is characteristic name -> localized value, shared across
	// every variant (material, collection, …). Variant-differentiating
	// characteristics (color, size, …) live on ProductVariant.Attributes.
	Details   map[string]LocalizedString
	Status    ProductStatus
	// PriceUnit is the unit each SKU's price is quoted per (piece vs square
	// meter) — shown next to the price on the public site.
	PriceUnit PriceUnit
	// HasStock is true when any active ProductVariant of this product has
	// an active ProductSKU with positive Inventory quantity in any
	// warehouse. It is system-maintained — recomputed by
	// inventory.Service.ApplyMovement whenever a movement changes the
	// owning SKU's stock (see that method's doc) — never client-settable,
	// so it carries no proto field of its own; it exists solely to back
	// ProductsListSortFieldInStock.
	HasStock  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // nil = active
	Etag      string     // OCC token; rolled on every write
}

// ProductNew creates a Product with a fresh ID, timestamps, and etag.
// Status defaults to Draft — Create has no status field on the wire, and a
// new product is not yet ready to sell until an admin flips it to Active.
func ProductNew(init ...func(*Product)) *Product {
	p := &Product{ID: uuid.NewString(), Status: ProductStatusDraft}
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
	Name        LocalizedString
	Description LocalizedString
	BrandID     string `normalize:"trim"`
	CategoryIDs []string
	Details     map[string]LocalizedString
	PriceUnit   PriceUnit
}

func (c *ProductCreate) Merge(dst *Product) *Product {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// Validate checks Name for the required locale (see
// LocalizedString.Validate). Description/Details are optional and never
// checked — unlike Name, it's fine to fill in only a non-default locale,
// or none at all, for any of their entries.
func (c *ProductCreate) Validate(requiredLocale string) error {
	if err := c.Name.Validate(requiredLocale); err != nil {
		return fmt.Errorf("name: %w", err)
	}
	return nil
}

// ProductUpdate is the Update input. Nil fields mean "leave unchanged".
// CategoryIDs is a full replacement applied only when present in the
// request's update mask (a repeated field has no wire-level presence), so
// the handler only sets it when requested.
type ProductUpdate struct {
	ID          string `normalize:"trim"`
	Name        LocalizedString
	Description LocalizedString
	BrandID     *string `normalize:"trim,nil_on_empty"`
	CategoryIDs []string
	Details     map[string]LocalizedString
	Status      *ProductStatus
	PriceUnit   *PriceUnit
	Etag        *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *ProductUpdate) Merge(dst *Product) *Product {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// Validate checks Name for the required locale (see
// LocalizedString.Validate). Description/Details are optional and never
// checked — see ProductCreate.Validate.
func (u *ProductUpdate) Validate(requiredLocale string) error {
	if err := u.Name.Validate(requiredLocale); err != nil {
		return fmt.Errorf("name: %w", err)
	}
	return nil
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
	ProductsListSortFieldInStock
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
	IDs               []string
	CreatedAt         *PeriodFilter
	Sort              ProductsListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
