package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// CategoryStatus is int32-aligned with crm.types.category.CategoryStatus so
// converter.Convert maps it both ways as a plain scalar.
type CategoryStatus int32

const (
	CategoryStatusUnspecified CategoryStatus = iota
	CategoryStatusActive
	CategoryStatusInactive
)

// Category is a peer classification a product may belong to; categories are
// not a strict hierarchy (a product may carry several at once), but each
// category may still nest under a ParentID for display grouping.
type Category struct {
	ID          string
	Name        LocalizedString
	Description LocalizedString
	ParentID    *string
	Status      CategoryStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // nil = active
	Etag        string     // OCC token; rolled on every write
}

// CategoryNew creates a Category with a fresh ID, timestamps, and etag.
func CategoryNew(init ...func(*Category)) *Category {
	c := &Category{ID: uuid.NewString()}
	if len(init) > 0 {
		init[0](c)
	}
	c.UpdateTimestamps()
	c.UpdateEtag()
	return c
}

func (c *Category) UpdateTimestamps() {
	now := time.Now().UTC()
	c.UpdatedAt = now
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
}

func (c *Category) UpdateEtag() { c.Etag = uuid.NewString() }

func (c *Category) BeforeUpdate() {
	c.UpdateTimestamps()
	c.UpdateEtag()
}

// CategoryCreate is the Create input. Merge applies it onto a freshly
// constructed Category via the converter.
type CategoryCreate struct {
	Name        LocalizedString
	Description LocalizedString
	ParentID    *string `normalize:"trim,nil_on_empty"`
}

func (c *CategoryCreate) Merge(dst *Category) *Category {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// CategoryUpdate is the Update input. Nil fields mean "leave unchanged".
type CategoryUpdate struct {
	ID          string `normalize:"trim"`
	Name        LocalizedString
	Description LocalizedString
	ParentID    *string `normalize:"trim,nil_on_empty"`
	Status      *CategoryStatus
	Etag        *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *CategoryUpdate) Merge(dst *Category) *Category {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// CategoryDelete is the Delete input.
type CategoryDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// CategoriesListSortField is int32-aligned with CategoriesListRequest.Sort.Field.
type CategoriesListSortField int32

const (
	CategoriesListSortFieldCreatedAt CategoriesListSortField = iota
	CategoriesListSortFieldName
)

type CategoriesListSort struct {
	Field     CategoriesListSortField
	Direction SortDirection
}

// CategoriesList is the single List input; scope/filters/sort/pagination all
// live inside it, per the List(ctx, in *XxxList) convention.
type CategoriesList struct {
	ParentID          *string `normalize:"trim,nil_on_empty"`
	Statuses          []CategoryStatus
	CreatedAt         *PeriodFilter
	Sort              CategoriesListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
