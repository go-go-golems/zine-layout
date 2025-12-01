package imagelayoutcmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout/engine"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newLayoutCommand() *cobra.Command {
	layoutCmd := &cobra.Command{
		Use:   "layout",
		Short: "Inspect LayoutRequest normalization stages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	frameCmd, err := newLayoutFrameCmd()
	if err != nil {
		panic(err)
	}
	layoutCmd.AddCommand(frameCmd)

	cropCmd, err := newLayoutCropCmd()
	if err != nil {
		panic(err)
	}
	layoutCmd.AddCommand(cropCmd)
	return layoutCmd
}

type layoutFrameGlazeCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*layoutFrameGlazeCommand)(nil)

func newLayoutFrameCmd() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "create glazed parameter layer")
	}
	cmd := &layoutFrameGlazeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"layout-frame",
			cmds.WithShort("Analyze frame inputs (use --spec or flags)"),
			cmds.WithFlags(frameVerbParameterDefinitions()...),
			cmds.WithLayersList(glazedLayer),
		),
	}
	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd, cli.WithParserConfig(cli.CobraParserConfig{
		ShortHelpLayers: []string{layers.DefaultSlug},
		MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
	}))
	if err != nil {
		return nil, err
	}
	cobraCmd.Use = "frame"
	cobraCmd.Short = "Analyze frame normalization"
	return cobraCmd, nil
}

func (c *layoutFrameGlazeCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &layoutParamSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	layout, meta, inputs, err := buildLayoutFromSettings(parsedLayers, settings)
	if err != nil {
		return err
	}
	_ = layout
	_ = meta
	analysis := engine.AnalyzeFrame(inputs.Frame)
	return printJSON(analysis)
}

type layoutCropGlazeCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*layoutCropGlazeCommand)(nil)

func newLayoutCropCmd() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "create glazed parameter layer")
	}
	cmd := &layoutCropGlazeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"layout-crop",
			cmds.WithShort("Analyze crop normalization (use --spec or flags)"),
			cmds.WithFlags(cropVerbParameterDefinitions()...),
			cmds.WithLayersList(glazedLayer),
		),
	}
	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd, cli.WithParserConfig(cli.CobraParserConfig{
		ShortHelpLayers: []string{layers.DefaultSlug},
		MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
	}))
	if err != nil {
		return nil, err
	}
	cobraCmd.Use = "crop"
	cobraCmd.Short = "Analyze crop normalization"
	return cobraCmd, nil
}

func (c *layoutCropGlazeCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &layoutParamSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}
	layout, meta, inputs, err := buildLayoutFromSettings(parsedLayers, settings)
	if err != nil {
		return err
	}
	_ = layout
	_ = meta
	frameAnalysis := engine.AnalyzeFrame(inputs.Frame)
	analysis := engine.AnalyzeCrop(inputs.Source, inputs.Crop, frameAnalysis.TargetRatio)
	return printJSON(analysis)
}

func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
