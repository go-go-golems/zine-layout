package spread

// ImageMeta describes source image dimensions when the server cannot read the file.
type ImageMeta struct {
    Width  int `json:"width"`
    Height int `json:"height"`
}

// Export defines output preferences.
type Export struct {
    Format           string `json:"format"`
    Quality          int    `json:"quality"`
    Background       string `json:"background"`
    OutDir           string `json:"out_dir"`
    FilenameTemplate string `json:"filename_template"`
}

// Settings are engine-agnostic resolved inputs for computing a spread.
type Settings struct {
    PaperWidthIn  float64 `json:"paper_width_in"`
    PaperHeightIn float64 `json:"paper_height_in"`
    DPI           float64 `json:"dpi"`
    Orientation   string  `json:"orientation"` // portrait|landscape

    MarginTopIn    float64 `json:"margin_top_in"`
    MarginRightIn  float64 `json:"margin_right_in"`
    MarginBottomIn float64 `json:"margin_bottom_in"`
    MarginLeftIn   float64 `json:"margin_left_in"`

    IsSpread bool    `json:"is_spread"`
    GutterIn float64 `json:"gutter_in"`

    // CropRatio is a numeric aspect ratio (w/h). Nil => original.
    CropRatio  *float64 `json:"crop_ratio"`
    CropToFill bool     `json:"crop_to_fill"`

    UserScale float64 `json:"user_scale"`

    PositionX float64 `json:"position_x"`
    PositionY float64 `json:"position_y"`
    Units     string  `json:"units"` // px|normalized

    Export Export `json:"export"`
}

// ComputeRequest requests geometry computation for a spread.
type ComputeRequest struct {
    Algorithm string   `json:"algorithm"`
    ImagePath string   `json:"image_path,omitempty"`
    Meta      *ImageMeta `json:"meta,omitempty"`
    Name      string   `json:"name,omitempty"`
    Settings  Settings `json:"settings"`
}


