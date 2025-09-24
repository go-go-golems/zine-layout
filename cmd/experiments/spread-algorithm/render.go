package main

import (
    "context"
    "fmt"
    "image"
    "image/color"
    imagedraw "image/draw"
    "image/jpeg"
    "image/png"
    "os"
    "path/filepath"
    "strings"

    xdraw "golang.org/x/image/draw"
    "github.com/rs/zerolog/log"
    "time"
    "math"
    simple "github.com/go-go-golems/zine-layout/pkg/spread/simple"
)

type RenderInfo struct {
    Index         int
    SpreadName    string
    ImageBaseName string
    Format        string
    Quality       int
    Background    string
    FilenameTmpl  string
    OutputDir     string
    PNGLevel      string // default|speed|max|none
    Scaler        string // fast|quality
    ParallelEncode bool
}

type SpreadOutput struct {
    Name        string
    PanelFiles  []string
    PanelLabels []string
    Logs        []string
    Timestamp   interface{}
}

func backgroundColor(bg string) (color.Color, bool) {
    if bg == "" || strings.EqualFold(bg, "transparent") {
        return color.Transparent, true
    }
    // accept #RRGGBB or #RRGGBBAA
    if strings.HasPrefix(bg, "#") {
        hex := strings.TrimPrefix(bg, "#")
        var r, g, b, a uint8 = 0, 0, 0, 255
        if len(hex) == 6 {
            var rv, gv, bv int
            if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &rv, &gv, &bv); err == nil {
                r, g, b = uint8(rv), uint8(gv), uint8(bv)
                return color.NRGBA{R: r, G: g, B: b, A: a}, true
            }
        } else if len(hex) == 8 {
            var rv, gv, bv, av int
            if _, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &rv, &gv, &bv, &av); err == nil {
                r, g, b, a = uint8(rv), uint8(gv), uint8(bv), uint8(av)
                return color.NRGBA{R: r, G: g, B: b, A: a}, true
            }
        }
    }
    return color.White, false
}

func resolveFilename(tmpl, spreadName, panel, imageBase string, idx int, ext string) string {
    if tmpl == "" {
        tmpl = "{index:03d}-{name}-{panel}.{ext}"
    }
    // simple replacements (no format spec other than {index:03d})
    name := tmpl
    // support {index:03d}
    name = strings.ReplaceAll(name, "{index:03d}", fmt.Sprintf("%03d", idx))
    name = strings.ReplaceAll(name, "{index}", fmt.Sprintf("%d", idx))
    name = strings.ReplaceAll(name, "{name}", sanitize(spreadName))
    name = strings.ReplaceAll(name, "{panel}", panel)
    name = strings.ReplaceAll(name, "{image_basename}", sanitize(imageBase))
    name = strings.ReplaceAll(name, "{ext}", ext)
    return name
}

func sanitize(s string) string {
    s = strings.TrimSpace(s)
    s = strings.ReplaceAll(s, " ", "-")
    s = strings.ReplaceAll(s, "/", "-")
    s = strings.ReplaceAll(s, "\\", "-")
    s = strings.ToLower(s)
    return s
}

func RenderSingle(ctx context.Context, src image.Image, res simple.Result, info RenderInfo) (string, error) {
    size := res.ExportSingle
    if size == nil {
        return "", fmt.Errorf("ExportSingle is nil")
    }
    w, h := size.W, size.H
    dst := image.NewNRGBA(image.Rect(0, 0, w, h))

    // Fill background
    if c, ok := backgroundColor(info.Background); ok {
        imagedraw.Draw(dst, dst.Bounds(), &image.Uniform{C: c}, image.Point{}, imagedraw.Src)
    }

    // Scale src_rect → dst_rect_global
    t0 := time.Now()
    drawRect := res.DstRectGlobal
    scaleAndDrawWith(info, src, dst, res.SrcRectGlobal, drawRect)
    log.Debug().Dur("scale_dur", time.Since(t0)).Int("w", w).Int("h", h).Msg("scaled single")

    ext := strings.ToLower(info.Format)
    if ext == "jpeg" {
        ext = "jpg"
    }
    fname := resolveFilename(info.FilenameTmpl, info.SpreadName, "single", info.ImageBaseName, info.Index, ext)
    fp := filepath.Join(info.OutputDir, fname)
    t1 := time.Now()
    if err := saveImage(fp, dst, ext, info.Quality, info.PNGLevel); err != nil {
        return "", err
    }
    log.Debug().Dur("encode_dur", time.Since(t1)).Str("file", fp).Msg("saved single")
    return fp, nil
}

