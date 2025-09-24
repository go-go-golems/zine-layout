package engine

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	xdraw "golang.org/x/image/draw"
)

type OutputFile struct {
	Panel string
	Path  string
}

// Render saves the computed spread into one or more files according to export settings.
func Render(result Result, src image.Image, outPaths map[string]string) ([]OutputFile, []TraceEntry, error) {
	trace := make([]TraceEntry, 0, 16)
	export := result.Settings.Export
	bg, err := parseBackground(export.Background, export.Format)
	if err != nil {
		return nil, trace, err
	}
	addTrace(&trace, "render", "background %s format %s (quality %d)", export.Background, export.Format, export.Quality)

	canvas := image.NewNRGBA(image.Rect(0, 0, int(result.PaperSizePx.Width), int(result.PaperSizePx.Height)))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	cropRect := image.Rect(result.SourceCrop.X, result.SourceCrop.Y, result.SourceCrop.X+result.SourceCrop.Width, result.SourceCrop.Y+result.SourceCrop.Height)
	cropped := cropImage(src, cropRect)
	addTrace(&trace, "render", "cropping source to rect (%d,%d,%d,%d)", cropRect.Min.X, cropRect.Min.Y, cropRect.Max.X, cropRect.Max.Y)

	targetW := int(math.Round(result.ImageDisplay.Width))
	targetH := int(math.Round(result.ImageDisplay.Height))
	if targetW <= 0 || targetH <= 0 {
		return nil, trace, fmt.Errorf("target size is zero")
	}
	scaled := image.NewNRGBA(image.Rect(0, 0, targetW, targetH))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), cropped, cropped.Bounds(), draw.Src, nil)
	addTrace(&trace, "render", "scaled crop to %dx%d px", targetW, targetH)

	dstPoint := image.Point{X: int(math.Round(result.VirtualPosition.X)), Y: int(math.Round(result.VirtualPosition.Y))}
	draw.Draw(canvas, image.Rectangle{Min: dstPoint, Max: dstPoint.Add(scaled.Bounds().Size())}, scaled, image.Point{}, draw.Over)
	addTrace(&trace, "render", "drawn at destination (%d,%d)", dstPoint.X, dstPoint.Y)

	outputs := []OutputFile{}

	if !result.Settings.IsSpread {
		path, ok := outPaths["single"]
		if !ok {
			return nil, trace, fmt.Errorf("output path for single panel missing")
		}
		if err := writeImage(canvas, export.Format, export.Quality, result.Settings.DPI, path); err != nil {
			return nil, trace, err
		}
		outputs = append(outputs, OutputFile{Panel: "single", Path: path})
		addTrace(&trace, "render", "wrote single panel to %s", path)
		return outputs, trace, nil
	}

	gutterHalf := result.GutterPx / 2
	leftWidth := result.EffectiveRect.X + result.EffectiveRect.Width + gutterHalf
	if leftWidth > canvas.Bounds().Dx() {
		leftWidth = canvas.Bounds().Dx()
	}
	rightX := leftWidth

	leftRect := image.Rect(0, 0, leftWidth, canvas.Bounds().Dy())
	rightRect := image.Rect(rightX, 0, canvas.Bounds().Dx(), canvas.Bounds().Dy())
	addTrace(&trace, "render", "split canvas: left width %d, right starts at %d", leftWidth, rightX)

	leftImg := cropImage(canvas, leftRect)
	rightImg := cropImage(canvas, rightRect)

	leftPath, ok := outPaths["left"]
	if !ok {
		return nil, trace, fmt.Errorf("output path for left panel missing")
	}
	if err := writeImage(leftImg, export.Format, export.Quality, result.Settings.DPI, leftPath); err != nil {
		return nil, trace, err
	}
	outputs = append(outputs, OutputFile{Panel: "left", Path: leftPath})
	addTrace(&trace, "render", "wrote left panel to %s", leftPath)

	rightPath, ok := outPaths["right"]
	if !ok {
		return nil, trace, fmt.Errorf("output path for right panel missing")
	}
	if err := writeImage(rightImg, export.Format, export.Quality, result.Settings.DPI, rightPath); err != nil {
		return nil, trace, err
	}
	outputs = append(outputs, OutputFile{Panel: "right", Path: rightPath})
	addTrace(&trace, "render", "wrote right panel to %s", rightPath)
	return outputs, trace, nil
}

func cropImage(src image.Image, rect image.Rectangle) image.Image {
	clipped := rect.Intersect(src.Bounds())
	cropped := image.NewNRGBA(image.Rect(0, 0, clipped.Dx(), clipped.Dy()))
	draw.Draw(cropped, cropped.Bounds(), src, clipped.Min, draw.Src)
	return cropped
}

func writeImage(img image.Image, format string, quality int, dpi float64, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	switch strings.ToLower(format) {
	case "png":
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return png.Encode(f, img)
	case "jpg":
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)
		opts := &jpeg.Options{Quality: quality}
		return jpeg.Encode(f, rgba, opts)
	case "pdf":
		return writePDF(img, dpi, path)
	default:
		return fmt.Errorf("unsupported format %s", format)
	}
}

func writePDF(img image.Image, dpi float64, path string) error {
	bounds := img.Bounds()
	widthPx := bounds.Dx()
	heightPx := bounds.Dy()
	if dpi <= 0 {
		dpi = 300
	}
	widthPt := float64(widthPx) / dpi * 72.0
	heightPt := float64(heightPx) / dpi * 72.0
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "pt",
		Size:    gofpdf.SizeType{Wd: widthPt, Ht: heightPt},
	})
	pdf.AddPage()

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	opt := gofpdf.ImageOptions{ImageType: "PNG"}
	pdf.RegisterImageOptionsReader("page", opt, &buf)
	pdf.ImageOptions("page", 0, 0, widthPt, heightPt, false, opt, 0, "")

	return pdf.OutputFileAndClose(path)
}

func parseBackground(value string, format string) (color.Color, error) {
	v := strings.TrimSpace(strings.ToLower(value))
	if v == "" || v == "transparent" {
		if strings.ToLower(format) == "jpg" {
			return color.White, nil
		}
		return color.NRGBA{0, 0, 0, 0}, nil
	}
	if strings.HasPrefix(v, "#") {
		hex := strings.TrimPrefix(v, "#")
		if len(hex) == 6 {
			r, err := strconv.ParseUint(hex[0:2], 16, 8)
			if err != nil {
				return nil, err
			}
			g, err := strconv.ParseUint(hex[2:4], 16, 8)
			if err != nil {
				return nil, err
			}
			b, err := strconv.ParseUint(hex[4:6], 16, 8)
			if err != nil {
				return nil, err
			}
			return color.NRGBA{uint8(r), uint8(g), uint8(b), 255}, nil
		}
	}
	return nil, fmt.Errorf("invalid background %q", value)
}
