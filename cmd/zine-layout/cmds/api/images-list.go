package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type ImagesListCommand struct {
	*cmds.CommandDescription
}

type ImagesListSettings struct {
	Server    string `glazed.parameter:"server"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *ImagesListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ImagesListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}

	url := fmt.Sprintf("%s/api/projects/%s/assets", settings.Server, settings.ProjectID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project %s not found", settings.ProjectID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Assets []struct {
			ID          string         `json:"id"`
			ProjectID   string         `json:"project_id"`
			Filename    string         `json:"filename"`
			Width       int            `json:"width"`
			Height      int            `json:"height"`
			ContentType string         `json:"content_type"`
			Bytes       int64          `json:"bytes"`
			UploadedAt  time.Time      `json:"uploaded_at"`
			Metadata    map[string]any `json:"metadata"`
			URL         string         `json:"url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, asset := range result.Assets {
		row := types.NewRow(
			types.MRP("project_id", asset.ProjectID),
			types.MRP("asset_id", asset.ID),
			types.MRP("filename", asset.Filename),
			types.MRP("width", asset.Width),
			types.MRP("height", asset.Height),
			types.MRP("bytes", asset.Bytes),
			types.MRP("content_type", asset.ContentType),
			types.MRP("uploaded_at", asset.UploadedAt.Format(time.RFC3339)),
			types.MRP("metadata", asset.Metadata),
			types.MRP("url", asset.URL),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func NewImagesListCommand() (*ImagesListCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ImagesListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"images-list",
			cmds.WithShort("List images in a zine-layout project"),
			cmds.WithLong(`
List all images in a specific project from the zine-layout server.

Examples:
  images-list --project-id prj-12345
  images-list --project-id prj-12345 --output json
  images-list --project-id prj-12345 --server http://localhost:8089
  images-list --project-id prj-12345 --fields name,width,height
			`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"server",
					parameters.ParameterTypeString,
					parameters.WithDefault("http://localhost:8088"),
					parameters.WithHelp("Server base URL"),
					parameters.WithShortFlag("s"),
				),
				parameters.NewParameterDefinition(
					"project-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Project ID (required)"),
					parameters.WithShortFlag("p"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ImagesListCommand{}
