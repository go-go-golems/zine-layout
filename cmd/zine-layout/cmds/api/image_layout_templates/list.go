package imagelayouttemplates

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

type ListCommand struct {
	*cmds.CommandDescription
}

type ListSettings struct {
	Server    string `glazed.parameter:"server"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *ListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	url := ""
	if settings.ProjectID != "" {
		url = fmt.Sprintf("%s/api/projects/%s/image-layout-templates", settings.Server, settings.ProjectID)
	} else {
		url = fmt.Sprintf("%s/api/image-layout-templates", settings.Server)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to list image layout templates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		if settings.ProjectID != "" {
			return fmt.Errorf("project %s not found", settings.ProjectID)
		}
		return fmt.Errorf("endpoint not found")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Templates []struct {
			ID          string         `json:"id"`
			ProjectID   *string        `json:"project_id"`
			Scope       string         `json:"scope"`
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Settings    map[string]any `json:"settings"`
			CreatedAt   time.Time      `json:"created_at"`
			UpdatedAt   time.Time      `json:"updated_at"`
		} `json:"templates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, tpl := range result.Templates {
		projectID := ""
		if tpl.ProjectID != nil {
			projectID = *tpl.ProjectID
		}
		row := types.NewRow(
			types.MRP("template_id", tpl.ID),
			types.MRP("scope", tpl.Scope),
			types.MRP("project_id", projectID),
			types.MRP("name", tpl.Name),
			types.MRP("description", tpl.Description),
			types.MRP("updated_at", tpl.UpdatedAt.Format(time.RFC3339)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func NewListCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &ListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List image layout templates"),
			cmds.WithLong(`
List image layout templates. Provide --project-id to include project-scoped templates plus globals; omit it to list global templates only.
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
					"project-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Project ID (optional)"),
					parameters.WithShortFlag("p"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return buildCommand(cmd)
}

var _ cmds.GlazeCommand = &ListCommand{}
