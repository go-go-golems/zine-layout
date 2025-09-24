package engine

import (
    "fmt"
    "math"

    "github.com/go-go-golems/zine-layout/pkg/spread/sonnet/config"
)

type Rectangle struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type FloatSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type FloatPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Panel struct {
	Name        string     `json:"name"`
	PanelRect   Rectangle  `json:"panel_rect"`
	ImageOffset FloatPoint `json:"image_offset"`
}

type Result struct {
	SpreadName      string                  `json:"spread_name"`
	ImagePath       string                  `json:"image_path"`
	PaperSizePx     FloatSize               `json:"paper_size_px"`
	PaperSizeIn     FloatSize               `json:"paper_size_in"`
	MarginsPx       MarginsPx               `json:"margins_px"`
	GutterPx        int                     `json:"gutter_px"`
	EffectiveRect   Rectangle               `json:"effective_rect"`
	SourceCrop      Rectangle               `json:"source_crop"`
	ImageDisplay    FloatSize               `json:"image_display"`
	VirtualPosition FloatPoint              `json:"virtual_position"`
	Panels          []Panel                 `json:"panels"`
	Trace           []TraceEntry            `json:"trace"`
	Settings        config.ResolvedSettings `json:"-"`
}

type MarginsPx struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

type SourceMeta struct {
	Width  int
	Height int
}

