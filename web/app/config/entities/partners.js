// Single source of truth for the Partner entity (TD §9.1, §10): plain
// fields like Client, plus a 0-100 commission percentage and a status
// enum. SideMenu, routing, EntityListPage/EntityForm all read only this
// file for Partner-specific behavior.
export default {
  key: 'partners',
  label: 'entities.partners.label',
  icon: 'i-lucide-handshake',
  route: '/partners',
  group: 'sales',
  permissions: {
    read: 'partners:read',
    create: 'partners:create',
    update: 'partners:update',
    delete: 'partners:delete',
  },

  list: {
    columns: [
      { key: 'name', label: 'fields.name' },
      { key: 'phone', label: 'fields.phone' },
      { key: 'email', label: 'fields.email' },
      { key: 'commissionPercentage', label: 'fields.commissionPercentage' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'partner' },
      { key: 'createdAt', label: 'fields.createdAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'statuses', type: 'multiselect', label: 'fields.status', optionsFrom: 'enum:PartnerStatus' },
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
      { key: 'name', type: 'text', label: 'fields.name', required: true, maxLength: 255 },
      { key: 'phone', type: 'text', label: 'fields.phone', required: true, maxLength: 32 },
      { key: 'email', type: 'text', label: 'fields.email', required: true, inputType: 'email' },
      { key: 'address', type: 'text', label: 'fields.address', maxLength: 500 },
      { key: 'commissionPercentage', type: 'number', label: 'fields.commissionPercentage', required: true, min: 0, max: 100 },
      { key: 'note', type: 'text', label: 'fields.note', maxLength: 1024 },
      { key: 'status', type: 'enum', enum: 'PartnerStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
