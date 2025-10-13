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

type laidOutImagesListCommand struct {
	*cmds.CommandDescription
}

type laidOutImagesListSettings struct {
	DataRoot  string `glazed.parameter:"data-root"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *laidOutImagesListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutImagesListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("--project-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	items, err := repos.LaidOutImages.ListByProject(settings.ProjectID)
	if err != nil {
		return err
	}

	for _, item := range items {
		row := types.NewRow(
			types.MRP("entity", "laid_out_image"),
			types.MRP("laid_out_image_id", item.ID),
			types.MRP("project_id", item.ProjectID),
			types.MRP("asset_id", item.AssetID),
			types.MRP("template_id", item.TemplateID),
			types.MRP("updated_at", item.UpdatedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newLaidOutImagesListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &laidOutImagesListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List laid-out images for a project"),
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
					parameters.WithRequired(true),
					parameters.WithHelp("Project identifier"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutImagesListCommand{}