func Compute(spec config.SpreadSpec, meta SourceMeta) (Result, error) {
	if meta.Width <= 0 || meta.Height <= 0 {
		return Result{}, fmt.Errorf("invalid source dimensions")
	}
	settings := spec.Settings
	dpi := settings.DPI
	trace := make([]TraceEntry, 0, 16)
	tracer := NewTracer(spec.Name, "compute", &trace)
	tracer.Log("source", "source image dimensions %dx%d px", meta.Width, meta.Height)

	paperWidthPx := inchesToPx(settings.PaperWidthIn, dpi)
	paperHeightPx := inchesToPx(settings.PaperHeightIn, dpi)
	if paperWidthPx <= 0 || paperHeightPx <= 0 {
		return Result{}, fmt.Errorf("paper dimensions resolve to zero")
	}
	tracer.Log("paper", "%.2fx%.2f in (%s) @ %.1f dpi → %dx%d px", settings.PaperWidthIn, settings.PaperHeightIn, settings.Orientation, dpi, paperWidthPx, paperHeightPx)

	marginTopPx := inchesToPx(settings.Margins.TopIn, dpi)
	marginRightPx := inchesToPx(settings.Margins.RightIn, dpi)
	marginBottomPx := inchesToPx(settings.Margins.BottomIn, dpi)
	marginLeftPx := inchesToPx(settings.Margins.LeftIn, dpi)
	gutterPx := inchesToPx(settings.GutterIn, dpi)
	tracer.Log("margins", "content margins (px) top=%d right=%d bottom=%d left=%d gutter=%d", marginTopPx, marginRightPx, marginBottomPx, marginLeftPx, gutterPx)

	effectiveWidth := paperWidthPx - marginLeftPx - marginRightPx
	if settings.IsSpread {
		effectiveWidth -= gutterPx
	}
	effectiveHeight := paperHeightPx - marginTopPx - marginBottomPx

	if effectiveWidth <= 0 || effectiveHeight <= 0 {
		return Result{}, fmt.Errorf("effective area is non-positive")
	}
	tracer.Log("effective-area", "effective area %dx%d px at (%d,%d)", effectiveWidth, effectiveHeight, marginLeftPx, marginTopPx)

	contentX0 := marginLeftPx
	contentY0 := marginTopPx

	var panels []Panel
	leftPanelWidth := effectiveWidth
	rightPanelWidth := 0
	if settings.IsSpread {
		leftPanelWidth = int(math.Floor(float64(effectiveWidth) / 2.0))
		rightPanelWidth = effectiveWidth - leftPanelWidth
	}

	leftPanelRect := Rectangle{X: contentX0, Y: contentY0, Width: leftPanelWidth, Height: effectiveHeight}
	panels = append(panels, Panel{Name: panelName(settings.IsSpread, true), PanelRect: leftPanelRect})

	var rightPanelRect Rectangle
	if settings.IsSpread {
		rightPanelRect = Rectangle{
			X:      contentX0 + leftPanelWidth + gutterPx,
			Y:      contentY0,
			Width:  rightPanelWidth,
			Height: effectiveHeight,
		}
		panels = append(panels, Panel{Name: "right", PanelRect: rightPanelRect})
		tracer.Log("spread", "spread panels: left %dx%d px, right %dx%d px (gutter %d px)", leftPanelRect.Width, leftPanelRect.Height, rightPanelRect.Width, rightPanelRect.Height, gutterPx)
	}
	if !settings.IsSpread {
		tracer.Log("spread", "single panel %dx%d px", leftPanelRect.Width, leftPanelRect.Height)
	}

	cropRect := centerCrop(meta.Width, meta.Height, settings.CropRatio)
	if settings.CropRatio != nil {
		tracer.Log("crop", "pre-crop ratio → %dx%d px (offset %d,%d)", cropRect.Width, cropRect.Height, cropRect.X, cropRect.Y)
	} else {
		tracer.Log("crop", "no pre-crop (using full image)")
	}

	sourceCrop := cropRect
	if settings.CropToFill {
		sourceCrop = cropToFill(sourceCrop, effectiveWidth, effectiveHeight, settings.Position)
		tracer.Log("crop", "crop-to-fill target %dx%d px with position (%.2f, %.2f %s)", sourceCrop.Width, sourceCrop.Height, settings.Position.X, settings.Position.Y, settings.Position.Units)
	}

	var imgDisplay FloatSize
	var virtualX, virtualY float64
	if settings.CropToFill {
		imgDisplay = FloatSize{Width: float64(effectiveWidth), Height: float64(effectiveHeight)}
		virtualX = float64(contentX0)
		virtualY = float64(contentY0)
		tracer.Log("scale", "fill mode → %dx%d px at origin (%d,%d)", effectiveWidth, effectiveHeight, contentX0, contentY0)
	} else {
		baseScale := math.Min(float64(effectiveWidth)/float64(sourceCrop.Width), float64(effectiveHeight)/float64(sourceCrop.Height))
		scale := baseScale * clampFloat(settings.UserScale, 0.1, 3.0)
		imgDisplay = FloatSize{Width: float64(sourceCrop.Width) * scale, Height: float64(sourceCrop.Height) * scale}

		cx := float64(contentX0) + (float64(effectiveWidth)-imgDisplay.Width)/2
		cy := float64(contentY0) + (float64(effectiveHeight)-imgDisplay.Height)/2

		panRangeX := math.Max(0, (imgDisplay.Width-float64(effectiveWidth))/2)
		panRangeY := math.Max(0, (imgDisplay.Height-float64(effectiveHeight))/2)
		px := positionValue(settings.Position.X, settings.Position.Units, panRangeX)
		py := positionValue(settings.Position.Y, settings.Position.Units, panRangeY)
		virtualX = cx - px
		virtualY = cy - py
		tracer.Log("scale", "fit mode: base scale %.4f × user %.4f = %.4f; display %.1fx%.1f px; pan offsets (%.1f, %.1f)", baseScale, settings.UserScale, scale, imgDisplay.Width, imgDisplay.Height, px, py)
	}

	panels[0].ImageOffset = FloatPoint{
		X: virtualX - float64(leftPanelRect.X),
		Y: virtualY - float64(leftPanelRect.Y),
	}
	if settings.IsSpread && len(panels) > 1 {
		panels[1].ImageOffset = FloatPoint{
			X: virtualX - float64(rightPanelRect.X),
			Y: virtualY - float64(rightPanelRect.Y),
		}
	}
	if !settings.IsSpread {
		panels[0].Name = "single"
	}
	for _, panel := range panels {
		tracer.Log("panels", "%s panel rect x=%d y=%d w=%d h=%d, image offset (%.1f, %.1f)", panel.Name, panel.PanelRect.X, panel.PanelRect.Y, panel.PanelRect.Width, panel.PanelRect.Height, panel.ImageOffset.X, panel.ImageOffset.Y)
	}

	result := Result{
		SpreadName:      spec.Name,
		ImagePath:       spec.ImagePath,
		PaperSizePx:     FloatSize{Width: float64(paperWidthPx), Height: float64(paperHeightPx)},
		PaperSizeIn:     FloatSize{Width: settings.PaperWidthIn, Height: settings.PaperHeightIn},
		MarginsPx:       MarginsPx{Top: marginTopPx, Right: marginRightPx, Bottom: marginBottomPx, Left: marginLeftPx},
		GutterPx:        gutterPx,
		EffectiveRect:   Rectangle{X: contentX0, Y: contentY0, Width: effectiveWidth, Height: effectiveHeight},
		SourceCrop:      sourceCrop,
		ImageDisplay:    imgDisplay,
		VirtualPosition: FloatPoint{X: virtualX, Y: virtualY},
		Panels:          panels,
		Trace:           trace,
		Settings:        settings,
	}
	return result, nil
}

