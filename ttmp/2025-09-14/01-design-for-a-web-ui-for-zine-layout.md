## Zine Layout — Web UI Design

### 1. Purpose and Scope

- **Goal**: Provide a friendly web interface to compose multi-page zines from input images using the existing `zine-layout` engine. Non-technical users should be able to configure layouts and generate print-ready PNG sheets without touching the CLI or YAML.
- **Backend**: Reuse current Go engine and YAML DSL. The web app translates UI inputs into the DSL, invokes the layout pipeline, and returns generated pages for preview/download.
- **Constraints from engine**:
  - All input images per render batch must have identical dimensions.
  - Rotation for cells is restricted to 0° or 180°.
  - Color accepts hex (e.g., `#000000`), known names (e.g., `black`), or `[R,G,B,A]`.
  - Input images are 1-based indexed per batch; images are consumed in groups of `rows*columns*len(output_pages)`.
  - Currently supports PNG input/outputs.

### 2. Primary Jobs To Be Done

- **Upload inputs**: Add a set of source images, auto-ordered, with simple reordering.
- **Pick a layout**: Choose from presets (2-up, 4-up, 8-sheet zine, 16-sheet zine) or define grid rows/columns.
- **Margins and borders**: Configure global page margins, per-page margins, and per-cell margins; toggle border styles and colors.
- **Rotation and placement**: Optional 0°/180° rotations per cell; drag images into grid cells.
- **Units and PPI**: Choose units, set PPI, and use expressions when needed (e.g., `1/8 in`).
- **Preview and export**: Render previews, paginate, and download PNG files.
- **Save/share**: Save a project with its layout spec and images metadata; export/import YAML.
- **Quick test mode**: Generate colored or BW test pages to explore layouts without uploads.

### 3. Information Architecture

- **Home/Dashboard**
  - Recent projects
  - Create New Project
  - Import YAML
- **Project Editor** (primary workspace)
  - Left sidebar: Project settings (Global + Page Setup)
  - Center: Grid canvas per output page with thumbnails
  - Right sidebar: Contextual properties (selected page/cell)
  - Bottom: Input image tray with reorder and selection
- **Render/Export Panel**
  - Preview pages, download all or selected pages
- **YAML View (Advanced)**
  - Two-way view; edit YAML or copy spec

### 4. Pages and Layouts

#### 4.1 Home / Dashboard

Functions: list recent, create new, import YAML.

ASCII Prototype:

```
+--------------------------------------------------------------+
| Zine Layout                                                  |
|                                                              |
| [ New Project ]   [ Import YAML ]                            |
|                                                              |
| Recent Projects                                              |
|  - SummerZine (edited 2h ago)   [Open] [Delete]              |
|  - Test 8-sheet (yesterday)     [Open] [Delete]              |
+--------------------------------------------------------------+
```

#### 4.2 Project Editor

High density workspace split into three columns + bottom tray.

```
+-----------------------------------------------------------------------------------+
| Sidebar (Global & Page Setup) |          Canvas (Output Page)         | Properties|
|                                |                                        |          |
| Project                                                             [Save] [Render]|
| - Name: [ My Zine      ]                                                 [YAML ⤵]|
| - PPI:  [ 300  v]                                                        [Export ]|
| Global Border: [ ] Enabled  Type: [plain v] Color: [#000000  ]                   |
|                                                                                  |
| Page Setup                                                                        |
| - Grid: Rows [ 2 ]  Columns [ 2 ]                                                |
| - Page Margin:  Top [ 0.25in ] Bottom [ 0.25in ] Left [ 0.25in ] Right [ 0.25in ]|
| - Page Border: [ ] Enabled  Type: [dotted v] Color: [gray     ]                   |
|                                                                                  |
| Output Pages                                                                      |
| - Pages: [ Page 1 v ]  [ + Add Page ]                                            |
|                                                                                  |
|                                +-------------------------------+                 |
|                                |   [0,0]     |    [0,1]        |                 |
|                                |             |                 |                 |
|                                |  [Drop]     |    [Drop]       |                 |
|                                |             |                 |                 |
|                                |-------------+-----------------|                 |
|                                |   [1,0]     |    [1,1]        |                 |
|                                |             |                 |                 |
|                                |  [Drop]     |    [Drop]       |                 |
|                                +-------------------------------+                 |
|                                                                                  |
|                                                                                  |
| Input Tray:  [ + Upload Images ]  [Select All] [Remove] [Reorder]                |
| [img#1] [img#2] [img#3] [img#4] ...                                             |
|                                                                                  |
+-----------------------------------------------------------------------------------+
```

