package cmd

import (
	"fmt"
	"os"

	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/roboalchemist/desearch-cli/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	cfgFile            string
	apiKey             string
	jsonOut            bool
	flagVerbose        bool
	flagQuiet          bool
	dispatchedToSubcmd bool // true when PreRunE manually dispatched to a subcommand
)

// version is set at build time via -ldflags "-X github.com/roboalchemist/desearch-cli/cmd.version=$(git describe --tags)"
var version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "desearch-cli",
	Short:         "A CLI tool for Desearch AI",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version,
	// Args wraps the default legacyArgs validator so that unknown subcommand
	// errors are tagged as UsageErrors (exit code 2) rather than generic errors.
	// When GNU "--" dispatch is used (e.g. "desearch -- search query"), Cobra stops
	// subcommand routing and passes args=["search", "query", ...] to root's Args.
	// We detect known subcommand names here and return nil so PreRunE can handle the
	// dispatch. Only truly unknown commands produce a usage error.
	Args: func(cmd *cobra.Command, args []string) error {
		if cmd.HasSubCommands() && !cmd.HasParent() && len(args) > 0 {
			subCmd, _, err := cmd.Find([]string{args[0]})
			if err == nil && subCmd != cmd {
				return nil
			}
			return errors.WrapUsage(fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath()))
		}
		return nil
	},
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
  Report bugs at: https://github.com/roboalchemist/desearch-cli/issues

To get started, you need an API key. Sign up at https://console.desearch.ai`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
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

		// GNU standard: "program -- subcommand args" should behave identically to
		// "program subcommand args". When -- is used, Cobra stops subcommand routing
		// and treats the subcommand name as a positional arg to root. We detect this
		// and manually dispatch to the named subcommand.
		if len(args) == 0 {
			return nil
		}
		subCmd, _, err := cmd.Find([]string{args[0]})
		if err != nil || subCmd == cmd {
			return nil // not a known subcommand name; let RunE handle it
		}
		dispatchedToSubcmd = true
		remaining := args[1:]
		// Parse flags for the subcommand BEFORE calling RunE, so that
		// PersistentPreRun hooks (which run before RunE in the normal flow)
		// can see the flag state.
		if err := subCmd.ParseFlags(remaining); err != nil {
			return err
		}
		// Also propagate root-level flags (parsed before --) to the subcommand.
		// e.g. desearch --api-key KEY -- search query --dry-run
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				_ = subCmd.Flags().Set(f.Name, f.Value.String())
			}
		})
		// Now run PersistentPreRun on subcmd so it can see parsed flags.
		// PersistentPreRunE returns error; nil means proceed.
		if subCmd.PersistentPreRunE != nil {
			if err := subCmd.PersistentPreRunE(subCmd, subCmd.Flags().Args()); err != nil {
				return err
			}
		} else if subCmd.PersistentPreRun != nil {
			subCmd.PersistentPreRun(subCmd, subCmd.Flags().Args())
		}
		// Execute the subcommand.
		if subCmd.RunE != nil {
			return subCmd.RunE(subCmd, subCmd.Flags().Args())
		}
		if subCmd.Run != nil {
			subCmd.Run(subCmd, subCmd.Flags().Args())
			return nil
		}
		return fmt.Errorf("subcommand %q has no Run function", args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// If PreRunE dispatched to a subcommand, RunE on root should not fire.
		// Cobra calls RunE on the LEAF command only, but after PreRunE returns nil
		// on root, root's RunE still fires before subcommand dispatch completes.
		// The dispatchedToSubcmd sentinel lets us skip gracefully.
		if dispatchedToSubcmd {
			return nil
		}
		// Show help when invoked with no args (or after dispatching).
		return cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip API key check for certain commands that don't need auth
		if isNoAuthCommand(cmd) {
			return
		}

		// Check if --dry-run or --fields is set on this command or any parent.
		// Also handle the GNU standard "-- subcommand args" dispatch case where
		// flags are parsed on the subcommand in PreRunE but PersistentPreRun
		// runs on root first. We detect this by checking os.Args for the pattern.
		if hasChangedFlag(cmd, "dry-run") || hasChangedFlag(cmd, "fields") || hasDryRunInArgs() {
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

// hasChangedFlag checks if a flag was changed on this command or any ancestor.
func hasChangedFlag(cmd *cobra.Command, name string) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Flags().Changed(name) {
			return true
		}
	}
	return false
}

// hasDryRunInArgs checks if --dry-run appears in os.Args after a "--" separator.
// This detects the GNU standard "program -- subcommand --dry-run" dispatch pattern
// where PersistentPreRun on root runs before PreRunE can parse the subcommand's flags.
func hasDryRunInArgs() bool {
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			// "--" found, check remaining args for --dry-run or --fields
			for _, a := range os.Args[i+1:] {
				if a == "--dry-run" || a == "--fields" || a == "-D" {
					return true
				}
			}
		}
	}
	return false
}

// isNoAuthCommand checks if the command or any of its ancestors don't require auth
func isNoAuthCommand(cmd *cobra.Command) bool {
	noAuthCommands := map[string]bool{
		"desearch-cli": true, // root command - help, version, etc don't need auth
		"version":      true,
		"help":         true,
		"docs":         true,
		"skill":        true,
		"print":        true,
		"add":          true,
		"completion":   true,
		"bash":         true,
		"zsh":          true,
		"fish":         true,
		"powershell":   true,
		"clear":        true,
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
	dispatchedToSubcmd = false // reset between invocations (e.g. test suites calling Execute twice)
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("desearch-cli {{.Version}}\nCopyright 2026 RoboAlchemist\n")
	return rootCmd.Execute()
}

// RootCmd returns the root cobra command for use by gendocs.
func RootCmd() *cobra.Command {
	return rootCmd
}

// GetJSONOut returns whether --json flag was set.
func GetJSONOut() bool {
	return jsonOut
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default ~/.config/desearch-cli/config.toml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication (overrides config file)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Show verbose progress output to stderr")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress stderr output except errors")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "silent", "", false, "Suppress stderr output except errors (alias for --quiet)")

	// Tag flag parse errors as UsageErrors so main.go can exit with code 2.
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return errors.WrapUsage(err)
	})

	// GNU standard: --help should end with "Report bugs" footer
	rootCmd.SetHelpTemplate(rootCmd.HelpTemplate() + "\nReport bugs at: https://github.com/roboalchemist/desearch-cli/issues\n")
}
