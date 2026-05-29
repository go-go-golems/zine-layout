package imagesequencescmd

import (
	"context"
	"fmt"

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

type imageSequencesReorderCommand struct {
	*cmds.CommandDescription
}

type imageSequencesReorderSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	SequenceID string `glazed.parameter:"sequence-id"`
	Items      string `glazed.parameter:"items"`
}

func (c *imageSequencesReorderCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &imageSequencesReorderSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SequenceID == "" {
		return fmt.Errorf("--sequence-id is required")
	}

	tokens := parseItemTokens(settings.Items)
	if len(tokens) == 0 {
		return fmt.Errorf("provide at least one item token")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	sequence, err := repos.ImageSequences.Get(settings.SequenceID)
	if err != nil {
		return err
	}

	replacements := make([]*repo.ImageSequenceItem, 0, len(tokens))
	for _, token := range tokens {
		if token == "" {
			replacements = append(replacements, &repo.ImageSequenceItem{
				SequenceID: settings.SequenceID,
				AssetID:    nil,
				IsGap:      true,
			})
			continue
		}
		if err := ensureAssetInProject(repos, sequence.ProjectID, token); err != nil {
			return err
		}
		itemToken := token
		replacements = append(replacements, &repo.ImageSequenceItem{
			SequenceID: settings.SequenceID,
			AssetID:    &itemToken,
			IsGap:      false,
		})
	}

	if err := repos.ImageSequences.ReplaceItems(settings.SequenceID, replacements); err != nil {
		return err
	}

	items, err := repos.ImageSequences.ListItems(settings.SequenceID)
	if err != nil {
		return err
	}

	for _, item := range items {
		row := types.NewRow(
			types.MRP("entity", "image_sequence_item"),
			types.MRP("sequence_id", item.SequenceID),
			types.MRP("position", item.Position),
			types.MRP("asset_id", item.AssetID),
			types.MRP("is_gap", item.IsGap),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newImageSequencesReorderCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &imageSequencesReorderCommand{
		CommandDescription: cmds.NewCommandDescription(
			"reorder",
			cmds.WithShort("Replace sequence items with a new ordering"),
			cmds.WithLong(`Provide comma-separated asset IDs; use 'gap' to insert blank slots.`),
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
					"items",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Comma-separated asset IDs; use 'gap' for blanks"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &imageSequencesReorderCommand{}