Right-side Properties (contextual):

```
Properties
----------
[Page | Cell]

When Page selected:
- ID: [ page-1 ]
- Layout Border: [ ] Enabled  Type: [corner v] Color: [black]

When Cell selected (r,c):
- Input Index: [ 1 ] (auto-updated on drop)
- Rotation: [ 0 v ]   (allow 0 or 180)
- Margin:  Top [ 5px ] Bottom [ 5px ] Left [ 5px ] Right [ 5px ]
- Inner Border: [ ] Enabled  Type: [dashed v] Color: [#999999]
```

Notes:
- PNG uploads only in v1 (engine decodes PNG). Future: auto-convert JPEG/PDF.
- Show image dimension badge on thumbnails; warn if sizes mismatch.

#### 4.3 Render/Export Panel

```
+---------------------------------- Render ---------------------------------------+
| Pages: [◀ Prev]  Page 1 / 2  [Next ▶]   Zoom: [-] 100% [+]                      |
|                                                                                  |
|   ┌───────────────────────────────────────────────────────────────────────────┐  |
|   │                                (Preview)                                 │  |
|   └───────────────────────────────────────────────────────────────────────────┘  |
|                                                                                  |
| [ Download current PNG ]  [ Download all as ZIP ]   [ Re-render ]                |
+----------------------------------------------------------------------------------+
```

#### 4.4 YAML View (Advanced)

```
+---------------------------- YAML (read-only by default) -------------------------+
| [ Edit YAML ] [ Copy ]                                                            |
|                                                                                  |
| global:                                                                          |
|   ppi: 300                                                                       |
| page_setup:                                                                      |
|   grid_size: { rows: 2, columns: 2 }                                             |
|   margin: { top: 0.25in, right: 0.25in, bottom: 0.25in, left: 0.25in }           |
|   border: { enabled: true, type: dotted, color: gray }                            |
| output_pages:                                                                     |
|   - id: page-1                                                                    |
|     margin: { top: 0px, right: 0px, bottom: 0px, left: 0px }                      |
|     border: { enabled: false, type: plain, color: black }                         |
|     layout:                                                                       |
|       - input_index: 1                                                            |
|         position: { row: 0, column: 0 }                                           |
|         rotation: 0                                                               |
|         margin: { top: 5px, right: 5px, bottom: 5px, left: 5px }                  |
+----------------------------------------------------------------------------------+
```

### 5. UX Flows

- **New Project**
  1) Click New Project → choose preset: 2-up, 4-up, 8-sheet, 16-sheet, or Custom grid
  2) Upload images → appear in tray, auto-fill cells by order
  3) Adjust margins/borders/rotation as needed
  4) Click Render → preview and download

- **Drag and Drop Placement**
  - Drag thumbnails to grid cells; dropping sets `input_index` for that cell.
  - Keyboard: arrows to change selected cell; Enter to open Properties.

- **Units/PPI**
  - Fields accept expressions: `10mm`, `0.25in`, `1/8 in`.
  - Inline validation and tooltips show computed pixels at current PPI.

- **Validation**
  - Warn if uploaded images have mismatched sizes (engine requires same size).
  - Show computed images-per-output and whether the number of inputs matches multiples.
  - Rotation limited to 0 or 180 (consistent with current engine behavior).

- **Saving/Exporting**
  - Save project stores generated YAML and lightweight references to uploaded images.
  - Export options: YAML spec only; rendered PNGs; ZIP of all outputs.

