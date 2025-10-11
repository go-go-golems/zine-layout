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

type GetCommand struct {
	*cmds.CommandDescription
}

type GetSettings struct {
	Server string `glazed.parameter:"server"`
	ID     string `glazed.parameter:"id"`
}

func (c *GetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &GetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ID == "" {
		return fmt.Errorf("id is required")
	}

	url := fmt.Sprintf("%s/api/laid-out-images/%s", settings.Server, settings.ID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch laid-out image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("laid-out image %s not found", settings.ID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		LaidOutImage struct {
			ID         string         `json:"id"`
			ProjectID  string         `json:"project_id"`
			AssetID    string         `json:"asset_id"`
			TemplateID string         `json:"template_id"`
			Overrides  map[string]any `json:"overrides"`
			Result     map[string]any `json:"result"`
			CreatedAt  time.Time      `json:"created_at"`
			UpdatedAt  time.Time      `json:"updated_at"`
		} `json:"laid_out_image"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	overridesJSON := ""
	if result.LaidOutImage.Overrides != nil {
		if data, err := json.MarshalIndent(result.LaidOutImage.Overrides, "", "  "); err == nil {
			overridesJSON = string(data)
		}
	}
	resultJSON := ""
	if result.LaidOutImage.Result != nil {
		if data, err := json.MarshalIndent(result.LaidOutImage.Result, "", "  "); err == nil {
			resultJSON = string(data)
		}
	}

	row := types.NewRow(
		types.MRP("laid_out_image_id", result.LaidOutImage.ID),
		types.MRP("project_id", result.LaidOutImage.ProjectID),
		types.MRP("asset_id", result.LaidOutImage.AssetID),
		types.MRP("template_id", result.LaidOutImage.TemplateID),
		types.MRP("created_at", result.LaidOutImage.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", result.LaidOutImage.UpdatedAt.Format(time.RFC3339)),
		types.MRP("overrides", overridesJSON),
		types.MRP("result", resultJSON),
	)
	return gp.AddRow(ctx, row)
}

func NewGetCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &GetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Fetch a laid-out image"),
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

var _ cmds.GlazeCommand = &GetCommand{}
