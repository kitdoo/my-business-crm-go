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

	// Smtp configures the outbound mail server NotificationsService.Send
	// delivers messages through (internal/pkg/mailer) when Resend is
	// absent. Optional: when both are absent, Send returns a "not
	// configured" error instead of failing app boot.
	Smtp *SMTPConfig `yaml:"smtp" default:"-"`

	// Resend configures delivery through the Resend HTTP API
	// (internal/pkg/mailer/resend). Takes precedence over Smtp when
	// both are set — see newMailerService in internal/fx/services.go.
	// Prefer this over Smtp: direct SMTP to providers like Gmail is
	// commonly blackholed from cloud/VPS source IPs on reputation
	// grounds, where an HTTPS API call is not.
	Resend *ResendConfig `yaml:"resend" default:"-"`

	// NotificationClients maps a client name (for audit logging — which
	// frontend sent this) to the static API key it must present via the
	// "x-client-key" gRPC metadata header to call
	// NotificationsService.Send — see internal/transports/grpc/
	// interceptors/clientkey. A caller whose key matches no entry here is
	// rejected; an absent/empty map denies every caller (fail closed),
	// same as an absent/empty RBAC table.
	NotificationClients map[string]redacted.RedactedString `yaml:"notificationClients"`

	// Invoice configures the seller letterhead, commercial terms, and VAT
	// rate used by InvoicesService (internal/services/invoice) to render a
	// Sale as a PDF — see internal/transports/http/handlers/invoice.
	// Optional: when absent, invoice generation fails with an explicit
	// "not configured" error instead of rendering a blank letterhead.
	Invoice *InvoiceConfig `yaml:"invoice" default:"-"`
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
		validation.Field(&c.Smtp, validation.NilOrNotEmpty),
		validation.Field(&c.Resend, validation.NilOrNotEmpty),
		validation.Field(&c.NotificationClients, validation.By(validateNotificationClients)),
		validation.Field(&c.Invoice, validation.NilOrNotEmpty),
	)
}

// validateNotificationClients rejects a client entry with a blank name or
// key — either would silently make that entry unreachable/unmatchable
// rather than failing config load.
func validateNotificationClients(value any) error {
	table, ok := value.(map[string]redacted.RedactedString)
	if !ok {
		return nil
	}
	for name, key := range table {
		if name == "" {
			return fmt.Errorf("notificationClients has an entry with an empty client name")
		}
		if key.Expose() == "" {
			return fmt.Errorf("notificationClients entry %q has an empty key", name)
		}
	}
	return nil
}

// SMTPConfig configures the outbound mail server NotificationsService.Send
// delivers messages through (internal/pkg/mailer) to To.
type SMTPConfig struct {
	// Host is the SMTP server hostname (e.g. "smtp.gmail.com").
	Host string `yaml:"host"`
	// Port is the SMTP server port. 587 (STARTTLS) and 465 (implicit TLS)
	// are both supported — see mailer.Service.
	Port int `yaml:"port" default:"587"`
	// Username authenticates to the SMTP server. Empty disables auth
	// (some internal relays accept unauthenticated mail).
	Username string `yaml:"username"`
	// Password authenticates to the SMTP server alongside Username.
	Password redacted.RedactedString `yaml:"password"`
	// From is the envelope/header From address emails are sent as.
	From string `yaml:"from"`
	// To is the single recipient address every message is delivered to.
	To string `yaml:"to"`
}

// Validate requires Host, Port, From, and To — a partially configured
// SMTP section would otherwise reach mailer.Service with a blank field
// instead of failing config load.
func (s *SMTPConfig) Validate() error {
	return config.ValidateStruct(s,
		validation.Field(&s.Host, validation.Required),
		validation.Field(&s.Port, validation.Required, validation.Min(1), validation.Max(65535)),
		validation.Field(&s.From, validation.Required),
		validation.Field(&s.To, validation.Required),
	)
}

// ResendConfig configures outbound mail delivery through the Resend HTTP
// API (internal/pkg/mailer/resend), to To.
type ResendConfig struct {
	// APIKey authenticates to the Resend API.
	APIKey redacted.RedactedString `yaml:"apiKey"`
	// From is the header From address emails are sent as. Must be on a
	// domain verified with Resend.
	From string `yaml:"from"`
	// To is the single recipient address every message is delivered to.
	To string `yaml:"to"`
}

