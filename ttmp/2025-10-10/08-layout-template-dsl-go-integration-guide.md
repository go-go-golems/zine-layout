# Phase 2 Deep Dive: Layout Template DSL & Go Integration

This guide explains how to give image layout templates the same expressive controls that power the interactive resizer in `ttmp/2025-10-10/02-image-resizer-code.tsx`. In this context, an image layout template defines **the destination viewport for a single asset**—how large the rendered frame should be, which portion of the source appears, and how anchors/zoom affect placement. It does *not* define the final page or spread composition, and it no longer captures gutter logic (that will be handled later when assembling pages). The guide introduces a YAML/JSON DSL that mirrors the TSX UI, outlines how to compile the DSL into the new `imagelayout.ViewportSettings` model described in `ttmp/2025-10-10/09-system-specification-after-phase1-and-phase2.md`, and describes the Go-side changes needed across repositories, services, HTTP routes, and CLI utilities now that `pkg/spread` is being removed.

---

## 1. Goals and Alignment
- Give template authors the same viewport knobs that exist in `02-image-resizer-code.tsx` (modes, anchors, margins, zoom, drag) so each asset renders into the intended destination frame.
- Store templates in a declarative DSL, not ad-hoc JSON blobs, so they can be shared, versioned, and linted.
- Replace `pkg/spread` with a dedicated `pkg/imagelayout` module that exposes typed `ViewportSettings` and pure layout computations (see §5).
- Allow UI and CLI clients to submit either DSL text or compiled JSON; the server compiles DSL → `imagelayout.ViewportSettings` and persists both.

---

## 2. Current Architecture Recap
- **Persistence** – `pkg/repo/sqlite/migrations.go` creates `image_layout_templates.settings_json` (JSON string). Templates are fetched through `pkg/repo/ImageLayoutTemplateRepository`.
- **Service layer** – `pkg/services/layout.go` currently deserializes template JSON into `spread.Settings`; the next iteration must switch to `imagelayout.ViewportSettings`.
- **HTTP API** – `pkg/serve/layout_templates_routes.go` accepts arbitrary `settings` maps when creating/updating templates; no DSL support or validation beyond JSON marshal.
- **UI** – `web/src/views/LayoutTemplateManager.tsx` exposes a raw JSON textarea. Phase 2 will replace this with controls derived from `02-image-resizer-code.tsx`.
- **CLI** – Phase 2 verbs (see `cmd/zine-layout/cmds/image-layout-templates/*`) forward JSON without transformation.

---

## 3. Target Flow & Responsibilities
1. **Authoring (UI/CLI)** – Users edit viewport parameters through TSX controls or write DSL snippets manually (no gutter fields).
2. **Transport** – Request body includes `settings_dsl` (string) and optionally `settings_json` (compiled cache).
3. **Backend compilation** (`pkg/imagelayout/dsl`) – Parse DSL → `TemplateSpec`; normalize into `imagelayout.ViewportSettings`.
4. **Persistence** – Save:
   - `settings_json`: canonical JSON encoding of `imagelayout.ViewportSettings`.
   - `settings_dsl`: original DSL for round-tripping (new column).
5. **Runtime usage** – Services that generate laid-out images use the compiled `imagelayout.ViewportSettings` to transform individual assets before they join sequences or pages.
6. **Round-trip editing** – When reading templates, server returns both `settings` (parsed JSON) and `settings_dsl` so the UI can rebuild the interactive form.

---

## 4. Layout Template DSL Specification (Version `layout-template/v1`)

