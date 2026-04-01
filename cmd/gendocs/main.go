package main

import (
	"log"

	"github.com/roboalchemist/desearch-cli/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	if err := doc.GenManTree(cmd.RootCmd(), nil, "docs/"); err != nil {
		log.Fatal(err)
	}
}