func RenderSpread(ctx context.Context, src image.Image, res simple.Result, info RenderInfo) (string, string, error) {
    if res.ExportSpread == nil || res.LeftPanel == nil || res.RightPanel == nil || res.DstRectLeft == nil || res.DstRectRight == nil {
        return "", "", fmt.Errorf("invalid spread result")
    }

    // Left
    lw, lh := res.ExportSpread.Left.W, res.ExportSpread.Left.H
    leftImg := image.NewNRGBA(image.Rect(0, 0, lw, lh))
    // Right
    rw, rh := res.ExportSpread.Right.W, res.ExportSpread.Right.H
    rightImg := image.NewNRGBA(image.Rect(0, 0, rw, rh))

    // Fill background
    if c, ok := backgroundColor(info.Background); ok {
        imagedraw.Draw(leftImg, leftImg.Bounds(), &image.Uniform{C: c}, image.Point{}, imagedraw.Src)
        imagedraw.Draw(rightImg, rightImg.Bounds(), &image.Uniform{C: c}, image.Point{}, imagedraw.Src)
    }

    // Compute clip rects for inner gutter margins
    pageW := res.ExportSpread.Left.W
    pageH := res.ExportSpread.Left.H
    leftVisibleW := int(math.Round(res.LeftPanel.W))
    rightVisibleW := int(math.Round(res.RightPanel.W))
    mLeft := pageW - leftVisibleW
    mRight := pageW - rightVisibleW
    if mLeft < 0 { mLeft = 0 }
    if mRight < 0 { mRight = 0 }
    clipLeft := image.Rect(0, 0, pageW - mLeft, pageH)
    clipRight := image.Rect(mRight, 0, pageW, pageH)
    log.Debug().Int("pageW", pageW).Int("mLeft", mLeft).Int("mRight", mRight).Msg("clip rects for pages")

    // Render sides concurrently
    type resErr struct{ err error }
    ch := make(chan resErr, 2)
    t0 := time.Now()
    go func() { scaleAndDrawWithClip(info, src, leftImg, res.SrcRectGlobal, *res.DstRectLeft, clipLeft); ch <- resErr{} }()
    go func() { scaleAndDrawWithClip(info, src, rightImg, res.SrcRectGlobal, *res.DstRectRight, clipRight); ch <- resErr{} }()
    // wait
    <-ch
    <-ch
    log.Debug().Dur("scale_dur", time.Since(t0)).Int("lw", lw).Int("lh", lh).Int("rw", rw).Int("rh", rh).Msg("scaled spread panels")

    ext := strings.ToLower(info.Format)
    if ext == "jpeg" {
        ext = "jpg"
    }
    leftName := resolveFilename(info.FilenameTmpl, info.SpreadName, "left", info.ImageBaseName, info.Index, ext)
    rightName := resolveFilename(info.FilenameTmpl, info.SpreadName, "right", info.ImageBaseName, info.Index, ext)
    leftPath := filepath.Join(info.OutputDir, leftName)
    rightPath := filepath.Join(info.OutputDir, rightName)
    if info.ParallelEncode {
        encCh := make(chan resErr, 2)
        t1 := time.Now()
        go func() { encCh <- resErr{saveImage(leftPath, leftImg, ext, info.Quality, info.PNGLevel)} }()
        go func() { encCh <- resErr{saveImage(rightPath, rightImg, ext, info.Quality, info.PNGLevel)} }()
        e1 := <-encCh
        e2 := <-encCh
        log.Debug().Dur("encode_dur", time.Since(t1)).Msg("encoded spread panels")
        if e1.err != nil { return "", "", e1.err }
        if e2.err != nil { return "", "", e2.err }
    } else {
        t1 := time.Now()
        if err := saveImage(leftPath, leftImg, ext, info.Quality, info.PNGLevel); err != nil {
            return "", "", err
        }
        if err := saveImage(rightPath, rightImg, ext, info.Quality, info.PNGLevel); err != nil {
            return "", "", err
        }
        log.Debug().Dur("encode_dur", time.Since(t1)).Msg("encoded spread panels (serial)")
    }
    return leftPath, rightPath, nil
}

