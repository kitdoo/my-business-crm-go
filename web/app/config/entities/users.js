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
  group: 'users',
  permissions: {
    read: 'users:read',
    create: 'users:create',
    update: 'users:update',
    delete: 'users:delete',
  },

  list: {
    // Lets useReferenceCacheStore batch RelationLabel lookups into one
    // filter.ids List call instead of one Get per row.
    idsFilterKey: 'ids',
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
    // Admin accounts can never be deleted (backend refuses it outright,
    // see internal/errs.ErrUserAdminProtected) — hide the button rather
    // than let the operator hit a confirm dialog that always fails.
    deleteGuard: (record) => record.role !== 'USER_ROLE_ADMIN',
    fields: [
      { key: 'name', type: 'localizedString', label: 'fields.name', required: true },
      { key: 'lastName', type: 'localizedString', label: 'fields.lastName' },
      { key: 'phone', type: 'text', label: 'fields.phone', required: true, maxLength: 32 },
      { key: 'email', type: 'text', label: 'fields.email', required: true, inputType: 'email' },
      { key: 'description', type: 'localizedString', label: 'fields.description' },
      // USER_ROLE_ADMIN excluded: the only admin account comes from
      // CRMConfig.BootstrapAdmin on first boot, the backend rejects
      // Create/Update with role admin outright (see
      // errs.ErrUserAdminCreateForbidden) — no point offering it here.
      { key: 'role', type: 'enum', enum: 'UserRole', label: 'fields.role', required: true, excludeOptions: ['USER_ROLE_ADMIN'] },
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
