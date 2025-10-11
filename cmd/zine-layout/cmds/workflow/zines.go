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

type zinesCreateCommand struct {
	*cmds.CommandDescription
}

type zinesCreateSettings struct {
	DataRoot    string   `glazed.parameter:"data-root"`
	ProjectID   string   `glazed.parameter:"project-id"`
	Name        string   `glazed.parameter:"name"`
	Description string   `glazed.parameter:"description"`
	Pages       []string `glazed.parameter:"pages"`
}

func (c *zinesCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &zinesCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("--project-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewZinesService(repos)
	zine, pages, err := service.CreateZine(settings.ProjectID, settings.Name, settings.Description, settings.Pages)
	if err != nil {
		return err
	}

	zineRow := types.NewRow(
		types.MRP("entity", "zine"),
		types.MRP("zine_id", zine.ID),
		types.MRP("project_id", zine.ProjectID),
		types.MRP("name", zine.Name),
		types.MRP("description", zine.Description),
		types.MRP("created_at", zine.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", zine.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, zineRow); err != nil {
		return err
	}

	for _, page := range pages {
		row := types.NewRow(
			types.MRP("entity", "zine_page"),
			types.MRP("zine_id", page.ZineID),
			types.MRP("position", page.Position),
			types.MRP("laid_out_page_id", page.LaidOutPageID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newZinesCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &zinesCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a zine"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("project-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Project id")),
				parameters.NewParameterDefinition("name", parameters.ParameterTypeString, parameters.WithDefault(""), parameters.WithHelp("Zine name")),
				parameters.NewParameterDefinition("description", parameters.ParameterTypeString, parameters.WithDefault(""), parameters.WithHelp("Zine description")),
				parameters.NewParameterDefinition("pages", parameters.ParameterTypeStringList, parameters.WithDefault([]string{}), parameters.WithHelp("Ordered laid-out page ids")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesCreateCommand{}

// list command

type zinesListCommand struct {
	*cmds.CommandDescription
}

type zinesListSettings struct {
	DataRoot  string `glazed.parameter:"data-root"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *zinesListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &zinesListSettings{}
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

	zines, err := repos.Zines.ListByProject(settings.ProjectID)
	if err != nil {
		return err
	}

	for _, z := range zines {
		row := types.NewRow(
			types.MRP("entity", "zine"),
			types.MRP("zine_id", z.ID),
			types.MRP("project_id", z.ProjectID),
			types.MRP("name", z.Name),
			types.MRP("updated_at", z.UpdatedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newZinesListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &zinesListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List zines for a project"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("project-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Project id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesListCommand{}

// get command

type zinesGetCommand struct {
	*cmds.CommandDescription
}

type zinesGetSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
	ZineID   string `glazed.parameter:"zine-id"`
}

func (c *zinesGetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &zinesGetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ZineID == "" {
		return fmt.Errorf("--zine-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewZinesService(repos)
	zine, pages, err := service.GetZineWithPages(settings.ZineID)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "zine"),
		types.MRP("zine_id", zine.ID),
		types.MRP("project_id", zine.ProjectID),
		types.MRP("name", zine.Name),
		types.MRP("description", zine.Description),
		types.MRP("updated_at", zine.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	for _, page := range pages {
		r := types.NewRow(
			types.MRP("entity", "zine_page"),
			types.MRP("zine_id", page.ZineID),
			types.MRP("position", page.Position),
			types.MRP("laid_out_page_id", page.LaidOutPageID),
		)
		if err := gp.AddRow(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func newZinesGetCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &zinesGetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Get a zine and its ordered pages"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("zine-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Zine id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesGetCommand{}

// set-pages command

type zinesSetPagesCommand struct {
	*cmds.CommandDescription
}

type zinesSetPagesSettings struct {
	DataRoot string   `glazed.parameter:"data-root"`
	ZineID   string   `glazed.parameter:"zine-id"`
	Pages    []string `glazed.parameter:"pages"`
}

func (c *zinesSetPagesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &zinesSetPagesSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ZineID == "" {
		return fmt.Errorf("--zine-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewZinesService(repos)
	if err := service.UpdateZinePages(settings.ZineID, settings.Pages); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "zine"),
		types.MRP("zine_id", settings.ZineID),
		types.MRP("status", "pages-updated"),
	)
	return gp.AddRow(ctx, row)
}

func newZinesSetPagesCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &zinesSetPagesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"set-pages",
			cmds.WithShort("Replace pages for a zine"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("zine-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Zine id")),
				parameters.NewParameterDefinition("pages", parameters.ParameterTypeStringList, parameters.WithDefault([]string{}), parameters.WithHelp("Ordered laid-out page ids")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesSetPagesCommand{}

// delete command

type zinesDeleteCommand struct {
	*cmds.CommandDescription
}

type zinesDeleteSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
	ZineID   string `glazed.parameter:"zine-id"`
}

func (c *zinesDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &zinesDeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ZineID == "" {
		return fmt.Errorf("--zine-id is required")
	}

	repos, db, err := openRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewZinesService(repos)
	if err := service.DeleteZine(settings.ZineID); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "zine"),
		types.MRP("zine_id", settings.ZineID),
		types.MRP("status", "deleted"),
	)
	return gp.AddRow(ctx, row)
}

func newZinesDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &zinesDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a zine"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("zine-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Zine id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesDeleteCommand{}

func newZinesCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "zines",
		Short: "Manage zines",
	}

	builders := []func() (*cobra.Command, error){
		newZinesCreateCommand,
		newZinesListCommand,
		newZinesGetCommand,
		newZinesSetPagesCommand,
		newZinesDeleteCommand,
	}

	for _, builder := range builders {
		cmd, err := builder()
		if err != nil {
			return nil, err
		}
		root.AddCommand(cmd)
	}

	return root, nil
}
