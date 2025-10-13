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
	"github.com/go-go-golems/zine-layout/pkg/repo"
	"github.com/spf13/cobra"
)

type imageSequencesCreateCommand struct {
	*cmds.CommandDescription
}

type imageSequencesCreateSettings struct {
	DataRoot    string `glazed.parameter:"data-root"`
	ProjectID   string `glazed.parameter:"project-id"`
	Name        string `glazed.parameter:"name"`
	Description string `glazed.parameter:"description"`
}

func (c *imageSequencesCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &imageSequencesCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if strings.TrimSpace(settings.ProjectID) == "" {
		return fmt.Errorf("--project-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := repos.Projects.Get(settings.ProjectID); err != nil {
		return fmt.Errorf("lookup project: %w", err)
	}

	name := strings.TrimSpace(settings.Name)
	if name == "" {
		name = "Untitled Sequence"
	}

	sequence := &repo.ImageSequence{
		ProjectID:   settings.ProjectID,
		Name:        name,
		Description: strings.TrimSpace(settings.Description),
	}
	if err := repos.ImageSequences.Create(sequence); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "image_sequence"),
		types.MRP("sequence_id", sequence.ID),
		types.MRP("project_id", sequence.ProjectID),
		types.MRP("name", sequence.Name),
		types.MRP("description", sequence.Description),
		types.MRP("created_at", sequence.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", sequence.UpdatedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newImageSequencesCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &imageSequencesCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create an image sequence"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"project-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Project identifier"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Sequence name"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Sequence description"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &imageSequencesCreateCommand{}
