package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of desearch",
	Example: `  desearch version
  desearch version -V`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("desearch-cli version", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
