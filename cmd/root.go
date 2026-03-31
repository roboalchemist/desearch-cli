package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "desearch",
	Short: "A CLI tool for Desearch AI",
	Long:  `CLI tool for Desearch AI - a contextual AI search engine that aggregates results across multiple sources.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}
