package imagelayout

// Rect describes a rectangle in floating point coordinates.
type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// ImageMeta captures source image properties when the file is not directly available.
type ImageMeta struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ExportOptions defines preferences for downstream rendering.
type ExportOptions struct {
	Format           string `json:"format"`
	Quality          int    `json:"quality"`
	Background       string `json:"background"`
	FilenameTemplate string `json:"filename_template"`
	OutDir           string `json:"out_dir"`
}

// LayoutRequest separates frame, crop, and presentation concerns for the
// refactored imagelayout pipeline.
type LayoutRequest struct {
	Frame  FrameSpec     `json:"frame"`
	Crop   CropSpec      `json:"crop"`
	Export ExportOptions `json:"export"`
}

// FrameSpec defines the destination box or aspect ratio before cropping.
type FrameSpec struct {
	Mode     string         `json:"mode"` // ratio | page | viewport
	Ratio    *float64       `json:"ratio,omitempty"`
	Page     *PageFrame     `json:"page,omitempty"`
	Viewport *ViewportFrame `json:"viewport,omitempty"`
}

// PageFrame captures physical media dimensions and margins (in inches).
type PageFrame struct {
	WidthIn     float64    `json:"width_in"`
	HeightIn    float64    `json:"height_in"`
	DPI         float64    `json:"dpi"`
	Orientation string     `json:"orientation,omitempty"` // portrait | landscape
	MarginsIn   BoxSpacing `json:"margins_in"`
}

// BoxSpacing represents the padding applied on each edge of a rectangle.
type BoxSpacing struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// ViewportFrame represents an explicit viewport in pixels (web/editor use).
type ViewportFrame struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// FocusPoint aligns a source coordinate with a target location in the viewport.
type FocusPoint struct {
	SourceX float64 `json:"source_x"`
	SourceY float64 `json:"source_y"`
	TargetX float64 `json:"target_x"`
	TargetY float64 `json:"target_y"`
}

// CropSpec declares how to cut the source image to match the requested frame.
type CropSpec struct {
	Strategy string      `json:"strategy"`         // auto | focus | anchor | manual
	Ratio    *float64    `json:"ratio,omitempty"`  // overrides frame ratio when set
	Zoom     float64     `json:"zoom,omitempty"`   // >1 zooms in, <1 zooms out
	Extent   float64     `json:"extent,omitempty"` // fraction (0,1] of the source to keep
	Anchor   string      `json:"anchor,omitempty"` // named anchor (top-left, center, etc.)
	Pan      Vec2        `json:"pan"`              // normalized pan offsets (-1..1)
	Focus    *FocusPoint `json:"focus,omitempty"`  // target-aware focus controls
	Units    string      `json:"units,omitempty"`  // normalized | px (for pan/focus targets)
}

// Vec2 stores normalized coordinates (-1..1) for pan/offset calculations.
type Vec2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ViewportResult describes the geometry outcome of placement.
type ViewportResult struct {
	SourceRect Rect    `json:"source_rect"`
	TargetRect Rect    `json:"target_rect"`
	CanvasRect Rect    `json:"canvas_rect"`
	Scale      float64 `json:"scale"`
}

// TraceStep captures intermediate calculations for debugging.
type TraceStep struct {
	Label string                 `json:"label"`
	Data  map[string]interface{} `json:"data"`
}

// Trace records human-readable diagnostics for placement.
type Trace struct {
	Inputs map[string]interface{} `json:"inputs"`
	Steps  []TraceStep            `json:"steps"`
}

// Computation bundles the layout request, result, and trace for persistence.
type Computation struct {
	Layout LayoutRequest  `json:"layout"`
	Result ViewportResult `json:"result"`
	Trace  *Trace         `json:"trace,omitempty"`
}
