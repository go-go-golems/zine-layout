package layoutsequences

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
		return fmt.Errorf("project-id is required")
	}

	name := strings.TrimSpace(settings.Name)
	if name == "" {
		name = "Untitled Layout Sequence"
	}

	payload := map[string]any{
		"name":        name,
		"description": strings.TrimSpace(settings.Description),
	}

	url := fmt.Sprintf("%s/api/projects/%s/layout-sequences", settings.Server, settings.ProjectID)
	data, err := httpPostJSON(url, payload)
	if err != nil {
		return err
	}

	var result struct {
		LayoutSequence struct {
			ID          string    `json:"id"`
			ProjectID   string    `json:"project_id"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"layout_sequence"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	row := types.NewRow(
		types.MRP("layout_sequence_id", result.LayoutSequence.ID),
		types.MRP("project_id", result.LayoutSequence.ProjectID),
		types.MRP("name", result.LayoutSequence.Name),
		types.MRP("description", result.LayoutSequence.Description),
		types.MRP("created_at", result.LayoutSequence.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", result.LayoutSequence.UpdatedAt.Format(time.RFC3339)),
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
			cmds.WithShort("Create a layout sequence"),
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
					parameters.WithRequired(true),
					parameters.WithHelp("Project ID"),
					parameters.WithShortFlag("p"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Layout sequence name"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Layout sequence description"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &CreateCommand{}
