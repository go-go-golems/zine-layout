package workflowcmd

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

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

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

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &pageTemplatesListCommand{}

// --- create command ---

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

	repos, db, err := openRepositories(settings.DataRoot)
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
	)
	return gp.AddRow(ctx, row)
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
					parameters.WithDefault(""),
					parameters.WithHelp("Optional project scope"),
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
					parameters.WithHelp("Inline JSON template body"),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to JSON template body"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &pageTemplatesCreateCommand{}

// --- get command ---

type pageTemplatesGetCommand struct {
	*cmds.CommandDescription
}

type pageTemplatesGetSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
	ID       string `glazed.parameter:"id"`
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
	if settings.ID == "" {
		return fmt.Errorf("--id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	tpl, err := repos.PageTemplates.Get(settings.ID)
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
	return gp.AddRow(ctx, row)
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
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Page template id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &pageTemplatesGetCommand{}

// --- delete command ---

type pageTemplatesDeleteCommand struct {
	*cmds.CommandDescription
}

type pageTemplatesDeleteSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
	ID       string `glazed.parameter:"id"`
}

func (c *pageTemplatesDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &pageTemplatesDeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ID == "" {
		return fmt.Errorf("--id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := repos.PageTemplates.Delete(settings.ID); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "page_template"),
		types.MRP("page_template_id", settings.ID),
		types.MRP("status", "deleted"),
	)
	return gp.AddRow(ctx, row)
}

func newPageTemplatesDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &pageTemplatesDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a page template"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Page template id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &pageTemplatesDeleteCommand{}

func newPageTemplatesCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "page-templates",
		Short: "Manage page templates",
	}

	listCmd, err := newPageTemplatesListCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(listCmd)

	createCmd, err := newPageTemplatesCreateCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(createCmd)

	getCmd, err := newPageTemplatesGetCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(getCmd)

	deleteCmd, err := newPageTemplatesDeleteCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(deleteCmd)

	return root, nil
}
