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
