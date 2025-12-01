package engine

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
)

var anchorPresets = map[string][2]float64{
	"top-left":     {-1, -1},
	"top":          {0, -1},
	"top-right":    {1, -1},
	"left":         {-1, 0},
	"center":       {0, 0},
	"right":        {1, 0},
	"bottom-left":  {-1, 1},
	"bottom":       {0, 1},
	"bottom-right": {1, 1},
}

// InputsFromSettings converts persisted settings and image metadata into algorithm inputs.
func InputsFromSettings(settings imagelayout.ViewportSettings, meta imagelayout.ImageMeta) (NormalizedInputs, error) {
	if meta.Width <= 0 || meta.Height <= 0 {
		return NormalizedInputs{}, fmt.Errorf("imagelayout: invalid source dimensions %dx%d", meta.Width, meta.Height)
	}
	if settings.DPI <= 0 {
		return NormalizedInputs{}, fmt.Errorf("imagelayout: dpi must be positive")
	}

	mode := strings.ToLower(strings.TrimSpace(settings.Mode))
	if mode == "" {
		mode = "page"
	}

	sourceRatio := float64(meta.Width) / float64(meta.Height)

	widthIn := settings.PaperWidthIn
	heightIn := settings.PaperHeightIn
	if strings.ToLower(settings.Orientation) == "landscape" {
		widthIn, heightIn = settings.PaperHeightIn, settings.PaperWidthIn
	}

	canvasW := widthIn * settings.DPI
	canvasH := heightIn * settings.DPI
	if canvasW <= 0 || canvasH <= 0 {
		return NormalizedInputs{}, fmt.Errorf("imagelayout: canvas dimensions must be positive")
	}

	mt := settings.MarginTopIn * settings.DPI
	mr := settings.MarginRightIn * settings.DPI
	mb := settings.MarginBottomIn * settings.DPI
	ml := settings.MarginLeftIn * settings.DPI

	contentW := canvasW - (ml + mr)
	contentH := canvasH - (mt + mb)
	if contentW <= 0 || contentH <= 0 {
		return NormalizedInputs{}, fmt.Errorf("imagelayout: margins exceed canvas size")
	}

	units := strings.ToLower(strings.TrimSpace(settings.Units))
	if units == "" {
		units = "normalized"
	}
	if units != "normalized" && units != "px" {
		return NormalizedInputs{}, fmt.Errorf("imagelayout: unsupported position units %q", settings.Units)
	}

	var ratio *float64
	if settings.CropRatio != nil {
		if *settings.CropRatio <= 0 {
			return NormalizedInputs{}, fmt.Errorf("imagelayout: crop ratio must be > 0")
		}
		value := *settings.CropRatio
		ratio = &value
	}

	scale := settings.UserScale
	if scale <= 0 {
		scale = 1.0
	}

	cropWidth := pointerValue(settings.CropWidthPx)
	cropHeight := pointerValue(settings.CropHeightPx)
	fitWidth := pointerValue(settings.FitWidthPx)
	fitHeight := pointerValue(settings.FitHeightPx)
	fitMode := strings.ToLower(strings.TrimSpace(settings.FitMode))

	switch mode {
	case "crop":
		if cropWidth <= 0 {
			if settings.CropRatio != nil && *settings.CropRatio > 0 && cropHeight > 0 {
				cropWidth = cropHeight * *settings.CropRatio
			} else {
				cropWidth = float64(meta.Width)
			}
		}
		if cropHeight <= 0 {
			if settings.CropRatio != nil && *settings.CropRatio > 0 {
				cropHeight = cropWidth / *settings.CropRatio
			} else {
				cropHeight = float64(meta.Height)
			}
		}
		canvasW = cropWidth
		canvasH = cropHeight
		contentW = cropWidth
		contentH = cropHeight
		mt, mr, mb, ml = 0, 0, 0, 0
	case "fit":
		switch fitMode {
		case "width":
			if fitWidth <= 0 {
				fitWidth = contentW
			}
			if fitHeight <= 0 {
				fitHeight = fitWidth / sourceRatio
			}
		case "height":
			if fitHeight <= 0 {
				fitHeight = contentH
			}
			if fitWidth <= 0 {
				fitWidth = fitHeight * sourceRatio
			}
		default:
			if fitWidth <= 0 && fitHeight > 0 {
				fitWidth = fitHeight * sourceRatio
			}
			if fitHeight <= 0 && fitWidth > 0 {
				fitHeight = fitWidth / sourceRatio
			}
			if fitWidth <= 0 {
				fitWidth = contentW
			}
			if fitHeight <= 0 {
				fitHeight = contentH
			}
		}
		canvasW = fitWidth
		canvasH = fitHeight
		contentW = fitWidth
		contentH = fitHeight
		mt, mr, mb, ml = 0, 0, 0, 0
	default:
		// page mode keeps calculated canvas/content values
	}

	if canvasW <= 0 || canvasH <= 0 {
		return NormalizedInputs{}, fmt.Errorf("imagelayout: resulting canvas dimensions must be positive")
	}

	frameRect := imagelayout.Rect{
		X: ml,
		Y: mt,
		W: contentW,
		H: contentH,
	}
	if mode != "page" {
		frameRect.X = 0
		frameRect.Y = 0
	}

	margins := MarginPixels{
		Top:    mt,
		Right:  mr,
		Bottom: mb,
		Left:   ml,
	}
	if mode != "page" {
		margins = MarginPixels{}
	}

	posX := resolveAnchor(settings, units, settings.PositionX, true)
	posY := resolveAnchor(settings, units, settings.PositionY, false)

	return NormalizedInputs{
		Source: SourceMeta{
			Width:  float64(meta.Width),
			Height: float64(meta.Height),
		},
		Frame: FrameInputs{
			Mode:       mode,
			CanvasRect: frameRect,
			Margins:    margins,
		},
		Crop: CropInputs{
			Ratio:      ratio,
			CropToFill: settings.CropToFill,
			Zoom:       1.0,
			Extent:     1.0,
			Units:      units,
			PanX:       posX,
			PanY:       posY,
			Focus:      settings.Focus,
		},
		Presentation: PresentationInputs{
			UserScale:     scale,
			OffsetUnits:   units,
			OffsetX:       posX,
			OffsetY:       posY,
			ClampToCanvas: true,
		},
	}, nil
}

