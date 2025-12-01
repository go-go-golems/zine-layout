package engine

import "github.com/go-go-golems/zine-layout/pkg/imagelayout"

// FrameAnalysis summarizes frame normalization results.
type FrameAnalysis struct {
	Mode        string           `json:"mode"`
	CanvasRect  imagelayout.Rect `json:"canvas_rect"`
	ContentRect imagelayout.Rect `json:"content_rect"`
	TargetRatio float64          `json:"target_ratio"`
	MarginsPx   MarginPixels     `json:"margins_px"`
}

// CropAnalysis summarizes crop normalization results.
type CropAnalysis struct {
	SourceRect  imagelayout.Rect `json:"source_rect"`
	Diagnostics map[string]any   `json:"diagnostics"`
}

// AnalyzeFrame executes the frame normalization stage.
func AnalyzeFrame(frame FrameInputs) FrameAnalysis {
	canvasRect, contentRect, ratio := buildFrame(frame)
	return FrameAnalysis{
		Mode:        string(frame.Mode),
		CanvasRect:  canvasRect,
		ContentRect: contentRect,
		TargetRatio: ratio,
		MarginsPx: MarginPixels{
			Top:    frame.Margins.Top,
			Right:  frame.Margins.Right,
			Bottom: frame.Margins.Bottom,
			Left:   frame.Margins.Left,
		},
	}
}

// AnalyzeCrop executes the crop normalization stage.
func AnalyzeCrop(source SourceMeta, crop CropInputs, targetRatio float64) CropAnalysis {
	sourceRatio := safeDiv(source.Width, source.Height)
	requestedRatio := determineRequestedRatio(crop, targetRatio, sourceRatio)
	sourceRect, diagnostics := resolveCrop(source, crop, requestedRatio, sourceRatio)
	return CropAnalysis{
		SourceRect:  sourceRect,
		Diagnostics: diagnostics,
	}
}
