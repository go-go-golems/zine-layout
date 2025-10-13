package projectscmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/go-go-golems/zine-layout/pkg/projects"
	"github.com/spf13/cobra"
)

type projectsDeleteCommand struct {
	*cmds.CommandDescription
}

type projectsDeleteSettings struct {
	DataRoot      string `glazed.parameter:"data-root"`
	ProjectID     string `glazed.parameter:"project-id"`
	KeepArtifacts bool   `glazed.parameter:"keep-artifacts"`
}

func (c *projectsDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &projectsDeleteSettings{}
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

	if err := repos.Projects.Delete(settings.ProjectID); err != nil {
		return err
	}

	if !settings.KeepArtifacts {
		root := workflowshared.ProjectsRoot(settings.DataRoot)
		_ = os.RemoveAll(projects.ProjectDir(root, settings.ProjectID))
	}

	row := types.NewRow(
		types.MRP("entity", "project"),
		types.MRP("project_id", settings.ProjectID),
		types.MRP("status", "deleted"),
		types.MRP("artifacts_removed", !settings.KeepArtifacts),
		types.MRP("deleted_at", time.Now().UTC().Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newProjectsDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &projectsDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a project and optionally its files"),
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
					parameters.WithHelp("Project identifier to delete"),
				),
				parameters.NewParameterDefinition(
					"keep-artifacts",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Keep on-disk project directories"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &projectsDeleteCommand{}
