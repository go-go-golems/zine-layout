package workflowcmd

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
	"github.com/go-go-golems/zine-layout/pkg/services"
	"github.com/spf13/cobra"
)

type laidOutPagesCreateCommand struct {
	*cmds.CommandDescription
}

type laidOutPagesCreateSettings struct {
	DataRoot   string   `glazed.parameter:"data-root"`
	ProjectID  string   `glazed.parameter:"project-id"`
	TemplateID string   `glazed.parameter:"template-id"`
	Inputs     []string `glazed.parameter:"inputs"`
}

func (c *laidOutPagesCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutPagesCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("--project-id is required")
	}
	if settings.TemplateID == "" {
		return fmt.Errorf("--template-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewPagesService(repos)
	page, inputs, err := service.CreatePage(settings.ProjectID, settings.TemplateID, settings.Inputs)
	if err != nil {
		return err
	}

	pageRow := types.NewRow(
		types.MRP("entity", "laid_out_page"),
		types.MRP("page_id", page.ID),
		types.MRP("project_id", page.ProjectID),
		types.MRP("page_template_id", page.PageTemplateID),
		types.MRP("created_at", page.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", page.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, pageRow); err != nil {
		return err
	}

	for _, input := range inputs {
		row := types.NewRow(
			types.MRP("entity", "laid_out_page_input"),
			types.MRP("page_id", input.PageID),
			types.MRP("position", input.InputIndex),
			types.MRP("laid_out_image_id", input.LaidOutImageID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newLaidOutPagesCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &laidOutPagesCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a laid-out page"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("project-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Project id")),
				parameters.NewParameterDefinition("template-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Page template id")),
				parameters.NewParameterDefinition("inputs", parameters.ParameterTypeStringList, parameters.WithDefault([]string{}), parameters.WithHelp("Ordered laid-out image ids")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesCreateCommand{}

// list command

type laidOutPagesListCommand struct {
	*cmds.CommandDescription
}

type laidOutPagesListSettings struct {
	DataRoot  string `glazed.parameter:"data-root"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *laidOutPagesListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutPagesListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if strings.TrimSpace(settings.ProjectID) == "" {
		return fmt.Errorf("--project-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	pages, err := repos.LaidOutPages.ListByProject(settings.ProjectID)
	if err != nil {
		return err
	}

	for _, page := range pages {
		row := types.NewRow(
			types.MRP("entity", "laid_out_page"),
			types.MRP("page_id", page.ID),
			types.MRP("project_id", page.ProjectID),
			types.MRP("page_template_id", page.PageTemplateID),
			types.MRP("updated_at", page.UpdatedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newLaidOutPagesListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &laidOutPagesListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List laid-out pages for a project"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("project-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Project id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesListCommand{}

// get command

type laidOutPagesGetCommand struct {
	*cmds.CommandDescription
}

type laidOutPagesGetSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
	PageID   string `glazed.parameter:"page-id"`
}

func (c *laidOutPagesGetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutPagesGetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.PageID == "" {
		return fmt.Errorf("--page-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewPagesService(repos)
	page, inputs, err := service.GetPageWithInputs(settings.PageID)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "laid_out_page"),
		types.MRP("page_id", page.ID),
		types.MRP("project_id", page.ProjectID),
		types.MRP("page_template_id", page.PageTemplateID),
		types.MRP("updated_at", page.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	for _, input := range inputs {
		r := types.NewRow(
			types.MRP("entity", "laid_out_page_input"),
			types.MRP("page_id", input.PageID),
			types.MRP("position", input.InputIndex),
			types.MRP("laid_out_image_id", input.LaidOutImageID),
		)
		if err := gp.AddRow(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func newLaidOutPagesGetCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &laidOutPagesGetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Get a laid-out page and its inputs"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("page-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Laid-out page id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesGetCommand{}

// set-inputs command

type laidOutPagesSetInputsCommand struct {
	*cmds.CommandDescription
}

type laidOutPagesSetInputsSettings struct {
	DataRoot string   `glazed.parameter:"data-root"`
	PageID   string   `glazed.parameter:"page-id"`
	Inputs   []string `glazed.parameter:"inputs"`
}

func (c *laidOutPagesSetInputsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutPagesSetInputsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.PageID == "" {
		return fmt.Errorf("--page-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewPagesService(repos)
	if err := service.UpdatePageInputs(settings.PageID, settings.Inputs); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "laid_out_page"),
		types.MRP("page_id", settings.PageID),
		types.MRP("status", "inputs-updated"),
	)
	return gp.AddRow(ctx, row)
}

func newLaidOutPagesSetInputsCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &laidOutPagesSetInputsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"set-inputs",
			cmds.WithShort("Replace inputs for a laid-out page"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("page-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Laid-out page id")),
				parameters.NewParameterDefinition("inputs", parameters.ParameterTypeStringList, parameters.WithDefault([]string{}), parameters.WithHelp("Ordered laid-out image ids")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesSetInputsCommand{}

// delete command

type laidOutPagesDeleteCommand struct {
	*cmds.CommandDescription
}

type laidOutPagesDeleteSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
	PageID   string `glazed.parameter:"page-id"`
}

func (c *laidOutPagesDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutPagesDeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.PageID == "" {
		return fmt.Errorf("--page-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewPagesService(repos)
	if err := service.DeletePage(settings.PageID); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "laid_out_page"),
		types.MRP("page_id", settings.PageID),
		types.MRP("status", "deleted"),
	)
	return gp.AddRow(ctx, row)
}

func newLaidOutPagesDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &laidOutPagesDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a laid-out page"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("page-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Laid-out page id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesDeleteCommand{}

func newLaidOutPagesCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "laid-out-pages",
		Short: "Manage laid-out pages",
	}

	cmdsList := []func() (*cobra.Command, error){
		newLaidOutPagesCreateCommand,
		newLaidOutPagesListCommand,
		newLaidOutPagesGetCommand,
		newLaidOutPagesSetInputsCommand,
		newLaidOutPagesDeleteCommand,
	}
	for _, builder := range cmdsList {
		sub, err := builder()
		if err != nil {
			return nil, err
		}
		root.AddCommand(sub)
	}

	return root, nil
}
