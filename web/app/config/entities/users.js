// Single source of truth for the User entity (TD §9.1, §10, §12.5).
// admin-only in practice (config/permissions.js doesn't grant any
// users:* permission to other roles, so SideMenu/EntityListPage hide it
// for everyone else automatically — no special-casing needed here).
// password is createOnly: set once at Create, changed only through the
// separate self-service ChangePassword flow (TopBar profile menu), never
// through this form's Update. status is editOnly: UserCreateRequest has
// no status field on the backend.
export default {
  key: 'users',
  label: 'entities.users.label',
  icon: 'i-lucide-user-cog',
  route: '/users',
  permissions: {
    read: 'users:read',
    create: 'users:create',
    update: 'users:update',
    delete: 'users:delete',
  },

  list: {
    columns: [
      { key: 'name', label: 'fields.name', component: 'LocalizedText' },
      { key: 'lastName', label: 'fields.lastName', component: 'LocalizedText' },
      { key: 'phone', label: 'fields.phone' },
      { key: 'email', label: 'fields.email' },
      { key: 'role', label: 'fields.role', component: 'EnumLabel' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'user' },
      { key: 'createdAt', label: 'fields.createdAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'statuses', type: 'multiselect', label: 'fields.status', optionsFrom: 'enum:UserStatus' },
      { key: 'roles', type: 'multiselect', label: 'fields.role', optionsFrom: 'enum:UserRole' },
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
      { key: 'lastName', type: 'localizedString', label: 'fields.lastName' },
      { key: 'phone', type: 'text', label: 'fields.phone', required: true, maxLength: 32 },
      { key: 'email', type: 'text', label: 'fields.email', required: true, inputType: 'email' },
      { key: 'description', type: 'localizedString', label: 'fields.description' },
      { key: 'role', type: 'enum', enum: 'UserRole', label: 'fields.role', required: true },
      {
        key: 'password',
        type: 'text',
        label: 'fields.password',
        required: true,
        createOnly: true,
        inputType: 'password',
        maxLength: 128,
      },
      { key: 'status', type: 'enum', enum: 'UserStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
