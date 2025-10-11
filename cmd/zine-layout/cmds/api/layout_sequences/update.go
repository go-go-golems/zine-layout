package layoutsequences

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
		return fmt.Errorf("sequence-id is required")
	}

	payload := map[string]any{}
	if strings.TrimSpace(settings.Name) != "" {
		payload["name"] = settings.Name
	}
	if strings.TrimSpace(settings.Description) != "" {
		payload["description"] = settings.Description
	}
	if len(payload) == 0 {
		return fmt.Errorf("no updates specified")
	}

	url := fmt.Sprintf("%s/api/layout-sequences/%s", settings.Server, settings.SequenceID)
	data, err := httpPutJSON(url, payload)
	if err != nil {
		return err
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
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	row := types.NewRow(
		types.MRP("layout_sequence_id", result.LayoutSequence.ID),
		types.MRP("project_id", result.LayoutSequence.ProjectID),
		types.MRP("name", result.LayoutSequence.Name),
		types.MRP("description", result.LayoutSequence.Description),
		types.MRP("updated_at", result.LayoutSequence.UpdatedAt.Format(time.RFC3339)),
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
			cmds.WithShort("Update a layout sequence"),
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
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New name"),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New description"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &UpdateCommand{}
