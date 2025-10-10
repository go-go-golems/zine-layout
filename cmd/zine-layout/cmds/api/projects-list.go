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

type ProjectsListCommand struct {
	*cmds.CommandDescription
}

type ProjectsListSettings struct {
	Server string `glazed.parameter:"server"`
}

func (c *ProjectsListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ProjectsListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/projects", settings.Server)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

var result struct {
	Projects []struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	} `json:"projects"`
}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, project := range result.Projects {
		row := types.NewRow(
			types.MRP("id", project.ID),
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

func NewProjectsListCommand() (*ProjectsListCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ProjectsListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"projects-list",
			cmds.WithShort("List all zine-layout projects"),
			cmds.WithLong(`
List all projects from the zine-layout server.

Examples:
  projects-list
  projects-list --server http://localhost:8089
  projects-list --output json
  projects-list --output csv --fields id,name,image_count
			`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"server",
					parameters.ParameterTypeString,
					parameters.WithDefault("http://localhost:8088"),
					parameters.WithHelp("Server base URL"),
					parameters.WithShortFlag("s"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ProjectsListCommand{}
