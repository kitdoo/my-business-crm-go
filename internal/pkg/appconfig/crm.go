package appconfig

import (
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/altessa-s/go-atlas/config"
	"github.com/altessa-s/go-atlas/core/types/redacted"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	grpcrbac "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/interceptors/rbac"
)

// knownRBACRoles are the only role names the RBAC interceptor ever checks
// a caller's role against (see entities.UserRole.String()) — any other
// key in CRMConfig.RBAC can never match a real caller and is almost
// certainly a typo.
var knownRBACRoles = map[string]bool{
	entities.UserRoleAdmin.String():    true,
	entities.UserRoleEmployee.String(): true,
	entities.UserRoleGuest.String():    true,
}

// validateRBAC rejects an unknown role key or an unknown, non-wildcard
// permission string in CRMConfig.RBAC — see grpcrbac.KnownPermissions.
func validateRBAC(value any) error {
	table, ok := value.(map[string][]string)
	if !ok {
		return nil
	}

	knownPermissions := grpcrbac.KnownPermissions()
	for role, perms := range table {
		if !knownRBACRoles[role] {
			return fmt.Errorf("unknown role %q (expected one of admin/employee/guest)", role)
		}
		for _, perm := range perms {
			if perm == "*" {
				continue
			}
			if !knownPermissions[perm] {
				return fmt.Errorf("role %q grants unknown permission %q", role, perm)
			}
		}
	}
	return nil
}

// CRMConfig holds business-domain configuration that has no go-atlas infra
// counterpart (bootstrap admin, RBAC permission table).
type CRMConfig struct {
	// BootstrapAdmin, when set, is created on first boot if no user with a
	// matching phone/email exists yet — the TD requires the initial admin
	// to come from config, never from the regular Create RPC.
	BootstrapAdmin *BootstrapAdminConfig `yaml:"bootstrapAdmin" default:"-"`

	// RBAC maps a role name (admin/employee/guest) to the permission
	// strings it grants ("*" grants everything). Enforced per-request by
	// the gRPC RBAC interceptor
	// (internal/transports/grpc/interceptors/rbac); a role with no
	// matching entry here is denied every permission-gated method.
	RBAC map[string][]string `yaml:"rbac"`

	// Currency is the single ISO 4217 code the whole system prices in (e.g.
	// "RSD"). Per the TD, it is set here, not chosen per-price.
	Currency string `yaml:"currency" default:"RSD"`

	// DefaultLocale is the locale code every non-empty LocalizedString
	// value (Brand/Category/Warehouse/User/Product name/description
	// fields) must include — see entities.LocalizedString.Validate. Sort
	// keys and the frontend's display fallback both rely on this locale
	// always being present.
	DefaultLocale string `yaml:"defaultLocale" default:"sr"`

	// Images configures the product-image upload endpoint (plain HTTP, not
	// gRPC — see internal/transports/http/handlers/image).
	Images *ImagesConfig `yaml:"images" default:"-"`

	// Auth configures session-token signing (internal/services/user).
	// Required for the service to boot — see AuthConfig.SigningKey.
	Auth *AuthConfig `yaml:"auth" default:"-"`
}

// Validate enforces that, when the crm section is present, auth is too
// (user.Service cannot function without a signing key — see
// AuthConfig.Validate) and that any other configured sub-section is
// internally valid.
func (c *CRMConfig) Validate() error {
	return config.ValidateStruct(c,
		validation.Field(&c.BootstrapAdmin, validation.NilOrNotEmpty),
		validation.Field(&c.Images, validation.NilOrNotEmpty),
		validation.Field(&c.Auth, validation.Required),
		validation.Field(&c.RBAC, validation.By(validateRBAC)),
	)
}

// AuthConfig configures the self-contained (stateless) session tokens
// issued by user.Service.Login. Tokens are never persisted anywhere — they
// carry the user id and an expiry, HMAC-SHA256-signed with SigningKey, and
// are verified purely by recomputing that signature. See
// internal/services/user/user/token.go.
type AuthConfig struct {
	// SigningKey is the HMAC-SHA256 secret used to sign and verify session
	// tokens. Required: user.Service construction fails without it. Every
	// node that must accept these tokens (every gRPC/HTTP replica) needs
	// the exact same key, and rotating it invalidates every token issued
	// under the old key (there is no key-id/rotation scheme here).
	SigningKey redacted.RedactedString `yaml:"signingKey"`

	// TokenTTL is how long a session token stays valid after Login, before
	// the caller must log in again.
	TokenTTL time.Duration `yaml:"tokenTTL" default:"72h"`
}

// Validate requires SigningKey — without it user.Service construction
// fails anyway (see internal/fx.newUserService), but failing here at
// config-load time gives a much clearer error than a deep DI wiring
// failure.
func (a *AuthConfig) Validate() error {
	return config.ValidateStruct(a,
		validation.Field(&a.SigningKey, validation.Required),
	)
}

// ImagesConfig configures where uploaded product images are stored on disk
// and the limits enforced on upload.
type ImagesConfig struct {
	// Dir is the directory files are saved under, one file per id (no
	// extension — MIME is sniffed when serving). Defaults to
	// "<var-dir>/images" when unset; that path is resolved at runtime, so
	// it cannot be expressed as a static `default` tag like the other
	// fields here.
	Dir string `yaml:"dir"`
	// MaxSizeBytes caps the accepted upload size.
	MaxSizeBytes int64 `yaml:"maxSizeBytes" default:"10485760"`
}

// Validate rejects a negative MaxSizeBytes (an explicit 0 never reaches
// here — the default tag replaces it before validation runs, same as any
// other zero-value defaulted field in this config).
func (i *ImagesConfig) Validate() error {
	return config.ValidateStruct(i,
		validation.Field(&i.MaxSizeBytes, validation.Min(int64(1))),
	)
}

// BootstrapAdminConfig is the plaintext admin account materialized on boot.
type BootstrapAdminConfig struct {
	Name     string `yaml:"name"`
	Phone    string `yaml:"phone"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

// Validate requires every field once bootstrapAdmin is present: a partial
// entry (e.g. missing password) would otherwise reach
// user.Service.Create with a blank field instead of failing config load —
// see internal/fx.registerBootstrapAdmin.
func (a *BootstrapAdminConfig) Validate() error {
	return config.ValidateStruct(a,
		validation.Field(&a.Name, validation.Required),
		validation.Field(&a.Phone, validation.Required),
		validation.Field(&a.Email, validation.Required),
		validation.Field(&a.Password, validation.Required),
	)
}
