package layoutsequences

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
	Server     string `glazed.parameter:"server"`
	SequenceID string `glazed.parameter:"sequence-id"`
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
	if settings.SequenceID == "" {
		return fmt.Errorf("sequence-id is required")
	}

	url := fmt.Sprintf("%s/api/layout-sequences/%s", settings.Server, settings.SequenceID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch layout sequence: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("layout sequence %s not found", settings.SequenceID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		LayoutSequence struct {
			ID          string    `json:"id"`
			ProjectID   string    `json:"project_id"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"layout_sequence"`
		Items []struct {
			Position       int    `json:"position"`
			LaidOutImageID string `json:"laid_out_image_id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	header := types.NewRow(
		types.MRP("layout_sequence_id", result.LayoutSequence.ID),
		types.MRP("project_id", result.LayoutSequence.ProjectID),
		types.MRP("name", result.LayoutSequence.Name),
		types.MRP("description", result.LayoutSequence.Description),
		types.MRP("created_at", result.LayoutSequence.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", result.LayoutSequence.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, header); err != nil {
		return err
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

func NewGetCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &GetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Fetch a layout sequence and its items"),
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
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &GetCommand{}
