package cmd

import (
	"fmt"
	"os"

	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	apiKey      string
	jsonOut     bool
	versionFlag bool
)

// version is set at build time via -ldflags "-X github.com/roboalchemist/desearch-cli/cmd.version=$(git describe --tags)"
var version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "desearch",
	Short:             "A CLI tool for Desearch AI",
	SilenceUsage:     true,
	SilenceErrors:    true,
	Version:           version,
	Long: `CLI tool for Desearch AI - a contextual AI search engine that aggregates results across multiple sources.

ENVIRONMENT
  DESEARCH_API_KEY  API key for authentication (overrides config file)

FILES
  ~/.config/desearch-cli/config.toml  Configuration file (mode 0600)

EXIT STATUS
  0  Success
  1  User error (invalid arguments, API error)
  2  Usage error (unknown flag or command)
  3+ System error (network failure, config error)

BUGS
  Report bugs at: https://gitea.roboalch.com/roboalchemist/desearch-cli/issues

To get started, you need an API key. Sign up at https://console.desearch.ai`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// Load config before any subcommand runs
		_, err := auth.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			// Continue anyway - flags may override the missing config
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip API key check for certain commands that don't need auth
		if cmd.Name() == "version" || cmd.Name() == "help" {
			return
		}

		// Validate API key is available (either from config or --api-key flag)
		key := apiKey
		if key == "" {
			key = auth.GetAPIKey()
		}
		if key == "" {
			// Hard error - PersistentPreRun cannot return errors, so we exit
			fmt.Fprintln(os.Stderr, "Error: No API key found.")
			fmt.Fprintln(os.Stderr, "Please provide an API key via the --api-key flag or configure one at ~/.config/desearch-cli/config.toml")
			fmt.Fprintln(os.Stderr, "Sign up at https://console.desearch.ai to get an API key")
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default ~/.config/desearch-cli/config.toml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication (overrides config file)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "V", false, "Print version")
}
