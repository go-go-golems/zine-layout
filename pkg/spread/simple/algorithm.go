package simple

import (
    "fmt"
    "math"
    "os"
)

type CropRatio struct {
    W float64
    H float64
}

type Point struct {
    X float64
    Y float64
}

type Rect struct {
    X float64
    Y float64
    W float64
    H float64
}

type Inputs struct {
    // Source image (pixels)
    SrcW float64
    SrcH float64

    // Paper and layout (inches)
    PaperWIn     float64
    PaperHIn     float64
    Orientation  string // portrait | landscape
    MarginTopIn    float64
    MarginRightIn  float64
    MarginBottomIn float64
    MarginLeftIn   float64
    DPI          float64

    // Spread / gutter
    IsSpread bool
    GutterIn float64

    // Crop & scale
    CropRatio  *CropRatio // nil => original
    CropToFill bool
    UserScale  float64 // s >= 0

    // Position
    ImagePosition Point
    PositionUnits string // px | normalized
}

type Result struct {
    // Step 1
    SrcRectGlobal Rect // in S
    DstRectGlobal Rect // in L

    ContentRect       Rect
    EffectiveSpreadW  float64
    GutterPx          float64

    LeftPanel   *Rect
    RightPanel  *Rect
    DstRectLeft  *Rect
    DstRectRight *Rect

    // Export sizes
    ExportSingle *Size
    ExportSpread *SpreadSizes
}

type Size struct {
    W int
    H int
}

type SpreadSizes struct {
    Left  Size
    Right Size
}

type Trace struct {
    Lines       []string
    EnableStdout bool
    UseZerolog  bool
    Structured  *PlacementTrace
    // Render traces populated by RenderSingle/RenderSpread
    RenderSingle *RenderSingleTrace
    RenderSpread *RenderSpreadTrace
}

func (t *Trace) Logf(format string, args ...interface{}) {
    s := fmt.Sprintf(format, args...)
    t.Lines = append(t.Lines, s)
    if t.EnableStdout {
        if t.UseZerolog {
            logMsg(s)
        } else {
            fmt.Fprintln(os.Stdout, s)
        }
    }
}

func logMsg(msg string) {
    // lightweight wrapper to avoid import cycle with zerolog here
    fmt.Fprintln(os.Stderr, msg)
}

func clamp(v, lo, hi float64) float64 {
    if v < lo {
        return lo
    }
    if v > hi {
        return hi
    }
    return v
}

// Trace schema types (Placement)
const TraceSchemaVersion = "simple-trace-v1"

type InputsSnapshot struct {
    SrcW, SrcH                                   float64
    PaperWIn, PaperHIn                           float64
    Orientation                                  string
    MarginTopIn, MarginRightIn, MarginBottomIn, MarginLeftIn float64
    DPI                                          float64
    IsSpread                                     bool
    GutterIn                                     float64
    CropRatio                                    *CropRatio
    CropToFill                                   bool
    UserScale                                    float64
    ImagePosition                                Point
    PositionUnits                                string
}

type PageSetupTrace struct {
    PageWpx, PageHpx, SpreadWpx float64
    MarginTopPx, MarginRightPx, MarginBottomPx, MarginLeftPx float64
    Content                     Rect
    GutterPx                    float64
    EffectiveSpreadW            float64
    TargetW, TargetH, TargetRatio float64
}

type CropTrace struct {
    SrcW, SrcH                      float64
    SourceRatio, RequestedRatio     float64
    Decision                        string // crop_width | crop_height | no_crop
    SX, SY, SW, SH                  float64
    PositionUnits                   string
    InputPosX, InputPosY            float64
    RangeX, RangeY                  float64
    NormalizedT                     float64
    AppliedOffsetX, AppliedOffsetY  float64
}

type ScalePlacementTrace struct {
    ScaleX, ScaleY      float64
    Mode                string // cover | fit
    Coverage            float64
    UserScale           float64
    FinalScale          float64
    DW, DH              float64
    CX, CY              float64
    PositionUnits       string
    Tx, Ty              float64
    FreeSpaceX, FreeSpaceY float64
    DX, DY              float64
}

