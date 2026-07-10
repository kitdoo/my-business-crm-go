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

	// ErrCategoryNotFound is returned when a category lookup finds no active row.
	ErrCategoryNotFound = errors.New("category not found")

	// ErrCategoryHasProducts is returned when a category delete is attempted
	// while products still reference it.
	ErrCategoryHasProducts = errors.New("category has products")

	// ErrCategoryParentNotFound is returned when a category's parentId does
	// not reference an existing category.
	ErrCategoryParentNotFound = errors.New("category parent not found")

	// ErrCategorySelfParent is returned when a category's parentId points to
	// itself.
	ErrCategorySelfParent = errors.New("category cannot be its own parent")

	// ErrWarehouseNotFound is returned when a warehouse lookup finds no
	// active row.
	ErrWarehouseNotFound = errors.New("warehouse not found")

	// ErrWarehouseHasStock is returned when a warehouse delete or deactivate
	// is attempted while it still carries inventory.
	ErrWarehouseHasStock = errors.New("warehouse has stock")

	// ErrPartnerNotFound is returned when a partner lookup finds no active row.
	ErrPartnerNotFound = errors.New("partner not found")

	// ErrPartnerPhoneConflict is returned when a partner's phone collides
	// with another active partner's phone.
	ErrPartnerPhoneConflict = errors.New("partner phone conflict")

	// ErrClientNotFound is returned when a client lookup finds no active row.
	ErrClientNotFound = errors.New("client not found")

	// ErrClientPhoneConflict is returned when a client's phone collides with
	// another active client's phone.
	ErrClientPhoneConflict = errors.New("client phone conflict")

	// ErrUserNotFound is returned when a user lookup finds no active row.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserPhoneConflict is returned when a user's phone collides with
	// another active user's phone.
	ErrUserPhoneConflict = errors.New("user phone conflict")

	// ErrUserEmailConflict is returned when a user's email collides with
	// another active user's email.
	ErrUserEmailConflict = errors.New("user email conflict")

	// ErrUserInvalidCredentials is returned when Login or ChangePassword is
	// given a login/password pair that does not match a stored user.
	ErrUserInvalidCredentials = errors.New("invalid credentials")

	// ErrUserInactive is returned when Login is attempted against a user
	// whose status is not active.
	ErrUserInactive = errors.New("user inactive")

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
