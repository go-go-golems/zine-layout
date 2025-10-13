package assetscmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/go-go-golems/zine-layout/pkg/projects"
	"github.com/go-go-golems/zine-layout/pkg/repo"
	"github.com/spf13/cobra"
)

type assetsCreateCommand struct {
	*cmds.CommandDescription
}

type assetsCreateSettings struct {
	DataRoot  string   `glazed.parameter:"data-root"`
	ProjectID string   `glazed.parameter:"project-id"`
	Files     []string `glazed.parameter:"file"`
}

func (c *assetsCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &assetsCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("--project-id is required")
	}
	if len(settings.Files) == 0 {
		return fmt.Errorf("provide at least one --file")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := repos.Projects.Get(settings.ProjectID); err != nil {
		return fmt.Errorf("lookup project: %w", err)
	}

	root := workflowshared.ProjectsRoot(settings.DataRoot)

	for _, file := range settings.Files {
		clean := strings.TrimSpace(file)
		if clean == "" {
			continue
		}
		saved, err := projects.SavePNGImageFromPath(root, settings.ProjectID, clean)
		if err != nil {
			return fmt.Errorf("ingest %s: %w", file, err)
		}

		relPath := filepath.ToSlash(filepath.Join("projects", settings.ProjectID, "images", saved.Filename))
		meta := map[string]any{
			"width":  saved.Width,
			"height": saved.Height,
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("encode metadata: %w", err)
		}

		uploadedAt := time.Now().UTC()
		asset := &repo.Asset{
			ProjectID:    settings.ProjectID,
			Filename:     saved.Filename,
			RelPath:      relPath,
			ContentType:  "image/png",
			Bytes:        saved.Bytes,
			Width:        saved.Width,
			Height:       saved.Height,
			MetadataJSON: string(metaJSON),
			UploadedAt:   uploadedAt,
		}
		if err := repos.Assets.Create(asset); err != nil {
			return fmt.Errorf("persist asset: %w", err)
		}

		row := types.NewRow(
			types.MRP("entity", "asset"),
			types.MRP("asset_id", asset.ID),
			types.MRP("project_id", asset.ProjectID),
			types.MRP("source", clean),
			types.MRP("filename", asset.Filename),
			types.MRP("rel_path", asset.RelPath),
			types.MRP("width", asset.Width),
			types.MRP("height", asset.Height),
			types.MRP("bytes", asset.Bytes),
			types.MRP("uploaded_at", asset.UploadedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func newAssetsCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &assetsCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Import PNG assets into a project"),
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
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeStringList,
					parameters.WithRequired(true),
					parameters.WithHelp("PNG file(s) to copy into the project"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &assetsCreateCommand{}
