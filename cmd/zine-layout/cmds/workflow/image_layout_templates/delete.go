package imagelayouttemplatescmd

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/spf13/cobra"
)

type templatesDeleteCommand struct {
	*cmds.CommandDescription
}

type templatesDeleteSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	TemplateID string `glazed.parameter:"template-id"`
}

func (c *templatesDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &templatesDeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.TemplateID == "" {
		return fmt.Errorf("--template-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := repos.ImageLayoutTemplates.Delete(settings.TemplateID); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("entity", "image_layout_template"),
		types.MRP("template_id", settings.TemplateID),
		types.MRP("status", "deleted"),
		types.MRP("deleted_at", time.Now().UTC().Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func newTemplatesDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &templatesDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete an image layout template"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"template-id",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Template identifier"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &templatesDeleteCommand{}
