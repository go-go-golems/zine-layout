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

type laidOutImagesUpdateCommand struct {
	*cmds.CommandDescription
}

type laidOutImagesUpdateSettings struct {
	DataRoot       string `glazed.parameter:"data-root"`
	ID             string `glazed.parameter:"id"`
	TemplateID     string `glazed.parameter:"template-id"`
	OverridesJSON  string `glazed.parameter:"overrides-json"`
	File           string `glazed.parameter:"file"`
	ClearOverrides bool   `glazed.parameter:"clear-overrides"`
}

func (c *laidOutImagesUpdateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutImagesUpdateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ID == "" {
		return fmt.Errorf("--id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	record, err := repos.LaidOutImages.Get(settings.ID)
	if err != nil {
		return err
	}

	if trimmed := strings.TrimSpace(settings.TemplateID); trimmed != "" {
		record.TemplateID = trimmed
	}

	if settings.ClearOverrides {
		record.OverridesJSON = nil
	} else {
		payload := strings.TrimSpace(settings.OverridesJSON)
		if payload == "" && strings.TrimSpace(settings.File) != "" {
			data, err := os.ReadFile(settings.File)
			if err != nil {
				return fmt.Errorf("read overrides file: %w", err)
			}
			payload = string(data)
		}
		if payload != "" {
			str := payload
			record.OverridesJSON = &str
		}
	}

	service := services.NewLayoutService(repos)
	if err := service.RecomputeLaidOutImage(record); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "laid_out_image"),
		types.MRP("laid_out_image_id", record.ID),
		types.MRP("project_id", record.ProjectID),
		types.MRP("asset_id", record.AssetID),
		types.MRP("template_id", record.TemplateID),
		types.MRP("updated_at", record.UpdatedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newLaidOutImagesUpdateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &laidOutImagesUpdateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Recompute a laid-out image with new settings"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Laid-out image identifier"),
				),
				parameters.NewParameterDefinition(
					"template-id",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("New template identifier"),
				),
				parameters.NewParameterDefinition(
					"overrides-json",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Replacement overrides JSON"),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Path to overrides JSON file"),
				),
				parameters.NewParameterDefinition(
					"clear-overrides",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Remove overrides before recomputing"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutImagesUpdateCommand{}
