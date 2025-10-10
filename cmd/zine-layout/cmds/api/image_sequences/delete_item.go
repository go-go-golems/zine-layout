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

type DeleteItemCommand struct {
	*cmds.CommandDescription
}

type DeleteItemSettings struct {
	Server     string `glazed.parameter:"server"`
	SequenceID string `glazed.parameter:"sequence-id"`
	Position   int    `glazed.parameter:"position"`
}

func (c *DeleteItemCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &DeleteItemSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.SequenceID == "" {
		return fmt.Errorf("sequence ID is required")
	}
	if settings.Position < 0 {
		return fmt.Errorf("position must be zero or positive")
	}

	url := fmt.Sprintf("%s/api/image-sequences/%s/items/%d", settings.Server, settings.SequenceID, settings.Position)
	if err := httpDelete(url); err != nil {
		return fmt.Errorf("failed to delete sequence item: %w", err)
	}

	row := types.NewRow(
		types.MRP("sequence_id", settings.SequenceID),
		types.MRP("position", settings.Position),
		types.MRP("status", "deleted"),
		types.MRP("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func NewDeleteItemCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &DeleteItemCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete-item",
			cmds.WithShort("Remove an item at a position from a sequence"),
			cmds.WithLong(`
Delete a single item from an image sequence by its zero-based position.

Examples:
  image-sequences delete-item --sequence-id seq-123 --position 0
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
				parameters.NewParameterDefinition(
					"position",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Zero-based position to remove"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &DeleteItemCommand{}
