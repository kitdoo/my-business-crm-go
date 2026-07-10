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

	// ErrProductNotFound is returned when a product lookup finds no active row.
	ErrProductNotFound = errors.New("product not found")

	// ErrProductSKUConflict is returned when a product's SKU collides with
	// another active product's SKU.
	ErrProductSKUConflict = errors.New("product sku conflict")

	// ErrProductBrandNotFound is returned when a product's brandId does not
	// reference an existing brand.
	ErrProductBrandNotFound = errors.New("product brand not found")

	// ErrProductCategoryNotFound is returned when one of a product's
	// categoryIds does not reference an existing category.
	ErrProductCategoryNotFound = errors.New("product category not found")

	// ErrProductPriceNotFound is returned when a price lookup finds no
	// active row.
	ErrProductPriceNotFound = errors.New("product price not found")

	// ErrProductPriceExists is returned when Create is attempted for a
	// product that already has an active price.
	ErrProductPriceExists = errors.New("product price already exists")

	// ErrInventoryNotFound is returned when no inventory row exists yet for
	// a (productId, warehouseId) pair.
	ErrInventoryNotFound = errors.New("inventory not found")

	// ErrInsufficientStock is returned when a movement's delta would drive
	// quantity below zero.
	ErrInsufficientStock = errors.New("insufficient stock")

	// ErrSaleNotFound is returned when a sale lookup finds no row.
	ErrSaleNotFound = errors.New("sale not found")

	// ErrSaleWarehouseInactive is returned when Create targets a warehouse
	// whose status is not active.
	ErrSaleWarehouseInactive = errors.New("sale warehouse inactive")

	// ErrSaleTerminalStatus is returned when UpdateStatus or Cancel is
	// attempted on a sale already in a terminal status (cancelled/refunded).
	ErrSaleTerminalStatus = errors.New("sale already in a terminal status")

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
