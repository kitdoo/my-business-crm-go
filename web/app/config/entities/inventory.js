// Single source of truth for the Inventory entity (TD §9.1, §10) —
// read-only: no create/update/delete permission keys at all (stock only
// changes through InventoryMovements), so EntityListPage/EntityDataTable
// hide the Create button and Actions column automatically the same way
// they hide anything the caller's role can't do — no special-casing
// needed for "this entity happens to be read-only for everyone".
export default {
  key: 'inventory',
  label: 'entities.inventory.label',
  icon: 'i-lucide-package-search',
  route: '/inventory',
  group: 'warehouse',
  permissions: {
    read: 'inventory:read',
  },

  list: {
    columns: [
      { key: 'productId', label: 'fields.product', component: 'RelationLabel', relation: 'products' },
      { key: 'warehouseId', label: 'fields.warehouse', component: 'RelationLabel', relation: 'warehouses' },
      { key: 'quantity', label: 'fields.quantity' },
      { key: 'updatedAt', label: 'fields.updatedAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'minQuantity', type: 'number', label: 'fields.minQuantity' },
      { key: 'maxQuantity', type: 'number', label: 'fields.maxQuantity' },
    ],
    // InventoryListRequest has no Sort message on the backend at all.
    sort: [],
    defaultSort: null,
  },

  form: { fields: [] },
}
