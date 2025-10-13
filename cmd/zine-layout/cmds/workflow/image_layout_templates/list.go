package imagelayouttemplatescmd

import (
	"context"
	"encoding/json"
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

type templatesListCommand struct {
	*cmds.CommandDescription
}

type templatesListSettings struct {
	DataRoot  string `glazed.parameter:"data-root"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *templatesListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &templatesListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	var templates []*repo.ImageLayoutTemplate
	if settings.ProjectID != "" {
		projectTemplates, err := repos.ImageLayoutTemplates.ListByProject(settings.ProjectID)
		if err != nil {
			return err
		}
		globalTemplates, err := repos.ImageLayoutTemplates.ListGlobal()
		if err != nil {
			return err
		}
		templates = append(projectTemplates, globalTemplates...)
	} else {
		templates, err = repos.ImageLayoutTemplates.ListGlobal()
		if err != nil {
			return err
		}
	}

	for _, tpl := range templates {
		settingsJSON := ""
		if tpl.SettingsJSON != "" {
			var anyMap map[string]any
			if err := json.Unmarshal([]byte(tpl.SettingsJSON), &anyMap); err == nil {
				if pretty, err := json.MarshalIndent(anyMap, "", "  "); err == nil {
					settingsJSON = string(pretty)
				}
			} else {
				settingsJSON = tpl.SettingsJSON
			}
		}

		projectID := ""
		scope := "global"
		if tpl.ProjectID != nil && *tpl.ProjectID != "" {
			projectID = *tpl.ProjectID
			scope = "project"
		}

		row := types.NewRow(
			types.MRP("entity", "image_layout_template"),
			types.MRP("template_id", tpl.ID),
			types.MRP("scope", scope),
			types.MRP("project_id", projectID),
			types.MRP("name", tpl.Name),
			types.MRP("description", tpl.Description),
			types.MRP("updated_at", tpl.UpdatedAt.Format(time.RFC3339)),
			types.MRP("settings", settingsJSON),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newTemplatesListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &templatesListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List image layout templates"),
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
					parameters.WithHelp("Project identifier to include scoped templates"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &templatesListCommand{}
