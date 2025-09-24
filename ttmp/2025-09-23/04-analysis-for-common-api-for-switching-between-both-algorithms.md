## Common API for Spread Algorithms (Simple and Sonnet)

### Purpose & Scope
This document analyzes the two existing spread placement/rendering algorithms — the "Simple" algorithm (`cmd/experiments/spread-algorithm/algorithm.go`) and the "Sonnet" engine (`cmd/experiments/spread-algorithm-actually-from-sonnet/internal/engine`) — and proposes a unified, reusable API under `zine-layout/pkg/...`. It also outlines how to switch between algorithms at runtime and exposes a REST API for computing geometry and generating preview/final images.

Goals:
- Define a common type system and interface so both algorithms can be used interchangeably.
- Extract and relocate algorithm code to `pkg/spread/...` (and subpackages per algorithm).
- Introduce a strategy/registry to pick the active engine (simple vs sonnet).
- Provide a REST API for: (1) compute-only previews (geometry) and (2) rendered images (preview/full).

---

## 1) Algorithm Overviews

### 1.1 Simple Algorithm (algorithm.go)
Location: `zine-layout/cmd/experiments/spread-algorithm/algorithm.go`

Key data types:
- Inputs: source size in px, paper and margins in inches (plus DPI), spread/gutter, crop ratio or fill, user scale, image position (px or normalized), orientation.
- Output Result includes:
  - `ContentRect` (content box inside margins, spread-aware)
  - `EffectiveSpreadW`, `GutterPx`
  - Crop window in source (`SrcRectGlobal`) and destination rect in layout (`DstRectGlobal`)
  - For spreads: `LeftPanel`, `RightPanel` (clip rects), and `DstRectLeft`/`DstRectRight` (dst relative to each page canvas origin)
  - Export sizes (single or per spread page)
  - Optional `Trace` for debugging

Notable behavior:
- Paper size is converted to px after applying orientation.
- Content area excludes margins. For spreads, gutter reduces effective content width.
- Crop selection: uses specified ratio, or if `CropToFill`, matches target aspect ratio for the content area.
- Placement/scaling: cover (fill) or contain (fit) based on `CropToFill` with `UserScale` and position offsets.
- Spread split: compute clip rectangles for left/right panels with gutter center.

Observations:
- Uses a global layout space (L) with origin at content origin (0,0) rather than paper origin.
- Provides both global destination rect and page-relative ones for spreads.
- Export sizes are computed but rendering is not included in this file.

### 1.2 Sonnet Engine (internal/engine)
Location: `zine-layout/cmd/experiments/spread-algorithm-actually-from-sonnet/internal/engine`

Key types (from `compute.go`):
- `Result` includes:
  - `PaperSizePx`, `PaperSizeIn`
  - `MarginsPx`, `GutterPx`
  - `EffectiveRect` (content rect including margins/gutter in px; origin is paper origin)
  - `SourceCrop` (crop in source px)
  - `ImageDisplay` (scaled image size)
  - `VirtualPosition` (top-left where image should be drawn on full paper canvas)
  - `Panels` with `PanelRect` (clip rect per page/panel) and `ImageOffset` (where to draw relative to each panel)
  - `Trace`
  - `Settings` (`ResolvedSettings` – derived from YAML config)

Key types (from `render.go`):
- Rendering draws on a full paper-size canvas with background color.
- Crops source to `SourceCrop`, scales to `ImageDisplay`, draws at `VirtualPosition` onto the canvas.
- For spreads, the final canvas is split into left/right images considering gutter.
- Writes to png/jpg/pdf, supports quality, background, and DPI for PDF.

Observations:
- Sonnet defines both geometry and a full rendering pipeline.
- Coordinates are paper-origin-based; `EffectiveRect` is at margin origin.
- Trace is structured and forwarded to zerolog.

---

## 2) Common Denominators and Differences

### Common Denominators
- Inputs: paper (inches + DPI), margins, spread/gutter, crop ratio vs fill, scale, position, orientation; source image px.
- Computed geometry: content rect (effective area), source crop, destination (scaled) rect, panels (single/left/right) with clip and offsets.
- Need for: preview geometry and export sizes; image rendering to PNG/JPG/PDF.

