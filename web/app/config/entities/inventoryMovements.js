// Single source of truth for the InventoryMovement entity (TD §9.1, §10,
// §12.4) — append-only ledger: Create + List/GetHistory only, no
// Get/Update/Delete on the backend at all, so no update/delete permission
// keys (EntityDataTable hides Actions/click-to-edit automatically, same
// mechanism as Inventory). The type select excludes MOVEMENT_TYPE_SALE
// (internal, created only by SalesService.Create) via ENUMS.MovementType.
export default {
  key: 'inventoryMovements',
  label: 'entities.inventoryMovements.label',
  icon: 'i-lucide-arrow-left-right',
  route: '/inventory-movements',
  group: 'warehouse',
  permissions: {
    read: 'inventorymovements:read',
    create: 'inventorymovements:create',
  },

  list: {
    columns: [
      { key: 'skuId', label: 'fields.sku', component: 'RelationLabel', relation: 'productSkus' },
      { key: 'warehouseId', label: 'fields.warehouse', component: 'RelationLabel', relation: 'warehouses' },
      { key: 'type', label: 'fields.movementType', component: 'EnumLabel' },
      { key: 'quantity', label: 'fields.quantity', component: 'QuantityAmountLabel' },
      { key: 'comment', label: 'fields.comment' },
      { key: 'saleId', label: 'fields.sale', component: 'SaleLink' },
      { key: 'createdAt', label: 'fields.createdAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'types', type: 'multiselect', label: 'fields.movementType', optionsFrom: 'enum:MovementType' },
      { key: 'createdAt', type: 'periodFilter', label: 'fields.createdAt' },
    ],
    // InventoryMovementsListRequest has no Sort message on the backend.
    sort: [],
    defaultSort: null,
  },

  form: {
    fields: [
      // Product -> Variant -> SKU cascade (SkuCascadeSelect), not a flat
      // relation dropdown — with a large catalog, a raw list of every SKU
      // (ProductSKUsService.List has no free-text search) is unusable.
      { key: 'skuId', type: 'skuCascade', label: 'fields.sku', required: true },
      // Scoped to warehouses that actually carry this SKU (Inventory.List
      // filtered by skuId) — a flat `warehouses` relation would offer
      // every warehouse, most of which never stocked it. Stays disabled
      // until skuId is picked (WarehouseStockSelect reads it off `form`
      // directly, same cross-field wiring EntityForm already does for
      // `capByStock` below).
      { key: 'warehouseId', type: 'warehouseStock', label: 'fields.warehouse', required: true },
      { key: 'type', type: 'enum', enum: 'MovementType', label: 'fields.movementType', required: true },
      // capByStock: floors the input at -(current on-hand quantity for the
      // selected skuId/warehouseId pair) once both are picked, so an
      // operator can't type a write-off/adjustment larger than what's
      // actually on the shelf (backend would reject it anyway via
      // Inventory.ApplyMovement, but catching it client-side is a much
      // clearer error than a generic insufficient-stock failure).
      {
        key: 'quantity',
        type: 'number',
        label: 'fields.quantity',
        required: true,
        capByStock: true,
        requiresFields: ['skuId', 'warehouseId'],
        // Wire value is hundredths of a unit (see
        // entities.InventoryMovement.Quantity) — EntityForm shows/accepts
        // the plain scaled-down number the operator types.
        scale: 100,
      },
      { key: 'comment', type: 'text', label: 'fields.comment', maxLength: 1024 },
      { key: 'saleId', type: 'relation', relation: 'sales', label: 'fields.sale' },
    ],
  },
}
