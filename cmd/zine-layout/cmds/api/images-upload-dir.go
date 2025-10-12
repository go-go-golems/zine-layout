package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type ImagesUploadDirCommand struct {
	*cmds.CommandDescription
}

type ImagesUploadDirSettings struct {
	Server    string `glazed.parameter:"server"`
	ProjectID string `glazed.parameter:"project-id"`
	Directory string `glazed.parameter:"directory"`
	Recursive bool   `glazed.parameter:"recursive"`
}

func (c *ImagesUploadDirCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ImagesUploadDirSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if settings.Directory == "" {
		return fmt.Errorf("directory is required")
	}

	// Check if directory exists
	if _, err := os.Stat(settings.Directory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", settings.Directory)
	}

	// Find PNG files in directory
	var pngFiles []string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// If not recursive, skip subdirectories
			if !settings.Recursive && path != settings.Directory {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a PNG file
		if strings.ToLower(filepath.Ext(path)) == ".png" {
			pngFiles = append(pngFiles, path)
		}
		return nil
	}

	if err := filepath.Walk(settings.Directory, walkFn); err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(pngFiles) == 0 {
		return fmt.Errorf("no PNG files found in directory: %s", settings.Directory)
	}

	fmt.Printf("Found %d PNG file(s) in %s\n", len(pngFiles), settings.Directory)

	// Upload in batches to avoid overwhelming the server
	const batchSize = 10
	for i := 0; i < len(pngFiles); i += batchSize {
		end := i + batchSize
		if end > len(pngFiles) {
			end = len(pngFiles)
		}

		batch := pngFiles[i:end]
		fmt.Printf("Uploading batch %d/%d (%d files)...\n", (i/batchSize)+1, (len(pngFiles)+batchSize-1)/batchSize, len(batch))

		url := fmt.Sprintf("%s/api/projects/%s/images", settings.Server, settings.ProjectID)
		respBytes, err := httpUploadFiles(url, batch)
		if err != nil {
			return fmt.Errorf("failed to upload batch: %w", err)
		}

		var result struct {
			Assets []struct {
				ID       string `json:"id"`
				Filename string `json:"filename"`
				Width    int    `json:"width"`
				Height   int    `json:"height"`
				URL      string `json:"url"`
			} `json:"assets"`
		}

		if err := json.Unmarshal(respBytes, &result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		// Output results for this batch
		for j, asset := range result.Assets {
			originalFile := batch[j]
			row := types.NewRow(
				types.MRP("project_id", settings.ProjectID),
				types.MRP("original_file", originalFile),
				types.MRP("asset_id", asset.ID),
				types.MRP("filename", asset.Filename),
				types.MRP("width", asset.Width),
				types.MRP("height", asset.Height),
				types.MRP("batch", (i/batchSize)+1),
				types.MRP("status", "uploaded"),
				types.MRP("url", asset.URL),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewImagesUploadDirCommand() (*ImagesUploadDirCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ImagesUploadDirCommand{
		CommandDescription: cmds.NewCommandDescription(
			"images-upload-dir",
			cmds.WithShort("Upload all PNG images from a directory to a zine-layout project"),
			cmds.WithLong(`
Upload all PNG images from a directory to a specific project.
Files are uploaded in batches to avoid overwhelming the server.

Examples:
  images-upload-dir --project-id prj-12345 --directory ./images
  images-upload-dir --project-id prj-12345 --directory ./photos --recursive
  images-upload-dir --project-id prj-12345 --directory ./assets --output json
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
				parameters.NewParameterDefinition(
					"directory",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Directory containing PNG files (required)"),
					parameters.WithShortFlag("d"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"recursive",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Include subdirectories recursively"),
					parameters.WithShortFlag("r"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ImagesUploadDirCommand{}