### Key Differences
- Coordinate origins:
  - Simple: content-origin-based `ContentRect{X:0,Y:0}`; provides `DstRectGlobal` in that space.
  - Sonnet: paper-origin-based; content rect has non-zero `X,Y`; `VirtualPosition` is absolute on paper canvas.
- Panel representation:
  - Simple: `LeftPanel/RightPanel` clip rects, plus `DstRectLeft/Right` relative to page origins.
  - Sonnet: `Panels[]` with `PanelRect` and `ImageOffset` (offset for drawing relative to each panel).
- Rendering:
  - Simple file lacks renderer; Sonnet includes renderer with background/color/pdf.
- Config:
  - Sonnet has a robust YAML model with defaults/assets/foreach and `ResolvedSettings`.
  - Simple has a smaller config (`config.go`) with immediate `ToInputs` for `algorithm.go`.

Implication: We can unify on a paper-origin layout space and retain enough fields to adapt in both directions.

---

## 3) Proposed Common Model (pkg/spread/types)

Introduce shared types in `zine-layout/pkg/spread/types.go`:

```go
package spread

type Rect struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
    W float64 `json:"w"`
    H float64 `json:"h"`
}

type Point struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type Size struct {
    W int `json:"w"`
    H int `json:"h"`
}

type Panel struct {
    Name      string `json:"name"` // single|left|right
    ClipRect  Rect   `json:"clip_rect"` // clip on full paper canvas
    Offset    Point  `json:"offset"`   // draw offset for scaled image relative to panel origin
}

// Settings are the resolved, algorithm-agnostic inputs
// (paper, margins, spread, crop, scale, position, export)
// Derived from YAML or direct REST JSON.
type Settings struct {
    PaperWidthIn  float64 `json:"paper_width_in"`
    PaperHeightIn float64 `json:"paper_height_in"`
    DPI           float64 `json:"dpi"`
    Orientation   string  `json:"orientation"` // portrait|landscape

    MarginTopIn    float64 `json:"margin_top_in"`
    MarginRightIn  float64 `json:"margin_right_in"`
    MarginBottomIn float64 `json:"margin_bottom_in"`
    MarginLeftIn   float64 `json:"margin_left_in"`

    IsSpread bool    `json:"is_spread"`
    GutterIn float64 `json:"gutter_in"`

    CropRatio  *float64 `json:"crop_ratio"` // nil => original
    CropToFill bool     `json:"crop_to_fill"`

    UserScale float64 `json:"user_scale"`

    PositionX float64 `json:"position_x"`
    PositionY float64 `json:"position_y"`
    Units     string  `json:"units"` // px|normalized

    Export Export `json:"export"`
}

type Export struct {
    Format           string `json:"format"`    // png|jpg|pdf
    Quality          int    `json:"quality"`   // 1..100
    Background       string `json:"background"`// transparent|#RRGGBB
    OutDir           string `json:"out_dir"`
    FilenameTemplate string `json:"filename_template"`
}

type ImageMeta struct {
    Width  int `json:"width"`
    Height int `json:"height"`
}

type ComputeRequest struct {
    Algorithm string    `json:"algorithm"` // simple|sonnet
    Meta      ImageMeta `json:"meta"`
    Settings  Settings  `json:"settings"`
}

type Geometry struct {
    PaperSizePx   Size   `json:"paper_size_px"`
    ContentRect   Rect   `json:"content_rect"` // content box on paper canvas (px)
    GutterPx      int    `json:"gutter_px"`
    IsSpread      bool   `json:"is_spread"`

    SourceCrop    Rect   `json:"source_crop"`  // in source px
    DrawRect      Rect   `json:"draw_rect"`    // scaled crop dst rect on paper canvas

    Panels        []Panel `json:"panels"`

    // Optional
    Trace         []string          `json:"trace,omitempty"`
    Extras        map[string]any    `json:"extras,omitempty"`
}

type ComputeResponse struct {
    Geometry Geometry `json:"geometry"`
}

type RenderRequest struct {
    Algorithm string   `json:"algorithm"`
    Settings  Settings `json:"settings"`

    // Source image input: exactly one should be provided
    ImagePath string  `json:"image_path,omitempty"`
    ImageURL  string  `json:"image_url,omitempty"`
    ImageBase64 string `json:"image_base64,omitempty"`

    // Rendering options
    Mode      string `json:"mode"` // preview|final
    Panel     string `json:"panel,omitempty"` // combined|single|left|right
    MaxLongEdge int  `json:"max_long_edge,omitempty"` // for preview downscale
}

type OutputFile struct {
    Panel string `json:"panel"`
    Path  string `json:"path"`
}

type RenderResponse struct {
    // When return=inline for preview
    MimeType string `json:"mime_type,omitempty"`
    DataBase64 string `json:"data_base64,omitempty"`

    // When writing to disk or returning multiple panels
    Outputs []OutputFile `json:"outputs,omitempty"`
}
```

