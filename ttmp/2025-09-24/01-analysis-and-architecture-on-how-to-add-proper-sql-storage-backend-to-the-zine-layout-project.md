## Purpose and scope

We will introduce a proper SQLite storage backend for zine-layout to persist projects, uploaded images (assets), book pages, and book spreads (two facing pages). The goal is to support:
- A first class uploaded image carousel within the BookSpreadDesigner
- Persisting layouts for pages and spreads as structured records
- Keeping compatibility with existing data-root disk layout for assets while adding relational data for project structure
- Repository interface pattern for testability and swappable storage

This document details requirements and designs for:
- Model design
- SQL schema
- Repository interfaces
- REST endpoints
- RTK Query additions
- React integration

## Current state summary

Server uses filesystem semantics over a data-root:
- Projects are directories with project.json, and images/ content
- Images are stored as files under projects/{id}/images
- Renders are stored in project subdirs
- API provides project list/create, per-project read/update/delete, image list/upload/reorder/delete, preset application, validation, and render/preview endpoints
- The web uses RTK Query in web/src/api.ts and presents a BookSpreadDesigner with upload section and preview flow
- All spread computation/rendering paths have been consolidated around the Simple algorithm; Sonnet code was removed in `pkg/spread/sonnet`. Simple-specific helpers now live in:
  - `pkg/spread/simple/inputs.go` (builds algorithm inputs straight from `spread.Settings`)
  - `pkg/spread/yaml.go` (Simple YAML schema v0.2, parsing and whole-book export)
  - `pkg/spread/defaults.go` (canonical defaults used when exporting data)
- The server exposes Simple-only REST handlers in `pkg/serve/server.go`:
  - `POST /api/v1/compute`, `/api/v1/preview`, `/api/v1/render` call `simple.InputsFromSettings` directly
  - `POST /api/v1/yaml` returns a Simple YAML snippet built by `spread.BuildSimpleYAML`
  - `POST /api/v1/yaml/render` parses Simple YAML via `spread.ParseSimpleBookYAML`
  - `GET /api/projects/{id}/yaml/book` assembles a full-book YAML export with `Server.buildProjectBookYAML`
- The web client (`web/src/api.ts`) gained RTK Query endpoints for the YAML features, including `exportBookYaml`, and `BookSpreadDesigner.tsx` now offers an “Export Book YAML” panel that hits the new endpoint.

Limitations:
- No relational persistence for image metadata beyond current directory scanning
- No persisted spreads/pages entities; existing compute endpoints generate previews on the fly
- Reordering is stored in project.json only; not queryable with filters

## Data model (domain)

Concepts:
- Project: container for a book project
- Asset: uploaded image file belonging to a project
- Page: a single page that references an asset and a layout result (placement and export parameters)
- Spread: two facing pages (left and right) for a project; references assets for left/right and layout results
- Preset: existing yaml-based design presets remain as files; we can store references to presets used per project

Core attributes:
- Project: id, name, createdAt, updatedAt, presetId (optional), coverAssetId (optional)
- Asset: id, projectId, filename, relPath, width, height, contentType, bytes, createdAt, sortIndex
- Page: id, projectId, pageNumber, assetId (nullable), settings JSON, result JSON, createdAt, updatedAt
- Spread: id, projectId, spreadNumber, leftPageId (nullable), rightPageId (nullable), settings JSON, result JSON, createdAt, updatedAt

Notes:
- settings JSON stores the spread settings used to compute the layout (paper size, dpi, margins, etc.)
- result JSON stores the computed placement results produced by the algorithm, and optionally render info
- We keep sortIndex on assets to implement the carousel order

