# my-business-crm-go

A production-oriented CRM backend for small-business retail operations —
built in Go, exposed over gRPC/HTTP, backed by MongoDB and Redis, and paired
with two Nuxt 3 frontends (an admin panel and a public storefront site).

## Purpose

Manages the core of a retail business: product catalog (brands, categories,
attributes, variants, SKUs, pricing), inventory across warehouses with a
full movement ledger, sales, clients/partners, and outbound notifications
(email). It exposes a typed gRPC API (with generated Go and JS/TS clients)
consumed by an internal admin web app and a public-facing storefront.

## Architecture

```
                 ┌──────────────┐        ┌───────────────────┐
                 │   web/       │        │   web-public/      │
                 │ Nuxt 3 admin │        │ Nuxt 3 public site  │
                 └──────┬───────┘        └─────────┬──────────┘
                        │ gRPC-Web / generated JS clients      │
                        ▼                                      ▼
                 ┌─────────────────────────────────────────────┐
                 │              my-business-crm (Go)            │
                 │  gRPC :7779           HTTP :9080 (images)     │
                 │  ── interceptors: auth · rbac · client-key    │
                 │  ── services: brand, category, product,       │
                 │     variant, sku, price, inventory,            │
                 │     inventory movement, warehouse, sale,       │
                 │     client, partner, user, notification,       │
                 │     report                                     │
                 └───────────────┬─────────────┬─────────────────┘
                                 ▼             ▼
                          ┌──────────┐   ┌──────────┐
                          │ MongoDB  │   │  Redis   │
                          └──────────┘   └──────────┘
```

Built on [go-atlas](https://github.com/altessa-s/go-atlas) for the
application skeleton: Cobra CLI, Uber Fx dependency injection, config
loading/validation, and standard probe/metrics endpoints.

## Key technical features

- **Protobuf-first API** — every domain type and RPC is defined in
  `proto/` (validated with [buf](https://buf.build) and
  [protovalidate](https://github.com/bufbuild/protovalidate)), with Go and
  JS/TS bindings generated for the backend and both frontends.
- **RBAC enforced per-RPC** — a gRPC interceptor checks the caller's role
  (admin / employee / guest) against a permission table defined entirely in
  config, with startup-time validation against the set of known
  permissions (`internal/transports/grpc/interceptors/rbac`).
- **Stateless session auth** — HMAC-signed, self-contained session tokens
  (`internal/services/user`), no server-side session store.
- **Client-key gated notifications** — the notification service only
  accepts calls carrying a pre-shared per-frontend key, so email sending
  can't be triggered by an arbitrary caller.
- **Pluggable mail delivery** — SMTP or the [Resend](https://resend.com)
  HTTP API, selected by config (`internal/pkg/mailer`).
- **Full inventory movement ledger** — every stock change is recorded as an
  auditable movement, not just a mutated quantity.
- **Config-driven bootstrap** — the first admin user, RBAC table, currency,
  and default locale are all declared in YAML, not created through the API.
- **Dockerized, multi-service deployment** — backend, admin web, and public
  site each ship as independent images; `docker-compose.yml` wires them up
  with MongoDB/Redis and log rotation for a long-running VM.

## Quick start

```bash
# Inspect available commands
go run ./cmd/my-business-crm --help

# Start the server (dev config)
go run ./cmd/my-business-crm server run --config ./configs/

# Print version
go run ./cmd/my-business-crm version --full --output json
```

Frontends (each is a standalone Nuxt 3 app):

```bash
cd web && npm install && npm run dev          # admin panel  (:3000)
cd web-public && npm install && npm run dev   # public site  (:3001)
```

Full local stack (backend + MongoDB + Redis, and optionally both
frontends) via `docker-compose.yml`.

## Project layout

```
my-business-crm-go/
├── cmd/my-business-crm/       # Binary entry + Cobra command tree
├── configs/                   # YAML config templates (opt-in, commented out by default)
├── internal/
│   ├── entities/               # Domain types
│   ├── services/               # Business logic, one package per domain (sale, inventory, ...)
│   ├── storages/                # MongoDB persistence, one package per domain
│   ├── transports/
│   │   ├── grpc/                # gRPC handlers + interceptors (auth, rbac, client-key)
│   │   └── http/                 # Image upload/serve endpoint
│   ├── rbac/                    # Permission model
│   ├── pkg/appconfig/            # Config struct + loader + validation
│   └── pkg/mailer/                # SMTP / Resend mail delivery
├── proto/                     # Protobuf definitions (buf-managed) + generated Go/JS
├── web/                       # Nuxt 3 admin panel
├── web-public/                # Nuxt 3 public storefront site
├── docs/                      # Technical design docs (backend, frontend, proto/service standards)
├── docker-compose.yml Dockerfile Makefile
└── .golangci.yml .optgen.yaml
```

## Build & quality tooling

```bash
make build            # build the binary into build/bin/<version>/<os>-<arch>/my-business-crm
make lint             # gci + golangci-lint
make test             # go test ./... -race -count=1
make coverage         # test coverage report (HTML)
make security-scan    # gosec + govulncheck
make proto            # regenerate Go + JS protobuf bindings from proto/
```

Docker images for all three components (backend, admin web, public site)
are built and published via `make build-docker-images`.

## Configuration

Configuration is loaded from one or more YAML files under `configs/`.
Every key may be overridden by a `CRM_`-prefixed environment variable with
`__` denoting nested keys, e.g.:

```bash
CRM_MONGODB__URI=mongodb://localhost:27017 \
CRM_LOGGER__LEVEL=debug \
go run ./cmd/my-business-crm server run --config ./configs/
```

The full config schema lives in `internal/pkg/appconfig/config.go` and
`internal/pkg/appconfig/crm.go` (business-domain config: bootstrap admin,
RBAC table, currency, locale, mail delivery, notification client keys).

## Deployment

`docker-compose.yml` runs backend + mongo + redis + web + web-public on
a single VM, pulling images published to
`ghcr.io/kitdoo/my-business-crm-go[-web|-web-public]` on every push to
`develop` (`.github/workflows/docker-publish.yml`), or built locally via
`make build-docker-images` / `docker compose up --build`.

```bash
cp .env.docker-compose.example .env   # fill in real secrets, never commit .env
docker compose up -d
```

Mongo/Redis and the backend's gRPC/HTTP ports aren't published to the
host, only reached over the compose-internal network; `web`/`web-public`
are published on `WEB_PORT`/`WEB_PUBLIC_PORT` (default `3000`/`3001`)
behind your own reverse proxy. All required secrets (Mongo credentials,
`CRM__AUTH__SIGNING_KEY`, `NUXT_SESSION_SECRET`, SMTP, RBAC table, etc.)
are documented inline in `.env.docker-compose.example`.

## License

Proprietary — see [LICENSE](./LICENSE). All rights reserved.
