package imagesequences

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type AddItemCommand struct {
	*cmds.CommandDescription
}

type AddItemSettings struct {
	Server     string `glazed.parameter:"server"`
	SequenceID string `glazed.parameter:"sequence-id"`
	AssetID    string `glazed.parameter:"asset-id"`
	IsGap      bool   `glazed.parameter:"gap"`
}

func (c *AddItemCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &AddItemSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.SequenceID == "" {
		return fmt.Errorf("sequence ID is required")
	}
	if !settings.IsGap && strings.TrimSpace(settings.AssetID) == "" {
		return fmt.Errorf("asset ID is required unless --gap is set")
	}

	var assetIDPtr *string
	if trimmed := strings.TrimSpace(settings.AssetID); trimmed != "" {
		assetIDPtr = &trimmed
	}

	body := map[string]any{
		"asset_id": assetIDPtr,
		"is_gap":   settings.IsGap,
	}

	url := fmt.Sprintf("%s/api/image-sequences/%s/items", settings.Server, settings.SequenceID)
	respBytes, err := httpPostJSON(url, body)
	if err != nil {
		return fmt.Errorf("failed to add item: %w", err)
	}

	var result struct {
		Items []struct {
			Position int     `json:"position"`
			AssetID  *string `json:"asset_id"`
			IsGap    bool    `json:"is_gap"`
		} `json:"items"`
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, item := range result.Items {
		row := types.NewRow(
			types.MRP("sequence_id", settings.SequenceID),
			types.MRP("position", item.Position),
			types.MRP("asset_id", item.AssetID),
			types.MRP("is_gap", item.IsGap),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func NewAddItemCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &AddItemCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add-item",
			cmds.WithShort("Append an item to an image sequence"),
			cmds.WithLong(`
Append a new item to the end of an image sequence. Provide either --asset-id or --gap.

Examples:
  image-sequences add-item --sequence-id seq-123 --asset-id ast-456
  image-sequences add-item --sequence-id seq-123 --gap
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
					"sequence-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Sequence ID (required)"),
					parameters.WithShortFlag("q"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"asset-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Asset ID to insert (omit when using --gap)"),
				),
				parameters.NewParameterDefinition(
					"gap",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Insert a gap placeholder"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &AddItemCommand{}
