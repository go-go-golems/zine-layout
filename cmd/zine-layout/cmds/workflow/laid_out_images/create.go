package laidoutimagescmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/go-go-golems/zine-layout/pkg/services"
	"github.com/spf13/cobra"
)

type laidOutImagesCreateCommand struct {
	*cmds.CommandDescription
}

type laidOutImagesCreateSettings struct {
	DataRoot      string `glazed.parameter:"data-root"`
	ProjectID     string `glazed.parameter:"project-id"`
	AssetID       string `glazed.parameter:"asset-id"`
	TemplateID    string `glazed.parameter:"template-id"`
	OverridesJSON string `glazed.parameter:"overrides-json"`
	File          string `glazed.parameter:"file"`
}

func (c *laidOutImagesCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutImagesCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ProjectID == "" {
		return fmt.Errorf("--project-id is required")
	}
	if settings.AssetID == "" {
		return fmt.Errorf("--asset-id is required")
	}
	if settings.TemplateID == "" {
		return fmt.Errorf("--template-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	var overrides *string
	payload := strings.TrimSpace(settings.OverridesJSON)
	if payload == "" && strings.TrimSpace(settings.File) != "" {
		data, err := os.ReadFile(settings.File)
		if err != nil {
			return fmt.Errorf("read overrides file: %w", err)
		}
		payload = string(data)
	}
	if payload != "" {
		clean := payload
		overrides = &clean
	}

	service := services.NewLayoutService(repos)
	record, err := service.CreateLaidOutImage(settings.ProjectID, settings.AssetID, settings.TemplateID, overrides)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "laid_out_image"),
		types.MRP("laid_out_image_id", record.ID),
		types.MRP("project_id", record.ProjectID),
		types.MRP("asset_id", record.AssetID),
		types.MRP("template_id", record.TemplateID),
		types.MRP("created_at", record.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", record.UpdatedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newLaidOutImagesCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &laidOutImagesCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Compute a laid-out image"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"project-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Project identifier"),
				),
				parameters.NewParameterDefinition(
					"asset-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Asset identifier"),
				),
				parameters.NewParameterDefinition(
					"template-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Image layout template identifier"),
				),
				parameters.NewParameterDefinition(
					"overrides-json",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Overrides JSON to merge onto the template"),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to JSON file containing overrides"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutImagesCreateCommand{}
