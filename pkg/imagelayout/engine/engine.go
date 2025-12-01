package engine

import (
	"math"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
)

func pointerValue(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

// ComputeViewport calculates viewport placement and trace data.
func ComputeViewport(inp NormalizedInputs) (imagelayout.ViewportResult, *imagelayout.Trace) {
	trace := &imagelayout.Trace{
		Inputs: map[string]interface{}{
			"source": map[string]float64{
				"width":  inp.Source.Width,
				"height": inp.Source.Height,
			},
			"frame": map[string]interface{}{
				"mode":        inp.Frame.Mode,
				"canvas_rect": inp.Frame.CanvasRect,
				"margins_px":  inp.Frame.Margins,
			},
			"crop": map[string]interface{}{
				"ratio":        pointerValue(inp.Crop.Ratio),
				"crop_to_fill": inp.Crop.CropToFill,
				"zoom":         inp.Crop.Zoom,
				"extent":       inp.Crop.Extent,
				"units":        inp.Crop.Units,
				"pan_x":        inp.Crop.PanX,
				"pan_y":        inp.Crop.PanY,
				"focus":        inp.Crop.Focus,
			},
			"presentation": map[string]interface{}{
				"user_scale":      inp.Presentation.UserScale,
				"offset_units":    inp.Presentation.OffsetUnits,
				"offset_x":        inp.Presentation.OffsetX,
				"offset_y":        inp.Presentation.OffsetY,
				"clamp_to_canvas": inp.Presentation.ClampToCanvas,
			},
		},
	}

	addStep := func(label string, data map[string]interface{}) {
		trace.Steps = append(trace.Steps, imagelayout.TraceStep{
			Label: label,
			Data:  data,
		})
	}

	canvasRect, contentRect, targetRatio := buildFrame(inp.Frame)
	sourceRatio := safeDiv(inp.Source.Width, inp.Source.Height)
	requestedRatio := determineRequestedRatio(inp.Crop, targetRatio, sourceRatio)

	sourceRect, cropStep := resolveCrop(inp.Source, inp.Crop, requestedRatio, sourceRatio)
	addStep("crop", cropStep)

	targetRect, mode, scale, scaleStep := composeTarget(inp.Crop, inp.Presentation, contentRect, sourceRect)
	addStep("scale", scaleStep)

	result := imagelayout.ViewportResult{
		SourceRect: sourceRect,
		TargetRect: targetRect,
		CanvasRect: canvasRect,
		Scale:      scale,
		Mode:       mode,
	}

	addStep("result", map[string]interface{}{
		"source_rect":  sourceRect,
		"target_rect":  targetRect,
		"canvas_rect":  canvasRect,
		"content_rect": contentRect,
	})

	return result, trace
}

func buildFrame(frame FrameInputs) (imagelayout.Rect, imagelayout.Rect, float64) {
	canvas := frame.CanvasRect
	content := frame.ContentRect
	if content.W == 0 || content.H == 0 {
		content = canvas
	}
	return canvas, content, safeDiv(content.W, content.H)
}

func determineRequestedRatio(crop CropInputs, targetRatio, sourceRatio float64) float64 {
	if crop.Ratio != nil && *crop.Ratio > 0 {
		return *crop.Ratio
	}
	if crop.CropToFill && targetRatio > 0 {
		return targetRatio
	}
	return sourceRatio
}

func resolveCrop(source SourceMeta, crop CropInputs, requestedRatio, sourceRatio float64) (imagelayout.Rect, map[string]interface{}) {
	sw := source.Width
	sh := source.Height

	if requestedRatio > 0 && source.Width > 0 && source.Height > 0 {
		switch {
		case sourceRatio > requestedRatio:
			sh = source.Height
			sw = sh * requestedRatio
		case sourceRatio < requestedRatio:
			sw = source.Width
			sh = sw / requestedRatio
		default:
			// keep source dimensions
		}
	}

	scale := computeCropScale(crop.Extent, crop.Zoom)
	sw = clampFloat(sw*scale, 1, source.Width)
	sh = clampFloat(sh*scale, 1, source.Height)

	rangeX := math.Max(0, source.Width-sw)
	rangeY := math.Max(0, source.Height-sh)

	sx := 0.0
	sy := 0.0
	focusApplied := false
	var focusInfo map[string]interface{}

	if crop.Focus != nil {
		focusApplied = true
		focusInfo = map[string]interface{}{}
		fx := clampFloat(crop.Focus.SourceX, 0, source.Width)
		fy := clampFloat(crop.Focus.SourceY, 0, source.Height)
		targetNX := resolveFocusTarget(crop.Focus.TargetX, sw)
		targetNY := resolveFocusTarget(crop.Focus.TargetY, sh)
		sx = clampFloat(fx-targetNX*sw, 0, rangeX)
		sy = clampFloat(fy-targetNY*sh, 0, rangeY)
		focusInfo["focus_source_x"] = fx
		focusInfo["focus_source_y"] = fy
		focusInfo["focus_target_x"] = targetNX
		focusInfo["focus_target_y"] = targetNY
	} else {
		sx = computeOffset(crop.Units, crop.PanX, rangeX)
		sy = computeOffset(crop.Units, crop.PanY, rangeY)
	}

	data := map[string]interface{}{
		"sx":              sx,
		"sy":              sy,
		"sw":              sw,
		"sh":              sh,
		"source_ratio":    sourceRatio,
		"requested_ratio": requestedRatio,
		"range_x":         rangeX,
		"range_y":         rangeY,
		"focus_applied":   focusApplied,
		"crop_zoom":       crop.Zoom,
		"crop_extent":     crop.Extent,
	}
	if focusApplied {
		data["focus"] = focusInfo
	}

	return imagelayout.Rect{X: sx, Y: sy, W: sw, H: sh}, data
}

func computeCropScale(extent, zoom float64) float64 {
	scale := 1.0
	if extent > 0 && extent <= 1 {
		scale *= extent
	}
	if zoom > 0 {
		scale /= zoom
	}
	if scale <= 0 {
		return 1.0
	}
	return scale
}

func composeTarget(crop CropInputs, presentation PresentationInputs, contentRect, sourceRect imagelayout.Rect) (imagelayout.Rect, string, float64, map[string]interface{}) {
	targetW := contentRect.W
	targetH := contentRect.H

	scaleX := safeDiv(targetW, sourceRect.W)
	scaleY := safeDiv(targetH, sourceRect.H)

	mode := "contain"
	scale := math.Min(scaleX, scaleY)
	if crop.CropToFill {
		mode = "cover"
		scale = math.Max(scaleX, scaleY)
	}
	scale *= presentation.UserScale

	dstW := sourceRect.W * scale
	dstH := sourceRect.H * scale
	offsetUnits := presentation.OffsetUnits
	if offsetUnits == "" {
		offsetUnits = crop.Units
	}
	tx, ty := positionOffsets(offsetUnits, presentation.OffsetX, presentation.OffsetY, targetW, targetH, dstW, dstH)

	targetRect := imagelayout.Rect{
		X: contentRect.X + tx,
		Y: contentRect.Y + ty,
		W: dstW,
		H: dstH,
	}

	data := map[string]interface{}{
		"scale_x":            scaleX,
		"scale_y":            scaleY,
		"final":              scale,
		"mode":               mode,
		"dst_w":              dstW,
		"dst_h":              dstH,
		"tx":                 tx,
		"ty":                 ty,
		"presentation_units": offsetUnits,
	}

	return targetRect, mode, scale, data
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func computeOffset(units string, value float64, rangePx float64) float64 {
	if rangePx <= 0 {
		return 0
	}
	if units == "normalized" {
		clamped := clampFloat(value, -1, 1)
		return (clamped + 1) * 0.5 * rangePx
	}
	half := rangePx / 2
	return clampFloat(value, -half, half) + half
}

func positionOffsets(units string, px, py float64, targetW, targetH, dstW, dstH float64) (float64, float64) {
	if units == "px" {
		return px, py
	}
	freeX := targetW - dstW
	freeY := targetH - dstH
	tx := (clampFloat(px, -1, 1) * freeX) / 2
	ty := (clampFloat(py, -1, 1) * freeY) / 2
	return tx, ty
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func resolveFocusTarget(value, length float64) float64 {
	if length <= 0 {
		return 0.5
	}
	if value < 0 || value > 1 {
		return clampFloat(value/length, 0, 1)
	}
	return clampFloat(value, 0, 1)
}
