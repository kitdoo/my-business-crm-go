package commands

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/altessa-s/go-atlas/core/runtime/appinfo"

	"github.com/kitdoo/my-business-crm-go/cmd/my-business-crm/commands/server"
	"github.com/kitdoo/my-business-crm-go/cmd/my-business-crm/commands/version"
)

// command wraps [cobra.Command] to add my-business-crm-specific configuration
// such as subcommand registration and output stream setup.
type command struct {
	*cobra.Command
}

// New builds the root [cobra.Command] with all subcommands (server, version)
// already registered. The returned command is ready to execute.
func New() *cobra.Command {
	root := &command{
		Command: &cobra.Command{
			Use:   "my-business-crm",
			Short: "my-business-crm service",
			Long: heredoc.Doc(`
				my-business-crm is a Uber Fx / Cobra service skeleton built on go-atlas.

				It boots MongoDB, Redis, TLS, metrics, tracing, and health-check
				infrastructure, then exposes it through gRPC and HTTP servers with the
				standard probe/metrics endpoints. Domain logic is added on top.
			`),
			Example: heredoc.Doc(`
				# Start the server
				my-business-crm server run

				# Start with a specific configuration file
				my-business-crm server run --config /path/to/config.yaml

				# Show version information
				my-business-crm version

				# Show detailed help for server commands
				my-business-crm server --help
			`),
			SilenceUsage: true,
			Version:      appinfo.Version,
		},
	}

	root.configure()

	return root.Command
}

// configure wires stderr/stdout, enables typo suggestions, sets the
// version template, and registers the version and server subcommands.
func (c *command) configure() {
	c.SetErr(os.Stderr)
	c.SetOut(os.Stdout)
	c.SuggestionsMinimumDistance = 1
	c.SetVersionTemplate(heredoc.Doc(`
		{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
	`))
	c.AddCommand(version.New())
	c.AddCommand(server.New())
}

// Run creates a fresh root command, sets the given CLI arguments, and
// executes it. It is the main entry point called by the binary's main
// function, and can also be called directly in integration tests.
func Run(args []string) error {
	cmd := New()
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		return err
	}

	return nil
}
