package api

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
)

type PresetsListCommand struct {
	*cmds.CommandDescription
}

type PresetsListSettings struct {
	Server string `glazed.parameter:"server"`
}

func (c *PresetsListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &PresetsListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/presets", settings.Server)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get presets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Presets []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"presets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, preset := range result.Presets {
		row := types.NewRow(
			types.MRP("id", preset.ID),
			types.MRP("name", preset.Name),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func NewPresetsListCommand() (*PresetsListCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &PresetsListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"presets-list",
			cmds.WithShort("List available zine-layout presets"),
			cmds.WithLong(`
List all available presets from the zine-layout server.

Examples:
  presets-list
  presets-list --server http://localhost:8089
  presets-list --output json
			`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"server",
					parameters.ParameterTypeString,
					parameters.WithDefault("http://localhost:8088"),
					parameters.WithHelp("Server base URL"),
					parameters.WithShortFlag("s"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &PresetsListCommand{}
