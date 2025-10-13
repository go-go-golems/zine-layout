package assetscmd

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

type assetsListCommand struct {
	*cmds.CommandDescription
}

type assetsListSettings struct {
	DataRoot  string `glazed.parameter:"data-root"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *assetsListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &assetsListSettings{}
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

	assets, err := repos.Assets.ListByProject(settings.ProjectID)
	if err != nil {
		return err
	}

	for _, asset := range assets {
		row := types.NewRow(
			types.MRP("entity", "asset"),
			types.MRP("asset_id", asset.ID),
			types.MRP("project_id", asset.ProjectID),
			types.MRP("filename", asset.Filename),
			types.MRP("rel_path", asset.RelPath),
			types.MRP("width", asset.Width),
			types.MRP("height", asset.Height),
			types.MRP("bytes", asset.Bytes),
			types.MRP("content_type", asset.ContentType),
			types.MRP("uploaded_at", asset.UploadedAt.Format(time.RFC3339)),
			types.MRP("metadata_json", asset.MetadataJSON),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newAssetsListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &assetsListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List assets for a project"),
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

var _ cmds.GlazeCommand = &assetsListCommand{}