// Validate requires APIKey, From, and To — a partially configured Resend
// section would otherwise reach resend.Service with a blank field instead
// of failing config load.
func (r *ResendConfig) Validate() error {
	return config.ValidateStruct(r,
		validation.Field(&r.APIKey, validation.Required),
		validation.Field(&r.From, validation.Required),
		validation.Field(&r.To, validation.Required),
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

// InvoiceCompanyConfig is the seller letterhead printed on every generated
// invoice — TOM STUDIO's own business details, never per-Sale data.
type InvoiceCompanyConfig struct {
	Name string `yaml:"name"`
	// City is the "Mesto izdavanja" (place of issue) printed on every
	// invoice — appended after Name in the letterhead and repeated in the
	// document-info box.
	City               string `yaml:"city"`
	Address            string `yaml:"address"`
	Phone              string `yaml:"phone"`
	Email              string `yaml:"email"`
	TaxID              string `yaml:"taxId"`
	RegistrationNumber string `yaml:"registrationNumber"`
	// ActivityCode is the "šifra delatnosti" printed on the letterhead.
	ActivityCode      string `yaml:"activityCode"`
	BankName          string `yaml:"bankName"`
	BankAccount       string `yaml:"bankAccount"`
	ResponsiblePerson string `yaml:"responsiblePerson"`
}

// Validate requires Name, Address, and BankAccount — the minimum a
// letterhead needs to not render visibly broken; every other field is
// optional and simply left blank on the PDF when unset.
func (c *InvoiceCompanyConfig) Validate() error {
	return config.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Address, validation.Required),
		validation.Field(&c.BankAccount, validation.Required),
	)
}

// InvoiceTermsConfig is the commercial-terms boilerplate printed on every
// invoice (payment/delivery conditions) — free text, not computed from the
// Sale.
type InvoiceTermsConfig struct {
	AdvancePercentage int `yaml:"advancePercentage" default:"100"`
	// DeliveryLeadTime is free text, e.g. "5-6 radnih dana od datuma
	// prihvatanja ponude i uplate avansa."
	DeliveryLeadTime  string `yaml:"deliveryLeadTime"`
	OfferValidityDays int    `yaml:"offerValidityDays" default:"5"`
}

// InvoiceVATConfig is the single system-wide VAT rate InvoicesService
// applies when a Generate request has includeVat=true — see
// entities.InvoiceGenerate.IncludeVAT. There is no per-invoice custom rate.
type InvoiceVATConfig struct {
	Percentage int `yaml:"percentage" default:"20"`
}

// Validate rejects a nonsensical percentage — a partially configured or
// mistyped rate would otherwise silently produce wrong VAT amounts on
// every invoice that requests it.
func (v *InvoiceVATConfig) Validate() error {
	return config.ValidateStruct(v,
		validation.Field(&v.Percentage, validation.Min(0), validation.Max(100)),
	)
}

// InvoiceConfig configures InvoicesService's PDF rendering — see
// CRMConfig.Invoice.
type InvoiceConfig struct {
	Company InvoiceCompanyConfig `yaml:"company"`
	// NumberPrefix prepends the Sale's own Number to build the printed
	// Offer number, e.g. prefix "POKU" + sale number 4 + issue date ->
	// "POKU_004-13-07-26".
	NumberPrefix string `yaml:"numberPrefix"`
	// ReceiptNumberPrefix is NumberPrefix's counterpart for the Receipt
	// (entities.InvoiceDocumentTypeReceipt). Falls back to NumberPrefix
	// when unset, since NumberPrefix's own default ("POKU" = Serbian
	// "Ponuda Kupca") wouldn't otherwise fit a Receipt's number.
	ReceiptNumberPrefix string             `yaml:"receiptNumberPrefix"`
	Terms               InvoiceTermsConfig `yaml:"terms"`
	VAT                 InvoiceVATConfig   `yaml:"vat"`
}

// Validate cascades into Company/VAT — Terms has no invalid shape (both its
// fields are plain, unconstrained numbers/text).
func (i *InvoiceConfig) Validate() error {
	return config.ValidateStruct(i,
		validation.Field(&i.Company, validation.Required),
		validation.Field(&i.VAT),
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
