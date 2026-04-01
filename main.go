package main

import (
	"fmt"
	"os"

	"github.com/roboalchemist/desearch-cli/cmd"
	"github.com/roboalchemist/desearch-cli/pkg/errors"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if errors.IsSystem(err) {
			os.Exit(3)
		}
		if errors.IsUsage(err) {
			os.Exit(2)
		}
		if cmd.GetJSONOut() {
			fmt.Fprintf(os.Stderr, "{\"error\": %q}\n", err.Error())
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
