package zinescmd

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
	"github.com/go-go-golems/zine-layout/pkg/export"
	"github.com/go-go-golems/zine-layout/pkg/services"
	"github.com/spf13/cobra"
)

type zinesExportCommand struct{ *cmds.CommandDescription }

type zinesExportSettings struct {
	DataRoot string  `glazed.parameter:"data-root"`
	ZineID   string  `glazed.parameter:"zine-id"`
	Preset   string  `glazed.parameter:"preset"`
	OutPath  string  `glazed.parameter:"out"`
	DPI      float64 `glazed.parameter:"dpi"`
}

func (c *zinesExportCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &zinesExportSettings{DPI: 300}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.ZineID == "" {
		return fmt.Errorf("--zine-id is required")
	}

	repos, db, err := workflowshared.OpenRepositories(settings.DataRoot)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	impose := services.NewImpositionService(repos)
	impose.SetDataRoot(settings.DataRoot)
	sheets, err := impose.ImposeZine(settings.ZineID, settings.Preset)
	if err != nil {
		return err
	}

	var f *os.File
	if settings.OutPath == "" {
		f = os.Stdout
	} else {
		f, err = os.Create(settings.OutPath)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
	}
	return export.SheetsToPDF(ctx, sheets, settings.DPI, f)
}

func newZinesExportCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	cmd := &zinesExportCommand{
		CommandDescription: cmds.NewCommandDescription(
			"export",
			cmds.WithShort("Export a zine to PDF (imposition preset)"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
				parameters.NewParameterDefinition("zine-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Zine ID")),
				parameters.NewParameterDefinition("preset", parameters.ParameterTypeString, parameters.WithDefault("10_8_sheet_zine"), parameters.WithHelp("Imposition preset id (YAML basename)")),
				parameters.NewParameterDefinition("out", parameters.ParameterTypeString, parameters.WithDefault(""), parameters.WithHelp("Output PDF path (defaults to stdout)")),
				parameters.NewParameterDefinition("dpi", parameters.ParameterTypeFloat, parameters.WithDefault(300.0), parameters.WithHelp("DPI for pixel-to-point conversion")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}
	return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &zinesExportCommand{}
