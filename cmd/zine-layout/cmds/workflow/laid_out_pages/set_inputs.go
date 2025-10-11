package laidoutpagescmd

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

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
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
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesSetInputsCommand{}
