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
