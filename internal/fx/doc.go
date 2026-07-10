// Package fx wires the service using Uber Fx dependency injection.
//
// It is organized into three modules:
//
//   - [InfrastructureModule] — MongoDB, Redis, TLS, metrics, tracing,
//     health coordinator, rate limiting, idempotency
//   - [ServicesModule] — domain aggregate storages, services, and gRPC handlers
//   - [TransportsModule] — gRPC and HTTP servers with handler registration
//
// The infrastructure constructors return `nil, nil` when their config section
// is absent, so optional subsystems can be left out of `configs/*.yaml`.
package fx
