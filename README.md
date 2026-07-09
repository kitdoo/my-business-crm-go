# my-business-crm-go

A minimal Go service skeleton built on [go-atlas](https://github.com/altessa-s/go-atlas):
Cobra CLI, Uber Fx dependency injection, MongoDB + Redis infrastructure, and
gRPC + HTTP servers with the standard probe/metrics endpoints.

## Quick start

```bash
# Inspect available commands
go run ./cmd/my-business-crm --help

# Start the server (dev config)
go run ./cmd/my-business-crm server run --config ./configs/

# Print version
go run ./cmd/my-business-crm version --full --output json
```

## Project layout

```
my-business-crm-go/
├── cmd/my-business-crm/     # Binary entry + Cobra command tree
├── configs/                 # YAML config templates (all opt-in, commented out by default)
├── internal/
│   ├── fx/                  # Uber Fx modules (infrastructure / transports)
│   └── pkg/appconfig/       # Config struct + loader
├── Dockerfile Makefile      # Build automation
└── .golangci.yml .optgen.yaml
```

## Build

```bash
make build       # build the binary into build/bin/<version>/<os>-<arch>/my-business-crm
make lint        # run gci + golangci-lint
make test        # go test ./... -race -count=1
```

## Configuration

Configuration is loaded from one or more YAML files under `configs/`.
Every key may be overridden by a `CRM_`-prefixed environment variable with
`__` denoting nested keys, e.g.:

```bash
CRM_MONGODB__URI=mongodb://localhost:27017 \
CRM_LOGGER__LEVEL=debug \
go run ./cmd/my-business-crm server run --config ./configs/
```

The full config schema lives in `internal/pkg/appconfig/config.go`.
