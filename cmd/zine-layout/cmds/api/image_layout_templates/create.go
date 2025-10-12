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

type CreateCommand struct {
	*cmds.CommandDescription
}

type CreateSettings struct {
	Server       string `glazed.parameter:"server"`
	ProjectID    string `glazed.parameter:"project-id"`
	Name         string `glazed.parameter:"name"`
	Description  string `glazed.parameter:"description"`
	SettingsFile string `glazed.parameter:"settings-file"`
	SettingsJSON string `glazed.parameter:"settings-json"`
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

	if strings.TrimSpace(settings.Name) == "" {
		settings.Name = "Untitled Template"
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
	if len(settingsMap) == 0 {
		return fmt.Errorf("template settings are required (provide --settings-file or --settings-json)")
	}

	payload := map[string]any{
		"name":        settings.Name,
		"description": settings.Description,
		"settings":    settingsMap,
	}

	var (
		data []byte
		err  error
	)
	if settings.ProjectID != "" {
		url := fmt.Sprintf("%s/api/projects/%s/image-layout-templates", settings.Server, settings.ProjectID)
		data, err = httpPostJSON(url, payload)
	} else {
		url := fmt.Sprintf("%s/api/image-layout-templates", settings.Server)
		data, err = httpPostJSON(url, payload)
	}
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
		types.MRP("created_at", result.Template.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", result.Template.UpdatedAt.Format(time.RFC3339)),
		types.MRP("settings", settingsJSON),
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
			cmds.WithShort("Create a new image layout template"),
			cmds.WithLong(`
Create a new image layout template. Provide --project-id to scope the template to a project, or omit it to create a global template. Settings must be provided via --settings-json or --settings-file.`),
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
					parameters.WithHelp("Project ID (optional)"),
					parameters.WithShortFlag("p"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Template name"),
					parameters.WithShortFlag("n"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Template description"),
					parameters.WithShortFlag("d"),
				),
				parameters.NewParameterDefinition(
					"settings-file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to JSON file containing template settings"),
				),
				parameters.NewParameterDefinition(
					"settings-json",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Raw JSON string describing template settings"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &CreateCommand{}
