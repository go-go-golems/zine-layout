package engine

import (
	"math"
	"testing"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
)

func floatPtr(v float64) *float64 { return &v }

func almostEqual(a, b float64) bool {
	const eps = 1e-6
	return math.Abs(a-b) <= eps
}

func TestComputeViewportContain(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	r := 4.0 / 3.0
	req.Frame.Mode = "ratio"
	req.Frame.Ratio = &r
	req.Frame.Fill = "contain"

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}

	result, trace := ComputeViewport(inputs)
	if result.Mode != "contain" {
		t.Fatalf("expected mode contain, got %s", result.Mode)
	}
	if !almostEqual(result.CanvasRect.W, 4000) || !almostEqual(result.CanvasRect.H, 3000) {
		t.Fatalf("unexpected canvas size %fx%f", result.CanvasRect.W, result.CanvasRect.H)
	}
	if result.Scale <= 0 {
		t.Fatalf("scale should be positive, got %f", result.Scale)
	}
	if trace == nil || len(trace.Steps) == 0 {
		t.Fatalf("expected trace steps for debugging data")
	}
}

func TestComputeViewportWithCropRatio(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	r := 4.0 / 3.0
	req.Frame.Mode = "ratio"
	req.Frame.Ratio = &r
	req.Crop.Ratio = floatPtr(1.0)

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}

	result, _ := ComputeViewport(inputs)

	if !almostEqual(result.SourceRect.W, 3000) {
		t.Fatalf("expected cropped width 3000, got %f", result.SourceRect.W)
	}
	if !almostEqual(result.SourceRect.X, 500) {
		t.Fatalf("expected centered crop offset 500, got %f", result.SourceRect.X)
	}
}

func TestComputeViewportCoverMode(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	req.Frame.Mode = "ratio"
	req.Frame.Fill = "cover"
	val := 16.0 / 9.0
	req.Frame.Ratio = &val

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}

	result, _ := ComputeViewport(inputs)
	if result.Mode != "cover" {
		t.Fatalf("expected cover mode, got %s", result.Mode)
	}
	if result.Scale <= 0 {
		t.Fatalf("expected positive scale, got %f", result.Scale)
	}
}

func TestAnchorPresetBottomRight(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	req.Crop.Strategy = "anchor"
	req.Crop.Anchor = "bottom-right"

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}
	if !almostEqual(inputs.Crop.PanX, 1) || !almostEqual(inputs.Crop.PanY, 1) {
		t.Fatalf("anchor preset not applied correctly: %f %f", inputs.Crop.PanX, inputs.Crop.PanY)
	}

	result, _ := ComputeViewport(inputs)
	if result.TargetRect.X < result.CanvasRect.X || result.TargetRect.Y < result.CanvasRect.Y {
		t.Fatalf("target rect unexpectedly outside canvas")
	}
}

func TestComputeViewportWithCropZoom(t *testing.T) {
	req := imagelayout.DefaultLayoutRequest()
	val := 1.0
	req.Frame.Mode = "ratio"
	req.Frame.Ratio = &val
	req.Crop.Strategy = "manual"
	req.Crop.Zoom = 2.0
	req.Crop.Pan = imagelayout.Vec2{X: 0, Y: 0}

	meta := imagelayout.ImageMeta{Width: 4000, Height: 4000}
	inputs, err := InputsFromRequest(req, meta)
	if err != nil {
		t.Fatalf("InputsFromRequest: %v", err)
	}

	result, _ := ComputeViewport(inputs)
	if !almostEqual(result.SourceRect.W, 2000) {
		t.Fatalf("expected zoomed width 2000, got %f", result.SourceRect.W)
	}
	if !almostEqual(result.SourceRect.H, 2000) {
		t.Fatalf("expected zoomed height 2000, got %f", result.SourceRect.H)
	}
}
