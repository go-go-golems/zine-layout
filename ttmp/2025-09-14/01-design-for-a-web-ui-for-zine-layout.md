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

- **Frontend**
  - Use Bootstrap + htmx for interactions; server-rendered HTML via `templ` for Go alignment.
  - Components (conceptual): ImageTray, GridCanvas, SidebarForms, PropertiesPanel, PreviewModal, YAMLView
  - State shape mirrors DSL: `global`, `page_setup`, `output_pages[]`

- **Backend** (Go)
  - Add a `serve` subcommand to the existing Cobra CLI (`zine-layout serve`), exposing HTTP endpoints.
  - Endpoints:
    - POST `/render` — multipart: `spec.yaml` (text) + `images[]` (files) OR JSON `{ uiState, images }` → returns PNG pages or ZIP
    - POST `/spec/from-ui` — `{ uiState }` → `{ yaml }`
    - POST `/spec/to-ui` — `{ yaml }` → `{ uiState }`
    - POST `/validate` — `{ uiState, imagesMeta }` → `{ ok, issues[] }`
    - GET `/presets` — returns example specs (2-up, 4-up, 8-sheet, 16-sheet)
  - Implementation strategy:
    - Convert `uiState` to `zinelayout.ZineLayout` and marshal to YAML with `yaml.Marshal`.
    - Decode uploaded PNGs in-memory; compute validation (same size, multiples of `rows*cols*pages`).
    - Group images by batch size and call `CreateOutputImage` for each output page.
    - Return rendered PNGs as files; for multiple, stream as ZIP.

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

- Arbitrary rotations other than 0°, 180°.
- Non-rectangular borders beyond provided four types.
- PDF export (future enhancement; start with PNG + ZIP).

### 13. Future Enhancements

- PDF/print layout export with bleed/crop marks.
- Per-preset wizards that explain folding/cutting for 8/16-sheet zines.
- Snap guides and nudging for inner borders and margins.
- Project templates and sharing via permalink (YAML + images manifest).

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

This design aligns with the current Go model and constraints observed in `pkg/zinelayout` and the CLI, and proposes a `templ` + htmx + Bootstrap web experience that maps 1:1 to the DSL while remaining approachable for non-technical users.