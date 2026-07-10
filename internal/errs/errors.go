// Package errs holds sentinel errors shared across services, storages, and
// transports. Compare with errors.Is, never string equality.
package errs

import "errors"

var (
	// ErrBrandNotFound is returned when a brand lookup finds no active row.
	ErrBrandNotFound = errors.New("brand not found")

	// ErrBrandHasProducts is returned when a brand delete is attempted while
	// products still reference it.
	ErrBrandHasProducts = errors.New("brand has products")

	// ErrStaleEntity is returned when a write's OCC precondition (etag) no
	// longer matches the stored row.
	ErrStaleEntity = errors.New("stale entity")

	// ErrInvalidListCursor is returned when a list cursor fails validation.
	ErrInvalidListCursor = errors.New("invalid list cursor")

	// ErrInvalidArgument is returned when a request argument fails domain
	// validation.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrNotImplemented marks a method whose body is a skeleton per the
	// service development standard.
	ErrNotImplemented = errors.New("not implemented")
)
