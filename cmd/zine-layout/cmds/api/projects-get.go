package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type ProjectsGetCommand struct {
	*cmds.CommandDescription
}

type ProjectsGetSettings struct {
	Server string `glazed.parameter:"server"`
	ID     string `glazed.parameter:"id"`
}

func (c *ProjectsGetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ProjectsGetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ID == "" {
		return fmt.Errorf("project ID is required")
	}

	url := fmt.Sprintf("%s/api/projects/%s", settings.Server, settings.ID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project %s not found", settings.ID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Project struct {
			ID          string    `json:"id"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"project"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	project := result.Project

	row := types.NewRow(
		types.MRP("id", project.ID),
		types.MRP("name", project.Name),
		types.MRP("description", project.Description),
		types.MRP("created_at", project.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", project.UpdatedAt.Format(time.RFC3339)),
	)

	return gp.AddRow(ctx, row)
}

func NewProjectsGetCommand() (*ProjectsGetCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ProjectsGetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"projects-get",
			cmds.WithShort("Get a specific zine-layout project"),
			cmds.WithLong(`
Get details of a specific project by ID from the zine-layout server.

Examples:
  projects-get --id prj-12345
  projects-get --id prj-12345 --output json
  projects-get --id prj-12345 --server http://localhost:8089
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
var _ cmds.GlazeCommand = &ProjectsGetCommand{}
