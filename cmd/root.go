package cmd

import (
	"fmt"
	"os"

	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/roboalchemist/desearch-cli/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	apiKey      string
	jsonOut     bool
	flagVerbose bool
	flagQuiet   bool
)

// version is set at build time via -ldflags "-X github.com/roboalchemist/desearch-cli/cmd.version=$(git describe --tags)"
var version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "desearch",
	Short:         "A CLI tool for Desearch AI",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version,
	Long: `CLI tool for Desearch AI - a contextual AI search engine that aggregates results across multiple sources.

ENVIRONMENT
  DESEARCH_API_KEY  API key for authentication (overrides config file)
  XDG_CONFIG_HOME   Config directory base (default ~/.config)
  NO_COLOR          Disable colored output when set to any non-empty value

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
			if errors.IsSystem(err) {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(3)
			}
			// For non-system errors (e.g. parse errors on an existing file),
			// print a warning and continue - flags may override the missing config.
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip API key check for certain commands that don't need auth
		if isNoAuthCommand(cmd) {
			return
		}

		// Validate API key is available (either from config or --api-key flag)
		// Skip check if --dry-run or --fields flag is set
		if cmd.Flags().Changed("dry-run") || cmd.Flags().Changed("fields") {
			return
		}
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

// isNoAuthCommand checks if the command or any of its ancestors don't require auth
func isNoAuthCommand(cmd *cobra.Command) bool {
	noAuthCommands := map[string]bool{
		"version":     true,
		"help":        true,
		"docs":        true,
		"skill":       true,
		"print":       true,
		"add":         true,
		"completion":  true,
		"ai":          true,
		"bash":        true,
		"zsh":         true,
		"fish":        true,
		"powershell":  true,
		"clear":       true,
	}
	for c := cmd; c != nil; c = c.Parent() {
		if noAuthCommands[c.Name()] {
			return true
		}
	}
	return false
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("desearch {{.Version}}\nCopyright 2026 RoboAlchemist\n")
	return rootCmd.Execute()
}

// RootCmd returns the root cobra command for use by gendocs.
func RootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default ~/.config/desearch-cli/config.toml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication (overrides config file)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "V", false, "Print version")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Show verbose progress output to stderr")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress stderr output except errors")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "silent", "", false, "Suppress stderr output except errors (alias for --quiet)")

	// GNU standard: --help should end with "Report bugs" footer
	rootCmd.SetHelpTemplate(rootCmd.HelpTemplate() + "\nReport bugs at: https://gitea.roboalch.com/roboalchemist/desearch-cli/issues\n")
}
