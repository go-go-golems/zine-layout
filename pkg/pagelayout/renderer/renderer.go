package renderer

import (
    "fmt"
    "image"
    "image/color"
    "image/draw"
    "log"

    xdraw "golang.org/x/image/draw"

    "github.com/go-go-golems/zine-layout/pkg/imagelayout"
    "github.com/go-go-golems/zine-layout/pkg/pagelayout"
    "github.com/go-go-golems/zine-layout/pkg/zinelayout"
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
    // LayoutResult provides the crop and target rectangles computed by the
    // imagelayout engine. When provided, the renderer will crop the source
    // image to LayoutResult.SourceRect before placement.
    LayoutResult   *imagelayout.ViewportResult
}

type PageRenderResult struct {
	Full     *image.RGBA
	Variants map[string]image.Image // thumbnail, combined, left, right, full
}

func RenderPage(ctx RenderContext) (*PageRenderResult, error) {
	log.Printf("[pagelayout] RenderPage: Starting render")
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
	log.Printf("[pagelayout] Canvas dimensions: %dx%d pixels (%.2fx%.2f inches @ %.0f DPI)", W, H, ctx.Settings.PageWidthIn, ctx.Settings.PageHeightIn, ctx.Settings.DPI)
	canvas := image.NewRGBA(image.Rect(0, 0, W, H))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: ctx.Background}, image.Point{}, draw.Src)

	target := ctx.Settings.ContentRectPx()
	log.Printf("[pagelayout] Content area: (%d,%d) to (%d,%d) = %dx%d pixels", target.Min.X, target.Min.Y, target.Max.X, target.Max.Y, target.Dx(), target.Dy())

    src := ctx.Source
    srcB := src.Bounds()
    log.Printf("[pagelayout] Source image: %dx%d pixels", srcB.Dx(), srcB.Dy())

    // If a viewport result is provided, crop the source to its SourceRect
    if ctx.LayoutResult != nil {
        log.Printf("[pagelayout] LayoutResult provided: cropping source to SourceRect (%.0f,%.0f,%.0f,%.0f)", 
            ctx.LayoutResult.SourceRect.X, ctx.LayoutResult.SourceRect.Y, ctx.LayoutResult.SourceRect.W, ctx.LayoutResult.SourceRect.H)
        src = cropSourceToRect(src, ctx.LayoutResult.SourceRect)
        srcB = src.Bounds()
        log.Printf("[pagelayout] Cropped source: %dx%d pixels", srcB.Dx(), srcB.Dy())
    }

	switch ctx.Settings.PositioningMode {
	case "absolute":
		// Place the image at absolute inches converted to pixels
		x := ctx.Settings.InchesToPixels(ctx.Settings.ImageXIn)
		y := ctx.Settings.InchesToPixels(ctx.Settings.ImageYIn)
		w := ctx.Settings.InchesToPixels(ctx.Settings.ImageWidthIn)
		h := ctx.Settings.InchesToPixels(ctx.Settings.ImageHeightIn)
		log.Printf("[pagelayout] Absolute mode: placing at (%d,%d) with size %dx%d pixels", x, y, w, h)
		if w <= 0 || h <= 0 { break }
		dst := image.Rect(x, y, x+w, y+h).Intersect(canvas.Bounds())
		if dst.Empty() { break }
		log.Printf("[pagelayout] Destination rect after intersection: (%d,%d) to (%d,%d)", dst.Min.X, dst.Min.Y, dst.Max.X, dst.Max.Y)
		xdraw.CatmullRom.Scale(canvas, dst, src, srcB, draw.Over, nil)
	default: // "fill" and "snap" behave the same for now
		// Scale-cover into target while preserving aspect ratio, then center-crop
		log.Printf("[pagelayout] Fill mode: scaling to cover target area")
		drawIntoTargetCover(canvas, target, src)
	}

    // Optional border: draw around the full page content area
    if ctx.Settings.BorderEnabled {
        c := parseBorderColor(ctx.Settings.BorderColor)
        bt := parseBorderType(ctx.Settings.BorderType)
        zinelayout.DrawBorder(canvas, canvas.Bounds(), c, bt)
    }

	variants := map[string]image.Image{}
	// full
	variants["full"] = canvas
	variants["combined"] = canvas
	log.Printf("[pagelayout] Generated 'full' and 'combined' variants: %dx%d", canvas.Bounds().Dx(), canvas.Bounds().Dy())
	// thumbnail
	thumb := makeThumbnail(canvas, ctx.ThumbnailMaxPx)
	variants["thumbnail"] = thumb
	log.Printf("[pagelayout] Generated 'thumbnail' variant: %dx%d", thumb.Bounds().Dx(), thumb.Bounds().Dy())
	// spread halves
    if ctx.Settings.IsSpread {
        gutterPx := ctx.Settings.InchesToPixels(ctx.Settings.GutterWidthIn)
        log.Printf("[pagelayout] Spread mode: splitting at center with gutter %.2f inches (%d pixels)", 
            ctx.Settings.GutterWidthIn, gutterPx)
        leftImg, rightImg := splitSpread(canvas, ctx.Settings)
        // Add gutter markers to left/right previews at inner edges
        marker := color.RGBA{0, 0, 0, 128}
        if li, ok := leftImg.(*image.RGBA); ok {
            drawDashedVertical(li, li.Bounds().Dx()-1, marker)
        } else {
            li := ensureRGBA(leftImg)
            drawDashedVertical(li, li.Bounds().Dx()-1, marker)
            leftImg = li
        }
        if ri, ok := rightImg.(*image.RGBA); ok {
            drawDashedVertical(ri, 0, marker)
        } else {
            ri := ensureRGBA(rightImg)
            drawDashedVertical(ri, 0, marker)
            rightImg = ri
        }
        variants["left"] = leftImg
        variants["right"] = rightImg
        log.Printf("[pagelayout] Generated 'left' variant: %dx%d", leftImg.Bounds().Dx(), leftImg.Bounds().Dy())
        log.Printf("[pagelayout] Generated 'right' variant: %dx%d", rightImg.Bounds().Dx(), rightImg.Bounds().Dy())
    }

	log.Printf("[pagelayout] RenderPage: Completed successfully, generated %d variants", len(variants))
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
	log.Printf("[pagelayout] Fill mode: scaleX=%.3f, scaleY=%.3f, using scale=%.3f (max)", scaleX, scaleY, scale)
	// New scaled size
	newW := int(float64(srcB.Dx())*scale + 0.5)
	newH := int(float64(srcB.Dy())*scale + 0.5)
	log.Printf("[pagelayout] Scaled image size: %dx%d pixels", newW, newH)
	// Destination rect centered in target
	offX := target.Min.X + (target.Dx()-newW)/2
	offY := target.Min.Y + (target.Dy()-newH)/2
	dstRect := image.Rect(offX, offY, offX+newW, offY+newH)
	log.Printf("[pagelayout] Destination rect: (%d,%d) to (%d,%d)", dstRect.Min.X, dstRect.Min.Y, dstRect.Max.X, dstRect.Max.Y)
	xdraw.CatmullRom.Scale(dst, dstRect, src, srcB, draw.Over, nil)
}

