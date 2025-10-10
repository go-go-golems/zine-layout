# Report: Current Codebase & Reuse Path for the Zine Layout Platform

## Architecture Snapshot
- **Back-end:** Go HTTP server (`pkg/serve`) backed by SQLite or filesystem. Routes expose project CRUD, image upload, page/spread persistence, and rendering utilities.
- **Persistence:** Repositories in `pkg/repo/sqlite` manage structured data; `pkg/projects` keeps backward-compatible JSON manifests under `data/projects/`.
- **Algorithms:** The `pkg/spread/simple` engine implements scaling, cropping, gutter handling, and export traces, orchestrated by `spread.Settings` and Simple YAML DSL helpers.
- **Front-end:** React + RTK Query app in `web/` offers project management, spread designer UI, YAML editor, and render/export panels.

## Stored Records & Persistence Hooks
Projects, assets, pages, and spreads are represented in the shared repository interfaces and backed by tables created at startup.

```6:66:zine-layout/pkg/repo/types.go
type Project struct {
    ID           string
    Name         string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    PresetID     *string
    CoverAssetID *string
}
// ... existing code ...
type Page struct {
    ProjectID    string
    PageNumber   int
    AssetID      *string
    SettingsJSON string
    ResultJSON   *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

```4:58:zine-layout/pkg/repo/sqlite/migrations.go
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    preset_id TEXT,
    cover_asset_id TEXT
);
// ... existing code ...
CREATE TABLE IF NOT EXISTS spreads (
    project_id TEXT NOT NULL,
    spread_number INTEGER NOT NULL,
    left_page_number INTEGER,
    right_page_number INTEGER,
    settings_json TEXT NOT NULL,
    result_json TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (project_id, spread_number)
);
```

`pkg/projects` continues to manage on-disk project manifests (`project.json`) and image file storage, providing helpers for image upload, ordering, and metadata extraction.

```17:157:zine-layout/pkg/projects/projects.go
type Project struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    Images    []string  `json:"images"`
    Order     []string  `json:"order"`
    PresetID  string    `json:"presetId,omitempty"`
}
// ... existing code ...
func SavePNGImage(projectsRoot, id string, fh *multipart.FileHeader) (*ImageItem, error) {
    // copies uploads to data/projects/{id}/images and updates project.json
}
```

## Mapping Existing Capabilities to the Target Workflow

| Target object | Current support | Reuse opportunity & gaps |
| --- | --- | --- |
| **Project** | Fully implemented via `Project` records and filesystem manifests. | Extend metadata to track sequencing presets or zine-level settings. |
| **Raw image** | Assets table + `/data/projects/<id>/images`. Upload, reorder, and delete available. | Surface additional metadata (filesize, capture date) by extending `repo.Asset` and upload handlers. |
| **Image sequence** | Only implicit ordering through `project.json` `Order` array and `assets.sort_index`. | Promote to first-class entity (e.g., new `image_sequences` table) or version existing order logic to support named sequences and gaps. |
| **Image layout template** | `spread.Settings`, presets under `data/presets/`, and Simple YAML DSL encode crop/margin/export expectations. | Introduce author-facing names/descriptions for reusable templates and persist them separately from spreads. |
| **Laid out image** | Pages table stores `settings_json` + `result_json` for a single asset/layout combo. | Repurpose `pages` as “laid-out image” records; add explicit template references and per-instance overrides. |
| **Layout sequence** | Spreads table links left/right page numbers and Settings. | Expand to sequence constructs beyond spreads (e.g., allow references to laid-out images instead of raw asset IDs). |
| **Page layout template** | Legacy zine-layout DSL (`pkg/zinelayout`) renders grid-based pages with margins and borders. | Wrap existing DSL as template engine for multi-image compositions; sync with new laid-out image outputs. |
| **Laid out page** | `render.DoProjectRender` assembles pages from zine-layout specs; Simple engine returns export rectangles. | Persist render outputs alongside metadata to avoid recomputation; attach to spreads or upcoming “pages” entity. |
| **Zine** | `spread.SimpleBookDocument` parses/serializes collections of spreads with defaults. | Use as canonical zine-level document; extend with references to layout/page templates and sequences for final assembly. |
| **Zine layout template** | Example YAML presets in `data/presets` + `examples/tests` capture multi-page folding schemes. | Catalog these presets in a dedicated table and expose via API for final imposition workflows. |

## Algorithms & Template Assets

The Simple algorithm already handles cover/contain scaling, gutters, and export sizing with detailed trace output—directly aligned with the proposed “laid out image” stage.

```204:265:zine-layout/pkg/spread/simple/algorithm.go
func ComputePlacement(inp Inputs, tr *Trace) Result {
    // computes crop window, scaling, gutter splits, and export sizes
    // Result.ExportSpread and Result.ExportSingle provide canvas dimensions
}
```

Template management, default overrides, and YAML serialization live in `pkg/spread/yaml.go`, enabling round-tripping between stored presets and runtime settings.

```93:197:zine-layout/pkg/spread/yaml.go
type SimpleBookDocument struct {
    Version  string
    Defaults Settings
    Spreads  []SimpleSpread
}
// ... existing code ...
func BuildBookYAML(defaults Settings, spreads []BookSpreadItem, baseDir string) (string, error) {
    // emits simple YAML for entire books/spreads
}
```

The older zine-layout grid renderer (`pkg/zinelayout`) still provides page-level composition, margins, and border drawing for multi-cell templates.

```13:205:zine-layout/pkg/zinelayout/layout.go
type ZineLayout struct {
    PageSetup   *PageSetup    `yaml:"page_setup"`
    OutputPages []*OutputPage `yaml:"output_pages"`
    Global      *Global       `yaml:"global"`
}
// ... existing code ...
func (zl *ZineLayout) CreateOutputImage(outputPage *OutputPage, inputImages []image.Image) (image.Image, error) {
    // computes grid cell sizes, applies margins, rotation, and draws borders
}
```

Preset YAML bundles are seeded and listed through `pkg/presets`, allowing us to version layout templates independently of projects.

```18:88:zine-layout/pkg/presets/presets.go
func SeedPresetsIfEmpty(presetsRoot string) error {
    // copies examples/tests YAML into data/presets on first run
}
```

## API Surface & Runtime Services

`pkg/serve/server.go` wires the REST API. Key resources that map to the future workflow:
- `/api/projects` — create/list/delete projects, apply presets.
- `/api/projects/{id}/images` — upload, list, reorder, delete raw assets.
- `/api/projects/{id}/pages` — upsert/list page records (laid out images).
- `/api/projects/{id}/spreads` — upsert/list spread records (layout sequence entries).
- `/api/v1/compute` & `/api/v1/preview` — run Simple algorithm and produce previews for ad-hoc layouts.

```709:1214:zine-layout/pkg/serve/server.go
mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        prjs, err := projects.ListProjects(s.projectsRoot)
        // ... existing code ...
    case http.MethodPost:
        // create project and optional preset application
    }
})
// ... existing code ...
mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
    // routes for images, pages, spreads, YAML export, renders
})
```

Because the server dual-writes to filesystem manifests and SQLite, it can be evolved to persist new entities (image sequences, layout templates) without breaking existing projects.

## Front-end Touchpoints

RTK Query endpoints encapsulate all server calls. Introducing new APIs (e.g., `/image-sequences`) means extending this central file.

```124:338:zine-layout/web/src/api.ts
export const api = createApi({
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Project', 'Image', 'Preset'],
  endpoints: (b) => ({
    getProjects: b.query({ query: () => '/projects' }),
    uploadImages: b.mutation({
      query: ({ id, files }) => ({ url: `/projects/${id}/images`, method: 'POST', body: fd })
    }),
    // ... existing code ...
    getSpreads: b.query({ query: ({ id }) => `/projects/${id}/spreads` }),
  }),
});
```

The spread designer view already demonstrates how to preview and persist layout settings, which aligns with the planned “laid out image” stage.

```70:338:zine-layout/web/src/views/BookSpreadDesigner.tsx
export const BookSpreadDesigner: React.FC = () => {
    const { data: projectsData } = useGetProjectsQuery();
    const imagesQuery = useGetImagesQuery({ id: selectedProjectId ?? '' }, { skip: !selectedProjectId });
    const { data: leftPreviewUrl } = useGetPreviewSpreadQuery(leftRequest!, { skip: !leftRequest || !isSpread });
    // ... existing code ...
    <ProjectAssetsPanel projectId={selectedProjectId || null} assets={assets} onSelectAsset={handleAssetSelect} />
}
```

`ProjectDetail` surfaces asset management and preset application, providing a natural place to add sequence editors or template assignments.

```27:213:zine-layout/web/src/views/ProjectDetail.tsx
export const ProjectDetail: React.FC = () => {
    const imagesQuery = useGetImagesQuery({ id }, { skip: !id });
    // ... existing code ...
    <ProjectAssetsPanel projectId={id || null} assets={assets} onSelectAsset={handleAssetSelect} />
}
```

Redux slice `bookSpreadSlice` mirrors `spread.Settings`, making it straightforward to map UI changes back to stored templates.

```65:172:zine-layout/web/src/store/bookSpreadSlice.ts
const initialState: BookSpreadState = {
    image: null,
    paperSize: '8x10',
    paperWidthIn: PAPER_SIZES['8x10'].width,
    paperHeightIn: PAPER_SIZES['8x10'].height,
    isSpread: false,
    // ... existing code ...
    gutterMargin: 0,
};
```

## CLI Surface

The `zine-layout` CLI (Glazed-based) bundles several developer tools found under `cmd/zine-layout`:
- `serve` – boot the HTTP server (`cmd/zine-layout serve --addr :8088`).
- `render` – run layout rendering against specs and assets (`cmd/zine-layout render --spec <file>`).
- `api` command group – REST clients under `cmd/zine-layout api` with sub-commands:
  - `projects-list|get|create|delete`
  - `images-list|upload|upload-dir|sync`
  - `presets-list`
  - `pages-list|get|put` and `spreads-list|get|put` (see `cmd/zine-layout/cmds/api` for full inventory).
- `build-web` – compile the web UI bundle (`cmd/build-web`).
- `image-info` – inspect image metadata (`cmd/image-info`).

These tools ensure parity between automated workflows (rendering, API smoke tests) and interactive management during the transition to the new model.

## Recommended Reuse & Extension Plan
1. **Elevate sequencing:** Introduce `image_sequences` (name, project_id, ordered asset IDs) and reuse ordering logic from `projects.SetProjectOrder` and `assets.UpdateOrder` for persistence.
2. **Template registry:** Persist `spread.Settings` snapshots as `image_layout_templates` (with human-readable metadata) and reuse existing preset seeding for defaults.
3. **Laid-out image entity:** Rebrand `pages` records to store `{template_id, asset_id, overrides, render_info}` rather than raw JSON strings; back-fill from existing data.
4. **Layout sequences and zines:** Map `spreads` to reference laid-out images instead of direct settings, then aggregate into a new `zines` table referencing ordered spread IDs. Leverage `SimpleBookDocument` for serialization/export.
5. **Page/layout templates:** Wrap `pkg/zinelayout` grid DSL as customizable page templates that can consume one or many laid-out images; store template definitions alongside YAML for reuse.
6. **API expansions:** Add endpoints for sequences, templates, laid-out images/pages, and zines, following existing REST patterns in `pkg/serve`. Update RTK Query and views to surface the new workflows.
7. **Rendering pipeline:** Reuse `render.DoProjectRender` and Simple preview endpoints to generate final assets; persist render metadata for zine layout templates (imposition rules) to consume.

## Observed Gaps & Considerations
- **Metadata enrichment:** Capture EXIF/date/ratio at upload time to satisfy “display info” requirements without rescanning images later.
- **Versioning:** Keep backward compatibility by migration scripts that seed default sequences/templates for legacy projects.
- **Traceability:** Store `simple.Trace` output (already produced) alongside laid-out images for debugging and audit trails.
- **Template lifecycle:** Provide copy/apply mechanisms on the front-end mirroring `applyPreset` flow to encourage reuse across projects.
- **Final imposition:** Map zine layout templates (folding schemes) to `render` outputs, reusing `data/presets` as starter definitions until a richer editor exists.

With these adjustments, the existing codebase already covers most of the heavy lifting—image ingestion, layout computation, previewing, rendering, and YAML serialization—making it a strong foundation for the layered workflow the new zine layout software envisions.


