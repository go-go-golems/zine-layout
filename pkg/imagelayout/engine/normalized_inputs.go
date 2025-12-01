package engine

import "github.com/go-go-golems/zine-layout/pkg/imagelayout"

// FrameMode enumerates supported frame modes.
type FrameMode string

const (
	FrameModeRatio    FrameMode = "ratio"
	FrameModePage     FrameMode = "page"
	FrameModeViewport FrameMode = "viewport"
)

// NormalizedInputs groups the normalized data that flows through the engine.
type NormalizedInputs struct {
	Source SourceMeta
	Frame  FrameInputs
	Crop   CropInputs
}

// SourceMeta describes the source image dimensions.
type SourceMeta struct {
	Width  float64
	Height float64
}

// FrameInputs contains the resolved viewport geometry.
type FrameInputs struct {
	Mode        FrameMode
	CanvasRect  imagelayout.Rect
	ContentRect imagelayout.Rect
	Margins     MarginPixels
}

// CropInputs captures the normalized crop configuration.
type CropInputs struct {
	Ratio  *float64
	Zoom   float64
	Extent float64
	Units  string
	PanX   float64
	PanY   float64
	Focus  *imagelayout.FocusPoint
}

// MarginPixels describes resolved margin sizes (in pixels).
type MarginPixels struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}
