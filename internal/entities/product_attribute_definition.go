package entities

import "time"

// ProductAttributeDefinition is a backend-defined catalog entry for a
// Product.Details key (a "characteristic" — material, color, size, …).
// Read-only from every frontend: there is no admin CRUD UI, entries are
// seeded/managed directly in the database by a developer. IsPublic controls
// whether the key may leave the CRM — web-public filters Product.Details
// down to only the keys of definitions with IsPublic = true before serving
// them to site visitors.
type ProductAttributeDefinition struct {
	ID        string
	Key       string
	Label     LocalizedString
	IsPublic  bool
	SortOrder int32
	CreatedAt time.Time
}

// ProductAttributeDefinitionsList is the single List input.
type ProductAttributeDefinitionsList struct {
	IsPublic   *bool
	Pagination ListPagination
}
