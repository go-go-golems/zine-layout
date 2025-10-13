package imagesequencescmd

import (
	"context"
	"fmt"
	"strings"
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

type imageSequencesUpdateCommand struct {
	*cmds.CommandDescription
}

type imageSequencesUpdateSettings struct {
	DataRoot    string `glazed.parameter:"data-root"`
	SequenceID  string `glazed.parameter:"sequence-id"`
	Name        string `glazed.parameter:"name"`
	Description string `glazed.parameter:"description"`
}

func (c *imageSequencesUpdateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &imageSequencesUpdateSettings{}
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
	defer db.Close()

	sequence, err := repos.ImageSequences.Get(settings.SequenceID)
	if err != nil {
		return err
	}

	if trimmed := strings.TrimSpace(settings.Name); trimmed != "" {
		sequence.Name = trimmed
	}
	if desc := strings.TrimSpace(settings.Description); desc != "" {
		sequence.Description = desc
	}
	sequence.UpdatedAt = time.Now().UTC()

	if err := repos.ImageSequences.Update(sequence); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "image_sequence"),
		types.MRP("sequence_id", sequence.ID),
		types.MRP("project_id", sequence.ProjectID),
		types.MRP("name", sequence.Name),
		types.MRP("description", sequence.Description),
		types.MRP("updated_at", sequence.UpdatedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newImageSequencesUpdateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &imageSequencesUpdateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Update sequence metadata"),
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
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New sequence name"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New sequence description"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &imageSequencesUpdateCommand{}
