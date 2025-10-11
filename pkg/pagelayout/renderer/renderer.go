package renderer

import (
	"image"
	"image/color"
	"image/draw"

	xdraw "golang.org/x/image/draw"

	"github.com/go-go-golems/zine-layout/pkg/pagelayout"
)

// RenderContext groups inputs for page-level rendering.
//
// Source is the decoded image to place on the page. Background defaults to white
// when nil. ThumbnailMaxPx controls the longest side of the generated thumbnail
// variant; if <= 0, a sensible default of 512 is used.
//
// Variant is optional and only relevant to callers that want to request a single
// variant; the renderer itself always computes the full set and returns them.

type RenderContext struct {
	Settings       pagelayout.PageLayoutSettings
	Source         image.Image
	Background     color.Color
	Variant        string
	ThumbnailMaxPx int
}

type PageRenderResult struct {
	Full     *image.RGBA
	Variants map[string]image.Image // thumbnail, combined, left, right, full
}

func RenderPage(ctx RenderContext) (*PageRenderResult, error) {
	if ctx.Background == nil {
		ctx.Background = color.White
	}
	if ctx.ThumbnailMaxPx <= 0 {
		ctx.ThumbnailMaxPx = 512
	}
	if err := ctx.Settings.Canonicalize(); err != nil {
		return nil, err
	}

	W := ctx.Settings.PixelWidth()
	H := ctx.Settings.PixelHeight()
	canvas := image.NewRGBA(image.Rect(0, 0, W, H))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: ctx.Background}, image.Point{}, draw.Src)

	target := ctx.Settings.ContentRectPx()

	src := ctx.Source
	srcB := src.Bounds()

	switch ctx.Settings.PositioningMode {
	case "absolute":
		// Place the image at absolute inches converted to pixels
		x := ctx.Settings.InchesToPixels(ctx.Settings.ImageXIn)
		y := ctx.Settings.InchesToPixels(ctx.Settings.ImageYIn)
		w := ctx.Settings.InchesToPixels(ctx.Settings.ImageWidthIn)
		h := ctx.Settings.InchesToPixels(ctx.Settings.ImageHeightIn)
		if w <= 0 || h <= 0 { break }
		dst := image.Rect(x, y, x+w, y+h).Intersect(canvas.Bounds())
		if dst.Empty() { break }
		xdraw.CatmullRom.Scale(canvas, dst, src, srcB, draw.Over, nil)
	default: // "fill" and "snap" behave the same for now
		// Scale-cover into target while preserving aspect ratio, then center-crop
		drawIntoTargetCover(canvas, target, src)
	}

	variants := map[string]image.Image{}
	// full
	variants["full"] = canvas
	variants["combined"] = canvas
	// thumbnail
	variants["thumbnail"] = makeThumbnail(canvas, ctx.ThumbnailMaxPx)
	// spread halves
	if ctx.Settings.IsSpread {
		leftImg, rightImg := splitSpread(canvas, ctx.Settings)
		variants["left"] = leftImg
		variants["right"] = rightImg
	}

	return &PageRenderResult{Full: canvas, Variants: variants}, nil
}

func drawIntoTargetCover(dst *image.RGBA, target image.Rectangle, src image.Image) {
	srcB := src.Bounds()
	if srcB.Empty() || target.Empty() {
		return
	}
	// Compute scale to cover
	scaleX := float64(target.Dx()) / float64(srcB.Dx())
	scaleY := float64(target.Dy()) / float64(srcB.Dy())
	scale := scaleX
	if scaleY > scale { scale = scaleY }
	// New scaled size
	newW := int(float64(srcB.Dx())*scale + 0.5)
	newH := int(float64(srcB.Dy())*scale + 0.5)
	// Destination rect centered in target
	offX := target.Min.X + (target.Dx()-newW)/2
	offY := target.Min.Y + (target.Dy()-newH)/2
	dstRect := image.Rect(offX, offY, offX+newW, offY+newH)
	xdraw.CatmullRom.Scale(dst, dstRect, src, srcB, draw.Over, nil)
}

func makeThumbnail(src image.Image, maxSide int) image.Image {
	if maxSide <= 0 { maxSide = 512 }
	srcB := src.Bounds()
	w := srcB.Dx()
	h := srcB.Dy()
	if w <= maxSide && h <= maxSide { return src }
	var newW, newH int
	if w >= h {
		newW = maxSide
		newH = int(float64(h) * float64(maxSide) / float64(w) + 0.5)
	} else {
		newH = maxSide
		newW = int(float64(w) * float64(maxSide) / float64(h) + 0.5)
	}
	out := image.NewRGBA(image.Rect(0, 0, newW, newH))
	xdraw.CatmullRom.Scale(out, out.Bounds(), src, srcB, draw.Over, nil)
	return out
}

func splitSpread(canvas *image.RGBA, s pagelayout.PageLayoutSettings) (image.Image, image.Image) {
	b := canvas.Bounds()
	W := b.Dx()
	// Split at center; optionally consider gutter to shrink both halves
	g := s.InchesToPixels(s.GutterWidthIn)
	center := W / 2
	leftEnd := center - g/2
	if leftEnd < 0 { leftEnd = center }
	rightStart := center + g/2
	if rightStart > W { rightStart = center }
	leftRect := image.Rect(0, 0, leftEnd, b.Dy())
	rightRect := image.Rect(rightStart, 0, W, b.Dy())
	left := image.NewRGBA(image.Rect(0, 0, leftRect.Dx(), leftRect.Dy()))
	right := image.NewRGBA(image.Rect(0, 0, rightRect.Dx(), rightRect.Dy()))
	draw.Draw(left, left.Bounds(), canvas, leftRect.Min, draw.Src)
	draw.Draw(right, right.Bounds(), canvas, rightRect.Min, draw.Src)
	return left, right
}
