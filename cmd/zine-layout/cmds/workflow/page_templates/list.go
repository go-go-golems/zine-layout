package pagetemplatescmd

import (
	"context"
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

type pageTemplatesListCommand struct {
	*cmds.CommandDescription
}

type pageTemplatesListSettings struct {
	DataRoot      string `glazed.parameter:"data-root"`
	ProjectID     string `glazed.parameter:"project-id"`
	IncludeGlobal bool   `glazed.parameter:"include-global"`
}

func (c *pageTemplatesListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &pageTemplatesListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	var templates []*repo.PageTemplate
	if settings.ProjectID == "" {
		templates, err = repos.PageTemplates.ListGlobal()
		if err != nil {
			return err
		}
	} else {
		templates, err = repos.PageTemplates.ListByProject(settings.ProjectID)
		if err != nil {
			return err
		}
		if settings.IncludeGlobal {
			global, err := repos.PageTemplates.ListGlobal()
			if err != nil {
				return err
			}
			templates = append(templates, global...)
		}
	}

	for _, tpl := range templates {
		scope := "project"
		projectID := ""
		if tpl.ProjectID == nil || *tpl.ProjectID == "" {
			scope = "global"
		} else {
			projectID = *tpl.ProjectID
		}
		row := types.NewRow(
			types.MRP("entity", "page_template"),
			types.MRP("page_template_id", tpl.ID),
			types.MRP("project_id", projectID),
			types.MRP("scope", scope),
			types.MRP("name", tpl.Name),
			types.MRP("description", tpl.Description),
			types.MRP("created_at", tpl.CreatedAt.Format(time.RFC3339)),
			types.MRP("updated_at", tpl.UpdatedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newPageTemplatesListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &pageTemplatesListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List page templates"),
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
					parameters.WithHelp("Project filter (optional)"),
				),
				parameters.NewParameterDefinition(
					"include-global",
					parameters.ParameterTypeBool,
					parameters.WithDefault(true),
					parameters.WithHelp("Include global templates when filtering by project"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &pageTemplatesListCommand{}
