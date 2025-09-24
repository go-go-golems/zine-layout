package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type ProjectsDeleteCommand struct {
	*cmds.CommandDescription
}

type ProjectsDeleteSettings struct {
	Server string `glazed.parameter:"server"`
	ID     string `glazed.parameter:"id"`
}

func (c *ProjectsDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ProjectsDeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ID == "" {
		return fmt.Errorf("project ID is required")
	}

	url := fmt.Sprintf("%s/api/projects/%s", settings.Server, settings.ID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project %s not found", settings.ID)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	row := types.NewRow(
		types.MRP("id", settings.ID),
		types.MRP("status", "deleted"),
	)

	return gp.AddRow(ctx, row)
}

func NewProjectsDeleteCommand() (*ProjectsDeleteCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ProjectsDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"projects-delete",
			cmds.WithShort("Delete a zine-layout project"),
			cmds.WithLong(`
Delete a project from the zine-layout server by ID.

Examples:
  projects-delete --id prj-12345
  projects-delete --id prj-12345 --server http://localhost:8089
			`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"server",
					parameters.ParameterTypeString,
					parameters.WithDefault("http://localhost:8088"),
					parameters.WithHelp("Server base URL"),
					parameters.WithShortFlag("s"),
				),
				parameters.NewParameterDefinition(
					"id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Project ID (required)"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ProjectsDeleteCommand{}
