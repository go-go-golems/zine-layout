## 2025-10-12

- Implemented page renderer integration (Stage A/2A):
  - Added cropping from `imagelayout.ViewportResult.SourceRect` in `pkg/pagelayout/renderer/renderer.go`.
  - Extended `RenderContext` to accept layout results.
  - Persisted preview variants (thumbnail/full/combined, left/right for spreads) to `projects/{project}/pages/{page}/`.
  - Implemented `PagesService.RenderPage` to marshal metadata and update `repo.LaidOutPage.ResultJSON`.
  - Updated preview endpoint to stream variant PNGs from disk.

- Smoke-tested via CLI:
 - Added caching headers (ETag/Last-Modified) to preview responses.
 - Added unit tests: renderer crop uses SourceRect; spreads split dims.
 - Added workflow CLI: `workflow laid-out-pages render` to print variant paths.
 - Added HTTP export endpoint: `/api/laid-out-pages/{id}/export?variant=combined` streams PNG.
 - Implemented `ImpositionService` at `pkg/services/imposition.go` to read YAML presets under `data/presets/`, load rendered page images, and generate imposed sheet images via `zinelayout.ZineLayout.CreateOutputImage`.
 - Server wires `ImpositionService` with data root for downstream PDF export work.
  - Created project, uploaded sample image, created image layout and page templates, created laid-out image and page, and fetched preview (`200`).
  - Files written: `thumbnail.png`, `full.png`, `combined.png` (and `left.png`, `right.png` for spreads).

# Changelog – Phase 2 Templates, Layouts & Sequences

## 2025-10-10T00:00Z – Kick-off Notes
- Documented Phase 2 start in `05-expansion-plan-for-zine-layout-platform.md`, linking this log for ongoing updates.
- Confirmed outstanding objectives: database support for layout templates, laid out images, layout sequences; corresponding services, CLI verbs, and UI workflows.

### What Worked
- Repo remains clean with Phase 1 artefacts; good baseline for the extended schema.

### What Didn't Work
- None yet — implementation work has not begun.

### What I Learned
- Adding a direct pointer from the expansion plan should help future contributors find progress context quickly.

### Attention Points
- Maintain incremental updates after each functional milestone (schema, repos, services, API, CLI, UI, tests).

---

## 2025-10-11T13:55Z – Phase 3 Schema Guard
- Added defensive migration logic to drop legacy Phase 2 tables (`laid_out_pages`, `laid_out_page_inputs`, `zine_pages`) when the new `laid_out_image_id` column is missing so the refreshed schema can be applied automatically.
- Verified the helper only runs once—fresh databases continue to use the new schema without data loss, and stale instances now migrate cleanly instead of crashing on startup.

### Attention Points
- Recreate print pages/zines after the drop since legacy data is removed; our workflow CLI can reseed fixtures quickly if needed.

---
## 2025-10-10T00:45Z – Schema & Repository Scaffolding
- Extended SQLite migration (`pkg/repo/sqlite/migrations.go`) with `image_layout_templates`, `laid_out_images`, `layout_sequences`, and item tables to support Phase 2 entities.
- Added repository interfaces and structs for layouts/laid-out assets/sequence ordering (`pkg/repo/types.go`), plus concrete SQLite adapters for each new entity.
- Wired repository constructors into `pkg/repo/sqlite/sqlite.go`; ensured helper conversion utilities handle nullable template/project fields.

### What Worked
- Reused existing ID generation and null-string helpers to keep the new repos compact and consistent with Phase 1 style.
- Incremental gofmt pass kept diffs readable despite multiple new files.

### What Didn't Work
- Nothing blocking yet; service layer still missing to exercise the repositories.

### What I Learned
- Keeping layout sequences gap-free simplifies the schema — no need for nullable IDs once laid-out images are mandatory.

### Attention Points
- Build service orchestration (`pkg/services/layout.go`) next so HTTP handlers can remain thin wrappers.

---

