// Single source of truth for the Category entity (TD §9.1, §10) —
// categories are peers, not a hierarchy at all; there is no parent
// category. SideMenu, routing, EntityListPage/EntityForm all read only
// this file for Category-specific behavior.
export default {
  key: 'categories',
  label: 'entities.categories.label',
  icon: 'i-lucide-shapes',
  route: '/categories',
  group: 'catalog',
  permissions: {
    read: 'categories:read',
    create: 'categories:create',
    update: 'categories:update',
    delete: 'categories:delete',
  },

  list: {
    columns: [
      { key: 'name', label: 'fields.name', component: 'LocalizedText' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'category' },
    ],
    filters: [{ key: 'statuses', type: 'select', label: 'fields.status', optionsFrom: 'enum:CategoryStatus' }],
    defaultFilter: { statuses: 'CATEGORY_STATUS_ACTIVE' },
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
      { key: 'status', type: 'enum', enum: 'CategoryStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
