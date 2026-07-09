// my-business-crm is the entry point for the service binary.
//
// It loads environment variables from a local .env file (if present),
// then delegates to [commands.Run] which builds the Cobra CLI tree and
// executes the matched subcommand. The process exits with code 1 on any
// command error.
//
// Usage:
//
//	my-business-crm server run --config ./configs/
//	my-business-crm version --full
package main

import (
	"os"

	"github.com/joho/godotenv"

	"github.com/kitdoo/my-business-crm-go/cmd/my-business-crm/commands"
)

func init() {
	_ = godotenv.Load(".env") //nolint:errcheck // .env is optional; absence is the documented happy path
}

func main() {
	if err := commands.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
