// Package rbac provides a config-driven role -> permission lookup. The TD
// requires role permissions to be configurable without code changes; this
// table is that lookup. It is enforced per-request by the gRPC RBAC
// interceptor (internal/transports/grpc/interceptors/rbac), which maps
// every authenticated method to a permission string and checks the
// caller's role against this table.
package rbac

// Table maps a role name to the permission strings it grants. A "*" entry
// grants every permission for that role.
type Table map[string][]string

// Allowed reports whether role grants permission.
func (t Table) Allowed(role, permission string) bool {
	for _, p := range t[role] {
		if p == "*" || p == permission {
			return true
		}
	}
	return false
}
