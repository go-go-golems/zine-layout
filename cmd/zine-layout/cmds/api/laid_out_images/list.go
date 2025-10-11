package laidoutimages

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
	"github.com/spf13/cobra"
)

type ListCommand struct {
	*cmds.CommandDescription
}

type ListSettings struct {
	Server    string `glazed.parameter:"server"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *ListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("project-id is required")
	}

	url := fmt.Sprintf("%s/api/projects/%s/laid-out-images", settings.Server, settings.ProjectID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to list laid-out images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project %s not found", settings.ProjectID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		LaidOutImages []struct {
			ID         string         `json:"id"`
			ProjectID  string         `json:"project_id"`
			AssetID    string         `json:"asset_id"`
			TemplateID string         `json:"template_id"`
			Overrides  map[string]any `json:"overrides"`
			Result     map[string]any `json:"result"`
			CreatedAt  time.Time      `json:"created_at"`
			UpdatedAt  time.Time      `json:"updated_at"`
		} `json:"laid_out_images"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, item := range result.LaidOutImages {
		row := types.NewRow(
			types.MRP("laid_out_image_id", item.ID),
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

func NewListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &ListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List laid-out images for a project"),
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
					parameters.WithRequired(true),
					parameters.WithHelp("Project ID"),
					parameters.WithShortFlag("p"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &ListCommand{}
