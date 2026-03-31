package cmd

import (
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search using Desearch AI",
	Long:  `Search the web using Desearch AI's contextual search engine.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Stub implementation - DC1-6 will implement full search functionality
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
