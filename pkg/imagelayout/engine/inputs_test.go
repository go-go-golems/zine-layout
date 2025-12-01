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

	if inputs.Frame.Mode != FrameModeRatio {
		t.Fatalf("expected ratio mode, got %s", inputs.Frame.Mode)
	}
	expectedCanvasW := ratio * float64(meta.Height)
	if !almostEqual(inputs.Frame.CanvasRect.W, expectedCanvasW) {
		t.Fatalf("unexpected canvas width: %f", inputs.Frame.CanvasRect.W)
	}
	if !almostEqual(inputs.Frame.ContentRect.W, expectedCanvasW) || !almostEqual(inputs.Frame.ContentRect.H, inputs.Frame.CanvasRect.H) {
		t.Fatalf("content rect should match canvas in ratio mode")
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

	if inputs.Frame.Mode != FrameModePage {
		t.Fatalf("expected page mode, got %s", inputs.Frame.Mode)
	}

	expectedCanvasW := float64(10 * 300)
	expectedCanvasH := float64(8 * 300)
	canvasW := inputs.Frame.CanvasRect.W
	canvasH := inputs.Frame.CanvasRect.H
	if !almostEqual(canvasW, expectedCanvasW) || !almostEqual(canvasH, expectedCanvasH) {
		t.Fatalf("unexpected canvas size %fx%f", canvasW, canvasH)
	}
	if !almostEqual(inputs.Frame.ContentRect.W, expectedCanvasW-inputs.Frame.Margins.Left-inputs.Frame.Margins.Right) {
		t.Fatalf("unexpected content width")
	}
	if !almostEqual(inputs.Frame.ContentRect.H, expectedCanvasH-inputs.Frame.Margins.Top-inputs.Frame.Margins.Bottom) {
		t.Fatalf("unexpected content height")
	}
	if inputs.Frame.Margins.Top <= 0 || inputs.Frame.Margins.Left <= 0 {
		t.Fatalf("expected positive margins for page mode")
	}
	if inputs.Frame.ContentRect.W >= canvasW || inputs.Frame.ContentRect.H >= canvasH {
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

	if inputs.Frame.Mode != FrameModeViewport {
		t.Fatalf("expected viewport mode, got %s", inputs.Frame.Mode)
	}
	expectedCanvasW := 900 * ratio
	if !almostEqual(inputs.Frame.CanvasRect.W, expectedCanvasW) {
		t.Fatalf("expected derived width %.2f got %.2f", expectedCanvasW, inputs.Frame.CanvasRect.W)
	}
	if !almostEqual(inputs.Frame.CanvasRect.H, 900) {
		t.Fatalf("expected canvas height 900 got %.2f", inputs.Frame.CanvasRect.H)
	}
	if !almostEqual(inputs.Frame.ContentRect.W, inputs.Frame.CanvasRect.W) || !almostEqual(inputs.Frame.ContentRect.H, inputs.Frame.CanvasRect.H) {
		t.Fatalf("content should match canvas in viewport mode")
	}
}
