package export

import (
    "context"
    "fmt"
    "image"
    "image/png"
    "io"
    "os"

    gofpdf "github.com/phpdave11/gofpdf"
    "github.com/go-go-golems/zine-layout/pkg/services"
)

// pixelsToPoints converts device pixels to PDF points at the provided DPI.
func pixelsToPoints(px int, dpi float64) float64 {
    if dpi <= 0 { dpi = 300 }
    return float64(px) * (72.0 / dpi)
}

// writeTempPNG writes an image to a temporary PNG file and returns its path.
func writeTempPNG(img image.Image) (string, error) {
    tmpDir := os.TempDir()
    fp, err := os.CreateTemp(tmpDir, "zine-sheet-*.png")
    if err != nil { return "", err }
    defer fp.Close()
    if err := png.Encode(fp, img); err != nil { return "", err }
    return fp.Name(), nil
}

// SheetsToPDF writes a PDF to out where each sheet becomes one PDF page sized to the sheet.
func SheetsToPDF(ctx context.Context, sheets []*services.SheetResult, dpi float64, out io.Writer) error {
    if len(sheets) == 0 {
        return fmt.Errorf("no sheets provided")
    }

    pdf := gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "pt", Size: gofpdf.SizeType{Wd: 595.28, Ht: 841.89}}) // default A4, will override per page
    pdf.SetCompression(true)

    // Register clean up for temp files
    var temps []string
    defer func() {
        for _, p := range temps { _ = os.Remove(p) }
    }()

    for _, sheet := range sheets {
        if sheet == nil || sheet.Image == nil { return fmt.Errorf("nil sheet image at index %d", sheet.Index) }
        wPt := pixelsToPoints(sheet.Width, dpi)
        hPt := pixelsToPoints(sheet.Height, dpi)
        pdf.AddPageFormat("P", gofpdf.SizeType{Wd: wPt, Ht: hPt})

        // Encode to temp PNG and register
        tmp, err := writeTempPNG(sheet.Image)
        if err != nil { return err }
        temps = append(temps, tmp)

        opts := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
        pdf.ImageOptions(tmp, 0, 0, wPt, hPt, false, opts, 0, "")

        // Early cancellation check
        select { case <-ctx.Done(): return ctx.Err(); default: }
    }

    return pdf.Output(out)
}


