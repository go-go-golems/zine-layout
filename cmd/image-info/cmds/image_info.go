package cmds

import (
    "bufio"
    "context"
    "image"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
    "math"
    "os"
    "path/filepath"
    "strings"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/go-go-golems/glazed/pkg/types"
    "github.com/pkg/errors"
)

type ImageInfoCommand struct {
    *cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*ImageInfoCommand)(nil)

type ImageInfoSettings struct {
    Files []*parameters.FileData `glazed.parameter:"files"`
    // Optional: override DPI if metadata missing
    DefaultDPI float64 `glazed.parameter:"default-dpi"`
}

func NewImageInfoCommand() (*ImageInfoCommand, error) {
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }

    desc := cmds.NewCommandDescription(
        "image-info",
        cmds.WithShort("Report image size, margins, and DPI for images"),
        cmds.WithLong(`Analyze images and detect uniform-color margins on each side. Outputs width/height (px), margins (px), DPI and sizes in inches.`),
        cmds.WithArguments(
            parameters.NewParameterDefinition("files", parameters.ParameterTypeFileList, parameters.WithHelp("Image files to analyze"), parameters.WithRequired(true)),
        ),
        cmds.WithFlags(
            parameters.NewParameterDefinition("default-dpi", parameters.ParameterTypeFloat, parameters.WithDefault(300.0), parameters.WithHelp("Fallback DPI if metadata is missing")),
        ),
        cmds.WithLayersList(glazedLayer),
    )

    return &ImageInfoCommand{CommandDescription: desc}, nil
}

