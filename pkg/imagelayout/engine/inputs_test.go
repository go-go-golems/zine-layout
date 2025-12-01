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

	if inputs.Mode != "ratio" {
		t.Fatalf("expected ratio mode, got %s", inputs.Mode)
	}
	expectedCanvasW := ratio * float64(meta.Height)
	if !almostEqual(inputs.CanvasW, expectedCanvasW) {
		t.Fatalf("unexpected canvas width: %f", inputs.CanvasW)
	}
	if inputs.CropRatio == nil || !almostEqual(*inputs.CropRatio, ratio) {
		t.Fatalf("expected crop ratio %.2f", ratio)
	}
	if inputs.MarginTopPx != 0 || inputs.MarginLeftPx != 0 {
		t.Fatalf("ratio mode should not have margins")
	}
	if !almostEqual(inputs.CropZoom, 1.0) || !almostEqual(inputs.CropExtent, 1.0) {
		t.Fatalf("expected default crop zoom/extent of 1")
	}
	if inputs.PresentationUnits != "normalized" {
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

	if inputs.Mode != "page" {
		t.Fatalf("expected page mode, got %s", inputs.Mode)
	}

	expectedCanvasW := float64(10 * 300)
	expectedCanvasH := float64(8 * 300)
	if !almostEqual(inputs.CanvasW, expectedCanvasW) || !almostEqual(inputs.CanvasH, expectedCanvasH) {
		t.Fatalf("unexpected canvas size %fx%f", inputs.CanvasW, inputs.CanvasH)
	}
	if inputs.MarginTopPx <= 0 || inputs.MarginLeftPx <= 0 {
		t.Fatalf("expected positive margins for page mode")
	}
	if inputs.ContentW >= inputs.CanvasW || inputs.ContentH >= inputs.CanvasH {
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

	if inputs.Mode != "viewport" {
		t.Fatalf("expected viewport mode, got %s", inputs.Mode)
	}
	expectedCanvasW := 900 * ratio
	if !almostEqual(inputs.CanvasW, expectedCanvasW) {
		t.Fatalf("expected derived width %.2f got %.2f", expectedCanvasW, inputs.CanvasW)
	}
	if !almostEqual(inputs.CanvasH, 900) {
		t.Fatalf("expected canvas height 900 got %.2f", inputs.CanvasH)
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

	if !almostEqual(inputs.CropZoom, 2.0) {
		t.Fatalf("expected crop zoom 2 got %.2f", inputs.CropZoom)
	}
	if !almostEqual(inputs.CropExtent, 0.8) {
		t.Fatalf("expected crop extent 0.8 got %.2f", inputs.CropExtent)
	}
	if inputs.Units != "normalized" {
		t.Fatalf("expected crop units normalized")
	}
	if inputs.PresentationUnits != "px" {
		t.Fatalf("expected presentation units px")
	}
	if !almostEqual(inputs.PresentationOffsetX, 24) || !almostEqual(inputs.PresentationOffsetY, -18) {
		t.Fatalf("unexpected presentation offsets %f %f", inputs.PresentationOffsetX, inputs.PresentationOffsetY)
	}
	if inputs.UserScale != 1.25 {
		t.Fatalf("user scale mismatch")
	}
	if inputs.ClampToCanvas {
		t.Fatalf("expected clamp to canvas disabled")
	}
}
