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
	settings := imagelayout.DefaultSettings()
	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}

	inputs, err := InputsFromSettings(settings, meta)
	if err != nil {
		t.Fatalf("InputsFromSettings: %v", err)
	}

	result, trace := ComputeViewport(inputs)
	if result.Mode != "contain" {
		t.Fatalf("expected mode contain, got %s", result.Mode)
	}

	if !almostEqual(result.CanvasRect.X, settings.MarginLeftIn*settings.DPI) {
		t.Fatalf("canvas rect x mismatch: got %f", result.CanvasRect.X)
	}
	if !almostEqual(result.TargetRect.W, 2250) {
		t.Fatalf("target width mismatch: got %f", result.TargetRect.W)
	}
	if !almostEqual(result.TargetRect.H, 1687.5) {
		t.Fatalf("target height mismatch: got %f", result.TargetRect.H)
	}

	if result.Scale <= 0 {
		t.Fatalf("scale should be positive, got %f", result.Scale)
	}

	if trace == nil || len(trace.Steps) == 0 {
		t.Fatalf("expected trace steps for debugging data")
	}
}

func TestComputeViewportWithCropRatio(t *testing.T) {
	settings := imagelayout.DefaultSettings()
	settings.CropRatio = floatPtr(1.0)

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromSettings(settings, meta)
	if err != nil {
		t.Fatalf("InputsFromSettings: %v", err)
	}

	result, _ := ComputeViewport(inputs)

	if !almostEqual(result.SourceRect.W, 3000) {
		t.Fatalf("expected cropped width 3000, got %f", result.SourceRect.W)
	}
	if !almostEqual(result.SourceRect.X, 500) {
		t.Fatalf("expected centered crop offset 500, got %f", result.SourceRect.X)
	}
	if !almostEqual(result.Scale, 0.75) {
		t.Fatalf("unexpected scale: %f", result.Scale)
	}
}

func TestComputeViewportCoverMode(t *testing.T) {
	settings := imagelayout.DefaultSettings()
	settings.CropToFill = true

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromSettings(settings, meta)
	if err != nil {
		t.Fatalf("InputsFromSettings: %v", err)
	}

	result, _ := ComputeViewport(inputs)

	if result.Mode != "cover" {
		t.Fatalf("expected cover mode, got %s", result.Mode)
	}
	if result.Scale <= 0 {
		t.Fatalf("expected positive scale, got %f", result.Scale)
	}
	if result.TargetRect.W < result.CanvasRect.W && result.TargetRect.H < result.CanvasRect.H {
		t.Fatalf("cover mode should fill one dimension fully")
	}
}

func TestInputsValidation(t *testing.T) {
	settings := imagelayout.DefaultSettings()
	settings.DPI = 0
	_, err := InputsFromSettings(settings, imagelayout.ImageMeta{Width: 100, Height: 100})
	if err == nil {
		t.Fatalf("expected error for invalid DPI")
	}
}

func TestComputeViewportCropMode(t *testing.T) {
	settings := imagelayout.DefaultSettings()
	settings.Mode = "crop"
	cw := 1200.0
	ch := 800.0
	settings.CropWidthPx = &cw
	settings.CropHeightPx = &ch
	settings.MarginTopIn = 0
	settings.MarginBottomIn = 0
	settings.MarginLeftIn = 0
	settings.MarginRightIn = 0

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromSettings(settings, meta)
	if err != nil {
		t.Fatalf("InputsFromSettings: %v", err)
	}

	result, _ := ComputeViewport(inputs)
	if !almostEqual(result.CanvasRect.W, cw) || !almostEqual(result.CanvasRect.H, ch) {
		t.Fatalf("expected canvas %fx%f got %fx%f", cw, ch, result.CanvasRect.W, result.CanvasRect.H)
	}
	if result.TargetRect.X != 0 || result.TargetRect.Y != 0 {
		t.Fatalf("expected target origin at 0, got (%f,%f)", result.TargetRect.X, result.TargetRect.Y)
	}
}

func TestComputeViewportFitWidth(t *testing.T) {
	settings := imagelayout.DefaultSettings()
	settings.Mode = "fit"
	settings.FitMode = "width"
	width := 1600.0
	settings.FitWidthPx = &width
	settings.MarginTopIn = 0
	settings.MarginBottomIn = 0
	settings.MarginLeftIn = 0
	settings.MarginRightIn = 0

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromSettings(settings, meta)
	if err != nil {
		t.Fatalf("InputsFromSettings: %v", err)
	}

	result, _ := ComputeViewport(inputs)
	if !almostEqual(result.CanvasRect.W, width) {
		t.Fatalf("expected canvas width %f got %f", width, result.CanvasRect.W)
	}
	if result.CanvasRect.H <= 0 {
		t.Fatalf("expected positive canvas height")
	}
}

func TestAnchorPresetBottomRight(t *testing.T) {
	settings := imagelayout.DefaultSettings()
	settings.AnchorPreset = "bottom-right"

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromSettings(settings, meta)
	if err != nil {
		t.Fatalf("InputsFromSettings: %v", err)
	}
	if !almostEqual(inputs.PositionX, 1) || !almostEqual(inputs.PositionY, 1) {
		t.Fatalf("anchor preset not applied correctly: %f %f", inputs.PositionX, inputs.PositionY)
	}

	result, _ := ComputeViewport(inputs)
	// With the preset applied, the target rectangle should not extend beyond canvas bounds.
	if result.TargetRect.X < result.CanvasRect.X || result.TargetRect.Y < result.CanvasRect.Y {
		t.Fatalf("target rect unexpectedly outside canvas")
	}
}

func TestFocusPointCentersSource(t *testing.T) {
	settings := imagelayout.DefaultSettings()
	settings.CropRatio = floatPtr(1.0)
	settings.Focus = &imagelayout.FocusPoint{
		SourceX: 3500,
		SourceY: 500,
		TargetX: 0.2,
		TargetY: 0.3,
	}

	meta := imagelayout.ImageMeta{Width: 4000, Height: 3000}
	inputs, err := InputsFromSettings(settings, meta)
	if err != nil {
		t.Fatalf("InputsFromSettings: %v", err)
	}

	result, trace := ComputeViewport(inputs)
	if !almostEqual(result.SourceRect.X, 1000) {
		t.Fatalf("expected focus to shift crop to 1000, got %f", result.SourceRect.X)
	}
	if trace == nil || len(trace.Steps) == 0 {
		t.Fatalf("expected trace data")
	}
	last := trace.Steps[0]
	if focus, ok := last.Data["focus"]; !ok || focus == nil {
		t.Fatalf("expected focus info in trace")
	}
}
