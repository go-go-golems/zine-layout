package engine

import "github.com/go-go-golems/zine-layout/pkg/imagelayout"

// FrameAnalysis summarizes frame normalization results.
type FrameAnalysis struct {
	Mode          string           `json:"mode"`
	CanvasRect    imagelayout.Rect `json:"canvas_rect"`
	ContentRect   imagelayout.Rect `json:"content_rect"`
	TargetRatio   float64          `json:"target_ratio"`
	MarginsPx     MarginPixels     `json:"margins_px"`
	ClampToCanvas bool             `json:"clamp_to_canvas"`
}

// CropAnalysis summarizes crop normalization results.
type CropAnalysis struct {
	SourceRect  imagelayout.Rect `json:"source_rect"`
	Diagnostics map[string]any   `json:"diagnostics"`
}

// PresentationAnalysis summarizes presentation normalization results.
type PresentationAnalysis struct {
	TargetRect  imagelayout.Rect `json:"target_rect"`
	Scale       float64          `json:"scale"`
	Mode        string           `json:"mode"`
	Diagnostics map[string]any   `json:"diagnostics"`
}

// AnalyzeFrame executes the frame normalization stage.
func AnalyzeFrame(inp NormalizedInputs) FrameAnalysis {
	canvasRect, ratio := buildFrame(inp.Frame)
	contentRect := imagelayout.Rect{
		X: 0,
		Y: 0,
		W: canvasRect.W,
		H: canvasRect.H,
	}
	return FrameAnalysis{
		Mode:        inp.Frame.Mode,
		CanvasRect:  canvasRect,
		ContentRect: contentRect,
		TargetRatio: ratio,
		MarginsPx: MarginPixels{
			Top:    inp.Frame.Margins.Top,
			Right:  inp.Frame.Margins.Right,
			Bottom: inp.Frame.Margins.Bottom,
			Left:   inp.Frame.Margins.Left,
		},
		ClampToCanvas: inp.Presentation.ClampToCanvas,
	}
}

// AnalyzeCrop executes the crop normalization stage.
func AnalyzeCrop(inp NormalizedInputs, targetRatio float64) CropAnalysis {
	sourceRatio := safeDiv(inp.Source.Width, inp.Source.Height)
	requestedRatio := determineRequestedRatio(inp.Crop, targetRatio, sourceRatio)
	sourceRect, diagnostics := resolveCrop(inp.Source, inp.Crop, requestedRatio, sourceRatio)
	return CropAnalysis{
		SourceRect:  sourceRect,
		Diagnostics: diagnostics,
	}
}

// AnalyzePresentation executes the presentation stage.
func AnalyzePresentation(inp NormalizedInputs, canvasRect, sourceRect imagelayout.Rect) PresentationAnalysis {
	targetRect, mode, scale, diagnostics := composeTarget(inp.Crop, inp.Presentation, canvasRect, sourceRect)
	return PresentationAnalysis{
		TargetRect:  targetRect,
		Scale:       scale,
		Mode:        mode,
		Diagnostics: diagnostics,
	}
}
