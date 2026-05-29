package zinescmd

import (
	"context"
	"fmt"

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

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

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
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesSetPagesCommand{}
