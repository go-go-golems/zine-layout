package imagesequencescmd

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

type imageSequencesListCommand struct {
	*cmds.CommandDescription
}

type imageSequencesListSettings struct {
	DataRoot  string `glazed.parameter:"data-root"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *imageSequencesListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &imageSequencesListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("--project-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	sequences, err := repos.ImageSequences.ListByProject(settings.ProjectID)
	if err != nil {
		return err
	}

	for _, seq := range sequences {
		row := types.NewRow(
			types.MRP("entity", "image_sequence"),
			types.MRP("sequence_id", seq.ID),
			types.MRP("project_id", seq.ProjectID),
			types.MRP("name", seq.Name),
			types.MRP("description", seq.Description),
			types.MRP("updated_at", seq.UpdatedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newImageSequencesListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &imageSequencesListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List image sequences for a project"),
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
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &imageSequencesListCommand{}
