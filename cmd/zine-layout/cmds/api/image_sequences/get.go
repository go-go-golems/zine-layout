package imagesequences

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
		return fmt.Errorf("sequence ID is required")
	}

	url := fmt.Sprintf("%s/api/image-sequences/%s", settings.Server, settings.SequenceID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch sequence: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("sequence %s not found", settings.SequenceID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Sequence struct {
			ID          string    `json:"id"`
			ProjectID   string    `json:"project_id"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"sequence"`
		Items []struct {
			Position int     `json:"position"`
			AssetID  *string `json:"asset_id"`
			IsGap    bool    `json:"is_gap"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	sequence := result.Sequence
	if len(result.Items) == 0 {
		row := types.NewRow(
			types.MRP("sequence_id", sequence.ID),
			types.MRP("project_id", sequence.ProjectID),
			types.MRP("name", sequence.Name),
			types.MRP("description", sequence.Description),
			types.MRP("created_at", sequence.CreatedAt.Format(time.RFC3339)),
			types.MRP("updated_at", sequence.UpdatedAt.Format(time.RFC3339)),
			types.MRP("position", nil),
			types.MRP("asset_id", nil),
			types.MRP("is_gap", nil),
		)
		return gp.AddRow(ctx, row)
	}

	for _, item := range result.Items {
		var assetID any
		if item.AssetID != nil {
			assetID = *item.AssetID
		}
		row := types.NewRow(
			types.MRP("sequence_id", sequence.ID),
			types.MRP("project_id", sequence.ProjectID),
			types.MRP("name", sequence.Name),
			types.MRP("description", sequence.Description),
			types.MRP("created_at", sequence.CreatedAt.Format(time.RFC3339)),
			types.MRP("updated_at", sequence.UpdatedAt.Format(time.RFC3339)),
			types.MRP("position", item.Position),
			types.MRP("asset_id", assetID),
			types.MRP("is_gap", item.IsGap),
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
			cmds.WithShort("Fetch a single image sequence and its items"),
			cmds.WithLong(`
Retrieve detailed information about an image sequence, including ordered items.

Examples:
  image-sequences-get --sequence-id seq-123
  image-sequences-get --sequence-id seq-123 --output json
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
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &GetCommand{}
