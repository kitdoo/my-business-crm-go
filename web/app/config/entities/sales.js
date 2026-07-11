// Single source of truth for the Sale entity's list rendering + SideMenu
// registration (TD §9.1, §10, §12.3). Not a generic CRUD entity — no
// Update/Delete on the backend at all (items immutable, status moves only
// through UpdateStatus/Cancel), so no update/delete permission keys
// (EntityDataTable hides Actions automatically) and no `form` — creation
// is a bespoke 2-step wizard (pages/sales/new.vue), editing is a bespoke
// detail page (pages/sales/[id].vue) with status-transition + cancel
// actions, not <EntityForm>.
export default {
  key: 'sales',
  label: 'entities.sales.label',
  icon: 'i-lucide-shopping-cart',
  route: '/sales',
  permissions: {
    read: 'sales:read',
    create: 'sales:create',
  },

  list: {
    columns: [
      { key: 'clientId', label: 'fields.client', component: 'RelationLabel', relation: 'clients' },
      { key: 'warehouseId', label: 'fields.warehouse', component: 'RelationLabel', relation: 'warehouses' },
      { key: 'totalAmount', label: 'fields.totalAmount', component: 'MoneyAmountLabel' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'sale' },
      { key: 'createdAt', label: 'fields.createdAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'statuses', type: 'multiselect', label: 'fields.status', optionsFrom: 'enum:SaleStatus' },
      { key: 'createdAt', type: 'periodFilter', label: 'fields.createdAt' },
    ],
    sort: [{ field: 'FIELD_CREATED_AT', label: 'sort.createdAt' }],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_DESC' },
  },

  form: { fields: [] },
}