## 2025-10-10T02:15Z – Layout Service & HTTP Restructure
- Introduced `pkg/services/layout.go` to encapsulate template application, override merging, and placement recompute logic (`simple.ComputePlacement` + trace capture).
- Expanded REST surface: project/global template endpoints, laid-out image CRUD with preview, layout sequence CRUD/item operations (all via new route files under `pkg/serve/`).
- Refactored the monolithic `server.go` into focused modules (`projects_routes.go`, `image_sequences_routes.go`, `layout_templates_routes.go`, etc.) and centralised response DTOs in `pkg/serve/types.go`.
- Added preview endpoint plumbing for laid-out images; export remains stubbed pending renderer hookup.

### What Worked
- Service layer keeps handlers thin—HTTP routes now orchestrate repositories/layout logic without duplicating JSON merging.
- Splitting handlers by concern made the import surface cleaner and improved `go build` output readability.
- Manual `go build ./...` passes; preview endpoint returns structured placement payloads ready for UI wiring.

### What Didn't Work
- Export endpoint currently returns `501` because renderer integration is deferred; callers must handle the placeholder for now.

### What I Learned
- Co-locating response adapters in `pkg/serve/types.go` avoids circular imports while keeping JSON shape changes discoverable.
- Having the layout service normalise overrides JSON prevents subtle drift between API payloads and stored state.

### Attention Points
- Follow up with actual image export implementation once rendering modes (tasks 2.7–2.9) land.
- Wire the new endpoints into CLI/UI layers before adding further behaviour to ensure contracts are exercised end-to-end.

---

## 2025-10-10T03:30Z – CLI Coverage for Templates & Layouts
- Added new Glazed verb groups covering image layout templates, laid-out images, and layout sequences with list/get/create/update/delete flows plus sequence item management (`cmd/zine-layout/cmds/api/*`).
- Exposed laid-out image preview command returning persisted placement traces; export command deferred until server endpoint is ready.
- Registered new verb groups in the API command root so `zine-layout api` now exposes all Phase 2 entities alongside existing project/sequence verbs.

### What Worked
- Reused common HTTP helpers to keep each verb implementation small and consistent with Phase 1 patterns.
- Manual `go build ./...` confirms CLI wiring compiles across all command sets.

### What Didn't Work
- No batch commands yet for `apply-template-to-sequence` or template previews—awaiting dedicated server endpoints before wiring.

### What I Learned
- Keeping command factories per verb group makes registration trivial; adding new groups was a drop-in change to `commands.go`.

### Attention Points
- Add the batch CLI verbs once supporting APIs exist (plan item 2.11).
- Consider shared helper package if future commands start duplicating HTTP helpers across directories.

---

## 2025-10-10T04:30Z – UI Scaffolding for Templates & Layout Sequences
- Updated RTK query layer (`web/src/api.ts`) with Phase 2 entities (layout templates, laid-out images, layout sequences) and verified with `npm run build`.
- Added `LayoutTemplateManager`, `LaidOutImageViewer`, and `LayoutSequenceEditor` views, integrating them into the project detail screen for an end-to-end Phase 2 workflow.
- Enabled preview rendering in the UI by hitting the new `/preview` endpoint; forms currently accept JSON overrides for rapid experimentation.

### What Worked
- Isolating each manager view kept state management focused and reduced churn inside `ProjectDetail`.
- Vite production build succeeded on the first pass, indicating API contracts line up with the new hooks.

### What Didn't Work
- Reordering uses simple buttons instead of drag-and-drop; usability improvements remain on the backlog.
- JSON textareas are functional but not yet user-friendly compared to dedicated form controls.

### What I Learned
- Surfacing preview payloads adjacent to edit forms shortens feedback loops when tweaking template overrides.

### Attention Points
- Add richer form validation and drag-and-drop ergonomics in a future iteration.
- Display global vs project template scope more prominently to prevent accidental edits.

---

