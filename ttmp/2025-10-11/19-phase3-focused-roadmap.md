# Phase 3 Focused Roadmap (Backend Rendering → Frontend Wiring → Imposition)
**Date:** 2025-10-11  
**Sources distilled:** `ttmp/2025-10-11/13-phase3-and-phase4-implementation-guide.md`, `ttmp/2025-10-11/14-changelog-for-ui-refactor-for-page-layouts.md`, `ttmp/2025-10-11/15-phase3-api-status-and-discrepancies.md`, `ttmp/2025-10-11/16-spread-rendering-visualization-guide.md`, `ttmp/2025-10-11/16-todo-notes.md`, `ttmp/2025-10-11/17-zinelayout-imposition-integration.md`, `ttmp/2025-10-11/18-fixes-for-image-layouts-tab-issues.md`

---

## 1. Current Status Snapshot
- **Phase 3 REST APIs**: Already live (`pkg/serve/page_templates_routes.go`, `pkg/serve/laid_out_pages_routes.go`, `pkg/serve/zines_routes.go`). Earlier guidance in `ttmp/2025-10-11/13-phase3-and-phase4-implementation-guide.md` that these need to be written is outdated; rely on the confirmed analysis in `ttmp/2025-10-11/15-phase3-api-status-and-discrepancies.md`.
- **Repositories & services**: CRUD flows exist (`pkg/services/pages.go` and `pkg/services/zines.go`). Rendering hooks still return `ErrPageRendererNotImplemented`.
- **Image layout math**: Implemented in `pkg/imagelayout/engine/engine.go`; use it (do not port TS logic).
- **Imposition**: `pkg/zinelayout` already handles sheet assembly from YAML presets (`ttmp/2025-10-11/17-zinelayout-imposition-integration.md`). Treat this as reusable infrastructure.
- **Frontend**: Tabs and RTK Query hooks are scaffolded, but Page/Zine tabs still use placeholder data (`web/src/views/tabs/PageLayoutsTab.tsx`, `web/src/views/tabs/ZineTab.tsx`). Image Layouts tab has a set of UX defects documented in `ttmp/2025-10-11/18-fixes-for-image-layouts-tab-issues.md`.

> **Mismatch to watch:** `ttmp/2025-10-11/13-phase3-and-phase4-implementation-guide.md` lists REST endpoints as TODO; ignore those steps and instead focus on rendering, frontend wiring, and export path. This document replaces the verbose instructions with an up-to-date action plan.

---

## 2. Implementation Roadmap

### Stage A – Page-Level Rendering Pipeline
1. **Introduce page layout settings types.**  
   - Location: create `pkg/pagelayout/settings.go`.  
   - Capture the schema hinted in `ttmp/2025-10-11/13-phase3-and-phase4-implementation-guide.md §3A` but trim unused fields (focus/anchor belong to laid-out image overrides per `ttmp/2025-10-11/18-fixes-for-image-layouts-tab-issues.md`).  
   - Pseudocode sketch:
     ```go
     type PageLayoutSettings struct {
       PageWidthIn, PageHeightIn, DPI float64
       MarginTopIn, MarginRightIn, MarginBottomIn, MarginLeftIn float64
       IsSpread bool
       GutterWidthIn, GutterOverlapIn float64 // only meaningful when IsSpread
       PositioningMode string // "fill" | "absolute" | "snap"
       AnchorPreset string   // normalized preset, align with imagelayout anchor map
       // Absolute overrides only used when PositioningMode == "absolute"
       ImageXIn, ImageYIn, ImageWidthIn, ImageHeightIn float64
     }
     
     func (s *PageLayoutSettings) Canonicalize() error {
       // enforce DPI > 0, margins >= 0, gutter < page width, etc.
     }
     ```
   - Ensure the JSON tags match API payloads (`page_templates_routes.go:58` expects `template` map; wire this struct via marshaling).

2. **Build renderer helper (thumbnail + full page).**  
   - Location: add `pkg/pagelayout/renderer/renderer.go`.  
   - Inputs: `PageLayoutSettings`, `imagelayout.LayoutComputation` (from `pkg/services/layout.go`), plus source image path.  
   - Use the thumbnail recipe in `ttmp/2025-10-10/17-imagelayout-rendering-algorithm-analysis.md §6` as the core algorithm: crop using `ViewportResult.SourceRect`, scale into the canvas defined by page settings, then split for spreads.  
   - Pseudocode structure:
     ```go
     func RenderPage(ctx RenderContext) (PageRenderResult, error) {
       layout := ctx.Layout // imagelayout.ViewportResult
       page := ctx.Settings // PageLayoutSettings
       canvas := newRGBA(page.PixelWidth(), page.PixelHeight())
       drawBackground(canvas, ctx.Background)
       placeImage(canvas, cropAndScale(ctx.SourceImage, layout, page, ctx.Mode))
       if page.IsSpread {
         return splitIntoVariants(canvas, page)
       }
       return PageRenderResult{Full: canvas}, nil
     }
     ```
   - Variants to support: `thumbnail`, `left`, `right`, `combined`, `full` (see `ttmp/2025-10-11/16-spread-rendering-visualization-guide.md`). Produce file paths + metadata for each.

