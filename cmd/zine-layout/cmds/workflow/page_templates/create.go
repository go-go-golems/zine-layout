package pagetemplatescmd

import (
	"context"
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

type pageTemplatesCreateCommand struct {
	*cmds.CommandDescription
}

type pageTemplatesCreateSettings struct {
	DataRoot     string `glazed.parameter:"data-root"`
	ProjectID    string `glazed.parameter:"project-id"`
	Name         string `glazed.parameter:"name"`
	Description  string `glazed.parameter:"description"`
	TemplateJSON string `glazed.parameter:"template-json"`
	File         string `glazed.parameter:"file"`
}

func (c *pageTemplatesCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &pageTemplatesCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	templateJSON := strings.TrimSpace(settings.TemplateJSON)
	if templateJSON == "" && strings.TrimSpace(settings.File) == "" {
		return fmt.Errorf("provide --template-json or --file")
	}
	if templateJSON == "" && settings.File != "" {
		data, err := os.ReadFile(settings.File)
		if err != nil {
			return fmt.Errorf("read template file: %w", err)
		}
		templateJSON = string(data)
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	tpl := &repo.PageTemplate{
		Name:         strings.TrimSpace(settings.Name),
		Description:  strings.TrimSpace(settings.Description),
		TemplateJSON: templateJSON,
	}
	if tpl.Name == "" {
		tpl.Name = "Untitled Page Template"
	}
	if strings.TrimSpace(settings.ProjectID) != "" {
		pid := strings.TrimSpace(settings.ProjectID)
		tpl.ProjectID = &pid
	}
	if err := repos.PageTemplates.Create(tpl); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "page_template"),
		types.MRP("page_template_id", tpl.ID),
		types.MRP("project_id", func() string {
			if tpl.ProjectID == nil {
				return ""
			}
			return *tpl.ProjectID
		}()),
		types.MRP("name", tpl.Name),
		types.MRP("description", tpl.Description),
		types.MRP("created_at", tpl.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", tpl.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	return nil
}

func newPageTemplatesCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &pageTemplatesCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a page template"),
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
					parameters.WithHelp("Project identifier (optional)"),
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
					"template-json",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Template JSON body"),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to template JSON file"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &pageTemplatesCreateCommand{}