Rationale:
- Common `Settings` map cleanly to both algorithms.
- `Geometry.DrawRect` matches Simple’s `DstRectGlobal` (but in paper-space) and Sonnet’s `VirtualPosition + ImageDisplay`.
- `Panels` can be computed by both engines and used by a shared renderer.
- `Trace` is transported as strings; engines may also fill `Extras` with engine-specific insights.

---

## 4) Engine Interface and Switching (Strategy)

In `zine-layout/pkg/spread/engine.go`:

```go
package spread

type Engine interface {
    ID() string
    Compute(ctx context.Context, req ComputeRequest) (ComputeResponse, error)
    Render(ctx context.Context, req RenderRequest) (RenderResponse, error)
}

type EngineFactory func() Engine

var registry = map[string]EngineFactory{}

func Register(id string, f EngineFactory) {
    registry[id] = f
}

func Get(id string) (Engine, error) {
    if f, ok := registry[id]; ok {
        return f(), nil
    }
    return nil, fmt.Errorf("unknown engine %q", id)
}

func Default() Engine { // fallback to sonnet or simple
    if e, err := Get("sonnet"); err == nil { return e }
    if e, err := Get("simple"); err == nil { return e }
    return nil
}
```

- Engines register themselves in `init()`:
  - `pkg/spread/engines/simple` registers `"simple"`.
  - `pkg/spread/engines/sonnet` registers `"sonnet"`.
- Switching: callers pass `Algorithm` in the request or choose via config/flag.

---

## 5) Implementation Adapters per Engine

### 5.1 Simple Engine Adapter (pkg/spread/engines/simple)
- Port logic from `algorithm.go` into a package function using the common `Settings`.
- Convert paper inches/DPI to px; compute `ContentRect`, `GutterPx`, etc.
- Compute crop in source px, scale, and final draw rect.
- For spreads, compute left/right `Panel.ClipRect` and per-panel `Offset`.
- Populate `Geometry` with paper-space values: set `ContentRect.X/Y` to margin offsets (paper origin), and set `DrawRect` as `VirtualPosition + size` on paper.
- Render: reuse shared renderer (see 6) or provide a thin adapter that delegates to shared rendering based on `Geometry`.

Notes:
- `algorithm.go` currently uses content-origin. Adapter must translate to paper-origin: add `(marginLeftPx, marginTopPx)` to destination.
- Preserve tracing: reuse the `Trace` struct, append formatted lines into `Geometry.Trace`.

### 5.2 Sonnet Engine Adapter (pkg/spread/engines/sonnet)
- Wrap `internal/engine.Compute` and map `engine.Result` → `Geometry`.
  - `PaperSizePx` → `Geometry.PaperSizePx`
  - `EffectiveRect` (x,y,w,h) → `ContentRect`
  - `SourceCrop` → `SourceCrop`
  - `VirtualPosition` + `ImageDisplay` → `DrawRect`
  - `Panels` → `Geometry.Panels` with `ClipRect`=PanelRect and `Offset`=ImageOffset
  - `Trace` → strings
- Render: wrap `internal/engine.Render`. For preview mode, scale output down or render onto smaller canvas.

Notes:
- We keep Sonnet renderer intact for feature completeness (JPG/PDF, BG color). We can also add a shared renderer for preview.

---

## 6) Shared Rendering Plan (pkg/spread/render)

Introduce a general-purpose renderer driven by `Geometry`:

- Inputs: `Geometry`, `Settings.Export`, and `image.Image` or decoded buffer.
- Steps:
  1) Create full paper-size canvas with background color.
  2) Crop source to `Geometry.SourceCrop` (source px).
  3) Scale to `Geometry.DrawRect.W/H` and draw at `Geometry.DrawRect.X/Y`.
  4) If `IsSpread`, split final canvas into left/right respecting gutter via `ContentRect` and `GutterPx`.
  5) Encode outputs to PNG/JPG/PDF with quality.

