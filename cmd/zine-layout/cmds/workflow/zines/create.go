package zinescmd

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
	if strings.TrimSpace(settings.ProjectID) == "" {
		return fmt.Errorf("--project-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

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
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesCreateCommand{}
