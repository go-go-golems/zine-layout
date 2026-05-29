package layoutsequencescmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/go-go-golems/zine-layout/pkg/repo"
	"github.com/spf13/cobra"
)

type layoutSequencesAddItemCommand struct {
	*cmds.CommandDescription
}

type layoutSequencesAddItemSettings struct {
	DataRoot       string `glazed.parameter:"data-root"`
	SequenceID     string `glazed.parameter:"sequence-id"`
	LaidOutImageID string `glazed.parameter:"laid-out-image-id"`
}

func (c *layoutSequencesAddItemCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &layoutSequencesAddItemSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SequenceID == "" {
		return fmt.Errorf("--sequence-id is required")
	}
	if strings.TrimSpace(settings.LaidOutImageID) == "" {
		return fmt.Errorf("--laid-out-image-id is required")
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

	if err := ensureLaidOutImageInProject(repos, sequence.ProjectID, settings.LaidOutImageID); err != nil {
		return err
	}

	item := &repo.LayoutSequenceItem{
		SequenceID:     settings.SequenceID,
		LaidOutImageID: strings.TrimSpace(settings.LaidOutImageID),
	}
	if err := repos.LayoutSequences.AddItem(item); err != nil {
		return err
	}

	items, err := repos.LayoutSequences.ListItems(settings.SequenceID)
	if err != nil {
		return err
	}

	for _, it := range items {
		row := types.NewRow(
			types.MRP("entity", "layout_sequence_item"),
			types.MRP("sequence_id", it.SequenceID),
			types.MRP("position", it.Position),
			types.MRP("laid_out_image_id", it.LaidOutImageID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newLayoutSequencesAddItemCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &layoutSequencesAddItemCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add-item",
			cmds.WithShort("Append a laid-out image to a layout sequence"),
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
					"laid-out-image-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Laid-out image identifier"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &layoutSequencesAddItemCommand{}