## 2025-10-10T06:15Z – Image Layout Engine Refresh + CLI Sanity Tool
- Replaced the legacy `pkg/spread` module with a focused `pkg/imagelayout` stack (`types`, `defaults`, `engine`) that now supports crop mode, fit-to-width/height, anchor presets, and focus-point alignment—matching the behaviours in `02-image-resizer-code.tsx`.
- Updated `LayoutService` and API payloads to persist the new computation shape (`settings`, `result`, `trace`) and added unit coverage in `pkg/imagelayout/engine/engine_test.go` for contain/cover, crop, fit, anchor, and focus scenarios.
- Introduced a local helper command `zine-layout imagelayout compute` that accepts either YAML specs or CLI flags, executes the engine directly, and prints the resulting computation JSON for quick debugging.
- Refreshed docs (`05-expansion-plan`, `08-layout-template-dsl-guide`, `09-system-spec`) to document the new settings fields, CLI workflow, and outstanding roadmap items.

### What Worked
- The new engine mirrors the TSX resizer math for page/crop/fit modes, keeping UI previews and backend results in sync.
- Reusing the service layer meant the API/CLI plumbing only needed minimal changes—existing verbs now emit the richer trace payload without breaking callers.
- The standalone CLI compute command made it easy to validate YAML templates locally before wiring them through the server.

### What Didn't Work
- Dual-page/spread handling remains unimplemented; templates still can’t express gutters or panel splitting.
- No renderer helpers yet, so export endpoints and CLI verbs still return `501` placeholders.

### What I Learned
- Normalising anchor presets and focus points in the engine drastically simplifies the DSL compiler requirements—TSX controls translate cleanly into `ViewportSettings`.
- Having a local CLI shortcut encourages rapid iteration on DSL changes without spinning up the full API stack.

### Attention Points
- Backfill service-level tests (`pkg/services/layout_test.go`) now that the engine is stable, and add integration coverage for template→laid-out image workflows.
- Implement renderer helpers (`pkg/imagelayout/renderer`) so export endpoints can graduate from stubs.
- Extend the DSL compiler to surface validation errors for the new fit/crop/focus fields, keeping UI feedback tight.

---

## 2025-10-10T07:45Z – Frontend API Typings & Global Template Hooks
- Reworked `web/src/api.ts` to mirror the Go-side `imagelayout` structs: added typed viewport settings/result/trace models, aliased `SpreadSettings`, and tightened laid-out overrides to `Partial<ImageLayoutViewportSettings>`.
- Added RTK Query endpoints for global image layout templates (`GET/POST /api/image-layout-templates`) plus single-template fetches, alongside new hooks for upcoming UI work.
- Renamed cache tag identifiers to the explicit “image layout” terminology so future page/zine layout DSLs can coexist without collisions.

### What Worked
- Porting the backend shape wholesale removed guesswork—existing Redux slices consume the richer data without intermediate casting.
- Tag renames kept cache invalidation predictable even with the added global listings.

### What Didn't Work
- Legacy React views still import the older hook names; compatibility aliases remain for now, so the UI workstream must fold in the new hooks to benefit fully.

### What I Learned
- Maintaining a `SpreadSettings` alias eased the transition while signalling the new terminology to downstream consumers.

### Attention Points
- Coordinate with the frontend implementer to adopt the new hooks and retire the temporary aliases.
- Backfill automated checks (typecheck/lint) once the UI migrates, ensuring the stricter typings catch regressions early.

## 2025-10-10T08:30Z – Removed Legacy Book Spread Frontend State
- Deleted the Redux `bookSpread` slice, supporting helpers, and the old spread designer components; kept only the assets panel (now `components/ProjectAssetsPanel.tsx`).
- Simplified the store to RTK Query plus the lightweight UI toast slice, reinforcing the API-driven data flow.
- Dropped the unused `SpreadSettings` alias from the frontend API typings to avoid reintroducing the retired surface.

### What Worked
- Pruning the unused slice removed Redux-specific TypeScript noise and clarified that image layout state now comes from backend computations.

### What Didn't Work
- UI scenes still need TSX refactors to fill the gap left by the deleted panels; awaiting the dedicated frontend pass.

### What I Learned
- Keeping asset uploads independent of Redux made the extraction painless—future components should follow similar patterns.

### Attention Points
- Partner with the UI owner to rebuild the designer views using the new hooks.
- Re-run `pnpm typecheck` once the TSX overhaul lands to clear the remaining component prop errors.

---

