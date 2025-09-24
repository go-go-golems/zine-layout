package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type PagesPutCommand struct {
	*cmds.CommandDescription
}

type PagesPutSettings struct {
	Server     string `glazed.parameter:"server"`
	ProjectID  string `glazed.parameter:"project-id"`
	PageNumber int    `glazed.parameter:"page-number"`
	AssetID    string `glazed.parameter:"asset-id"`
	Settings   string `glazed.parameter:"settings"`
}

func (c *PagesPutCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &PagesPutSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if settings.PageNumber == 0 {
		return fmt.Errorf("page number is required")
	}

	// Create the page data
	pageData := map[string]any{}
	
	if settings.AssetID != "" {
		pageData["asset_id"] = settings.AssetID
	}

	// Parse settings JSON or use defaults
	if settings.Settings != "" {
		var settingsObj any
		if err := json.Unmarshal([]byte(settings.Settings), &settingsObj); err != nil {
			return fmt.Errorf("invalid settings JSON: %w", err)
		}
		pageData["settings"] = settingsObj
	} else {
		// Default settings for a simple page
		pageData["settings"] = map[string]any{
			"paper_size": "letter",
			"dpi":        300,
			"margins": map[string]float64{
				"top":    0.5,
				"bottom": 0.5,
				"left":   0.5,
				"right":  0.5,
			},
		}
	}

	url := fmt.Sprintf("%s/api/projects/%s/pages/%d", settings.Server, settings.ProjectID, settings.PageNumber)
	respBytes, err := httpPutJSON(url, pageData)
	if err != nil {
		return fmt.Errorf("failed to create/update page: %w", err)
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

	if err := json.Unmarshal(respBytes, &result); err != nil {
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
		types.MRP("status", "saved"),
	)

	return gp.AddRow(ctx, row)
}

func NewPagesPutCommand() (*PagesPutCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &PagesPutCommand{
		CommandDescription: cmds.NewCommandDescription(
			"pages-put",
			cmds.WithShort("Create or update a page in a zine-layout project"),
			cmds.WithLong(`
Create or update a page in a project with settings and optional asset reference.

Examples:
  pages-put --project-id prj-12345 --page-number 1
  pages-put --project-id prj-12345 --page-number 1 --asset-id img-123
  pages-put --project-id prj-12345 --page-number 1 --settings '{"paper_size":"a4","dpi":300}'
  pages-put --project-id prj-12345 --page-number 1 --asset-id img-123 --output json
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
				parameters.NewParameterDefinition(
					"asset-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Asset ID to associate with the page"),
					parameters.WithShortFlag("a"),
				),
				parameters.NewParameterDefinition(
					"settings",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Settings JSON for the page (defaults provided if not specified)"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &PagesPutCommand{}
