package pagelayoutcmd

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
	"github.com/go-go-golems/zine-layout/pkg/pagelayout"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type computeCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*computeCommand)(nil)

type computeSettings struct {
	Spec            string  `glazed.parameter:"spec"`
	PageWidthIn     float64 `glazed.parameter:"page-width-in"`
	PageHeightIn    float64 `glazed.parameter:"page-height-in"`
	DPI             float64 `glazed.parameter:"dpi"`
	MarginTopIn     float64 `glazed.parameter:"margin-top-in"`
	MarginRightIn   float64 `glazed.parameter:"margin-right-in"`
	MarginBottomIn  float64 `glazed.parameter:"margin-bottom-in"`
	MarginLeftIn    float64 `glazed.parameter:"margin-left-in"`
	IsSpread        bool    `glazed.parameter:"is-spread"`
	GutterWidthIn   float64 `glazed.parameter:"gutter-width-in"`
	GutterOverlapIn float64 `glazed.parameter:"gutter-overlap-in"`
	PositioningMode string  `glazed.parameter:"positioning-mode"`
	AnchorPreset    string  `glazed.parameter:"anchor-preset"`
	ImageXIn        float64 `glazed.parameter:"image-x-in"`
	ImageYIn        float64 `glazed.parameter:"image-y-in"`
	ImageWidthIn    float64 `glazed.parameter:"image-width-in"`
	ImageHeightIn   float64 `glazed.parameter:"image-height-in"`
	BorderEnabled   bool    `glazed.parameter:"border-enabled"`
	BorderColor     string  `glazed.parameter:"border-color"`
	BorderType      string  `glazed.parameter:"border-type"`
}

type computeOutput struct {
	Settings pagelayout.PageLayoutSettings `json:"settings"`
	Metrics  struct {
		PixelWidth   int                `json:"pixelWidth"`
		PixelHeight  int                `json:"pixelHeight"`
		ContentRect  contentRectOutput  `json:"contentRect"`
		SpreadSplitX int                `json:"spreadSplitX"`
	} `json:"metrics"`
	Valid bool `json:"valid"`
}

type contentRectOutput struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// NewComputeCommand wires the pagelayout compute command.
func NewComputeCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "create glazed parameter layer")
	}

	cmd := &computeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"compute",
			cmds.WithShort("Compute page metrics and validate settings"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"spec",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML/JSON page settings file"),
				),
				parameters.NewParameterDefinition(
					"page-width-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Page width in inches"),
				),
				parameters.NewParameterDefinition(
					"page-height-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Page height in inches"),
				),
				parameters.NewParameterDefinition(
					"dpi",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Dots per inch"),
				),
				parameters.NewParameterDefinition(
					"margin-top-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Top margin in inches"),
				),
				parameters.NewParameterDefinition(
					"margin-right-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Right margin in inches"),
				),
				parameters.NewParameterDefinition(
					"margin-bottom-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Bottom margin in inches"),
				),
				parameters.NewParameterDefinition(
					"margin-left-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Left margin in inches"),
				),
				parameters.NewParameterDefinition(
					"is-spread",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Enable spread mode (two-page spread)"),
				),
				parameters.NewParameterDefinition(
					"gutter-width-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Gutter width in inches (for spreads)"),
				),
				parameters.NewParameterDefinition(
					"gutter-overlap-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Gutter overlap in inches (for spreads)"),
				),
				parameters.NewParameterDefinition(
					"positioning-mode",
					parameters.ParameterTypeString,
					parameters.WithHelp("Positioning mode: fill|absolute|snap"),
				),
				parameters.NewParameterDefinition(
					"anchor-preset",
					parameters.ParameterTypeString,
					parameters.WithHelp("Anchor preset (reserved for future use)"),
				),
				parameters.NewParameterDefinition(
					"image-x-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Image X position in inches (absolute mode)"),
				),
				parameters.NewParameterDefinition(
					"image-y-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Image Y position in inches (absolute mode)"),
				),
				parameters.NewParameterDefinition(
					"image-width-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Image width in inches (absolute mode)"),
				),
				parameters.NewParameterDefinition(
					"image-height-in",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("Image height in inches (absolute mode)"),
				),
				parameters.NewParameterDefinition(
					"border-enabled",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Enable page border"),
				),
				parameters.NewParameterDefinition(
					"border-color",
					parameters.ParameterTypeString,
					parameters.WithHelp("Border color (#RRGGBB, #RRGGBBAA, or r,g,b,a)"),
				),
				parameters.NewParameterDefinition(
					"border-type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Border type: plain|dotted|dashed|corner"),
				),
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

	settings := pagelayout.PageLayoutSettings{}

	// Load from spec file if provided
	specPath := strings.TrimSpace(opts.Spec)
	if specPath != "" {
		data, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("read spec: %w", err)
		}
		if err := unmarshalByExt(specPath, data, &settings); err != nil {
			return fmt.Errorf("parse spec: %w", err)
		}
	}

	// Override with flags if provided
	if parameterSet(parsedLayers, "page-width-in") {
		settings.PageWidthIn = opts.PageWidthIn
	}
	if parameterSet(parsedLayers, "page-height-in") {
		settings.PageHeightIn = opts.PageHeightIn
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
	if parameterSet(parsedLayers, "is-spread") {
		settings.IsSpread = opts.IsSpread
	}
	if parameterSet(parsedLayers, "gutter-width-in") {
		settings.GutterWidthIn = opts.GutterWidthIn
	}
	if parameterSet(parsedLayers, "gutter-overlap-in") {
		settings.GutterOverlapIn = opts.GutterOverlapIn
	}
	if parameterSet(parsedLayers, "positioning-mode") {
		settings.PositioningMode = opts.PositioningMode
	}
	if parameterSet(parsedLayers, "anchor-preset") {
		settings.AnchorPreset = opts.AnchorPreset
	}
	if parameterSet(parsedLayers, "image-x-in") {
		settings.ImageXIn = opts.ImageXIn
	}
	if parameterSet(parsedLayers, "image-y-in") {
		settings.ImageYIn = opts.ImageYIn
	}
	if parameterSet(parsedLayers, "image-width-in") {
		settings.ImageWidthIn = opts.ImageWidthIn
	}
	if parameterSet(parsedLayers, "image-height-in") {
		settings.ImageHeightIn = opts.ImageHeightIn
	}
	if parameterSet(parsedLayers, "border-enabled") {
		settings.BorderEnabled = opts.BorderEnabled
	}
	if parameterSet(parsedLayers, "border-color") {
		settings.BorderColor = opts.BorderColor
	}
	if parameterSet(parsedLayers, "border-type") {
		settings.BorderType = opts.BorderType
	}

	// Validate settings
	if err := settings.Canonicalize(); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	// Compute metrics
	contentRect := settings.ContentRectPx()
	output := computeOutput{
		Settings: settings,
		Valid:    true,
	}
	output.Metrics.PixelWidth = settings.PixelWidth()
	output.Metrics.PixelHeight = settings.PixelHeight()
	output.Metrics.ContentRect.X = contentRect.Min.X
	output.Metrics.ContentRect.Y = contentRect.Min.Y
	output.Metrics.ContentRect.Width = contentRect.Dx()
	output.Metrics.ContentRect.Height = contentRect.Dy()
	output.Metrics.SpreadSplitX = settings.SpreadSplitX()

	// Output JSON
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

