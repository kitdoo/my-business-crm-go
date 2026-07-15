package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// ProductSkuStatus is int32-aligned with
// crm.types.product_sku.ProductSkuStatus so converter.Convert maps it
// both ways as a plain scalar.
type ProductSkuStatus int32

const (
	ProductSkuStatusUnspecified ProductSkuStatus = iota
	ProductSkuStatusDraft
	ProductSkuStatusActive
	ProductSkuStatusInactive
)

// ProductSKU is the purchasable unit of a ProductVariant — one sku, one
// set of characteristics that affect price or availability (size,
// thickness, packaging, …). Price lives in ProductPrice and stock in
// Inventory, both keyed by SKUID (not VariantID).
type ProductSKU struct {
	ID        string
	VariantID string
	SKU       string // unique among active rows
	// Attributes is characteristic name -> localized value, affecting
	// price/availability (size, thickness, packaging, …).
	// ProductVariant.Attributes holds characteristics that change
	// appearance instead.
	Attributes map[string]LocalizedString
	Status     ProductSkuStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time // nil = active
	Etag       string     // OCC token; rolled on every write
}

// ProductSKUNew creates a ProductSKU with a fresh ID, timestamps, and
// etag. Status defaults to Draft, matching ProductVariant's
// ProductVariantNew.
func ProductSKUNew(init ...func(*ProductSKU)) *ProductSKU {
	s := &ProductSKU{ID: uuid.NewString(), Status: ProductSkuStatusDraft}
	if len(init) > 0 {
		init[0](s)
	}
	s.UpdateTimestamps()
	s.UpdateEtag()
	return s
}

func (s *ProductSKU) UpdateTimestamps() {
	now := time.Now().UTC()
	s.UpdatedAt = now
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
}

func (s *ProductSKU) UpdateEtag() { s.Etag = uuid.NewString() }

func (s *ProductSKU) BeforeUpdate() {
	s.UpdateTimestamps()
	s.UpdateEtag()
}

// ProductSKUCreate is the Create input. Merge applies it onto a freshly
// constructed ProductSKU via the converter.
type ProductSKUCreate struct {
	VariantID  string `normalize:"trim"`
	SKU        string `normalize:"trim"`
	Attributes map[string]LocalizedString
}

func (c *ProductSKUCreate) Merge(dst *ProductSKU) *ProductSKU {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// Validate requires VariantID and SKU; Attributes is optional and never
// checked, same as ProductVariantCreate.
func (c *ProductSKUCreate) Validate() error {
	if c.VariantID == "" {
		return fmt.Errorf("variantId: required")
	}
	if c.SKU == "" {
		return fmt.Errorf("sku: required")
	}
	return nil
}

// ProductSKUUpdate is the Update input. Nil fields mean "leave
// unchanged". VariantID/SKU are immutable and have no field here.
type ProductSKUUpdate struct {
	ID         string `normalize:"trim"`
	Attributes map[string]LocalizedString
	Status     *ProductSkuStatus
	Etag       *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *ProductSKUUpdate) Merge(dst *ProductSKU) *ProductSKU {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// ProductSKUDelete is the Delete input.
type ProductSKUDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// ProductSKUsListSortField is int32-aligned with
// ProductSKUsListRequest.Sort.Field.
type ProductSKUsListSortField int32

const (
	ProductSKUsListSortFieldCreatedAt ProductSKUsListSortField = iota
	ProductSKUsListSortFieldSKU
)

type ProductSKUsListSort struct {
	Field     ProductSKUsListSortField
	Direction SortDirection
}

// ProductSKUsList is the single List input; scope/filters/sort/pagination
// all live inside it, per the List(ctx, in *XxxList) convention.
type ProductSKUsList struct {
	VariantID         *string `normalize:"trim,nil_on_empty"`
	Statuses          []ProductSkuStatus
	VariantIDs        []string
	SKUs              []string
	CreatedAt         *PeriodFilter
	Sort              ProductSKUsListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