3. **Persist render outputs.**  
   - Extend `repo.LaidOutPage` to store render metadata if not already available (check `pkg/repo/models/laid_out_page.go`). Add columns for `render_root`, `thumbnail_rel_path`, `updated_at`.  
   - Update migrations if necessary (create new migration file under `pkg/repo/sqlite/migrations`).  
   - Keep files under project-scoped directory (`{data-root}/projects/{id}/pages/{pageID}/`).

4. **Integrate renderer with service layer.**  
   - Modify `pkg/services/pages.go:39-118`. After creating/updating a page, invoke renderer and store result JSON + file paths.  
   - Provide `RenderPage` and `GetPreview` helpers that `pkg/serve/laid_out_pages_routes.go` can call for `/preview` and `/export`.  
   - On updates, delete stale files when recomputing.

5. **Wire HTTP endpoints.**  
   - Replace 501 responses in `pkg/serve/laid_out_pages_routes.go:170-209` with real handlers that stream thumbnails/exports.  
   - Accept query params `variant`, `debug`, `format`; rely on renderer output map.  
   - Add caching headers (Etag/Last-Modified) to avoid recompute storms.

6. **CLI verbs.**  
   - Follow the Glazed command patterns documented in `ttmp/2025-10-10/index.md` (“CLI verbs should use Glazed, directories follow verb groups, one file per verb”).  
   - Create a new verb group directory `cmd/zine-layout/cmds/pages/` if it does not exist. Inside, add `render.go` (single command file) that embeds a Glazed command struct.  
   - Wire middleware layers (parameters, output) exactly as shown in existing verbs, e.g., `cmd/zine-layout/cmds/imagelayout/compute.go`.  
   - Ensure flags follow the standard: positional project/page identifiers, `--data-root` layer injection, and `--file` spec support when applicable.  
   - Update `cmd/zine-layout/main.go` (or the appropriate root command registration file) to include the new verb group.

### Stage B – Frontend Alignment
1. **PageLayoutsTab real data.**  
   - Files: `web/src/views/tabs/PageLayoutsTab.tsx`, `web/src/api.ts`.  
   - Replace mock arrays with RTK Query hooks (`useGetPageTemplatesQuery`, `useGetLaidOutPagesQuery`, etc., see definitions around `web/src/api.ts:340`).  
   - Adjust form state to use `PageLayoutSettings` subsets (remove template-level focus/position controls per `16-todo-notes.md`).  
   - Use server thumbnails: point `<img>` to `/api/laid-out-pages/{id}/preview?variant=thumbnail`.

2. **ImageLayoutsTab cleanup.**  
   - Implement fixes listed in `ttmp/2025-10-11/18-fixes-for-image-layouts-tab-issues.md`: orientation swap handler, always-visible create form, preselect recent template, functional edit drawer, etc.  
   - Remove template-level focus/anchor overrides; keep anchor preset default only.  
   - Ensure preview fallback gracefully handles missing thumbnails until backend renderer lands.

3. **ZineTab integration.**  
   - File: `web/src/views/tabs/ZineTab.tsx`.  
   - Consume `/api/projects/{id}/zines` and `/api/zines/{id}/pages` via hooks.  
   - Provide ordering UI that matches data contract (see `15-... §Zine API`).  
   - Add placeholder button for exporting, disabled until backend export exists.

4. **SequencesTab asset column.**  
   - Confirm implementation from `ttmp/2025-10-11/18-fixes-for-image-layouts-tab-issues.md` (Fix 1) is in place; otherwise port the pseudocode layout (three-column grid).

### Stage C – Imposition & Export
1. **Reuse `pkg/zinelayout`.**  
   - Wrap existing YAML-driven layouts in a new service: `pkg/services/imposition.go`.  
   - Pseudocode:
     ```go
     func (s *ImpositionService) Impose(zineID, presetID string) ([]SheetResult, error) {
       zine, pages := s.repos.Zines.GetWithPages(zineID)
       inputs := loadRenderedPages(pages) // reuse renderer outputs
       layout := zinelayout.LoadPreset(presetID)
       return layout.Render(inputs)
     }
     ```
   - Support preset lookup from disk (`data/presets`) or database table (future).

