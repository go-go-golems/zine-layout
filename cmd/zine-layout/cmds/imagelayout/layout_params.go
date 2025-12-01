package imagelayoutcmd

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout/engine"
)

type layoutParamSettings struct {
	Spec                string  `glazed.parameter:"spec"`
	SourceWidth         int     `glazed.parameter:"source-width"`
	SourceHeight        int     `glazed.parameter:"source-height"`
	FrameMode           string  `glazed.parameter:"frame-mode"`
	FrameRatio          float64 `glazed.parameter:"frame-ratio"`
	FramePageWidth      float64 `glazed.parameter:"frame-page-width-in"`
	FramePageHeight     float64 `glazed.parameter:"frame-page-height-in"`
	FramePageDPI        float64 `glazed.parameter:"frame-page-dpi"`
	FramePageOrient     string  `glazed.parameter:"frame-page-orientation"`
	FrameMarginTop      float64 `glazed.parameter:"frame-margin-top-in"`
	FrameMarginRight    float64 `glazed.parameter:"frame-margin-right-in"`
	FrameMarginBottom   float64 `glazed.parameter:"frame-margin-bottom-in"`
	FrameMarginLeft     float64 `glazed.parameter:"frame-margin-left-in"`
	FrameViewportWidth  float64 `glazed.parameter:"frame-viewport-width"`
	FrameViewportHeight float64 `glazed.parameter:"frame-viewport-height"`
	CropStrategy        string  `glazed.parameter:"crop-strategy"`
	CropRatio           float64 `glazed.parameter:"crop-ratio"`
	CropZoom            float64 `glazed.parameter:"crop-zoom"`
	CropExtent          float64 `glazed.parameter:"crop-extent"`
	CropAnchor          string  `glazed.parameter:"crop-anchor"`
	CropPanX            float64 `glazed.parameter:"crop-pan-x"`
	CropPanY            float64 `glazed.parameter:"crop-pan-y"`
	FocalSourceX        float64 `glazed.parameter:"crop-focus-source-x"`
	FocalSourceY        float64 `glazed.parameter:"crop-focus-source-y"`
	FocalTargetX        float64 `glazed.parameter:"crop-focus-target-x"`
	FocalTargetY        float64 `glazed.parameter:"crop-focus-target-y"`
	CropUnits           string  `glazed.parameter:"crop-units"`
}

func layoutParameterDefinitions() []*parameters.ParameterDefinition {
	return computeParameterDefinitions()
}

func frameVerbParameterDefinitions() []*parameters.ParameterDefinition {
	return joinParameterDefs(
		specParameterDefinitions(),
		frameOnlyParameterDefinitions(),
	)
}

func cropVerbParameterDefinitions() []*parameters.ParameterDefinition {
	return joinParameterDefs(
		specParameterDefinitions(),
		frameOnlyParameterDefinitions(),
		cropOnlyParameterDefinitions(),
	)
}

func computeParameterDefinitions() []*parameters.ParameterDefinition {
	return joinParameterDefs(
		specParameterDefinitions(),
		frameOnlyParameterDefinitions(),
		cropOnlyParameterDefinitions(),
	)
}

func specParameterDefinitions() []*parameters.ParameterDefinition {
	return []*parameters.ParameterDefinition{
		parameters.NewParameterDefinition("spec", parameters.ParameterTypeString, parameters.WithHelp("Path to LayoutRequest spec (YAML/JSON)")),
		parameters.NewParameterDefinition("source-width", parameters.ParameterTypeInteger, parameters.WithHelp("Source image width in pixels")),
		parameters.NewParameterDefinition("source-height", parameters.ParameterTypeInteger, parameters.WithHelp("Source image height in pixels")),
	}
}

func frameOnlyParameterDefinitions() []*parameters.ParameterDefinition {
	return []*parameters.ParameterDefinition{
		parameters.NewParameterDefinition("frame-mode", parameters.ParameterTypeString, parameters.WithHelp("Frame mode: ratio|page|viewport")),
		parameters.NewParameterDefinition("frame-ratio", parameters.ParameterTypeFloat, parameters.WithHelp("Frame ratio when mode=ratio")),
		parameters.NewParameterDefinition("frame-page-width-in", parameters.ParameterTypeFloat, parameters.WithHelp("Page width (inches) when mode=page")),
		parameters.NewParameterDefinition("frame-page-height-in", parameters.ParameterTypeFloat, parameters.WithHelp("Page height (inches) when mode=page")),
		parameters.NewParameterDefinition("frame-page-dpi", parameters.ParameterTypeFloat, parameters.WithHelp("Page DPI when mode=page")),
		parameters.NewParameterDefinition("frame-page-orientation", parameters.ParameterTypeString, parameters.WithHelp("Page orientation (portrait|landscape)")),
		parameters.NewParameterDefinition("frame-margin-top-in", parameters.ParameterTypeFloat, parameters.WithHelp("Top margin in inches")),
		parameters.NewParameterDefinition("frame-margin-right-in", parameters.ParameterTypeFloat, parameters.WithHelp("Right margin in inches")),
		parameters.NewParameterDefinition("frame-margin-bottom-in", parameters.ParameterTypeFloat, parameters.WithHelp("Bottom margin in inches")),
		parameters.NewParameterDefinition("frame-margin-left-in", parameters.ParameterTypeFloat, parameters.WithHelp("Left margin in inches")),
		parameters.NewParameterDefinition("frame-viewport-width", parameters.ParameterTypeFloat, parameters.WithHelp("Viewport width in pixels (mode=viewport)")),
		parameters.NewParameterDefinition("frame-viewport-height", parameters.ParameterTypeFloat, parameters.WithHelp("Viewport height in pixels (mode=viewport)")),
	}
}

