package appconfig

import (
	"fmt"

	"github.com/go-ozzo/ozzo-validation/v4"

	"github.com/altessa-s/go-atlas/config"
	"github.com/altessa-s/go-atlas/config/loader"
	"github.com/altessa-s/go-atlas/config/loader/backend/yaml3"
	"github.com/altessa-s/go-atlas/core/runtime/appinfo"
)

// Config is the top-level configuration of the service. Only Logger, Node,
// MongoDB, and Grpc are required at boot — every other subsystem is
// optional and is silently skipped when its section is absent.
type Config struct {
	// Logger configuration for the service.
	Logger *config.Logger `yaml:"logger" default:"-"`

	// Node-specific configuration (node id, regions).
	Node *config.Node `yaml:"node" default:"-"`

	// MongoDB is the primary persistence backend.
	MongoDB *config.Mongodb `yaml:"mongodb" default:"-"`

	// Grpc is the gRPC server configuration.
	Grpc *config.Grpc `yaml:"grpc"`

	// Http is the HTTP server configuration (probes, metrics, public API).
	Http *config.Http `yaml:"http" default:"-"`

	// Redis is the optional cache / shared-state backend.
	Redis *config.Redis `yaml:"redis" default:"-"`

	// TlsProvider holds the TLS certificate provider configuration.
	TlsProvider *config.TlsProvider `yaml:"tlsProvider" default:"-"`

	// Limiter is the optional shared rate-limiter configuration.
	Limiter *config.TokenBucketLimiter `yaml:"limiter" default:"-"`

	// Idempotency is the optional shared idempotency-keeper configuration.
	Idempotency *config.Idempotency `yaml:"idempotency" default:"-"`

	// Observability bundles metrics + tracing collectors.
	Observability *config.Observability `yaml:"observability" default:"-"`

	// CRM holds business-domain configuration (bootstrap admin, RBAC).
	CRM *CRMConfig `yaml:"crm" default:"-"`
}

// Load reads YAML configuration from filePath (file or directory) and
// merges CRM_-prefixed environment overrides on top.
func Load(filePath string) (*Config, error) {
	opts := []loader.Option{
		loader.WithEnvPrefix(appinfo.EnvPrefix),
	}

	if filePath != "" {
		opts = append(opts, loader.WithPath(filePath))
	}

	cfg := loader.New(&yaml3.Backend{}, opts...)
	if _, err := cfg.Load((*Config)(nil)); err != nil {
		return nil, err
	}

	conf, ok := cfg.Config().(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected config type: %T", cfg.Config())
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return conf, nil
}

// Validate enforces presence of the required sections.
func (c *Config) Validate() error {
	return config.ValidateStruct(c,
		validation.Field(&c.Logger, validation.NilOrNotEmpty),
		validation.Field(&c.Node, validation.Required),
		validation.Field(&c.MongoDB, validation.Required),
		validation.Field(&c.Grpc, validation.Required),
		validation.Field(&c.Http, validation.NilOrNotEmpty),
		validation.Field(&c.Redis, validation.NilOrNotEmpty),
		validation.Field(&c.Limiter, validation.NilOrNotEmpty),
		validation.Field(&c.Idempotency, validation.NilOrNotEmpty),
		validation.Field(&c.Observability, validation.NilOrNotEmpty),
		validation.Field(&c.CRM, validation.NilOrNotEmpty),
	)
}

// IsTlsProviderConfigured reports whether the TLS provider section is present.
func (c *Config) IsTlsProviderConfigured() bool { return c.TlsProvider != nil }
