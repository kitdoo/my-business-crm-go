// Package entities holds plain domain structs with no database or transport
// annotations. This file collects the primitives shared by every aggregate's
// list/soft-delete inputs so individual entity files only define what is
// specific to their resource.
package entities

import "time"

// SortDirection is int32-aligned with crm.types.common.SortDirection so
// converter.Convert maps it directly (ASC = 0, DESC = 1).
type SortDirection int32

const (
	SortDirectionAsc SortDirection = iota
	SortDirectionDesc
)

// ListPagination carries cursor-based pagination input shared by every
// aggregate's List method.
type ListPagination struct {
	Limit  int64
	Cursor string `normalize:"trim"`
}

// List is the generic paginated result returned by every aggregate's List method.
type List[T any] struct {
	Items      []*T
	Total      int64
	NextCursor string
}

// PeriodFilter is an inclusive, optional Unix-second range shared by every
// aggregate's created-at (or similar) list filter.
type PeriodFilter struct {
	From *time.Time
	To   *time.Time
}

// SoftDelete carries the values the service rolled via BeforeUpdate for a
// single-row soft delete; storage generates nothing.
type SoftDelete struct {
	ID           string
	Etag         string // OCC filter; "" skips the check
	NewUpdatedAt time.Time
	NewEtag      string
}

// BulkSoftDelete carries the values the service rolled for a cascading
// soft delete across multiple rows.
type BulkSoftDelete struct {
	NewUpdatedAt time.Time
	NewEtag      string
}
