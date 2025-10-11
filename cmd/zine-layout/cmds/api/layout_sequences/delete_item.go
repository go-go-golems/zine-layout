package layoutsequences

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type DeleteItemCommand struct {
	*cmds.CommandDescription
}

type DeleteItemSettings struct {
	Server     string `glazed.parameter:"server"`
	SequenceID string `glazed.parameter:"sequence-id"`
	Position   int    `glazed.parameter:"position"`
}

func (c *DeleteItemCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &DeleteItemSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SequenceID == "" {
		return fmt.Errorf("sequence-id is required")
	}
	if settings.Position < 0 {
		return fmt.Errorf("position must be >= 0")
	}

	url := fmt.Sprintf("%s/api/layout-sequences/%s/items/%d", settings.Server, settings.SequenceID, settings.Position)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("item position %d not found", settings.Position)
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	row := types.NewRow(
		types.MRP("sequence_id", settings.SequenceID),
		types.MRP("position", strconv.Itoa(settings.Position)),
		types.MRP("status", "deleted"),
	)
	return gp.AddRow(ctx, row)
}

func NewDeleteItemCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &DeleteItemCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete-item",
			cmds.WithShort("Remove an item from a layout sequence by position"),
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
					"position",
					parameters.ParameterTypeInteger,
					parameters.WithRequired(true),
					parameters.WithHelp("Zero-based position to delete"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &DeleteItemCommand{}
