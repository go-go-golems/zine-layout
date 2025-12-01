package imagelayoutcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout/engine"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type computeCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*computeCommand)(nil)

// NewComputeCommand wires the imagelayout compute command.
func NewComputeCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "create glazed parameter layer")
	}

	cmd := &computeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"compute",
			cmds.WithShort("Compute viewport placement from a LayoutRequest spec or flags"),
			cmds.WithFlags(computeParameterDefinitions()...),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return cli.BuildCobraCommandFromCommand(
		cmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
}

func (c *computeCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &layoutParamSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if strings.TrimSpace(settings.Spec) == "" &&
		(settings.SourceWidth <= 0 || settings.SourceHeight <= 0) {
		return fmt.Errorf("--spec or --source-width/--source-height must be provided")
	}

	layout, _, inputs, err := buildLayoutFromSettings(parsedLayers, settings)
	if err != nil {
		return err
	}

	result, trace := engine.ComputeViewport(inputs)
	output := imagelayout.Computation{
		Layout: layout,
		Result: result,
		Trace:  trace,
	}
	pretty, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(pretty))
	return nil
}
