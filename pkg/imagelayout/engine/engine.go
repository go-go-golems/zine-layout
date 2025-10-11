package engine

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
)

// Inputs collects normalized values derived from settings + image metadata.
type Inputs struct {
	SourceW float64
	SourceH float64

	CanvasW float64
	CanvasH float64

	ContentW float64
	ContentH float64

	EffectiveW float64

	MarginTopPx    float64
	MarginRightPx  float64
	MarginBottomPx float64
	MarginLeftPx   float64

	CropRatio  *float64
	CropToFill bool

	UserScale float64
	PositionX float64
	PositionY float64
	Units     string
}

// InputsFromSettings converts persisted settings and image metadata into algorithm inputs.
func InputsFromSettings(settings imagelayout.ViewportSettings, meta imagelayout.ImageMeta) (Inputs, error) {
	if meta.Width <= 0 || meta.Height <= 0 {
		return Inputs{}, fmt.Errorf("imagelayout: invalid source dimensions %dx%d", meta.Width, meta.Height)
	}
	if settings.DPI <= 0 {
		return Inputs{}, fmt.Errorf("imagelayout: dpi must be positive")
	}
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

	return Inputs{
		SourceW:        float64(meta.Width),
		SourceH:        float64(meta.Height),
		CanvasW:        canvasW,
		CanvasH:        canvasH,
		ContentW:       contentW,
		ContentH:       contentH,
		EffectiveW:     contentW,
		MarginTopPx:    mt,
		MarginRightPx:  mr,
		MarginBottomPx: mb,
		MarginLeftPx:   ml,
		CropRatio:      ratio,
		CropToFill:     settings.CropToFill,
		UserScale:      scale,
		PositionX:      settings.PositionX,
		PositionY:      settings.PositionY,
		Units:          units,
	}, nil
}

// ComputeViewport calculates viewport placement and trace data.
func ComputeViewport(inp Inputs) (imagelayout.ViewportResult, *imagelayout.Trace) {
	trace := &imagelayout.Trace{
		Inputs: map[string]interface{}{
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

	targetW := inp.ContentW
	targetH := inp.ContentH
	targetRatio := safeDiv(targetW, targetH)

	sourceRatio := safeDiv(inp.SourceW, inp.SourceH)
	requestedRatio := sourceRatio
	if inp.CropRatio != nil {
		requestedRatio = *inp.CropRatio
	} else if inp.CropToFill && targetRatio > 0 {
		requestedRatio = targetRatio
	}

	var sx, sy, sw, sh float64 = 0, 0, inp.SourceW, inp.SourceH
	if requestedRatio > 0 && inp.SourceW > 0 && inp.SourceH > 0 {
		switch {
		case sourceRatio > requestedRatio:
			// crop width
			sh = inp.SourceH
			sw = sh * requestedRatio
			offset := computeOffset(inp.Units, inp.PositionX, inp.SourceW-sw)
			sx = offset
		case sourceRatio < requestedRatio:
			sw = inp.SourceW
			sh = sw / requestedRatio
			offset := computeOffset(inp.Units, inp.PositionY, inp.SourceH-sh)
			sy = offset
		}
	}
	addStep("crop", map[string]interface{}{
		"sx": sx, "sy": sy, "sw": sw, "sh": sh,
		"source_ratio":    sourceRatio,
		"requested_ratio": requestedRatio,
	})

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

	canvasRect := imagelayout.Rect{
		X: inp.MarginLeftPx,
		Y: inp.MarginTopPx,
		W: inp.ContentW,
		H: inp.ContentH,
	}
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
		clamped := math.Max(-1, math.Min(1, value))
		return (clamped + 1) * 0.5 * rangePx
	}
	half := rangePx / 2
	offset := value
	if offset < -half {
		offset = -half
	}
	if offset > half {
		offset = half
	}
	return offset + half
}

func positionOffsets(units string, px, py float64, targetW, targetH, dstW, dstH float64) (float64, float64) {
	if units == "px" {
		return px, py
	}
	freeX := targetW - dstW
	freeY := targetH - dstH
	tx := (math.Max(-1, math.Min(1, px)) * freeX) / 2
	ty := (math.Max(-1, math.Min(1, py)) * freeY) / 2
	return tx, ty
}