### 4.1 Top-Level Shape
```yaml
version: layout-template/v1
name: Full Bleed Portrait
description: Optional human notes
mode: page            # page | crop | fit

canvas:
  width: 1920         # numeric
  height: 1080
  units: px           # px | in
  dpi: 300            # required if units == px

margins:
  top: 100
  right: 100
  bottom: 100
  left: 100
  units: px | in
  linked: false       # aligns with marginLinked toggle

fit:
  strategy: cover     # cover | contain | fitWidth | fitHeight
  crop_to_fill: true  # maps to CropToFill
  width: 1600         # optional fit target width (px)
  height: 1200        # optional fit target height (px)

position:
  mode: anchor        # anchor | drag | focus
  anchor:
    x: center         # left|center|right or 0.0-1.0
    y: middle         # top|middle|bottom or 0.0-1.0
  drag:
    x: 0
    y: 0
  focus:
    source: { x: 1200, y: 800 }
    target: { x: 960, y: 540 }
  units: normalized   # normalized | px

anchor_preset: center   # center|top-left|bottom-right etc.

adjustments:
  zoom: 1.0           # maps to imagelayout.ViewportSettings.UserScale

crop:
  aspect: "16:9"          # matches TSX dropdown
  width: 1000             # optional override (px)
  height: 562

focus:
  source: { x: 3200, y: 1400 }  # pixel in source image
  target: { x: 0.25, y: 0.2 }   # normalized position inside viewport

output:
  format: png
  quality: 90
  background: "#ffffff"
  filename_template: "{index:03d}-{name}.{ext}"
  out_dir: "./out"
```

### 4.2 Mode Semantics
| Mode    | UI behaviour (from `02-image-resizer-code.tsx`) | DSL interpretation | `imagelayout.ViewportSettings` impact |
|---------|--------------------------------------------------|--------------------|---------------------------|
| `page`  | Single viewport with margins                     | Use `canvas`, `margins`, `fit`, `position`, `adjustments`. | `IsSpread=false`, `GutterIn=0`. Margins become `Margin*In`. |
| `crop`  | Fixed crop frame with anchor/focus controls       | Use `crop.aspect/width/height` to derive `CropRatio`; `fit.strategy` toggles cover/contain. | Set `CropRatio` and `CropToFill`. Viewport size comes from `canvas`. |
| `fit`   | Fit-to-dimension viewport                         | Same as `page` but margins default to zero; `fit.strategy` + `fit.width/height` define canvas size. | Sets `FitMode`, `FitWidthPx`, `FitHeightPx`; engine keeps `IsSpread=false`. |

### 4.3 Conversion Rules
- **Units** – Allow `px` or `in` on `canvas` and `margins`. When `px`, convert to inches via `value / dpi`.
- **Anchors** – `anchor_preset` resolves to normalized offsets (left=-1, center=0, right=1; top=-1, middle=0, bottom=1). Additional drag offsets are applied afterward (converted if units=`normalized`).
- **Focus point** – Convert the `focus` block into `imagelayout.FocusPoint`. Source values are clamped to the image bounds; target values treat `0..1` as normalized and `>1` as pixels inside the viewport.
- **Crop aspect** – Parse strings like `16:9`; store as `CropRatio`. If `width/height` explicit, compute ratio = width/height.
- **Fit strategy** – `fit.strategy` guides placement: `cover` ⇒ `CropToFill=true`; `contain` ⇒ `CropToFill=false`; `fitWidth`/`fitHeight` populate `FitMode` + `FitWidthPx`/`FitHeightPx` for the engine.
- **Zoom** – Directly assign to `UserScale`.
- **Drag limits** – The UI enforces limits; backend simply accepts computed offsets.

---

## 5. Go Implementation Blueprint

### 5.1 Package Layout
- Create `pkg/imagelayout/dsl` (new package).
  - `document.go` – DTO structs (`TemplateDocument`, `Canvas`, `Margins`, `Fit`, `Position`, `Crop`, `Output`).
  - `parse.go` – Parse YAML/JSON into `TemplateDocument` using `gopkg.in/yaml.v3` (see parsing helpers in `pkg/imagelayout/spec` when they land).
  - `compile.go` – Convert `TemplateDocument` → `imagelayout.ViewportSettings`.
  - `roundtrip.go` – Reverse conversion for editing (optional Phase 2.1).
  - `validation.go` – Human-friendly error messages (include field paths).