func (c *ImageInfoCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &ImageInfoSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return errors.Wrap(err, "parse settings")
    }

    for _, fd := range s.Files {
        info, err := analyzeImage(fd.AbsolutePath, s.DefaultDPI)
        if err != nil {
            // Emit an error row for this file
            row := types.NewRow(
                types.MRP("image", fd.Path),
                types.MRP("error", err.Error()),
            )
            _ = gp.AddRow(ctx, row)
            continue
        }

        row := types.NewRow(
            types.MRP("image", info.Name),
            types.MRP("width_px", info.Width),
            types.MRP("height_px", info.Height),
            types.MRP("margin_top_px", info.MarginTop),
            types.MRP("margin_right_px", info.MarginRight),
            types.MRP("margin_bottom_px", info.MarginBottom),
            types.MRP("margin_left_px", info.MarginLeft),
            types.MRP("dpi_x", info.DPIX),
            types.MRP("dpi_y", info.DPIY),
            types.MRP("width_in", round4(info.WidthIn)),
            types.MRP("height_in", round4(info.HeightIn)),
            types.MRP("margin_top_in", round4(info.MarginTopIn)),
            types.MRP("margin_right_in", round4(info.MarginRightIn)),
            types.MRP("margin_bottom_in", round4(info.MarginBottomIn)),
            types.MRP("margin_left_in", round4(info.MarginLeftIn)),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    return nil
}

type imageInfo struct {
    Name          string
    Width         int
    Height        int
    MarginTop     int
    MarginRight   int
    MarginBottom  int
    MarginLeft    int
    DPIX          float64
    DPIY          float64
    WidthIn       float64
    HeightIn      float64
    MarginTopIn   float64
    MarginRightIn float64
    MarginBottomIn float64
    MarginLeftIn  float64
}

func analyzeImage(path string, defaultDPI float64) (*imageInfo, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()

    // Keep a buffered reader for magic sniffing when needed
    br := bufio.NewReader(f)
    // Decode fully (we need pixel data to detect margins)
    img, format, err := image.Decode(br)
    if err != nil { return nil, errors.Wrap(err, "decode image") }

    // After Decode, the reader consumed data; reopen for metadata reads if needed
    _ = f.Close()

    // Reopen for header parsing
    f2, err := os.Open(path)
    if err != nil { return nil, err }
    defer f2.Close()
    br2 := bufio.NewReader(f2)

    dpix, dpiy := extractDPI(br2, format, defaultDPI)

    bounds := img.Bounds()
    w := bounds.Dx()
    h := bounds.Dy()

    mt, mr, mb, ml := detectUniformMargins(img)

    // Convert to inches using dpix/dpiy
    win := float64(w) / dpix
    hin := float64(h) / dpiy
    mtIn := float64(mt) / dpiy
    mrIn := float64(mr) / dpix
    mbIn := float64(mb) / dpiy
    mlIn := float64(ml) / dpix

    return &imageInfo{
        Name: filepath.Base(path),
        Width: w,
        Height: h,
        MarginTop: mt,
        MarginRight: mr,
        MarginBottom: mb,
        MarginLeft: ml,
        DPIX: dpix,
        DPIY: dpiy,
        WidthIn: win,
        HeightIn: hin,
        MarginTopIn: mtIn,
        MarginRightIn: mrIn,
        MarginBottomIn: mbIn,
        MarginLeftIn: mlIn,
    }, nil
}

// extractDPI tries to read DPI from JPEG EXIF (ResolutionUnit/XResolution/YResolution) or PNG pHYs.
// Falls back to defaultDPI for both axes if not found. Returns dpix, dpiy.
func extractDPI(r *bufio.Reader, format string, defaultDPI float64) (float64, float64) {
    // Default if nothing found
    dpix := defaultDPI
    dpiy := defaultDPI

    // We need to peek file header to decide specific handling
    // For PNG, search for pHYs chunk to get pixels per unit
    // For JPEG, the standard library doesn't expose EXIF; we'll parse common APP0 JFIF density if present

    // Read a small prefix for format detection without consuming
    // Note: r is already positioned at start
    header, _ := r.Peek(16)
    if len(header) >= 8 && string(header[:8]) == "\x89PNG\r\n\x1a\n" {
        // PNG: scan chunks for pHYs
        if x, y, ok := parsePNGpHYs(r); ok {
            dpix = x
            dpiy = y
        }
        return dpix, dpiy
    }

    // JPEG: look for JFIF APP0 density fields
    if len(header) >= 2 && header[0] == 0xFF && header[1] == 0xD8 {
        if x, y, ok := parseJPEGJFIFDensity(r); ok {
            dpix = x
            dpiy = y
        }
        return dpix, dpiy
    }

    // Fallback: if we know format string from decoder
    switch strings.ToLower(format) {
    case "png":
        // if not found because peek missed, try again
        if x, y, ok := parsePNGpHYs(r); ok { return x, y }
    case "jpeg", "jpg":
        if x, y, ok := parseJPEGJFIFDensity(r); ok { return x, y }
    }
    return dpix, dpiy
}

// parsePNGpHYs parses the PNG pHYs chunk to get DPI. It expects r at file start.
func parsePNGpHYs(r *bufio.Reader) (float64, float64, bool) {
    // Reset by reopening reader at file start expected by caller
    // Read signature
    sig, err := r.Peek(8)
    if err != nil { return 0, 0, false }
    if string(sig) != "\x89PNG\r\n\x1a\n" { return 0, 0, false }

    // Discard signature
    _, _ = r.Discard(8)

    // Iterate chunks
    for {
        // Read chunk length (4), type (4)
        hdr, err := r.Peek(8)
        if err != nil { return 0, 0, false }
        // Length is big-endian
        length := int(uint32(hdr[0])<<24 | uint32(hdr[1])<<16 | uint32(hdr[2])<<8 | uint32(hdr[3]))
        ctype := string(hdr[4:8])
        // Consume header
        _, _ = r.Discard(8)

        if ctype == "pHYs" {
            if length != 9 { return 0, 0, false }
            data := make([]byte, 9)
            if _, err := r.Read(data); err != nil { return 0, 0, false }
            // Skip CRC
            _, _ = r.Discard(4)
            // Pixels per unit, X and Y, 4 bytes each big-endian
            ppux := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
            ppuy := uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])
            unit := data[8] // 1 = meter
            if unit == 1 {
                // Convert pixels per meter to DPI (dots per inch)
                const inchesPerMeter = 39.37007874015748
                xdpi := float64(ppux) / inchesPerMeter
                ydpi := float64(ppuy) / inchesPerMeter
                return xdpi, ydpi, true
            }
            // Unit is unknown: cannot convert
            return 0, 0, false
        }

        // Skip chunk data + CRC
        toSkip := length + 4
        if _, err := r.Discard(toSkip); err != nil { return 0, 0, false }
        if ctype == "IEND" { break }
    }
    return 0, 0, false
}

