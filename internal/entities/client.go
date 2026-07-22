package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// Client is a retail buyer. Name/Address/Note are factual data, not
// localized.
type Client struct {
	ID string
	// Phone is unique across active clients.
	Phone     string
	Name      string
	Email     string
	Address   string
	Note      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // nil = active
	Etag      string     // OCC token; rolled on every write
}

// ClientNew creates a Client with a fresh ID, timestamps, and etag.
func ClientNew(init ...func(*Client)) *Client {
	c := &Client{ID: uuid.NewString()}
	if len(init) > 0 {
		init[0](c)
	}
	c.UpdateTimestamps()
	c.UpdateEtag()
	return c
}

func (c *Client) UpdateTimestamps() {
	now := time.Now().UTC()
	c.UpdatedAt = now
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
}

func (c *Client) UpdateEtag() { c.Etag = uuid.NewString() }

func (c *Client) BeforeUpdate() {
	c.UpdateTimestamps()
	c.UpdateEtag()
}

// ClientCreate is the Create input. Merge applies it onto a freshly
// constructed Client via the converter.
type ClientCreate struct {
	Name    string `normalize:"trim"`
	Phone   string `normalize:"trim"`
	Email   string `normalize:"trim"`
	Address string `normalize:"trim"`
	Note    string `normalize:"trim"`
}

func (c *ClientCreate) Merge(dst *Client) *Client {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// ClientUpdate is the Update input. Nil fields mean "leave unchanged".
type ClientUpdate struct {
	ID      string  `normalize:"trim"`
	Name    *string `normalize:"trim,nil_on_empty"`
	Phone   *string `normalize:"trim,nil_on_empty"`
	Email   *string `normalize:"trim,nil_on_empty"`
	Address *string `normalize:"trim,nil_on_empty"`
	Note    *string `normalize:"trim,nil_on_empty"`
	Etag    *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *ClientUpdate) Merge(dst *Client) *Client {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// ClientDelete is the Delete input.
type ClientDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// ClientsListSortField is int32-aligned with ClientsListRequest.Sort.Field.
type ClientsListSortField int32

const (
	ClientsListSortFieldCreatedAt ClientsListSortField = iota
	ClientsListSortFieldName
)

type ClientsListSort struct {
	Field     ClientsListSortField
	Direction SortDirection
}

// ClientsList is the single List input; scope/filters/sort/pagination all
// live inside it, per the List(ctx, in *XxxList) convention.
type ClientsList struct {
	// Email is an exact match — used to find a client by email before
	// creating a sale (SalesService.Create's find-or-create).
	Email             *string `normalize:"trim,nil_on_empty"`
	IDs               []string
	CreatedAt         *PeriodFilter
	Sort              ClientsListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
