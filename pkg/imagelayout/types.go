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
	PaperWidthIn  float64 `json:"paper_width_in"`
	PaperHeightIn float64 `json:"paper_height_in"`
	DPI           float64 `json:"dpi"`
	Orientation   string  `json:"orientation"` // portrait|landscape

	MarginTopIn    float64 `json:"margin_top_in"`
	MarginRightIn  float64 `json:"margin_right_in"`
	MarginBottomIn float64 `json:"margin_bottom_in"`
	MarginLeftIn   float64 `json:"margin_left_in"`

	CropRatio  *float64 `json:"crop_ratio,omitempty"`
	CropToFill bool     `json:"crop_to_fill"`

	UserScale float64 `json:"user_scale"`
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	Units     string  `json:"units"` // normalized|px

	Export ExportOptions `json:"export"`
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

// Computation bundles settings, result, and trace for persistence.
type Computation struct {
	Settings ViewportSettings `json:"settings"`
	Result   ViewportResult   `json:"result"`
	Trace    *Trace           `json:"trace,omitempty"`
}
