### How the simple spread computation & rendering algorithm works

This guide explains the end‑to‑end behavior of the simple spread algorithm: how it computes placement (crop, scale, position, split into panels) and how it renders outputs (single pages or left/right panels for spreads).

The implementation lives in `zine-layout/pkg/spread/simple/` and is composed of four files:
- `algorithm.go`: core math and placement computation
- `render.go`: raster rendering, scaling, clipping, encoding
- `inputs.go`: translation from public settings to algorithm inputs
- `html.go`: optional HTML gallery index with trace logs

---

## 1) Concepts and coordinate spaces

- **Source space (S)**: pixel space of the input image. Units are pixels of the original source. Crop windows are defined in S as `Rect{X,Y,W,H}` with floating point values for subpixel accuracy.
- **Layout space (L)**: pixel space of the destination canvas inside the page content area (margins removed). All computed placement coordinates (destination rectangles, panels) live in L.
- **Page vs Spread**:
  - Page size is based on physical paper dimensions and `DPI` (dots per inch) plus margins.
  - Spread uses two facing pages. The content area is split into left and right panels with an optional inner `gutter` that removes a band in the middle from the visible area.

Terminology used in code:
- `SrcRectGlobal` (in S): what region of the source image is used.
- `DstRectGlobal` (in L): where the (scaled) image lands relative to the content area origin.
- `LeftPanel`/`RightPanel` (in L): visible rectangles for each page in a spread; they account for the gutter.
- `DstRectLeft`/`DstRectRight` (in each page’s own local origin): destination rects re‑expressed per‑page to simplify page‑local drawing + clipping.

---

## 2) Inputs model

The algorithm uses the following inputs (see `Inputs`):
- **Source image**: `SrcW`, `SrcH` in pixels.
- **Paper & layout** (inches): `PaperWIn`, `PaperHIn`, `Orientation` (`portrait|landscape`), `Margin*In`, and `DPI`.
- **Spread & gutter**: `IsSpread` (bool), `GutterIn` (inches); only meaningful when `IsSpread` is true.
- **Crop & scale**:
  - `CropRatio`: optional fixed aspect ratio of the crop window (W:H). If nil, the ratio is either the source ratio or the target ratio (when `CropToFill` is true).
  - `CropToFill`: when true, the crop ratio defaults to the target layout aspect ratio, and final scaling uses cover behavior (no letterboxing).
  - `UserScale`: multiplicative factor ≥ 0 to further scale the placed image after the algorithm’s baseline fit/cover scale.
- **Position**:
  - `ImagePosition`: user translation or crop alignment hint. Interpreted differently depending on the step.
  - `PositionUnits`: `normalized` or `px`.
    - For crop alignment in S: `normalized` treats values in [-1, 1] → [left/top, right/bottom] inside the allowable crop range; `px` offsets from center are clamped to the valid range.
    - For placement translation in L (non‑fill mode): `normalized` maps to half‑range of free space, `px` is direct pixel translation.

`inputs.go` provides helpers to build `Inputs` from the public `spread.Settings` while applying defaults for unset fields.

---

## 3) Placement computation (algorithm.go)

The core of the algorithm is `ComputePlacement(inputs, trace) → Result`. It executes 5 ordered steps and logs a trace that can be shown in the HTML index.

### Step 1 – Page → px, content rect, gutter, target box

1. Convert paper inches to page pixels using `DPI`, applying `Orientation`.
2. Compute spread width: one page width for single pages, double for spreads.
3. Subtract margins to get the content area `contentRect` in L.
4. If `IsSpread`, compute `gutterPx = max(0, GutterIn*DPI)` and set `effectiveSpreadW = max(0, contentW - gutterPx)`; otherwise `effectiveSpreadW = contentW`.
5. Define the target box for image placement as `targetW = effectiveSpreadW`, `targetH = contentH`, and `targetRatio = targetW/targetH`.

Effect: the gutter is removed from the target width used for fitting/covering. This ensures the composed image covers the visible left+right panels without being scaled down by the invisible inner gap.

Emitted structured trace (Step 1):
```json
{
  "Step1": {
    "PageWpx": 2550,
    "PageHpx": 3300,
    "SpreadWpx": 2550,
    "MarginTopPx": 150,
    "MarginRightPx": 150,
    "MarginBottomPx": 150,
    "MarginLeftPx": 150,
    "Content": { "X": 0, "Y": 0, "W": 2250, "H": 3000 },
    "GutterPx": 0,
    "EffectiveSpreadW": 2250,
    "TargetW": 2250,
    "TargetH": 3000,
    "TargetRatio": 0.75
  }
}
```

