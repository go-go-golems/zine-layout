package imagesequences

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
	"github.com/spf13/cobra"
)

type ListCommand struct {
	*cmds.CommandDescription
}

type ListSettings struct {
	Server    string `glazed.parameter:"server"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *ListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}

	url := fmt.Sprintf("%s/api/projects/%s/image-sequences", settings.Server, settings.ProjectID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to list image sequences: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project %s not found", settings.ProjectID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Sequences []struct {
			ID          string    `json:"id"`
			ProjectID   string    `json:"project_id"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"sequences"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, seq := range result.Sequences {
		row := types.NewRow(
			types.MRP("project_id", seq.ProjectID),
			types.MRP("sequence_id", seq.ID),
			types.MRP("name", seq.Name),
			types.MRP("description", seq.Description),
			types.MRP("created_at", seq.CreatedAt.Format(time.RFC3339)),
			types.MRP("updated_at", seq.UpdatedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func NewListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &ListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List image sequences for a project"),
			cmds.WithLong(`
List all image sequences associated with a project.

Examples:
  image-sequences-list --project-id prj-12345
  image-sequences-list --project-id prj-12345 --output json
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
					"project-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Project ID (required)"),
					parameters.WithShortFlag("p"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &ListCommand{}
