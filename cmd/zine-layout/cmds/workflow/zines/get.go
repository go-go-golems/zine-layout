package zinescmd

import (
	"context"
	"fmt"
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

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
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
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesGetCommand{}
