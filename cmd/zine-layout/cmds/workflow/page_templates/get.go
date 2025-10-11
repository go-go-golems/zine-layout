package pagetemplatescmd

import (
	"context"
	"fmt"
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

type pageTemplatesGetCommand struct {
	*cmds.CommandDescription
}

type pageTemplatesGetSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	TemplateID string `glazed.parameter:"page-template-id"`
}

func (c *pageTemplatesGetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &pageTemplatesGetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if strings.TrimSpace(settings.TemplateID) == "" {
		return fmt.Errorf("page-template-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	tpl, err := repos.PageTemplates.Get(settings.TemplateID)
	if err != nil {
		return err
	}

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
		types.MRP("template_json", tpl.TemplateJSON),
		types.MRP("created_at", tpl.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", tpl.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	return nil
}

func newPageTemplatesGetCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &pageTemplatesGetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Get a page template"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"page-template-id",
					parameters.ParameterTypeString,
					parameters.WithHelp("Template identifier"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &pageTemplatesGetCommand{}
