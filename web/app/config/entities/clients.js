// Single source of truth for the Client entity (TD §9.1, §10) — the
// simplest entity in the registry: plain (non-localized) fields, no
// relations, no enum/status. SideMenu, routing, EntityListPage/EntityForm
// all read only this file for Client-specific behavior.
export default {
  key: 'clients',
  label: 'entities.clients.label',
  icon: 'i-lucide-user-round',
  route: '/clients',
  group: 'sales',
  permissions: {
    read: 'clients:read',
    create: 'clients:create',
    update: 'clients:update',
    delete: 'clients:delete',
  },

  list: {
    // Lets useReferenceCacheStore batch RelationLabel lookups into one
    // filter.ids List call instead of one Get per row.
    idsFilterKey: 'ids',
    columns: [
      { key: 'name', label: 'fields.name' },
      { key: 'phone', label: 'fields.phone' },
      { key: 'email', label: 'fields.email' },
      { key: 'address', label: 'fields.address' },
    ],
    filters: [],
    sort: [
      { field: 'FIELD_CREATED_AT', label: 'sort.createdAt' },
      { field: 'FIELD_NAME', label: 'sort.name' },
    ],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_DESC' },
  },

  // Viewing a client preloads their sales history underneath the record
  // fields (EntityViewDrawer's relatedSales — SalesListRequest.clientId).
  view: {
    relatedSales: (record) => ({ clientId: record.id }),
  },

  form: {
    fields: [
      { key: 'name', type: 'text', label: 'fields.name', required: true, maxLength: 255 },
      { key: 'phone', type: 'text', label: 'fields.phone', required: true, maxLength: 32 },
      { key: 'email', type: 'text', label: 'fields.email', required: true, inputType: 'email' },
      { key: 'address', type: 'text', label: 'fields.address', required: true, maxLength: 512 },
      { key: 'note', type: 'text', label: 'fields.note', maxLength: 1024 },
    ],
  },
}