### 5.2 Conversion Pseudocode
```go
func Compile(doc *TemplateDocument) (TemplateCompilation, error) {
    if doc.Version == "" { doc.Version = "layout-template/v1" }
    if doc.Mode == "" { doc.Mode = "page" }

    dpi := doc.Canvas.DPI
    if doc.Canvas.Units == "px" && dpi <= 0 {
        return TemplateCompilation{}, fmt.Errorf("canvas.dpi required when units=px")
    }

    conv := newUnitConverter(dpi)
    settings := imagelayout.ViewportSettings{
        Mode:         doc.Mode,
        DPI:          chooseDPI(doc.Canvas, defaults),
        PaperWidthIn: conv.ToInches(doc.Canvas.Width, doc.Canvas.Units),
        PaperHeightIn: conv.ToInches(doc.Canvas.Height, doc.Canvas.Units),
        Orientation:  inferOrientation(doc.Canvas, doc.Mode),
        MarginTopIn:    conv.ToInches(doc.Margins.Top, doc.Margins.Units),
        MarginRightIn:  conv.ToInches(doc.Margins.Right, doc.Margins.Units),
        MarginBottomIn: conv.ToInches(doc.Margins.Bottom, doc.Margins.Units),
        MarginLeftIn:   conv.ToInches(doc.Margins.Left, doc.Margins.Units),
        CropToFill:   doc.Fit.Strategy == "cover",
        UserScale:    clampZoom(doc.Adjustments.Zoom),
        Units:        normalizeUnits(doc.Position.Units),
    }

    ratio, err := resolveCropRatio(doc)
    if err != nil { return TemplateCompilation{}, err }
    settings.CropRatio = ratio
    settings.CropWidthPx, settings.CropHeightPx = doc.Crop.OutputDimensions()
    settings.FitMode, settings.FitWidthPx, settings.FitHeightPx = doc.Fit.OutputDimensions()

    posX, posY, err := resolvePosition(doc.Position, doc.Mode, doc.Canvas, conv)
    if err != nil { return TemplateCompilation{}, err }
    settings.PositionX = posX
    settings.PositionY = posY
    settings.AnchorPreset = doc.Position.AnchorPreset
    settings.Focus = doc.Position.FocusPointer()

    export := imagelayout.ExportOptions{
        Format:           firstNonEmpty(doc.Output.Format, "png"),
        Quality:          coalesceQuality(doc.Output.Quality),
        Background:       doc.Output.Background,
        OutDir:           doc.Output.OutDir,
        FilenameTemplate: doc.Output.FilenameTemplate,
    }
    settings.Export = export

    return TemplateCompilation{
        Settings: settings,
        Metadata: TemplateMetadata{
            Name: doc.Name,
            Mode: doc.Mode,
            DSL:  doc.RawPayload,
        },
    }, nil
}
```
`TemplateCompilation` should keep the normalized `imagelayout.ViewportSettings`, derived metadata (mode, named anchors resolved), and any validation warnings to echo back to the caller. Because templates no longer handle gutters, `IsSpread` remains `false` in compiled settings.

### 5.3 Helper Strategies
- `unitConverter` – centralize inch/pixel conversion to avoid drift.
- `resolvePosition` – Mirror the math in `anchorToTranslation` and related helpers from `02-image-resizer-code.tsx` so anchor + drag behaviour matches the UI.
- `resolveAnchorPreset` – Translate `position.anchor.preset` (or DSL shorthand) to normalized offsets.
- `resolveFocus` – Convert DSL focus blocks into `imagelayout.FocusPoint`, clamping targets and source coords.
- `resolveCropRatio` – Accept numeric, `"w:h"` string, or fallback to canvas aspect.
- `resolveFitDimensions` – Derive fit width/height defaults when DSL omits explicit values.
- `inferOrientation` – Use width/height ratio or explicit override.
- Validation should echo friendly errors: `fmt.Errorf("margins.left requires non-negative number, got %v", doc.Margins.Left)`.

