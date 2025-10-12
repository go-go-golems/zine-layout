package imagesequences

import (
	"context"
	"fmt"
	"time"

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
		return fmt.Errorf("sequence ID is required")
	}

	url := fmt.Sprintf("%s/api/image-sequences/%s", settings.Server, settings.SequenceID)
	if err := httpDelete(url); err != nil {
		return fmt.Errorf("failed to delete sequence: %w", err)
	}

	row := types.NewRow(
		types.MRP("sequence_id", settings.SequenceID),
		types.MRP("status", "deleted"),
		types.MRP("timestamp", time.Now().UTC().Format(time.RFC3339)),
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
			cmds.WithShort("Delete an image sequence"),
			cmds.WithLong(`
Delete an image sequence and all of its items.

Examples:
  image-sequences delete --sequence-id seq-123
			`),
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
					parameters.WithDefault(""),
					parameters.WithHelp("Sequence ID (required)"),
					parameters.WithShortFlag("q"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &DeleteCommand{}
