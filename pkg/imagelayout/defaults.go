package imagelayout

// DefaultSettings returns baseline viewport settings used when
// no explicit template values are provided.
func DefaultSettings() ViewportSettings {
	return ViewportSettings{
		Mode:           "page",
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
		AnchorPreset:   "center",
		Export: ExportOptions{
			Format:           "png",
			Quality:          90,
			Background:       "white",
			FilenameTemplate: "{name}-{index}.{ext}",
			OutDir:           "./out",
		},
	}
}

// DefaultLayoutRequest returns baseline frame/crop/presentation specs for the
// refactored imagelayout API.
func DefaultLayoutRequest() LayoutRequest {
	ratio := 4.0 / 3.0
	return LayoutRequest{
		Frame: FrameSpec{
			Mode:  "ratio",
			Ratio: &ratio,
			Fill:  "contain",
			Page: &PageFrame{
				WidthIn:     8.0,
				HeightIn:    10.0,
				DPI:         300,
				Orientation: "portrait",
				MarginsIn: BoxSpacing{
					Top:    0.25,
					Right:  0.25,
					Bottom: 0.25,
					Left:   0.25,
				},
			},
		},
		Crop: CropSpec{
			Strategy: "auto",
			Zoom:     1.0,
			Extent:   1.0,
			Pan: Vec2{
				X: 0,
				Y: 0,
			},
			Units: "normalized",
		},
		Presentation: PresentationSpec{
			UserScale: 1.0,
			OffsetPx: Vec2Px{
				X: 0,
				Y: 0,
			},
			ClampToCanvas: true,
		},
		Export: ExportOptions{
			Format:           "png",
			Quality:          90,
			Background:       "white",
			FilenameTemplate: "{name}-{index}.{ext}",
			OutDir:           "./out",
		},
	}
}
