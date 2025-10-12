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

type ReorderCommand struct {
	*cmds.CommandDescription
}

type ReorderSettings struct {
	Server     string `glazed.parameter:"server"`
	SequenceID string `glazed.parameter:"sequence-id"`
	Items      string `glazed.parameter:"items"`
}

func (c *ReorderCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ReorderSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.SequenceID == "" {
		return fmt.Errorf("sequence ID is required")
	}

	tokens := parseItemTokens(settings.Items)
	if len(tokens) == 0 {
		return fmt.Errorf("provide at least one item token")
	}

	replacements := make([]map[string]any, 0, len(tokens))
	for _, token := range tokens {
		if token == "" {
			replacements = append(replacements, map[string]any{"asset_id": nil, "is_gap": true})
			continue
		}
		replacements = append(replacements, map[string]any{"asset_id": token, "is_gap": false})
	}

	body := map[string]any{"items": replacements}
	url := fmt.Sprintf("%s/api/image-sequences/%s/items", settings.Server, settings.SequenceID)
	respBytes, err := httpPutJSON(url, body)
	if err != nil {
		return fmt.Errorf("failed to reorder items: %w", err)
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

func NewReorderCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &ReorderCommand{
		CommandDescription: cmds.NewCommandDescription(
			"reorder",
			cmds.WithShort("Replace the items of an image sequence"),
			cmds.WithLong(`
Replace the ordered items of an image sequence. Provide a comma-separated list of tokens.
Use the literal "gap" to insert empty slots. Example: ast-1,gap,ast-3

Examples:
  image-sequences reorder --sequence-id seq-123 --items ast-1,ast-2,ast-3
  image-sequences reorder --sequence-id seq-123 --items ast-1,gap,ast-3
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
					"items",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Comma-separated list of asset IDs; use 'gap' for blanks"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

func parseItemTokens(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		lower := strings.ToLower(token)
		if lower == "gap" || lower == "_" || lower == "-" {
			items = append(items, "")
			continue
		}
		items = append(items, token)
	}
	return items
}

var _ cmds.GlazeCommand = &ReorderCommand{}