- **Quick Test Mode**
  - Toggle: [ Use Test Images ]
  - Options: Colored or Black/White, dimensions fields (`width,height`) with unit parsing.

### 6. Technical Design Overview

- **Frontend (React + Redux Toolkit + RTK Query + TypeScript)**
  - Stack: React 18 + TypeScript + Vite, Redux Toolkit for state, RTK Query for API, React Router v6 for routing. Optional UI: Bootstrap/Tailwind.
  - Routes:
    - `/` → Dashboard (recent projects, new/import)
    - `/projects/:projectId` → Project Editor
    - `/projects/:projectId/render` → Render/Export
    - `/projects/:projectId/yaml` → YAML View
  - Component map:
    - `App` (router + layout shell)
    - `DashboardPage`
    - `ProjectEditorPage`
      - `SidebarGlobal` (PPI, global border)
      - `SidebarPageSetup` (grid, margins, page border)
      - `OutputPagesNav` (add/select page)
      - `GridCanvas` (cells, drag-and-drop)
      - `PropertiesPanel` (page/cell props)
      - `ImageTray` (thumbnail list, upload, reorder)
      - `PresetModal`
    - `RenderPage` (preview, download)
    - `YamlPage` (editor/viewer)
    - `Toast/Notifications`, `ConfirmDialog`
  - Store (Redux Toolkit):
    - Slices for UI-only state: `uiSlice` (selection, modals, toasts), `editorSlice` (local edits before sync if needed).
    - RTK Query `api` slice for server data.
    - Configure store with `api.middleware` and slice reducers.
  - RTK Query API design:
    - `tagTypes`: `['Project','Spec','Image','Render','Preset']`
    - Endpoints (typed):
      - Projects: `getProjects`, `createProject`, `getProject`, `updateProject`, `deleteProject`
      - Presets: `getPresets`, `getPreset`
      - Spec: `getYaml`, `putYaml`, `fromUI`, `toUI`
      - Images: `getImages`, `uploadImages`, `deleteImage`, `reorderImages`, `getImage`
      - Validate: `validateSpec`
      - Render: `render`, `getRenders`, `getRenderFile`, `downloadZip`
    - Cache policy:
      - `getProjects` → provides `Project`
      - Mutations on project → invalidate `Project`
      - `getImages` → provides `Image`; `upload/delete/reorder` → invalidate `Image` and `Spec`
      - `getYaml`/`putYaml`/`fromUI` → provides/invalidates `Spec`
      - `render` → invalidates `Render`; `getRenders` provides `Render`
  - Key types (TS):
    ```ts
    type BorderType = 'plain'|'dotted'|'dashed'|'corner';
    type Color = string | [number,number,number,number];
    interface Margin { top: string; bottom: string; left: string; right: string; }
    interface GridSize { rows: number; columns: number; }
    interface Position { row: number; column: number; }
    interface LayoutItem {
      input_index: number; // 1-based index into images order
      position: Position;
      rotation: 0|180;
      margin: Margin;
      border?: { enabled: boolean; type: BorderType; color: Color; };
    }
    interface OutputPage {
      id: string;
      margin: Margin;
      border?: { enabled: boolean; type: BorderType; color: Color; };
      layout: LayoutItem[];
    }
    interface SpecUI {
      global: { ppi: number; border?: { enabled: boolean; type: BorderType; color: Color; } };
      page_setup: { grid_size: GridSize; margin: Margin; border?: { enabled: boolean; type: BorderType; color: Color; } };
      output_pages: OutputPage[];
    }
    interface ImageItem { id: string; name: string; width: number; height: number; }
    interface Project { id: string; name: string; createdAt: string; updatedAt: string; }
    ```
  - API slice sketch:
    ```ts
    export const api = createApi({
      reducerPath: 'api',
      baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
      tagTypes: ['Project','Spec','Image','Render','Preset'],
      endpoints: (b) => ({
        getProjects: b.query<{projects:Project[]}, void>({
          query: () => '/projects', providesTags: ['Project']
        }),
        createProject: b.mutation<{project:Project}, {name?:string,presetId?:string}>({
          query: (body) => ({ url: '/projects', method: 'POST', body }),
          invalidatesTags: ['Project']
        }),
        getImages: b.query<{images:ImageItem[],order:string[]}, string>({
          query: (id) => `/projects/${id}/images`, providesTags: ['Image']
        }),
        uploadImages: b.mutation<{images:ImageItem[]}, {id:string, files:FileList}>({
          query: ({id,files}) => ({ url: `/projects/${id}/images`, method: 'POST', body: toFormData(files) }),
          invalidatesTags: ['Image','Spec']
        }),
        getYaml: b.query<string, string>({ query: (id) => ({ url: `/projects/${id}/yaml`, responseHandler: (r)=>r.text() }), providesTags: ['Spec'] }),
        putYaml: b.mutation<{ok:true}, {id:string, yaml:string}>({
          query: ({id,yaml}) => ({ url: `/projects/${id}/yaml`, method: 'PUT', body: yaml, headers: {'Content-Type':'text/plain'} }),
          invalidatesTags: ['Spec']
        }),
        render: b.mutation<{renderId:string,files:string[]}, {id:string, opts?:{test?:boolean,test_bw?:boolean,test_dimensions?:string}} >({
          query: ({id,opts}) => ({ url: `/projects/${id}/render`, method: 'POST', body: opts||{} }),
          invalidatesTags: ['Render']
        })
      })
    });
    ```
  - Store setup sketch:
    ```ts
    export const store = configureStore({
      reducer: { [api.reducerPath]: api.reducer, ui: uiReducer, editor: editorReducer },
      middleware: (gDM) => gDM().concat(api.middleware)
    });
    ```
  - Drag-and-drop: HTML5 DnD on `GridCanvas` and `ImageTray`. Dropping sets `layout.input_index` to the image's current 1-based position in the project `order`. On reorder, recompute all `layout.input_index` by mapping `imageId -> index`:
    ```ts
    function remapInputIndexes(layouts: LayoutItem[], order: string[], idByIndex: (idx:number)=>string) {
      const posById = new Map(order.map((id, i) => [id, i+1]));
      for (const li of layouts) {
        const imageId = idByIndex(li.input_index-1);
        li.input_index = posById.get(imageId) ?? li.input_index;
      }
    }
    ```
  - Validation: call `useValidateSpecMutation()` on significant changes; surface issues inline.
  - YAML View: editable code editor; save via `usePutYamlMutation`, refresh via `useToUIQuery`.

