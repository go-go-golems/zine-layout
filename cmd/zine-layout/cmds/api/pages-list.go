package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type PagesListCommand struct {
	*cmds.CommandDescription
}

type PagesListSettings struct {
	Server    string `glazed.parameter:"server"`
	ProjectID string `glazed.parameter:"project-id"`
}

func (c *PagesListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &PagesListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}

	url := fmt.Sprintf("%s/api/projects/%s/pages", settings.Server, settings.ProjectID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get pages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project %s not found", settings.ProjectID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Pages []struct {
			PageNumber  int    `json:"page_number"`
			AssetID     string `json:"asset_id,omitempty"`
			Settings    any    `json:"settings"`
			Result      any    `json:"result,omitempty"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		} `json:"pages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, page := range result.Pages {
		settingsJson, _ := json.Marshal(page.Settings)
		resultJson, _ := json.Marshal(page.Result)
		
		row := types.NewRow(
			types.MRP("project_id", settings.ProjectID),
			types.MRP("page_number", page.PageNumber),
			types.MRP("asset_id", page.AssetID),
			types.MRP("settings", string(settingsJson)),
			types.MRP("result", string(resultJson)),
			types.MRP("created_at", page.CreatedAt),
			types.MRP("updated_at", page.UpdatedAt),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func NewPagesListCommand() (*PagesListCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &PagesListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"pages-list",
			cmds.WithShort("List pages in a zine-layout project"),
			cmds.WithLong(`
List all pages in a specific project from the zine-layout server.

Examples:
  pages-list --project-id prj-12345
  pages-list --project-id prj-12345 --output json
  pages-list --project-id prj-12345 --server http://localhost:8089
  pages-list --project-id prj-12345 --fields page_number,asset_id
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
					parameters.WithHelp("Project ID (required)"),
					parameters.WithShortFlag("p"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &PagesListCommand{}
