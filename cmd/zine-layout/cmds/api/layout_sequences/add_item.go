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

type AddItemCommand struct {
	*cmds.CommandDescription
}

type AddItemSettings struct {
	Server         string `glazed.parameter:"server"`
	SequenceID     string `glazed.parameter:"sequence-id"`
	LaidOutImageID string `glazed.parameter:"laid-out-image-id"`
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
		return fmt.Errorf("sequence-id is required")
	}
	if strings.TrimSpace(settings.LaidOutImageID) == "" {
		return fmt.Errorf("laid-out-image-id is required")
	}

	payload := map[string]any{
		"laid_out_image_id": settings.LaidOutImageID,
	}
	url := fmt.Sprintf("%s/api/layout-sequences/%s/items", settings.Server, settings.SequenceID)
	data, err := httpPostJSON(url, payload)
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

func NewAddItemCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &AddItemCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add-item",
			cmds.WithShort("Append a laid-out image to a layout sequence"),
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
					"laid-out-image-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Laid-out image ID to append"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &AddItemCommand{}