## 2025-10-11T04:07Z – Page/Zine Persistence & Workflow CLI
- Extended the SQLite schema with `page_templates`, `laid_out_pages`, `laid_out_page_inputs`, `zines`, and `zine_pages`, wiring corresponding repo adapters and service layers (`pkg/services/pages.go`, `pkg/services/zines.go`).
- Introduced the `zine-layout workflow` Glazed command group for direct testing: `page-templates`, `laid-out-pages`, and `zines` verbs now operate via repositories/services without needing REST.
- Added Go service helpers for laid-out page creation/reordering and zine page sequencing; `RenderPage` currently returns `ErrPageRendererNotImplemented` pending renderer hookup.
- Verified compilation with `go test ./...`.

### What Worked
- Reusing the repository patterns from Phase 2 kept the new SQLite adapters concise and consistent.
- Glazed commands made it straightforward to pipe results into other tooling while bypassing the unfinished HTTP layer.

### What Didn't Work
- The rendering path remains stubbed—no PNG export yet, so CLI consumers must handle the `ErrPageRendererNotImplemented` sentinel.

### What I Learned
- Keeping the workflow CLI separate from the API commands provides a safe playground for future backend changes without blocking on REST contract decisions.

### Attention Points
- Implement the actual renderer/export pipeline so `RenderPage` and future zine exports can return artefacts.
- Mirror the new CLI capabilities in forthcoming REST handlers and frontend views.

---

## 2025-10-11T04:25Z – Workflow CLI Smoke Tests & Imagelayout Check
- Exercised the new workflow verbs against a fresh `tmp-workflow-test` data root:
  - Created a global page template (`ptpl-…`), laid-out page (`lpg-…`), and single-page zine (`zne-…`), confirming list/get/set/delete flows across page and zine services.
  - Observed FK requirements (manual `projects` seed via sqlite3) and ensured schema migrations no longer drop tables on startup.
- Ran `go run ./cmd/zine-layout imagelayout compute` with explicit flags—placement math returns expected crop/cover results and diagnostic trace.
- Noticed that `--spec` alone still yields “canvas dimensions must be positive”; YAML deserialization isn’t populating the settings struct yet (likely missing field tags). Added this to the follow-up list.

### What Worked
- Workflow commands produce tidy tabular/JSON output and provide a practical harness while REST/UI remain unfinished.
- Layout engine continues to match expectations for cover/contain scenarios when inputs are provided via flags.

### What Didn't Work
- Imagelayout spec parsing ignores YAML settings, forcing manual flag overrides.

### What I Learned
- Having repo-level CLI verbs is invaluable for quick regression checks; keeping a scratch data root avoids polluting main workspace assets.

### Attention Points
- Patch the imagelayout spec loader with proper YAML tags so CLI specs work out-of-the-box.
- Consider adding workflow verbs for seeding projects/assets to avoid manual sqlite3 scripting during tests.

---

## 2025-10-11T05:10Z – REST API for Page Templates, Print Pages & Zines
- Added dedicated route files in `pkg/serve` for page templates, laid-out pages, and zines, wired into `server.go` with new `PagesService` and `ZinesService` instances.
- Responses now expose normalized JSON payloads (`page_templates`, `laid_out_pages`, `zines`) with helper structs in `pkg/serve/types.go`.
- Extended `web/src/api.ts` with strongly-typed RTK Query endpoints covering all Phase 3 entities, including list/detail mutations and tag invalidation semantics.

### What Worked
- Reusing the Phase 2 routing pattern (one file per entity) made it easy to bolt on the new handlers without bloating `server.go`.
- RTK Query’s tag system let us keep cache invalidation predictable across project-scoped and global template lists.

### What Didn't Work
- Preview/export endpoints remain stubs; `/api/laid-out-pages/{id}/preview` still surfaces `ErrPageRendererNotImplemented`.

### What I Learned
- Surfacing shared results (global + project templates) benefits from tagging both scopes so UI caches refresh automatically when globals change.

### Attention Points
- Implement the renderer/export pipeline so the preview/export endpoints can return binary payloads.
- Once the frontend views land, ensure they use the new hooks instead of the deprecated store slices.

---