### Step 2 – Choose crop window in source space S

1. Compute `sourceRatio = SrcW/SrcH`.
2. Choose required crop ratio `reqRatio`:
   - If `CropRatio != nil`, use it.
   - Else if `CropToFill`, use `targetRatio`.
   - Else use `sourceRatio` (no aspect change needed).
3. Based on comparison of `sourceRatio` and `reqRatio`:
   - If source is wider, crop width: `sw = H*reqRatio`, `sh = H`.
   - If source is taller, crop height: `sw = W`, `sh = W/reqRatio`.
   - Else keep full image.
4. Position the crop window using `ImagePosition`:
   - For width‑crop: slide horizontally in `[0, W-sw]`.
   - For height‑crop: slide vertically in `[0, H-sh]`.
   - With `PositionUnits = normalized`, `[-1, 1]` maps to `[0, range]`. With `px`, offsets are interpreted around center and clamped.

Output: `SrcRectGlobal = (sx, sy, sw, sh)`.

Emitted structured trace (Step 2):
```json
{
  "Step2": {
    "SrcW": 3000,
    "SrcH": 2000,
    "SourceRatio": 1.5,
    "RequestedRatio": 1.5,
    "Decision": "no_crop",
    "SX": 0,
    "SY": 0,
    "SW": 3000,
    "SH": 2000,
    "PositionUnits": "normalized",
    "InputPosX": 0,
    "InputPosY": 0,
    "RangeX": 0,
    "RangeY": 0,
    "NormalizedT": 0,
    "AppliedOffsetX": 0,
    "AppliedOffsetY": 0
  }
}
```

### Step 3 – Scale and place in layout space L

Let `scaleX = targetW/sw`, `scaleY = targetH/sh`.

- If `CropToFill` (cover):
  - `coverage = max(scaleX, scaleY)` ensures fill in both dimensions.
  - `finalScale = coverage * max(1, UserScale)`; user scale cannot reduce below cover.
  - Translation is ignored (`tx = ty = 0`) because cover already aligns to the target box.
- Else (fit/contain):
  - `finalScale = min(scaleX, scaleY) * UserScale`.
  - Compute output dims: `dw = sw*finalScale`, `dh = sh*finalScale`.
  - Center inside target box: `(cx, cy) = (targetW/2, targetH/2)`, then `(dx, dy) = (cx - dw/2, cy - dh/2)`.
  - Apply user translation:
    - `normalized`: `tx = posX * (targetW - dw)/2`, `ty = posY * (targetH - dh)/2`.
    - `px`: `tx = posX`, `ty = posY`.
  - Final destination rect: `DstRectGlobal = (dx+tx, dy+ty, dw, dh)`.

Emitted structured trace (Step 3):
```json
{
  "Step3": {
    "Mode": "fit",
    "ScaleX": 0.75,
    "ScaleY": 1.5,
    "Coverage": 0.75,
    "UserScale": 1,
    "FinalScale": 0.75,
    "DW": 2250,
    "DH": 1500,
    "CX": 1125,
    "CY": 1500,
    "FreeSpaceX": 0,
    "FreeSpaceY": 1500,
    "PositionUnits": "normalized",
    "Tx": 0,
    "Ty": 0,
    "DX": 0,
    "DY": 750
  }
}
```

### Step 4 – Panels & page‑local rects (spreads only)

For spreads, the content area is split into two pages of width `pageW = contentW/2`. The gutter reduces the visible width of each facing page by half the gutter:

- Let `m = gutterPx/2`.
- `LeftPanel = Rect{X: 0, Y: 0, W: pageW - m, H: contentH}`.
- `RightPanel = Rect{X: pageW + m, Y: 0, W: pageW - m, H: contentH}`.

For drawing convenience, compute destination rects in page‑local coordinates (origins `(0,0)` at each page’s left/top):

- `DstRectLeft = (dx - 0, dy, dw, dh)`.
- `DstRectRight = (dx - pageW, dy, dw, dh)`.

These rects will be drawn with explicit clipping so that content falling inside the inner gutter band is not rendered.

