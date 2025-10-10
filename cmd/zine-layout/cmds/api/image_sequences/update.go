package imagesequences

import (
	"context"
	"encoding/json"
	"fmt"
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
	Server      string `glazed.parameter:"server"`
	SequenceID  string `glazed.parameter:"sequence-id"`
	Name        string `glazed.parameter:"name"`
	Description string `glazed.parameter:"description"`
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

	if settings.SequenceID == "" {
		return fmt.Errorf("sequence ID is required")
	}

	payload := map[string]string{}
	if name := strings.TrimSpace(settings.Name); name != "" {
		payload["name"] = name
	}
	if desc := strings.TrimSpace(settings.Description); desc != "" {
		payload["description"] = desc
	}
	if len(payload) == 0 {
		return fmt.Errorf("provide at least one of --name or --description")
	}

	url := fmt.Sprintf("%s/api/image-sequences/%s", settings.Server, settings.SequenceID)
	respBytes, err := httpPutJSON(url, payload)
	if err != nil {
		return fmt.Errorf("failed to update sequence: %w", err)
	}

	var result struct {
		Sequence struct {
			ID          string    `json:"id"`
			ProjectID   string    `json:"project_id"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"sequence"`
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	seq := result.Sequence
	row := types.NewRow(
		types.MRP("sequence_id", seq.ID),
		types.MRP("project_id", seq.ProjectID),
		types.MRP("name", seq.Name),
		types.MRP("description", seq.Description),
		types.MRP("updated_at", seq.UpdatedAt.Format(time.RFC3339)),
		types.MRP("status", "updated"),
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
			cmds.WithShort("Update an image sequence's metadata"),
			cmds.WithLong(`
Update the name and/or description of an image sequence.

Examples:
  image-sequences-update --sequence-id seq-123 --name "New Name"
  image-sequences-update --sequence-id seq-123 --description "Updated details"
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
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New name (optional)"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New description (optional)"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &UpdateCommand{}