func panelName(isSpread, left bool) string {
	if !isSpread {
		return "single"
	}
	if left {
		return "left"
	}
	return "right"
}

func inchesToPx(in float64, dpi float64) int {
	return int(math.Round(in * dpi))
}

func centerCrop(srcW, srcH int, ratio *float64) Rectangle {
	if ratio == nil {
		return Rectangle{X: 0, Y: 0, Width: srcW, Height: srcH}
	}
	target := *ratio
	srcAR := float64(srcW) / float64(srcH)
	if math.Abs(srcAR-target) < 1e-9 {
		return Rectangle{X: 0, Y: 0, Width: srcW, Height: srcH}
	}
	if srcAR > target {
		cropW := int(math.Round(float64(srcH) * target))
		if cropW < 1 {
			cropW = 1
		}
		cropX := (srcW - cropW) / 2
		return Rectangle{X: cropX, Y: 0, Width: cropW, Height: srcH}
	}
	cropH := int(math.Round(float64(srcW) / target))
	if cropH < 1 {
		cropH = 1
	}
	cropY := (srcH - cropH) / 2
	return Rectangle{X: 0, Y: cropY, Width: srcW, Height: cropH}
}

func cropToFill(rect Rectangle, effW, effH int, pos config.PositionValues) Rectangle {
	targetAR := float64(effW) / float64(effH)
	currentAR := float64(rect.Width) / float64(rect.Height)
	maxShiftX := 0
	maxShiftY := 0
	if currentAR > targetAR {
		targetW := int(math.Round(float64(rect.Height) * targetAR))
		if targetW < 1 {
			targetW = 1
		}
		maxShiftX = rect.Width - targetW
		offsetX := cropOffset(pos.X, pos.Units, maxShiftX)
		return Rectangle{X: rect.X + offsetX, Y: rect.Y, Width: targetW, Height: rect.Height}
	}
	if currentAR < targetAR {
		targetH := int(math.Round(float64(rect.Width) / targetAR))
		if targetH < 1 {
			targetH = 1
		}
		maxShiftY = rect.Height - targetH
		offsetY := cropOffset(pos.Y, pos.Units, maxShiftY)
		return Rectangle{X: rect.X, Y: rect.Y + offsetY, Width: rect.Width, Height: targetH}
	}
	return rect
}

func cropOffset(value float64, units string, maxShift int) int {
	if maxShift <= 0 {
		return 0
	}
	switch units {
	case "px":
		half := float64(maxShift) / 2
		v := clampFloat(value, -half, half)
		return int(math.Round(v + half))
	default: // normalized
		n := clampFloat(value, -1, 1)
		return int(math.Round(((n + 1) / 2) * float64(maxShift)))
	}
}

func positionValue(value float64, units string, rangeVal float64) float64 {
	if rangeVal <= 0 {
		return 0
	}
	switch units {
	case "px":
		return clampFloat(value, -rangeVal, rangeVal)
	default:
		normalized := clampFloat(value, -1, 1)
		return normalized * rangeVal
	}
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
