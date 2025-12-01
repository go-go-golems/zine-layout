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
	Spec         string `glazed.parameter:"spec"`
	SourceWidth  int    `glazed.parameter:"source-width"`
	SourceHeight int    `glazed.parameter:"source-height"`
}

type specDocument struct {
	Layout imagelayout.LayoutRequest `json:"layout" yaml:"layout"`
	Image  struct {
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
			cmds.WithShort("Compute viewport placement from a LayoutRequest spec"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"spec",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML/JSON imagelayout spec"),
				),
				parameters.NewParameterDefinition(
					"source-width",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Override source image width in pixels"),
				),
				parameters.NewParameterDefinition(
					"source-height",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Override source image height in pixels"),
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

	if strings.TrimSpace(opts.Spec) == "" {
		return fmt.Errorf("--spec is required")
	}
	layout := imagelayout.DefaultLayoutRequest()
	meta := imagelayout.ImageMeta{}

	specPath := strings.TrimSpace(opts.Spec)
	data, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("read spec: %w", err)
	}
	var doc specDocument
	if err := unmarshalByExt(specPath, data, &doc); err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}
	layout = doc.Layout
	meta.Width = doc.Image.Width
	meta.Height = doc.Image.Height

	if parameterSet(parsedLayers, "source-width") {
		meta.Width = opts.SourceWidth
	}
	if parameterSet(parsedLayers, "source-height") {
		meta.Height = opts.SourceHeight
	}
	if meta.Width <= 0 || meta.Height <= 0 {
		return fmt.Errorf("source-width and source-height must be positive")
	}

	inputs, err := engine.InputsFromRequest(layout, meta)
	if err != nil {
		return err
	}

	result, trace := engine.ComputeViewport(inputs)
	output := imagelayout.Computation{
		Layout: layout,
		Result: result,
		Trace:  trace,
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
