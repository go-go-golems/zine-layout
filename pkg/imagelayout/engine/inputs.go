package engine

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
)

var anchorPresets = map[string][2]float64{
	"top-left":      {-1, -1},
	"top":           {0, -1},
	"top-right":     {1, -1},
	"left":          {-1, 0},
	"center":        {0, 0},
	"middle-center": {0, 0},
	"right":         {1, 0},
	"bottom-left":   {-1, 1},
	"bottom":        {0, 1},
	"bottom-right":  {1, 1},
}

// InputsFromRequest converts the refactored LayoutRequest payload into engine inputs.
func InputsFromRequest(req imagelayout.LayoutRequest, meta imagelayout.ImageMeta) (NormalizedInputs, error) {
	if meta.Width <= 0 || meta.Height <= 0 {
		return NormalizedInputs{}, fmt.Errorf("imagelayout: invalid source dimensions %dx%d", meta.Width, meta.Height)
	}

	normalized := normalizeLayoutRequest(req)

	rawFrameMode := strings.ToLower(strings.TrimSpace(normalized.Frame.Mode))
	if rawFrameMode == "" {
		rawFrameMode = string(FrameModeRatio)
	}
	frameMode := FrameMode(rawFrameMode)

	var (
		canvasW  float64
		canvasH  float64
		contentW float64
		contentH float64
		mt       float64
		mr       float64
		mb       float64
		ml       float64
		contentX float64
		contentY float64
	)

	switch frameMode {
	case FrameModePage:
		page := normalized.Frame.Page
		if page == nil {
			return NormalizedInputs{}, fmt.Errorf("imagelayout: page mode requires page dimensions")
		}
		dpi := page.DPI
		if dpi <= 0 {
			return NormalizedInputs{}, fmt.Errorf("imagelayout: page dpi must be positive")
		}
		widthIn := page.WidthIn
		heightIn := page.HeightIn
		if strings.ToLower(page.Orientation) == "landscape" {
			widthIn, heightIn = page.HeightIn, page.WidthIn
		}
		if widthIn <= 0 || heightIn <= 0 {
			return NormalizedInputs{}, fmt.Errorf("imagelayout: page dimensions must be positive")
		}
		canvasW = widthIn * dpi
		canvasH = heightIn * dpi
		mt = page.MarginsIn.Top * dpi
		mr = page.MarginsIn.Right * dpi
		mb = page.MarginsIn.Bottom * dpi
		ml = page.MarginsIn.Left * dpi
		contentW = canvasW - (ml + mr)
		contentH = canvasH - (mt + mb)
		if contentW <= 0 || contentH <= 0 {
			return NormalizedInputs{}, fmt.Errorf("imagelayout: margins exceed canvas size")
		}
	case FrameModeViewport:
		vp := normalized.Frame.Viewport
		if vp == nil {
			return NormalizedInputs{}, fmt.Errorf("imagelayout: viewport mode requires viewport dimensions")
		}
		dims, err := resolveViewportDims(vp, normalized.Frame.Ratio, meta)
		if err != nil {
			return NormalizedInputs{}, err
		}
		canvasW = dims[0]
		canvasH = dims[1]
		contentW = canvasW
		contentH = canvasH
	default: // ratio (fallback)
		frameMode = FrameModeRatio
		ratio := normalized.Frame.Ratio
		var ratioValue float64
		if ratio != nil && *ratio > 0 {
			ratioValue = *ratio
		} else {
			ratioValue = safeDiv(float64(meta.Width), float64(meta.Height))
		}
		canvasH = float64(meta.Height)
		if canvasH <= 0 {
			canvasH = 1
		}
		canvasW = ratioValue * canvasH
		if canvasW <= 0 {
			canvasW = float64(meta.Width)
		}
		contentW = canvasW
		contentH = canvasH
	}

	cropUnits := strings.ToLower(strings.TrimSpace(normalized.Crop.Units))
	if cropUnits == "" {
		cropUnits = "normalized"
	}
	if cropUnits != "normalized" && cropUnits != "px" {
		return NormalizedInputs{}, fmt.Errorf("imagelayout: unsupported crop units %q", normalized.Crop.Units)
	}

	cropRatio := normalized.Crop.Ratio
	if cropRatio == nil {
		value := safeDiv(canvasW, canvasH)
		cropRatio = &value
	}

	cropToFill := strings.ToLower(normalized.Frame.Fill) == "cover"

	posX, posY := resolveCropPosition(normalized.Crop)

	cropZoom := normalized.Crop.Zoom
	if cropZoom <= 0 {
		cropZoom = 1.0
	}
	cropExtent := normalized.Crop.Extent
	if cropExtent <= 0 || cropExtent > 1 {
		cropExtent = 1.0
	}

	canvasRect := imagelayout.Rect{
		X: 0,
		Y: 0,
		W: canvasW,
		H: canvasH,
	}

	if frameMode != "page" {
		contentX, contentY = 0, 0
		mt, mr, mb, ml = 0, 0, 0, 0
	} else {
		contentX = ml
		contentY = mt
	}
	contentRect := imagelayout.Rect{
		X: contentX,
		Y: contentY,
		W: contentW,
		H: contentH,
	}

	margins := MarginPixels{
		Top:    mt,
		Right:  mr,
		Bottom: mb,
		Left:   ml,
	}
	if frameMode != "page" {
		margins = MarginPixels{}
	}

	inputs := NormalizedInputs{
		Source: SourceMeta{
			Width:  float64(meta.Width),
			Height: float64(meta.Height),
		},
		Frame: FrameInputs{
			Mode:        frameMode,
			CanvasRect:  canvasRect,
			ContentRect: contentRect,
			Margins:     margins,
		},
		Crop: CropInputs{
			Ratio:      cropRatio,
			CropToFill: cropToFill,
			Zoom:       cropZoom,
			Extent:     cropExtent,
			Units:      cropUnits,
			PanX:       posX,
			PanY:       posY,
			Focus:      normalized.Crop.Focus,
		},
	}

	return inputs, nil
}