func cropOnlyParameterDefinitions() []*parameters.ParameterDefinition {
	return []*parameters.ParameterDefinition{
		parameters.NewParameterDefinition("crop-strategy", parameters.ParameterTypeString, parameters.WithHelp("Crop strategy (auto|anchor|focus|manual)")),
		parameters.NewParameterDefinition("crop-ratio", parameters.ParameterTypeFloat, parameters.WithHelp("Requested crop ratio")),
		parameters.NewParameterDefinition("crop-zoom", parameters.ParameterTypeFloat, parameters.WithHelp("Crop zoom factor (>1 zooms in)")),
		parameters.NewParameterDefinition("crop-extent", parameters.ParameterTypeFloat, parameters.WithHelp("Crop extent (0-1 fraction of source to keep)")),
		parameters.NewParameterDefinition("crop-anchor", parameters.ParameterTypeString, parameters.WithHelp("Anchor preset when strategy=anchor")),
		parameters.NewParameterDefinition("crop-pan-x", parameters.ParameterTypeFloat, parameters.WithHelp("Manual crop pan X (-1..1 or px)")),
		parameters.NewParameterDefinition("crop-pan-y", parameters.ParameterTypeFloat, parameters.WithHelp("Manual crop pan Y (-1..1 or px)")),
		parameters.NewParameterDefinition("crop-focus-source-x", parameters.ParameterTypeFloat, parameters.WithHelp("Focus source X in pixels")),
		parameters.NewParameterDefinition("crop-focus-source-y", parameters.ParameterTypeFloat, parameters.WithHelp("Focus source Y in pixels")),
		parameters.NewParameterDefinition("crop-focus-target-x", parameters.ParameterTypeFloat, parameters.WithHelp("Focus target X (0..1 or px)")),
		parameters.NewParameterDefinition("crop-focus-target-y", parameters.ParameterTypeFloat, parameters.WithHelp("Focus target Y (0..1 or px)")),
		parameters.NewParameterDefinition("crop-units", parameters.ParameterTypeString, parameters.WithHelp("Units for manual pan offsets (normalized|px)")),
	}
}

func joinParameterDefs(groups ...[]*parameters.ParameterDefinition) []*parameters.ParameterDefinition {
	var combined []*parameters.ParameterDefinition
	for _, group := range groups {
		combined = append(combined, group...)
	}
	return combined
}

