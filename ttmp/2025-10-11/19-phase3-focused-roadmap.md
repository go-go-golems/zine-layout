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
   - Implemented at `pkg/pagelayout/renderer/renderer.go`.  
   - Inputs: `PageLayoutSettings`, `imagelayout.ViewportResult` (from `LayoutComputation.Result`), plus source image.  
   - Done: cropping via `SourceRect`, scaling into page `ContentRectPx`, spread split (left/right), optional borders.  
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
   - Done: metadata in `LaidOutPage.ResultJSON` with variant rel paths; files under `{data-root}/projects/{id}/pages/{pageID}/`.

4. **Integrate renderer with service layer.**  
   - Done: `PagesService.RenderPage` generates files and persists metadata; server injects data root.  
   - Preview endpoint calls render and streams files with ETag/Last-Modified.  
   - TODO: garbage collect stale files on updates.

5. **Wire HTTP endpoints.**  
   - Done: `/api/laid-out-pages/{id}/preview?variant=...` streams PNG; `/api/laid-out-pages/{id}/export?variant=...` streams PNG for download.  
   - Accepts `variant` param; uses stored variant paths.  
   - Caching headers added to preview responses.

6. **CLI verbs.**  
   - Follow the Glazed command patterns documented in `ttmp/2025-10-10/index.md` (“CLI verbs should use Glazed, directories follow verb groups, one file per verb”).  
   - Added `workflow laid-out-pages render` to print variant paths for a page.

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
   - Implement `pkg/services/imposition.go`: load preset YAML (from `data/presets`), load rendered page PNGs, map to inputs, call `ZineLayout.CreateOutputImage` for each sheet.  
   - Return sheet images + metadata (sizes, positions) for downstream export.

2. **Export formats.**  
   - Next: implement `pkg/export/pdf.go` to generate a single PDF: each imposed sheet → PDF page.  
   - Provide CLI: `go run ./cmd/zine-layout workflow zines export --data-root ... --zine-id ... --preset ...`.

3. **HTTP export endpoints.**  
   - Add `/api/zines/{id}/export?preset=...` streaming the PDF.  
   - Consider job logging for long-running presets (future).

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
