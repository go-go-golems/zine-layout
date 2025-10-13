# Zine Layout Platform – Project Handover & Onboarding

Date: 2025-10-12

## 1. Purpose and Scope
This repository implements a photo/zine layout platform. It takes uploaded images, computes layout placements, renders print pages (including spreads), and imposes them onto printable sheets based on YAML presets, with export to PDF. There is a REST API server, a React frontend, and a set of CLI verbs (API clients and direct workflow commands).

Read this first when onboarding; it links to all the critical code paths and living docs.

## 2. High-Level Architecture
- Backend (Go): repositories (SQLite), services, REST server, renderers.
- Frontend (React/RTK Query): UI to manage projects, templates, pages, zines.
- Data roots: on-disk project storage (`projects/`, `uploads/`, `presets/`).
- Rendering pipeline:
  1) Asset upload → imagelayout (crop/scale math) → laid-out images metadata.
  2) Page renderer takes a laid-out image + page template → PNG variants (thumbnail/full/combined/left/right) + metadata.
  3) Imposition engine arranges rendered pages into sheet images per YAML preset → export to PDF (one sheet per PDF page).

## 3. Key Directories and Files
- CLI: `cmd/zine-layout/`
  - `cmds/api/*`: HTTP client verbs (requires server)
  - `cmds/workflow/*`: direct DB/service verbs (no server required)
  - `cmds/serve/command.go`: server launcher (`serve`)
  - `cmds/render/command.go`: standalone YAML imposition renderer
- Backend:
  - Repositories (SQLite): `pkg/repo/sqlite/*.go` (one per entity), types in `pkg/repo/types.go`
  - Services: `pkg/services/*.go` (layout, pages, zines, imposition)
  - REST server: `pkg/serve/*.go` (routes per entity, `server.go` bootstraps)
  - Renderer: `pkg/pagelayout/renderer/renderer.go`, settings in `pkg/pagelayout/settings.go`
  - Imposition engine: `pkg/zinelayout/*` (YAML → sheet images)
  - PDF export: `pkg/export/pdf.go` (`SheetsToPDF`)
- Frontend: `web/` (React app, `src/api.ts`, views in `src/views/*`)
- Documentation: `ttmp/2025-10-10/*`, `ttmp/2025-10-11/*`

See also: `ttmp/2025-10-10/index.md` for the codebase index and orientation.

## 4. Data Flow and Core Symbols
- Imagelayout math: `pkg/imagelayout/engine.ComputeViewport` (settings/result types in `pkg/imagelayout`)
- Layout service: `pkg/services/layout.go`
  - `CreateLaidOutImage`, `RecomputeLaidOutImage`, `ApplyTemplateToSequence`
- Page renderer and service:
  - Settings: `pkg/pagelayout/settings.go` (`PageLayoutSettings`)
  - Render helper: `pkg/pagelayout/renderer/renderer.go` (`RenderPage(RenderContext)`)
  - Pages service: `pkg/services/pages.go` (`CreatePage`, `UpdatePageImage`, `RenderPage`)
  - REST: `pkg/serve/laid_out_pages_routes.go` (`/api/laid-out-pages/{id}/preview|export`)
- Imposition and export:
  - Presets YAML → images: `pkg/zinelayout/layout.go` (`CreateOutputImage`)
  - Imposition service: `pkg/services/imposition.go` (`ImposeZine`, `SaveSheetsAsPNGs`)
  - PDF export: `pkg/export/pdf.go` (`SheetsToPDF`)
  - REST: `pkg/serve/zines_routes.go` (`/api/zines/{id}/export`)

## 5. What Was Recently Implemented
- Stage A (Page-level rendering – 2A):
  - Added `PageLayoutSettings` and renderer (`renderer.RenderPage`) with crop from `imagelayout.ViewportResult.SourceRect`, scaling into `ContentRectPx`, split for spreads, optional borders.
  - Service `PagesService.RenderPage` writes PNG variants under `projects/{project}/pages/{page}/` and stores variant paths JSON (`ResultJSON`).
  - Preview endpoint now streams variant PNGs and sets `ETag/Last-Modified` for caching.
  - Unit tests for renderer crop and spread splits.
