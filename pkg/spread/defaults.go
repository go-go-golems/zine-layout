package spread

// DefaultSettings returns the canonical defaults for the Simple algorithm.
func DefaultSettings() Settings {
	return Settings{
		PaperWidthIn:   8.0,
		PaperHeightIn:  10.0,
		DPI:            300,
		Orientation:    "portrait",
		MarginTopIn:    0.25,
		MarginRightIn:  0.25,
		MarginBottomIn: 0.25,
		MarginLeftIn:   0.25,
		IsSpread:       false,
		GutterIn:       0,
		CropRatio:      nil,
		CropToFill:     false,
		UserScale:      1.0,
		PositionX:      0,
		PositionY:      0,
		Units:          "normalized",
		Export: Export{
			Format:           "png",
			Quality:          90,
			Background:       "white",
			OutDir:           "./out",
			FilenameTemplate: "{name}-{panel}.{ext}",
		},
	}
}
