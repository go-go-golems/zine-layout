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

type ImagesUploadCommand struct {
	*cmds.CommandDescription
}

type ImagesUploadSettings struct {
	Server    string   `glazed.parameter:"server"`
	ProjectID string   `glazed.parameter:"project-id"`
	Files     []string `glazed.parameter:"files"`
}

func (c *ImagesUploadCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ImagesUploadSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if len(settings.Files) == 0 {
		return fmt.Errorf("at least one file is required")
	}

	// Validate files exist and are PNG images
	var validFiles []string
	for _, file := range settings.Files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", file)
		}
		
		// Check file extension (basic validation)
		ext := strings.ToLower(filepath.Ext(file))
		if ext != ".png" {
			fmt.Printf("Warning: %s is not a PNG file, skipping\n", file)
			continue
		}
		
		validFiles = append(validFiles, file)
	}

	if len(validFiles) == 0 {
		return fmt.Errorf("no valid PNG files found")
	}

	fmt.Printf("Uploading %d file(s) to project %s...\n", len(validFiles), settings.ProjectID)

	url := fmt.Sprintf("%s/api/projects/%s/images", settings.Server, settings.ProjectID)
	respBytes, err := httpUploadFiles(url, validFiles)
	if err != nil {
		return fmt.Errorf("failed to upload files: %w", err)
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

	for i, image := range result.Images {
		originalFile := validFiles[i] // Assuming server returns images in same order
		row := types.NewRow(
			types.MRP("project_id", settings.ProjectID),
			types.MRP("original_file", originalFile),
			types.MRP("id", image.ID),
			types.MRP("name", image.Name),
			types.MRP("width", image.Width),
			types.MRP("height", image.Height),
			types.MRP("status", "uploaded"),
			types.MRP("url", fmt.Sprintf("%s/api/projects/%s/images/%s", settings.Server, settings.ProjectID, image.ID)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func NewImagesUploadCommand() (*ImagesUploadCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ImagesUploadCommand{
		CommandDescription: cmds.NewCommandDescription(
			"images-upload",
			cmds.WithShort("Upload images to a zine-layout project"),
			cmds.WithLong(`
Upload one or more PNG images to a specific project.

Examples:
  images-upload --project-id prj-12345 --files image1.png
  images-upload --project-id prj-12345 --files image1.png,image2.png,image3.png
  images-upload --project-id prj-12345 --files "*.png"
  images-upload --project-id prj-12345 --files img1.png,img2.png --output json
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
					"files",
					parameters.ParameterTypeStringList,
					parameters.WithDefault([]string{}),
					parameters.WithHelp("List of PNG files to upload"),
					parameters.WithShortFlag("f"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ImagesUploadCommand{}
