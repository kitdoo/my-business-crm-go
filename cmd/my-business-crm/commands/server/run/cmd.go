package run

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

// command wraps [cobra.Command] for the "server run" leaf command.
type command struct {
	*cobra.Command
}

// New returns the "run" [cobra.Command] wired to start [App].
func New() *cobra.Command {
	r := &command{
		Command: &cobra.Command{
			Use:   "run [flags]",
			Short: "Start and run the my-business-crm server",
			Long: heredoc.Doc(`
				Start the my-business-crm server and begin serving requests.

				The server will load configuration from the specified file or directory,
				initialize all required infrastructure, and start the gRPC and HTTP
				listeners.

				The server runs until it receives a termination signal (SIGTERM or SIGINT),
				at which point it will gracefully shutdown all connections and services.
			`),
			Example: heredoc.Doc(`
				# Start the server with the default configuration path
				my-business-crm server run

				# Start the server with a specific configuration file
				my-business-crm server run --config /path/to/config.yaml
			`),
			SilenceUsage: true,
		},
	}

	r.configure()
	return r.Command
}

func (c *command) configure() {
	c.RunE = c.run
}

func (c *command) run(cmd *cobra.Command, args []string) error {
	app := NewApp()
	return app.Run(cmd, args)
}
