package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// PartnerStatus is int32-aligned with crm.types.partner.PartnerStatus so
// converter.Convert maps it both ways as a plain scalar.
type PartnerStatus int32

const (
	PartnerStatusUnspecified PartnerStatus = iota
	PartnerStatusActive
	PartnerStatusInactive
)

// Partner is a reseller who earns a commission on sales made under their
// account. Name/Phone/Note are factual data, not localized.
type Partner struct {
	ID   string
	Name string
	// Phone is unique across active partners.
	Phone string
	Email string
	// CommissionPercentage applied to Sale.TotalAmount, 0-100.
	CommissionPercentage int32
	Note                 string
	Status               PartnerStatus
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time // nil = active
	Etag                 string     // OCC token; rolled on every write
}

// PartnerNew creates a Partner with a fresh ID, timestamps, and etag.
// Status defaults to Active — Create has no status field on the wire, so
// this is the only place a newly created partner's status is set.
func PartnerNew(init ...func(*Partner)) *Partner {
	p := &Partner{ID: uuid.NewString(), Status: PartnerStatusActive}
	if len(init) > 0 {
		init[0](p)
	}
	p.UpdateTimestamps()
	p.UpdateEtag()
	return p
}

func (p *Partner) UpdateTimestamps() {
	now := time.Now().UTC()
	p.UpdatedAt = now
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
}

func (p *Partner) UpdateEtag() { p.Etag = uuid.NewString() }

func (p *Partner) BeforeUpdate() {
	p.UpdateTimestamps()
	p.UpdateEtag()
}

// PartnerCreate is the Create input. Merge applies it onto a freshly
// constructed Partner via the converter.
type PartnerCreate struct {
	Name                 string `normalize:"trim"`
	Phone                string `normalize:"trim"`
	Email                string `normalize:"trim"`
	CommissionPercentage int32
	Note                 string `normalize:"trim"`
}

func (c *PartnerCreate) Merge(dst *Partner) *Partner {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// PartnerUpdate is the Update input. Nil fields mean "leave unchanged".
type PartnerUpdate struct {
	ID                   string  `normalize:"trim"`
	Name                 *string `normalize:"trim,nil_on_empty"`
	Phone                *string `normalize:"trim,nil_on_empty"`
	Email                *string `normalize:"trim,nil_on_empty"`
	CommissionPercentage *int32
	Note                 *string `normalize:"trim,nil_on_empty"`
	Status               *PartnerStatus
	Etag                 *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *PartnerUpdate) Merge(dst *Partner) *Partner {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// PartnerDelete is the Delete input.
type PartnerDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// PartnersListSortField is int32-aligned with PartnersListRequest.Sort.Field.
type PartnersListSortField int32

const (
	PartnersListSortFieldCreatedAt PartnersListSortField = iota
	PartnersListSortFieldName
)

type PartnersListSort struct {
	Field     PartnersListSortField
	Direction SortDirection
}

// PartnersList is the single List input; scope/filters/sort/pagination all
// live inside it, per the List(ctx, in *XxxList) convention.
type PartnersList struct {
	Statuses          []PartnerStatus
	CreatedAt         *PeriodFilter
	Sort              PartnersListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
