// The frontend's own copy of role -> permissions (TD §4.3). The backend
// does not expose this table over RPC, so this file is a second,
// independent source of truth — keep it in sync with configs/crm.yaml /
// internal/transports/grpc/interceptors/rbac/permissions.go by hand at
// each release (see DoD checklist). UI-hiding here is UX only; the real
// enforcement is server-side RBAC — every mutation still goes through
// useApiErrorHandler for PermissionDenied (TD §4.2/§4.4).
export const ROLE_PERMISSIONS = {
  admin: ['*'],
  employee: [
    'products:read', 'products:create',
    'categories:read', 'categories:create',
    'brands:read', 'brands:create',
    'sales:create',
  ],
  guest: [],
}
