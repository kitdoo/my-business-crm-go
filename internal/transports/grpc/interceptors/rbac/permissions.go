package rbac

import (
	brandsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/brand/v1"
	categorysvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/category/v1"
	clientsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/client/v1"
	inventorysvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/inventory/v1"
	inventorymovementsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/inventory_movement/v1"
	partnersvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/partner/v1"
	pricesvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/price/v1"
	productsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/product/v1"
	productattributedefinitionsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/product_attribute_definition/v1"
	productvariantsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/product_variant/v1"
	reportsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/report/v1"
	salesvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/sale/v1"
	usersvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/user/v1"
	warehousesvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/warehouse/v1"

	serviceinfopb "github.com/altessa-s/proto-gen-go/serviceinfo/v1"
)

const (
	actionRead   = "read"
	actionCreate = "create"
	actionUpdate = "update"
	actionDelete = "delete"

	// permissionAuthenticatedOnly is a sentinel permission value meaning
	// "any authenticated caller, regardless of role" — for self-service
	// calls where the request itself already proves the caller's right to
	// act (ChangePassword requires the current password) or diagnostic
	// endpoints that carry no business data (ServiceInfo).
	permissionAuthenticatedOnly = ""
)

// permission builds the "<resource>:<action>" strings used as both the
// permissions map's values here and the entries operators list under a
// role in CRMConfig.RBAC / configs/crm.yaml (e.g. "products:read").
func permission(resource, action string) string { return resource + ":" + action }

