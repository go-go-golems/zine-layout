# Stage C – Imposition and PDF Export: Implementation Context

Date: 2025-10-12

This document equips a new developer/intern with all the context to continue with PDF export and remaining Stage C tasks.

## What’s Implemented

- Page rendering pipeline
  - `pkg/pagelayout/renderer/renderer.go`
    - Symbol: `RenderPage(ctx RenderContext) (*PageRenderResult, error)`
    - Crops source using `imagelayout.ViewportResult.SourceRect`, composes into `PageLayoutSettings.ContentRectPx`, split spreads into `left`/`right`, produces variants map (`thumbnail`, `full`, `combined`, optionally `left`/`right`).
  - `pkg/services/pages.go`
    - Symbol: `(*PagesService).RenderPage(pageID string)`
    - Loads `Asset`, `LaidOutImage` (decode layout computation), `PageTemplate` → `PageLayoutSettings`, renders, writes PNGs to `projects/{project}/pages/{page}/`, stores result paths in `LaidOutPage.ResultJSON`.
  - HTTP preview/export for single page
    - `pkg/serve/laid_out_pages_routes.go`: `GET /api/laid-out-pages/{id}/preview?variant=...` (ETag/Last-Modified), `GET /api/laid-out-pages/{id}/export?variant=...` streams PNG.

- Imposition service (Stage C groundwork)
  - `pkg/services/imposition.go`
    - Type: `ImpositionService` with `SetDataRoot`.
    - Symbol: `(*ImpositionService).ImposeZine(zineID, presetID string) ([]*SheetResult, error)`
      - Loads zine and ordered pages via repos.
      - Ensures each page is rendered (uses `PagesService.RenderPage`).
      - Reads the best variant (combined → full → thumbnail) and decodes into `image.Image`.
      - Loads preset YAML from `{dataRoot}/presets/{presetID}.yaml` into `zinelayout.ZineLayout`.
      - For each `OutputPage`, invokes `ZineLayout.CreateOutputImage` to generate a sheet image.
    - Symbol: `(*ImpositionService).SaveSheetsAsPNGs(projectID, zineID string, sheets []*SheetResult)`
      - Utility to dump imposed sheets for inspection under `projects/{project}/zines/{zine}/imposed/`.
  - Server wiring
    - `pkg/serve/server.go`: `Server.impose` initialized in `initDatabase()` and injected with data root.

## What Remains (You Will Implement)

1) PDF Export
  - Add `pkg/export/pdf.go` with a function like:
    - `func SheetsToPDF(ctx context.Context, sheets []*services.SheetResult, dpi float64, out io.Writer) error`
    - Library: prefer `github.com/phpdave11/gofpdf` (maintained fork of gofpdf). Add to `go.mod` when implementing.
    - Each sheet → one PDF page. Compute points from pixels using `dpi` (1 pt = 1/72 in): `points = pixels * (72.0 / dpi)`.
    - Register each PNG as an image and place at origin with page size matching the sheet in points. Avoid recompression if possible.
    - Suggested helpers in `pkg/export/pdf.go`:
      - `func pixelsToPoints(px int, dpi float64) float64`
      - `func addSheetPage(pdf *gofpdf.Fpdf, imgPath string, wPx, hPx int, dpi float64) error`
    - Error handling: fail fast on missing files, propagate context cancellation.

2) HTTP Export Endpoint
  - `pkg/serve/zines_routes.go`: add route `GET /api/zines/{id}/export?preset=<id>` to stream a generated PDF.
  - Flow:
    - Parse `preset` query; default to a sensible preset if missing (optional).
    - Call `s.impose.ImposeZine(id, preset)` → get sheets.
    - Use `export.SheetsToPDF` to write to `w` with headers:
      - `Content-Type: application/pdf`
      - `Content-Disposition: attachment; filename="<zineID>-<preset>.pdf"`
    - Optional query: `dpi` (default 300). Validate `dpi > 0`.
    - Consider `ETag` on content-addressed temp files later; for now, stream directly.