Emitted structured trace (Step 4 – when IsSpread=true):
```json
{
  "Step4": {
    "ContentW": 4500,
    "ContentH": 3000,
    "PageW": 2250,
    "GutterPx": 150,
    "HalfGutter": 75,
    "LeftPanel": { "X": 0, "Y": 0, "W": 2175, "H": 3000 },
    "RightPanel": { "X": 2325, "Y": 0, "W": 2175, "H": 3000 },
    "PageLeftOriginX": 0,
    "PageRightOriginX": 2250,
    "DstLeft": { "X": -100, "Y": 700, "W": 2400, "H": 1600 },
    "DstRight": { "X": -2350, "Y": 700, "W": 2400, "H": 1600 }
  }
}
```

### Step 5 – Export sizes

- Single page: export canvas is `contentW x contentH`.
- Spread: export two images of size `(contentW/2) x contentH` each. The inner gutter is not part of the visible width of each page; it is handled by clipping during rendering, not by shrinking the export canvas.

The function returns a `Result` bundling all of the above plus the computed export sizes.

Emitted structured trace (Step 5):
```json
{
  "Step5": {
    "IsSpread": false,
    "SingleCanvas": { "W": 2250, "H": 3000 },
    "LeftCanvas": null,
    "RightCanvas": null
  }
}
```

---

## 4) Rendering (render.go)

Two top‑level functions render to images using the precomputed `Result`:
- `RenderSingle(ctx, src, res, info) → filePath`
- `RenderSpread(ctx, src, res, info) → (leftPath, rightPath)`

Key behavior:
- A new `NRGBA` canvas is allocated with the export size.
- Optional background fill is applied via `Background` (supports `transparent`, `#RRGGBB[AA]`, or named `white` fallback).
- The source crop (`SrcRectGlobal` in S) is scaled into the destination rectangle using either quality or fast scaler:
  - `Scaler = quality` → Catmull‑Rom
  - `Scaler = fast` → ApproxBiLinear
- For spreads, each page is rendered independently with clipping to exclude the inner gutter band:
  - Compute per‑page clip rectangles from `LeftPanel`/`RightPanel` widths.
  - Use `scaleAndDrawWithClip` to intersect the destination with the clip, map back into source S, and scale just the visible subarea.

Encoding & filenames:
- `Format`: `png` or `jpg` (`jpeg` mapped to `jpg`).
- `Quality` for JPEG; `PNGLevel` (`default|speed|max|none`).
- Optional `ParallelEncode` writes left/right concurrently.
- Filenames are produced by a template with tokens:
  - `{index}`, `{index:03d}`, `{name}`, `{panel}` (`single|left|right`), `{image_basename}`, `{ext}`
- `OutputDir` controls the directory; `PathOverrides` can inject absolute output paths per panel.

Emitted structured trace (RenderSingle):
```json
{
  "RenderSingle": {
    "Canvas": { "W": 2250, "H": 3000 },
    "ScaleRects": {
      "SrcRect": { "min": {"x": 0, "y": 0}, "max": {"x": 3000, "y": 2000} },
      "DstRect": { "min": {"x": 0, "y": 750}, "max": {"x": 2250, "y": 2250} }
    },
    "Scaler": "CatmullRom",
    "Background": "white",
    "OutputPath": "./out/001-spread-single.png"
  }
}
```

Emitted structured trace (RenderSpread):
```json
{
  "RenderSpread": {
    "LeftCanvas": { "W": 2250, "H": 3000 },
    "RightCanvas": { "W": 2250, "H": 3000 },
    "LeftClip": {
      "Clip": { "min": {"x": 0, "y": 0}, "max": {"x": 2175, "y": 3000} },
      "OrigDst": { "min": {"x": -100, "y": 700}, "max": {"x": 2300, "y": 2300} },
      "IntersectDst": { "min": {"x": 0, "y": 700}, "max": {"x": 2175, "y": 2300} },
      "OffX": 100,
      "OffY": 0,
      "SrcRect": { "min": {"x": 133, "y": 87}, "max": {"x": 2933, "y": 1713} }
    },
    "RightClip": {
      "Clip": { "min": {"x": 75, "y": 0}, "max": {"x": 2250, "y": 3000} },
      "OrigDst": { "min": {"x": -2350, "y": 700}, "max": {"x": 50, "y": 2300} },
      "IntersectDst": { "min": {"x": 75, "y": 700}, "max": {"x": 50, "y": 2300} },
      "OffX": 2425,
      "OffY": 0,
      "SrcRect": { "min": {"x": 1450, "y": 87}, "max": {"x": 3025, "y": 1713} }
    },
    "Scaler": "CatmullRom",
    "EncodedInParallel": true,
    "LeftOutputPath": "./out/001-spread-left.png",
    "RightOutputPath": "./out/001-spread-right.png"
  }
}
```

---