2. **Export formats.**  
   - Implement PDF assembly in `pkg/export/pdf.go` using `github.com/jung-kurt/gofpdf` or similar. Each `SheetResult` becomes a PDF page.  
   - Provide CLI verb `zine-layout workflow zines export --id ... --preset ...`.

3. **HTTP export endpoints.**  
   - Add `/api/zines/{id}/export` streaming PDF/zip files.  
   - Ensure long-running jobs write status logs (maybe store in `exports` table).

4. **CLI verbs (workflow/zines).**  
   - Per `ttmp/2025-10-10/index.md`, place Glazed command files under `cmd/zine-layout/cmds/workflow/zines/`, keeping one verb per file (for example, `export.go` and `impose.go`).  
   - Reuse the shared layers helpers from existing workflow verbs to expose `--project-id`, `--data-root`, `--preset`, and optional `--file` inputs.  
   - Register the verb group in the workflow root so `go run ./cmd/zine-layout workflow zines ...` lists the new commands.  
   - Mirror how `cmd/zine-layout/cmds/imagelayout/compute.go` handles spec documents when adding YAML support.

### Stage D – Testing & Tooling
1. **Unit tests.**  
   - `pkg/pagelayout/renderer/renderer_test.go`: cover portrait, landscape, spread variants (compare bounding boxes).  
   - `pkg/services/pages_test.go`: verify renderer trigger + metadata persistence.  
   - `pkg/services/imposition_test.go`: ensure preset mapping yields correct sheet counts.

2. **CLI smoke.**  
   - Extend `scripts/run_cli_playbook.py` to cover new verbs (render page, export zine).  
   - Document updates in `ttmp/2025-10-10/11-playbook-for-cli-testing.md`.

3. **Frontend tests.**  
   - Add component tests for PageLayoutsTab flow (creating template → page → thumbnail).  
   - Include regression for orientation toggle and create-form defaults from `ttmp/2025-10-11/18-fixes-for-image-layouts-tab-issues.md`.

---

## 3. Reference Map

| Area | Key Files / Symbols | Notes |
|------|---------------------|-------|
| Page templates API | `pkg/serve/page_templates_routes.go` | Already functional; ensure new settings struct marshals cleanly. |
| Page rendering service | `pkg/services/pages.go`, `pkg/pagelayout/renderer/*` (to add) | Replace `ErrPageRendererNotImplemented`. |
| Image layout source data | `pkg/services/layout.go`, `pkg/imagelayout/engine/engine.go` | Use `LayoutComputation.Result` instead of duplicating math. |
| Spread variants | `ttmp/2025-10-11/16-spread-rendering-visualization-guide.md` | Drive renderer outputs (`left/right/combined/full`). |
| Imposition engine | `pkg/zinelayout/layout.go` et al. | Already handles YAML-defined sheet layouts. |
| Frontend tabs | `web/src/views/tabs/{ImageLayoutsTab,PageLayoutsTab,ZineTab}.tsx` | Hook to live APIs and backend thumbnails. |
| Outstanding UI bugs | `ttmp/2025-10-11/16-todo-notes.md`, `ttmp/2025-10-11/18-fixes-for-image-layouts-tab-issues.md` | Treat as acceptance checklist after wiring. |

---

## 4. Cleanup & Consistency Notes
- Archive or cross-link the superseded procedural details in `ttmp/2025-10-11/13-phase3-and-phase4-implementation-guide.md`; use this roadmap as the canonical guide going forward.
- Keep documentation in sync: whenever you finish a stage, update `ttmp/2025-10-10/05-expansion-plan-for-zine-layout-platform.md` (Phase 3 section) and `ttmp/2025-10-10/07-phase2-backend-and-ui-progress-changelog.md` (add Phase 3 entries) per the cadence outlined in `ttmp/2025-10-10/index.md`.
- Before enabling frontend previews that rely on backend renders, ensure the renderer gracefully handles missing images and returns informative HTTP errors.

---

## 5. Next Developer Checklist
1. Set up fresh data root (`tmp-phase3-dev`) and run:  
   ```bash
   go run ./cmd/zine-layout serve --data-root tmp-phase3-dev --addr :8090
   npm --prefix web run dev
   ```
2. Implement Stage A steps sequentially; verify `/api/laid-out-pages/{id}/preview` returns PNG for both single pages and spreads.
3. Wire PageLayoutsTab to live data; confirm CRUD + previews operate end-to-end.
4. Integrate imposition service and provide a single preset export via CLI and HTTP.
5. Update docs/changelog/index as you complete each milestone.

This condensed roadmap replaces the sprawling instructions in `ttmp/2025-10-11/13-phase3-and-phase4-implementation-guide.md` while preserving essential context. Follow it to deliver backend rendering, frontend wiring, and export workflows with confidence.
