# Zine Layout Platform — System Specification (Phase 1 & 2)

**Version:** 1.0  
**Last Updated:** October 11, 2025  
**Status:** Phase 1 & 2 Complete — Foundation + Image Layout Engine Operational

---

## Overview

The Zine Layout Platform is a structured workflow system for producing print-ready photobooks, zines, and multi-page layouts. Users progress through discrete stages—uploading raw images, sequencing them, applying layout templates, composing pages, assembling zines, and exporting for print.

**Core Philosophy:**
- **First-class entities:** Every concept (project, image, sequence, template, laid-out image, page, zine) is a proper database record with full CRUD lifecycle.
- **Separation of concerns:** Templates define reusable settings; instances apply those settings to specific assets; sequences order instances.
- **Repository-backed persistence:** All data lives in SQLite; no filesystem manifests (project.json removed).
- **Service-layer orchestration:** Business logic lives in `pkg/services`; API handlers delegate to services.

---

## Target Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Workflow                           │
├─────────────────────────────────────────────────────────────────┤
│ 1. Upload images → 2. Create sequences → 3. Apply templates    │
│ 4. Review laid-out images → 5. Compose pages → 6. Build zine   │
│ 7. Export for print                                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      Application Layers                         │
├─────────────────────────────────────────────────────────────────┤
│  Web UI (React + RTK Query)                                     │
│    ↓                                                             │
│  REST API (pkg/serve)                                           │
│    ↓                                                             │
│  Services (pkg/services) ← Layout computation, workflows        │
│    ↓                                                             │
│  Repositories (pkg/repo/sqlite) ← Data access                   │
│    ↓                                                             │
│  SQLite Database + Filesystem (images on disk)                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     CLI (Glazed-based)                          │
│  zine-layout api <entity> <verb> [--flags]                      │
│  Direct calls to REST API for automation/scripting              │
└─────────────────────────────────────────────────────────────────┘
```

---

## Data Model (Phases 1 & 2)

### Entity Hierarchy

```
Project (workspace)
  ├── Assets (raw uploaded images)
  ├── Image Sequences (ordered asset lists)
  ├── Image Layout Templates (reusable resize/crop settings)
  ├── Laid Out Images (asset + template + overrides → computed layout)
  └── Layout Sequences (ordered laid-out image lists)

[Phase 3 will add:]
  ├── Page Templates (multi-image composition grids)
  ├── Laid Out Pages (template + laid-out images → printable page)
  └── Zines (ordered page collections)

[Phase 4 will add:]
  └── Zine Layout Templates (imposition schemes for print output)
