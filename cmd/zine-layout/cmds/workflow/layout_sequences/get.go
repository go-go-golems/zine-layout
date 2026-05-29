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

type layoutSequencesGetCommand struct {
	*cmds.CommandDescription
}

type layoutSequencesGetSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	SequenceID string `glazed.parameter:"sequence-id"`
}

func (c *layoutSequencesGetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &layoutSequencesGetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SequenceID == "" {
		return fmt.Errorf("--sequence-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	sequence, err := repos.LayoutSequences.Get(settings.SequenceID)
	if err != nil {
		return err
	}
	items, err := repos.LayoutSequences.ListItems(settings.SequenceID)
	if err != nil {
		return err
	}

	header := types.NewRow(
		types.MRP("entity", "layout_sequence"),
		types.MRP("sequence_id", sequence.ID),
		types.MRP("project_id", sequence.ProjectID),
		types.MRP("name", sequence.Name),
		types.MRP("description", sequence.Description),
		types.MRP("created_at", sequence.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", sequence.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, header); err != nil {
		return err
	}

	for _, item := range items {
		row := types.NewRow(
			types.MRP("entity", "layout_sequence_item"),
			types.MRP("sequence_id", item.SequenceID),
			types.MRP("position", item.Position),
			types.MRP("laid_out_image_id", item.LaidOutImageID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newLayoutSequencesGetCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &layoutSequencesGetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Show a layout sequence and its items"),
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
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &layoutSequencesGetCommand{}
