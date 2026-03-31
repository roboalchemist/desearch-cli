package main

import (
	"github.com/roboalchemist/desearch-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
