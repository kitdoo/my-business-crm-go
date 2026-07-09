// Package run implements the "my-business-crm server run" command, which is
// the primary way to start the service.
//
// Startup sequence:
//
//  1. Resolve the YAML configuration path (--config flag or CONFIG_FILE env).
//  2. Load and validate configuration via [appconfig.Load].
//  3. Initialize structured logging and create required filesystem directories.
//  4. Build the Uber Fx dependency graph (infrastructure, transports).
//  5. Start all Fx lifecycle hooks within [startTimeout].
//  6. Block until SIGINT / SIGTERM, then drain within [shutdownTimeout].
package run
