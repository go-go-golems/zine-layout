package laidoutpagescmd

import (
	"context"
	"fmt"
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

type laidOutPagesCreateCommand struct {
	*cmds.CommandDescription
}

type laidOutPagesCreateSettings struct {
	DataRoot       string `glazed.parameter:"data-root"`
	ProjectID      string `glazed.parameter:"project-id"`
	TemplateID     string `glazed.parameter:"template-id"`
	LaidOutImageID string `glazed.parameter:"laid-out-image-id"`
}

func (c *laidOutPagesCreateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &laidOutPagesCreateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if strings.TrimSpace(settings.ProjectID) == "" {
		return fmt.Errorf("--project-id is required")
	}
	if strings.TrimSpace(settings.TemplateID) == "" {
		return fmt.Errorf("--template-id is required")
	}
	if strings.TrimSpace(settings.LaidOutImageID) == "" {
		return fmt.Errorf("--laid-out-image-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	service := services.NewPagesService(repos)
	page, err := service.CreatePage(settings.ProjectID, settings.TemplateID, settings.LaidOutImageID)
	if err != nil {
		return err
	}

	pageRow := types.NewRow(
		types.MRP("entity", "laid_out_page"),
		types.MRP("page_id", page.ID),
		types.MRP("project_id", page.ProjectID),
		types.MRP("page_template_id", page.PageTemplateID),
		types.MRP("laid_out_image_id", page.LaidOutImageID),
		types.MRP("created_at", page.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", page.UpdatedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, pageRow)
}

func newLaidOutPagesCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &laidOutPagesCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a print-ready page from a laid-out image"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("project-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Project id")),
				parameters.NewParameterDefinition("template-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Page layout template id")),
				parameters.NewParameterDefinition("laid-out-image-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Laid-out image id to place on page")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesCreateCommand{}