func saveImage(path string, img image.Image, ext string, quality int, pngLevel string) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return err
    }
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    switch ext {
    case "png":
        enc := png.Encoder{CompressionLevel: mapPNGLevel(pngLevel)}
        return enc.Encode(f, img)
    case "jpg":
        if quality <= 0 || quality > 100 {
            quality = 90
        }
        opts := &jpeg.Options{Quality: quality}
        return jpeg.Encode(f, img, opts)
    default:
        return fmt.Errorf("unsupported format: %s", ext)
    }
}

func mapPNGLevel(level string) png.CompressionLevel {
    switch strings.ToLower(level) {
    case "speed":
        return png.BestSpeed
    case "max", "best":
        return png.BestCompression
    case "none":
        return png.NoCompression
    case "default", "":
        fallthrough
    default:
        return png.DefaultCompression
    }
}

func scaleAndDrawWith(info RenderInfo, src image.Image, dst *image.NRGBA, srcRect simple.Rect, dstRect simple.Rect) {
    // Extract crop from source
    // Map float rects to pixel rectangles
    sRect := image.Rect(
        int(srcRect.X+0.5),
        int(srcRect.Y+0.5),
        int(srcRect.X+srcRect.W+0.5),
        int(srcRect.Y+srcRect.H+0.5),
    )
    // Destination rectangle in dst image
    dRect := image.Rect(
        int(dstRect.X+0.5),
        int(dstRect.Y+0.5),
        int(dstRect.X+dstRect.W+0.5),
        int(dstRect.Y+dstRect.H+0.5),
    )

    // Create a temp image of the exact dst rect size for high-quality scaling
    tmp := image.NewNRGBA(image.Rect(0, 0, dRect.Dx(), dRect.Dy()))
    // Choose scaler
    var scaler xdraw.Scaler
    if strings.ToLower(info.Scaler) == "fast" {
        scaler = xdraw.ApproxBiLinear
    } else {
        scaler = xdraw.CatmullRom
    }
    scaler.Scale(tmp, tmp.Bounds(), src, sRect, xdraw.Over, nil)

    // Blit tmp into dst at dRect
    imagedraw.Draw(dst, dRect, tmp, image.Point{}, imagedraw.Over)
    log.Debug().Int("dx", dRect.Min.X).Int("dy", dRect.Min.Y).Int("dw", dRect.Dx()).Int("dh", dRect.Dy()).Msg("scaleAndDraw done")
}

func scaleAndDrawWithClip(info RenderInfo, src image.Image, dst *image.NRGBA, srcRect simple.Rect, dstRect simple.Rect, clip image.Rectangle) {
    // Intersect destination with clip
    dRect := image.Rect(
        int(dstRect.X+0.5),
        int(dstRect.Y+0.5),
        int(dstRect.X+dstRect.W+0.5),
        int(dstRect.Y+dstRect.H+0.5),
    )
    dRect = dRect.Intersect(clip)
    if dRect.Empty() {
        return
    }
    // Compute proportional source sub-rectangle corresponding to clipped destination
    // Map dRect relative to original dstRect
    origD := image.Rect(
        int(dstRect.X+0.5),
        int(dstRect.Y+0.5),
        int(dstRect.X+dstRect.W+0.5),
        int(dstRect.Y+dstRect.H+0.5),
    )
    scaleX := srcRect.W / float64(origD.Dx())
    scaleY := srcRect.H / float64(origD.Dy())
    // Offset within orig dst
    offX := dRect.Min.X - origD.Min.X
    offY := dRect.Min.Y - origD.Min.Y
    sRect := image.Rect(
        int(srcRect.X+float64(offX)*scaleX+0.5),
        int(srcRect.Y+float64(offY)*scaleY+0.5),
        int(srcRect.X+float64(offX+dRect.Dx())*scaleX+0.5),
        int(srcRect.Y+float64(offY+dRect.Dy())*scaleY+0.5),
    )

    // Scale into temp and blit
    tmp := image.NewNRGBA(image.Rect(0, 0, dRect.Dx(), dRect.Dy()))
    var scaler xdraw.Scaler
    if strings.ToLower(info.Scaler) == "fast" {
        scaler = xdraw.ApproxBiLinear
    } else {
        scaler = xdraw.CatmullRom
    }
    scaler.Scale(tmp, tmp.Bounds(), src, sRect, xdraw.Over, nil)
    imagedraw.Draw(dst, dRect, tmp, image.Point{}, imagedraw.Over)
}