## 5) Settings → Inputs and RenderInfo (inputs.go)

`InputsFromSettings(settings, meta)` applies defaults and validates the image dimensions. Notable rules:
- If `CropRatio` is provided as a single `float64`, it is interpreted as W:1.
- `Units` default to `normalized` if empty.
- All margin and paper values are passed straight through; `Orientation` must be `portrait` or `landscape`.

`RenderInfoFromExport(export, index, spreadName, imageBase, opts)` prepares the rendering parameters with safe defaults:
- Default `Format=png`, `Background=white`, `FilenameTemplate={index:03d}-{name}-{panel}.{ext}`.
- Default `OutputDir=./out`, `Quality=90`, `PNGLevel=default`, `Scaler=quality`.

---

## 6) Trace logging & HTML index (html.go)

`Trace` collects human‑readable lines that explain each step’s key numbers. When assembled into a `SpreadOutput`, these logs can be embedded into an HTML gallery via `WriteHTMLIndex(path, results)` to aid troubleshooting.

The annotations include page/content sizes, gutter handling, crop window, scale factors, destination rects, and export dimensions.

---

## 7) Numerical behavior and edge cases

- All divisions are guarded (`safeDiv`) to avoid NaNs.
- Gutter and effective spread widths are clamped to ≥ 0.
- Crop/position inputs are clamped to valid ranges. In `normalized` mode `[-1,1]` maps to the crop/placement range; in `px` mode, crop offsets are interpreted around the center and clamped to `[0, range]`.
- Rounding: destination and source rectangles are rounded at draw time with `+0.5` to minimize drift; export sizes are rounded with `Round`.
- In `CropToFill` mode, `UserScale < 1` is ignored by design (cannot shrink below the cover scale). In fit mode, `UserScale` can enlarge beyond fit.
- If the visible area gets fully clipped by gutter (extreme cases), nothing is drawn for that part.

---

## 8) Worked mini‑example (single page)

Given `Src=3000x2000`, paper `8.5"×11"` portrait, `DPI=300`, margins `0.5"` all around, `IsSpread=false`, `CropToFill=false`, `UserScale=1`, centered position:
- Page px: `(2550×3300)`; content `(2250×3000)`.
- Target box: `(W=2250,H=3000)`.
- Source ratio `1.5`, target ratio `0.75`; fit uses `finalScale = min(2250/3000, 3000/2000) = 0.75` → placed `dw×dh = 2250×1500`, centered with `dy = (3000-1500)/2 = 750`.

The resulting image letterboxes vertically inside the content area.

---

## 9) Typical integration (pseudocode)

```go
inputs := InputsFromSettings(settings, meta)
res := ComputePlacement(inputs, &Trace{EnableStdout: false})
if settings.IsSpread {
  left, right, err := RenderSpread(ctx, srcImage, res, RenderInfoFromExport(export, idx, spreadName, imageBase, opts))
  _ = left; _ = right; _ = err
} else {
  out, err := RenderSingle(ctx, srcImage, res, RenderInfoFromExport(export, idx, spreadName, imageBase, opts))
  _ = out; _ = err
}
```

---

## 10) FAQ and practical notes

- To push the image away from the gutter on spreads (fit mode), use `normalized` positions with small translations. In cover mode, translations are ignored because the image is expanded to fill.
- To tightly control aspect, specify `CropRatio`; it overrides `CropToFill`’s implicit ratio.
- Choose `Scaler=quality` for final renders and `Scaler=fast` for quick previews.
- Use `Background=transparent` with `png` for transparent page backgrounds.

---

## 11) What the algorithm guarantees

- With `CropToFill`, the visible left+right panel region is fully covered (no letterboxing) after removing the gutter width from the target box.
- With fit mode, the entire cropped source is visible inside the target box and centered unless you translate it.
- Spread exports are page‑sized; the inner gutter band is empty because of explicit clipping.

---

## 12) File map

- `pkg/spread/simple/algorithm.go`: `Inputs`, `Result`, `Trace`, `ComputePlacement`
- `pkg/spread/simple/render.go`: `RenderSingle`, `RenderSpread`, scaling + clipping + encoding
- `pkg/spread/simple/inputs.go`: `InputsFromSettings`, `RenderInfoFromExport`
- `pkg/spread/simple/html.go`: `WriteHTMLIndex`

This document should serve as a stable reference when adjusting crop rules, gutter semantics, or rendering behavior.

---

## 13) Structured trace schema (design)

