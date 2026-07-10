package appconfig

// CRMConfig holds business-domain configuration that has no go-atlas infra
// counterpart (bootstrap admin, RBAC permission table).
type CRMConfig struct {
	// BootstrapAdmin, when set, is created on first boot if no user with a
	// matching phone/email exists yet — the TD requires the initial admin
	// to come from config, never from the regular Create RPC.
	BootstrapAdmin *BootstrapAdminConfig `yaml:"bootstrapAdmin" default:"-"`

	// RBAC maps a role name (admin/employee/guest) to the permission
	// strings it grants ("*" grants everything). Not enforced by an
	// interceptor yet — see internal/rbac.Table for the lookup helper.
	RBAC map[string][]string `yaml:"rbac" default:"-"`
}

// BootstrapAdminConfig is the plaintext admin account materialized on boot.
type BootstrapAdminConfig struct {
	Name     string `yaml:"name"`
	Phone    string `yaml:"phone"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}
