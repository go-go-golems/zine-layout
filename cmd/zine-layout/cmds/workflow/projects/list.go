package projectscmd

import (
	"context"
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

type projectsListCommand struct {
	*cmds.CommandDescription
}

type projectsListSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
}

func (c *projectsListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &projectsListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	projects, err := repos.Projects.List()
	if err != nil {
		return err
	}

	for _, project := range projects {
		row := types.NewRow(
			types.MRP("entity", "project"),
			types.MRP("project_id", project.ID),
			types.MRP("name", project.Name),
			types.MRP("description", project.Description),
			types.MRP("created_at", project.CreatedAt.Format(time.RFC3339)),
			types.MRP("updated_at", project.UpdatedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newProjectsListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &projectsListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List projects from the local repository"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &projectsListCommand{}
