package assetscmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

type assetsDeleteCommand struct {
	*cmds.CommandDescription
}

type assetsDeleteSettings struct {
	DataRoot string `glazed.parameter:"data-root"`
	AssetID  string `glazed.parameter:"asset-id"`
}

func (c *assetsDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &assetsDeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.AssetID == "" {
		return fmt.Errorf("--asset-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	asset, err := repos.Assets.Get(settings.AssetID)
	if err != nil {
		return err
	}

	dataRoot := settings.DataRoot
	if dataRoot == "" {
		dataRoot = "./data"
	}
	assetPath := filepath.Join(dataRoot, filepath.FromSlash(asset.RelPath))
	if err := os.Remove(assetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete asset file: %w", err)
	}

	if err := repos.Assets.Delete(settings.AssetID); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "asset"),
		types.MRP("asset_id", asset.ID),
		types.MRP("project_id", asset.ProjectID),
		types.MRP("status", "deleted"),
		types.MRP("deleted_at", time.Now().UTC().Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newAssetsDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &assetsDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete an asset and its stored PNG"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"asset-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Asset identifier"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &assetsDeleteCommand{}
