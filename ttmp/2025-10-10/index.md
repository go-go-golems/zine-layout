# Zine Layout Platform – Codebase Index & Orientation

Welcome! This guide is handed to every task runner so you can quickly locate the right files, documents, and commands in the repository. Keep it close while you work—and update it whenever major structure changes land.

---

## 1. Repository Map

| Path | Purpose |
|------|---------|
| `cmd/zine-layout/` | All end-user CLIs. Subfolders under `cmds/` contain Glazed command groups (`api`, `workflow`, `imagelayout`, `serve`, `render`). |
| `pkg/repo/` | Database access layer. `sqlite/` holds migrations and repository implementations; `types.go` defines persistent entities. |
| `pkg/services/` | Application logic (layout computation, page/zine orchestration). Services wrap repositories and encapsulate workflows. |
| `pkg/imagelayout/` | Layout engine types, defaults, and math. Mirrors behaviours documented in the resizing algorithm notes. |
| `pkg/serve/` | HTTP server configuration and REST handlers. Routes are split per entity (projects, image sequences, layouts, etc.). |
| `web/` | React + RTK Query frontend (`src/api.ts`, `src/views/*`, `src/components/*`). `package.json`/`pnpm-lock.yaml` define the dev toolchain. |
| `data/` | Default storage for projects/uploads when the server runs locally. Safe to wipe during testing. |
| `ttmp/2025-10-10/` | Living documentation for the current re-architecture (see section 3). |

> Tip: `go run ./cmd/zine-layout --help` lists every top-level CLI. Combine with `GLAZED_OUTPUT=json` for machine-friendly output.

---

## 2. Key Code Paths by Task

- **Database changes**  
  - Schema: `pkg/repo/sqlite/migrations.go`  
  - Repositories: `pkg/repo/sqlite/*.go` (one file per entity)  
  - Types/interfaces: `pkg/repo/types.go`

- **Business logic / workflows**  
  - Layout computation: `pkg/services/layout.go`  
  - Page & zine orchestration: `pkg/services/pages.go`, `pkg/services/zines.go`  
  - Imagelayout math: `pkg/imagelayout/engine/*`

- **HTTP / API**  
  - Server bootstrap: `pkg/serve/server.go`  
  - Route handlers: `pkg/serve/*_routes.go`  
  - JSON response helpers: `pkg/serve/types.go`

- **CLI**  
  - All verbs use the **Glazed** framework (see `glazed/pkg/doc/tutorials/...` for patterns).  
  - Directory structure mirrors verb groups: each group lives in `cmd/zine-layout/cmds/<verb-group>/`, with **one Go file per verb** (e.g., `image_sequences/create.go`).  
  - API client verbs (`rest`-oriented): `cmd/zine-layout/cmds/api/*`  
  - Workflow verbs (direct DB/service access): `cmd/zine-layout/cmds/workflow/*`  
  - Imagelayout tooling: `cmd/zine-layout/cmds/imagelayout/command.go`

- **Frontend**  
  - API layer & RTK hooks: `web/src/api.ts`  
  - Project dashboard: `web/src/views/ProjectDetail.tsx`  
  - Template/layout editors: `web/src/views/LayoutTemplateManager.tsx`, `web/src/views/LaidOutImageViewer.tsx`, `web/src/views/LayoutSequenceEditor.tsx`  
  - Shared UI primitives: `web/src/components/ui/`

---

## 3. Essential Documentation (ttmp/2025-10-10)

Keep these updated as you work:

| File | Why it matters |
|------|----------------|
| `index.md` *(this file)* | Master index / onboarding sheet. Update when directories or docs change. |
| `05-expansion-plan-for-zine-layout-platform.md` | Implementation checklist per phase with cross-links to specs/changelog. Mark progress and add instructions for future steps. |
| `07-phase2-backend-and-ui-progress-changelog.md` | Day-by-day engineering log. Append an entry after **every** significant task (schema change, CLI test, bugfix). |
| `09-system-specification-after-phase1-and-phase2.md` | Canonical architecture/spec. Update immediately when schemas, services, or CLI inventories change. |
| `../2025-10-11/19-phase3-focused-roadmap.md` | Up-to-date Phase 3 plan (Stages A–D). |
| `../2025-10-11/20-imposition-and-pdf-export-context.md` | How to finish Stage C: imposition → PDF export (symbols, commands). |
| `04-report-on-the-current-codebase-and-how-it-can-be-used-for-the-zine-layout-software.md` | Current state analysis prior to Phase 2; useful historical context. |
| `01-algorithm-for-resizing.md`, `02-image-resizer-code.tsx`, `03-template-resize-dsl.md` | Algorithmic references for imagelayout behaviour and DSL expectations. |
| `06-phase1-backend-and-cli-progress-changelog.md` | Phase 1 log (legacy but informative). |
| `08-layout-template-dsl-go-integration-guide.md` | How to translate the TSX DSL into Go structures. |
| `10-ui-design-for-the-zine-photo-layout-software.md` | Figma/UI guidelines for the ongoing frontend refresh. |

**Update cadence**  
- Every code change → log it in `07-…-changelog.md`.  
- Architecture/API shifts → mirror them in `09-…-system-specification.md`.  
- Whenever a checklist item moves forward → tick it in `05-…-expansion-plan.md` and reference the supporting changelog/spec sections.

---

## 4. Getting Started Quickly

1. **Install toolchains**  
   - Go ≥ 1.22, Node ≥ 20 (pnpm 10.x).  
   - `pnpm install --frozen-lockfile` inside `web/`.

2. **Run the server & web app**  
   ```bash
   go run ./cmd/zine-layout serve --data-root ./data --root web/dist
   pnpm --dir web dev
   ```

3. **CLI smoke tests**  
   - API verbs: `go run ./cmd/zine-layout api projects list`  
   - Workflow verbs: `go run ./cmd/zine-layout workflow page-templates list --data-root ./tmp-workflow`  
   - Imagelayout math: `go run ./cmd/zine-layout imagelayout compute --source-width 4000 --source-height 3000 --mode page --paper-width-in 8.5 --paper-height-in 11 --dpi 300`

4. **Testing**  
   - Backend: `go test ./...`  
   - Frontend lint/typecheck: `pnpm --dir web lint`, `pnpm --dir web typecheck`

---

## 5. Conventions & Best Practices

- **SQLite migrations** are idempotent—never drop tables in `schemaSQL`; apply additive migrations only.
- **Workflow CLI** is safe for local experiments but bypasses REST/validation; mirror important flows in the API before shipping.
- **Documentation** is part of the deliverable. If you touch major functionality without updating `05`, `07`, and `09`, the next contributor will miss vital context.
- **Data roots** (`--data-root`) are disposable. Create per-task directories (e.g., `tmp-workflow-test`) to keep experiments isolated.

---

Need something that isn’t listed here? Add it! This index should evolve with the codebase so new contributors can stay productive from day one.