- **Backend (Go, same binary, new `serve` command)**
  - Run: `zine-layout serve --root /path/to/data --addr :8080`
  - Storage layout (on disk):
    ```
    <root>/
      projects/
        <projectId>/
          project.json
          spec.yaml
          images/
            0001.png
            0002.png
          renders/
            <renderId>/
              output1_1.png
              output1_2.png
          thumbnails/ (optional)
      presets/
        two_pages_two_inputs.yaml
        10_8_sheet_zine.yaml
        11_16_sheet_zine.yaml
    ```
  - `project.json`:
    ```json
    {
      "id": "abcd1234",
      "name": "My Zine",
      "createdAt": "2025-09-15T10:00:00Z",
      "updatedAt": "2025-09-15T11:00:00Z",
      "images": [
        { "id": "0001", "filename": "0001.png", "width": 600, "height": 800 },
        { "id": "0002", "filename": "0002.png", "width": 600, "height": 800 }
      ],
      "order": ["0001","0002"],
      "lastRenderId": "r-20250915-1100"
    }
    ```
    - Maintain image order by `order[]`; file names remain stable.
    - Guarantee: all uploaded images are saved on disk under `images/` alongside `project.json`. No transient-only storage.
  - Endpoints (prefix `/api`):
    - Projects
      - `GET /projects` → list `{ projects: Project[] }`
      - `POST /projects` `{ name?: string, presetId?: string }` → `{ project: Project }` (creates dir, copies preset to `spec.yaml`)
      - `GET /projects/:id` → `{ project, specYamlExists: boolean }`
      - `PUT /projects/:id` `{ name }` → `{ project }`
      - `DELETE /projects/:id` → `{ ok: true }`
    - Presets
      - `GET /presets` → `{ presets: { id, name }[] }`
      - `GET /presets/:id` → raw YAML
    - Spec
      - `GET /projects/:id/yaml` → raw YAML
      - `PUT /projects/:id/yaml` (text/plain) → `{ ok: true }` (saves to disk)
      - `POST /projects/:id/spec/from-ui` `{ uiState: SpecUI }` → `{ yaml: string }` (also saves `spec.yaml`)
      - `POST /projects/:id/spec/to-ui` (no body) → `{ uiState: SpecUI }` (parses `spec.yaml`)
    - Images
      - `GET /projects/:id/images` → `{ images: ImageItem[], order: string[] }`
      - `POST /projects/:id/images` (multipart `images[]`) → `{ images: ImageItem[] }` (validates PNG, stores as `images/NNNN.png`, updates `order`)
      - `DELETE /projects/:id/images/:imageId` → `{ ok: true }`
      - `POST /projects/:id/images/reorder` `{ order: string[] }` → `{ order: string[] }`
      - `GET /projects/:id/images/:imageId` → serves image file (Content-Type: image/png)
    - Validate
      - `POST /projects/:id/validate` `{ options? }` → `{ ok: boolean, issues: string[] }` (checks same size, images-per-output multiple)
    - Render
      - `POST /projects/:id/render` `{ test?: boolean, test_bw?: boolean, test_dimensions?: string }` → `{ renderId, files: string[] }`
      - `GET /projects/:id/renders` → `{ renders: { id, createdAt, files: string[] }[] }`
      - `GET /projects/:id/renders/:renderId/files/:name` → serves PNG
      - `GET /projects/:id/download.zip?renderId=<id>` → ZIP stream of that render (or latest)
  - Render flow:
    - Read `spec.yaml` and process with Emrichen+Sprig (same as CLI). Unmarshal to `zinelayout.ZineLayout`. Apply optional overrides (ppi, borders) if included in request.
    - Source images: if `test*` flags present, generate via `GenerateTestImages`/`GenerateTestImagesBW`. Else, load project images in `order[]`.
    - Validate: ensure equal sizes and count multiple of `rows*cols*len(output_pages)`; return issues otherwise.
    - For each batch group: for each `output_page`, call `CreateOutputImage`, save as `renders/<renderId>/output<batchIndex>_<pageIndex>.png`.
    - Persist `render.json` metadata: list of files, dimensions, timings.
  - Notes:
    - Images are always persisted to disk together with `project.json` (source of truth: the on-disk folder).
    - PNG-only uploads in v1; future: auto-convert JPEG→PNG server-side.
    - Thumbnails optional (`thumbnails/`); can be generated lazily on request.
    - Concurrency: parallelize per-output-page generation with `errgroup`.
    - Limits: max image size and count per project (configurable).
 
