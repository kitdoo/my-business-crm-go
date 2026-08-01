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
      { key: 'skuId', label: 'fields.sku', component: 'RelationLabel', relation: 'productSkus' },
      { key: 'warehouseId', label: 'fields.warehouse', component: 'RelationLabel', relation: 'warehouses' },
      { key: 'quantity', label: 'fields.quantity', component: 'QuantityAmountLabel' },
      { key: 'updatedAt', label: 'fields.updatedAt', component: 'DateLabel' },
    ],
    filters: [
      {
        key: 'quantity',
        type: 'numberRange',
        label: 'fields.quantity',
        minKey: 'minQuantity',
        maxKey: 'maxQuantity',
        scale: 100,
      },
    ],
    // Quantity is the only sortable field — Inventory carries no product
    // name to sort by (would need a join against Products).
    sort: [{ field: 'FIELD_QUANTITY', label: 'sort.quantity' }],
    defaultSort: null,
  },

  form: { fields: [] },
}
