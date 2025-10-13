package imagesequencescmd

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

type imageSequencesAddItemCommand struct {
	*cmds.CommandDescription
}

type imageSequencesAddItemSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	SequenceID string `glazed.parameter:"sequence-id"`
	AssetID    string `glazed.parameter:"asset-id"`
	Gap        bool   `glazed.parameter:"gap"`
}

func (c *imageSequencesAddItemCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &imageSequencesAddItemSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SequenceID == "" {
		return fmt.Errorf("--sequence-id is required")
	}
	if !settings.Gap && strings.TrimSpace(settings.AssetID) == "" {
		return fmt.Errorf("--asset-id is required unless --gap is set")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	sequence, err := repos.ImageSequences.Get(settings.SequenceID)
	if err != nil {
		return err
	}

	var assetIDPtr *string
	if !settings.Gap {
		assetID := strings.TrimSpace(settings.AssetID)
		if err := ensureAssetInProject(repos, sequence.ProjectID, assetID); err != nil {
			return err
		}
		assetIDPtr = &assetID
	}

	item := &repo.ImageSequenceItem{
		SequenceID: settings.SequenceID,
		AssetID:    assetIDPtr,
		IsGap:      settings.Gap,
	}
	if err := repos.ImageSequences.AddItem(item); err != nil {
		return err
	}

	items, err := repos.ImageSequences.ListItems(settings.SequenceID)
	if err != nil {
		return err
	}

	for _, it := range items {
		row := types.NewRow(
			types.MRP("entity", "image_sequence_item"),
			types.MRP("sequence_id", it.SequenceID),
			types.MRP("position", it.Position),
			types.MRP("asset_id", it.AssetID),
			types.MRP("is_gap", it.IsGap),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newImageSequencesAddItemCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &imageSequencesAddItemCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add-item",
			cmds.WithShort("Append an item to an image sequence"),
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
					"asset-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Asset identifier (omit with --gap)"),
				),
				parameters.NewParameterDefinition(
					"gap",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Insert a gap item"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &imageSequencesAddItemCommand{}
