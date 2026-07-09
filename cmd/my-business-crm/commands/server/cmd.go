package server

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/kitdoo/my-business-crm-go/cmd/my-business-crm/commands/server/run"
)

// command wraps [cobra.Command] for the "server" subcommand group.
type command struct {
	*cobra.Command
}

// New returns the "server" [cobra.Command] with the --config/-c persistent
// flag and all server subcommands (currently only "run") already registered.
func New() *cobra.Command {
	srv := &command{
		Command: &cobra.Command{
			Use:   "server [command]",
			Short: "Commands to manage the my-business-crm server",
			Long: heredoc.Doc(`
				Server commands provide control over the my-business-crm lifecycle.

				The server supports various runtime modes and configuration options
				to accommodate different deployment scenarios. All server operations
				require proper configuration files to specify storage connections
				and other essential settings.
			`),
			Example: heredoc.Doc(`
				# Start the server with the default configuration path
				my-business-crm server run

				# Start the server with a specific configuration file
				my-business-crm server run --config /path/to/config.yaml

				# Start the server with a configuration directory
				my-business-crm server run -c ./configs/
			`),
			SilenceUsage: true,
		},
	}

	srv.configure()
	return srv.Command
}

func (c *command) configure() {
	c.Command.PersistentFlags().StringP("config", "c", "", heredoc.Doc(`
		Path to configuration file or directory containing configuration files.

		If a directory is specified, all .yaml files in the directory will be
		loaded and merged. Environment variables can override any configuration
		value using the CRM_ prefix.
	`))

	c.AddCommand(run.New())
}
