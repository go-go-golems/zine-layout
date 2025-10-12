package layoutsequences

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
	Order      string `glazed.parameter:"order"`
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
		return fmt.Errorf("sequence-id is required")
	}
	orderRaw := strings.TrimSpace(settings.Order)
	if orderRaw == "" {
		return fmt.Errorf("order is required")
	}

	ids := []string{}
	for _, part := range strings.Split(orderRaw, ",") {
		id := strings.TrimSpace(part)
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("order must contain at least one laid-out image id")
	}

	replacements := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		replacements = append(replacements, map[string]any{
			"laid_out_image_id": id,
		})
	}

	payload := map[string]any{"items": replacements}
	url := fmt.Sprintf("%s/api/layout-sequences/%s/items", settings.Server, settings.SequenceID)
	data, err := httpPutJSON(url, payload)
	if err != nil {
		return err
	}

	var result struct {
		Items []struct {
			Position       int    `json:"position"`
			LaidOutImageID string `json:"laid_out_image_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, item := range result.Items {
		row := types.NewRow(
			types.MRP("position", item.Position),
			types.MRP("laid_out_image_id", item.LaidOutImageID),
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
			cmds.WithShort("Replace layout sequence ordering"),
			cmds.WithLong(`Provide a comma-separated list of laid-out image IDs to define the new order.`),
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
					parameters.WithRequired(true),
					parameters.WithHelp("Layout sequence ID"),
				),
				parameters.NewParameterDefinition(
					"order",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Comma-separated laid-out image IDs in desired order"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &ReorderCommand{}
