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

Implementation: provide pkg/repo/sqlite with concrete types and a NewSQLiteRepositories(db *sql.DB) factory returning a struct with these repos. Use BEGIN IMMEDIATE transactions where appropriate (reorder, batch upserts).

## REST API design

Keep existing routes and add DB-backed routes. New resources:

- Assets
  - GET /api/projects/{id}/assets -> { assets: Asset[] }
  - POST /api/projects/{id}/assets (multipart form: file) -> { asset: Asset }
  - POST /api/projects/{id}/assets/reorder -> { ok: true }
  - DELETE /api/projects/{id}/assets/{assetId} -> { ok: true }

Note: Maintain /uploads endpoint for generic uploads; for project assets prefer the project assets endpoint so we persist metadata.

- Pages
  - GET /api/projects/{id}/pages -> { pages: Page[] }
  - PUT /api/projects/{id}/pages/{pageNumber} body: { assetId?, settings, result? } -> { page: Page }
  - DELETE /api/projects/{id}/pages/{pageNumber} -> { ok: true }

- Spreads
  - GET /api/projects/{id}/spreads -> { spreads: Spread[] }
  - PUT /api/projects/{id}/spreads/{spreadNumber} body: { leftPageNumber?, rightPageNumber?, settings, result? } -> { spread: Spread }
  - DELETE /api/projects/{id}/spreads/{spreadNumber} -> { ok: true }

Interplay with preview and render:
- Existing /api/v1/preview and /api/v1/render stay as-is; we can add variants that reference a Page or Spread by number which lookup asset(s) from DB.

Error handling:
- 404 when project or asset not found
- 400 for invalid body
- 409 for reorder mismatches

## RTK Query additions

New types in web/src/api.ts:
```ts
export interface Asset { id: string; projectId: string; filename: string; rel_path: string; content_type: string; bytes: number; width: number; height: number; sort_index: number; created_at: string }
export interface Page { id: string; projectId: string; pageNumber: number; assetId?: string; settings: any; result?: any; createdAt: string; updatedAt: string }
export interface Spread { id: string; projectId: string; spreadNumber: number; leftPageId?: string; rightPageId?: string; settings: any; result?: any; createdAt: string; updatedAt: string }
```

Add endpoints:
```ts
getAssets: b.query<{ assets: Asset[] }, { id: string }>({ query: ({ id }) => `/projects/${id}/assets`, providesTags: (_r,_e,arg) => [{ type: 'Image', id: arg.id }] }),
uploadAsset: b.mutation<{ asset: Asset }, { id: string; file: File }>({ query: ({ id, file }) => { const fd = new FormData(); fd.append('file', file); return { url: `/projects/${id}/assets`, method: 'POST', body: fd }; }, invalidatesTags: ['Image'] }),
reorderAssets: b.mutation<{ ok: boolean }, { id: string; order: string[] }>({ query: ({ id, order }) => ({ url: `/projects/${id}/assets/reorder`, method: 'POST', body: { order } }), invalidatesTags: ['Image'] }),
deleteAsset: b.mutation<{ ok: boolean }, { id: string; assetId: string }>({ query: ({ id, assetId }) => ({ url: `/projects/${id}/assets/${encodeURIComponent(assetId)}`, method: 'DELETE' }), invalidatesTags: ['Image'] }),

getPages: b.query<{ pages: Page[] }, { id: string }>({ query: ({ id }) => `/projects/${id}/pages` }),
putPage: b.mutation<{ page: Page }, { id: string; pageNumber: number; page: Partial<Page> }>({ query: ({ id, pageNumber, page }) => ({ url: `/projects/${id}/pages/${pageNumber}`, method: 'PUT', body: page }), invalidatesTags: ['Project'] }),
deletePage: b.mutation<{ ok: boolean }, { id: string; pageNumber: number }>({ query: ({ id, pageNumber }) => ({ url: `/projects/${id}/pages/${pageNumber}`, method: 'DELETE' }), invalidatesTags: ['Project'] }),

getSpreads: b.query<{ spreads: Spread[] }, { id: string }>({ query: ({ id }) => `/projects/${id}/spreads` }),
putSpread: b.mutation<{ spread: Spread }, { id: string; spreadNumber: number; spread: Partial<Spread> }>({ query: ({ id, spreadNumber, spread }) => ({ url: `/projects/${id}/spreads/${spreadNumber}`, method: 'PUT', body: spread }), invalidatesTags: ['Project'] }),
deleteSpread: b.mutation<{ ok: boolean }, { id: string; spreadNumber: number }>({ query: ({ id, spreadNumber }) => ({ url: `/projects/${id}/spreads/${spreadNumber}`, method: 'DELETE' }), invalidatesTags: ['Project'] }),
```

Cache tags:
- Continue using tagTypes: Project, Image, Preset. Assets reuse Image tag for now to avoid churn.

## React integration plan

Image carousel:
- Replace current Images endpoints usage with Assets endpoints
- Component shows assets ordered by sortIndex; drag and drop to reorder -> calls reorderAssets
- Clicking an asset selects it as the current image for preview and for page assignment

Pages management:
- Add panel to define number of pages and map assets to pages (pageNumber -> assetId)
- When user updates settings (paper size, margins, dpi) or selects a new asset for a page, PUT to pages to persist settings_json; preview endpoints can accept either raw request or reference to pageNumber for convenience

Spreads management:
- Add a list of spreads with spreadNumber; allow assigning left/right page numbers
- Persist spread settings_json; compute result via preview and save result_json
- Allow render from spread selection

Routing and state:
- Extend Redux slice to include selectedPageNumber and selectedSpreadNumber
- Update selectors used by usePreviewRequest to optionally read settings from the current page or spread record

Migration plan:
- Introduce sqlite-backed repositories and keep existing filesystem code for compatibility
- Add a feature flag to switch endpoints to DB-backed variants
- Start with assets for carousel and progressively add pages and spreads

## Implementation checklist

- [ ] Create pkg/repo with interfaces and pkg/repo/sqlite implementation
- [ ] Add sqlite initialization in server startup with migration execution
- [ ] Implement assets endpoints backed by DB and disk storage
- [ ] Add pages and spreads endpoints
- [ ] Extend RTK Query endpoints and types
- [ ] Build React components for carousel, pages, spreads, and integrate with preview
- [ ] Add tests for repository methods and endpoint handlers

Appendix: Using the repository interface pattern allows swapping sqlite for filesystem or future backends and enables unit testing by mocking repositories.


