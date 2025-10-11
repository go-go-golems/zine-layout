package imagelayoutcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout/engine"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type computeCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*computeCommand)(nil)

type computeSettings struct {
	Spec           string  `glazed.parameter:"spec"`
	SourceWidth    int     `glazed.parameter:"source-width"`
	SourceHeight   int     `glazed.parameter:"source-height"`
	Mode           string  `glazed.parameter:"mode"`
	Orientation    string  `glazed.parameter:"orientation"`
	Units          string  `glazed.parameter:"units"`
	AnchorPreset   string  `glazed.parameter:"anchor-preset"`
	PaperWidthIn   float64 `glazed.parameter:"paper-width-in"`
	PaperHeightIn  float64 `glazed.parameter:"paper-height-in"`
	DPI            float64 `glazed.parameter:"dpi"`
	MarginTopIn    float64 `glazed.parameter:"margin-top-in"`
	MarginRightIn  float64 `glazed.parameter:"margin-right-in"`
	MarginBottomIn float64 `glazed.parameter:"margin-bottom-in"`
	MarginLeftIn   float64 `glazed.parameter:"margin-left-in"`
	CropToFill     bool    `glazed.parameter:"crop-to-fill"`
	CropRatio      float64 `glazed.parameter:"crop-ratio"`
	CropWidth      float64 `glazed.parameter:"crop-width"`
	CropHeight     float64 `glazed.parameter:"crop-height"`
	FitMode        string  `glazed.parameter:"fit-mode"`
	FitWidth       float64 `glazed.parameter:"fit-width"`
	FitHeight      float64 `glazed.parameter:"fit-height"`
	UserScale      float64 `glazed.parameter:"user-scale"`
	PositionX      float64 `glazed.parameter:"position-x"`
	PositionY      float64 `glazed.parameter:"position-y"`
	FocusSourceX   float64 `glazed.parameter:"focus-source-x"`
	FocusSourceY   float64 `glazed.parameter:"focus-source-y"`
	FocusTargetX   float64 `glazed.parameter:"focus-target-x"`
	FocusTargetY   float64 `glazed.parameter:"focus-target-y"`
}

type specDocument struct {
	Settings imagelayout.ViewportSettings `json:"settings" yaml:"settings"`
	Image    struct {
		Width  int `json:"width" yaml:"width"`
		Height int `json:"height" yaml:"height"`
	} `json:"image" yaml:"image"`
}

