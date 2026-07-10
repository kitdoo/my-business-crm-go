package fx

import (
	"context"
	"errors"
	"log/slog"

	"go.uber.org/fx"

	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
	"github.com/kitdoo/my-business-crm-go/internal/storages/users"
)

// registerBootstrapAdmin creates the config-defined admin account on first
// boot if no user with its phone/email exists yet. The TD requires the
// initial admin to come from config, never from the regular Create RPC.
func registerBootstrapAdmin(lc fx.Lifecycle, cfg *appconfig.Config, storage users.Storage, svc usersvc.Service) {
	admin := getBootstrapAdmin(cfg.CRM)
	if admin == nil {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger := slog.Default().With(slogx.Module("bootstrap:admin"))

			if _, err := storage.GetByLogin(ctx, admin.Email); err == nil {
				return nil
			} else if !errors.Is(err, errs.ErrUserNotFound) {
				return err
			}

			_, err := svc.Create(ctx, &entities.UserCreate{
				Name:     entities.LocalizedString{"sr": admin.Name},
				Phone:    admin.Phone,
				Email:    admin.Email,
				Role:     entities.UserRoleAdmin,
				Password: admin.Password,
			})
			if err != nil {
				if errors.Is(err, errs.ErrUserPhoneConflict) || errors.Is(err, errs.ErrUserEmailConflict) {
					return nil
				}
				return err
			}
			logger.InfoContext(ctx, "bootstrap admin created", slog.String("email", admin.Email))
			return nil
		},
	})
}

// GetBootstrapAdmin is a nil-safe accessor: cfg.CRM is optional, so this
// keeps registerBootstrapAdmin a one-line nil check.
func getBootstrapAdmin(cfg *appconfig.CRMConfig) *appconfig.BootstrapAdminConfig {
	if cfg == nil {
		return nil
	}
	return cfg.BootstrapAdmin
}