// parseJPEGJFIFDensity parses JFIF APP0 density fields to DPI.
func parseJPEGJFIFDensity(r *bufio.Reader) (float64, float64, bool) {
    // Expect SOI 0xFFD8
    b, err := r.Peek(2)
    if err != nil || !(b[0] == 0xFF && b[1] == 0xD8) { return 0, 0, false }
    // Consume SOI
    _, _ = r.Discard(2)
    for {
        // Read marker 0xFFxx
        m1, err := r.ReadByte()
        if err != nil { return 0, 0, false }
        if m1 != 0xFF { continue }
        // Skip fill bytes 0xFF
        var m2 byte
        for {
            m2, err = r.ReadByte()
            if err != nil { return 0, 0, false }
            if m2 != 0xFF { break }
        }
        if m2 == 0xD9 { // EOI
            return 0, 0, false
        }
        // Read segment length (2 bytes, includes length bytes)
        lb, err := r.Peek(2)
        if err != nil { return 0, 0, false }
        segLen := int(lb[0])<<8 | int(lb[1])
        // Consume length bytes
        _, _ = r.Discard(2)

        if m2 == 0xE0 { // APP0
            // Read segment data
            data := make([]byte, segLen-2)
            if _, err := r.Read(data); err != nil { return 0, 0, false }
            if len(data) >= 14 && string(data[:5]) == "JFIF\x00" {
                // Density unit at offset 7: 0=no units, 1=dpi, 2=dpcm
                unit := data[7]
                xden := int(data[8])<<8 | int(data[9])
                yden := int(data[10])<<8 | int(data[11])
                switch unit {
                case 1: // dpi
                    return float64(xden), float64(yden), true
                case 2: // dpcm → convert to dpi
                    return float64(xden) * 2.54, float64(yden) * 2.54, true
                default:
                    // unitless: cannot convert reliably
                    return 0, 0, false
                }
            }
        } else {
            // Skip this segment
            if _, err := r.Discard(segLen-2); err != nil { return 0, 0, false }
        }
    }
}

func detectUniformMargins(img image.Image) (top, right, bottom, left int) {
    b := img.Bounds()
    w := b.Dx()
    h := b.Dy()

    // Helper to get RGBA at pixel
    rgbaAt := func(x, y int) (r, g, bl, a uint32) {
        cr, cg, cb, ca := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
        // Normalize to 8-bit
        return cr >> 8, cg >> 8, cb >> 8, ca >> 8
    }

    // Reference colors from outermost edges
    // We consider a side to have uniform margin if entire rows/columns share the same color, expanding inward.

    // Top
    top = 0
    for y := 0; y < h; y++ {
        r0, g0, b0, a0 := rgbaAt(0, y)
        uniform := true
        for x := 1; x < w; x++ {
            r, g, b, a := rgbaAt(x, y)
            if r != r0 || g != g0 || b != b0 || a != a0 {
                uniform = false
                break
            }
        }
        if uniform { top++ } else { break }
    }

    // Bottom
    bottom = 0
    for y := h-1; y >= 0; y-- {
        r0, g0, b0, a0 := rgbaAt(0, y)
        uniform := true
        for x := 1; x < w; x++ {
            r, g, b, a := rgbaAt(x, y)
            if r != r0 || g != g0 || b != b0 || a != a0 {
                uniform = false
                break
            }
        }
        if uniform { bottom++ } else { break }
    }

    // Left
    left = 0
    for x := 0; x < w; x++ {
        r0, g0, b0, a0 := rgbaAt(x, 0)
        uniform := true
        for y := 1; y < h; y++ {
            r, g, b, a := rgbaAt(x, y)
            if r != r0 || g != g0 || b != b0 || a != a0 {
                uniform = false
                break
            }
        }
        if uniform { left++ } else { break }
    }

    // Right
    right = 0
    for x := w-1; x >= 0; x-- {
        r0, g0, b0, a0 := rgbaAt(x, 0)
        uniform := true
        for y := 1; y < h; y++ {
            r, g, b, a := rgbaAt(x, y)
            if r != r0 || g != g0 || b != b0 || a != a0 {
                uniform = false
                break
            }
        }
        if uniform { right++ } else { break }
    }

    // Clamp margins to not exceed image
    if top+bottom > h { top = h/2; bottom = h - top }
    if left+right > w { left = w/2; right = w - left }
    return
}

func round4(v float64) float64 {
    return math.Round(v*10000) / 10000
}


