package imagelayout

// DefaultLayoutRequest returns a safe baseline LayoutRequest.
func DefaultLayoutRequest() LayoutRequest {
	return LayoutRequest{
		Frame: FrameSpec{
			Mode:  "ratio",
			Ratio: func() *float64 { v := 4.0 / 3.0; return &v }(),
		},
		Crop: CropSpec{
			Strategy: "auto",
			Pan:      Vec2{X: 0, Y: 0},
			Zoom:     1.0,
			Units:    "normalized",
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
