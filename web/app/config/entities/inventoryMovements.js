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
      { key: 'productId', label: 'fields.product', component: 'RelationLabel', relation: 'products' },
      { key: 'warehouseId', label: 'fields.warehouse', component: 'RelationLabel', relation: 'warehouses' },
      { key: 'type', label: 'fields.movementType', component: 'EnumLabel' },
      { key: 'quantity', label: 'fields.quantity' },
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
      { key: 'productId', type: 'relation', relation: 'products', label: 'fields.product', required: true },
      { key: 'warehouseId', type: 'relation', relation: 'warehouses', label: 'fields.warehouse', required: true },
      { key: 'type', type: 'enum', enum: 'MovementType', label: 'fields.movementType', required: true },
      { key: 'quantity', type: 'number', label: 'fields.quantity', required: true },
      { key: 'comment', type: 'text', label: 'fields.comment', maxLength: 1024 },
      { key: 'saleId', type: 'relation', relation: 'sales', label: 'fields.sale' },
    ],
  },
}
