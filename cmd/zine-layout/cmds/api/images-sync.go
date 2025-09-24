package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

type ImagesSyncCommand struct {
	*cmds.CommandDescription
}

type ImagesSyncSettings struct {
	Server    string `glazed.parameter:"server"`
	ProjectID string `glazed.parameter:"project-id"`
	Directory string `glazed.parameter:"directory"`
	Recursive bool   `glazed.parameter:"recursive"`
	DryRun    bool   `glazed.parameter:"dry-run"`
}

func (c *ImagesSyncCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ImagesSyncSettings{}
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

	// Get existing images from the project
	url := fmt.Sprintf("%s/api/projects/%s/images", settings.Server, settings.ProjectID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get existing images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d when getting images", resp.StatusCode)
	}

	var existingResult struct {
		Images []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"images"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&existingResult); err != nil {
		return fmt.Errorf("failed to decode existing images response: %w", err)
	}

	// Create a set of existing image names
	existingImages := make(map[string]bool)
	for _, img := range existingResult.Images {
		existingImages[img.Name] = true
	}

	// Find PNG files in directory
	var allPngFiles []string
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
			allPngFiles = append(allPngFiles, path)
		}
		return nil
	}

	if err := filepath.Walk(settings.Directory, walkFn); err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	// Filter out files that already exist in the project
	var newFiles []string
	var skippedFiles []string
	
	for _, file := range allPngFiles {
		fileName := filepath.Base(file)
		if existingImages[fileName] {
			skippedFiles = append(skippedFiles, file)
		} else {
			newFiles = append(newFiles, file)
		}
	}

	fmt.Printf("Found %d PNG files, %d new, %d already exist\n", len(allPngFiles), len(newFiles), len(skippedFiles))

	// Report skipped files
	for _, file := range skippedFiles {
		row := types.NewRow(
			types.MRP("project_id", settings.ProjectID),
			types.MRP("file", file),
			types.MRP("name", filepath.Base(file)),
			types.MRP("status", "skipped"),
			types.MRP("reason", "already exists"),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// If no new files, we're done
	if len(newFiles) == 0 {
		fmt.Println("No new files to upload")
		return nil
	}

	// If dry run, just report what would be uploaded
	if settings.DryRun {
		for _, file := range newFiles {
			row := types.NewRow(
				types.MRP("project_id", settings.ProjectID),
				types.MRP("file", file),
				types.MRP("name", filepath.Base(file)),
				types.MRP("status", "would upload"),
				types.MRP("reason", "dry run mode"),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
		return nil
	}

	// Upload new files in batches
	const batchSize = 10
	for i := 0; i < len(newFiles); i += batchSize {
		end := i + batchSize
		if end > len(newFiles) {
			end = len(newFiles)
		}
		
		batch := newFiles[i:end]
		fmt.Printf("Uploading batch %d/%d (%d files)...\n", (i/batchSize)+1, (len(newFiles)+batchSize-1)/batchSize, len(batch))
		
		uploadUrl := fmt.Sprintf("%s/api/projects/%s/images", settings.Server, settings.ProjectID)
		respBytes, err := httpUploadFiles(uploadUrl, batch)
		if err != nil {
			return fmt.Errorf("failed to upload batch: %w", err)
		}

		var result struct {
			Images []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"images"`
		}

		if err := json.Unmarshal(respBytes, &result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		// Output results for this batch
		for j, image := range result.Images {
			originalFile := batch[j]
			row := types.NewRow(
				types.MRP("project_id", settings.ProjectID),
				types.MRP("file", originalFile),
				types.MRP("id", image.ID),
				types.MRP("name", image.Name),
				types.MRP("width", image.Width),
				types.MRP("height", image.Height),
				types.MRP("batch", (i/batchSize)+1),
				types.MRP("status", "uploaded"),
				types.MRP("url", fmt.Sprintf("%s/api/projects/%s/images/%s", settings.Server, settings.ProjectID, image.ID)),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewImagesSyncCommand() (*ImagesSyncCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ImagesSyncCommand{
		CommandDescription: cmds.NewCommandDescription(
			"images-sync",
			cmds.WithShort("Sync PNG images from directory to project (upload missing only)"),
			cmds.WithLong(`
Synchronize PNG images from a directory to a project, uploading only missing files.
This compares local files with existing images in the project and uploads only new ones.

Examples:
  images-sync --project-id prj-12345 --directory ./images
  images-sync --project-id prj-12345 --directory ./photos --recursive
  images-sync --project-id prj-12345 --directory ./assets --dry-run
  images-sync --project-id prj-12345 --directory ./images --output json
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
				parameters.NewParameterDefinition(
					"dry-run",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Show what would be uploaded without actually uploading"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ImagesSyncCommand{}