Benefits:
- Both engines can rely on a single path for preview rendering.
- Sonnet’s existing renderer can still be used for final outputs; or we can unify entirely over time.

---

## 7) Package Layout (proposed)

- `pkg/spread/types.go` — shared types
- `pkg/spread/engine.go` — interface, registry, helpers
- `pkg/spread/render/renderer.go` — shared renderer (preview-focused, also supports final)
- `pkg/spread/engines/simple/compute.go` — adapter for Simple
- `pkg/spread/engines/sonnet/compute.go` — adapter for Sonnet compute
- `pkg/spread/engines/sonnet/render.go` — thin wrapper to Sonnet render (optional)
- `pkg/spread/config/...` — unify config reading and `ResolvedSettings` (primarily reuse Sonnet’s model)
- `cmd/spreadd/main.go` — REST server for compute/preview/render
- `cmd/spread-cli` — optional CLI unified over new pkg

Migration plan:
- Move `internal/engine` to `pkg/spread/engines/sonnet` (update imports).
- Move Simple `algorithm.go` logic into `pkg/spread/engines/simple` and expose as engine.
- Provide mapping helpers between Sonnet `ResolvedSettings` and common `Settings`.

---

## 8) REST API Design

Base path: `/api/v1`

### 8.1 Compute Geometry
- Endpoint: `POST /api/v1/compute`
- Request (JSON):
```json
{
  "algorithm": "sonnet",
  "meta": {"width": 5472, "height": 3648},
  "settings": {
    "paper_width_in": 8.5,
    "paper_height_in": 11.0,
    "dpi": 300,
    "orientation": "portrait",
    "margin_top_in": 0.25,
    "margin_right_in": 0.25,
    "margin_bottom_in": 0.25,
    "margin_left_in": 0.25,
    "is_spread": true,
    "gutter_in": 0.25,
    "crop_ratio": null,
    "crop_to_fill": true,
    "user_scale": 1.0,
    "position_x": 0,
    "position_y": 0,
    "units": "normalized",
    "export": {"format": "png", "quality": 90, "background": "transparent", "out_dir": "./out", "filename_template": "{index:03d}-{name}-{panel}.{ext}"}
  }
}
```
- Response (JSON):
```json
{
  "geometry": {
    "paper_size_px": {"w": 2550, "h": 3300},
    "content_rect": {"x": 75, "y": 75, "w": 2400, "h": 3150},
    "gutter_px": 75,
    "is_spread": true,
    "source_crop": {"x": 0, "y": 336, "w": 5472, "h": 3000},
    "draw_rect": {"x": 75, "y": 75, "w": 2400, "h": 3150},
    "panels": [
      {"name": "left",  "clip_rect": {"x": 0,    "y": 0, "w": 1237, "h": 3300}, "offset": {"x": 0,   "y": 0}},
      {"name": "right", "clip_rect": {"x": 1312, "y": 0, "w": 1238, "h": 3300}, "offset": {"x": 0,   "y": 0}}
    ],
    "trace": ["..."]
  }
}
```

### 8.2 Preview Image
- Endpoint: `POST /api/v1/preview`
- Request (JSON or multipart): same as Render (see below) but with `mode=preview` and optional `max_long_edge` (default 1600).
- Response:
  - `Content-Type: image/png` if `Accept: image/*`, or JSON with `{ mime_type, data_base64 }` if `Accept: application/json`.
  - Default `panel=combined` renders the full paper canvas (including both panels for spreads). `panel=left|right|single` returns only that panel cropped.

### 8.3 Final Render
- Endpoint: `POST /api/v1/render`
- Request (JSON):
```json
{
  "algorithm": "sonnet",
  "settings": { /* same as above */ },
  "image_path": "/abs/path/to/image.jpg",
  "mode": "final",
  "panel": "left"  
}
```
- Response (JSON):
```json
{ "outputs": [{"panel": "left", "path": "/abs/out/001-spread-left.png"}] }
```
- For spreads without `panel`, it returns both left/right outputs.
- For `panel=combined` it writes a single combined canvas image.

