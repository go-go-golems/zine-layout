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

type layoutSequencesReorderCommand struct {
	*cmds.CommandDescription
}

type layoutSequencesReorderSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	SequenceID string `glazed.parameter:"sequence-id"`
	Order      string `glazed.parameter:"order"`
}

func (c *layoutSequencesReorderCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &layoutSequencesReorderSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SequenceID == "" {
		return fmt.Errorf("--sequence-id is required")
	}
	orderRaw := strings.TrimSpace(settings.Order)
	if orderRaw == "" {
		return fmt.Errorf("--order must contain laid-out image ids")
	}

	ids := []string{}
	for _, part := range strings.Split(orderRaw, ",") {
		id := strings.TrimSpace(part)
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("--order must include at least one laid-out image id")
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

	replacements := make([]*repo.LayoutSequenceItem, 0, len(ids))
	for pos, id := range ids {
		if err := ensureLaidOutImageInProject(repos, sequence.ProjectID, id); err != nil {
			return err
		}
		replacements = append(replacements, &repo.LayoutSequenceItem{
			SequenceID:     settings.SequenceID,
			Position:       pos,
			LaidOutImageID: id,
		})
	}

	if err := repos.LayoutSequences.ReplaceItems(settings.SequenceID, replacements); err != nil {
		return err
	}

	items, err := repos.LayoutSequences.ListItems(settings.SequenceID)
	if err != nil {
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

func newLayoutSequencesReorderCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &layoutSequencesReorderCommand{
		CommandDescription: cmds.NewCommandDescription(
			"reorder",
			cmds.WithShort("Replace layout sequence ordering"),
			cmds.WithLong(`Provide a comma-separated list of laid-out image IDs to define the sequence order.`),
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
					"order",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Comma-separated laid-out image IDs"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &layoutSequencesReorderCommand{}
