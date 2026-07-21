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

	// ErrUserAdminProtected is returned when Delete is attempted against a
	// user whose role is admin — an admin account can never be deleted
	// through the regular Delete RPC, regardless of caller role, so a
	// locked-out operator always has at least one account left to recover
	// through.
	ErrUserAdminProtected = errors.New("admin user cannot be deleted")

	// ErrUserAdminCreateForbidden is returned when Create or Update is
	// attempted with role admin — the only admin account is the one
	// materialized from CRMConfig.BootstrapAdmin on first boot (see
	// internal/fx.registerBootstrapAdmin); there is no RPC path to mint a
	// second one.
	ErrUserAdminCreateForbidden = errors.New("admin role cannot be assigned through this endpoint")

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

	// ErrProductVariantNotFound is returned when a variant lookup finds no
	// active row.
	ErrProductVariantNotFound = errors.New("product variant not found")

	// ErrProductVariantProductNotFound is returned when a variant's
	// productId does not reference an existing product.
	ErrProductVariantProductNotFound = errors.New("product variant product not found")

	// ErrProductSkuNotFound is returned when a SKU lookup finds no active row.
	ErrProductSkuNotFound = errors.New("product sku not found")

	// ErrProductSkuSKUConflict is returned when a SKU's sku string collides
	// with another active SKU's.
	ErrProductSkuSKUConflict = errors.New("product sku conflict")

	// ErrProductSkuVariantNotFound is returned when a SKU's variantId does
	// not reference an existing variant.
	ErrProductSkuVariantNotFound = errors.New("product sku variant not found")

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

	// ErrProductAttributeDefinitionNotFound is returned when a
	// characteristics-catalog lookup finds no row for the given id.
	ErrProductAttributeDefinitionNotFound = errors.New("product attribute definition not found")

	// ErrSaleNotFound is returned when a sale lookup finds no row.
	ErrSaleNotFound = errors.New("sale not found")

	// ErrSaleWarehouseInactive is returned when Create targets a warehouse
	// whose status is not active.
	ErrSaleWarehouseInactive = errors.New("sale warehouse inactive")

	// ErrSaleTerminalStatus is returned when UpdateStatus or Cancel is
	// attempted on a sale already in a terminal status (cancelled/refunded).
	ErrSaleTerminalStatus = errors.New("sale already in a terminal status")

	// ErrSaleMissingClientOrPartner is returned when Create names neither a
	// client (clientId/newClient) nor a partner — every sale needs at
	// least one counterparty.
	ErrSaleMissingClientOrPartner = errors.New("sale must have a client or a partner")

	// ErrUnauthenticated is returned when a request carries no valid session
	// token (missing, malformed, or not matching an active user). Generic
	// across every non-gRPC endpoint that needs auth (currently only image
	// upload); gRPC handlers have their own sentinels for this.
	ErrUnauthenticated = errors.New("unauthenticated")

	// ErrForbidden is returned when an authenticated caller's role does not
	// permit the requested action.
	ErrForbidden = errors.New("forbidden")

	// ErrStaleEntity is returned when a write's OCC precondition (etag) no
	// longer matches the stored row.
	ErrStaleEntity = errors.New("stale entity")

	// ErrInvalidListCursor is returned when a list cursor fails validation.
	ErrInvalidListCursor = errors.New("invalid list cursor")

	// ErrInvalidArgument is returned when a request argument fails domain
	// validation.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrLocalizedStringMissingRequiredLocale is returned when a
	// LocalizedString value is set but does not include the configured
	// required locale (CRMConfig.DefaultLocale), which every entity with a
	// LocalizedString field requires (see
	// entities.LocalizedString.Validate). The message deliberately does
	// not name the locale itself — every gRPC handler's MapError returns
	// this sentinel's own static text verbatim, discarding anything
	// wrapped around it, so the specific locale could never reach the
	// caller through that path anyway.
	ErrLocalizedStringMissingRequiredLocale = errors.New("localized string missing required locale")

	// ErrNotImplemented marks a method whose body is a skeleton per the
	// service development standard.
	ErrNotImplemented = errors.New("not implemented")

	// ErrMailerNotConfigured is returned by mailer.Service.Send when
	// neither CRMConfig.Resend nor CRMConfig.Smtp is set — form
	// submissions cannot be delivered anywhere until an operator
	// configures one of them.
	ErrMailerNotConfigured = errors.New("mailer not configured")

	// ErrImageNotFound is returned when an uploaded-image metadata lookup
	// finds no row for the given id.
	ErrImageNotFound = errors.New("image not found")
)
