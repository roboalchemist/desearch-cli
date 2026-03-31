package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X main.version=$(git describe --tags)"
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of desearch",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("desearch-cli version", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
