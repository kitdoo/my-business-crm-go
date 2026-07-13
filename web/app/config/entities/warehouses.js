// Single source of truth for the Warehouse entity (TD §9.1, §10, §12.2).
// status is intentionally NOT a form field — WarehouseUpdateRequest on the
// backend doesn't carry a status field at all; the only way to change it
// is the dedicated Deactivate RPC, wired here as a declarative
// form.actions entry (EntityForm renders it generically, no bespoke Vue
// file needed).
export default {
  key: 'warehouses',
  label: 'entities.warehouses.label',
  icon: 'i-lucide-warehouse',
  route: '/warehouses',
  group: 'warehouse',
  permissions: {
    read: 'warehouses:read',
    create: 'warehouses:create',
    update: 'warehouses:update',
    delete: 'warehouses:delete',
  },

  list: {
    columns: [
      { key: 'name', label: 'fields.name', component: 'LocalizedText' },
      { key: 'address', label: 'fields.address' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'warehouse' },
      { key: 'createdAt', label: 'fields.createdAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'statuses', type: 'multiselect', label: 'fields.status', optionsFrom: 'enum:WarehouseStatus' },
      { key: 'createdAt', type: 'periodFilter', label: 'fields.createdAt' },
    ],
    sort: [
      { field: 'FIELD_CREATED_AT', label: 'sort.createdAt' },
      { field: 'FIELD_NAME', label: 'sort.name' },
    ],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_DESC' },
  },

  form: {
    fields: [
      { key: 'name', type: 'localizedString', label: 'fields.name', required: true },
      { key: 'description', type: 'localizedString', label: 'fields.description' },
      { key: 'address', type: 'text', label: 'fields.address', required: true, maxLength: 512 },
    ],
    actions: [
      {
        key: 'deactivate',
        label: 'entities.warehouses.deactivate',
        endpoint: '/api/warehouses/deactivate',
        method: 'POST',
        permission: 'update',
        confirmTitle: 'entities.warehouses.deactivateConfirmTitle',
        confirmBody: 'entities.warehouses.deactivateConfirmBody',
        visibleWhen: (record) => record.status === 'WAREHOUSE_STATUS_ACTIVE',
      },
    ],
  },
}