The current text trace is useful for humans but hard to consume programmatically. This section designs a structured, stable trace schema that records all inputs, intermediate values, decisions, and outputs at each step, enabling faithful reproduction and debugging.

### 13.1 Goals

- **Complete**: capture every numeric input and intermediate derived value used in decisions.
- **Deterministic**: values reflect pre‑rounding floats and the exact rounding used at draw time.
- **Versioned**: schema includes a `version` to support evolution.
- **Composable**: placement and rendering traces are separate but linkable.

### 13.2 Core types (to be implemented in Go later)

```go
// Version tag for forwards-compat
const TraceSchemaVersion = "simple-trace-v1"

type RectF struct { X, Y, W, H float64 }
type SizeF struct { W, H float64 }
type SizeI struct { W, H int }

type Units string       // "normalized" | "px"
type Orientation string // "portrait" | "landscape"
type CropDecision string // "crop_width" | "crop_height" | "no_crop"
type ScaleMode string    // "cover" | "fit"

type InputsSnapshot struct {
  SrcW, SrcH float64
  PaperWIn, PaperHIn float64
  Orientation Orientation
  MarginTopIn, MarginRightIn, MarginBottomIn, MarginLeftIn float64
  DPI float64
  IsSpread bool
  GutterIn float64
  CropRatio *struct{ W, H float64 } // nil means not set
  CropToFill bool
  UserScale float64
  ImagePosition struct{ X, Y float64 }
  PositionUnits Units
}

type PlacementTrace struct {
  Version string
  Inputs InputsSnapshot
  Step1 PageSetupTrace
  Step2 CropTrace
  Step3 ScalePlacementTrace
  Step4 PanelsTrace // nil when !IsSpread
  Step5 ExportTrace
}
```

### 13.3 Placement steps (ComputePlacement)

#### Step 1 – Page setup and target box

```go
type PageSetupTrace struct {
  // Derived page sizes (px)
  PageWpx float64
  PageHpx float64
  SpreadWpx float64

  // Margins in px and resulting content rect in L
  MarginTopPx, MarginRightPx, MarginBottomPx, MarginLeftPx float64
  Content RectF

  // Gutter handling
  GutterPx float64
  EffectiveSpreadW float64 // content.W minus gutter (for spreads)

  // Target box used for fit/cover
  TargetW float64
  TargetH float64
  TargetRatio float64 // 0 when invalid
}
```

Captured decisions: orientation mapping (paperW/H swap), `IsSpread` fan‑out, gutter clamp to ≥ 0.

#### Step 2 – Source crop window (S)

```go
type CropTrace struct {
  SrcW float64
  SrcH float64
  SourceRatio float64
  RequestedRatio float64 // from CropRatio or TargetRatio or SourceRatio
  Decision CropDecision

  // Computed crop window
  SX, SY, SW, SH float64

  // Positioning internals
  PositionUnits Units
  InputPosX float64 // as provided
  InputPosY float64
  RangeX float64 // max move in X when cropping width
  RangeY float64 // max move in Y when cropping height
  NormalizedT float64 // resolved t in [0,1] when Units==normalized; 0 otherwise
  AppliedOffsetX float64 // final sx offset chosen (delta from centered suggestion when Units==px)
  AppliedOffsetY float64 // final sy offset chosen
}
```

Captured decisions: width vs height vs none; resolved `RequestedRatio`; clamping behavior for both normalized and px positioning.

#### Step 3 – Scale and placement (L)

```go
type ScalePlacementTrace struct {
  ScaleX float64 // TargetW / SW (0 if SW==0)
  ScaleY float64 // TargetH / SH (0 if SH==0)
  Mode ScaleMode // cover when CropToFill else fit

  Coverage float64 // max(ScaleX, ScaleY) for cover; min(...) for fit (pre UserScale)
  UserScale float64
  FinalScale float64

  // Output dims and center
  DW float64
  DH float64
  CX float64
  CY float64

  // Translation components
  PositionUnits Units
  Tx float64
  Ty float64
  FreeSpaceX float64 // TargetW - DW (fit only)
  FreeSpaceY float64 // TargetH - DH (fit only)

  // Final destination rect in L
  DX float64
  DY float64
}
```

Captured decisions: cover vs fit; enforcement of `max(1, UserScale)` in cover; normalized translation mapping to half free‑space.

#### Step 4 – Panels, pages, and clipping (spreads)

