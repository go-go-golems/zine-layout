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

type PagesGetCommand struct {
	*cmds.CommandDescription
}

type PagesGetSettings struct {
	Server     string `glazed.parameter:"server"`
	ProjectID  string `glazed.parameter:"project-id"`
	PageNumber int    `glazed.parameter:"page-number"`
}

func (c *PagesGetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &PagesGetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if settings.PageNumber == 0 {
		return fmt.Errorf("page number is required")
	}

	url := fmt.Sprintf("%s/api/projects/%s/pages/%d", settings.Server, settings.ProjectID, settings.PageNumber)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("page %d not found in project %s", settings.PageNumber, settings.ProjectID)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result struct {
		Page struct {
			PageNumber  int    `json:"page_number"`
			AssetID     string `json:"asset_id,omitempty"`
			Settings    any    `json:"settings"`
			Result      any    `json:"result,omitempty"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		} `json:"page"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	page := result.Page
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

	return gp.AddRow(ctx, row)
}

func NewPagesGetCommand() (*PagesGetCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &PagesGetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"pages-get",
			cmds.WithShort("Get a specific page in a zine-layout project"),
			cmds.WithLong(`
Get details of a specific page by page number from a project.

Examples:
  pages-get --project-id prj-12345 --page-number 1
  pages-get --project-id prj-12345 --page-number 1 --output json
  pages-get --project-id prj-12345 --page-number 1 --server http://localhost:8089
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
				parameters.NewParameterDefinition(
					"page-number",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(0),
					parameters.WithHelp("Page number (required)"),
					parameters.WithShortFlag("n"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &PagesGetCommand{}
