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
    - `func SheetsToPDF(sheets []*services.SheetResult, out io.Writer) error`
    - Use a maintained PDF lib (e.g., `github.com/jung-kurt/gofpdf` or `github.com/phpdave11/gofpdf`) to create a PDF.
    - Each sheet is one PDF page; scale 1:1 pixel to point with a DPI assumption (or fit to page size from preset’s `PPI`).

2) HTTP Export Endpoint
  - `pkg/serve/zines_routes.go`: add route `GET /api/zines/{id}/export?preset=<id>` to stream a generated PDF.
  - Flow:
    - Parse `preset` query; default to a sensible preset if missing (optional).
    - Call `s.impose.ImposeZine(id, preset)` → get sheets.
    - Use `export.SheetsToPDF` to write to `w` with `Content-Type: application/pdf` and `Content-Disposition: attachment`.

3) CLI Verb (Workflow)
  - New file: `cmd/zine-layout/cmds/workflow/zines/export.go`.
  - Flags: `--data-root`, `--zine-id`, `--preset`, `--out` (path; if empty, write to stdout).
  - Use `ImpositionService` then `SheetsToPDF` to produce the PDF.

## Key Files and Symbols

- Repos and services
  - `pkg/repo/types.go`: `Zine`, `ZinePage`, `LaidOutPage` models.
  - `pkg/services/zines.go`: `GetZineWithPages`, `UpdateZinePages`.
  - `pkg/services/pages.go`: `RenderPage` writes PNGs, stores `ResultJSON` variant paths.
  - `pkg/services/imposition.go`: `ImposeZine`, `SaveSheetsAsPNGs`.
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

## Testing Plan (Post-PDF)

- Unit tests for `SheetsToPDF`: ensure page count matches sheets, basic metadata.
- HTTP smoke: `GET /api/zines/{id}/export?preset=...` returns `200` with `application/pdf`.
- CLI smoke: `workflow zines export --zine-id ... --preset ... --out test.pdf` and validate file exists and non-zero size.

## Notes and Caveats

- Ensure all pages used by a zine belong to the same project (already enforced in services). The imposition uses the project’s data root to load PNGs.
- If a page has no `combined` variant (non-spread), fallback to `full` (then `thumbnail`).
- Large sheets: consider memory—stream PDF page-by-page if the library supports it.