func pointerValue(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func resolveAnchor(settings imagelayout.ViewportSettings, units string, fallback float64, isX bool) float64 {
	if units != "normalized" {
		return fallback
	}
	preset := strings.ToLower(strings.TrimSpace(settings.AnchorPreset))
	if preset == "" {
		return fallback
	}
	vec, ok := anchorPresets[preset]
	if !ok {
		return fallback
	}
	if isX {
		return vec[0]
	}
	return vec[1]
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

	canvasRect, targetRatio := buildFrame(inp.Frame)
	sourceRatio := safeDiv(inp.Source.Width, inp.Source.Height)
	requestedRatio := determineRequestedRatio(inp.Crop, targetRatio, sourceRatio)

	sourceRect, cropStep := resolveCrop(inp.Source, inp.Crop, requestedRatio, sourceRatio)
	addStep("crop", cropStep)

	targetRect, mode, scale, scaleStep := composeTarget(inp.Crop, inp.Presentation, canvasRect, sourceRect)
	addStep("scale", scaleStep)

	result := imagelayout.ViewportResult{
		SourceRect: sourceRect,
		TargetRect: targetRect,
		CanvasRect: canvasRect,
		Scale:      scale,
		Mode:       mode,
	}

	addStep("result", map[string]interface{}{
		"source_rect": sourceRect,
		"target_rect": targetRect,
		"canvas_rect": canvasRect,
	})

	return result, trace
}

func buildFrame(frame FrameInputs) (imagelayout.Rect, float64) {
	rect := frame.CanvasRect
	return rect, safeDiv(rect.W, rect.H)
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

func composeTarget(crop CropInputs, presentation PresentationInputs, canvasRect, sourceRect imagelayout.Rect) (imagelayout.Rect, string, float64, map[string]interface{}) {
	targetW := canvasRect.W
	targetH := canvasRect.H

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
		X: canvasRect.X + tx,
		Y: canvasRect.Y + ty,
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
