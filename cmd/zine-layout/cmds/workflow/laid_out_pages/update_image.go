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

type laidOutPagesUpdateImageCommand struct {
	*cmds.CommandDescription
}

type laidOutPagesUpdateImageSettings struct {
	DataRoot       string `glazed.parameter:"data-root"`
	PageID         string `glazed.parameter:"page-id"`
	LaidOutImageID string `glazed.parameter:"laid-out-image-id"`
}

func (c *laidOutPagesUpdateImageCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutPagesUpdateImageSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.PageID == "" {
		return fmt.Errorf("--page-id is required")
	}
	if settings.LaidOutImageID == "" {
		return fmt.Errorf("--laid-out-image-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewPagesService(repos)
	if err := service.UpdatePageImage(settings.PageID, settings.LaidOutImageID); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "laid_out_page"),
		types.MRP("page_id", settings.PageID),
		types.MRP("laid_out_image_id", settings.LaidOutImageID),
		types.MRP("status", "image-updated"),
	)
	return gp.AddRow(ctx, row)
}

func newLaidOutPagesUpdateImageCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &laidOutPagesUpdateImageCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update-image",
			cmds.WithShort("Change the laid-out image for a print page"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("page-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Print page id")),
				parameters.NewParameterDefinition("laid-out-image-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("New laid-out image id")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesUpdateImageCommand{}
