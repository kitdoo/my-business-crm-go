package entities

import "github.com/kitdoo/my-business-crm-go/internal/errs"

// Validate reports errs.ErrLocalizedStringMissingRequiredLocale if l is
// non-empty but missing requiredLocale, or requiredLocale is present but
// blank. An empty l is valid on its own — some LocalizedString fields
// (e.g. Description) are themselves optional; whether the field is
// required at all is enforced separately (proto buf.validate for
// required fields on Create).
//
// requiredLocale comes from CRMConfig.DefaultLocale (see
// internal/fx.resolveDefaultLocale), not a hardcoded value — callers
// (each *Create/*Update.Validate) are handed it by their owning service,
// which resolves it once from config.
func (l LocalizedString) Validate(requiredLocale string) error {
	if len(l) == 0 {
		return nil
	}
	if l[requiredLocale] == "" {
		return errs.ErrLocalizedStringMissingRequiredLocale
	}
	return nil
}