// permissions maps every gRPC method that requires authentication (i.e.
// every method the auth interceptor does not exempt — see
// internal/transports/grpc/interceptors/auth) to the permission string
// internal/rbac.Table checks the caller's role against. A method absent
// from this map is denied for every role except one holding the "*"
// wildcard (see Table.Allowed) — fail closed for anything new that lands
// here without an explicit entry.
//
// Login, ProductsService.List, PricesService.Get, CategoriesService.List
// and ProductAttributeDefinitionsService.List are intentionally absent: the
// auth interceptor exempts them outright (see auth.New), so they never
// reach the RBAC check.
var permissions = map[string]string{
	brandsvcpb.BrandsService_Create_FullMethodName: permission("brands", actionCreate),
	brandsvcpb.BrandsService_Get_FullMethodName:    permission("brands", actionRead),
	brandsvcpb.BrandsService_List_FullMethodName:   permission("brands", actionRead),
	brandsvcpb.BrandsService_Update_FullMethodName: permission("brands", actionUpdate),
	brandsvcpb.BrandsService_Delete_FullMethodName: permission("brands", actionDelete),

	categorysvcpb.CategoriesService_Create_FullMethodName: permission("categories", actionCreate),
	categorysvcpb.CategoriesService_Get_FullMethodName:    permission("categories", actionRead),
	categorysvcpb.CategoriesService_List_FullMethodName:   permission("categories", actionRead),
	categorysvcpb.CategoriesService_Update_FullMethodName: permission("categories", actionUpdate),
	categorysvcpb.CategoriesService_Delete_FullMethodName: permission("categories", actionDelete),

	clientsvcpb.ClientsService_Create_FullMethodName: permission("clients", actionCreate),
	clientsvcpb.ClientsService_Get_FullMethodName:    permission("clients", actionRead),
	clientsvcpb.ClientsService_List_FullMethodName:   permission("clients", actionRead),
	clientsvcpb.ClientsService_Update_FullMethodName: permission("clients", actionUpdate),
	clientsvcpb.ClientsService_Delete_FullMethodName: permission("clients", actionDelete),

	inventorysvcpb.InventoryService_Get_FullMethodName:  permission("inventory", actionRead),
	inventorysvcpb.InventoryService_List_FullMethodName: permission("inventory", actionRead),

	inventorymovementsvcpb.InventoryMovementsService_Create_FullMethodName:     permission("inventorymovements", actionCreate),
	inventorymovementsvcpb.InventoryMovementsService_List_FullMethodName:       permission("inventorymovements", actionRead),
	inventorymovementsvcpb.InventoryMovementsService_GetHistory_FullMethodName: permission("inventorymovements", actionRead),

	partnersvcpb.PartnersService_Create_FullMethodName: permission("partners", actionCreate),
	partnersvcpb.PartnersService_Get_FullMethodName:    permission("partners", actionRead),
	partnersvcpb.PartnersService_List_FullMethodName:   permission("partners", actionRead),
	partnersvcpb.PartnersService_Update_FullMethodName: permission("partners", actionUpdate),
	partnersvcpb.PartnersService_Delete_FullMethodName: permission("partners", actionDelete),

	pricesvcpb.PricesService_Create_FullMethodName:     permission("prices", actionCreate),
	pricesvcpb.PricesService_Get_FullMethodName:        permission("prices", actionRead),
	pricesvcpb.PricesService_GetHistory_FullMethodName: permission("prices", actionRead),
	pricesvcpb.PricesService_Update_FullMethodName:     permission("prices", actionUpdate),
	pricesvcpb.PricesService_Delete_FullMethodName:     permission("prices", actionDelete),

	productsvcpb.ProductsService_Create_FullMethodName: permission("products", actionCreate),
	productsvcpb.ProductsService_Get_FullMethodName:    permission("products", actionRead),
	productsvcpb.ProductsService_List_FullMethodName:   permission("products", actionRead),
	productsvcpb.ProductsService_Update_FullMethodName: permission("products", actionUpdate),
	productsvcpb.ProductsService_Delete_FullMethodName: permission("products", actionDelete),

	// Create/Update/Delete don't exist on this service (TD: seeded via
	// migrations, read-only over the wire) — List is auth-exempt (see
	// auth.New), so Get is the only method that needs an entry here.
	productattributedefinitionsvcpb.ProductAttributeDefinitionsService_Get_FullMethodName: permission("productattributedefinitions", actionRead),

	// List is auth-exempt (see auth.New) — it backs the public catalog.
	productvariantsvcpb.ProductVariantsService_Create_FullMethodName: permission("productvariants", actionCreate),
	productvariantsvcpb.ProductVariantsService_Get_FullMethodName:    permission("productvariants", actionRead),
	productvariantsvcpb.ProductVariantsService_Update_FullMethodName: permission("productvariants", actionUpdate),
	productvariantsvcpb.ProductVariantsService_Delete_FullMethodName: permission("productvariants", actionDelete),

	// Reports are all read-only aggregation endpoints.
	reportsvcpb.ReportsService_GetSalesReport_FullMethodName:     permission("reports", actionRead),
	reportsvcpb.ReportsService_GetSalesByStaff_FullMethodName:    permission("reports", actionRead),
	reportsvcpb.ReportsService_GetSalesByPartner_FullMethodName:  permission("reports", actionRead),
	reportsvcpb.ReportsService_GetPopularProducts_FullMethodName: permission("reports", actionRead),
	reportsvcpb.ReportsService_GetTurnover_FullMethodName:        permission("reports", actionRead),
	reportsvcpb.ReportsService_GetStockLevels_FullMethodName:     permission("reports", actionRead),

	salesvcpb.SalesService_Create_FullMethodName: permission("sales", actionCreate),
	salesvcpb.SalesService_Get_FullMethodName:    permission("sales", actionRead),
	salesvcpb.SalesService_List_FullMethodName:   permission("sales", actionRead),
	// UpdateStatus/Cancel are both just status transitions on an existing
	// sale, so both fall under the single "sales:update" permission rather
	// than minting a separate "sales:cancel".
	salesvcpb.SalesService_UpdateStatus_FullMethodName: permission("sales", actionUpdate),
	salesvcpb.SalesService_Cancel_FullMethodName:       permission("sales", actionUpdate),

	usersvcpb.UsersService_Create_FullMethodName: permission("users", actionCreate),
	usersvcpb.UsersService_Get_FullMethodName:    permission("users", actionRead),
	usersvcpb.UsersService_List_FullMethodName:   permission("users", actionRead),
	usersvcpb.UsersService_Update_FullMethodName: permission("users", actionUpdate),
	usersvcpb.UsersService_Delete_FullMethodName: permission("users", actionDelete),
	// ChangePassword requires knowing the current password (verified in
	// user.Service.ChangePassword), so it is self-authorizing — every
	// authenticated role may call it for its own account.
	usersvcpb.UsersService_ChangePassword_FullMethodName: permissionAuthenticatedOnly,

	warehousesvcpb.WarehousesService_Create_FullMethodName:     permission("warehouses", actionCreate),
	warehousesvcpb.WarehousesService_Get_FullMethodName:        permission("warehouses", actionRead),
	warehousesvcpb.WarehousesService_List_FullMethodName:       permission("warehouses", actionRead),
	warehousesvcpb.WarehousesService_Update_FullMethodName:     permission("warehouses", actionUpdate),
	warehousesvcpb.WarehousesService_Delete_FullMethodName:     permission("warehouses", actionDelete),
	warehousesvcpb.WarehousesService_Deactivate_FullMethodName: permission("warehouses", actionUpdate),

	// Diagnostic endpoint, no business data: any authenticated caller.
	serviceinfopb.ServiceInfoService_Get_FullMethodName: permissionAuthenticatedOnly,
}

// KnownPermissions returns every distinct non-wildcard, non-self-service
// permission string a role can be granted (i.e. every value that can
// legally appear under a role in CRMConfig.RBAC / configs/crm.yaml,
// besides the universal "*" wildcard). Used by appconfig.CRMConfig.Validate
// to catch a typo'd permission string (e.g. "prodcuts:read") at
// config-load time — otherwise it would silently deny every role that
// wasn't supposed to hold that permission anyway, and nobody would notice
// until a caller who genuinely needed it hit PermissionDenied.
func KnownPermissions() map[string]bool {
	known := make(map[string]bool, len(permissions))
	for _, perm := range permissions {
		if perm != permissionAuthenticatedOnly {
			known[perm] = true
		}
	}
	return known
}
