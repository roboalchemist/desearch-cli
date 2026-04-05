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
			fmt.Fprintln(os.Stdout, string(data))
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

		fmt.Fprintf(os.Stdout, "API Key:               %s\n", apiKeyDisplay)
		fmt.Fprintf(os.Stdout, "Default Tools:         %v\n", cfg.DefaultTools)
		fmt.Fprintf(os.Stdout, "Default Date Filter:   %s\n", cfg.DefaultDateFilter)
		fmt.Fprintf(os.Stdout, "Default Count:         %d\n", cfg.DefaultCount)
		fmt.Fprintf(os.Stdout, "History Enabled:       %v\n", cfg.HistoryEnabled)
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
			fmt.Fprintln(os.Stdout, "No config file to clear.")
			return nil
		}

		if err := os.Remove(xdgPath); err != nil {
			return fmt.Errorf("removing config file: %w", err)
		}

		fmt.Fprintln(os.Stdout, "Configuration cleared.")
		return nil
	},
}

var (
	flagAPIKey            string
	flagDefaultTools      []string
	flagDefaultDateFilter string
	flagDefaultCount      int
	flagForce             bool
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(clearCmd)

	configCmd.Flags().StringVar(&flagAPIKey, "api-key", "", "Set API key")
	configCmd.Flags().StringSliceVar(&flagDefaultTools, "default-tool", nil, "Set default sources (can be specified multiple times)")
	configCmd.Flags().StringVar(&flagDefaultDateFilter, "default-date-filter", "", "Set date filter (e.g., PAST_24_HOURS, PAST_WEEK, PAST_MONTH)")
	configCmd.Flags().IntVar(&flagDefaultCount, "default-count", 0, "Set default result count per source (10-200, or 0 to clear)")
	configCmd.Flags().Bool("history-enabled", false, "Enable or disable history logging (use --history-enabled=true or --history-enabled=false)")
	clearCmd.Flags().BoolVarP(&flagForce, "force", "f", false, "Force clear without confirmation")

	// Wire all config flags to run the set subcommand implicitly
	configCmd.RunE = func(cmd *cobra.Command, args []string) error {
		historyEnabledChanged := cmd.Flags().Changed("history-enabled")
		defaultCountChanged := cmd.Flags().Changed("default-count")

		// If no flags were provided, show help
		if flagAPIKey == "" && flagDefaultDateFilter == "" && len(flagDefaultTools) == 0 &&
			!historyEnabledChanged && !defaultCountChanged {
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
		if historyEnabledChanged {
			val, _ := cmd.Flags().GetBool("history-enabled")
			cfg.HistoryEnabled = val
		}
		if defaultCountChanged {
			if flagDefaultCount != 0 && (flagDefaultCount < 10 || flagDefaultCount > 200) {
				return fmt.Errorf("--default-count must be 0 (to clear) or between 10 and 200, got %d", flagDefaultCount)
			}
			cfg.DefaultCount = flagDefaultCount
		}

		if err := auth.SaveConfig(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Fprintln(os.Stdout, "Configuration saved.")
		return nil
	}
}
