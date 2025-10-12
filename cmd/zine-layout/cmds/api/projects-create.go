package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type ProjectsCreateCommand struct {
	*cmds.CommandDescription
}

type ProjectsCreateSettings struct {
	Server      string `glazed.parameter:"server"`
	Name        string `glazed.parameter:"name"`
	Description string `glazed.parameter:"description"`
}

func (c *ProjectsCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ProjectsCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.Name == "" {
		return fmt.Errorf("project name is required")
	}

	body := map[string]string{
		"name":        settings.Name,
		"description": settings.Description,
	}

	respBytes, err := httpPostJSON(fmt.Sprintf("%s/api/projects", settings.Server), body)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	var result struct {
		Project struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		} `json:"project"`
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	project := result.Project
	row := types.NewRow(
		types.MRP("id", project.ID),
		types.MRP("name", project.Name),
		types.MRP("description", project.Description),
		types.MRP("created_at", project.CreatedAt),
		types.MRP("updated_at", project.UpdatedAt),
		types.MRP("status", "created"),
	)

	return gp.AddRow(ctx, row)
}

func NewProjectsCreateCommand() (*ProjectsCreateCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ProjectsCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"projects-create",
			cmds.WithShort("Create a new zine-layout project"),
			cmds.WithLong(`
Create a new project in the zine-layout server.

Examples:
  projects-create --name "My New Project"
  projects-create --name "Test Project" --server http://localhost:8089
  projects-create --name "Demo" --output json
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
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Project name (required)"),
					parameters.WithShortFlag("n"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Optional project description"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ProjectsCreateCommand{}
