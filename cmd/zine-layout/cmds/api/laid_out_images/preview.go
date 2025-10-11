package laidoutimages

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type PreviewCommand struct {
	*cmds.CommandDescription
}

type PreviewSettings struct {
	Server string `glazed.parameter:"server"`
	ID     string `glazed.parameter:"id"`
}

func (c *PreviewCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &PreviewSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ID == "" {
		return fmt.Errorf("id is required")
	}

	url := fmt.Sprintf("%s/api/laid-out-images/%s/preview", settings.Server, settings.ID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch preview: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotImplemented {
		return fmt.Errorf("preview not available: server returned 501")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("laid-out image %s not found", settings.ID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	previewJSON := ""
	if data, err := json.MarshalIndent(result, "", "  "); err == nil {
		previewJSON = string(data)
	}

	row := types.NewRow(
		types.MRP("laid_out_image_id", settings.ID),
		types.MRP("preview", previewJSON),
	)
	return gp.AddRow(ctx, row)
}

func NewPreviewCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &PreviewCommand{
		CommandDescription: cmds.NewCommandDescription(
			"preview",
			cmds.WithShort("Retrieve preview data for a laid-out image"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"server",
					parameters.ParameterTypeString,
					parameters.WithDefault("http://localhost:8088"),
					parameters.WithHelp("Server base URL"),
					parameters.WithShortFlag("s"),
				),
				parameters.NewParameterDefinition(
					"id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Laid-out image ID"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &PreviewCommand{}
