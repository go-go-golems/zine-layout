package imagelayout

// DefaultLayoutRequest returns a safe baseline LayoutRequest.
func DefaultLayoutRequest() LayoutRequest {
	return LayoutRequest{
		Frame: FrameSpec{
			Mode:  "ratio",
			Ratio: func() *float64 { v := 4.0 / 3.0; return &v }(),
			Fill:  "contain",
		},
		Crop: CropSpec{
			Strategy: "auto",
			Pan:      Vec2{X: 0, Y: 0},
			Zoom:     1.0,
			Units:    "normalized",
		},
		Presentation: PresentationSpec{
			UserScale:     1.0,
			OffsetPx:      Vec2Px{X: 0, Y: 0},
			ClampToCanvas: true,
		},
		Export: ExportOptions{
			Format:           "png",
			Quality:          90,
			Background:       "white",
			FilenameTemplate: "{name}-{panel}.{ext}",
			OutDir:           "./out",
		},
	}
}
