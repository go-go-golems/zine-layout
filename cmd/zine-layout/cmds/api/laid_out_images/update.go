package laidoutimages

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type UpdateCommand struct {
	*cmds.CommandDescription
}

type UpdateSettings struct {
	Server        string `glazed.parameter:"server"`
	ID            string `glazed.parameter:"id"`
	TemplateID    string `glazed.parameter:"template-id"`
	OverridesFile string `glazed.parameter:"overrides-file"`
	OverridesJSON string `glazed.parameter:"overrides-json"`
}

func (c *UpdateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &UpdateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ID == "" {
		return fmt.Errorf("id is required")
	}

	payload := map[string]any{}
	if strings.TrimSpace(settings.TemplateID) != "" {
		payload["template_id"] = settings.TemplateID
	}

	overrides := map[string]any{}
	if strings.TrimSpace(settings.OverridesJSON) != "" {
		if err := json.Unmarshal([]byte(settings.OverridesJSON), &overrides); err != nil {
			return fmt.Errorf("invalid overrides json: %w", err)
		}
	}
	if settings.OverridesFile != "" {
		data, err := os.ReadFile(settings.OverridesFile)
		if err != nil {
			return fmt.Errorf("failed to read overrides file: %w", err)
		}
		if err := json.Unmarshal(data, &overrides); err != nil {
			return fmt.Errorf("invalid json in overrides file: %w", err)
		}
	}
	if len(overrides) > 0 {
		payload["overrides"] = overrides
	}

	if len(payload) == 0 {
		return fmt.Errorf("no updates specified")
	}

	url := fmt.Sprintf("%s/api/laid-out-images/%s", settings.Server, settings.ID)
	data, err := httpPutJSON(url, payload)
	if err != nil {
		return err
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
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	overridesJSON := ""
	if result.LaidOutImage.Overrides != nil {
		if raw, err := json.MarshalIndent(result.LaidOutImage.Overrides, "", "  "); err == nil {
			overridesJSON = string(raw)
		}
	}
	resultJSON := ""
	if result.LaidOutImage.Result != nil {
		if raw, err := json.MarshalIndent(result.LaidOutImage.Result, "", "  "); err == nil {
			resultJSON = string(raw)
		}
	}

	row := types.NewRow(
		types.MRP("laid_out_image_id", result.LaidOutImage.ID),
		types.MRP("project_id", result.LaidOutImage.ProjectID),
		types.MRP("asset_id", result.LaidOutImage.AssetID),
		types.MRP("template_id", result.LaidOutImage.TemplateID),
		types.MRP("updated_at", result.LaidOutImage.UpdatedAt.Format(time.RFC3339)),
		types.MRP("overrides", overridesJSON),
		types.MRP("result", resultJSON),
	)
	return gp.AddRow(ctx, row)
}

func NewUpdateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &UpdateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Update a laid-out image"),
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
				parameters.NewParameterDefinition(
					"template-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New template ID"),
				),
				parameters.NewParameterDefinition(
					"overrides-file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to overrides JSON"),
				),
				parameters.NewParameterDefinition(
					"overrides-json",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Overrides JSON string"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &UpdateCommand{}