func resolveViewportDims(vp *imagelayout.ViewportFrame, ratio *float64, meta imagelayout.ImageMeta) ([2]float64, error) {
	width := vp.Width
	height := vp.Height
	ratioValue := pointerValue(ratio)
	sourceRatio := safeDiv(float64(meta.Width), float64(meta.Height))

	if ratioValue <= 0 {
		ratioValue = sourceRatio
	}

	switch {
	case width > 0 && height > 0:
		// already defined
	case width > 0 && height <= 0:
		height = width / ratioValue
	case height > 0 && width <= 0:
		width = height * ratioValue
	default:
		if ratioValue <= 0 {
			return [2]float64{}, fmt.Errorf("imagelayout: viewport requires at least one dimension or ratio")
		}
		height = float64(meta.Height)
		if height <= 0 {
			height = 1
		}
		width = ratioValue * height
	}

	if width <= 0 || height <= 0 {
		return [2]float64{}, fmt.Errorf("imagelayout: viewport dimensions must be positive")
	}
	return [2]float64{width, height}, nil
}

func normalizeLayoutRequest(req imagelayout.LayoutRequest) imagelayout.LayoutRequest {
	defaults := imagelayout.DefaultLayoutRequest()

	out := defaults

	// Frame overrides
	if req.Frame.Mode != "" {
		out.Frame.Mode = req.Frame.Mode
	}
	if req.Frame.Ratio != nil {
		out.Frame.Ratio = req.Frame.Ratio
	}
	if req.Frame.Fill != "" {
		out.Frame.Fill = req.Frame.Fill
	}
	if req.Frame.Page != nil {
		out.Frame.Page = clonePageFrame(*req.Frame.Page)
	}
	if req.Frame.Viewport != nil {
		vp := *req.Frame.Viewport
		out.Frame.Viewport = &vp
	}
	if req.Frame.FitAxis != "" {
		out.Frame.FitAxis = req.Frame.FitAxis
	}

	// Crop overrides
	if req.Crop.Strategy != "" {
		out.Crop.Strategy = req.Crop.Strategy
	}
	if req.Crop.Ratio != nil {
		out.Crop.Ratio = req.Crop.Ratio
	}
	if req.Crop.Zoom != 0 {
		out.Crop.Zoom = req.Crop.Zoom
	}
	if req.Crop.Extent != 0 {
		out.Crop.Extent = req.Crop.Extent
	}
	if req.Crop.Anchor != "" {
		out.Crop.Anchor = req.Crop.Anchor
	}
	if req.Crop.Pan != (imagelayout.Vec2{}) {
		out.Crop.Pan = req.Crop.Pan
	}
	if req.Crop.Focus != nil {
		out.Crop.Focus = req.Crop.Focus
	}
	if req.Crop.Units != "" {
		out.Crop.Units = req.Crop.Units
	}

	return out
}

func clonePageFrame(in imagelayout.PageFrame) *imagelayout.PageFrame {
	page := in
	page.MarginsIn = imagelayout.BoxSpacing{
		Top:    in.MarginsIn.Top,
		Right:  in.MarginsIn.Right,
		Bottom: in.MarginsIn.Bottom,
		Left:   in.MarginsIn.Left,
	}
	return &page
}

func resolveCropPosition(spec imagelayout.CropSpec) (float64, float64) {
	strategy := strings.ToLower(strings.TrimSpace(spec.Strategy))
	switch strategy {
	case "anchor":
		key := strings.ToLower(strings.TrimSpace(spec.Anchor))
		if vec, ok := anchorPresets[key]; ok {
			return vec[0], vec[1]
		}
	case "manual":
		return spec.Pan.X, spec.Pan.Y
	}
	return 0, 0
}
