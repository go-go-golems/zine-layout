package imagelayoutcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout/engine"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type specDocument struct {
	Settings imagelayout.ViewportSettings `json:"settings" yaml:"settings"`
	Image    struct {
		Width  int `json:"width" yaml:"width"`
		Height int `json:"height" yaml:"height"`
	} `json:"image" yaml:"image"`
}

// NewCommand returns the `imagelayout` command group.
func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "imagelayout",
		Short: "Local image layout helpers",
	}
	root.AddCommand(newComputeCommand())
	return root
}

func newComputeCommand() *cobra.Command {
	var (
		specPath string
		metaW    int
		metaH    int
	)

	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Compute viewport placement from settings or YAML spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			settings := imagelayout.DefaultSettings()
			meta := imagelayout.ImageMeta{}

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

			if cmd.Flags().Changed("source-width") {
				metaW, _ = cmd.Flags().GetInt("source-width")
				meta.Width = metaW
			}
			if cmd.Flags().Changed("source-height") {
				metaH, _ = cmd.Flags().GetInt("source-height")
				meta.Height = metaH
			}

			if meta.Width <= 0 || meta.Height <= 0 {
				return fmt.Errorf("source-width and source-height must be positive")
			}

			applyFloatOverride(cmd, "paper-width-in", &settings.PaperWidthIn)
			applyFloatOverride(cmd, "paper-height-in", &settings.PaperHeightIn)
			applyFloatOverride(cmd, "dpi", &settings.DPI)
			applyFloatOverride(cmd, "margin-top-in", &settings.MarginTopIn)
			applyFloatOverride(cmd, "margin-right-in", &settings.MarginRightIn)
			applyFloatOverride(cmd, "margin-bottom-in", &settings.MarginBottomIn)
			applyFloatOverride(cmd, "margin-left-in", &settings.MarginLeftIn)
			applyFloatOverride(cmd, "user-scale", &settings.UserScale)
			applyFloatOverride(cmd, "position-x", &settings.PositionX)
			applyFloatOverride(cmd, "position-y", &settings.PositionY)

			if cmd.Flags().Changed("mode") {
				settings.Mode, _ = cmd.Flags().GetString("mode")
			}
			if cmd.Flags().Changed("orientation") {
				settings.Orientation, _ = cmd.Flags().GetString("orientation")
			}
			if cmd.Flags().Changed("units") {
				settings.Units, _ = cmd.Flags().GetString("units")
			}
			if cmd.Flags().Changed("anchor-preset") {
				settings.AnchorPreset, _ = cmd.Flags().GetString("anchor-preset")
			}
			if cmd.Flags().Changed("crop-to-fill") {
				settings.CropToFill, _ = cmd.Flags().GetBool("crop-to-fill")
			}
			if cmd.Flags().Changed("crop-ratio") {
				val, _ := cmd.Flags().GetFloat64("crop-ratio")
				settings.CropRatio = floatPtr(val)
			}
			if cmd.Flags().Changed("crop-width") {
				val, _ := cmd.Flags().GetFloat64("crop-width")
				settings.CropWidthPx = floatPtr(val)
			}
			if cmd.Flags().Changed("crop-height") {
				val, _ := cmd.Flags().GetFloat64("crop-height")
				settings.CropHeightPx = floatPtr(val)
			}
			if cmd.Flags().Changed("fit-mode") {
				settings.FitMode, _ = cmd.Flags().GetString("fit-mode")
			}
			if cmd.Flags().Changed("fit-width") {
				val, _ := cmd.Flags().GetFloat64("fit-width")
				settings.FitWidthPx = floatPtr(val)
			}
			if cmd.Flags().Changed("fit-height") {
				val, _ := cmd.Flags().GetFloat64("fit-height")
				settings.FitHeightPx = floatPtr(val)
			}
			if cmd.Flags().Changed("focus-source-x") || cmd.Flags().Changed("focus-source-y") ||
				cmd.Flags().Changed("focus-target-x") || cmd.Flags().Changed("focus-target-y") {
				if settings.Focus == nil {
					settings.Focus = &imagelayout.FocusPoint{}
				}
				if cmd.Flags().Changed("focus-source-x") {
					val, _ := cmd.Flags().GetFloat64("focus-source-x")
					settings.Focus.SourceX = val
				}
				if cmd.Flags().Changed("focus-source-y") {
					val, _ := cmd.Flags().GetFloat64("focus-source-y")
					settings.Focus.SourceY = val
				}
				if cmd.Flags().Changed("focus-target-x") {
					val, _ := cmd.Flags().GetFloat64("focus-target-x")
					settings.Focus.TargetX = val
				}
				if cmd.Flags().Changed("focus-target-y") {
					val, _ := cmd.Flags().GetFloat64("focus-target-y")
					settings.Focus.TargetY = val
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
		},
	}

	cmd.Flags().StringVar(&specPath, "spec", "", "Path to YAML/JSON specification")
	cmd.Flags().IntVar(&metaW, "source-width", 0, "Source image width in pixels")
	cmd.Flags().IntVar(&metaH, "source-height", 0, "Source image height in pixels")

	cmd.Flags().String("mode", "", "Layout mode: page|crop|fit")
	cmd.Flags().String("orientation", "", "Page orientation override")
	cmd.Flags().Float64("paper-width-in", 0, "Paper width in inches")
	cmd.Flags().Float64("paper-height-in", 0, "Paper height in inches")
	cmd.Flags().Float64("dpi", 0, "Dots per inch")
	cmd.Flags().Float64("margin-top-in", 0, "Top margin in inches")
	cmd.Flags().Float64("margin-right-in", 0, "Right margin in inches")
	cmd.Flags().Float64("margin-bottom-in", 0, "Bottom margin in inches")
	cmd.Flags().Float64("margin-left-in", 0, "Left margin in inches")
	cmd.Flags().Bool("crop-to-fill", false, "Enable cover mode")
	cmd.Flags().Float64("crop-ratio", 0, "Crop aspect ratio (width/height)")
	cmd.Flags().Float64("crop-width", 0, "Output crop width in pixels (crop mode)")
	cmd.Flags().Float64("crop-height", 0, "Output crop height in pixels (crop mode)")
	cmd.Flags().String("fit-mode", "", "Fit mode: width|height|auto")
	cmd.Flags().Float64("fit-width", 0, "Target width in pixels (fit mode)")
	cmd.Flags().Float64("fit-height", 0, "Target height in pixels (fit mode)")
	cmd.Flags().Float64("user-scale", 0, "Additional user scale multiplier")
	cmd.Flags().Float64("position-x", 0, "Position offset X (-1..1 or px)")
	cmd.Flags().Float64("position-y", 0, "Position offset Y (-1..1 or px)")
	cmd.Flags().String("units", "", "Position units: normalized|px")
	cmd.Flags().String("anchor-preset", "", "Anchor preset (e.g. center, top-left)")
	cmd.Flags().Float64("focus-source-x", 0, "Focus source X (pixels)")
	cmd.Flags().Float64("focus-source-y", 0, "Focus source Y (pixels)")
	cmd.Flags().Float64("focus-target-x", 0, "Focus target X (0..1 or pixels)")
	cmd.Flags().Float64("focus-target-y", 0, "Focus target Y (0..1 or pixels)")

	return cmd
}

func unmarshalByExt(path string, data []byte, out any) error {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, out)
	case ".json":
		return json.Unmarshal(data, out)
	default:
		// try yaml first then json
		if err := yaml.Unmarshal(data, out); err == nil {
			return nil
		}
		return json.Unmarshal(data, out)
	}
}

func applyFloatOverride(cmd *cobra.Command, flag string, target *float64) {
	if cmd.Flags().Changed(flag) {
		val, _ := cmd.Flags().GetFloat64(flag)
		*target = val
	}
}

func floatPtr(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}
