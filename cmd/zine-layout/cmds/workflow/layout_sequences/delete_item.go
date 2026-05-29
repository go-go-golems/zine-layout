package layoutsequencescmd

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
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/spf13/cobra"
)

type layoutSequencesDeleteItemCommand struct {
	*cmds.CommandDescription
}

type layoutSequencesDeleteItemSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	SequenceID string `glazed.parameter:"sequence-id"`
	Position   int    `glazed.parameter:"position"`
}

func (c *layoutSequencesDeleteItemCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &layoutSequencesDeleteItemSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SequenceID == "" {
		return fmt.Errorf("--sequence-id is required")
	}
	if settings.Position < 0 {
		return fmt.Errorf("--position must be zero or greater")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := repos.LayoutSequences.DeleteItem(settings.SequenceID, settings.Position); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "layout_sequence_item"),
		types.MRP("sequence_id", settings.SequenceID),
		types.MRP("position", settings.Position),
		types.MRP("status", "deleted"),
		types.MRP("deleted_at", time.Now().UTC().Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newLayoutSequencesDeleteItemCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &layoutSequencesDeleteItemCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete-item",
			cmds.WithShort("Remove a laid-out image from a layout sequence"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"sequence-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Sequence identifier"),
				),
				parameters.NewParameterDefinition(
					"position",
					parameters.ParameterTypeInteger,
					parameters.WithRequired(true),
					parameters.WithHelp("Zero-based position to delete"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &layoutSequencesDeleteItemCommand{}
