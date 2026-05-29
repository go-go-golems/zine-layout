package renderer

import (
	"image"
	"image/color"
	"testing"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"github.com/go-go-golems/zine-layout/pkg/pagelayout"
)

// makeTestSplitImage creates a 100x100 image with left half red and right half blue.
func makeTestSplitImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if x < 50 {
				img.Set(x, y, red)
			} else {
				img.Set(x, y, blue)
			}
		}
	}
	return img
}

func TestRenderPage_UsesLayoutCrop(t *testing.T) {
	// Crop to the right half (blue)
	src := makeTestSplitImage()
	settings := pagelayout.PageLayoutSettings{
		PageWidthIn:  1,
		PageHeightIn: 1,
		DPI:          100,
		// No margins for simplicity
		PositioningMode: "fill",
	}
	lr := imagelayout.ViewportResult{
		SourceRect: imagelayout.Rect{X: 50, Y: 0, W: 50, H: 100},
	}
	ctx := RenderContext{Settings: settings, Source: src, LayoutResult: &lr}
	result, err := RenderPage(ctx)
	if err != nil {
		t.Fatalf("RenderPage failed: %v", err)
	}
	if result == nil || result.Full == nil {
		t.Fatalf("nil render result")
	}
	// Sample center pixel; should be blue (since cropped to right half only)
	center := result.Full.At(result.Full.Bounds().Dx()/2, result.Full.Bounds().Dy()/2)
	r, g, b, _ := center.RGBA()
	// Allow some tolerance due to scaling; but dominant channel should be blue
	if b <= r || b <= g {
		t.Fatalf("expected center to be predominantly blue, got R=%d G=%d B=%d", r, g, b)
	}
}

func TestRenderPage_SpreadSplitDimensions(t *testing.T) {
	// Single-color source; check left/right widths with gutter considered
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	green := color.RGBA{0, 255, 0, 255}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			src.Set(x, y, green)
		}
	}
	settings := pagelayout.PageLayoutSettings{
		PageWidthIn:     4,   // 400 px
		PageHeightIn:    1,   // 100 px
		DPI:             100, // px per inch
		IsSpread:        true,
		GutterWidthIn:   1, // 100 px gutter
		PositioningMode: "fill",
	}
	ctx := RenderContext{Settings: settings, Source: src}
	result, err := RenderPage(ctx)
	if err != nil {
		t.Fatalf("RenderPage failed: %v", err)
	}
	left, okL := result.Variants["left"]
	right, okR := result.Variants["right"]
	if !okL || !okR {
		t.Fatalf("expected left and right variants for spread")
	}
	if left.Bounds().Dy() != 100 || right.Bounds().Dy() != 100 {
		t.Fatalf("expected left/right height 100, got left=%d right=%d", left.Bounds().Dy(), right.Bounds().Dy())
	}
	// With 400 total width and 100 gutter: left 150, right 150
	if left.Bounds().Dx() != 150 || right.Bounds().Dx() != 150 {
		t.Fatalf("unexpected split widths: left=%d right=%d", left.Bounds().Dx(), right.Bounds().Dx())
	}
}
