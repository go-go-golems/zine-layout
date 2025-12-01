package engine

import "github.com/go-go-golems/zine-layout/pkg/imagelayout"

// NormalizedInputs groups the normalized data that flows through the engine.
type NormalizedInputs struct {
	Source       SourceMeta
	Frame        FrameInputs
	Crop         CropInputs
	Presentation PresentationInputs
}

// SourceMeta describes the source image dimensions.
type SourceMeta struct {
	Width  float64
	Height float64
}

// FrameInputs contains the resolved viewport geometry.
type FrameInputs struct {
	Mode       string
	CanvasRect imagelayout.Rect
	Margins    MarginPixels
}

// CropInputs captures the normalized crop configuration.
type CropInputs struct {
	Ratio      *float64
	CropToFill bool
	Zoom       float64
	Extent     float64
	Units      string
	PanX       float64
	PanY       float64
	Focus      *imagelayout.FocusPoint
}

// PresentationInputs stores post-crop adjustments.
type PresentationInputs struct {
	UserScale     float64
	OffsetUnits   string
	OffsetX       float64
	OffsetY       float64
	ClampToCanvas bool
}

// MarginPixels describes resolved margin sizes (in pixels).
type MarginPixels struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}
