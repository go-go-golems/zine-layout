package projectscmd

import (
	"context"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/go-go-golems/zine-layout/pkg/projects"
	"github.com/go-go-golems/zine-layout/pkg/repo"
	"github.com/spf13/cobra"
)

type projectsCreateCommand struct {
	*cmds.CommandDescription
}

type projectsCreateSettings struct {
	DataRoot     string `glazed.parameter:"data-root"`
	Name         string `glazed.parameter:"name"`
	Description  string `glazed.parameter:"description"`
	SkipScaffold bool   `glazed.parameter:"skip-scaffold"`
}

func (c *projectsCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &projectsCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	name := strings.TrimSpace(settings.Name)
	if name == "" {
		name = "Untitled Project"
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	project := &repo.Project{
		Name:        name,
		Description: strings.TrimSpace(settings.Description),
	}
	if err := repos.Projects.Create(project); err != nil {
		return err
	}

	if !settings.SkipScaffold {
		if err := projects.EnsureProjectDirs(workflowshared.ProjectsRoot(settings.DataRoot), project.ID); err != nil {
			return err
		}
	}

	row := types.NewRow(
		types.MRP("entity", "project"),
		types.MRP("project_id", project.ID),
		types.MRP("name", project.Name),
		types.MRP("description", project.Description),
		types.MRP("created_at", project.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", project.UpdatedAt.Format(time.RFC3339)),
		types.MRP("scaffold", !settings.SkipScaffold),
	)
	return gp.AddRow(ctx, row)
}

func newProjectsCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &projectsCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a new project"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Project name (defaults to 'Untitled Project')"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Project description"),
				),
				parameters.NewParameterDefinition(
					"skip-scaffold",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Skip creating on-disk project folders"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &projectsCreateCommand{}