- **Data Model (UI State)**
  - Mirrors Go structs with UI extras:
    - `inputs[]`: `{ id, name, width, height }`
    - `global`: `{ ppi, border{ enabled, type, color } }`
    - `page_setup`: `{ grid_size{ rows, columns }, margin{ top,bottom,left,right }, border }`
    - `output_pages[]`: `{ id, margin, border (as layout_border), layout[] }`
    - `layout[]`: `{ input_index, position{row,column}, rotation, margin, border (inner) }`

### 7. Controls and Widgets

- **Color picker**: supports hex, named colors, or RGBA; validate to Go parser constraints.
- **Border type**: dropdown with plain/dotted/dashed/corner; preview swatch.
- **Units input**: free text with hint; computed px shown in-line.
- **Grid editor**:
  - Change rows/columns → canvas re-renders dotted grid.
  - Empty cells show “Drop here”; filled cells show thumbnail + input index.
  - Context menu: Clear cell, Rotate 180°, Set margins.

### 8. Presets

- 2-up landscape (1x2)
- 2x2 (4-up)
- 8-sheet zine (2x4 with specific placement/rotations)
- 16-sheet zine (4x4 or 2x8 depending on spec)

Each preset is defined as a YAML template and loaded into the editor, with visual helper text for folding/cutting where applicable.

