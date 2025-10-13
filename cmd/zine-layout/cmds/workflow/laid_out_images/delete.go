package laidoutimagescmd

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
	"github.com/spf13/cobra"
)

type laidOutImagesDeleteCommand struct {
	*cmds.CommandDescription
}

type laidOutImagesDeleteSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
	ID       string `glazed.parameter:"id"`
}

func (c *laidOutImagesDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutImagesDeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ID == "" {
		return fmt.Errorf("--id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := repos.LaidOutImages.Delete(settings.ID); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "laid_out_image"),
		types.MRP("laid_out_image_id", settings.ID),
		types.MRP("status", "deleted"),
		types.MRP("deleted_at", time.Now().UTC().Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newLaidOutImagesDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &laidOutImagesDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a laid-out image"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Laid-out image identifier"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutImagesDeleteCommand{}