```go
type PanelsTrace struct {
  ContentW float64
  ContentH float64
  PageW float64 // contentW/2
  GutterPx float64
  HalfGutter float64 // m = GutterPx/2

  LeftPanel RectF  // in L
  RightPanel RectF // in L

  // Page-local destination rects
  PageLeftOriginX float64 // 0
  PageRightOriginX float64 // PageW
  DstLeft RectF
  DstRight RectF
}
```

Captured decisions: visible widths `(pageW - m)` and origins; conversion of global destination rect into page‑local coordinates.

#### Step 5 – Export sizes

```go
type ExportTrace struct {
  IsSpread bool

  // When single
  SingleCanvas *SizeI

  // When spread
  LeftCanvas  *SizeI
  RightCanvas *SizeI
}
```

Captured decisions: rounding to nearest int for export sizes.

### 13.4 Rendering trace (RenderSingle/RenderSpread)

Rendering produces its own trace because rounding and clipping are applied at draw time.

```go
type RenderInfoSnapshot struct {
  Format string // png|jpg
  Quality int
  Background string // e.g. #RRGGBBAA or "transparent"
  PNGLevel string // default|speed|max|none
  Scaler string   // fast|quality
  ParallelEncode bool
  FilenameTemplate string
  OutputDir string
}

type ScaleRectsInt struct {
  SrcRect image.Rectangle // rounded S
  DstRect image.Rectangle // rounded L canvas space
}

type RenderSingleTrace struct {
  Version string
  Info RenderInfoSnapshot
  Canvas SizeI
  ScaleRects ScaleRectsInt
  Scaler string // resolved (ApproxBiLinear|CatmullRom)
  OutputPath string
}

type ClipMapping struct {
  Clip image.Rectangle // page clip in canvas space
  OrigDst image.Rectangle // pre-clip dst rect
  IntersectDst image.Rectangle // dst ∩ clip
  OffX int // IntersectDst.Min.X - OrigDst.Min.X
  OffY int
  SrcRect image.Rectangle // corresponding mapped source rect used for scaling
}

type RenderSpreadTrace struct {
  Version string
  Info RenderInfoSnapshot
  LeftCanvas SizeI
  RightCanvas SizeI
  LeftClip ClipMapping
  RightClip ClipMapping
  Scaler string
  LeftOutputPath string
  RightOutputPath string
  EncodedInParallel bool
}
```

Captured decisions: background color resolution, scaler selection, exact integer rectangles used, clip → source mapping math (`scaleX/scaleY`, offsets), and output paths.

### 13.5 Example payloads

Example (Step 2 – crop):

```json
{
  "Step2": {
    "SrcW": 6000,
    "SrcH": 4000,
    "SourceRatio": 1.5,
    "RequestedRatio": 0.75,
    "Decision": "crop_height",
    "SX": 0,
    "SY": 500,
    "SW": 6000,
    "SH": 8000,
    "PositionUnits": "normalized",
    "InputPosX": 0,
    "InputPosY": -0.5,
    "RangeX": 0,
    "RangeY": 4000,
    "NormalizedT": 0.25,
    "AppliedOffsetX": 0,
    "AppliedOffsetY": 1000
  }
}
```

Example (Render spread – left clip mapping):

```json
{
  "RenderSpread": {
    "LeftCanvas": { "W": 2250, "H": 3000 },
    "LeftClip": {
      "Clip": { "min": {"x": 0, "y": 0}, "max": {"x": 2100, "y": 3000} },
      "OrigDst": { "min": {"x": -200, "y": 100}, "max": {"x": 2300, "y": 2900} },
      "IntersectDst": { "min": {"x": 0, "y": 100}, "max": {"x": 2100, "y": 2900} },
      "OffX": 200,
      "OffY": 0,
      "SrcRect": { "min": {"x": 120, "y": 80}, "max": {"x": 2520, "y": 1920} }
    }
  }
}
```

Notes:
- Rectangle JSON shape is illustrative; the Go impl will likely serialize with `{X,Y,W,H}` for float rects and `{Min:{X,Y},Max:{X,Y}}` for int rects to match `image.Rectangle`.

### 13.6 Emission contracts

- Emit the full `PlacementTrace` once per `ComputePlacement` call.
- Emit a `RenderSingleTrace` or `RenderSpreadTrace` per render call. When spreads are encoded in parallel, capture both panels’ clip mappings.
- Include both pre‑rounded floats and the post‑rounded integer rectangles used by the scaler.
- Record clamping by reporting both the raw input (e.g., `InputPosX/Y`) and the resolved/applied offsets (`AppliedOffsetX/Y`).
- Never omit fields; when not applicable, use zero values or explicit `nil` pointers as designed above.


