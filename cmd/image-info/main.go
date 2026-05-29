package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/zine-layout/cmd/image-info/cmds"
)

func main() {
	// Build the glazed command
	c, err := cmds.NewImageInfoCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating command: %v\n", err)
		os.Exit(1)
	}

	// Convert into cobra command
	cobraCmd, err := cli.BuildCobraCommand(c,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building cobra command: %v\n", err)
		os.Exit(1)
	}

	// Execute
	if err := cobraCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
