package cmd

import (
	"encoding/json"
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
	Example: `  desearch config --api-key sk-xxx  # Set API key
  desearch config --show              # Show current config
  desearch config --clear             # Clear all config`,
}

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Display current configuration",
	Example: `  desearch config --show`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := auth.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if jsonOut {
			// Output full config as JSON
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config as JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
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
		return nil
	},
}

var clearCmd = &cobra.Command{
	Use:     "clear",
	Short:   "Reset configuration to defaults",
	Example: `  desearch config --clear`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := auth.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		xdgPath, err := auth.ConfigPath()
		if err != nil {
			return fmt.Errorf("getting config path: %w", err)
		}

		if _, err := os.Stat(xdgPath); os.IsNotExist(err) {
			fmt.Println("No config file to clear.")
			return nil
		}

		if err := os.Remove(xdgPath); err != nil {
			return fmt.Errorf("removing config file: %w", err)
		}

		fmt.Println("Configuration cleared.")
		return nil
	},
}

var (
	flagAPIKey            string
	flagDefaultTools      []string
	flagDefaultDateFilter string
	flagForce             bool
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(clearCmd)

	configCmd.Flags().StringVar(&flagAPIKey, "api-key", "", "Set API key")
	configCmd.Flags().StringSliceVar(&flagDefaultTools, "default-tool", nil, "Set default sources (can be specified multiple times)")
	configCmd.Flags().StringVar(&flagDefaultDateFilter, "default-date-filter", "", "Set date filter (e.g., PAST_24_HOURS, PAST_WEEK, PAST_MONTH)")
	clearCmd.Flags().BoolVarP(&flagForce, "force", "f", false, "Force clear without confirmation")

	// Wire --api-key, --default-tool, --default-date-filter to run the set subcommand implicitly
	configCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// If no flags were provided, show help
		if flagAPIKey == "" && flagDefaultDateFilter == "" && len(flagDefaultTools) == 0 {
			_ = cmd.Help()
			return nil
		}

		cfg := &auth.Config{}

		// Load existing config to preserve values not being set
		existing, err := auth.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Start with existing values
		*cfg = *existing

		// Apply flag overrides
		if flagAPIKey != "" {
			if strings.TrimSpace(flagAPIKey) == "" {
				return fmt.Errorf("API key cannot be empty")
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
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Println("Configuration saved.")
		return nil
	}
}
