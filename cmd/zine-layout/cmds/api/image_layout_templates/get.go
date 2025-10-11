package imagelayouttemplates

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

type GetCommand struct {
	*cmds.CommandDescription
}

type GetSettings struct {
	Server     string `glazed.parameter:"server"`
	TemplateID string `glazed.parameter:"template-id"`
}

func (c *GetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &GetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.TemplateID == "" {
		return fmt.Errorf("template-id is required")
	}

	url := fmt.Sprintf("%s/api/image-layout-templates/%s", settings.Server, settings.TemplateID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch image layout template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("template %s not found", settings.TemplateID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Template struct {
			ID          string         `json:"id"`
			ProjectID   *string        `json:"project_id"`
			Scope       string         `json:"scope"`
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Settings    map[string]any `json:"settings"`
			CreatedAt   time.Time      `json:"created_at"`
			UpdatedAt   time.Time      `json:"updated_at"`
		} `json:"template"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	settingsJSON := ""
	if result.Template.Settings != nil {
		if data, err := json.MarshalIndent(result.Template.Settings, "", "  "); err == nil {
			settingsJSON = string(data)
		}
	}

	projectID := ""
	if result.Template.ProjectID != nil {
		projectID = *result.Template.ProjectID
	}

	row := types.NewRow(
		types.MRP("template_id", result.Template.ID),
		types.MRP("scope", result.Template.Scope),
		types.MRP("project_id", projectID),
		types.MRP("name", result.Template.Name),
		types.MRP("description", result.Template.Description),
		types.MRP("created_at", result.Template.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", result.Template.UpdatedAt.Format(time.RFC3339)),
		types.MRP("settings", settingsJSON),
	)
	return gp.AddRow(ctx, row)
}

func NewGetCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &GetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Retrieve a single image layout template"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"server",
					parameters.ParameterTypeString,
					parameters.WithDefault("http://localhost:8088"),
					parameters.WithHelp("Server base URL"),
					parameters.WithShortFlag("s"),
				),
				parameters.NewParameterDefinition(
					"template-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Template ID to load"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &GetCommand{}
