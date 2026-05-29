package imagelayouttemplatescmd

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
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/spf13/cobra"
)

type templatesGetCommand struct {
	*cmds.CommandDescription
}

type templatesGetSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	TemplateID string `glazed.parameter:"template-id"`
}

func (c *templatesGetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &templatesGetSettings{}
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

	tpl, err := repos.ImageLayoutTemplates.Get(settings.TemplateID)
	if err != nil {
		return err
	}

	settingsJSON := tpl.SettingsJSON
	if tpl.SettingsJSON != "" {
		var pretty map[string]any
		if err := json.Unmarshal([]byte(tpl.SettingsJSON), &pretty); err == nil {
			if formatted, err := json.MarshalIndent(pretty, "", "  "); err == nil {
				settingsJSON = string(formatted)
			}
		}
	}

	scope := "global"
	projectID := ""
	if tpl.ProjectID != nil && *tpl.ProjectID != "" {
		scope = "project"
		projectID = *tpl.ProjectID
	}

	row := types.NewRow(
		types.MRP("entity", "image_layout_template"),
		types.MRP("template_id", tpl.ID),
		types.MRP("scope", scope),
		types.MRP("project_id", projectID),
		types.MRP("name", tpl.Name),
		types.MRP("description", tpl.Description),
		types.MRP("created_at", tpl.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", tpl.UpdatedAt.Format(time.RFC3339)),
		types.MRP("settings", settingsJSON),
	)
	return gp.AddRow(ctx, row)
}

func newTemplatesGetCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &templatesGetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Show a single image layout template"),
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
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &templatesGetCommand{}