### 5.4 Tests
- Add table-driven tests in `pkg/imagelayout/dsl/compile_test.go`.
  - Each mode (page, crop, fit) with fixtures covering anchor and drag scenarios.
  - Round-trip test: compile → `imagelayout.ViewportSettings` → rehydrate DSL (when reverse implemented).
  - Negative tests: missing DPI, invalid mode, malformed aspect string.

---

## 6. Persistence & API Changes

### 6.1 Database
- Update `pkg/repo/sqlite/migrations.go`:
  ```sql
  ALTER TABLE image_layout_templates ADD COLUMN settings_dsl TEXT;
  ALTER TABLE image_layout_templates ADD COLUMN settings_version TEXT DEFAULT 'layout-template/v1';
  ```
  Ensure migration is idempotent; extend `schemaSQL`.

- Reflect new fields in `pkg/repo/types.go` (`ImageLayoutTemplate` struct) and repository implementations under `pkg/repo/sqlite`.

### 6.2 Repository Methods
- When creating/updating templates, accept both `SettingsJSON` and `SettingsDSL`. Add helper:
  ```go
  func (r *ImageLayoutTemplateRepository) CreateWithDSL(t *repo.ImageLayoutTemplate) error
  ```
  or just reuse `Create` with populated fields.

### 6.3 HTTP Layer (`pkg/serve/layout_templates_routes.go`)
- Extend request payloads:
  ```go
  var req struct {
      Name        string         `json:"name"`
      Description string         `json:"description"`
      Settings    map[string]any `json:"settings"`       // optional fallback
      SettingsDSL string         `json:"settings_dsl"`   // new
  }
  ```
- Preference order:
  1. If `settings_dsl` present → compile via `pkg/imagelayout/dsl`.
  2. Else if `settings` present → use as today.
  3. Reject empty payload.
- On success, respond with both:
  ```json
  {
    "template": {
      "id": "...",
      "name": "...",
      "settings": { ... },          // parsed JSON
      "settings_dsl": "version...\n"
    }
  }
  ```
  `projection` is optional metadata (e.g., resolved anchors) returned by the compiler for UI hints.

### 6.4 Service Layer (`pkg/services/layout.go`)
- Replace manual JSON merge with DSL-aware helper:
  ```go
  func mergeTemplateSettings(baseJSON, overridesJSON *string, dsl *string) (imagelayout.ViewportSettings, *string, error)
  ```
- If template has `SettingsDSL`, compile it once at load time, cache the resulting JSON, and use that for overrides. Consider storing compiled JSON alongside DSL during template creation to avoid runtime compilation.

### 6.5 CLI Commands
- Update `cmd/zine-layout/cmds/image-layout-templates/*` to accept `--dsl path/to/template.yaml`.
- Provide helper that loads DSL, posts to `/api/...` with `settings_dsl`.

---

## 7. Mapping TSX Controls → DSL → `imagelayout.ViewportSettings`

