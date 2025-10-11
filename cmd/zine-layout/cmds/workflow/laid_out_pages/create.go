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
	DataRoot   string   `glazed.parameter:"data-root"`
	ProjectID  string   `glazed.parameter:"project-id"`
	TemplateID string   `glazed.parameter:"template-id"`
	Inputs     []string `glazed.parameter:"inputs"`
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

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	service := services.NewPagesService(repos)
	page, inputs, err := service.CreatePage(settings.ProjectID, settings.TemplateID, settings.Inputs)
	if err != nil {
		return err
	}

	pageRow := types.NewRow(
		types.MRP("entity", "laid_out_page"),
		types.MRP("page_id", page.ID),
		types.MRP("project_id", page.ProjectID),
		types.MRP("page_template_id", page.PageTemplateID),
		types.MRP("created_at", page.CreatedAt.Format(time.RFC3339)),
		types.MRP("updated_at", page.UpdatedAt.Format(time.RFC3339)),
	)
	if err := gp.AddRow(ctx, pageRow); err != nil {
		return err
	}

	for _, input := range inputs {
		row := types.NewRow(
			types.MRP("entity", "laid_out_page_input"),
			types.MRP("page_id", input.PageID),
			types.MRP("position", input.InputIndex),
			types.MRP("laid_out_image_id", input.LaidOutImageID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func newLaidOutPagesCreateCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &laidOutPagesCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a laid-out page"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("project-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Project id")),
				parameters.NewParameterDefinition("template-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Page template id")),
				parameters.NewParameterDefinition("inputs", parameters.ParameterTypeStringList, parameters.WithDefault([]string{}), parameters.WithHelp("Ordered laid-out image ids")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesCreateCommand{}
