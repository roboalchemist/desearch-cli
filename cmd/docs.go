package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed README.md
var readme string

var docsCmd = &cobra.Command{
	Use:     "docs",
	Aliases: []string{"readme"},
	Short:   "Print full documentation",
	Long:    "Print the full README documentation to stdout",
	Example: "desearch docs",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(readme)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}