### 9. Accessibility and Internationalization

- Keyboard navigation for grid and forms; ARIA labels on inputs.
- Unit picker localized labels; color names friendly to non-English by hex fallback.

### 10. Error Handling and Messaging

- Inline field errors (unit parse failures, invalid color/type).
- Global warning bar for engine-level constraints (image sizes, count multiples).
- Non-blocking to let users experiment; render disables until valid.

### 11. Security and Privacy

- Images processed in-memory and not persisted unless user saves project.
- If deployed online: max file size and rate limiting.

### 12. Non-goals (Initial Version)

- PDF export (future enhancement; start with PNG + ZIP).

### 13. Future Enhancements

- PDF/print layout export with bleed/crop marks.
- Per-preset wizards that explain folding/cutting for 8/16-sheet zines.

### 14. Additional ASCII Wireframes

Preset chooser modal:

```
+---------------------- Choose a preset ----------------------+
|  ( ) Custom grid     Rows [  ]  Columns [  ]                |
|  (•) 2-up landscape  (1x2)                                   |
|  ( ) 2x2 (4-up)                                             |
|  ( ) 8-sheet zine (2x4)                                     |
|  ( ) 16-sheet zine                                          |
|                                                              |
|                                   [ Cancel ]   [  Continue ] |
+--------------------------------------------------------------+
```

Cell context menu:

```
┌───────────────┐
│ Clear cell    │
│ Rotate 180°   │
│ Set margins…  │
└───────────────┘
```

### 15. Mapping UI → DSL

- Global
  - PPI → `global.ppi`
  - Global border → `global.border`
- Page Setup
  - Grid rows/cols → `page_setup.grid_size`
  - Margin → `page_setup.margin`
  - Page border → `page_setup.border`
- Output Page (selected page)
  - ID → `output_pages[i].id`
  - Margin → `output_pages[i].margin`
  - Layout border → `output_pages[i].border` (named as Layout Border in UI)
- Cell
  - Input selection/order → `layout[].input_index`
  - Position → `layout[].position.{row,column}`
  - Rotation → `layout[].rotation` (0|180)
  - Margin → `layout[].margin`
  - Inner border → `layout[].border`

This design aligns with the current Go model and constraints observed in `pkg/zinelayout` and the CLI, and proposes a React + Redux Toolkit + RTK Query web app that maps 1:1 to the DSL while remaining approachable for non-technical users.

### 16. Incremental Implementation Plan (Testable Milestones)

Below is a step-by-step plan to build and validate the system incrementally. Each step includes backend endpoints, minimal frontend UI, and test instructions.

1) Serve skeleton and health check
   - Backend:
     - Command: `zine-layout serve --root <path> --addr :8080`
     - Endpoint: `GET /api/health` → `{ ok: true }`
     - Creates `<root>/projects/` and `<root>/presets/` if missing.
   - Frontend:
     - Boot React app with Router and a simple Health status banner (calls `/api/health`).
   - Test:
     - `curl -s localhost:8080/api/health | jq`
     - Load `/` and confirm “Server OK”.

2) Projects CRUD (directories + project.json)
   - Backend:
     - `GET /api/projects` → list projects from `<root>/projects/*/project.json`.
     - `POST /api/projects` `{ name? }` → create dir, write minimal `project.json` with `images:[], order:[]`.
     - `GET /api/projects/:id` → metadata.
     - `PUT /api/projects/:id` `{ name }` → rename in json (directory name unchanged).
     - `DELETE /api/projects/:id` → delete directory recursively.
   - Frontend:
     - Dashboard lists projects with Open/Delete; New Project modal posts to API.
     - RTK Query: `getProjects`, `createProject`, `getProject`, `updateProject`, `deleteProject`.
   - Test:
     - `curl -X POST localhost:8080/api/projects -d '{"name":"My Zine"}' -H 'Content-Type: application/json'`
     - Verify directory and `project.json` on disk.

