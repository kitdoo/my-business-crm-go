// Package fx wires the service using Uber Fx dependency injection.
//
// It is organized into two modules:
//
//   - [InfrastructureModule] — MongoDB, Redis, TLS, metrics, tracing,
//     health coordinator, rate limiting, idempotency
//   - [TransportsModule] — gRPC and HTTP servers with handler registration
//
// The infrastructure constructors return `nil, nil` when their config section
// is absent, so optional subsystems can be left out of `configs/*.yaml`.
package fx
