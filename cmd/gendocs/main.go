package main

import (
	"log"
	"os"

	"github.com/roboalchemist/desearch-cli/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	if err := doc.GenMan(cmd.RootCmd(), nil, os.Stdout); err != nil {
		log.Fatal(err)
	}
}