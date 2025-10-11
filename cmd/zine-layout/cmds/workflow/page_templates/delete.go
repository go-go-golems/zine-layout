package pagetemplatescmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/spf13/cobra"
)

type pageTemplatesDeleteCommand struct {
	*cmds.CommandDescription
}

type pageTemplatesDeleteSettings struct {
	DataRoot   string `glazed.parameter:"data-root"`
	TemplateID string `glazed.parameter:"page-template-id"`
}

func (c *pageTemplatesDeleteCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &pageTemplatesDeleteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if strings.TrimSpace(settings.TemplateID) == "" {
		return fmt.Errorf("page-template-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := repos.PageTemplates.Delete(settings.TemplateID); err != nil {
		return err
	}

	return nil
}

func newPageTemplatesDeleteCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmd := &pageTemplatesDeleteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a page template"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-root",
					parameters.ParameterTypeString,
					parameters.WithDefault("./data"),
					parameters.WithHelp("Path to data directory"),
				),
				parameters.NewParameterDefinition(
					"page-template-id",
					parameters.ParameterTypeString,
					parameters.WithHelp("Template identifier"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &pageTemplatesDeleteCommand{}
