// Package rbac provides a config-driven role -> permission lookup. The TD
// requires role permissions to be configurable without code changes; this
// table is that lookup. Wiring it into a gRPC interceptor that enforces it
// per-request is a separate cross-cutting task shared by every handler, not
// specific to any one aggregate, and is left for that follow-up.
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