## SQL schema (SQLite)

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  preset_id TEXT,
  cover_asset_id TEXT,
  FOREIGN KEY (cover_asset_id) REFERENCES assets(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS assets (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  filename TEXT NOT NULL,
  rel_path TEXT NOT NULL,
  content_type TEXT,
  bytes INTEGER NOT NULL,
  width INTEGER NOT NULL,
  height INTEGER NOT NULL,
  sort_index INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assets_project ON assets(project_id, sort_index);

CREATE TABLE IF NOT EXISTS pages (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  page_number INTEGER NOT NULL,
  asset_id TEXT,
  settings_json TEXT NOT NULL,
  result_json TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE(project_id, page_number),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS spreads (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  spread_number INTEGER NOT NULL,
  left_page_id TEXT,
  right_page_id TEXT,
  settings_json TEXT NOT NULL,
  result_json TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE(project_id, spread_number),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY (left_page_id) REFERENCES pages(id) ON DELETE SET NULL,
  FOREIGN KEY (right_page_id) REFERENCES pages(id) ON DELETE SET NULL
);
```

IDs: use ULIDs or UUIDv4 strings generated in Go. The rest of the code already uses string IDs like prj-..., we can keep that for projects and use new ID scheme for others.

Files on disk:
- Keep storing raw image files under data-root uploads or project images directories. assets.rel_path stores the relative file path for retrieval. The API for /uploads can remain, and we add a project asset upload that also writes a DB row.
- When persisting to SQL, plan for dual-source truth during the migration: render/export endpoints currently derive layout settings from request payloads and Simple YAML. New tables must capture the same `spread.Settings` (JSON) the server consumes so the Simple pathway remains stateless. See `pkg/serve/server.go:701` for how settings/results are bundled today when producing YAML previews.

## Repository interfaces

Define small interfaces under zine-layout/pkg/repo to encapsulate DB operations. Use context in all methods.

```go
package repo

import "context"

type Project struct {
  ID string
  Name string
  CreatedAt int64
  UpdatedAt int64
  PresetID *string
  CoverAssetID *string
}

type Asset struct {
  ID string
  ProjectID string
  Filename string
  RelPath string
  ContentType string
  Bytes int64
  Width int
  Height int
  SortIndex int
  CreatedAt int64
}

type Page struct {
  ID string
  ProjectID string
  PageNumber int
  AssetID *string
  SettingsJSON string
  ResultJSON *string
  CreatedAt int64
  UpdatedAt int64
}

type Spread struct {
  ID string
  ProjectID string
  SpreadNumber int
  LeftPageID *string
  RightPageID *string
  SettingsJSON string
  ResultJSON *string
  CreatedAt int64
  UpdatedAt int64
}

type ProjectRepository interface {
  Create(ctx context.Context, p *Project) error
  Update(ctx context.Context, p *Project) error
  Get(ctx context.Context, id string) (*Project, error)
  List(ctx context.Context) ([]*Project, error)
  Delete(ctx context.Context, id string) error
}

type AssetRepository interface {
  Create(ctx context.Context, a *Asset) error
  Get(ctx context.Context, id string) (*Asset, error)
  ListByProject(ctx context.Context, projectID string) ([]*Asset, error)
  UpdateOrder(ctx context.Context, projectID string, orderedIDs []string) error
  Delete(ctx context.Context, id string) error
}

type PageRepository interface {
  Upsert(ctx context.Context, p *Page) error
  GetByNumber(ctx context.Context, projectID string, pageNumber int) (*Page, error)
  List(ctx context.Context, projectID string) ([]*Page, error)
  Delete(ctx context.Context, id string) error
}

type SpreadRepository interface {
  Upsert(ctx context.Context, s *Spread) error
  GetByNumber(ctx context.Context, projectID string, spreadNumber int) (*Spread, error)
  List(ctx context.Context, projectID string) ([]*Spread, error)
  Delete(ctx context.Context, id string) error
}
```

When materializing `SettingsJSON` and `ResultJSON`, follow the structures already emitted in the Simple pipeline:
- `pkg/spread/simple/inputs.go` defines the fields the preview/render endpoints expect.
- `pkg/spread/simple/algorithm.go` and `pkg/spread/simple/render.go` describe the `Result` layout you should persist.
- `pkg/spread/yaml.go` shows how these shapes serialize to YAML; keeping JSON payloads parallel ensures `/api/v1/yaml`, `/api/v1/yaml/render`, and `/api/projects/{id}/yaml/book` can round-trip between SQL rows and exported documents.

Implementation: provide pkg/repo/sqlite with concrete types and a NewSQLiteRepositories(db *sql.DB) factory returning a struct with these repos. Use BEGIN IMMEDIATE transactions where appropriate (reorder, batch upserts).

Implementation snapshot:
- `pkg/repo/types.go` contains the shared structs and interfaces.
- `pkg/repo/sqlite` wires SQLite via `modernc.org/sqlite`, applies inline migrations (see `migrations.go`), and exposes CRUD via `NewRepositories`.
- The HTTP server (`pkg/serve/server.go`) opens `data/zine-layout.db` on startup, seeds missing rows from the legacy filesystem, and keeps both SQLite and the existing `project.json` files in sync.
- New/updated endpoints:
  - `GET/POST /api/projects/{id}/images`, `POST /api/projects/{id}/images/reorder`, `DELETE /api/projects/{id}/images/{imageId}` (assets backed by `repo.Assets`).
  - `GET/PUT/DELETE /api/projects/{id}/pages/{pageNumber}` and `/api/projects/{id}/spreads/{spreadNumber}` persist `spread.Settings` / `simple.Result` blobs as JSON strings.
- Frontend RTK Query stubs (`web/src/api.ts`) mirror the new pages/spreads endpoints so future UI work can hook into the DB-backed data model.

## REST API design

We kept the established route surface and layered persistence beneath it:

- Images (`/api/projects/{id}/images`) continue to return `{images, order}` so existing components stay untouched, but the handlers now read/write via the asset repository and sync disk + DB state.
- Pages use the new `/api/projects/{id}/pages` collection with numeric keys. Requests accept `{ settings, asset_id?, result? }` bodies and respond with serialized `spread.Settings` & `simple.Result` documents.
- Spreads follow the same pattern under `/api/projects/{id}/spreads`, linking to pages by number instead of IDs to match the Simple workflow.

Preview/render endpoints (`/api/v1/*`) remain payload based today. Future enhancements can add variants that fetch stored page/spread settings before invoking the Simple algorithm.

Errors surface as standard HTTP responses: 400 for malformed JSON, 404 for missing records, and 409 for reorder mismatches during asset updates.

## RTK Query additions

`web/src/api.ts` now defines the wire types that match the server payloads:
```ts
export interface ImageItem { id: string; name: string; width: number; height: number }
export interface PersistedPage {
  page_number: number;
  asset_id?: string;
  settings: SpreadSettings;
  result?: any;
  created_at: string;
  updated_at: string;
}
export interface PersistedSpread {
  spread_number: number;
  left_page_number?: number;
  right_page_number?: number;
  settings: SpreadSettings;
  result?: any;
  created_at: string;
  updated_at: string;
}
```

Endpoints align closely with the server routes:
```ts
getImages: b.query<{ images: ImageItem[]; order: string[] }, { id: string }>({
  query: ({ id }) => `/projects/${id}/images`,
  providesTags: ['Image'],
}),
uploadImages: b.mutation<{ images: ImageItem[] }, { id: string; files: FileList | File[] }>({
  query: ({ id, files }) => { /* multi-upload */ },
  invalidatesTags: ['Image'],
}),
getPages: b.query<{ pages: PersistedPage[] }, { id: string }>({
  query: ({ id }) => `/projects/${id}/pages`,
}),
putPage: b.mutation<{ page: PersistedPage }, { id: string; pageNumber: number; page: { asset_id?: string; settings: SpreadSettings; result?: any } }>({
  query: ({ id, pageNumber, page }) => ({ url: `/projects/${id}/pages/${pageNumber}`, method: 'PUT', body: page }),
}),
getSpreads: b.query<{ spreads: PersistedSpread[] }, { id: string }>({
  query: ({ id }) => `/projects/${id}/spreads`,
}),
putSpread: b.mutation<{ spread: PersistedSpread }, { id: string; spreadNumber: number; spread: { left_page_number?: number; right_page_number?: number; settings: SpreadSettings; result?: any } }>({
  query: ({ id, spreadNumber, spread }) => ({ url: `/projects/${id}/spreads/${spreadNumber}`, method: 'PUT', body: spread }),
}),
```

Everything shares the existing `Project`/`Image` tag types, so cache invalidation keeps behaving the same while we phase in new UI consumers.

## React integration plan

Image carousel:
- Keep using `useGetImagesQuery` but surface the persisted order/sort index in component state so drag-and-drop can call `useReorderImagesMutation`.
- After upload, refresh the carousel from SQL to verify the backfill path (`ensureProjectAssets`) keeps legacy projects consistent.

Pages management:
- Add UI panels to load `useGetPagesQuery`, edit `SpreadSettings`, and call `usePutPageMutation` so per-page tweaks survive reloads.
- Offer a “preview from page” action that pipes the stored settings/result into the Simple compute endpoint before writing the updated layout back to SQL.

Spreads management:
- Render spreads from `useGetSpreadsQuery`, with controls to map left/right page numbers and persist results via `usePutSpreadMutation`.
- Once spread edits are persisted, hook `exportBookYaml` so it can derive defaults and overrides from the SQL-backed data rather than solely from `spec.yaml`.

Routing and state:
- Extend the Redux slice with `selectedPageNumber` / `selectedSpreadNumber` values that reference the persisted records.
- Update selectors powering previews/renders to read from the chosen record when available, falling back to in-flight form state otherwise.

### React UI wiring snapshot (2025-09-24)

- `web/src/views/BookSpreadDesigner.tsx` orchestrates the new store-first workflow: it loads assets via `useGetImagesQuery`, selects thumbnails with `setImage`, and exposes persistence controls through `LayoutPersistencePanel`.
- Asset management lives in `web/src/components/bookSpread/ProjectAssetsPanel.tsx`, which wraps `useUploadImagesMutation`, `useDeleteImageMutation`, and `useReorderImagesMutation` to keep the SQL-backed carousel authoritative.
- Page/spread CRUD hooks are consumed inside `web/src/components/bookSpread/LayoutPersistencePanel.tsx`, translating UI form fields into the `SpreadSettings` JSON the server persists.
- Legacy `ImageTray` (used by `web/src/views/ProjectDetail.tsx`) still works against the same endpoints; refactoring it to reuse the new ProjectAssets panel would give a consistent UX across entry points.
- All of the new panels rely on Redux actions from `web/src/store/bookSpreadSlice.ts` (`setImage`, `loadSettings`, `setPaperSize`, `setMargins`, `setOrientation`) to hydrate designer state from persisted records.

## Follow-up roadmap

- Remove the remaining filesystem fallbacks once the React layer reads persisted pages/spreads by default.
- Add a nightly task or CLI to materialize `spec.yaml` from SQL, keeping YAML exports aligned with the DB representation.
- Introduce repository-level tests with a temp SQLite database (look at `modernc.org/sqlite` in-memory DSN) to lock down migrations and JSON encoding.
- Consolidate project views by migrating `web/src/views/ProjectDetail.tsx` to the shared asset/persistence panels so uploads and page saves behave consistently across the app.

## Implementation status

- [x] Repository interfaces plus a SQLite-backed implementation live under `pkg/repo` and `pkg/repo/sqlite` (see `pkg/repo/sqlite/sqlite.go:22`).
- [x] Server startup now opens `data/zine-layout.db`, applies migrations, and seeds missing state from disk (`pkg/serve/server.go:706`).
- [x] Legacy image upload/list/reorder routes persist via the asset repository while continuing to serve the old response shape (`pkg/serve/server.go:174`).
- [x] Page and spread CRUD endpoints are backed directly by SQL with JSON encoding helpers mirroring the Simple pipeline (`pkg/serve/server.go:344`).
- [x] RTK Query exposes typed hooks for the new page/spread resources (`web/src/api.ts:233`).
- [x] React components now load and persist assets/pages/spreads through RTK Query endpoints (`web/src/views/BookSpreadDesigner.tsx`).
- [x] Basic HTTP integration coverage exercises the pages/spreads lifecycle through the REST API (`pkg/serve/server_rest_test.go`).
- [ ] Repository-focused unit tests are still pending; consider lightweight sqlite-in-memory suites for CRUD edge cases.

Appendix: Using the repository interface pattern allows swapping sqlite for filesystem or future backends and enables unit testing by mocking repositories.

## Repository layer deep dive

The new persistence package cleanly separates domain structs from storage details:

- Structs and narrow interfaces live in `pkg/repo/types.go:5`, keeping timestamps as `time.Time` and avoiding direct SQL dependencies in callers.
- `pkg/repo/sqlite/migrations.go:1` executes on every boot, enabling WAL mode and foreign keys before creating the `projects`, `assets`, `pages`, and `spreads` tables.
- Repositories share helpers that translate nullable fields and timestamps (`pkg/repo/sqlite/sqlite.go:44`). This keeps JSON payloads and DB rows consistent with the Simple renderer.
- Each repo performs the smallest necessary mutation: for example, pages/spreads use `INSERT ... ON CONFLICT` so upserts remain idempotent while preserving `created_at` (`pkg/repo/sqlite/pages.go:18`, `pkg/repo/sqlite/spreads.go:18`).

Key behaviours to know:

- Assets are keyed by `(project_id, id)` and retain filesystem order through the `sort_index` column (`pkg/repo/sqlite/assets.go:43`). Reordering runs inside a single transaction to avoid partially-applied sort orders (`pkg/repo/sqlite/assets.go:82`).
- Pages and spreads store their algorithm payloads verbatim as JSON text. The `settings_json` column is required, while `result_json` remains nullable for draft stages (`pkg/repo/sqlite/pages.go:21`, `pkg/repo/sqlite/spreads.go:22`).
- Foreign keys tie pages to assets and spreads back to pages, letting cascading deletes remove associated layout data automatically (see schema definitions in `pkg/repo/sqlite/migrations.go:9`).

## Server integration notes

`serve.Server` now orchestrates both disk and SQL concerns:

- Startup sets project/preset/upload roots, opens SQLite with a busy timeout + WAL DSN, and wires repositories (`pkg/serve/server.go:680`). Errors during repo construction close the DB to avoid leaking handles.
- `upsertProjectRecord` mirrors `project.json` metadata into the `projects` table, preserving existing timestamps and preset choices (`pkg/serve/server.go:62`).
- `ensureProjectAssets` backfills SQL rows from the on-disk `images/` directory whenever a project lacks asset records (`pkg/serve/server.go:113`). This keeps the carousel populated for older projects without requiring a manual migration.
- Asset responses still return the legacy `{images, order}` payload expected by the UI (`pkg/serve/server.go:174`), even though repository rows carry richer metadata.
- Settings/results for pages and spreads are serialized via dedicated helpers so round-tripping through JSON matches Simple compute/render semantics (`pkg/serve/server.go:344`). Missing JSON values surface as 400s before hitting the DB, keeping persisted data clean.
- `handleUpsertPage` and `handleUpsertSpread` retain each record’s original `created_at` timestamp when updating, ensuring deterministic ordering for history views (`pkg/serve/server.go:462`, `pkg/serve/server.go:573`).
- Deleting pages/spreads bumps the project `updated_at`, keeping cross-layer sync accurate (`pkg/serve/server.go:505`, `pkg/serve/server.go:604`).

### Seeding & dual-write strategy

- Asset uploads write the PNG to disk first, then immediately upsert a matching SQL row with the computed dimensions and byte size (`pkg/serve/server.go:194`).
- The YAML export path still reads disk metadata but now benefits from ordered assets seeded via the repo, so in future we can hydrate YAML entirely from SQL (`pkg/serve/server.go:1471`).
- Until the React layer consumes `GET /pages` and `GET /spreads`, the server continues to accept ad-hoc compute requests; persisted records are optional but ready for the UI to adopt.

## REST endpoints & client usage

- The assets collection accepts multipart uploads, reorder POSTs, and deletions while persisting through the repo (`pkg/serve/server.go:174`).
- `GET /api/projects/{id}/pages` and `PUT /api/projects/{id}/pages/{page_number}` expose typed layout storage; both return the canonical JSON produced by `pageRecordToResponse` (`pkg/serve/server.go:383`).
- `GET /api/projects/{id}/spreads` plus the spread upsert/delete mirror the page semantics and allow linking pages by number (`pkg/serve/server.go:522`).
- RTK Query surfaces ergonomic hooks (`useGetPagesQuery`, `usePutPageMutation`, etc.) and serializes the request bodies to the shape expected by the server (`web/src/api.ts:233`). Existing hooks for images continue to work, sharing the `Image` cache tag.

For new feature work, start by wiring the BookSpreadDesigner to these hooks so spread edits persist even after refresh. Once the UI depends on the repos, the filesystem JSON can gradually become a legacy compatibility layer.

## Operational tips

- `go test ./...` exercises the repository glue by hitting the server package; add focused unit tests for SQLite repos to catch migration regressions.
- Use `make serve-tmux` (or `tmux new -s zine-layout -- make serve`) to keep the Go API and Vite dev server running side-by-side while tailing the sqlite-backed logs.
- The SQLite database lives at `data/zine-layout.db`; remove it to rebuild from the filesystem if you need a clean slate during development.
