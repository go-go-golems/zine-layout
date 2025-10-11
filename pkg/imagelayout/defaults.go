package imagelayout

// DefaultSettings returns baseline viewport settings used when
// no explicit template values are provided.
func DefaultSettings() ViewportSettings {
	return ViewportSettings{
		PaperWidthIn:   8.0,
		PaperHeightIn:  10.0,
		DPI:            300,
		Orientation:    "portrait",
		MarginTopIn:    0.25,
		MarginRightIn:  0.25,
		MarginBottomIn: 0.25,
		MarginLeftIn:   0.25,
		CropRatio:      nil,
		CropToFill:     false,
		UserScale:      1.0,
		PositionX:      0,
		PositionY:      0,
		Units:          "normalized",
		Export: ExportOptions{
			Format:           "png",
			Quality:          90,
			Background:       "white",
			FilenameTemplate: "{name}-{index}.{ext}",
			OutDir:           "./out",
		},
	}
}