// cropSourceToRect crops the given source image to the floating-point rect
// expressed in source pixel coordinates, clamped to the source bounds.
// If the underlying image type does not support SubImage, a new RGBA is
// allocated and pixels are copied.
func cropSourceToRect(src image.Image, r imagelayout.Rect) image.Image {
    if src == nil { return src }
    srcB := src.Bounds()
    if srcB.Empty() { return src }
    // Convert float rect to integer rectangle and clamp
    x0 := int(r.X + 0.5)
    y0 := int(r.Y + 0.5)
    x1 := int(r.X + r.W + 0.5)
    y1 := int(r.Y + r.H + 0.5)
    crop := image.Rect(x0, y0, x1, y1).Intersect(srcB)
    if crop.Empty() {
        return src
    }
    type subImager interface{ SubImage(r image.Rectangle) image.Image }
    if si, ok := src.(subImager); ok {
        return si.SubImage(crop)
    }
    out := image.NewRGBA(image.Rect(0, 0, crop.Dx(), crop.Dy()))
    draw.Draw(out, out.Bounds(), src, crop.Min, draw.Src)
    return out
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
	log.Printf("[pagelayout] Spread split: center=%d, gutter=%d, leftEnd=%d, rightStart=%d", center, g, leftEnd, rightStart)
	leftRect := image.Rect(0, 0, leftEnd, b.Dy())
	rightRect := image.Rect(rightStart, 0, W, b.Dy())
	log.Printf("[pagelayout] Left rect: (%d,%d) to (%d,%d) = %dx%d", leftRect.Min.X, leftRect.Min.Y, leftRect.Max.X, leftRect.Max.Y, leftRect.Dx(), leftRect.Dy())
	log.Printf("[pagelayout] Right rect: (%d,%d) to (%d,%d) = %dx%d", rightRect.Min.X, rightRect.Min.Y, rightRect.Max.X, rightRect.Max.Y, rightRect.Dx(), rightRect.Dy())
	left := image.NewRGBA(image.Rect(0, 0, leftRect.Dx(), leftRect.Dy()))
	right := image.NewRGBA(image.Rect(0, 0, rightRect.Dx(), rightRect.Dy()))
	draw.Draw(left, left.Bounds(), canvas, leftRect.Min, draw.Src)
	draw.Draw(right, right.Bounds(), canvas, rightRect.Min, draw.Src)
	return left, right
}

