package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// ProductVariantStatus is int32-aligned with
// crm.types.product_variant.ProductVariantStatus so converter.Convert maps
// it both ways as a plain scalar.
type ProductVariantStatus int32

const (
	ProductVariantStatusUnspecified ProductVariantStatus = iota
	ProductVariantStatusDraft
	ProductVariantStatusActive
	ProductVariantStatusInactive
)

// ProductVariant is the purchasable unit of a Product — one sku, one set
// of images, one set of differentiating characteristics (color, size,
// …). Price lives in ProductPrice and stock in Inventory, both keyed by
// VariantID (not ProductID).
type ProductVariant struct {
	ID        string
	ProductID string
	SKU       string // unique among active rows
	// Attributes is characteristic name -> localized value, differentiating
	// this variant from its siblings (color, size, …). Product.Details
	// holds characteristics shared by every variant instead.
	Attributes map[string]LocalizedString
	// ImageIDs are display order, first is main; populated via Update only.
	ImageIDs  []string
	Status    ProductVariantStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // nil = active
	Etag      string     // OCC token; rolled on every write
}

// ProductVariantNew creates a ProductVariant with a fresh ID, timestamps,
// and etag. Status defaults to Draft, matching Product's ProductNew.
func ProductVariantNew(init ...func(*ProductVariant)) *ProductVariant {
	v := &ProductVariant{ID: uuid.NewString(), Status: ProductVariantStatusDraft}
	if len(init) > 0 {
		init[0](v)
	}
	v.UpdateTimestamps()
	v.UpdateEtag()
	return v
}

func (v *ProductVariant) UpdateTimestamps() {
	now := time.Now().UTC()
	v.UpdatedAt = now
	if v.CreatedAt.IsZero() {
		v.CreatedAt = now
	}
}

func (v *ProductVariant) UpdateEtag() { v.Etag = uuid.NewString() }

func (v *ProductVariant) BeforeUpdate() {
	v.UpdateTimestamps()
	v.UpdateEtag()
}

// ProductVariantCreate is the Create input. Merge applies it onto a
// freshly constructed ProductVariant via the converter.
type ProductVariantCreate struct {
	ProductID  string `normalize:"trim"`
	SKU        string `normalize:"trim"`
	Attributes map[string]LocalizedString
}

func (c *ProductVariantCreate) Merge(dst *ProductVariant) *ProductVariant {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// Validate requires ProductID and SKU; Attributes is optional and never
// checked, same as Product.Details.
func (c *ProductVariantCreate) Validate() error {
	if c.ProductID == "" {
		return fmt.Errorf("productId: required")
	}
	if c.SKU == "" {
		return fmt.Errorf("sku: required")
	}
	return nil
}

// ProductVariantUpdate is the Update input. Nil fields mean "leave
// unchanged". ProductID/SKU are immutable and have no field here.
// ImageIDs is a full replacement applied only when present in the
// request's update mask (a repeated field has no wire-level presence), so
// the handler only sets it when requested.
type ProductVariantUpdate struct {
	ID         string `normalize:"trim"`
	Attributes map[string]LocalizedString
	ImageIDs   []string
	Status     *ProductVariantStatus
	Etag       *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *ProductVariantUpdate) Merge(dst *ProductVariant) *ProductVariant {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// ProductVariantDelete is the Delete input.
type ProductVariantDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// ProductVariantsListSortField is int32-aligned with
// ProductVariantsListRequest.Sort.Field.
type ProductVariantsListSortField int32

const (
	ProductVariantsListSortFieldCreatedAt ProductVariantsListSortField = iota
	ProductVariantsListSortFieldSKU
)

type ProductVariantsListSort struct {
	Field     ProductVariantsListSortField
	Direction SortDirection
}

// ProductVariantsList is the single List input; scope/filters/sort/
// pagination all live inside it, per the List(ctx, in *XxxList) convention.
type ProductVariantsList struct {
	ProductID         *string `normalize:"trim,nil_on_empty"`
	Statuses          []ProductVariantStatus
	ProductIDs        []string
	SKUs              []string
	CreatedAt         *PeriodFilter
	Sort              ProductVariantsListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