func buildLayoutFromSettings(parsedLayers *layers.ParsedLayers, settings *layoutParamSettings) (imagelayout.LayoutRequest, imagelayout.ImageMeta, engine.NormalizedInputs, error) {
	layout := imagelayout.DefaultLayoutRequest()
	meta := imagelayout.ImageMeta{}

	if strings.TrimSpace(settings.Spec) != "" {
		loaded, loadedMeta, err := loadLayoutSpec(strings.TrimSpace(settings.Spec))
		if err != nil {
			return imagelayout.LayoutRequest{}, imagelayout.ImageMeta{}, engine.NormalizedInputs{}, err
		}
		layout = loaded
		meta = loadedMeta
	}

	if settings.SourceWidth > 0 {
		meta.Width = settings.SourceWidth
	}
	if settings.SourceHeight > 0 {
		meta.Height = settings.SourceHeight
	}
	if meta.Width <= 0 || meta.Height <= 0 {
		return imagelayout.LayoutRequest{}, imagelayout.ImageMeta{}, engine.NormalizedInputs{}, fmt.Errorf("source-width and source-height must be positive (spec or flags)")
	}

	if parameterSet(parsedLayers, "frame-mode") {
		layout.Frame.Mode = settings.FrameMode
	}
	if parameterSet(parsedLayers, "frame-ratio") && settings.FrameRatio > 0 {
		layout.Frame.Ratio = &settings.FrameRatio
	}

	if parameterSet(parsedLayers, "frame-page-width-in") ||
		parameterSet(parsedLayers, "frame-page-height-in") ||
		parameterSet(parsedLayers, "frame-page-dpi") ||
		parameterSet(parsedLayers, "frame-page-orientation") ||
		parameterSet(parsedLayers, "frame-margin-top-in") ||
		parameterSet(parsedLayers, "frame-margin-right-in") ||
		parameterSet(parsedLayers, "frame-margin-bottom-in") ||
		parameterSet(parsedLayers, "frame-margin-left-in") {
		if layout.Frame.Page == nil {
			layout.Frame.Page = &imagelayout.PageFrame{
				WidthIn:     8.0,
				HeightIn:    10.0,
				DPI:         300,
				Orientation: "portrait",
				MarginsIn: imagelayout.BoxSpacing{
					Top:    0.25,
					Right:  0.25,
					Bottom: 0.25,
					Left:   0.25,
				},
			}
		}
		if parameterSet(parsedLayers, "frame-page-width-in") {
			layout.Frame.Page.WidthIn = settings.FramePageWidth
		}
		if parameterSet(parsedLayers, "frame-page-height-in") {
			layout.Frame.Page.HeightIn = settings.FramePageHeight
		}
		if parameterSet(parsedLayers, "frame-page-dpi") && settings.FramePageDPI > 0 {
			layout.Frame.Page.DPI = settings.FramePageDPI
		}
		if parameterSet(parsedLayers, "frame-page-orientation") && settings.FramePageOrient != "" {
			layout.Frame.Page.Orientation = settings.FramePageOrient
		}
		if parameterSet(parsedLayers, "frame-margin-top-in") {
			layout.Frame.Page.MarginsIn.Top = settings.FrameMarginTop
		}
		if parameterSet(parsedLayers, "frame-margin-right-in") {
			layout.Frame.Page.MarginsIn.Right = settings.FrameMarginRight
		}
		if parameterSet(parsedLayers, "frame-margin-bottom-in") {
			layout.Frame.Page.MarginsIn.Bottom = settings.FrameMarginBottom
		}
		if parameterSet(parsedLayers, "frame-margin-left-in") {
			layout.Frame.Page.MarginsIn.Left = settings.FrameMarginLeft
		}
	}

	if parameterSet(parsedLayers, "frame-viewport-width") ||
		parameterSet(parsedLayers, "frame-viewport-height") {
		if layout.Frame.Viewport == nil {
			layout.Frame.Viewport = &imagelayout.ViewportFrame{}
		}
		if parameterSet(parsedLayers, "frame-viewport-width") {
			layout.Frame.Viewport.Width = settings.FrameViewportWidth
		}
		if parameterSet(parsedLayers, "frame-viewport-height") {
			layout.Frame.Viewport.Height = settings.FrameViewportHeight
		}
	}

	if parameterSet(parsedLayers, "crop-strategy") && settings.CropStrategy != "" {
		layout.Crop.Strategy = settings.CropStrategy
	}
	if parameterSet(parsedLayers, "crop-ratio") && settings.CropRatio > 0 {
		ratio := settings.CropRatio
		layout.Crop.Ratio = &ratio
	}
	if parameterSet(parsedLayers, "crop-zoom") && settings.CropZoom > 0 {
		layout.Crop.Zoom = settings.CropZoom
	}
	if parameterSet(parsedLayers, "crop-extent") && settings.CropExtent > 0 {
		layout.Crop.Extent = settings.CropExtent
	}
	if parameterSet(parsedLayers, "crop-anchor") && settings.CropAnchor != "" {
		layout.Crop.Anchor = settings.CropAnchor
	}
	if parameterSet(parsedLayers, "crop-pan-x") || parameterSet(parsedLayers, "crop-pan-y") {
		layout.Crop.Pan = imagelayout.Vec2{X: settings.CropPanX, Y: settings.CropPanY}
	}
	if parameterSet(parsedLayers, "crop-focus-source-x") || parameterSet(parsedLayers, "crop-focus-source-y") ||
		parameterSet(parsedLayers, "crop-focus-target-x") || parameterSet(parsedLayers, "crop-focus-target-y") {
		if layout.Crop.Focus == nil {
			layout.Crop.Focus = &imagelayout.FocusPoint{}
		}
		if parameterSet(parsedLayers, "crop-focus-source-x") {
			layout.Crop.Focus.SourceX = settings.FocalSourceX
		}
		if parameterSet(parsedLayers, "crop-focus-source-y") {
			layout.Crop.Focus.SourceY = settings.FocalSourceY
		}
		if parameterSet(parsedLayers, "crop-focus-target-x") {
			layout.Crop.Focus.TargetX = settings.FocalTargetX
		}
		if parameterSet(parsedLayers, "crop-focus-target-y") {
			layout.Crop.Focus.TargetY = settings.FocalTargetY
		}
	}
	if parameterSet(parsedLayers, "crop-units") && settings.CropUnits != "" {
		layout.Crop.Units = settings.CropUnits
	}

	inputs, err := engine.InputsFromRequest(layout, meta)
	if err != nil {
		return imagelayout.LayoutRequest{}, imagelayout.ImageMeta{}, engine.NormalizedInputs{}, err
	}
	return layout, meta, inputs, nil
}

func parameterSet(parsedLayers *layers.ParsedLayers, name string) bool {
	_, ok := parsedLayers.GetParameter(layers.DefaultSlug, name)
	return ok
}
