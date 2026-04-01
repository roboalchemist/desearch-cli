package main

import (
	"os"

	"github.com/roboalchemist/desearch-cli/cmd"
	"github.com/roboalchemist/desearch-cli/pkg/errors"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if errors.IsSystem(err) {
			os.Exit(3)
		}
		os.Exit(1)
	}
}
