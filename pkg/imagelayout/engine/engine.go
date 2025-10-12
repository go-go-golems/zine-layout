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

// Inputs collects normalized values derived from settings + image metadata.
type Inputs struct {
	Mode string

	SourceW float64
	SourceH float64

	CanvasW float64
	CanvasH float64

	ContentW float64
	ContentH float64

	MarginTopPx    float64
	MarginRightPx  float64
	MarginBottomPx float64
	MarginLeftPx   float64

	CropRatio    *float64
	CropToFill   bool
	CropWidthPx  float64
	CropHeightPx float64

	FitMode     string
	FitWidthPx  float64
	FitHeightPx float64

	UserScale float64
	PositionX float64
	PositionY float64
	Units     string

	Focus *imagelayout.FocusPoint
}

// InputsFromSettings converts persisted settings and image metadata into algorithm inputs.
func InputsFromSettings(settings imagelayout.ViewportSettings, meta imagelayout.ImageMeta) (Inputs, error) {
	if meta.Width <= 0 || meta.Height <= 0 {
		return Inputs{}, fmt.Errorf("imagelayout: invalid source dimensions %dx%d", meta.Width, meta.Height)
	}
	if settings.DPI <= 0 {
		return Inputs{}, fmt.Errorf("imagelayout: dpi must be positive")
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
		return Inputs{}, fmt.Errorf("imagelayout: canvas dimensions must be positive")
	}

	mt := settings.MarginTopIn * settings.DPI
	mr := settings.MarginRightIn * settings.DPI
	mb := settings.MarginBottomIn * settings.DPI
	ml := settings.MarginLeftIn * settings.DPI

	contentW := canvasW - (ml + mr)
	contentH := canvasH - (mt + mb)
	if contentW <= 0 || contentH <= 0 {
		return Inputs{}, fmt.Errorf("imagelayout: margins exceed canvas size")
	}

	units := strings.ToLower(strings.TrimSpace(settings.Units))
	if units == "" {
		units = "normalized"
	}
	if units != "normalized" && units != "px" {
		return Inputs{}, fmt.Errorf("imagelayout: unsupported position units %q", settings.Units)
	}

	var ratio *float64
	if settings.CropRatio != nil {
		if *settings.CropRatio <= 0 {
			return Inputs{}, fmt.Errorf("imagelayout: crop ratio must be > 0")
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
		return Inputs{}, fmt.Errorf("imagelayout: resulting canvas dimensions must be positive")
	}

	return Inputs{
		Mode:           mode,
		SourceW:        float64(meta.Width),
		SourceH:        float64(meta.Height),
		CanvasW:        canvasW,
		CanvasH:        canvasH,
		ContentW:       contentW,
		ContentH:       contentH,
		MarginTopPx:    mt,
		MarginRightPx:  mr,
		MarginBottomPx: mb,
		MarginLeftPx:   ml,
		CropRatio:      ratio,
		CropToFill:     settings.CropToFill,
		CropWidthPx:    cropWidth,
		CropHeightPx:   cropHeight,
		FitMode:        fitMode,
		FitWidthPx:     fitWidth,
		FitHeightPx:    fitHeight,
		UserScale:      scale,
		PositionX:      resolveAnchor(settings, units, settings.PositionX, true),
		PositionY:      resolveAnchor(settings, units, settings.PositionY, false),
		Units:          units,
		Focus:          settings.Focus,
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
func ComputeViewport(inp Inputs) (imagelayout.ViewportResult, *imagelayout.Trace) {
	trace := &imagelayout.Trace{
		Inputs: map[string]interface{}{
			"mode":         inp.Mode,
			"source_w":     inp.SourceW,
			"source_h":     inp.SourceH,
			"canvas_w":     inp.CanvasW,
			"canvas_h":     inp.CanvasH,
			"content_w":    inp.ContentW,
			"content_h":    inp.ContentH,
			"user_scale":   inp.UserScale,
			"units":        inp.Units,
			"crop_to_fill": inp.CropToFill,
		},
	}

	addStep := func(label string, data map[string]interface{}) {
		trace.Steps = append(trace.Steps, imagelayout.TraceStep{
			Label: label,
			Data:  data,
		})
	}

	canvasRect := imagelayout.Rect{
		X: inp.MarginLeftPx,
		Y: inp.MarginTopPx,
		W: inp.ContentW,
		H: inp.ContentH,
	}
	if inp.Mode != "page" {
		canvasRect = imagelayout.Rect{
			X: 0,
			Y: 0,
			W: inp.ContentW,
			H: inp.ContentH,
		}
	}

	targetW := canvasRect.W
	targetH := canvasRect.H
	targetRatio := safeDiv(targetW, targetH)

	sourceRatio := safeDiv(inp.SourceW, inp.SourceH)
	requestedRatio := sourceRatio
	if inp.CropRatio != nil {
		requestedRatio = *inp.CropRatio
	} else if inp.CropToFill && targetRatio > 0 {
		requestedRatio = targetRatio
	}

	var sx, sy, sw, sh float64 = 0, 0, inp.SourceW, inp.SourceH
	var rangeX, rangeY float64
	focusApplied := false
	if requestedRatio > 0 && inp.SourceW > 0 && inp.SourceH > 0 {
		switch {
		case sourceRatio > requestedRatio:
			// crop width
			sh = inp.SourceH
			sw = sh * requestedRatio
			rangeX = math.Max(0, inp.SourceW-sw)
			if inp.Focus == nil {
				sx = computeOffset(inp.Units, inp.PositionX, rangeX)
			}
		case sourceRatio < requestedRatio:
			sw = inp.SourceW
			sh = sw / requestedRatio
			rangeY = math.Max(0, inp.SourceH-sh)
			if inp.Focus == nil {
				sy = computeOffset(inp.Units, inp.PositionY, rangeY)
			}
		default:
			rangeX = math.Max(0, inp.SourceW-sw)
			rangeY = math.Max(0, inp.SourceH-sh)
			if inp.Focus == nil {
				sx = computeOffset(inp.Units, inp.PositionX, rangeX)
				sy = computeOffset(inp.Units, inp.PositionY, rangeY)
			}
		}
	} else {
		rangeX = math.Max(0, inp.SourceW-sw)
		rangeY = math.Max(0, inp.SourceH-sh)
	}

	focusInfo := map[string]interface{}{}
	if inp.Focus != nil {
		focusApplied = true
		fx := clampFloat(inp.Focus.SourceX, 0, inp.SourceW)
		fy := clampFloat(inp.Focus.SourceY, 0, inp.SourceH)
		targetNX := resolveFocusTarget(inp.Focus.TargetX, sw)
		targetNY := resolveFocusTarget(inp.Focus.TargetY, sh)
		sx = clampFloat(fx-targetNX*sw, 0, rangeX)
		sy = clampFloat(fy-targetNY*sh, 0, rangeY)
		focusInfo["focus_source_x"] = fx
		focusInfo["focus_source_y"] = fy
		focusInfo["focus_target_x"] = targetNX
		focusInfo["focus_target_y"] = targetNY
	}

	addStep("crop", map[string]interface{}{
		"sx": sx, "sy": sy, "sw": sw, "sh": sh,
		"source_ratio":    sourceRatio,
		"requested_ratio": requestedRatio,
		"range_x":         rangeX,
		"range_y":         rangeY,
		"focus_applied":   focusApplied,
	})
	if focusApplied {
		trace.Steps[len(trace.Steps)-1].Data["focus"] = focusInfo
	}

	scaleX := safeDiv(targetW, sw)
	scaleY := safeDiv(targetH, sh)
	mode := "contain"
	scale := math.Min(scaleX, scaleY)
	if inp.CropToFill {
		mode = "cover"
		scale = math.Max(scaleX, scaleY)
	}
	scale *= inp.UserScale

	dstW := sw * scale
	dstH := sh * scale
	tx, ty := positionOffsets(inp.Units, inp.PositionX, inp.PositionY, targetW, targetH, dstW, dstH)
	addStep("scale", map[string]interface{}{
		"scale_x": scaleX,
		"scale_y": scaleY,
		"final":   scale,
		"mode":    mode,
		"dst_w":   dstW,
		"dst_h":   dstH,
		"tx":      tx,
		"ty":      ty,
	})

	targetRect := imagelayout.Rect{
		X: canvasRect.X + tx,
		Y: canvasRect.Y + ty,
		W: dstW,
		H: dstH,
	}
	sourceRect := imagelayout.Rect{X: sx, Y: sy, W: sw, H: sh}

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
