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

// ViewportSettings represents the canonical input for positioning an image inside a viewport.
type ViewportSettings struct {
	Mode string `json:"mode,omitempty"` // page | crop | fit

	PaperWidthIn  float64 `json:"paper_width_in"`
	PaperHeightIn float64 `json:"paper_height_in"`
	DPI           float64 `json:"dpi"`
	Orientation   string  `json:"orientation"` // portrait|landscape

	MarginTopIn    float64 `json:"margin_top_in"`
	MarginRightIn  float64 `json:"margin_right_in"`
	MarginBottomIn float64 `json:"margin_bottom_in"`
	MarginLeftIn   float64 `json:"margin_left_in"`

	CropRatio    *float64 `json:"crop_ratio,omitempty"`
	CropToFill   bool     `json:"crop_to_fill"`
	CropWidthPx  *float64 `json:"crop_width_px,omitempty"`
	CropHeightPx *float64 `json:"crop_height_px,omitempty"`

	FitMode     string   `json:"fit_mode,omitempty"`      // width|height|auto
	FitWidthPx  *float64 `json:"fit_width_px,omitempty"`  // target width in pixels
	FitHeightPx *float64 `json:"fit_height_px,omitempty"` // target height in pixels

	UserScale float64 `json:"user_scale"`
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	Units     string  `json:"units"` // normalized|px

	AnchorPreset string      `json:"anchor_preset,omitempty"` // e.g. center, top-right
	Focus        *FocusPoint `json:"focus,omitempty"`

	Export ExportOptions `json:"export"`
}

// LayoutRequest separates frame, crop, and presentation concerns for the
// refactored imagelayout pipeline.
type LayoutRequest struct {
	Frame        FrameSpec        `json:"frame"`
	Crop         CropSpec         `json:"crop"`
	Presentation PresentationSpec `json:"presentation"`
	Export       ExportOptions    `json:"export"`
}

// FrameSpec defines the destination box or aspect ratio before cropping.
type FrameSpec struct {
	Mode     string         `json:"mode"` // ratio | page | viewport
	Ratio    *float64       `json:"ratio,omitempty"`
	Fill     string         `json:"fill,omitempty"` // contain | cover
	Page     *PageFrame     `json:"page,omitempty"`
	Viewport *ViewportFrame `json:"viewport,omitempty"`
	FitAxis  string         `json:"fit_axis,omitempty"` // width | height | auto
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

// Vec2Px stores offsets in absolute pixels for presentation adjustments.
type Vec2Px struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// PresentationSpec captures user adjustments applied after frame+crop.
type PresentationSpec struct {
	UserScale     float64 `json:"user_scale"`
	OffsetPx      Vec2Px  `json:"offset_px"`
	ClampToCanvas bool    `json:"clamp_to_canvas"`
}

// ViewportResult describes the geometry outcome of placement.
type ViewportResult struct {
	SourceRect Rect    `json:"source_rect"`
	TargetRect Rect    `json:"target_rect"`
	CanvasRect Rect    `json:"canvas_rect"`
	Scale      float64 `json:"scale"`
	Mode       string  `json:"mode"` // cover|contain
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