```

### Entity Details

#### 1. Project
**Purpose:** Top-level workspace containing all assets and layout artifacts.

**Schema:**
```sql
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
```

**Go Type:**
```go
type Project struct {
    ID          string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Lifecycle:**
- Created via API or CLI with a name
- ID auto-generated: `prj-20251011T153045Z-a8x9mz`
- Files stored under `data/projects/{id}/images/`
- Deletion cascades to all child entities

---

#### 2. Asset
**Purpose:** Raw uploaded image with metadata (dimensions, filesize, upload date).

**Schema:**
```sql
CREATE TABLE assets (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    rel_path TEXT NOT NULL,
    content_type TEXT NOT NULL,
    bytes INTEGER NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    uploaded_at INTEGER NOT NULL,
    metadata_json TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

**Go Type:**
```go
type Asset struct {
    ID           string
    ProjectID    string
    Filename     string
    RelPath      string
    ContentType  string
    Bytes        int64
    Width        int
    Height       int
    UploadedAt   time.Time
    MetadataJSON string  // Future: EXIF data, capture date, GPS
}
```

**Storage:**
- Physical file: `data/projects/{project_id}/images/{filename}`
- Metadata record: SQLite `assets` table
- URL for serving: `/projects/{project_id}/images/{filename}`

**Lifecycle:**
- User uploads PNG via web UI or CLI
- Server generates ID: `img-20251011T153045Z-k2p7wq`
- Dimensions extracted automatically
- Deletion removes both DB record and file on disk

---

#### 3. Image Sequence
**Purpose:** Named, ordered collection of assets for organizing raw images (e.g., "Summer 2025", "Portrait Series").

**Schema:**
```sql
CREATE TABLE image_sequences (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE image_sequence_items (
    sequence_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    asset_id TEXT,
    is_gap INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (sequence_id, position),
    FOREIGN KEY (sequence_id) REFERENCES image_sequences(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE SET NULL,
    CHECK (is_gap IN (0, 1)),
    CHECK (is_gap = 1 OR asset_id IS NOT NULL)
);
```

**Go Types:**
```go
type ImageSequence struct {
    ID          string
    ProjectID   string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type ImageSequenceItem struct {
    SequenceID string
    Position   int
    AssetID    *string  // NULL if is_gap = true
    IsGap      bool
}
```

**Features:**
- **Gaps:** Placeholder positions for spreads or spacing (is_gap=true, asset_id=NULL)
- **Reordering:** Items can be rearranged via drag-and-drop or API calls
- **Deletion:** Deleting an asset sets item.asset_id to NULL (doesn't break sequence)

**Lifecycle:**
- Created with name and optional description
- Items added individually or bulk-populated from all project assets
- Reordered via `ReplaceItems(sequenceID, items)` transaction
- Used as input to batch template application

---

#### 4. Image Layout Template
**Purpose:** Reusable settings for image cropping, scaling, positioning, and export (based on `imagelayout.ViewportSettings`).

**Schema:**
```sql
CREATE TABLE image_layout_templates (
    id TEXT PRIMARY KEY,
    project_id TEXT,  -- NULL = global template
    name TEXT NOT NULL,
    description TEXT,
    settings_json TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

**Go Type:**
```go
type ImageLayoutTemplate struct {
    ID           string
    ProjectID    *string  // NULL for global templates
    Name         string
    Description  string
    SettingsJSON string   // Serialized imagelayout.ViewportSettings
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Settings Structure (JSON):**
```typescript
{
  paper_width_in: 8.0,
  paper_height_in: 10.0,
  dpi: 300,
  orientation: "portrait",
  margin_top_in: 0.25,
  margin_right_in: 0.25,
  margin_bottom_in: 0.25,
  margin_left_in: 0.25,
  // Viewport templates no longer encode spreads or gutters
  crop_ratio: null,     // or numeric (e.g., 1.5 for 3:2)
  crop_to_fill: false,
  user_scale: 1.0,
  position_x: 0,
  position_y: 0,
  units: "normalized",
  export: {
    format: "png",
    quality: 90,
    background: "white",
    out_dir: "./out",
    filename_template: "{name}-{panel}.{ext}"
  }
}
```

**Scope:**
- **Global:** `project_id IS NULL` — available to all projects (e.g., "Instagram Square 1:1")
- **Project-specific:** `project_id = 'prj-...'` — custom templates for one project

**Lifecycle:**
- Created from Book Spread Designer UI or CLI
- Captures current UI state (margins, crop, scale, position)
- Applied to assets to generate laid-out images
- Deletion restricted if referenced by laid-out images

---

#### 5. Laid Out Image
**Purpose:** Computed layout result for an asset processed through a template with optional overrides.

**Schema:**
```sql
CREATE TABLE laid_out_images (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    asset_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    overrides_json TEXT,
    result_json TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES image_layout_templates(id) ON DELETE RESTRICT
);
```

**Go Type:**
```go
type LaidOutImage struct {
    ID            string
    ProjectID     string
    AssetID       string
    TemplateID    string
    OverridesJSON *string  // Partial viewport settings to override template
    ResultJSON    string   // services.LayoutComputation
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**Result Structure (JSON):**
```typescript
{
  settings: { /* merged imagelayout.ViewportSettings */ },
  result: {
    source_rect: { x, y, w, h },  // Crop window in source pixels
    target_rect: { x, y, w, h },  // Placement inside canvas rect
    canvas_rect: { x, y, w, h },
    scale: number,
    mode: "cover" | "contain"
  },
  trace?: {
    inputs: Record<string, unknown>,
    steps: Array<{ label: string; data: Record<string, unknown> }>
  }
}
```

**Computation:**
- Service layer (`services.LayoutService.CreateLaidOutImage`) orchestrates:
  1. Fetch asset (for dimensions)
  2. Fetch template (for base settings)
  3. Merge overrides (user_scale, position adjustments)
  4. Run `simple.ComputePlacement(inputs, trace)`
  5. Serialize result as JSON
  6. Persist record

**Overrides Example:**
```json
{
  "user_scale": 1.2,
  "position_x": 0.1,
  "position_y": -0.05
}
```

**Lifecycle:**
- Created when user applies template to asset
- Can be recomputed if template or overrides change
- Used as input to layout sequences and page composition

---

#### 6. Layout Sequence
**Purpose:** Ordered collection of laid-out images for review/export workflows.

**Schema:**
```sql
CREATE TABLE layout_sequences (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE layout_sequence_items (
    sequence_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    laid_out_image_id TEXT NOT NULL,
    PRIMARY KEY (sequence_id, position),
    FOREIGN KEY (sequence_id) REFERENCES layout_sequences(id) ON DELETE CASCADE,
    FOREIGN KEY (laid_out_image_id) REFERENCES laid_out_images(id) ON DELETE CASCADE
);
```

**Go Types:**
```go
type LayoutSequence struct {
    ID          string
    ProjectID   string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type LayoutSequenceItem struct {
    SequenceID     string
    Position       int
    LaidOutImageID string
}
```

**Use Cases:**
- **Preview workflow:** See sequence of cropped/resized images before page composition
- **Export batch:** Generate final image files for all items in sequence
- **Page input:** Reference sequence items when building multi-image pages (Phase 3)

**Lifecycle:**
- Created manually or auto-generated when applying template to image sequence
- Items can be reordered without recomputation
- Deleting a laid-out image removes it from sequence (CASCADE)

---

## API Surface (REST Endpoints)

### Project Management

#### `GET /api/projects`
**Returns:** List of all projects
```json
{
  "projects": [
    {
      "id": "prj-20251011T153045Z-a8x9mz",
      "name": "Summer Photobook 2025",
      "description": "Vacation photos",
      "created_at": "2025-10-11T15:30:45Z",
      "updated_at": "2025-10-11T16:22:10Z"
    }
  ]
}
```

#### `POST /api/projects`
**Body:** `{ "name": "My Project", "description": "..." }`  
**Returns:** Created project

#### `GET /api/projects/{id}`
**Returns:** Single project details

#### `PUT /api/projects/{id}`
**Body:** `{ "name": "Updated Name", "description": "..." }`  
**Returns:** Updated project

#### `DELETE /api/projects/{id}`
**Effect:** Cascades deletion to all assets, sequences, templates, laid-out images

---

### Asset Management

#### `GET /api/projects/{id}/assets`
**Returns:** All assets in project (ordered by upload date)
```json
{
  "assets": [
    {
      "id": "img-20251011T153100Z-k2p7wq",
      "project_id": "prj-...",
      "filename": "img-20251011T153100Z-k2p7wq.png",
      "rel_path": "projects/prj-.../images/img-....png",
      "content_type": "image/png",
      "bytes": 2456789,
      "width": 4032,
      "height": 3024,
      "uploaded_at": "2025-10-11T15:31:00Z",
      "metadata": {},
      "url": "/projects/prj-.../images/img-....png"
    }
  ]
}
```

#### `POST /api/projects/{id}/images`
**Body:** Multipart form with `images[]` files  
**Returns:** Array of created assets

#### `DELETE /api/assets/{id}`
**Effect:** Deletes database record and file; sets referencing sequence items to NULL

---

### Image Sequences

#### `GET /api/projects/{id}/image-sequences`
**Returns:** All sequences in project
```json
{
  "sequences": [
    {
      "id": "seq-20251011T160000Z-xyz123",
      "project_id": "prj-...",
      "name": "Best of Summer",
      "description": "Top 20 photos for book",
      "created_at": "2025-10-11T16:00:00Z",
      "updated_at": "2025-10-11T16:15:00Z"
    }
  ]
}
```

#### `POST /api/projects/{id}/image-sequences`
**Body:** `{ "name": "Sequence Name", "description": "..." }`  
**Returns:** Created sequence (initially empty)

#### `GET /api/image-sequences/{id}`
**Returns:** Sequence metadata + items
```json
{
  "sequence": { "id": "seq-...", "name": "...", ... },
  "items": [
    { "position": 0, "asset_id": "img-...", "is_gap": false },
    { "position": 1, "asset_id": "img-...", "is_gap": false },
    { "position": 2, "asset_id": null, "is_gap": true }
  ]
}
```

#### `PUT /api/image-sequences/{id}`
**Body:** `{ "name": "New Name", "description": "..." }`  
**Returns:** Updated sequence

#### `DELETE /api/image-sequences/{id}`
**Effect:** Deletes sequence and all items

#### `GET /api/image-sequences/{id}/items`
**Returns:** Ordered list of sequence items

#### `PUT /api/image-sequences/{id}/items`
**Body:** Complete new item list (replaces all)
```json
{
  "items": [
    { "position": 0, "asset_id": "img-...", "is_gap": false },
    { "position": 1, "asset_id": null, "is_gap": true },
    { "position": 2, "asset_id": "img-...", "is_gap": false }
  ]
}
```
**Effect:** Transactional replacement of all items

#### `POST /api/image-sequences/{id}/items`
**Body:** `{ "position": 3, "asset_id": "img-...", "is_gap": false }`  
**Returns:** Added item

#### `DELETE /api/image-sequences/{id}/items/{position}`
**Effect:** Removes item at position

---

### Image Layout Templates

#### `GET /api/image-layout-templates`
**Returns:** All global templates
```json
{
  "templates": [
    {
      "id": "tmpl-20251011T120000Z-abc123",
      "project_id": null,
      "scope": "global",
      "name": "Instagram Square 1:1",
      "description": "Crop to 1:1 ratio, fill, 300 DPI",
      "settings": { "paper_width_in": 8, ... },
      "created_at": "2025-10-11T12:00:00Z",
      "updated_at": "2025-10-11T12:00:00Z"
    }
  ]
}
```

#### `GET /api/projects/{id}/image-layout-templates`
**Returns:** Global templates + project-specific templates (merged list)

#### `POST /api/projects/{id}/image-layout-templates`
**Body:**
```json
{
  "name": "Portrait 2:3",
  "description": "Vertical crop for book pages",
  "settings": {
    "paper_width_in": 8,
    "paper_height_in": 10,
    "dpi": 300,
    "crop_ratio": 0.667,
    "crop_to_fill": true,
    ...
  }
}
```
**Returns:** Created template (scoped to project)

#### `POST /api/image-layout-templates`
**Body:** Same as above  
**Returns:** Created global template (project_id = NULL)

#### `GET /api/image-layout-templates/{id}`
**Returns:** Single template with full settings

#### `PUT /api/image-layout-templates/{id}`
**Body:** `{ "name": "...", "description": "...", "settings": {...} }`  
**Returns:** Updated template

#### `DELETE /api/image-layout-templates/{id}`
**Effect:** Fails if referenced by laid-out images (RESTRICT constraint)

---

### Laid Out Images

#### `GET /api/projects/{id}/laid-out-images`
**Returns:** All laid-out images in project
```json
{
  "laid_out_images": [
    {
      "id": "loi-20251011T170000Z-def456",
      "project_id": "prj-...",
      "asset_id": "img-...",
      "template_id": "tmpl-...",
      "overrides": { "user_scale": 1.2 },
      "result": {
        "settings": { ... },
        "result": {
          "src_rect_global": { "x": 100, "y": 50, "w": 3000, "h": 2000 },
          "dst_rect_global": { "x": 0, "y": 0, "w": 2250, "h": 2850 },
          "export_single": { "w": 2250, "h": 2850 }
        },
        "placement_trace": { ... }
      },
      "created_at": "2025-10-11T17:00:00Z",
      "updated_at": "2025-10-11T17:00:00Z"
    }
  ]
}
```

#### `POST /api/projects/{id}/laid-out-images`
**Body:**
```json
{
  "asset_id": "img-...",
  "template_id": "tmpl-...",
  "overrides": {
    "user_scale": 1.1,
    "position_x": 0.05
  }
}
```
**Returns:** Created laid-out image with computed result

**Process:**
1. Fetch asset → get width/height
2. Fetch template → get base settings
3. Merge overrides → final settings
4. Run `simple.ComputePlacement(inputs, trace)`
5. Store result JSON

#### `GET /api/laid-out-images/{id}`
**Returns:** Single laid-out image with full result

#### `PUT /api/laid-out-images/{id}`
**Body:** `{ "overrides": { "user_scale": 1.3 } }`  
**Effect:** Merges new overrides, recomputes placement, updates record

#### `DELETE /api/laid-out-images/{id}`
**Effect:** Deletes record; cascades removal from layout sequences

#### `GET /api/laid-out-images/{id}/preview`
**Returns:** JSON payload suitable for client-side rendering
```json
{
  "asset_url": "/projects/prj-.../images/img-....png",
  "asset_dimensions": { "width": 4032, "height": 3024 },
  "computation": {
    "settings": { ... },
    "result": { ... }
  }
}
```
**Client responsibility:** Render the image using result coordinates

#### `GET /api/laid-out-images/{id}/export`
**Status:** 501 Not Implemented (Phase 2 stub)  
**Future:** Server-side rendering to PNG/JPEG

---

### Layout Sequences

#### `GET /api/projects/{id}/layout-sequences`
**Returns:** All layout sequences in project

#### `POST /api/projects/{id}/layout-sequences`
**Body:** `{ "name": "Final Book Order", "description": "..." }`  
**Returns:** Created sequence

#### `GET /api/layout-sequences/{id}`
**Returns:** Sequence metadata + items

#### `PUT /api/layout-sequences/{id}`
**Body:** `{ "name": "...", "description": "..." }`  
**Returns:** Updated sequence

#### `DELETE /api/layout-sequences/{id}`
**Effect:** Deletes sequence and all items

#### `GET /api/layout-sequences/{id}/items`
**Returns:** Ordered list of items
```json
{
  "items": [
    { "position": 0, "laid_out_image_id": "loi-..." },
    { "position": 1, "laid_out_image_id": "loi-..." }
  ]
}
```

#### `PUT /api/layout-sequences/{id}/items`
**Body:** Complete new item list (transactional replace)

#### `POST /api/layout-sequences/{id}/items`
**Body:** `{ "position": 5, "laid_out_image_id": "loi-..." }`

#### `DELETE /api/layout-sequences/{id}/items/{position}`

---

## Core Processes

### Process 1: Image Upload & Organization

**Actors:** User, Web UI, Server, Repositories

**Flow:**
1. User selects project in UI
2. Drags/drops PNG files into upload zone
3. Frontend sends `POST /api/projects/{id}/images` with multipart form
4. Server handler:
   - Calls `projects.SavePNGImage()` → writes file to disk, extracts dimensions
   - Creates `repo.Asset` record with metadata
   - Returns asset response with URL
5. Frontend invalidates `Asset` cache tag → UI refreshes
6. User can:
   - View assets in gallery
   - Create image sequence
   - Add assets to sequence via drag-and-drop
   - Reorder sequence items

**Validation:**
- Only PNG files accepted
- Filesize limits enforced (64MB)
- Dimensions extracted automatically
- Duplicate filenames get unique IDs

---

### Process 2: Template Creation & Application

**Actors:** User, Book Spread Designer UI, Server, LayoutService

**Flow (Template Creation):**
1. User opens Book Spread Designer
2. Selects project and asset
3. Adjusts settings (paper size, margins, crop ratio, scale, position)
4. Clicks "Save as Template"
5. Enters name and description
6. Frontend sends `POST /api/projects/{id}/image-layout-templates`
   ```json
   {
     "name": "8x10 Portrait Crop",
     "settings": { /* current bookSpreadSlice state */ }
   }
   ```
7. Server creates template record
8. Template appears in template list (available for reuse)

**Flow (Template Application):**
1. User selects asset and template in UI
2. Optionally tweaks overrides (zoom, nudge position)
3. Clicks "Apply Template"
4. Frontend sends `POST /api/projects/{id}/laid-out-images`
   ```json
   {
     "asset_id": "img-...",
     "template_id": "tmpl-...",
     "overrides": { "user_scale": 1.2 }
   }
   ```
5. Server orchestrates via `LayoutService.CreateLaidOutImage()`:
   - Fetches asset → `{width: 4032, height: 3024}`
   - Fetches template → base settings
   - Merges overrides → final settings
   - Builds `simple.Inputs` from settings + asset dimensions
   - Calls `simple.ComputePlacement(inputs, trace)`
   - Result contains crop/scale/position geometry
   - Serializes `LayoutComputation` to JSON
   - Persists `LaidOutImage` record
6. Returns laid-out image with result
7. Frontend caches result for preview rendering

**Batch Application (Image Sequence → Template):**
1. User selects image sequence (e.g., 50 photos)
2. Selects template
3. Clicks "Apply to All"
4. Frontend sends `POST /api/projects/{id}/laid-out-images/batch`
   ```json
   {
     "sequence_id": "seq-...",
     "template_id": "tmpl-..."
   }
   ```
5. Service calls `ApplyTemplateToSequence()`:
   - Fetches sequence items
   - For each non-gap asset:
     - Creates laid-out image (template + asset)
   - Returns array of created records
6. Optionally creates corresponding layout sequence

---

### Process 3: Layout Review & Export

**Actors:** User, Layout Sequence Editor, Preview Service

**Flow (Preview):**
1. User selects layout sequence in UI
2. Frontend fetches sequence items
3. For each item, frontend requests `GET /api/laid-out-images/{id}/preview`
4. Server returns computation payload (settings + result + asset URL)
5. Frontend renders using Canvas 2D:
   ```typescript
   const { src_rect_global, dst_rect_global } = result;
   ctx.drawImage(
     img,
     src_rect_global.x, src_rect_global.y,
     src_rect_global.w, src_rect_global.h,
     dst_rect_global.x, dst_rect_global.y,
     dst_rect_global.w, dst_rect_global.h
   );
   ```
6. User can:
   - Navigate through sequence (prev/next)
   - Reorder items
   - Edit individual overrides and recompute

**Flow (Export - Future):**
1. User clicks "Export All" on layout sequence
2. Frontend sends `POST /api/layout-sequences/{id}/export`
3. Server renders each laid-out image to file
4. Returns ZIP archive for download

---

### Process 4: Algorithm Execution (Image Layout Engine)

**Input:** `imagelayout/engine.Inputs` (derived from `imagelayout.ViewportSettings` + image metadata)

**Algorithm Location:** `pkg/imagelayout/engine/engine.go :: ComputeViewport()`

**Steps:**
1. **Canvas Setup:** Convert paper size + DPI + margins → pixel dimensions
   - Apply orientation (swap width/height for landscape)
   - Compute content rectangle after subtracting margins

2. **Source Cropping:** Determine crop window in source image
   - Honor explicit `crop_ratio` when present
   - Fall back to cover/contain behaviour based on `crop_to_fill`
   - Apply anchor/drag offsets according to `units`

3. **Scale & Placement:** Fit cropped region into the content rectangle
   - Compute cover/contain scale and apply `user_scale`
   - Translate to final destination within the content rectangle

4. **Result Assembly:** Emit source, target, and canvas rectangles plus scale/mode metadata
   - Persist trace steps for debugging (inputs snapshot + crop/scale data)

**Output:** `imagelayout.ViewportResult` with optional `imagelayout.Trace` for debugging

**Trace Structure:**
```json
{
  "inputs": { "source_w": 4032, "source_h": 3024, ... },
  "steps": [
    { "label": "crop", "data": { "sx": 120, "sw": 3600 } },
    { "label": "scale", "data": { "mode": "cover", "scale": 1.42 } },
    { "label": "result", "data": { "target_rect": { "x": 48, "y": 32, "w": 2800, "h": 2100 } } }
  ]
}
```

---

## CLI Interface

All commands follow Glazed framework patterns with structured output.

### Command Groups

#### `zine-layout api projects [verb]`
- `list` — List all projects
- `get --id <id>` — Show project details
- `create --name <name>` — Create new project
- `delete --id <id>` — Delete project

**Example:**
```bash
$ zine-layout api projects list --output table
ID                             NAME                  CREATED_AT
prj-20251011T153045Z-a8x9mz    Summer Photobook      2025-10-11 15:30:45
prj-20251011T140000Z-xyz123    Winter Zine           2025-10-11 14:00:00

$ zine-layout api projects create --name "New Book" --output json
{"id":"prj-20251011T180000Z-qwerty","name":"New Book",...}
```

#### `zine-layout api assets [verb]`
- `list --project-id <id>` — List project assets
- `delete --id <id>` — Delete asset

#### `zine-layout api image-sequences [verb]`
- `list --project-id <id>`
- `get --id <id>` (includes items)
- `create --project-id <id> --name <name>`
- `update --id <id> --name <name>`
- `delete --id <id>`
- `add-item --sequence-id <id> --asset-id <id> --position <n>`
- `reorder --sequence-id <id> --items <json>`
- `delete-item --sequence-id <id> --position <n>`

**Example:**
```bash
$ zine-layout api image-sequences create \
    --project-id prj-... \
    --name "Summer Best" \
    --output json

$ zine-layout api image-sequences add-item \
    --sequence-id seq-... \
    --asset-id img-... \
    --position 0
```

#### `zine-layout api image-layout-templates [verb]`
- `list` (global)
- `list --project-id <id>` (global + project)
- `get --id <id>`
- `create --name <name> --settings <json>`
- `update --id <id> --name <name> --settings <json>`
- `delete --id <id>`

#### `zine-layout api laid-out-images [verb]`
- `list --project-id <id>`
- `get --id <id>`
- `create --project-id <id> --asset-id <id> --template-id <id> [--overrides <json>]`
- `update --id <id> --overrides <json>` (recomputes)
- `delete --id <id>`
- `preview --id <id>` (outputs preview JSON)

**Example:**
```bash
$ zine-layout api laid-out-images create \
    --project-id prj-... \
    --asset-id img-... \
    --template-id tmpl-... \
    --overrides '{"user_scale":1.2}' \
    --output json
```

#### `zine-layout api layout-sequences [verb]`
- `list --project-id <id>`
- `get --id <id>` (includes items)
- `create --project-id <id> --name <name>`
- `update --id <id> --name <name>`
- `delete --id <id>`
- `add-item --sequence-id <id> --laid-out-image-id <id> --position <n>`
- `reorder --sequence-id <id> --items <json>`
- `delete-item --sequence-id <id> --position <n>`

---

## Frontend Components (Phase 1 & 2)

### Views

#### `/projects` — Projects.tsx
- Grid of project cards
- Create new project button
- Navigate to project detail

#### `/projects/:id` — ProjectDetail.tsx
- **Asset Gallery:** Upload, view, delete images
- **Image Sequences:** List, create, edit sequences
  - Drag-and-drop asset ordering
  - Gap insertion
  - Slideshow preview
- **Layout Templates:** Browse available templates
- **Quick Actions:** Apply template to sequence

#### `/projects/:id/templates` — LayoutTemplateManager.tsx
- List global + project templates
- Create template from current settings
- Edit template settings
- Preview template on sample asset
- Delete template (if not in use)

#### `/projects/:id/laid-out-images` — LaidOutImageViewer.tsx
- Grid of laid-out images
- Preview panel showing cropped/scaled result
- Edit overrides (zoom, position nudges)
- Recompute button
- Delete laid-out image

#### `/projects/:id/layout-sequences` — LayoutSequenceEditor.tsx
- List layout sequences
- Create new sequence
- Drag-and-drop laid-out images
- Preview sequence as slideshow
- Export options

---

## Service Layer (pkg/services)

### LayoutService

**Location:** `pkg/services/layout.go`

**Responsibilities:**
- Template + asset → laid-out image computation
- Batch application to sequences
- Settings merging and validation

**Key Methods:**

```go
func (s *LayoutService) CreateLaidOutImage(
    projectID, assetID, templateID string,
    overridesJSON *string,
) (*repo.LaidOutImage, error)
```
- Fetches asset and template
- Merges settings + overrides
- Runs `simple.ComputePlacement()`
- Persists result

```go
func (s *LayoutService) RecomputeLaidOutImage(
    record *repo.LaidOutImage,
) error
```
- Updates existing record with fresh computation
- Used when template or overrides change

```go
func (s *LayoutService) ApplyTemplateToSequence(
    projectID, sequenceID, templateID string,
) ([]*repo.LaidOutImage, error)
```
- Batch operation: creates laid-out image for each asset in sequence
- Skips gaps
- Returns array of created records

---

## Repository Layer (pkg/repo)

### Interfaces

All repositories follow standard CRUD pattern:
- `Create(entity)` → insert with auto-generated ID
- `Get(id)` → fetch single entity
- `List(parentID)` → fetch all children
- `Update(entity)` → modify existing
- `Delete(id)` → remove (cascade/restrict per schema)

**Sequence-specific:**
- `AddItem(item)` → append to sequence
- `ListItems(sequenceID)` → ordered items
- `ReplaceItems(sequenceID, items)` → transactional reorder
- `DeleteItem(sequenceID, position)` → remove one item

**Template-specific:**
- `ListGlobal()` → templates with project_id IS NULL
- `ListByProject(projectID)` → project-specific templates

---

## Data Flow Diagrams

### Upload → Sequence → Template → Layout Sequence

```
┌─────────┐
│  User   │
└────┬────┘
     │ 1. Upload PNGs
     ↓
┌────────────────┐     ┌──────────────┐
│ POST /images   │ ──→ │ Asset Table  │
└────────────────┘     └──────┬───────┘
                              │
     ┌────────────────────────┘
     │ 2. Create sequence
     ↓
┌──────────────────────┐     ┌─────────────────────┐
│ POST /image-sequences│ ──→ │ ImageSequence Table │
└──────────────────────┘     └──────┬──────────────┘
                                    │
     ┌──────────────────────────────┘
     │ 3. Add items to sequence
     ↓
┌────────────────────────┐     ┌────────────────────────┐
│ PUT /sequences/../items│ ──→ │ ImageSequenceItem Table│
└────────────────────────┘     └────────────────────────┘
                                         │
     ┌───────────────────────────────────┘
     │ 4. Create/select template
     ↓
┌──────────────────────────────┐     ┌────────────────────────┐
│ POST /image-layout-templates │ ──→ │ LayoutTemplate Table   │
└──────────────────────────────┘     └──────┬─────────────────┘
                                            │
     ┌──────────────────────────────────────┴─────────┐
     │ 5. Apply template to asset (or batch to sequence)
     ↓                                                 
┌──────────────────────┐     ┌──────────────┐     ┌──────────────┐
│ POST /laid-out-images│ ──→ │LayoutService │ ──→ │ComputePlace- │
└──────────────────────┘     └──────┬───────┘     │    ment()    │
                                    │             └──────┬───────┘
                                    ↓                    │
                            ┌─────────────────┐         │
                            │ LaidOutImage    │ ←───────┘
                            │ Table (w/result)│
                            └────────┬────────┘
                                     │
     ┌───────────────────────────────┘
     │ 6. Create layout sequence
     ↓
┌───────────────────────┐     ┌────────────────────┐
│ POST /layout-sequences│ ──→ │LayoutSequence Table│
└───────────────────────┘     └────────────────────┘
                                     │
     ┌───────────────────────────────┘
     │ 7. Add laid-out images
     ↓
┌────────────────────────────┐     ┌──────────────────────────┐
│ PUT /.../items             │ ──→ │LayoutSequenceItem Table  │
└────────────────────────────┘     └──────────────────────────┘
```

### Computation Detail Flow

```
Asset (4032×3024px)  +  Template (8×10", 300dpi, crop 2:3, fill)
          │                           │
          └───────────┬───────────────┘
                      ↓
          ┌─────────────────────────┐
          │ mergeTemplateSettings() │
          └─────────┬───────────────┘
                    ↓
          Final Settings (merged with overrides)
                    ↓
          ┌─────────────────────────┐
          │ InputsFromSettings()    │  Convert inches→px, build Inputs
          └─────────┬───────────────┘
                    ↓
          simple.Inputs {
            SrcW: 4032, SrcH: 3024,
            PaperWIn: 8, PaperHIn: 10,
            DPI: 300,
            CropRatio: {W: 2, H: 3},
            CropToFill: true,
            ...
          }
                    ↓
          ┌─────────────────────────┐
          │ ComputePlacement()      │  5-step algorithm
          └─────────┬───────────────┘
                    ↓
          simple.Result {
            SrcRectGlobal: {x:516, y:0, w:3000, h:3024},
            DstRectGlobal: {x:0, y:0, w:2250, h:2850},
            ExportSingle: {W:2250, H:2850}
          }
                    ↓
          ┌─────────────────────────┐
          │ Serialize to JSON       │
          └─────────┬───────────────┘
                    ↓
          LaidOutImage.ResultJSON (stored in DB)
```

---

## Repository Implementation Notes

### Transaction Patterns

**Reorder operations use transactions:**
```go
func (r *imageSequenceRepo) ReplaceItems(sequenceID string, items []*ImageSequenceItem) error {
    tx, _ := r.db.Begin()
    defer tx.Rollback()
    
    // 1. Delete all existing items
    tx.Exec(`DELETE FROM image_sequence_items WHERE sequence_id = ?`, sequenceID)
    
    // 2. Insert new items
    for _, item := range items {
        tx.Exec(`INSERT INTO image_sequence_items ...`, ...)
    }
    
    return tx.Commit()
}
```

**Why:** Ensures atomic updates; no partial states where positions are duplicated or missing.

### Foreign Key Constraints

- **CASCADE:** `assets.project_id`, `sequences.project_id` → deleting project cleans up all children
- **SET NULL:** `image_sequence_items.asset_id` → deleting asset doesn't break sequence (gap created)
- **RESTRICT:** `laid_out_images.template_id` → can't delete template if in use (prevents orphaned layouts)

### ID Generation

**Pattern:** `{prefix}-{timestamp}-{random}`
- `prj-20251011T153045Z-a8x9mz` (project)
- `img-20251011T153100Z-k2p7wq` (asset)
- `seq-20251011T160000Z-xyz123` (sequence)
- `tmpl-20251011T120000Z-abc123` (template)
- `loi-20251011T170000Z-def456` (laid-out image)

**Implementation:**
```go
func generateID(prefix string) string {
    ts := time.Now().UTC().Format("20060102T150405Z")
    const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, 6)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return fmt.Sprintf("%s-%s-%s", prefix, ts, string(b))
}
```

**Benefits:**
- Chronological sorting by ID
- No collisions (timestamp + random suffix)
- Human-readable timestamps for debugging

---

## Viewport Settings Schema (`imagelayout.ViewportSettings`)

**Current Fields (Phase 2):**
```go
type ViewportSettings struct {
    PaperWidthIn  float64
    PaperHeightIn float64
    DPI           float64
    Orientation   string // "portrait" | "landscape"

    MarginTopIn    float64
    MarginRightIn  float64
    MarginBottomIn float64
    MarginLeftIn   float64

    CropRatio  *float64
    CropToFill bool

    UserScale float64
    PositionX float64
    PositionY float64
    Units     string // "normalized" | "px"

    Export ExportOptions
}

type ExportOptions struct {
    Format           string
    Quality          int
    Background       string
    OutDir           string
    FilenameTemplate string
}
```

**Future Extensions:**
- Rich crop descriptors (focus points, aspect presets)
- Alternate coordinate systems (percent, relative to short edge)
- Output directives for downstream page composition (e.g., logical slot hints)

---

## Filesystem Layout

```
data/
├── zine-layout.db          # SQLite database (all metadata)
├── zine-layout.db-shm      # Shared memory (WAL mode)
├── zine-layout.db-wal      # Write-ahead log
└── projects/
    ├── prj-20251011T153045Z-a8x9mz/
    │   ├── images/
    │   │   ├── img-20251011T153100Z-k2p7wq.png
    │   │   ├── img-20251011T153105Z-m8n4tz.png
    │   │   └── img-20251011T153110Z-p9q5rx.png
    │   └── renders/              # Phase 3: page render outputs
    │       └── 20251011-170000/
    │           ├── page-001.png
    │           └── page-002.png
    └── prj-20251011T140000Z-xyz123/
        └── images/ ...
```

**Key Points:**
- Database is single source of truth for all metadata
- Images stored by auto-generated ID (not original filename)
- Renders cached on disk for export workflows
- No `project.json` files (removed in Phase 1)

---

## Security & Validation

### Input Validation

**Server-side checks:**
- File uploads: PNG only, max 64MB
- Required fields: project name, asset ID, template ID
- Foreign key integrity: server returns 404 if referenced entity doesn't exist
- JSON schema: settings validated by `simple.InputsFromSettings()`

### Path Traversal Prevention

**All file paths sanitized:**
```go
filepath.Join(projectsRoot, projectID, "images", filepath.Base(filename))
```
**Effect:** `../../etc/passwd` becomes `etc/passwd` (relative to project dir)

### Cascade Policies

- Deleting project → cascades to all assets, sequences, templates, laid-out images
- Deleting asset → sets sequence items to NULL (preserves sequence structure)
- Deleting template → blocked if referenced by laid-out images (RESTRICT)

---

## Performance Characteristics

### Query Patterns

**Optimized indexes:**
- `idx_assets_project` → fast listing by project
- `idx_laid_out_images_asset` → find all layouts for one asset
- `idx_sequences_project` → project sequence listing
- Position-based primary keys on items → ordered traversal

**Caching Strategy (Frontend):**
- RTK Query caches API responses (30s default)
- Invalidation tags ensure freshness after mutations
- Preview images rendered client-side from cached computation

### Scalability Limits

**Current constraints:**
- SQLite single-writer (WAL mode mitigates read contention)
- Asset storage: filesystem-based (no CDN/blob storage yet)
- Preview rendering: client-side Canvas 2D (no server rendering in Phase 2)
- Batch operations: synchronous (no background jobs)

**Sufficient for:**
- ~1000 projects
- ~10,000 assets per project
- ~100 sequences per project
- Real-time UI updates with <100ms latency

---

## Remaining Work (Phase 3 & 4)

### Phase 3: Page Templates + Laid Out Pages + Zines

**Missing entities:**
- `page_templates` table (ZineLayout DSL storage)
- `laid_out_pages` table (rendered page metadata)
- `laid_out_page_inputs` table (image-to-page mappings)
- `zines` table
- `zine_pages` table

**Missing processes:**
- Multi-image page composition using `pkg/zinelayout`
- Zine assembly workflows
- Server-side page rendering

### Phase 4: Zine Export & Imposition

**Missing entities:**
- `zine_layout_templates` table (folding schemes)

**Missing processes:**
- Imposition algorithms (8-page fold, 16-page booklet)
- PDF generation with crop marks
- Print-ready ZIP export

---

## Key Design Decisions

### Why Separate Image Sequences from Layout Sequences?

**Rationale:**
- Image sequences order *raw assets* (user's original organization)
- Layout sequences order *laid-out images* (computed results ready for pages)
- Allows user to experiment with different templates on same image sequence
- Prevents tight coupling between raw image ordering and final output ordering

**Example Workflow:**
1. Create image sequence "Best 20 Summer Photos"
2. Apply template "8×10 Portrait" → creates 20 laid-out images
3. Create layout sequence "Book Order v1" with those 20 images
4. Later: apply different template "Square 1:1" → creates new 20 laid-out images
5. Create layout sequence "Book Order v2 (square variant)"
6. Compare both sequences; choose winner

### Why Store Computation Results?

**Rationale:**
- **Performance:** Avoid re-running placement algorithm on every preview
- **Consistency:** Locked-in geometry survives template changes
- **Traceability:** Placement trace captures inputs for debugging
- **Offline rendering:** Server can render without re-computing (Phase 3+)

**Trade-off:**
- Storage overhead (~2KB JSON per laid-out image)
- Requires explicit recompute when template changes
- **Accepted:** Storage is cheap; computation determinism is valuable

### Why Template Scopes (Global vs. Project)?

**Rationale:**
- **Global templates:** Reusable across all projects (e.g., "Instagram 1:1", "Standard 8×10")
- **Project templates:** Experiment within one project without polluting global namespace
- **Discoverability:** Project templates appear alongside globals in UI

**Implementation:**
- `project_id IS NULL` → global
- `project_id = 'prj-...'` → project-scoped
- List endpoints merge both for convenience

---

## Error Handling Patterns

### Repository Layer
```go
if err := r.db.QueryRow(...).Scan(&result); err != nil {
    return nil, errNotFound(err)  // Converts sql.ErrNoRows to semantic error
}
```

### Service Layer
```go
if asset.ProjectID != projectID {
    return nil, fmt.Errorf("asset %s does not belong to project %s", assetID, projectID)
}
```

### API Layer
```go
if errors.Is(err, sql.ErrNoRows) {
    http.NotFound(w, r)
    return
}
respondError(w, http.StatusInternalServerError, err.Error())
```

### Frontend
```typescript
const { data, error, isLoading } = useGetAssetsQuery({ projectId });

if (error) {
  return <div>Failed to load assets</div>;
}
```

**Pattern:** Errors bubble up with context; handlers decide HTTP status; frontend shows user-friendly messages.

---

## Testing Strategy

### Repository Tests
**File:** `pkg/repo/sqlite/*_test.go`

**Pattern:**
```go
func TestImageSequenceRepo(t *testing.T) {
    db, _ := sql.Open("sqlite", ":memory:")
    repos, _ := NewRepositories(db)
    
    // Create
    seq := &repo.ImageSequence{ProjectID: "prj-test", Name: "Test"}
    repos.ImageSequences.Create(seq)
    
    // Get
    fetched, err := repos.ImageSequences.Get(seq.ID)
    assert.NoError(t, err)
    assert.Equal(t, "Test", fetched.Name)
    
    // List
    all, _ := repos.ImageSequences.ListByProject("prj-test")
    assert.Len(t, all, 1)
}
```

### Integration Tests
**File:** `pkg/serve/integration_test.go`

**Pattern:** HTTP requests → in-memory SQLite → validate responses

### CLI Smoke Tests
**File:** `cmd/zine-layout/cmds/api/phase1_test.sh`

**Pattern:** Shell script calling CLI commands, validating outputs

---

## Deployment Configuration

### Server Startup
```bash
zine-layout serve \
  --addr :8088 \
  --data-root ./data \
  --web-root ./dist
```

**Flags:**
- `--addr` — HTTP listen address
- `--data-root` — SQLite DB + project files location
- `--web-root` — Built React app bundle (index.html, assets/)

### Database Initialization
1. Server calls `initDatabase()` on startup
2. Opens SQLite connection with WAL mode
3. Executes `schemaSQL` (idempotent; uses `CREATE TABLE IF NOT EXISTS`)
4. Instantiates all repositories
5. Ready to serve requests

**Location:** `pkg/serve/server.go :: prepare()`

---

## Frontend State Management

### Redux Store Structure

```typescript
{
  api: {
    // RTK Query cache
    queries: {
      'getProjects(undefined)': { data: [...], status: 'fulfilled' },
      'getAssets({"projectId":"prj-..."})': { data: [...] },
      ...
    },
    mutations: { ... }
  }
}
```

**No separate Redux slices needed:** RTK Query manages all server state.

**Local component state only for:**
- Form inputs before submission
- UI toggles (modals, dropdowns)
- Drag-and-drop temporary states

---

## Algorithm Integration (Image Layout Engine)

### Current Implementation Status

**Supported Capabilities:**
- ✅ Page viewport with margin handling
- ✅ Cover / contain behaviour with optional crop ratio overrides
- ✅ Normalized and pixel-based positioning
- ✅ User scale multiplier for fine adjustments

**Not Yet Supported (from 01-algorithm-for-resizing.md):**
- ❌ Dedicated crop presets (e.g. 4:5 portrait) expressed as named tokens
- ❌ Fit-to-width / fit-to-height helpers that ignore page margins entirely
- ❌ Anchor presets (9-point grid) for quick alignment
- ❌ Focus point heuristic for intelligent cropping
- ❌ Dual-page (spread) viewport splitting — deferred to page layout phase

**Roadmap:** Extend `imagelayout/engine` to cover the outstanding modes and expose ergonomic helpers in the DSL/UI.

### Extension Plan

**Engine TODOs:**
- Introduce layout modes (`page`, `crop`, `fit`) with dedicated helper structs
- Add anchor presets + focus point math for intuitive positioning
- Surface validation errors that map cleanly back to DSL form fields
- Plug render helpers into future export workflow (image and PDF rendering)

---

## API Route Summary (Phase 1 & 2)

### Project Routes
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/projects` | List all projects |
| POST | `/api/projects` | Create project |
| GET | `/api/projects/{id}` | Get project |
| PUT | `/api/projects/{id}` | Update project |
| DELETE | `/api/projects/{id}` | Delete project |

### Asset Routes
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/projects/{id}/assets` | List project assets |
| POST | `/api/projects/{id}/images` | Upload images (multipart) |
| GET | `/api/assets/{id}` | Get asset metadata |
| DELETE | `/api/assets/{id}` | Delete asset + file |

### Image Sequence Routes
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/projects/{id}/image-sequences` | List sequences |
| POST | `/api/projects/{id}/image-sequences` | Create sequence |
| GET | `/api/image-sequences/{id}` | Get sequence + items |
| PUT | `/api/image-sequences/{id}` | Update sequence |
| DELETE | `/api/image-sequences/{id}` | Delete sequence |
| GET | `/api/image-sequences/{id}/items` | List items |
| PUT | `/api/image-sequences/{id}/items` | Replace all items |
| POST | `/api/image-sequences/{id}/items` | Add item |
| DELETE | `/api/image-sequences/{id}/items/{pos}` | Delete item |

### Image Layout Template Routes
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/image-layout-templates` | List global templates |
| POST | `/api/image-layout-templates` | Create global template |
| GET | `/api/projects/{id}/image-layout-templates` | List global + project templates |
| POST | `/api/projects/{id}/image-layout-templates` | Create project template |
| GET | `/api/image-layout-templates/{id}` | Get template |
| PUT | `/api/image-layout-templates/{id}` | Update template |
| DELETE | `/api/image-layout-templates/{id}` | Delete template |

### Laid Out Image Routes
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/projects/{id}/laid-out-images` | List laid-out images |
| POST | `/api/projects/{id}/laid-out-images` | Create laid-out image |
| GET | `/api/laid-out-images/{id}` | Get laid-out image |
| PUT | `/api/laid-out-images/{id}` | Update overrides, recompute |
| DELETE | `/api/laid-out-images/{id}` | Delete laid-out image |
| GET | `/api/laid-out-images/{id}/preview` | Get preview payload |
| GET | `/api/laid-out-images/{id}/export` | Export rendered image (stub) |

### Layout Sequence Routes
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/projects/{id}/layout-sequences` | List sequences |
| POST | `/api/projects/{id}/layout-sequences` | Create sequence |
| GET | `/api/layout-sequences/{id}` | Get sequence + items |
| PUT | `/api/layout-sequences/{id}` | Update sequence |
| DELETE | `/api/layout-sequences/{id}` | Delete sequence |
| GET | `/api/layout-sequences/{id}/items` | List items |
| PUT | `/api/layout-sequences/{id}/items` | Replace all items |
| POST | `/api/layout-sequences/{id}/items` | Add item |
| DELETE | `/api/layout-sequences/{id}/items/{pos}` | Delete item |

**Total:** 42 endpoints implemented (Phase 1 & 2)

---

## CLI Command Summary

### Structure
```
zine-layout api <entity> <verb> [--flags]
```

### Implemented Commands (Phase 1 & 2)

**Projects:**
- `projects list`
- `projects get --id <id>`
- `projects create --name <name>`
- `projects delete --id <id>`

**Assets:**
- `assets list --project-id <id>`
- `assets delete --id <id>`

**Image Sequences:**
- `image-sequences list --project-id <id>`
- `image-sequences get --id <id>`
- `image-sequences create --project-id <id> --name <name>`
- `image-sequences update --id <id> --name <name>`
- `image-sequences delete --id <id>`
- `image-sequences add-item --sequence-id <id> --asset-id <id> --position <n>`
- `image-sequences reorder --sequence-id <id> --items <json>`
- `image-sequences delete-item --sequence-id <id> --position <n>`

**Image Layout Templates:**
- `image-layout-templates list [--project-id <id>]`
- `image-layout-templates get --id <id>`
- `image-layout-templates create --name <name> --settings <json> [--project-id <id>]`
- `image-layout-templates update --id <id> --name <name> --settings <json>`
- `image-layout-templates delete --id <id>`

**Laid Out Images:**
- `laid-out-images list --project-id <id>`
- `laid-out-images get --id <id>`
- `laid-out-images create --project-id <id> --asset-id <id> --template-id <id> [--overrides <json>]`
- `laid-out-images update --id <id> --overrides <json>`
- `laid-out-images delete --id <id>`
- `laid-out-images preview --id <id>`

**Layout Sequences:**
- `layout-sequences list --project-id <id>`
- `layout-sequences get --id <id>`
- `layout-sequences create --project-id <id> --name <name>`
- `layout-sequences update --id <id> --name <name>`
- `layout-sequences delete --id <id>`
- `layout-sequences add-item --sequence-id <id> --laid-out-image-id <id> --position <n>`
- `layout-sequences reorder --sequence-id <id> --items <json>`
- `layout-sequences delete-item --sequence-id <id> --position <n>`

**Total:** 38 CLI commands operational

---

## Frontend Component Inventory (Phase 1 & 2)

### Views

| Component | Route | Purpose |
|-----------|-------|---------|
| `Projects.tsx` | `/projects` | Project gallery, create new |
| `ProjectDetail.tsx` | `/projects/:id` | Asset management, sequence editing |
| `LayoutTemplateManager.tsx` | `/projects/:id/templates` | Browse/create/edit templates |
| `LaidOutImageViewer.tsx` | `/projects/:id/laid-out-images` | Grid of computed layouts |
| `LayoutSequenceEditor.tsx` | `/projects/:id/layout-sequences` | Order laid-out images |

### Reusable Components

| Component | Location | Purpose |
|-----------|----------|---------|
| `ProjectAssetsPanel.tsx` | `components/` | Upload, gallery, drag-and-drop |
| `ImageSequencePanel.tsx` | `components/` | Sequence list, item editor |
| `TemplatePreview.tsx` | `components/` | Canvas preview of layout |
| `Card, Button, Input` | `components/ui/` | Reusable UI primitives |

---

## Example User Journey (Complete Phase 1 & 2 Workflow)

### Day 1: Upload & Organize
1. User creates project "Summer Zine 2025"
2. Uploads 50 PNG files via web UI
3. Creates sequence "Best 20" 
4. Drags 20 favorite assets into sequence
5. Inserts gap at position 10 for spread break
6. Reorders via drag-and-drop

### Day 2: Template Creation
1. Opens Book Spread Designer
2. Selects first asset from sequence
3. Configures: 8×10" paper, 0.25" margins, crop to 2:3, fill
4. Previews in real-time
5. Saves as template "Portrait 2:3"
6. Repeats for square variant: saves as "Square 1:1"

### Day 3: Layout Application
1. Navigates to Laid Out Images view
2. Clicks "Apply Template to Sequence"
3. Selects "Best 20" sequence + "Portrait 2:3" template
4. Server creates 19 laid-out images (skips 1 gap)
5. User reviews grid of previews
6. Tweaks one image: adjusts `user_scale: 1.15`
7. Clicks recompute → updates that one record

### Day 4: Final Sequence
1. Creates layout sequence "Final Book Order"
2. Adds all 19 laid-out images
3. Reorders to taste
4. Exports preview slideshow
5. **[Phase 3]:** Composes pages, builds zine
6. **[Phase 4]:** Applies 8-page fold template, exports for print

---

## Migration Notes (From Old System)

### What Changed
- ❌ **Removed:** `project.json` files, filesystem-based project manifests
- ❌ **Removed:** `pkg/projects` CRUD functions (now in repositories)
- ✅ **Added:** SQLite as primary data store
- ✅ **Added:** First-class sequences, templates, laid-out images
- ✅ **Added:** Service layer for workflow orchestration

### Breaking Changes
- Old projects must be re-imported (no automatic migration)
- YAML specs from old system won't load (schema changed)
- CLI commands have new syntax (`api <entity> <verb>` vs. old flat structure)

### Deprecation Timeline
- **Phase 1-2:** Clean break; database reset acceptable
- **Phase 3-4:** Consider migration tools if user data exists
- **Post-launch:** Semantic versioning for schema changes

---

## Future Enhancements (Post-Phase 4)

### Potential Features
- **Batch rendering:** Background job queue for large sequences
- **Cloud storage:** S3/CDN integration for asset hosting
- **Collaborative editing:** Multi-user projects with conflict resolution
- **Version control:** Snapshot/restore project state
- **AI cropping:** Auto-detect faces/subjects for smart crop windows
- **PDF output:** Direct-to-PDF rendering with embedded fonts
- **Print services:** Integration with print-on-demand APIs

### Scalability Path
- **Horizontal scaling:** Read replicas for SQLite (if needed)
- **Caching layer:** Redis for computed layouts
- **Microservices:** Split rendering to dedicated workers
- **CDN:** Serve static assets from edge locations

---

## Glossary

| Term | Definition |
|------|------------|
| **Asset** | Raw uploaded image (original file) |
| **Image Sequence** | Ordered list of assets (for organization) |
| **Template** | Reusable crop/scale/margin settings |
| **Laid-Out Image** | Computed layout result (asset + template) |
| **Layout Sequence** | Ordered list of laid-out images |
| **Page Template** | Multi-image grid composition (Phase 3) |
| **Laid-Out Page** | Rendered page (template + images, Phase 3) |
| **Zine** | Complete book (ordered pages, Phase 3) |
| **Imposition** | Print layout scheme (folding/booklet, Phase 4) |
| **Image Layout Engine** | Placement computation engine (`pkg/imagelayout/engine`) |
| **Spread** | Two facing pages in a book |
| **Gutter** | Center binding area between pages |
| **Cover** | Scale mode that fills target (may crop) |
| **Contain** | Scale mode that fits entire image (may letterbox) |

---

## Appendix: Sample API Requests

### Create Project and Upload Images
```bash
# Create project
curl -X POST http://localhost:8088/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"My Zine","description":"Test project"}' \
  | jq -r '.project.id'
# → prj-20251011T180000Z-abc123

# Upload images
curl -X POST http://localhost:8088/api/projects/prj-.../images \
  -F "images[]=@photo1.png" \
  -F "images[]=@photo2.png"
```

### Create Sequence and Add Items
```bash
# Create sequence
curl -X POST http://localhost:8088/api/projects/prj-.../image-sequences \
  -H "Content-Type: application/json" \
  -d '{"name":"Summer Photos"}' \
  | jq -r '.sequence.id'
# → seq-20251011T180100Z-xyz789

# Add items
curl -X PUT http://localhost:8088/api/image-sequences/seq-.../items \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"position":0, "asset_id":"img-...", "is_gap":false},
      {"position":1, "asset_id":"img-...", "is_gap":false}
    ]
  }'
```

### Create Template and Apply
```bash
# Create template
curl -X POST http://localhost:8088/api/image-layout-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "8x10 Portrait",
    "settings": {
      "paper_width_in": 8,
      "paper_height_in": 10,
      "dpi": 300,
      "crop_ratio": 0.667,
      "crop_to_fill": true
    }
  }' \
  | jq -r '.template.id'

# Apply to asset
curl -X POST http://localhost:8088/api/projects/prj-.../laid-out-images \
  -H "Content-Type: application/json" \
  -d '{
    "asset_id": "img-...",
    "template_id": "tmpl-...",
    "overrides": {"user_scale": 1.1}
  }'
```

---

## Conclusion

**Status:** Foundation complete. The system supports end-to-end workflows from image upload through template application and layout sequence management.

**Ready for:** Phase 3 (page composition) and Phase 4 (print export).

**Stable interfaces:** API routes, database schema, and CLI commands are production-ready and versioned.

**Next milestone:** Implement page templates and zine assembly to complete the authoring workflow.
