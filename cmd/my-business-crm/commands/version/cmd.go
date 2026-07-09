package version

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/altessa-s/go-atlas/core/runtime/appinfo"

	coreerrs "github.com/altessa-s/go-atlas/core/errors"
)

// command wraps [cobra.Command] for the "version" subcommand.
type command struct {
	*cobra.Command
}

// New returns the "version" [cobra.Command] with --full/-f and
// --output/-o flags configured. Supported output formats are "text"
// (default) and "json".
func New() *cobra.Command {
	v := &command{
		Command: &cobra.Command{
			Use:   "version",
			Short: "Display version information",
			Long: heredoc.Doc(`
				Display version information for my-business-crm.

				By default, only the version number is shown. Use --full to display
				additional build information including commit hash, build time, and
				project details.
			`),
			Example: heredoc.Doc(`
				# Display the current version
				my-business-crm version

				# Display full version information
				my-business-crm version --full

				# Display version information in JSON format
				my-business-crm version --output json
			`),
			SilenceUsage: true,
		},
	}

	v.configure()
	return v.Command
}

func (c *command) configure() {
	c.Flags().BoolP("full", "f", false, "Display full version information")
	c.Flags().StringP("output", "o", "text", "Output format: text (default), json")

	c.RunE = c.run
}

func (c *command) run(cmd *cobra.Command, _ []string) error {
	output, _ := cmd.Flags().GetString("output") //nolint:errcheck // flag is registered as String in configure
	full, _ := cmd.Flags().GetBool("full")       //nolint:errcheck // flag is registered as Bool in configure

	switch output {
	case "json":
		return printJSON(full)
	case "text":
		printText(full)
		return nil
	default:
		return fmt.Errorf("invalid output format: %s. Valid formats are: text, json", output)
	}
}

// printText writes version info to stdout. When full is true the output
// includes commit hash, build time, and platform details via [appinfo.AppVersion].
func printText(full bool) {
	if full {
		fmt.Println(appinfo.AppVersion())
	} else {
		fmt.Println(appinfo.Version)
	}
}

// printJSON writes version info to stdout as indented JSON.
func printJSON(full bool) error {
	var data any

	if full {
		data = struct {
			Version   string `json:"version"`
			BuildTime string `json:"build_time"`
			Commit    string `json:"commit"`
			Project   string `json:"project"`
			EnvPrefix string `json:"env_prefix"`
			GoVersion string `json:"go_version"`
			Platform  string `json:"platform"`
		}{
			Version:   appinfo.Version,
			BuildTime: appinfo.BuildTime,
			Commit:    appinfo.Commit,
			Project:   appinfo.Project,
			EnvPrefix: appinfo.EnvPrefix,
			GoVersion: appinfo.BuildGoVersion(),
			Platform:  appinfo.BuildPlatform(),
		}
	} else {
		data = struct {
			Version string `json:"version"`
		}{Version: appinfo.Version}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return coreerrs.WrapOperation(err, "marshal version info")
	}
	fmt.Println(string(jsonData))
	return nil
}
