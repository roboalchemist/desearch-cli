package main

import (
	"os"

	"github.com/roboalchemist/desearch-cli/cmd"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