// Helpers to parse border options from settings
func ensureRGBA(img image.Image) *image.RGBA {
    if v, ok := img.(*image.RGBA); ok { return v }
    b := img.Bounds()
    out := image.NewRGBA(b)
    draw.Draw(out, b, img, b.Min, draw.Src)
    return out
}

func drawDashedVertical(img *image.RGBA, x int, c color.Color) {
    b := img.Bounds()
    if x < b.Min.X { x = b.Min.X }
    if x >= b.Max.X { x = b.Max.X - 1 }
    dash := 6
    for y := b.Min.Y; y < b.Max.Y; y++ {
        if (y-b.Min.Y)%dash < dash/2 {
            img.Set(x, y, c)
        }
    }
}

func parseBorderColor(s string) color.Color {
    if s == "" {
        return color.RGBA{0,0,0,255}
    }
    // Accept formats: #RRGGBB, #RRGGBBAA, or r,g,b,a
    if len(s) > 0 && s[0] == '#' {
        // Very small parser: only #RRGGBB and #RRGGBBAA
        hex := s[1:]
        var r, g, b, a uint8
        switch len(hex) {
        case 6:
            var rv, gv, bv int
            _, err := fmt.Sscanf(hex, "%02x%02x%02x", &rv, &gv, &bv)
            if err == nil { r, g, b, a = uint8(rv), uint8(gv), uint8(bv), 255 }
        case 8:
            var rv, gv, bv, av int
            _, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &rv, &gv, &bv, &av)
            if err == nil { r, g, b, a = uint8(rv), uint8(gv), uint8(bv), uint8(av) }
        }
        if a == 0 { a = 255 }
        return color.RGBA{r,g,b,a}
    }
    var r, g, b, a int
    if _, err := fmt.Sscanf(s, "%d,%d,%d,%d", &r,&g,&b,&a); err == nil {
        if a == 0 { a = 255 }
        return color.RGBA{uint8(r),uint8(g),uint8(b),uint8(a)}
    }
    return color.RGBA{0,0,0,255}
}

func parseBorderType(s string) zinelayout.BorderType {
    switch s {
    case string(zinelayout.BorderTypeDotted):
        return zinelayout.BorderTypeDotted
    case string(zinelayout.BorderTypeDashed):
        return zinelayout.BorderTypeDashed
    case string(zinelayout.BorderTypeCorner):
        return zinelayout.BorderTypeCorner
    default:
        return zinelayout.BorderTypePlain
    }
}
