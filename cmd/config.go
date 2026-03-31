package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage API key and default settings",
	Long:  `Manage the CLI configuration including API key and default search settings.`,
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := auth.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		apiKeyDisplay := ""
		if cfg.APIKey != "" {
			if len(cfg.APIKey) <= 4 {
				apiKeyDisplay = "****"
			} else {
				apiKeyDisplay = cfg.APIKey[:4] + "****"
			}
		}

		fmt.Printf("API Key:               %s\n", apiKeyDisplay)
		fmt.Printf("Default Tools:         %v\n", cfg.DefaultTools)
		fmt.Printf("Default Date Filter:   %s\n", cfg.DefaultDateFilter)
		fmt.Printf("Default Count:         %d\n", cfg.DefaultCount)
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Reset configuration to defaults",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := auth.LoadConfig()
		_ = path // not used; just checking if config exists
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		xdgPath, err := configPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
			os.Exit(1)
		}

		if _, err := os.Stat(xdgPath); os.IsNotExist(err) {
			fmt.Println("No config file to clear.")
			return
		}

		if err := os.Remove(xdgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Configuration cleared.")
	},
}

func configPath() (string, error) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		xdgConfigHome = home + "/.config"
	}
	return xdgConfigHome + "/desearch-cli/config.toml", nil
}

var (
	flagAPIKey          string
	flagDefaultTools     []string
	flagDefaultDateFilter string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(clearCmd)

	configCmd.Flags().StringVar(&flagAPIKey, "api-key", "", "Set API key")
	configCmd.Flags().StringSliceVar(&flagDefaultTools, "default-tool", nil, "Set default sources (can be specified multiple times)")
	configCmd.Flags().StringVar(&flagDefaultDateFilter, "default-date-filter", "", "Set date filter (e.g., PAST_24_HOURS, PAST_WEEK, PAST_MONTH)")

	// Wire --api-key, --default-tool, --default-date-filter to run the set subcommand implicitly
	configCmd.Run = func(cmd *cobra.Command, args []string) {
		// If no flags were provided, show help
		if flagAPIKey == "" && flagDefaultDateFilter == "" && len(flagDefaultTools) == 0 {
			cmd.Help()
			return
		}

		cfg := &auth.Config{}

		// Load existing config to preserve values not being set
		existing, err := auth.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Start with existing values
		*cfg = *existing

		// Apply flag overrides
		if flagAPIKey != "" {
			if strings.TrimSpace(flagAPIKey) == "" {
				fmt.Fprintln(os.Stderr, "Error: API key cannot be empty")
				os.Exit(1)
			}
			cfg.APIKey = flagAPIKey
		}
		if flagDefaultDateFilter != "" {
			cfg.DefaultDateFilter = flagDefaultDateFilter
		}
		if len(flagDefaultTools) > 0 {
			cfg.DefaultTools = flagDefaultTools
		}

		if err := auth.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Configuration saved.")
	}
}
