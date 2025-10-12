package imagelayouttemplates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

type UpdateCommand struct {
	*cmds.CommandDescription
}

type UpdateSettings struct {
	Server       string `glazed.parameter:"server"`
	TemplateID   string `glazed.parameter:"template-id"`
	Name         string `glazed.parameter:"name"`
	Description  string `glazed.parameter:"description"`
	SettingsFile string `glazed.parameter:"settings-file"`
	SettingsJSON string `glazed.parameter:"settings-json"`
}

func (c *UpdateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &UpdateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.TemplateID == "" {
		return fmt.Errorf("template-id is required")
	}

	payload := map[string]any{}
	if strings.TrimSpace(settings.Name) != "" {
		payload["name"] = settings.Name
	}
	if strings.TrimSpace(settings.Description) != "" {
		payload["description"] = settings.Description
	}

	settingsMap := map[string]any{}
	if strings.TrimSpace(settings.SettingsJSON) != "" {
		if err := json.Unmarshal([]byte(settings.SettingsJSON), &settingsMap); err != nil {
			return fmt.Errorf("invalid --settings-json payload: %w", err)
		}
	}
	if settings.SettingsFile != "" {
		data, err := os.ReadFile(settings.SettingsFile)
		if err != nil {
			return fmt.Errorf("failed to read settings file: %w", err)
		}
		if err := json.Unmarshal(data, &settingsMap); err != nil {
			return fmt.Errorf("invalid JSON in settings file: %w", err)
		}
	}
	if len(settingsMap) > 0 {
		payload["settings"] = settingsMap
	}

	if len(payload) == 0 {
		return fmt.Errorf("no updates specified")
	}

	url := fmt.Sprintf("%s/api/image-layout-templates/%s", settings.Server, settings.TemplateID)
	data, err := httpPutJSON(url, payload)
	if err != nil {
		return err
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
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	projectID := ""
	if result.Template.ProjectID != nil {
		projectID = *result.Template.ProjectID
	}
	settingsJSON := ""
	if result.Template.Settings != nil {
		if raw, err := json.MarshalIndent(result.Template.Settings, "", "  "); err == nil {
			settingsJSON = string(raw)
		}
	}

	row := types.NewRow(
		types.MRP("template_id", result.Template.ID),
		types.MRP("scope", result.Template.Scope),
		types.MRP("project_id", projectID),
		types.MRP("name", result.Template.Name),
		types.MRP("description", result.Template.Description),
		types.MRP("updated_at", result.Template.UpdatedAt.Format(time.RFC3339)),
		types.MRP("settings", settingsJSON),
	)
	return gp.AddRow(ctx, row)
}

func NewUpdateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &UpdateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Update an existing image layout template"),
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
					parameters.WithHelp("Template ID to update"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New name (optional)"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New description (optional)"),
				),
				parameters.NewParameterDefinition(
					"settings-file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to JSON file with replacement settings"),
				),
				parameters.NewParameterDefinition(
					"settings-json",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Raw JSON string for replacement settings"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &UpdateCommand{}
