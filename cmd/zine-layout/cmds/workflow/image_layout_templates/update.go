package imagelayouttemplatescmd

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
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/spf13/cobra"
)

type templatesUpdateCommand struct {
	*cmds.CommandDescription
}

type templatesUpdateSettings struct {
	DataRoot     string `glazed.parameter:"data-root"`
	TemplateID   string `glazed.parameter:"template-id"`
	Name         string `glazed.parameter:"name"`
	Description  string `glazed.parameter:"description"`
	ProjectID    string `glazed.parameter:"project-id"`
	ClearProject bool   `glazed.parameter:"clear-project"`
	SettingsJSON string `glazed.parameter:"settings-json"`
	File         string `glazed.parameter:"file"`
}

func (c *templatesUpdateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &templatesUpdateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.TemplateID == "" {
		return fmt.Errorf("--template-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	template, err := repos.ImageLayoutTemplates.Get(settings.TemplateID)
	if err != nil {
		return err
	}

	if trimmed := strings.TrimSpace(settings.Name); trimmed != "" {
		template.Name = trimmed
	}
	if desc := strings.TrimSpace(settings.Description); desc != "" {
		template.Description = desc
	}

	if settings.ClearProject {
		template.ProjectID = nil
	} else if trimmed := strings.TrimSpace(settings.ProjectID); trimmed != "" {
		template.ProjectID = &trimmed
	}

	payload := strings.TrimSpace(settings.SettingsJSON)
	if payload == "" && strings.TrimSpace(settings.File) != "" {
		data, err := os.ReadFile(settings.File)
		if err != nil {
			return fmt.Errorf("read settings file: %w", err)
		}
		payload = string(data)
	}
	if payload != "" {
		var settingsMap map[string]any
		if err := json.Unmarshal([]byte(payload), &settingsMap); err != nil {
			return fmt.Errorf("parse template settings: %w", err)
		}
		normalized, err := json.Marshal(settingsMap)
		if err != nil {
			return fmt.Errorf("normalize settings: %w", err)
		}
		template.SettingsJSON = string(normalized)
	}

	template.UpdatedAt = time.Now().UTC()
	if err := repos.ImageLayoutTemplates.Update(template); err != nil {
		return err
	}

	scope := "global"
	projectID := ""
	if template.ProjectID != nil && *template.ProjectID != "" {
		scope = "project"
		projectID = *template.ProjectID
	}

	row := types.NewRow(
		types.MRP("entity", "image_layout_template"),
		types.MRP("template_id", template.ID),
		types.MRP("scope", scope),
		types.MRP("project_id", projectID),
		types.MRP("name", template.Name),
		types.MRP("description", template.Description),
		types.MRP("updated_at", template.UpdatedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newTemplatesUpdateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &templatesUpdateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Update an image layout template"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"template-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Template identifier"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New template name"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New template description"),
				),
				parameters.NewParameterDefinition(
					"project-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Set or change project scope"),
				),
				parameters.NewParameterDefinition(
					"clear-project",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Remove project scope and make template global"),
				),
				parameters.NewParameterDefinition(
					"settings-json",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Replacement settings JSON"),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to JSON file with replacement settings"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &templatesUpdateCommand{}