### 8.4 Notes
- CORS enabled for dev (`http://localhost:*`).
- Errors return JSON with `error` and `details`.
- For `image_base64`/`image_url`, the server decodes/downloads into memory; max size limit guarded by config.

---

## 9) CLI & Server

- New server: `cmd/spreadd/main.go` (cobra) with flags:
  - `--addr :8080`, `--max-upload-mb`, `--default-algorithm=sonnet`.
- Endpoints registered with `net/http` or `chi`.
- Optional unified CLI: `cmd/spread-cli` for compute/render (replacing experiment CLIs), forwarding to `pkg/spread`.

---

## 10) Testing Strategy

- Golden tests comparing Simple vs Sonnet for a matrix of scenarios:
  - portrait/landscape, single/spread, various margins/gutters
  - crop ratio equal/unequal, fill vs fit, different positions and scales
- Validate invariants:
  - `Geometry.DrawRect` inside paper bounds (allow overdraw for fill)
  - Panel clip rects partition the canvas with gutter in between
  - For `CropToFill`, aspect ratio tolerance
- Render smoke tests: produce small preview PNGs and checksum them.

---

## 11) Step-by-Step Implementation Plan

1) Create common types and engine interface
- [ ] Add `pkg/spread/types.go` with shared types (Settings, Geometry, etc.)
- [ ] Add `pkg/spread/engine.go` with `Engine`, registry, and helpers

2) Move/Adapt Sonnet engine
- [ ] Move `internal/engine` → `pkg/spread/engines/sonnet`
- [ ] Wrap Compute: map `engine.Result` → `Geometry`
- [ ] Wrap Render: adapt to `RenderRequest` and return `RenderResponse`
- [ ] Register engine `sonnet`

3) Move/Adapt Simple engine
- [ ] Port `algorithm.go` core compute into `pkg/spread/engines/simple/compute.go`
- [ ] Translate content-origin → paper-origin
- [ ] Implement Render via shared renderer
- [ ] Register engine `simple`

4) Shared renderer (initially for preview)
- [ ] Add `pkg/spread/render/renderer.go` implementing draw from `Geometry`
- [ ] Encode PNG/JPG/PDF; quality/bg; support `panel=combined|left|right|single`

5) REST server
- [ ] Add `cmd/spreadd/` with cobra + HTTP server
- [ ] Implement `/compute`, `/preview`, `/render` endpoints
- [ ] Add CORS and size limits

6) Config unification
- [ ] Add `pkg/spread/config` with loader → `Settings`
- [ ] Provide adapters to/from existing Sonnet config model

7) Integration & UI
- [ ] Wire frontend preview component to `/preview` (algorithm param)
- [ ] Add selection control for algorithm

8) Tests & CI
- [ ] Add geometry goldens and preview PNG checksums
- [ ] GitHub Actions job to run tests

---

## 12) Risks & Mitigations
- Coordinate origin mismatch (content vs paper) → explicit translation layer in Simple adapter.
- Divergent export semantics (page widths/gutter split) → centralize split math in shared renderer.
- Performance for large images in preview → downscale early; allow `max_long_edge` and `quality=speed` (use `ApproxBiLinear`).
- API churn while moving code → keep experiment CLIs by re-exporting new pkg types to avoid breaking local scripts.

---

## 13) Example: Using the API in Go

```go
eng, _ := spread.Get("sonnet")
comp, _ := eng.Compute(ctx, spread.ComputeRequest{
    Algorithm: "sonnet",
    Meta: spread.ImageMeta{Width: 5472, Height: 3648},
    Settings: settings,
})

rend, _ := eng.Render(ctx, spread.RenderRequest{
    Algorithm: "sonnet",
    Settings:  settings,
    ImagePath: "/path/img.jpg",
    Mode:      "preview",
    Panel:     "combined",
    MaxLongEdge: 1600,
})
```

---

## 14) Implementation Notes
- Keep zerolog integration via a `Tracer` adapter that collects messages and logs.
- Maintain existing filename templating; unify under `Export.FilenameTemplate`.
- Avoid removing any `XXX` comments or debug logging per project guidelines.
- Do not create directories explicitly outside of renderer file writes; respect existing paths.

---

## 15) Next Steps
- Proceed with Step 1 in the plan: add `pkg/spread/types.go` and `pkg/spread/engine.go`.
- After each step, run lints and minimal tests, then continue sequentially.
