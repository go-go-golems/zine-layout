package layoutsequences

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type DeleteCommand struct {
	*cmds.CommandDescription
}

type DeleteSettings struct {
	Server     string `glazed.parameter:"server"`
	SequenceID string `glazed.parameter:"sequence-id"`
}

func (c *DeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &DeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SequenceID == "" {
		return fmt.Errorf("sequence-id is required")
	}

	url := fmt.Sprintf("%s/api/layout-sequences/%s", settings.Server, settings.SequenceID)
	if err := httpDelete(url); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("layout_sequence_id", settings.SequenceID),
		types.MRP("status", "deleted"),
	)
	return gp.AddRow(ctx, row)
}

func NewDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &DeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a layout sequence"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"server",
					parameters.ParameterTypeString,
					parameters.WithDefault("http://localhost:8088"),
					parameters.WithHelp("Server base URL"),
					parameters.WithShortFlag("s"),
				),
				parameters.NewParameterDefinition(
					"sequence-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Layout sequence ID"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &DeleteCommand{}
