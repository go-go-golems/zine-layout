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
	"github.com/go-go-golems/zine-layout/pkg/repo"
	"github.com/spf13/cobra"
)

type templatesCreateCommand struct {
	*cmds.CommandDescription
}

type templatesCreateSettings struct {
	DataRoot     string `glazed.parameter:"data-root"`
	ProjectID    string `glazed.parameter:"project-id"`
	Name         string `glazed.parameter:"name"`
	Description  string `glazed.parameter:"description"`
	SettingsJSON string `glazed.parameter:"settings-json"`
	File         string `glazed.parameter:"file"`
}

func (c *templatesCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &templatesCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	payload := strings.TrimSpace(settings.SettingsJSON)
	if payload == "" && strings.TrimSpace(settings.File) == "" {
		return fmt.Errorf("provide --settings-json or --file")
	}
	if payload == "" && settings.File != "" {
		data, err := os.ReadFile(settings.File)
		if err != nil {
			return fmt.Errorf("read settings file: %w", err)
		}
		payload = string(data)
	}

	var settingsMap map[string]any
	if err := json.Unmarshal([]byte(payload), &settingsMap); err != nil {
		return fmt.Errorf("parse template settings: %w", err)
	}
	normalizedBytes, err := json.Marshal(settingsMap)
	if err != nil {
		return fmt.Errorf("normalize settings: %w", err)
	}

	name := strings.TrimSpace(settings.Name)
	if name == "" {
		name = "Untitled Template"
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	template := &repo.ImageLayoutTemplate{
		Name:         name,
		Description:  strings.TrimSpace(settings.Description),
		SettingsJSON: string(normalizedBytes),
	}
	if strings.TrimSpace(settings.ProjectID) != "" {
		projectID := strings.TrimSpace(settings.ProjectID)
		template.ProjectID = &projectID
	}

	if err := repos.ImageLayoutTemplates.Create(template); err != nil {
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
		types.MRP("created_at", template.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", template.UpdatedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newTemplatesCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &templatesCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create an image layout template"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"project-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Project identifier (optional; omit for global template)"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Template name"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Template description"),
				),
				parameters.NewParameterDefinition(
					"settings-json",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Template settings JSON payload"),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to JSON file containing template settings"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &templatesCreateCommand{}