type PanelsTrace struct {
    ContentW, ContentH float64
    PageW              float64
    GutterPx           float64
    HalfGutter         float64
    LeftPanel          Rect
    RightPanel         Rect
    PageLeftOriginX    float64
    PageRightOriginX   float64
    DstLeft            Rect
    DstRight           Rect
}

type ExportTrace struct {
    IsSpread     bool
    SingleCanvas *Size
    LeftCanvas   *Size
    RightCanvas  *Size
}

type PlacementTrace struct {
    Version string
    Inputs  InputsSnapshot
    Step1   PageSetupTrace
    Step2   CropTrace
    Step3   ScalePlacementTrace
    Step4   *PanelsTrace
    Step5   ExportTrace
}

func ComputePlacement(inp Inputs, tr *Trace) Result {
    if tr == nil {
        tr = &Trace{}
    }
    // Initialize structured trace
    st := &PlacementTrace{Version: TraceSchemaVersion}
    st.Inputs = InputsSnapshot{
        SrcW: inp.SrcW, SrcH: inp.SrcH,
        PaperWIn: inp.PaperWIn, PaperHIn: inp.PaperHIn,
        Orientation: inp.Orientation,
        MarginTopIn: inp.MarginTopIn, MarginRightIn: inp.MarginRightIn, MarginBottomIn: inp.MarginBottomIn, MarginLeftIn: inp.MarginLeftIn,
        DPI: inp.DPI,
        IsSpread: inp.IsSpread,
        GutterIn: inp.GutterIn,
        CropRatio: inp.CropRatio,
        CropToFill: inp.CropToFill,
        UserScale: inp.UserScale,
        ImagePosition: inp.ImagePosition,
        PositionUnits: inp.PositionUnits,
    }
    // 1) Paper → px (apply orientation), content rect in L
    var pageWpx, pageHpx float64
    if inp.Orientation == "landscape" {
        pageWpx = inp.PaperHIn * inp.DPI
        pageHpx = inp.PaperWIn * inp.DPI
    } else {
        pageWpx = inp.PaperWIn * inp.DPI
        pageHpx = inp.PaperHIn * inp.DPI
    }
    spreadWpx := pageWpx
    if inp.IsSpread {
        spreadWpx = 2 * pageWpx
    }

    Mt := inp.MarginTopIn * inp.DPI
    Mr := inp.MarginRightIn * inp.DPI
    Mb := inp.MarginBottomIn * inp.DPI
    Ml := inp.MarginLeftIn * inp.DPI

    contentW := spreadWpx - (Ml + Mr)
    contentH := pageHpx - (Mt + Mb)
    contentRect := Rect{X: 0, Y: 0, W: contentW, H: contentH}

    gutterPx := 0.0
    if inp.IsSpread {
        gutterPx = math.Max(0, inp.GutterIn*inp.DPI)
    }
    effectiveSpreadW := contentW
    if inp.IsSpread {
        effectiveSpreadW = math.Max(0, contentW-gutterPx)
    }
    targetW := effectiveSpreadW
    targetH := contentH
    targetRatio := 0.0
    if targetW > 0 && targetH > 0 {
        targetRatio = targetW / targetH
    }
    tr.Logf("[1] page_px=(%.2f,%.2f) spreadW_px=%.2f content=(%.2f,%.2f) gutter=%.2f effectiveSpreadW=%.2f", pageWpx, pageHpx, spreadWpx, contentW, contentH, gutterPx, effectiveSpreadW)
    st.Step1 = PageSetupTrace{
        PageWpx: pageWpx, PageHpx: pageHpx, SpreadWpx: spreadWpx,
        MarginTopPx: Mt, MarginRightPx: Mr, MarginBottomPx: Mb, MarginLeftPx: Ml,
        Content: contentRect,
        GutterPx: gutterPx,
        EffectiveSpreadW: effectiveSpreadW,
        TargetW: targetW, TargetH: targetH, TargetRatio: targetRatio,
    }

    // 2) Source crop window in S
    W := inp.SrcW
    H := inp.SrcH
    sourceRatio := safeDiv(W, H)
    reqRatio := sourceRatio
    if inp.CropRatio != nil {
        reqRatio = safeDiv(inp.CropRatio.W, inp.CropRatio.H)
    } else if inp.CropToFill {
        reqRatio = targetRatio
    }

    sx, sy, sw, sh := 0.0, 0.0, W, H
    rangeX := 0.0
    rangeY := 0.0
    normT := 0.0
    appliedOffX := 0.0
    appliedOffY := 0.0
    decision := "no_crop"
    if W > 0 && H > 0 && reqRatio > 0 {
        if sourceRatio > reqRatio { // crop width
            decision = "crop_width"
            sh = H
            sw = H * reqRatio
            rangeX = math.Max(0, W-sw)
            if inp.PositionUnits == "normalized" {
                normT = clamp((inp.ImagePosition.X+1)/2, 0, 1)
                sx = normT * rangeX
                appliedOffX = sx
            } else { // px
                sx = clamp(rangeX/2+inp.ImagePosition.X, 0, rangeX)
                appliedOffX = sx - rangeX/2
            }
            sy = 0
        } else if sourceRatio < reqRatio { // crop height
            decision = "crop_height"
            sw = W
            sh = safeDiv(W, reqRatio)
            rangeY = math.Max(0, H-sh)
            if inp.PositionUnits == "normalized" {
                normT = clamp((inp.ImagePosition.Y+1)/2, 0, 1)
                sy = normT * rangeY
                appliedOffY = sy
            } else {
                sy = clamp(rangeY/2+inp.ImagePosition.Y, 0, rangeY)
                appliedOffY = sy - rangeY/2
            }
            sx = 0
        } else {
            sx, sy, sw, sh = 0, 0, W, H
        }
    }
    tr.Logf("[2] crop S: sx=%.2f sy=%.2f sw=%.2f sh=%.2f reqRatio=%.6f srcRatio=%.6f", sx, sy, sw, sh, reqRatio, sourceRatio)
    st.Step2 = CropTrace{
        SrcW: W, SrcH: H, SourceRatio: sourceRatio, RequestedRatio: reqRatio, Decision: decision,
        SX: sx, SY: sy, SW: sw, SH: sh,
        PositionUnits: inp.PositionUnits,
        InputPosX: inp.ImagePosition.X, InputPosY: inp.ImagePosition.Y,
        RangeX: rangeX, RangeY: rangeY,
        NormalizedT: normT,
        AppliedOffsetX: appliedOffX, AppliedOffsetY: appliedOffY,
    }

    // 3) Scale and position in L
    scaleX := 0.0
    if sw > 0 {
        scaleX = targetW / sw
    }
    scaleY := 0.0
    if sh > 0 {
        scaleY = targetH / sh
    }
    finalScale := 0.0
    tx := 0.0
    ty := 0.0
    if inp.CropToFill {
        coverage := math.Max(scaleX, scaleY)
        finalScale = coverage * math.Max(1, inp.UserScale)
        tx, ty = 0, 0
    } else {
        finalScale = math.Min(scaleX, scaleY) * inp.UserScale
        dwTmp := sw * finalScale
        dhTmp := sh * finalScale
        if inp.PositionUnits == "normalized" {
            tx = inp.ImagePosition.X * (targetW - dwTmp) / 2
            ty = inp.ImagePosition.Y * (targetH - dhTmp) / 2
        } else {
            tx = inp.ImagePosition.X
            ty = inp.ImagePosition.Y
        }
    }

    dw := sw * finalScale
    dh := sh * finalScale
    cx := targetW / 2
    cy := targetH / 2
    dx := cx - dw/2 + tx
    dy := cy - dh/2 + ty
    tr.Logf("[3] scale: sx=%.6f sy=%.6f final=%.6f dw=%.2f dh=%.2f dx=%.2f dy=%.2f", scaleX, scaleY, finalScale, dw, dh, dx, dy)
    mode := "fit"
    coverage := math.Min(scaleX, scaleY)
    if inp.CropToFill {
        mode = "cover"
        coverage = math.Max(scaleX, scaleY)
    }
    st.Step3 = ScalePlacementTrace{
        ScaleX: scaleX, ScaleY: scaleY,
        Mode: mode,
        Coverage: coverage,
        UserScale: inp.UserScale,
        FinalScale: finalScale,
        DW: dw, DH: dh,
        CX: cx, CY: cy,
        PositionUnits: inp.PositionUnits,
        Tx: tx, Ty: ty,
        FreeSpaceX: targetW - dw, FreeSpaceY: targetH - dh,
        DX: dx, DY: dy,
    }

    srcRectGlobal := Rect{X: sx, Y: sy, W: sw, H: sh}
    dstRectGlobal := Rect{X: dx, Y: dy, W: dw, H: dh}

    // 4) Panels, pages & clipping (spread split)
    var leftPanel, rightPanel, dstLeft, dstRight *Rect
    var pageW float64
    if inp.IsSpread {
        // Page canvas width is half of content width
        pageW = contentW / 2
        m := gutterPx / 2
        // Visible areas (clip) in global L coordinates
        l := Rect{X: 0, Y: 0, W: pageW - m, H: contentH}
        r := Rect{X: pageW + m, Y: 0, W: pageW - m, H: contentH}
        leftPanel, rightPanel = &l, &r

        // Destination rects computed relative to page canvas origins
        pageLeftX := 0.0
        pageRightX := pageW
        dl := Rect{X: dx - pageLeftX, Y: dy, W: dw, H: dh}
        dr := Rect{X: dx - pageRightX, Y: dy, W: dw, H: dh}
        dstLeft, dstRight = &dl, &dr
        tr.Logf("[4] panels: left=(%.0fx%.0f) right=(%.0fx%.0f) gutter=%.1f pageW=%.0f", l.W, l.H, r.W, r.H, gutterPx, pageW)
        st.Step4 = &PanelsTrace{
            ContentW: contentW, ContentH: contentH,
            PageW: pageW,
            GutterPx: gutterPx,
            HalfGutter: m,
            LeftPanel: l, RightPanel: r,
            PageLeftOriginX: pageLeftX, PageRightOriginX: pageRightX,
            DstLeft: dl, DstRight: dr,
        }
    }

    // 5) Export sizes
    var exportSingle *Size
    var exportSpread *SpreadSizes
    if !inp.IsSpread {
        exportSingle = &Size{W: int(math.Round(contentW)), H: int(math.Round(contentH))}
        tr.Logf("[5] export single: %dx%d", exportSingle.W, exportSingle.H)
    } else {
        // Export each page at page canvas width (includes inner gutter margins)
        lw := int(math.Round(contentW / 2))
        rw := lw
        left := Size{W: lw, H: int(math.Round(contentH))}
        right := Size{W: rw, H: int(math.Round(contentH))}
        exportSpread = &SpreadSizes{Left: left, Right: right}
        tr.Logf("[5] export spread: left=%dx%d right=%dx%d (pageW=%.0f, gutter=%.1f)", left.W, left.H, right.W, right.H, contentW/2, gutterPx)
    }

    st.Step5 = ExportTrace{IsSpread: inp.IsSpread}
    if exportSingle != nil {
        st.Step5.SingleCanvas = &Size{W: exportSingle.W, H: exportSingle.H}
    }
    if exportSpread != nil {
        st.Step5.LeftCanvas = &Size{W: exportSpread.Left.W, H: exportSpread.Left.H}
        st.Step5.RightCanvas = &Size{W: exportSpread.Right.W, H: exportSpread.Right.H}
    }

    res := Result{
        SrcRectGlobal:    srcRectGlobal,
        DstRectGlobal:    dstRectGlobal,
        ContentRect:      contentRect,
        EffectiveSpreadW: effectiveSpreadW,
        GutterPx:         gutterPx,
        LeftPanel:        leftPanel,
        RightPanel:       rightPanel,
        DstRectLeft:      dstLeft,
        DstRectRight:     dstRight,
        ExportSingle:     exportSingle,
        ExportSpread:     exportSpread,
    }
    tr.Structured = st
    return res
}

func safeDiv(a, b float64) float64 {
    if b == 0 {
        return 0
    }
    return a / b
}


