package cmd

import (
	"fmt"
	"os"

	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	apiKey  string
	jsonOut bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "desearch",
	Short: "A CLI tool for Desearch AI",
	Long: `CLI tool for Desearch AI - a contextual AI search engine that aggregates results across multiple sources.

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
			fmt.Fprintln(os.Stderr, "Error: No API key found.")
			fmt.Fprintln(os.Stderr, "Please provide an API key via the --api-key flag or configure one at ~/.config/desearch-cli/config.toml")
			fmt.Fprintln(os.Stderr, "Sign up at https://console.desearch.ai to get an API key")
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default ~/.config/desearch-cli/config.toml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication (overrides config file)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
}
