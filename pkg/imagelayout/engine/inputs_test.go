package engine

import (
	"testing"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
)

func TestInputsFromRequestRatioFrame(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	ratio := 3.0 / 2.0
	req.Frame.Mode = "ratio"
	req.Frame.Ratio = &ratio
	req.Frame.Page = nil

	meta := imagelayout.ImageMeta{Width: 4000, Height: 2000}

	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}

	if inputs.Frame.Mode != "ratio" {
		t.Fatalf("expected ratio mode, got %s", inputs.Frame.Mode)
	}
	expectedCanvasW := ratio * float64(meta.Height)
	if !almostEqual(inputs.Frame.CanvasRect.W, expectedCanvasW) {
		t.Fatalf("unexpected canvas width: %f", inputs.Frame.CanvasRect.W)
	}
	if inputs.Crop.Ratio == nil || !almostEqual(*inputs.Crop.Ratio, ratio) {
		t.Fatalf("expected crop ratio %.2f", ratio)
	}
	if inputs.Frame.Margins.Top != 0 || inputs.Frame.Margins.Left != 0 {
		t.Fatalf("ratio mode should not have margins")
	}
	if !almostEqual(inputs.Crop.Zoom, 1.0) || !almostEqual(inputs.Crop.Extent, 1.0) {
		t.Fatalf("expected default crop zoom/extent of 1")
	}
	if inputs.Presentation.OffsetUnits != "normalized" {
		t.Fatalf("expected normalized presentation units")
	}
}

func TestInputsFromRequestPageFrame(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	req.Frame.Mode = "page"
	req.Frame.Page = &imagelayout.PageFrame{
		WidthIn:  10,
		HeightIn: 8,
		DPI:      300,
		MarginsIn: imagelayout.BoxSpacing{
			Top:    0.5,
			Right:  0.25,
			Bottom: 0.5,
			Left:   0.25,
		},
	}

	meta := imagelayout.ImageMeta{Width: 3000, Height: 2000}

	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}

	if inputs.Frame.Mode != "page" {
		t.Fatalf("expected page mode, got %s", inputs.Frame.Mode)
	}

	expectedCanvasW := float64(10 * 300)
	expectedCanvasH := float64(8 * 300)
	canvasW := inputs.Frame.CanvasRect.W + inputs.Frame.Margins.Left + inputs.Frame.Margins.Right
	canvasH := inputs.Frame.CanvasRect.H + inputs.Frame.Margins.Top + inputs.Frame.Margins.Bottom
	if !almostEqual(canvasW, expectedCanvasW) || !almostEqual(canvasH, expectedCanvasH) {
		t.Fatalf("unexpected canvas size %fx%f", canvasW, canvasH)
	}
	if inputs.Frame.Margins.Top <= 0 || inputs.Frame.Margins.Left <= 0 {
		t.Fatalf("expected positive margins for page mode")
	}
	if inputs.Frame.CanvasRect.W >= canvasW || inputs.Frame.CanvasRect.H >= canvasH {
		t.Fatalf("content should be smaller than canvas once margins applied")
	}
}

func TestInputsFromRequestViewportFrame(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	req.Frame.Mode = "viewport"
	req.Frame.Viewport = &imagelayout.ViewportFrame{
		Width:  0,
		Height: 900,
	}
	ratio := 16.0 / 9.0
	req.Frame.Ratio = &ratio

	meta := imagelayout.ImageMeta{Width: 4000, Height: 2500}

	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}

	if inputs.Frame.Mode != "viewport" {
		t.Fatalf("expected viewport mode, got %s", inputs.Frame.Mode)
	}
	expectedCanvasW := 900 * ratio
	if !almostEqual(inputs.Frame.CanvasRect.W, expectedCanvasW) {
		t.Fatalf("expected derived width %.2f got %.2f", expectedCanvasW, inputs.Frame.CanvasRect.W)
	}
	if !almostEqual(inputs.Frame.CanvasRect.H, 900) {
		t.Fatalf("expected canvas height 900 got %.2f", inputs.Frame.CanvasRect.H)
	}
}

func TestInputsFromRequestCropZoomAndPresentation(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	req.Frame.Mode = "ratio"
	val := 1.0
	req.Frame.Ratio = &val
	req.Crop.Strategy = "manual"
	req.Crop.Pan = imagelayout.Vec2{X: 0.25, Y: -0.25}
	req.Crop.Zoom = 2.0
	req.Crop.Extent = 0.8
	req.Crop.Units = "normalized"
	req.Presentation.UserScale = 1.25
	req.Presentation.OffsetPx = imagelayout.Vec2Px{X: 24, Y: -18}
	req.Presentation.ClampToCanvas = false

	meta := imagelayout.ImageMeta{Width: 5000, Height: 5000}

	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}

	if !almostEqual(inputs.Crop.Zoom, 2.0) {
		t.Fatalf("expected crop zoom 2 got %.2f", inputs.Crop.Zoom)
	}
	if !almostEqual(inputs.Crop.Extent, 0.8) {
		t.Fatalf("expected crop extent 0.8 got %.2f", inputs.Crop.Extent)
	}
	if inputs.Crop.Units != "normalized" {
		t.Fatalf("expected crop units normalized")
	}
	if inputs.Presentation.OffsetUnits != "px" {
		t.Fatalf("expected presentation units px")
	}
	if !almostEqual(inputs.Presentation.OffsetX, 24) || !almostEqual(inputs.Presentation.OffsetY, -18) {
		t.Fatalf("unexpected presentation offsets %f %f", inputs.Presentation.OffsetX, inputs.Presentation.OffsetY)
	}
	if inputs.Presentation.UserScale != 1.25 {
		t.Fatalf("user scale mismatch")
	}
	if inputs.Presentation.ClampToCanvas {
		t.Fatalf("expected clamp to canvas disabled")
	}
}