// NewComputeCommand wires the imagelayout compute command.
func NewComputeCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "create glazed parameter layer")
	}

	cmd := &computeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"compute",
			cmds.WithShort("Compute viewport placement from settings or YAML spec"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"spec",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML/JSON imagelayout spec"),
				),
				parameters.NewParameterDefinition(
					"source-width",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Source image width in pixels"),
				),
				parameters.NewParameterDefinition(
					"source-height",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Source image height in pixels"),
				),
				parameters.NewParameterDefinition("mode", parameters.ParameterTypeString, parameters.WithHelp("Layout mode: page|crop|fit")),
				parameters.NewParameterDefinition("orientation", parameters.ParameterTypeString, parameters.WithHelp("Page orientation override")),
				parameters.NewParameterDefinition("units", parameters.ParameterTypeString, parameters.WithHelp("Position units: normalized|px")),
				parameters.NewParameterDefinition("anchor-preset", parameters.ParameterTypeString, parameters.WithHelp("Anchor preset (e.g. center, top-left)")),
				parameters.NewParameterDefinition("paper-width-in", parameters.ParameterTypeFloat, parameters.WithHelp("Paper width in inches")),
				parameters.NewParameterDefinition("paper-height-in", parameters.ParameterTypeFloat, parameters.WithHelp("Paper height in inches")),
				parameters.NewParameterDefinition("dpi", parameters.ParameterTypeFloat, parameters.WithHelp("Dots per inch")),
				parameters.NewParameterDefinition("margin-top-in", parameters.ParameterTypeFloat, parameters.WithHelp("Top margin in inches")),
				parameters.NewParameterDefinition("margin-right-in", parameters.ParameterTypeFloat, parameters.WithHelp("Right margin in inches")),
				parameters.NewParameterDefinition("margin-bottom-in", parameters.ParameterTypeFloat, parameters.WithHelp("Bottom margin in inches")),
				parameters.NewParameterDefinition("margin-left-in", parameters.ParameterTypeFloat, parameters.WithHelp("Left margin in inches")),
				parameters.NewParameterDefinition("crop-to-fill", parameters.ParameterTypeBool, parameters.WithHelp("Enable cover mode")),
				parameters.NewParameterDefinition("crop-ratio", parameters.ParameterTypeFloat, parameters.WithHelp("Crop aspect ratio (width/height)")),
				parameters.NewParameterDefinition("crop-width", parameters.ParameterTypeFloat, parameters.WithHelp("Output crop width in pixels")),
				parameters.NewParameterDefinition("crop-height", parameters.ParameterTypeFloat, parameters.WithHelp("Output crop height in pixels")),
				parameters.NewParameterDefinition("fit-mode", parameters.ParameterTypeString, parameters.WithHelp("Fit mode: width|height|auto")),
				parameters.NewParameterDefinition("fit-width", parameters.ParameterTypeFloat, parameters.WithHelp("Target width in pixels (fit mode)")),
				parameters.NewParameterDefinition("fit-height", parameters.ParameterTypeFloat, parameters.WithHelp("Target height in pixels (fit mode)")),
				parameters.NewParameterDefinition("user-scale", parameters.ParameterTypeFloat, parameters.WithHelp("Additional user scale multiplier")),
				parameters.NewParameterDefinition("position-x", parameters.ParameterTypeFloat, parameters.WithHelp("Position offset X (-1..1 or px)")),
				parameters.NewParameterDefinition("position-y", parameters.ParameterTypeFloat, parameters.WithHelp("Position offset Y (-1..1 or px)")),
				parameters.NewParameterDefinition("focus-source-x", parameters.ParameterTypeFloat, parameters.WithHelp("Focus source X in pixels")),
				parameters.NewParameterDefinition("focus-source-y", parameters.ParameterTypeFloat, parameters.WithHelp("Focus source Y in pixels")),
				parameters.NewParameterDefinition("focus-target-x", parameters.ParameterTypeFloat, parameters.WithHelp("Focus target X (0..1 or pixels)")),
				parameters.NewParameterDefinition("focus-target-y", parameters.ParameterTypeFloat, parameters.WithHelp("Focus target Y (0..1 or pixels)")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return cli.BuildCobraCommandFromCommand(
		cmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
}

func (c *computeCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	opts := &computeSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, opts); err != nil {
		return err
	}

	settings := imagelayout.DefaultSettings()
	meta := imagelayout.ImageMeta{}

	specPath := strings.TrimSpace(opts.Spec)
	if specPath != "" {
		data, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("read spec: %w", err)
		}
		var doc specDocument
		if err := unmarshalByExt(specPath, data, &doc); err != nil {
			return fmt.Errorf("parse spec: %w", err)
		}
		settings = doc.Settings
		meta.Width = doc.Image.Width
		meta.Height = doc.Image.Height
	}

	if parameterSet(parsedLayers, "source-width") {
		meta.Width = opts.SourceWidth
	}
	if parameterSet(parsedLayers, "source-height") {
		meta.Height = opts.SourceHeight
	}
	if meta.Width <= 0 || meta.Height <= 0 {
		return fmt.Errorf("source-width and source-height must be positive")
	}

	if parameterSet(parsedLayers, "paper-width-in") {
		settings.PaperWidthIn = opts.PaperWidthIn
	}
	if parameterSet(parsedLayers, "paper-height-in") {
		settings.PaperHeightIn = opts.PaperHeightIn
	}
	if parameterSet(parsedLayers, "dpi") {
		settings.DPI = opts.DPI
	}
	if parameterSet(parsedLayers, "margin-top-in") {
		settings.MarginTopIn = opts.MarginTopIn
	}
	if parameterSet(parsedLayers, "margin-right-in") {
		settings.MarginRightIn = opts.MarginRightIn
	}
	if parameterSet(parsedLayers, "margin-bottom-in") {
		settings.MarginBottomIn = opts.MarginBottomIn
	}
	if parameterSet(parsedLayers, "margin-left-in") {
		settings.MarginLeftIn = opts.MarginLeftIn
	}
	if parameterSet(parsedLayers, "user-scale") {
		settings.UserScale = opts.UserScale
	}
	if parameterSet(parsedLayers, "position-x") {
		settings.PositionX = opts.PositionX
	}
	if parameterSet(parsedLayers, "position-y") {
		settings.PositionY = opts.PositionY
	}

	if parameterSet(parsedLayers, "mode") {
		settings.Mode = opts.Mode
	}
	if parameterSet(parsedLayers, "orientation") {
		settings.Orientation = opts.Orientation
	}
	if parameterSet(parsedLayers, "units") {
		settings.Units = opts.Units
	}
	if parameterSet(parsedLayers, "anchor-preset") {
		settings.AnchorPreset = opts.AnchorPreset
	}
	if parameterSet(parsedLayers, "crop-to-fill") {
		settings.CropToFill = opts.CropToFill
	}
	if parameterSet(parsedLayers, "crop-ratio") {
		settings.CropRatio = floatPtr(opts.CropRatio)
	}
	if parameterSet(parsedLayers, "crop-width") {
		settings.CropWidthPx = floatPtr(opts.CropWidth)
	}
	if parameterSet(parsedLayers, "crop-height") {
		settings.CropHeightPx = floatPtr(opts.CropHeight)
	}
	if parameterSet(parsedLayers, "fit-mode") {
		settings.FitMode = opts.FitMode
	}
	if parameterSet(parsedLayers, "fit-width") {
		settings.FitWidthPx = floatPtr(opts.FitWidth)
	}
	if parameterSet(parsedLayers, "fit-height") {
		settings.FitHeightPx = floatPtr(opts.FitHeight)
	}

	focusProvided := parameterSet(parsedLayers, "focus-source-x") ||
		parameterSet(parsedLayers, "focus-source-y") ||
		parameterSet(parsedLayers, "focus-target-x") ||
		parameterSet(parsedLayers, "focus-target-y")

	if focusProvided {
		if settings.Focus == nil {
			settings.Focus = &imagelayout.FocusPoint{}
		}
		if parameterSet(parsedLayers, "focus-source-x") {
			settings.Focus.SourceX = opts.FocusSourceX
		}
		if parameterSet(parsedLayers, "focus-source-y") {
			settings.Focus.SourceY = opts.FocusSourceY
		}
		if parameterSet(parsedLayers, "focus-target-x") {
			settings.Focus.TargetX = opts.FocusTargetX
		}
		if parameterSet(parsedLayers, "focus-target-y") {
			settings.Focus.TargetY = opts.FocusTargetY
		}
	}

	inputs, err := engine.InputsFromSettings(settings, meta)
	if err != nil {
		return err
	}

	result, trace := engine.ComputeViewport(inputs)
	output := imagelayout.Computation{
		Settings: settings,
		Result:   result,
		Trace:    trace,
	}
	pretty, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(pretty))
	return nil
}

func parameterSet(parsedLayers *layers.ParsedLayers, name string) bool {
	_, ok := parsedLayers.GetParameter(layers.DefaultSlug, name)
	return ok
}

func unmarshalByExt(path string, data []byte, out any) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, out)
	case ".json":
		return json.Unmarshal(data, out)
	default:
		return fmt.Errorf("unsupported spec extension: %s", ext)
	}
}

func floatPtr(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}
