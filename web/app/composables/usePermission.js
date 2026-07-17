import { ROLE_PERMISSIONS } from '~/config/permissions'

// USER_ROLE_ADMIN -> 'admin' (matches proto-loader's string enum output
// and the role keys used in app/config/permissions.js).
export function normalizeRole(role) {
  if (!role) return null
  return role.replace(/^USER_ROLE_/, '').toLowerCase()
}

/**
 * Single entry point for permission checks (TD §4.3) — never compare
 * `user.role === 'admin'` directly in components/pages, always go through
 * can(). Route guards, menu generation, and <UButton v-if="can(...)"> all
 * use this.
 */
export function usePermission() {
  const { user } = useAuth()
  const role = computed(() => normalizeRole(user.value?.role))

  function can(permission) {
    // No permission key at all means the action doesn't exist for this
    // entity (e.g. Inventory has no create/update/delete on the backend,
    // full stop) — not "admin can do it via the wildcard". A falsy
    // permission is never granted, regardless of role.
    if (!permission || !user.value) return false
    const perms = ROLE_PERMISSIONS[role.value] ?? []
    return perms.includes('*') || perms.includes(permission)
  }

  return { can, role }
}
