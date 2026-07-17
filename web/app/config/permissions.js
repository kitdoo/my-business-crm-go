// The frontend's own copy of role -> permissions (TD §4.3). The backend
// does not expose this table over RPC, so this file is a second,
// independent source of truth — keep it in sync with configs/crm.yaml /
// internal/transports/grpc/interceptors/rbac/permissions.go by hand at
// each release (see DoD checklist). UI-hiding here is UX only; the real
// enforcement is server-side RBAC — every mutation still goes through
// useApiErrorHandler for PermissionDenied (TD §4.2/§4.4).
// Every resource read permission that exists on the backend (see
// permissions.go) — both guest and employee can view the whole catalog;
// only the action set and (for employee) the users resource differ.
const ALL_READ = [
  'brands:read',
  'categories:read',
  'clients:read',
  'inventory:read',
  'inventorymovements:read',
  'partners:read',
  'prices:read',
  'products:read',
  'productattributedefinitions:read',
  'productvariants:read',
  'productskus:read',
  'reports:read',
  'sales:read',
  'users:read',
  'warehouses:read',
]

export const ROLE_PERMISSIONS = {
  admin: ['*'],
  // Full read access except users, plus the only mutation employees get:
  // creating/editing sales. Note SideMenu additionally restricts which of
  // these read-visible entities show up as menu items (config/navigation.js
  // ROLE_MENU_KEYS) — the extra reads here back relation lookups (brand
  // name on a product card, etc), not menu entries.
  employee: [...ALL_READ.filter((p) => p !== 'users:read'), 'sales:create', 'sales:update'],
  // Full read access, no mutations at all — every create/update/delete
  // control in the UI is gated through can() and stays hidden.
  guest: ALL_READ,
}
