package pagelayoutcmd

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/zine-layout/pkg/app"
	"github.com/go-go-golems/zine-layout/pkg/pagelayout"
	"github.com/go-go-golems/zine-layout/pkg/pagelayout/renderer"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type renderCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*renderCommand)(nil)

type renderSettings struct {
	SourceImage     string  `glazed.parameter:"source-image"`
	Settings        string  `glazed.parameter:"settings"`
	OutputDir       string  `glazed.parameter:"output-dir"`
	Variant         string  `glazed.parameter:"variant"`
	BackgroundColor string  `glazed.parameter:"background-color"`
	ThumbnailMaxPx  int     `glazed.parameter:"thumbnail-max-px"`
	Test            bool    `glazed.parameter:"test"`
	TestWidth       int     `glazed.parameter:"test-width"`
	TestHeight      int     `glazed.parameter:"test-height"`
	// Page settings flags (same as compute)
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

// NewRenderCommand wires the pagelayout render command.
func NewRenderCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "create glazed parameter layer")
	}

	cmd := &renderCommand{
		CommandDescription: cmds.NewCommandDescription(
			"render",
			cmds.WithShort("Render a source image onto a page using page settings"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"source-image",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to source image file (required unless --test)"),
				),
				parameters.NewParameterDefinition(
					"settings",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML/JSON page settings file (or use flags)"),
				),
				parameters.NewParameterDefinition(
					"output-dir",
					parameters.ParameterTypeString,
					parameters.WithDefault("./output"),
					parameters.WithHelp("Output directory for variants"),
				),
				parameters.NewParameterDefinition(
					"variant",
					parameters.ParameterTypeString,
					parameters.WithHelp("Specific variant to generate (thumbnail|full|combined|left|right). If not specified, generates all."),
				),
				parameters.NewParameterDefinition(
					"background-color",
					parameters.ParameterTypeString,
					parameters.WithHelp("Background color (default: white)"),
				),
				parameters.NewParameterDefinition(
					"thumbnail-max-px",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(512),
					parameters.WithHelp("Maximum side length for thumbnail"),
				),
				parameters.NewParameterDefinition(
					"test",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Generate test image instead of reading from file"),
				),
				parameters.NewParameterDefinition(
					"test-width",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(4000),
					parameters.WithHelp("Test image width"),
				),
				parameters.NewParameterDefinition(
					"test-height",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(3000),
					parameters.WithHelp("Test image height"),
				),
				// Page settings flags (same as compute)
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

func (c *renderCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	opts := &renderSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, opts); err != nil {
		return err
	}

	// Parse settings (reuse logic from compute)
	settings, err := parsePageSettings(parsedLayers, opts.Settings, opts)
	if err != nil {
		return fmt.Errorf("parse settings: %w", err)
	}

	// Load or generate source image
	var srcImg image.Image
	if opts.Test {
		testImgs, err := app.GenerateTestImages(1, opts.TestWidth, opts.TestHeight, false)
		if err != nil {
			return fmt.Errorf("generate test image: %w", err)
		}
		if len(testImgs) == 0 {
			return fmt.Errorf("test image generation returned no images")
		}
		srcImg = testImgs[0]
	} else {
		if opts.SourceImage == "" {
			return fmt.Errorf("--source-image is required unless --test is specified")
		}
		f, err := os.Open(opts.SourceImage)
		if err != nil {
			return fmt.Errorf("open source image: %w", err)
		}
		defer f.Close()
		srcImg, _, err = image.Decode(f)
		if err != nil {
			return fmt.Errorf("decode image: %w", err)
		}
	}

	// Parse background color
	var bgColor color.Color = color.White
	if opts.BackgroundColor != "" {
		bgColor = parseColor(opts.BackgroundColor)
	}

	// Create render context
	renderCtx := renderer.RenderContext{
		Settings:       settings,
		Source:         srcImg,
		Background:     bgColor,
		ThumbnailMaxPx: opts.ThumbnailMaxPx,
	}

	// Render
	result, err := renderer.RenderPage(renderCtx)
	if err != nil {
		return fmt.Errorf("render page: %w", err)
	}

	// Save variants
	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	variantsToSave := result.Variants
	if opts.Variant != "" {
		if variant, ok := result.Variants[opts.Variant]; ok {
			variantsToSave = map[string]image.Image{opts.Variant: variant}
		} else {
			return fmt.Errorf("variant %q not found (available: %v)", opts.Variant, getVariantKeys(result.Variants))
		}
	}

	// Save variants and print summary
	fmt.Println("Rendered page variants:")
	for name, img := range variantsToSave {
		if img == nil {
			continue
		}
		path := filepath.Join(opts.OutputDir, name+".png")
		if err := savePNG(img, path); err != nil {
			return fmt.Errorf("save %s: %w", name, err)
		}
		bounds := img.Bounds()
		fmt.Printf("  %s: %s (%dx%d)\n", name, path, bounds.Dx(), bounds.Dy())
	}

	return nil
}

func parsePageSettings(parsedLayers *layers.ParsedLayers, specPath string, opts *renderSettings) (pagelayout.PageLayoutSettings, error) {
	settings := pagelayout.PageLayoutSettings{}

	// Load from spec file if provided
	if specPath != "" {
		data, err := os.ReadFile(specPath)
		if err != nil {
			return settings, fmt.Errorf("read spec: %w", err)
		}
		if err := unmarshalByExt(specPath, data, &settings); err != nil {
			return settings, fmt.Errorf("parse spec: %w", err)
		}
	}

	// Override with flags if provided (same logic as compute)
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

	// Validate
	if err := settings.Canonicalize(); err != nil {
		return settings, fmt.Errorf("invalid settings: %w", err)
	}

	return settings, nil
}

func parseColor(s string) color.Color {
	// Simple color parser - supports #RRGGBB, #RRGGBBAA, or r,g,b,a
	if len(s) > 0 && s[0] == '#' {
		hex := s[1:]
		var r, g, b, a uint8 = 0, 0, 0, 255
		switch len(hex) {
		case 6:
			_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
			if err != nil {
				return color.White
			}
		case 8:
			_, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
			if err != nil {
				return color.White
			}
		}
		if a == 0 {
			a = 255
		}
		return color.RGBA{r, g, b, a}
	}
	// Try comma-separated format
	var r, g, b, a int
	if _, err := fmt.Sscanf(s, "%d,%d,%d,%d", &r, &g, &b, &a); err == nil {
		if a == 0 {
			a = 255
		}
		return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	}
	// Default to white if parsing fails
	return color.White
}

func savePNG(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

func getVariantKeys(variants map[string]image.Image) []string {
	keys := make([]string, 0, len(variants))
	for k := range variants {
		keys = append(keys, k)
	}
	return keys
}

