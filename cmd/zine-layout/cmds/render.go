package cmds

import (
	"context"
	"fmt"
	"image"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/zine-layout/pkg/app"
	"github.com/go-go-golems/zine-layout/pkg/zinelayout"
	"github.com/pkg/errors"
)

type RenderCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*RenderCommand)(nil)

func NewRenderCommand() (*RenderCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &RenderCommand{
		CommandDescription: cmds.NewCommandDescription(
			"render",
			cmds.WithShort("Render output pages from a layout spec and input images"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"input-files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Input image files (required unless --test)"),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition("spec", parameters.ParameterTypeString, parameters.WithDefault("layout.yaml"), parameters.WithHelp("Path to the YAML layout specification")),
				parameters.NewParameterDefinition("output-dir", parameters.ParameterTypeString, parameters.WithDefault("."), parameters.WithHelp("Directory to save output images")),
				parameters.NewParameterDefinition("verbose", parameters.ParameterTypeBool, parameters.WithDefault(false), parameters.WithHelp("Verbose output")),
				parameters.NewParameterDefinition("global-border", parameters.ParameterTypeBool, parameters.WithDefault(false), parameters.WithHelp("Enable global border")),
				parameters.NewParameterDefinition("page-border", parameters.ParameterTypeBool, parameters.WithDefault(false), parameters.WithHelp("Enable page border")),
				parameters.NewParameterDefinition("layout-border", parameters.ParameterTypeBool, parameters.WithDefault(false), parameters.WithHelp("Enable layout border")),
				parameters.NewParameterDefinition("inner-border", parameters.ParameterTypeBool, parameters.WithDefault(false), parameters.WithHelp("Enable inner layout border")),
				parameters.NewParameterDefinition("border-color", parameters.ParameterTypeString, parameters.WithDefault(""), parameters.WithHelp("Border color R,G,B,A (0-255) or color name or #hex")),
				parameters.NewParameterDefinition("border-type", parameters.ParameterTypeChoice, parameters.WithChoices("plain", "dotted", "dashed", "corner"), parameters.WithDefault(""), parameters.WithHelp("Border type")),
				parameters.NewParameterDefinition("test", parameters.ParameterTypeBool, parameters.WithDefault(false), parameters.WithHelp("Generate test images instead of reading inputs")),
				parameters.NewParameterDefinition("test-bw", parameters.ParameterTypeBool, parameters.WithDefault(false), parameters.WithHelp("Use black and white test images")),
				parameters.NewParameterDefinition("test-dimensions", parameters.ParameterTypeString, parameters.WithDefault(""), parameters.WithHelp("Test image size: 'WIDTH,HEIGHT' (e.g. 600px,800px)")),
				parameters.NewParameterDefinition("ppi", parameters.ParameterTypeInteger, parameters.WithDefault(0), parameters.WithHelp("Override layout PPI")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

type RenderSettings struct {
	InputFiles     []string `glazed.parameter:"input-files"`
	Spec           string   `glazed.parameter:"spec"`
	OutputDir      string   `glazed.parameter:"output-dir"`
	Verbose        bool     `glazed.parameter:"verbose"`
	GlobalBorder   bool     `glazed.parameter:"global-border"`
	PageBorder     bool     `glazed.parameter:"page-border"`
	LayoutBorder   bool     `glazed.parameter:"layout-border"`
	InnerBorder    bool     `glazed.parameter:"inner-border"`
	BorderColor    string   `glazed.parameter:"border-color"`
	BorderType     string   `glazed.parameter:"border-type"`
	Test           bool     `glazed.parameter:"test"`
	TestBW         bool     `glazed.parameter:"test-bw"`
	TestDimensions string   `glazed.parameter:"test-dimensions"`
	PPI            int      `glazed.parameter:"ppi"`
}

func (c *RenderCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &RenderSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	if !s.Test && len(s.InputFiles) == 0 {
		return fmt.Errorf("no input files provided; pass --test or specify input files")
	}

	// Load layouts
	env := map[string]interface{}{}
	layouts, err := app.LoadLayoutsFromSpec(s.Spec, env)
	if err != nil {
		return err
	}

	for _, zl := range layouts {
		if err := app.ApplyOverrides(&zl, app.Overrides{
			GlobalBorder: s.GlobalBorder,
			PageBorder:   s.PageBorder,
			LayoutBorder: s.LayoutBorder,
			InnerBorder:  s.InnerBorder,
			BorderColor:  s.BorderColor,
			BorderType:   s.BorderType,
			PPI:          s.PPI,
		}); err != nil {
			return err
		}

		ppi := zl.Global.PPI
		if ppi == 0 {
			ppi = 300
		}

		// Prepare inputs
		var inputImages []image.Image
		if s.Test {
			maxIndex := 0
			for _, op := range zl.OutputPages {
				for _, l := range op.Layout {
					if l.InputIndex > maxIndex {
						maxIndex = l.InputIndex
					}
				}
			}
			w, h, err := app.ParseTestDimensions(s.TestDimensions, ppi)
			if err != nil {
				return err
			}
			inputImages, err = app.GenerateTestImages(maxIndex, w, h, s.TestBW)
			if err != nil {
				return err
			}
		} else {
			inputImages, err = app.ReadInputImages(s.InputFiles)
			if err != nil {
				return err
			}
		}

		if !zinelayout.AllImagesSameSize(inputImages) {
			return fmt.Errorf("input images are not the same size")
		}

		if s.Verbose {
			fmt.Println("Parsed ZineLayout:")
			app.DebugPrintZineLayout(zl)
			fmt.Println()
		}

		written, err := app.RenderOutputs(&zl, inputImages, s.OutputDir)
		if err != nil {
			return err
		}
		for _, fp := range written {
			if fi, err := os.Stat(fp); err == nil {
				fmt.Printf("Saved output image: %s (Size: %d bytes)\n", fp, fi.Size())
			}
		}
	}

	return nil
}
