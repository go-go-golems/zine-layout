package imagesequences

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type CreateCommand struct {
	*cmds.CommandDescription
}

type CreateSettings struct {
	Server      string `glazed.parameter:"server"`
	ProjectID   string `glazed.parameter:"project-id"`
	Name        string `glazed.parameter:"name"`
	Description string `glazed.parameter:"description"`
}

func (c *CreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &CreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if settings.Name == "" {
		return fmt.Errorf("sequence name is required")
	}

	body := map[string]string{
		"name":        settings.Name,
		"description": settings.Description,
	}

	url := fmt.Sprintf("%s/api/projects/%s/image-sequences", settings.Server, settings.ProjectID)
	respBytes, err := httpPostJSON(url, body)
	if err != nil {
		return fmt.Errorf("failed to create sequence: %w", err)
	}

	var result struct {
		Sequence struct {
			ID          string    `json:"id"`
			ProjectID   string    `json:"project_id"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"sequence"`
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	seq := result.Sequence
	row := types.NewRow(
		types.MRP("project_id", seq.ProjectID),
		types.MRP("sequence_id", seq.ID),
		types.MRP("name", seq.Name),
		types.MRP("description", seq.Description),
		types.MRP("created_at", seq.CreatedAt.Format(time.RFC3339)),
		types.MRP("status", "created"),
	)

	return gp.AddRow(ctx, row)
}

func NewCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &CreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a new image sequence for a project"),
			cmds.WithLong(`
Create a new image sequence associated with a project.

Examples:
  image-sequences-create --project-id prj-123 --name "Contact Sheet"
  image-sequences-create --project-id prj-123 --name "My Seq" --description "Favorites"
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
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Sequence name (required)"),
					parameters.WithShortFlag("n"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Optional description"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &CreateCommand{}