- Stage C (Imposition & Export):
  - Implemented `ImpositionService.ImposeZine` to load YAML presets (seeded to `presets/`), read rendered page variants, generate sheet images via `ZineLayout.CreateOutputImage`.
  - Implemented `pkg/export/pdf.go` and CLI/HTTP export paths:
    - HTTP: `GET /api/zines/{id}/export?preset=10_8_sheet_zine&dpi=300` streams a PDF.
    - CLI: `workflow zines export --data-root <root> --zine-id <id> --preset 10_8_sheet_zine --out out.pdf --dpi 300`.
  - Smoke-tested end-to-end; sample PDF saved and validated.

See change logs/specs:
- `ttmp/2025-10-10/07-phase2-backend-and-ui-progress-changelog.md`
- `ttmp/2025-10-10/09-system-specification-after-phase1-and-phase2.md`
- `ttmp/2025-10-11/19-phase3-focused-roadmap.md`
- `ttmp/2025-10-11/20-imposition-and-pdf-export-context.md`
- `ttmp/2025-10-11/21-end-to-end-cli-test-plan.md`
- `ttmp/2025-10-11/22-cli-verbs-inventory-and-mapping.md`

## 6. How to Run and Test Quickly
- Start server:
  - `go run ./cmd/zine-layout serve --data-root ./data --addr :8088`
- End-to-end CLI test (creates and reuses JSON outputs):
  - Follow `ttmp/2025-10-11/21-end-to-end-cli-test-plan.md`. It generates images, creates project/assets/templates, renders pages, creates a zine, and exports PDFs (HTTP + CLI), saving JSON state to `./tmp-e2e/*.json`.
- Health/Preview checks:
  - `curl -sS http://localhost:8088/api/health`
  - `curl -sS http://localhost:8088/api/laid-out-pages/<PAGE_ID>/preview?variant=thumbnail -o thumb.png`

## 7. CLI Verbs Overview
- API verbs (require server): `cmd/zine-layout/cmds/api/*`
  - Projects, Assets (uploads/list), Image Layout Templates, Laid-Out Images, Image/Layout Sequences
- Workflow verbs (direct DB/services): `cmd/zine-layout/cmds/workflow/*`
  - Page Templates, Laid-Out Pages (incl. `render`), Zines (`create`, `set-pages`, `export`)
- Inventory and mappings with flags/files: `ttmp/2025-10-11/22-cli-verbs-inventory-and-mapping.md`

## 8. Data Layout & Presets
- Data root (`--data-root`) contains:
  - `uploads/` – raw uploads
  - `projects/{project}/pages/{page}/` – page variant PNGs
  - `presets/` – YAML imposition presets; seeded on server startup if missing
- Preset example: `examples/tests/10_8_sheet_zine.yaml` (copied into data root)

## 9. Known Gaps / Next Steps
- CLI coverage (serverless) missing for: Projects, Assets, Image Layout Templates, Laid-Out Images, Image/Layout Sequences. See proposals in `22-cli-verbs-inventory-and-mapping.md`.
- Frontend: wire Page/Zine tabs to live data & page thumbnails; fix listed UX issues.
- Garbage collect stale page render files upon updates.
- More tests: services (`pages`, `imposition`), HTTP handlers, CLI integration.

## 10. Handover Checklist
- Read: `ttmp/2025-10-10/index.md` (index), `19-phase3-focused-roadmap.md` (plan), `20-imposition-and-pdf-export-context.md` (Stage C details), `21-end-to-end-cli-test-plan.md` (steps), `22-cli-verbs-inventory-and-mapping.md` (verbs/flags/mapping).
- Build and run server; execute the E2E plan on a fresh data root.
- Review `pkg/services/*` and `pkg/serve/*` for business logic and routes.
- Extend workflow CLI per proposals if you need serverless operation for more flows.

## 11. Quick Reference – Critical Files
- Server and routes: `pkg/serve/server.go`, `pkg/serve/*_routes.go`
- Renderer: `pkg/pagelayout/renderer/renderer.go`, `pkg/pagelayout/settings.go`
- Services: `pkg/services/layout.go`, `pkg/services/pages.go`, `pkg/services/zines.go`, `pkg/services/imposition.go`
- Export: `pkg/export/pdf.go`
- Repos/Types: `pkg/repo/types.go`, `pkg/repo/sqlite/*.go`
- CLI: `cmd/zine-layout/cmds/*`

If you need deeper internals, open the files above and follow the symbols listed here; the code is intentionally organized to mirror the domain workflows.