| TSX State (see `02-image-resizer-code.tsx`) | DSL Field | `imagelayout.ViewportSettings` |
|---------------------------------------------|-----------|-------------------|
| `mode`                                      | `mode`    | Determines which sections apply; compiled settings keep `IsSpread=false`. |
| `pageWidth`, `pageHeight`                   | `canvas.width/height` | `PaperWidthIn`, `PaperHeightIn` |
| `margin{Top,Right,Bottom,Left}`             | `margins.*` | `Margin*In` |
| `pageCropFill`, `fitCropFill`                   | `fit.crop_to_fill` | `CropToFill` |
| `pageAnchorX/Y`, `cropAnchorX/Y`, etc.      | `position.anchor` | `PositionX`, `PositionY` |
| `dragX`, `dragY`                            | `position.drag` | `PositionX`, `PositionY` adjustments |
| `zoom`                                      | `adjustments.zoom` | `UserScale` |
| Crop ratio dropdown                         | `crop.aspect` | `CropRatio` |
| `fitMode` / `cropFitMode`                   | `fit.strategy` | `CropToFill`, `FitMode` |
| Fit width / height inputs                   | `fit.width`, `fit.height` | `FitWidthPx`, `FitHeightPx` |
| Anchor preset dropdown                      | `anchor_preset` | `AnchorPreset` |
| Focus picker                                | `focus.{source,target}` | `Focus` |
| Export panel (future UI)                    | `output.*` | `imagelayout.ExportOptions` |

Use this table when wiring the new React controls and when writing unit tests for the compiler.

---

## 8. Regression-Resistant Testing Strategy
- **Compiler tests** (`pkg/imagelayout/dsl`): golden fixtures for each mode; compare against expected `imagelayout.ViewportSettings` JSON.
- **Service smoke tests** (`pkg/serve/server_rest_test.go` or new file): POST template with DSL, GET should echo DSL + JSON; creating laid-out image uses compiled settings.
- **Integration CLI test**: extend Phase 2 smoke script to run `zine-layout image-layout-templates create --dsl fixtures/full-bleed.yaml` and verify API response.
- **UI contract**: Add JSON Schema (under `web/src/utils/layoutTemplateSchema.ts`) derived from DSL spec; use Ajv to validate compiled payload before submission.

---

## 9. Migration & Rollout Steps
1. **Schema migration** – Add `settings_dsl`/`settings_version` columns and backfill existing rows with minimal stub DSL (`mode: page` plus values from JSON).
2. **Compiler package** – Implement parser + compiler + tests.
3. **Server wiring** – Update repositories, services, and routes to use the compiler.
4. **CLI update** – Add `--dsl` option and documentation in `README.md`.
5. **UI refresh** – Replace textarea with designer component; ensure it serializes to DSL and shows YAML editor for advanced tweaks.
6. **Docs** – Update `05-expansion-plan...` and `07-phase2...` after landing each milestone.

Roll out in feature flags if necessary: accept DSL first, keep JSON fallback, then deprecate raw JSON once the UI ships.

---

## 10. Reference Files
- `ttmp/2025-10-10/02-image-resizer-code.tsx` – UI behaviour source of truth.
- `ttmp/2025-10-10/03-template-resize-dsl.md` – Prior DSL concepts; reuse terminology.
- `pkg/imagelayout/types.go` (to add) – Canonical `ViewportSettings` definition.
- `pkg/imagelayout/engine` (to add) – Placement computations that supersede `pkg/spread/simple`.
- `pkg/imagelayout/dsl` (to add) – DSL compiler described in this guide.
- `pkg/services/layout.go` – Consumers of compiled settings.
- `pkg/serve/layout_templates_routes.go` – API entry-point to hook DSL compilation.
- `cmd/zine-layout/cmds/image-layout-templates/*` – CLI verbs to extend with DSL support.
- `web/src/views/LayoutTemplateManager.tsx` – UI component to replace with the new designer.

---

## 11. CLI Helpers
- Local sanity check: `go run ./cmd/zine-layout imagelayout compute --source-width 4000 --source-height 3000 --mode crop --crop-width 1600 --crop-height 1600 --focus-source-x 2800 --focus-source-y 1500 --focus-target-x 0.3 --focus-target-y 0.4`
- From YAML: `zine-layout imagelayout compute --spec viewport.yaml` emits the merged `imagelayout.Computation` (settings + result + trace).

---

By following this guide, the Go backend will gain a robust translation layer from the user-friendly template DSL to the new image layout engine, enabling richer UI controls while we retire `pkg/spread` and align the codebase with the Phase 2 specification.