3) Image management (persist images on disk)
   - Backend:
     - `POST /api/projects/:id/images` multipart `images[]` → save as `images/NNNN.png`, update `project.json.images` and `order`.
     - `GET /api/projects/:id/images` → `{ images, order }` (width/height read via decoder).
     - `DELETE /api/projects/:id/images/:imageId` → remove file, update `project.json`.
     - `POST /api/projects/:id/images/reorder` `{ order: string[] }` → persist new order in `project.json`.
     - `GET /api/projects/:id/images/:imageId` → serve image file.
     - Invariants: images are saved on disk together with `project.json` (no in-memory-only).
   - Frontend:
     - `ImageTray` with Upload (input type=file multiple), list, delete, drag-reorder.
     - RTK Query: `getImages`, `uploadImages`, `deleteImage`, `reorderImages`, `getImage`.
   - Test:
     - `curl -F images[]=@one.png -F images[]=@two.png localhost:8080/api/projects/<id>/images`
     - Verify files under `<root>/projects/<id>/images/` and updated `project.json`.

4) Presets (read-only)
   - Backend:
     - `<root>/presets/*.yaml` shipped from `zine-layout/examples`.
     - `GET /api/presets` → list names/ids.
     - `GET /api/presets/:id` → raw YAML.
   - Frontend:
     - Preset chooser to create project from preset (calls `POST /projects` with `presetId`).
   - Test:
     - `curl localhost:8080/api/presets | jq` and fetch one YAML.

5) Spec YAML endpoints
   - Backend:
     - `GET /api/projects/:id/yaml` → raw.
     - `PUT /api/projects/:id/yaml` (text/plain) → write `spec.yaml`.
     - `POST /api/projects/:id/spec/to-ui` → `{ uiState }` from `spec.yaml`.
     - `POST /api/projects/:id/spec/from-ui` `{ uiState }` → `{ yaml }` and write `spec.yaml`.
   - Frontend:
     - YAML page with editor; Save → `putYaml` then refresh.
     - Editor page in Project Editor shows computed read-only YAML from UI state.
   - Test:
     - `curl -T spec.yaml -H 'Content-Type: text/plain' localhost:8080/api/projects/<id>/yaml`
     - `curl -X POST localhost:8080/api/projects/<id>/spec/to-ui`

6) Validation
   - Backend:
     - `POST /api/projects/:id/validate` → verifies same image sizes and multiples of `rows*cols*pages`.
   - Frontend:
     - Show issues panel in Project Editor; disable render when invalid.
   - Test:
     - Upload mismatched images and confirm issues.

7) Render
   - Backend:
     - `POST /api/projects/:id/render` `{ test?, test_bw?, test_dimensions? }` → create `renders/<renderId>/outputX_Y.png`.
     - `GET /api/projects/:id/renders` → list; `GET /.../files/:name` → serve PNG; `GET /.../download.zip` → zip.
   - Frontend:
     - Render page with preview carousel; Download PNG/ZIP.
   - Test:
     - `curl -X POST localhost:8080/api/projects/<id>/render -H 'Content-Type: application/json' -d '{}' | jq`

8) Grid canvas and placement
   - Frontend:
     - Implement `GridCanvas` and `PropertiesPanel` (DnD from tray assigns `input_index`).
     - Persist via `fromUI` (write `spec.yaml`), display computed YAML; validate on change.
   - Test:
     - Place images, verify `spec.yaml` content and successful render.

9) Polish and robustness
   - Backend:
     - Add size limits, content-type checks, atomic writes of `project.json`.
     - Concurrency: `errgroup` for rendering per page.
   - Frontend:
     - Error handling, toasts, optimistic updates with RTK Query.
   - Test:
     - Simulate failures and confirm robust UX.