3) CLI Verb (Workflow)
  - New file: `cmd/zine-layout/cmds/workflow/zines/export.go`.
  - Flags: `--data-root`, `--zine-id`, `--preset`, `--out` (path; if empty, write to stdout).
  - Flags (optional): `--dpi` (default 300).
  - Use `ImpositionService` then `SheetsToPDF` to produce the PDF.
  - Pseudocode sketch:
    ```go
    func newZinesExportCommand() (*cobra.Command, error) {
      // parse flags: data-root, zine-id, preset, out, dpi
      sheets, err := impose.ImposeZine(zineID, preset)
      // open out (or stdout), defer close
      err = export.SheetsToPDF(ctx, sheets, dpi, out)
    }
    ```

## Key Files and Symbols

- Repos and services
  - `pkg/repo/types.go`: `Zine`, `ZinePage`, `LaidOutPage` models.
  - `pkg/services/zines.go`: `GetZineWithPages`, `UpdateZinePages`.
  - `pkg/services/pages.go`: `RenderPage` writes PNGs, stores `ResultJSON` variant paths.
  - `pkg/services/imposition.go`: `ImposeZine`, `SaveSheetsAsPNGs`.
- Export
  - `pkg/export/pdf.go`: `SheetsToPDF(ctx context.Context, sheets []*services.SheetResult, dpi float64, out io.Writer) error`.
- Zine layout engine
  - `pkg/zinelayout/layout.go`: `ZineLayout`, `CreateOutputImage`.
  - `pkg/presets/presets.go`: helpers for presets on disk (list/seed/apply).
- HTTP
  - `pkg/serve/server.go`: service wiring and router setup.
  - `pkg/serve/zines_routes.go`: add export handler.
  - `pkg/serve/laid_out_pages_routes.go`: reference for streaming files and caching approach.
- CLI
  - `cmd/zine-layout/cmds/workflow/zines/`: add `export.go`.
  - `cmd/zine-layout/main.go`: already registers the workflow group.

## Data Layout Expectations

- Data root: server and workflow commands use `--data-root`.
- Pages: `projects/{projectID}/pages/{pageID}/{variant}.png`.
- Presets: `{dataRoot}/presets/{presetID}.yaml`.
- Imposed sheets (optional): `projects/{projectID}/zines/{zineID}/imposed/sheet-XX.png`.
  - For PDF export, images can be used directly from render locations; no need to save intermediary sheet PNGs unless debugging.

## Testing Plan (Post-PDF)

- Unit tests for `SheetsToPDF`: ensure page count matches sheets, basic metadata.
- HTTP smoke: `GET /api/zines/{id}/export?preset=...` returns `200` with `application/pdf`.
- CLI smoke: `workflow zines export --zine-id ... --preset ... --out test.pdf` and validate file exists and non-zero size.
  - End-to-end (server): `curl -sS -o out.pdf "http://localhost:8090/api/zines/<ZINE_ID>/export?preset=<PRESET>&dpi=300"` then `file out.pdf` and `pdfinfo out.pdf`.

## Notes and Caveats

- Ensure all pages used by a zine belong to the same project (already enforced in services). The imposition uses the project’s data root to load PNGs.
- If a page has no `combined` variant (non-spread), fallback to `full` (then `thumbnail`).
- Large sheets: consider memory—stream PDF page-by-page if the library supports it.
 - Orientation: page size in PDF should match sheet width/height from `ZineLayout.CreateOutputImage` (landscape vs portrait preserved).
 - DPI: Choose a single `dpi` for conversion; mismatched DPI will change on-paper size but not pixel fidelity.
 - Errors: if a page has no usable variant, surface a clear error with page ID; do not silently skip.

---

## Implementation Checklist (Copy/Paste when working)

- [ ] Add dependency: `github.com/phpdave11/gofpdf`.
- [ ] Create `pkg/export/pdf.go` with `SheetsToPDF` and helpers.
- [ ] Implement HTTP handler in `pkg/serve/zines_routes.go` (`handleZineExport`).
- [ ] Add workflow CLI `cmd/zine-layout/cmds/workflow/zines/export.go`.
- [ ] Test via CLI and HTTP; update `ttmp/2025-10-10/07-phase2-backend-and-ui-progress-changelog.md`.


