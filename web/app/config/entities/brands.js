// Single source of truth for the Brand entity (TD §9.1): how to list it,
// edit it, and which permissions gate each action. SideMenu, routing,
// EntityListPage/EntityForm all read only this file for Brand-specific
// behavior — adding a field here is the whole change.
export default {
  key: 'brands',
  label: 'entities.brands.label',
  icon: 'i-lucide-tag',
  route: '/brands',
  group: 'catalog',
  permissions: {
    read: 'brands:read',
    create: 'brands:create',
    update: 'brands:update',
    delete: 'brands:delete',
  },

  list: {
    columns: [
      { key: 'name', label: 'fields.name', component: 'LocalizedText' },
      { key: 'description', label: 'fields.description', component: 'LocalizedText' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'brand' },
      { key: 'createdAt', label: 'fields.createdAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'statuses', type: 'multiselect', label: 'fields.status', optionsFrom: 'enum:BrandStatus' },
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
      { key: 'status', type: 'enum', enum: 'BrandStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
