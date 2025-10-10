# Expansion Plan: Transition to First-Class Zine Layout Entities
*Detailed Implementation Guide for New Contributors*

This document provides step-by-step instructions for refactoring the zine-layout codebase to support first-class entities for image sequences, layout templates, laid-out images/pages, and complete zines. Since we're not maintaining backwards compatibility, we can start fresh with a clean database schema.

---

## SECTION 1: Persistence & Data Model Overhaul

### 1.1 Database Schema Design

**Goal:** Define complete SQLite schema for all new entities with proper relationships.

**File to modify:** `pkg/repo/sqlite/migrations.go`

**Current state:** The file contains a single `schemaSQL` constant with tables for `projects`, `assets`, `pages`, and `spreads`.

**Steps:**

#### 1.1.1 Clear existing schema (fresh start)
```go
// In pkg/repo/sqlite/migrations.go
const schemaSQL = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- Drop existing tables if starting fresh
DROP TABLE IF EXISTS spreads;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS projects;
```

#### 1.1.2 Define core entity tables

Add these table definitions to `schemaSQL`:

```sql
-- Projects: Top-level container for all work
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_projects_updated ON projects(updated_at DESC);

-- Assets: Raw uploaded images
CREATE TABLE IF NOT EXISTS assets (
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

CREATE INDEX IF NOT EXISTS idx_assets_project ON assets(project_id, uploaded_at DESC);

-- Image Sequences: Named, ordered collections of assets with optional gaps
CREATE TABLE IF NOT EXISTS image_sequences (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sequences_project ON image_sequences(project_id, name);

-- Image Sequence Items: Ordering within sequences
CREATE TABLE IF NOT EXISTS image_sequence_items (
    sequence_id TEXT NOT NULL,
    asset_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    is_gap BOOLEAN NOT NULL DEFAULT 0,
    PRIMARY KEY (sequence_id, position),
    FOREIGN KEY (sequence_id) REFERENCES image_sequences(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE
);

-- Image Layout Templates: Reusable resize/crop/position settings
-- Stores spread.Settings as JSON
CREATE TABLE IF NOT EXISTS image_layout_templates (
    id TEXT PRIMARY KEY,
    project_id TEXT,  -- NULL = global template
    name TEXT NOT NULL,
    description TEXT,
    settings_json TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_templates_project ON image_layout_templates(project_id, name);

-- Laid Out Images: Asset + Template + Overrides = Resized/Cropped result
CREATE TABLE IF NOT EXISTS laid_out_images (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    asset_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    overrides_json TEXT,  -- Partial spread.Settings to override template
    result_json TEXT,  -- simple.Result from computation
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES image_layout_templates(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_laid_images_project ON laid_out_images(project_id);
CREATE INDEX IF NOT EXISTS idx_laid_images_asset ON laid_out_images(asset_id);

-- Layout Sequences: Ordered sequence of laid-out images (like current spreads but cleaner)
CREATE TABLE IF NOT EXISTS layout_sequences (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_layout_seqs_project ON layout_sequences(project_id);

-- Layout Sequence Items: Ordering within layout sequences
CREATE TABLE IF NOT EXISTS layout_sequence_items (
    sequence_id TEXT NOT NULL,
    laid_out_image_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    is_gap BOOLEAN NOT NULL DEFAULT 0,
    PRIMARY KEY (sequence_id, position),
    FOREIGN KEY (sequence_id) REFERENCES layout_sequences(id) ON DELETE CASCADE,
    FOREIGN KEY (laid_out_image_id) REFERENCES laid_out_images(id) ON DELETE CASCADE
);

-- Page Templates: How to position laid-out images on a printable page
-- Stores margins, gutter info, multi-image positioning from zinelayout DSL
CREATE TABLE IF NOT EXISTS page_templates (
    id TEXT PRIMARY KEY,
    project_id TEXT,  -- NULL = global
    name TEXT NOT NULL,
    description TEXT,
    template_json TEXT NOT NULL,  -- Serialized ZineLayout struct
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_page_templates_project ON page_templates(project_id);

-- Laid Out Pages: Page Template + Laid Out Images = Final printable page
CREATE TABLE IF NOT EXISTS laid_out_pages (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    page_template_id TEXT NOT NULL,
    result_json TEXT,  -- Render metadata, dimensions, file paths
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (page_template_id) REFERENCES page_templates(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_laid_pages_project ON laid_out_pages(project_id);

-- Laid Out Page Inputs: Which laid-out images go into which page
CREATE TABLE IF NOT EXISTS laid_out_page_inputs (
    page_id TEXT NOT NULL,
    laid_out_image_id TEXT NOT NULL,
    input_index INTEGER NOT NULL,  -- Maps to ZineLayout input slots
    PRIMARY KEY (page_id, input_index),
    FOREIGN KEY (page_id) REFERENCES laid_out_pages(id) ON DELETE CASCADE,
    FOREIGN KEY (laid_out_image_id) REFERENCES laid_out_images(id) ON DELETE CASCADE
);

-- Zines: Complete book consisting of an ordered sequence of pages
CREATE TABLE IF NOT EXISTS zines (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_zines_project ON zines(project_id);

-- Zine Pages: Ordering of pages within a zine
CREATE TABLE IF NOT EXISTS zine_pages (
    zine_id TEXT NOT NULL,
    laid_out_page_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    PRIMARY KEY (zine_id, position),
    FOREIGN KEY (zine_id) REFERENCES zines(id) ON DELETE CASCADE,
    FOREIGN KEY (laid_out_page_id) REFERENCES laid_out_pages(id) ON DELETE CASCADE
);

-- Zine Layout Templates: Imposition/folding schemes (8-page zine, 16-page booklet, etc.)
CREATE TABLE IF NOT EXISTS zine_layout_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    template_json TEXT NOT NULL,  -- Serialized imposition rules
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
```

**Why this structure:**
- Foreign keys enforce referential integrity
- CASCADE deletes clean up dependent records
- RESTRICT prevents accidental deletion of referenced templates
- Indexes optimize common queries (list by project, order by date)
- JSON columns store complex settings until we need structured access

---

### 1.2 Repository Layer Types

**Goal:** Define Go structs and interfaces for all entities.

**File to modify:** `pkg/repo/types.go`

#### 1.2.1 Add new struct definitions

After the existing `Project`, `Asset`, `Page`, `Spread` structs, add:

```go
// ImageSequence represents a named, ordered collection of assets
type ImageSequence struct {
    ID          string
    ProjectID   string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// ImageSequenceItem represents an item in a sequence
type ImageSequenceItem struct {
    SequenceID string
    AssetID    string
    Position   int
    IsGap      bool
}

// ImageLayoutTemplate represents reusable layout settings
type ImageLayoutTemplate struct {
    ID           string
    ProjectID    *string  // NULL for global templates
    Name         string
    Description  string
    SettingsJSON string   // Serialized spread.Settings
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// LaidOutImage represents an asset processed through a template
type LaidOutImage struct {
    ID            string
    ProjectID     string
    AssetID       string
    TemplateID    string
    OverridesJSON *string  // Partial spread.Settings overrides
    ResultJSON    *string  // simple.Result from computation
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

// LayoutSequence represents an ordered sequence of laid-out images
type LayoutSequence struct {
    ID          string
    ProjectID   string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// LayoutSequenceItem represents an item in a layout sequence
type LayoutSequenceItem struct {
    SequenceID      string
    LaidOutImageID  string
    Position        int
    IsGap           bool
}

// PageTemplate represents how to compose laid-out images on a page
type PageTemplate struct {
    ID           string
    ProjectID    *string  // NULL for global
    Name         string
    Description  string
    TemplateJSON string   // Serialized ZineLayout
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// LaidOutPage represents a rendered page
type LaidOutPage struct {
    ID             string
    ProjectID      string
    PageTemplateID string
    ResultJSON     *string  // Render metadata
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// LaidOutPageInput maps laid-out images to page input slots
type LaidOutPageInput struct {
    PageID         string
    LaidOutImageID string
    InputIndex     int
}

// Zine represents a complete book
type Zine struct {
    ID          string
    ProjectID   string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// ZinePage represents a page within a zine
type ZinePage struct {
    ZineID         string
    LaidOutPageID  string
    Position       int
}

// ZineLayoutTemplate represents imposition schemes
type ZineLayoutTemplate struct {
    ID           string
    Name         string
    Description  string
    TemplateJSON string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### 1.2.2 Add repository interfaces

After the existing repository interfaces, add:

```go
// ImageSequenceRepository manages image sequences
type ImageSequenceRepository interface {
    Create(seq *ImageSequence) error
    Get(id string) (*ImageSequence, error)
    List(projectID string) ([]*ImageSequence, error)
    Update(seq *ImageSequence) error
    Delete(id string) error
    
    // Item management
    AddItem(item *ImageSequenceItem) error
    GetItems(sequenceID string) ([]*ImageSequenceItem, error)
    ReorderItems(sequenceID string, items []*ImageSequenceItem) error
    DeleteItem(sequenceID string, position int) error
}

// ImageLayoutTemplateRepository manages layout templates
type ImageLayoutTemplateRepository interface {
    Create(tmpl *ImageLayoutTemplate) error
    Get(id string) (*ImageLayoutTemplate, error)
    List(projectID *string) ([]*ImageLayoutTemplate, error)  // NULL = global
    Update(tmpl *ImageLayoutTemplate) error
    Delete(id string) error
}

// LaidOutImageRepository manages laid-out images
type LaidOutImageRepository interface {
    Create(img *LaidOutImage) error
    Get(id string) (*LaidOutImage, error)
    List(projectID string) ([]*LaidOutImage, error)
    ListByAsset(assetID string) ([]*LaidOutImage, error)
    Update(img *LaidOutImage) error
    Delete(id string) error
}

// LayoutSequenceRepository manages layout sequences
type LayoutSequenceRepository interface {
    Create(seq *LayoutSequence) error
    Get(id string) (*LayoutSequence, error)
    List(projectID string) ([]*LayoutSequence, error)
    Update(seq *LayoutSequence) error
    Delete(id string) error
    
    AddItem(item *LayoutSequenceItem) error
    GetItems(sequenceID string) ([]*LayoutSequenceItem, error)
    ReorderItems(sequenceID string, items []*LayoutSequenceItem) error
    DeleteItem(sequenceID string, position int) error
}

// PageTemplateRepository manages page templates
type PageTemplateRepository interface {
    Create(tmpl *PageTemplate) error
    Get(id string) (*PageTemplate, error)
    List(projectID *string) ([]*PageTemplate, error)
    Update(tmpl *PageTemplate) error
    Delete(id string) error
}

// LaidOutPageRepository manages laid-out pages
type LaidOutPageRepository interface {
    Create(page *LaidOutPage) error
    Get(id string) (*LaidOutPage, error)
    List(projectID string) ([]*LaidOutPage, error)
    Update(page *LaidOutPage) error
    Delete(id string) error
    
    SetInputs(pageID string, inputs []*LaidOutPageInput) error
    GetInputs(pageID string) ([]*LaidOutPageInput, error)
}

// ZineRepository manages zines
type ZineRepository interface {
    Create(zine *Zine) error
    Get(id string) (*Zine, error)
    List(projectID string) ([]*Zine, error)
    Update(zine *Zine) error
    Delete(id string) error
    
    SetPages(zineID string, pages []*ZinePage) error
    GetPages(zineID string) ([]*ZinePage, error)
}

// ZineLayoutTemplateRepository manages zine layout templates
type ZineLayoutTemplateRepository interface {
    Create(tmpl *ZineLayoutTemplate) error
    Get(id string) (*ZineLayoutTemplate, error)
    List() ([]*ZineLayoutTemplate, error)
    Update(tmpl *ZineLayoutTemplate) error
    Delete(id string) error
}
```

#### 1.2.3 Update Repositories struct

Replace the existing `Repositories` struct:

```go
// Repositories bundles all repository interfaces
type Repositories struct {
    Projects            ProjectRepository
    Assets              AssetRepository
    ImageSequences      ImageSequenceRepository
    ImageLayoutTemplates ImageLayoutTemplateRepository
    LaidOutImages       LaidOutImageRepository
    LayoutSequences     LayoutSequenceRepository
    PageTemplates       PageTemplateRepository
    LaidOutPages        LaidOutPageRepository
    Zines               ZineRepository
    ZineLayoutTemplates ZineLayoutTemplateRepository
}
```

---

### 1.3 SQLite Repository Implementations

**Goal:** Implement concrete repositories following existing patterns.

**Files to create/modify:** New files in `pkg/repo/sqlite/`

#### 1.3.1 Helper utilities

Update `pkg/repo/sqlite/sqlite.go` to add ID generation:

```go
// generateID creates a unique ID with timestamp and random suffix
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

#### 1.3.2 Create `pkg/repo/sqlite/image_sequences.go`

```go
package sqlite

import (
    "database/sql"
    "fmt"
    "time"
    "github.com/go-go-golems/zine-layout/pkg/repo"
)

type imageSequenceRepo struct {
    db *sql.DB
}

func (r *imageSequenceRepo) Create(seq *repo.ImageSequence) error {
    if seq.ID == "" {
        seq.ID = generateID("seq")
    }
    if seq.CreatedAt.IsZero() {
        seq.CreatedAt = time.Now().UTC()
    }
    if seq.UpdatedAt.IsZero() {
        seq.UpdatedAt = seq.CreatedAt
    }
    
    _, err := r.db.Exec(`INSERT INTO image_sequences 
        (id, project_id, name, description, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)`,
        seq.ID, seq.ProjectID, seq.Name, seq.Description,
        toUnix(seq.CreatedAt), toUnix(seq.UpdatedAt))
    return err
}

func (r *imageSequenceRepo) Get(id string) (*repo.ImageSequence, error) {
    var seq repo.ImageSequence
    var created, updated int64
    var desc sql.NullString
    
    err := r.db.QueryRow(`SELECT id, project_id, name, description, created_at, updated_at
        FROM image_sequences WHERE id = ?`, id).
        Scan(&seq.ID, &seq.ProjectID, &seq.Name, &desc, &created, &updated)
    
    if err != nil {
        return nil, errNotFound(err)
    }
    
    seq.Description = desc.String
    seq.CreatedAt = fromUnix(created)
    seq.UpdatedAt = fromUnix(updated)
    return &seq, nil
}

func (r *imageSequenceRepo) List(projectID string) ([]*repo.ImageSequence, error) {
    rows, err := r.db.Query(`SELECT id, project_id, name, description, created_at, updated_at
        FROM image_sequences WHERE project_id = ? ORDER BY name`, projectID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var seqs []*repo.ImageSequence
    for rows.Next() {
        var seq repo.ImageSequence
        var created, updated int64
        var desc sql.NullString
        
        if err := rows.Scan(&seq.ID, &seq.ProjectID, &seq.Name, &desc, &created, &updated); err != nil {
            return nil, err
        }
        
        seq.Description = desc.String
        seq.CreatedAt = fromUnix(created)
        seq.UpdatedAt = fromUnix(updated)
        seqs = append(seqs, &seq)
    }
    return seqs, rows.Err()
}

func (r *imageSequenceRepo) Update(seq *repo.ImageSequence) error {
    seq.UpdatedAt = time.Now().UTC()
    res, err := r.db.Exec(`UPDATE image_sequences 
        SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
        seq.Name, seq.Description, toUnix(seq.UpdatedAt), seq.ID)
    if err != nil {
        return err
    }
    if rows, err := res.RowsAffected(); err == nil && rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *imageSequenceRepo) Delete(id string) error {
    res, err := r.db.Exec(`DELETE FROM image_sequences WHERE id = ?`, id)
    if err != nil {
        return err
    }
    if rows, err := res.RowsAffected(); err == nil && rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *imageSequenceRepo) AddItem(item *repo.ImageSequenceItem) error {
    _, err := r.db.Exec(`INSERT INTO image_sequence_items 
        (sequence_id, asset_id, position, is_gap) VALUES (?, ?, ?, ?)`,
        item.SequenceID, item.AssetID, item.Position, item.IsGap)
    return err
}

func (r *imageSequenceRepo) GetItems(sequenceID string) ([]*repo.ImageSequenceItem, error) {
    rows, err := r.db.Query(`SELECT sequence_id, asset_id, position, is_gap
        FROM image_sequence_items WHERE sequence_id = ? ORDER BY position`, sequenceID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var items []*repo.ImageSequenceItem
    for rows.Next() {
        var item repo.ImageSequenceItem
        if err := rows.Scan(&item.SequenceID, &item.AssetID, &item.Position, &item.IsGap); err != nil {
            return nil, err
        }
        items = append(items, &item)
    }
    return items, rows.Err()
}

func (r *imageSequenceRepo) ReorderItems(sequenceID string, items []*repo.ImageSequenceItem) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Delete existing items
    if _, err := tx.Exec(`DELETE FROM image_sequence_items WHERE sequence_id = ?`, sequenceID); err != nil {
        return err
    }
    
    // Insert new order
    for _, item := range items {
        if _, err := tx.Exec(`INSERT INTO image_sequence_items 
            (sequence_id, asset_id, position, is_gap) VALUES (?, ?, ?, ?)`,
            sequenceID, item.AssetID, item.Position, item.IsGap); err != nil {
            return err
        }
    }
    
    return tx.Commit()
}

func (r *imageSequenceRepo) DeleteItem(sequenceID string, position int) error {
    _, err := r.db.Exec(`DELETE FROM image_sequence_items 
        WHERE sequence_id = ? AND position = ?`, sequenceID, position)
    return err
}
```

**Pattern to follow:** Implement similar repositories for each entity in separate files:
- `image_layout_templates.go`
- `laid_out_images.go`  
- `layout_sequences.go`
- `page_templates.go`
- `laid_out_pages.go`
- `zines.go`
- `zine_layout_templates.go`

Each follows the same pattern: struct with `db *sql.DB`, CRUD methods, helper methods for relationships.

#### 1.3.3 Update repository factory

In `pkg/repo/sqlite/sqlite.go`, update `NewRepositories`:

```go
func NewRepositories(db *sql.DB) (*repo.Repositories, error) {
    // Execute schema
    if _, err := db.Exec(schemaSQL); err != nil {
        return nil, fmt.Errorf("execute schema: %w", err)
    }
    
    return &repo.Repositories{
        Projects:             &projectRepo{db},
        Assets:               &assetRepo{db},
        ImageSequences:       &imageSequenceRepo{db},
        ImageLayoutTemplates: &imageLayoutTemplateRepo{db},
        LaidOutImages:        &laidOutImageRepo{db},
        LayoutSequences:      &layoutSequenceRepo{db},
        PageTemplates:        &pageTemplateRepo{db},
        LaidOutPages:         &laidOutPageRepo{db},
        Zines:                &zineRepo{db},
        ZineLayoutTemplates:  &zineLayoutTemplateRepo{db},
    }, nil
}
```

---

### 1.4 Remove Filesystem Persistence

**Goal:** Delete old JSON manifest code since we're repository-first now.

#### 1.4.1 Gut `pkg/projects/projects.go`

Keep only the helper functions for file paths and image uploads. Remove:
- `Project` struct (now in `pkg/repo/types.go`)
- `ReadProject`, `WriteProject` functions
- `ListProjects`, `CreateProject`, `DeleteProject` functions  
- Order management functions

Keep:
- `ProjectDir`, `ProjectImagesDir` helpers
- `SavePNGImage` (update to only save file, not write JSON)
- `readImageSize` helper

Updated `SavePNGImage`:

```go
func SavePNGImage(projectsRoot, projectID string, fh *multipart.FileHeader) (string, int, int, error) {
    // Returns: assetID, width, height, error
    if fh.Size == 0 {
        return "", 0, 0, fmt.Errorf("empty file")
    }
    
    dir := ProjectImagesDir(projectsRoot, projectID)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return "", 0, 0, err
    }
    
    assetID := generateAssetID()
    dstPath := filepath.Join(dir, assetID+".png")
    
    src, err := fh.Open()
    if err != nil {
        return "", 0, 0, err
    }
    defer src.Close()
    
    dst, err := os.Create(dstPath)
    if err != nil {
        return "", 0, 0, err
    }
    defer dst.Close()
    
    if _, err := io.Copy(dst, src); err != nil {
        os.Remove(dstPath)
        return "", 0, 0, err
    }
    
    w, h, err := readImageSize(dstPath)
    if err != nil {
        os.Remove(dstPath)
        return "", 0, 0, err
    }
    
    return assetID, w, h, nil
}

func generateAssetID() string {
    ts := time.Now().UTC().Format("20060102T150405Z")
    const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, 6)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return fmt.Sprintf("img-%s-%s", ts, string(b))
}
```

---

## SECTION 2: API Surface & Service Layer

### 2.1 Service Layer for Business Logic

**Goal:** Create service package that orchestrates repository operations and implements workflows.

**New file:** `pkg/services/layout.go`

```go
package services

import (
    "encoding/json"
    "fmt"
    "github.com/go-go-golems/zine-layout/pkg/repo"
    "github.com/go-go-golems/zine-layout/pkg/spread"
    "github.com/go-go-golems/zine-layout/pkg/spread/simple"
)

type LayoutService struct {
    repos *repo.Repositories
}

func NewLayoutService(repos *repo.Repositories) *LayoutService {
    return &LayoutService{repos: repos}
}

// CreateLaidOutImage takes an asset + template and computes the layout
func (s *LayoutService) CreateLaidOutImage(
    projectID, assetID, templateID string,
    overrides *spread.Settings,
) (*repo.LaidOutImage, error) {
    // Fetch asset for dimensions
    asset, err := s.repos.Assets.Get(projectID, assetID)
    if err != nil {
        return nil, fmt.Errorf("asset not found: %w", err)
    }
    
    // Fetch template
    tmpl, err := s.repos.ImageLayoutTemplates.Get(templateID)
    if err != nil {
        return nil, fmt.Errorf("template not found: %w", err)
    }
    
    // Parse template settings
    var settings spread.Settings
    if err := json.Unmarshal([]byte(tmpl.SettingsJSON), &settings); err != nil {
        return nil, fmt.Errorf("invalid template settings: %w", err)
    }
    
    // Apply overrides if provided
    if overrides != nil {
        settings = mergeSettings(settings, *overrides)
    }
    
    // Run computation
    meta := spread.ImageMeta{Width: asset.Width, Height: asset.Height}
    inputs, err := simple.InputsFromSettings(settings, meta)
    if err != nil {
        return nil, fmt.Errorf("invalid settings: %w", err)
    }
    
    trace := &simple.Trace{}
    result := simple.ComputePlacement(inputs, trace)
    
    // Serialize result
    resultJSON, err := json.Marshal(result)
    if err != nil {
        return nil, err
    }
    resultStr := string(resultJSON)
    
    // Serialize overrides
    var overridesJSON *string
    if overrides != nil {
        data, err := json.Marshal(overrides)
        if err != nil {
            return nil, err
        }
        str := string(data)
        overridesJSON = &str
    }
    
    // Create laid-out image record
    laidOut := &repo.LaidOutImage{
        ProjectID:     projectID,
        AssetID:       assetID,
        TemplateID:    templateID,
        OverridesJSON: overridesJSON,
        ResultJSON:    &resultStr,
    }
    
    if err := s.repos.LaidOutImages.Create(laidOut); err != nil {
        return nil, err
    }
    
    return laidOut, nil
}

func mergeSettings(base, override spread.Settings) spread.Settings {
    // Apply non-zero overrides to base
    result := base
    if override.UserScale != 0 {
        result.UserScale = override.UserScale
    }
    if override.PositionX != 0 || override.PositionY != 0 {
        result.PositionX = override.PositionX
        result.PositionY = override.PositionY
    }
    // ... etc for all fields
    return result
}
```

**Similar service files to create:**
- `pkg/services/sequences.go` – manage image/layout sequences
- `pkg/services/pages.go` – create laid-out pages from templates
- `pkg/services/zines.go` – assemble zines and apply layout templates

---

### 2.2 REST API Endpoints

**Goal:** Add clean CRUD endpoints for all entities in `pkg/serve/server.go`.

#### 2.2.1 Image Sequences API

Add to `Routes()` method in `pkg/serve/server.go`:

```go
// Image Sequences
mux.HandleFunc("/api/projects/{id}/image-sequences", func(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/")
    projectID := parts[0]
    
    switch r.Method {
    case http.MethodGet:
        s.handleListImageSequences(w, projectID)
    case http.MethodPost:
        s.handleCreateImageSequence(w, r, projectID)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
})

mux.HandleFunc("/api/projects/{projectId}/image-sequences/{id}", func(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/")
    projectID := parts[0]
    sequenceID := parts[2]
    
    switch r.Method {
    case http.MethodGet:
        s.handleGetImageSequence(w, sequenceID)
    case http.MethodPut:
        s.handleUpdateImageSequence(w, r, projectID, sequenceID)
    case http.MethodDelete:
        s.handleDeleteImageSequence(w, sequenceID)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
})

// Sequence items
mux.HandleFunc("/api/image-sequences/{id}/items", func(w http.ResponseWriter, r *http.Request) {
    sequenceID := strings.TrimPrefix(r.URL.Path, "/api/image-sequences/")
    sequenceID = strings.TrimSuffix(sequenceID, "/items")
    
    switch r.Method {
    case http.MethodGet:
        s.handleGetSequenceItems(w, sequenceID)
    case http.MethodPut:
        s.handleReorderSequenceItems(w, r, sequenceID)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
})
```

#### 2.2.2 Handler implementations

Add handler methods to `Server` struct:

```go
func (s *Server) handleListImageSequences(w http.ResponseWriter, projectID string) {
    sequences, err := s.repos.ImageSequences.List(projectID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"sequences": sequences})
}

func (s *Server) handleCreateImageSequence(w http.ResponseWriter, r *http.Request, projectID string) {
    var req struct {
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    
    seq := &repo.ImageSequence{
        ProjectID:   projectID,
        Name:        req.Name,
        Description: req.Description,
    }
    
    if err := s.repos.ImageSequences.Create(seq); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    writeJSON(w, http.StatusCreated, map[string]any{"sequence": seq})
}

func (s *Server) handleGetImageSequence(w http.ResponseWriter, sequenceID string) {
    seq, err := s.repos.ImageSequences.Get(sequenceID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Also fetch items
    items, err := s.repos.ImageSequences.GetItems(sequenceID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]any{
        "sequence": seq,
        "items":    items,
    })
}

func (s *Server) handleUpdateImageSequence(w http.ResponseWriter, r *http.Request, projectID, sequenceID string) {
    var req struct {
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    
    seq, err := s.repos.ImageSequences.Get(sequenceID)
    if err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    
    seq.Name = req.Name
    seq.Description = req.Description
    
    if err := s.repos.ImageSequences.Update(seq); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]any{"sequence": seq})
}

func (s *Server) handleDeleteImageSequence(w http.ResponseWriter, sequenceID string) {
    if err := s.repos.ImageSequences.Delete(sequenceID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleGetSequenceItems(w http.ResponseWriter, sequenceID string) {
    items, err := s.repos.ImageSequences.GetItems(sequenceID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleReorderSequenceItems(w http.ResponseWriter, r *http.Request, sequenceID string) {
    var req struct {
        Items []*repo.ImageSequenceItem `json:"items"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    
    if err := s.repos.ImageSequences.ReorderItems(sequenceID, req.Items); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
```

**Pattern to follow:** Implement similar endpoint groups for:
- `/api/projects/{id}/image-layout-templates`
- `/api/projects/{id}/laid-out-images`
- `/api/projects/{id}/layout-sequences`
- `/api/projects/{id}/page-templates`
- `/api/projects/{id}/laid-out-pages`
- `/api/projects/{id}/zines`
- `/api/zine-layout-templates` (global, no project scope)

Each follows the same REST pattern: list, create, get, update, delete.

---

## SECTION 3: Tooling, CLI, and UI Integration

### 3.1 CLI Commands

**Goal:** Add Glazed commands for all new entities.

#### 3.1.1 Image Sequences Commands

**New file:** `cmd/zine-layout/cmds/api/image-sequences-list.go`

```go
package api

import (
    "context"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/go-go-golems/glazed/pkg/types"
)

type ImageSequencesListCommand struct {
    *cmds.CommandDescription
}

type ImageSequencesListSettings struct {
    ProjectID string `glazed.parameter:"project-id"`
}

func NewImageSequencesListCommand() (*ImageSequencesListCommand, error) {
    return &ImageSequencesListCommand{
        CommandDescription: cmds.NewCommandDescription(
            "image-sequences-list",
            cmds.WithShort("List image sequences in a project"),
            cmds.WithFlags(
                parameters.NewParameterDefinition(
                    "project-id",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Project ID"),
                    parameters.WithRequired(true),
                ),
            ),
            cmds.WithLayersList(defaultGlazedLayer),
        ),
    }, nil
}

func (c *ImageSequencesListCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &ImageSequencesListSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    url := fmt.Sprintf("%s/projects/%s/image-sequences", baseURL(), s.ProjectID)
    
    var resp struct {
        Sequences []map[string]interface{} `json:"sequences"`
    }
    
    if err := httpGetJSON(url, &resp); err != nil {
        return err
    }
    
    for _, seq := range resp.Sequences {
        row := types.NewRow(
            types.MRP("id", seq["id"]),
            types.MRP("name", seq["name"]),
            types.MRP("description", seq["description"]),
            types.MRP("created_at", seq["created_at"]),
            types.MRP("updated_at", seq["updated_at"]),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```

**Similarly, create:**
- `image-sequences-get.go`
- `image-sequences-create.go`
- `image-sequences-update.go`
- `image-sequences-delete.go`
- `image-sequences-add-item.go`
- `image-sequences-reorder.go`

And similar command files for:
- `image-layout-templates-*`
- `laid-out-images-*`
- `layout-sequences-*`
- `page-templates-*`
- `laid-out-pages-*`
- `zines-*`
- `zine-layout-templates-*`

#### 3.1.2 Register commands

Update `cmd/zine-layout/cmds/api/commands.go`:

```go
func AddAllAPICommands(cmd *cobra.Command) error {
    // ... existing commands ...
    
    // Image Sequences
    if err := addCommand(cmd, NewImageSequencesListCommand); err != nil {
        return err
    }
    if err := addCommand(cmd, NewImageSequencesGetCommand); err != nil {
        return err
    }
    if err := addCommand(cmd, NewImageSequencesCreateCommand); err != nil {
        return err
    }
    // ... etc for all new commands
    
    return nil
}
```

### 3.2 Frontend Integration

**Goal:** Update RTK Query API and add Redux state management for new entities.

#### 3.2.1 Extend RTK Query API

**File:** `web/src/api.ts`

Add type definitions:

```typescript
export interface ImageSequence {
  id: string;
  project_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface ImageSequenceItem {
  sequence_id: string;
  asset_id: string;
  position: number;
  is_gap: boolean;
}

export interface ImageLayoutTemplate {
  id: string;
  project_id?: string;
  name: string;
  description?: string;
  settings_json: string;
  created_at: string;
  updated_at: string;
}

export interface LaidOutImage {
  id: string;
  project_id: string;
  asset_id: string;
  template_id: string;
  overrides_json?: string;
  result_json?: string;
  created_at: string;
  updated_at: string;
}

// ... similar for other entities
```

Add endpoints:

```typescript
export const api = createApi({
  // ... existing config ...
  tagTypes: ['Project', 'Asset', 'ImageSequence', 'ImageLayoutTemplate', 
             'LaidOutImage', 'LayoutSequence', 'PageTemplate', 
             'LaidOutPage', 'Zine'],
  endpoints: (b) => ({
    // ... existing endpoints ...
    
    // Image Sequences
    getImageSequences: b.query<{ sequences: ImageSequence[] }, { projectId: string }>({
      query: ({ projectId }) => `/projects/${projectId}/image-sequences`,
      providesTags: ['ImageSequence'],
    }),
    
    createImageSequence: b.mutation<{ sequence: ImageSequence }, { 
      projectId: string; 
      name: string; 
      description?: string;
    }>({
      query: ({ projectId, ...body }) => ({
        url: `/projects/${projectId}/image-sequences`,
        method: 'POST',
        body,
      }),
      invalidatesTags: ['ImageSequence'],
    }),
    
    updateImageSequence: b.mutation<{ sequence: ImageSequence }, {
      projectId: string;
      sequenceId: string;
      name: string;
      description?: string;
    }>({
      query: ({ projectId, sequenceId, ...body }) => ({
        url: `/projects/${projectId}/image-sequences/${sequenceId}`,
        method: 'PUT',
        body,
      }),
      invalidatesTags: ['ImageSequence'],
    }),
    
    deleteImageSequence: b.mutation<{ ok: boolean }, {
      projectId: string;
      sequenceId: string;
    }>({
      query: ({ projectId, sequenceId }) => ({
        url: `/projects/${projectId}/image-sequences/${sequenceId}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['ImageSequence'],
    }),
    
    getSequenceItems: b.query<{ items: ImageSequenceItem[] }, { sequenceId: string }>({
      query: ({ sequenceId }) => `/image-sequences/${sequenceId}/items`,
      providesTags: ['ImageSequence'],
    }),
    
    reorderSequenceItems: b.mutation<{ ok: boolean }, {
      sequenceId: string;
      items: ImageSequenceItem[];
    }>({
      query: ({ sequenceId, items }) => ({
        url: `/image-sequences/${sequenceId}/items`,
        method: 'PUT',
        body: { items },
      }),
      invalidatesTags: ['ImageSequence'],
    }),
    
    // ... similar endpoints for other entities
  }),
});

export const {
  // ... existing hooks ...
  useGetImageSequencesQuery,
  useCreateImageSequenceMutation,
  useUpdateImageSequenceMutation,
  useDeleteImageSequenceMutation,
  useGetSequenceItemsQuery,
  useReorderSequenceItemsMutation,
  // ... hooks for other entities
} = api;
```

#### 3.2.2 Create UI Components

**New file:** `web/src/views/ImageSequenceEditor.tsx`

```typescript
import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import {
  useGetImageSequencesQuery,
  useCreateImageSequenceMutation,
  useGetSequenceItemsQuery,
  useReorderSequenceItemsMutation,
  useGetImagesQuery,
} from '../api';
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd';

export const ImageSequenceEditor: React.FC = () => {
  const { projectId } = useParams();
  const [selectedSequenceId, setSelectedSequenceId] = useState<string | null>(null);
  
  const { data: sequences } = useGetImageSequencesQuery({ projectId: projectId! });
  const { data: assets } = useGetImagesQuery({ id: projectId! });
  const { data: sequenceItems } = useGetSequenceItemsQuery(
    { sequenceId: selectedSequenceId! },
    { skip: !selectedSequenceId }
  );
  
  const [createSequence] = useCreateImageSequenceMutation();
  const [reorderItems] = useReorderSequenceItemsMutation();
  
  const handleCreateSequence = async () => {
    const name = prompt('Sequence name:');
    if (name) {
      await createSequence({ projectId: projectId!, name });
    }
  };
  
  const handleDragEnd = async (result: any) => {
    if (!result.destination || !sequenceItems) return;
    
    const items = Array.from(sequenceItems.items);
    const [removed] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, removed);
    
    // Update positions
    const reordered = items.map((item, index) => ({
      ...item,
      position: index,
    }));
    
    await reorderItems({ sequenceId: selectedSequenceId!, items: reordered });
  };
  
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Image Sequences</h1>
      
      <div className="grid grid-cols-3 gap-4">
        {/* Sequence List */}
        <div className="bg-white p-4 rounded shadow">
          <div className="flex justify-between items-center mb-4">
            <h2 className="font-semibold">Sequences</h2>
            <button
              onClick={handleCreateSequence}
              className="px-3 py-1 bg-blue-500 text-white rounded text-sm"
            >
              + New
            </button>
          </div>
          
          <div className="space-y-2">
            {sequences?.sequences.map((seq) => (
              <button
                key={seq.id}
                onClick={() => setSelectedSequenceId(seq.id)}
                className={`w-full text-left p-2 rounded ${
                  selectedSequenceId === seq.id ? 'bg-blue-100' : 'hover:bg-gray-100'
                }`}
              >
                {seq.name}
              </button>
            ))}
          </div>
        </div>
        
        {/* Sequence Items */}
        <div className="bg-white p-4 rounded shadow">
          <h2 className="font-semibold mb-4">Items</h2>
          
          {selectedSequenceId && sequenceItems && (
            <DragDropContext onDragEnd={handleDragEnd}>
              <Droppable droppableId="sequence-items">
                {(provided) => (
                  <div
                    {...provided.droppableProps}
                    ref={provided.innerRef}
                    className="space-y-2"
                  >
                    {sequenceItems.items.map((item, index) => {
                      const asset = assets?.images.find((a) => a.id === item.asset_id);
                      return (
                        <Draggable
                          key={`${item.asset_id}-${index}`}
                          draggableId={`${item.asset_id}-${index}`}
                          index={index}
                        >
                          {(provided) => (
                            <div
                              ref={provided.innerRef}
                              {...provided.draggableProps}
                              {...provided.dragHandleProps}
                              className="p-2 bg-gray-50 rounded flex items-center gap-2"
                            >
                              <span className="text-gray-500">{index + 1}.</span>
                              <span>{asset?.name || item.asset_id}</span>
                            </div>
                          )}
                        </Draggable>
                      );
                    })}
                    {provided.placeholder}
                  </div>
                )}
              </Droppable>
            </DragDropContext>
          )}
        </div>
        
        {/* Available Assets */}
        <div className="bg-white p-4 rounded shadow">
          <h2 className="font-semibold mb-4">Available Assets</h2>
          <div className="space-y-2">
            {assets?.images.map((asset) => (
              <div key={asset.id} className="p-2 bg-gray-50 rounded text-sm">
                {asset.name}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
```

**Similar components to create:**
- `LayoutTemplateManager.tsx` – browse/create/edit layout templates
- `LaidOutImageViewer.tsx` – preview laid-out images
- `PageComposer.tsx` – assemble pages from laid-out images
- `ZineBuilder.tsx` – compose zines from pages

---

### 3.3 Testing Strategy

**Goal:** Ensure all pieces work together end-to-end.

#### 3.3.1 Integration test structure

**New file:** `pkg/serve/integration_test.go`

```go
package serve_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/go-go-golems/zine-layout/pkg/serve"
    "github.com/go-go-golems/zine-layout/pkg/repo"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*serve.Server, func()) {
    // Create in-memory SQLite
    db, err := sql.Open("sqlite", ":memory:")
    require.NoError(t, err)
    
    repos, err := sqliterepo.NewRepositories(db)
    require.NoError(t, err)
    
    server := serve.New(serve.Settings{
        Root: "./dist",
        DataRoot: t.TempDir(),
        Addr: ":0",
    })
    server.SetRepos(repos)
    
    cleanup := func() {
        db.Close()
    }
    
    return server, cleanup
}

func TestImageSequenceWorkflow(t *testing.T) {
    server, cleanup := setupTestServer(t)
    defer cleanup()
    
    handler := server.Routes()
    
    // 1. Create project
    reqBody := bytes.NewBufferString(`{"name":"Test Project"}`)
    req := httptest.NewRequest("POST", "/api/projects", reqBody)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    var createResp struct {
        Project repo.Project `json:"project"`
    }
    json.NewDecoder(w.Body).Decode(&createResp)
    projectID := createResp.Project.ID
    
    // 2. Create image sequence
    reqBody = bytes.NewBufferString(`{"name":"Summer Photos","description":"Best of 2025"}`)
    req = httptest.NewRequest("POST", "/api/projects/"+projectID+"/image-sequences", reqBody)
    w = httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    
    // 3. List sequences
    req = httptest.NewRequest("GET", "/api/projects/"+projectID+"/image-sequences", nil)
    w = httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    var listResp struct {
        Sequences []repo.ImageSequence `json:"sequences"`
    }
    json.NewDecoder(w.Body).Decode(&listResp)
    assert.Len(t, listResp.Sequences, 1)
    assert.Equal(t, "Summer Photos", listResp.Sequences[0].Name)
}

// Similar tests for other workflows:
// - TestLayoutTemplateCreation
// - TestLaidOutImageComputation
// - TestPageComposition
// - TestZineAssembly
```

#### 3.3.2 CLI smoke tests

**New file:** `cmd/zine-layout/cmds/api/integration_test.sh`

```bash
#!/bin/bash
set -e

# Start server in background
./zine-layout serve --addr :8089 --data-root ./test-data &
SERVER_PID=$!
trap "kill $SERVER_PID" EXIT

sleep 2

# Test workflow
echo "Creating project..."
PROJECT_ID=$(./zine-layout api projects-create --name "CLI Test" --output json | jq -r '.id')

echo "Creating image sequence..."
SEQ_ID=$(./zine-layout api image-sequences-create --project-id "$PROJECT_ID" --name "Test Seq" --output json | jq -r '.id')

echo "Listing sequences..."
./zine-layout api image-sequences-list --project-id "$PROJECT_ID"

echo "✓ All tests passed"
```

---

## Implementation Checklist

### Phase 1: Projects + Images + Image Sequences

**Goal:** Build foundation for project and asset management with sequencing support.
> Status: Backend + CLI layers complete; frontend and automated validation still outstanding.

**Backend:**
- [x] 1.1 Update database schema in `pkg/repo/sqlite/migrations.go`
  - Add `projects`, `assets`, `image_sequences`, `image_sequence_items` tables
  - Remove old `pages` and `spreads` tables for now
- [x] 1.2 Add entity types to `pkg/repo/types.go`
  - `Project` (simplified: id, name, description, timestamps)
  - `Asset` (with metadata_json for EXIF/date/ratio)
  - `ImageSequence`, `ImageSequenceItem`
- [x] 1.3 Implement SQLite repositories
  - `pkg/repo/sqlite/projects.go` – Create/Get/List/Update/Delete
  - `pkg/repo/sqlite/assets.go` – Create/Get/ListByProject/Delete, capture metadata on upload
  - `pkg/repo/sqlite/image_sequences.go` – CRUD + item management (Add/Get/Reorder/Delete items)
- [x] 1.4 Update `pkg/repo/sqlite/sqlite.go`
  - Wire up new repositories in `NewRepositories()`
  - Add `generateID()` helper for ID generation
- [x] 1.5 Gut `pkg/projects/projects.go`
  - Remove JSON manifest read/write functions
  - Keep only `ProjectDir`, `ProjectImagesDir`, updated `SavePNGImage` (returns assetID, w, h)
- [x] 1.6 Update `pkg/serve/server.go`
  - Refactor `/api/projects` to use repos only (no filesystem manifests)
  - Refactor `/api/projects/{id}/images` upload handler to create Asset records with metadata
  - Add `/api/projects/{id}/image-sequences` endpoints (list, create)
  - Add `/api/projects/{id}/image-sequences/{id}` endpoints (get, update, delete)
  - Add `/api/image-sequences/{id}/items` endpoints (get, reorder)

**CLI:**
- [x] 1.7 Add Glazed commands in `cmd/zine-layout/cmds/api/image_sequences/` (verb group with `list|get|create|update|delete|add-item|reorder|delete-item`)
- [x] 1.8 Register commands in `cmd/zine-layout/cmds/api/commands.go`

**Frontend:**
- [ ] 1.9 Extend `web/src/api.ts`
  - Add `ImageSequence`, `ImageSequenceItem` types
  - Add RTK Query endpoints: `getImageSequences`, `createImageSequence`, `getSequenceItems`, `reorderSequenceItems`, etc.
- [ ] 1.10 Create `web/src/views/ImageSequenceEditor.tsx`
  - List sequences in project
  - Create/rename/delete sequences
  - Drag-and-drop interface for reordering assets within sequence
  - Support gap markers
- [ ] 1.11 Add route in `web/src/routes/*.tsx` for sequence editor

**Testing:**
- [ ] 1.14 Create CLI smoke test script `cmd/zine-layout/cmds/api/phase1_test.sh`

**Validation:**
- [ ] 1.15 Test end-to-end: create project → upload images → create sequence → reorder → view in UI

---

### Phase 2: Image Layout Templates + Laid Out Images + Layout Sequences

**Goal:** Implement template system and image layout computation using algorithms from `01-algorithm-for-resizing.md`.

**Backend:**
- [ ] 2.1 Add tables to schema in `pkg/repo/sqlite/migrations.go`
  - `image_layout_templates`
  - `laid_out_images`
  - `layout_sequences`
  - `layout_sequence_items`
- [ ] 2.2 Add entity types to `pkg/repo/types.go`
  - `ImageLayoutTemplate`
  - `LaidOutImage`
  - `LayoutSequence`, `LayoutSequenceItem`
- [ ] 2.3 Implement repositories
  - `pkg/repo/sqlite/image_layout_templates.go` – CRUD, support global (project_id NULL) and project-specific templates
  - `pkg/repo/sqlite/laid_out_images.go` – CRUD + ListByAsset
  - `pkg/repo/sqlite/layout_sequences.go` – CRUD + item management
- [ ] 2.4 Create service layer in `pkg/services/layout.go`
  - `CreateLaidOutImage(projectID, assetID, templateID, overrides)` – fetches asset/template, runs `simple.ComputePlacement`, stores result
  - `ApplyTemplateToSequence(projectID, sequenceID, templateID)` – batch-creates laid-out images for all assets in sequence
  - `mergeSettings(base, override)` – helper to apply overrides to template settings
- [ ] 2.5 Add REST endpoints in `pkg/serve/server.go`
  - `/api/projects/{id}/image-layout-templates` (list, create)
  - `/api/image-layout-templates` (list global templates)
  - `/api/image-layout-templates/{id}` (get, update, delete)
  - `/api/projects/{id}/laid-out-images` (list, create)
  - `/api/laid-out-images/{id}` (get, update, delete)
  - `/api/projects/{id}/layout-sequences` (list, create)
  - `/api/layout-sequences/{id}` (get, update, delete)
  - `/api/layout-sequences/{id}/items` (get, reorder)
- [ ] 2.6 Add preview/render endpoints
  - `/api/laid-out-images/{id}/preview` – render preview using stored result
  - `/api/laid-out-images/{id}/export` – export final cropped/resized image

**Algorithms Implementation:**
- [ ] 2.7 Adapt `pkg/spread/simple/algorithm.go` to support all modes from `01-algorithm-for-resizing.md`
  - Current: covers Page+margins and basic spread
  - Add: Fixed crop format mode, Fit to width/height mode, Spread with gutter position/overlap
  - Update `Inputs` struct with new fields (gutter position, overlap, mode selector)
  - Update `Result` struct to capture all output variants
- [ ] 2.8 Extend `spread.Settings` type in `pkg/spread/types.go`
  - Add fields for mode selection, gutter position, overlap
  - Maintain backwards compatibility with existing Simple YAML
- [ ] 2.9 Update rendering in `pkg/spread/simple/render.go`
  - Support new output modes (crop-only, fit, spread with overlap)
  - Generate all required export variants per mode

**CLI:**
- [ ] 2.10 Add Glazed commands
  - `image-layout-templates/*` (list, get, create, update, delete)
  - `laid-out-images/*` (list, get, create, delete, preview, export)
  - `layout-sequences/*` (list, create, get, update, delete, add-item, reorder)
- [ ] 2.11 Add batch operation commands
  - `apply-template-to-sequence` – create laid-out images for entire sequence
  - `preview-template` – preview template settings on sample image

**Frontend:**
- [ ] 2.12 Extend `web/src/api.ts`
  - Add types: `ImageLayoutTemplate`, `LaidOutImage`, `LayoutSequence`
  - Add all CRUD endpoints and preview/export endpoints
- [ ] 2.13 Create `web/src/views/LayoutTemplateManager.tsx`
  - List templates (global + project-specific)
  - Create/edit templates using settings from `bookSpreadSlice`
  - Preview template on selected asset
  - Save current Book Spread Designer settings as new template
- [ ] 2.14 Create `web/src/views/LaidOutImageViewer.tsx`
  - Grid view of laid-out images
  - Preview panel showing result
  - Edit overrides (zoom, position) without changing template
  - Re-compute button
- [ ] 2.15 Create `web/src/views/LayoutSequenceEditor.tsx`
  - Similar to ImageSequenceEditor but for laid-out images
  - Preview panel showing sequence in order
  - Drag-and-drop reordering

**Testing:**
- [ ] 2.16 Write service layer tests in `pkg/services/layout_test.go`
- [ ] 2.17 Write algorithm tests for new modes in `pkg/spread/simple/algorithm_test.go`
- [ ] 2.18 Integration test: create template → apply to asset → verify result dimensions
- [ ] 2.19 CLI smoke test covering template creation and application

**Validation:**
- [ ] 2.20 Test workflow: create template → apply to image → preview → create layout sequence → export

---

### Phase 3: Page Templates + Laid Out Pages + Zines

**Goal:** Implement multi-image page composition and zine assembly.

**Backend:**
- [ ] 3.1 Add tables to schema
  - `page_templates`
  - `laid_out_pages`
  - `laid_out_page_inputs`
  - `zines`
  - `zine_pages`
- [ ] 3.2 Add entity types to `pkg/repo/types.go`
  - `PageTemplate`
  - `LaidOutPage`, `LaidOutPageInput`
  - `Zine`, `ZinePage`
- [ ] 3.3 Implement repositories
  - `pkg/repo/sqlite/page_templates.go`
  - `pkg/repo/sqlite/laid_out_pages.go` – include SetInputs/GetInputs for page-to-image mappings
  - `pkg/repo/sqlite/zines.go` – include SetPages/GetPages for zine-to-page ordering
- [ ] 3.4 Create service layer in `pkg/services/pages.go`
  - `CreateLaidOutPage(projectID, pageTemplateID, laidOutImageIDs)` – validates inputs, renders page using `pkg/zinelayout`
  - `RenderPage(pageID)` – fetches page + inputs + template, calls `ZineLayout.CreateOutputImage()`
- [ ] 3.5 Create service layer in `pkg/services/zines.go`
  - `CreateZine(projectID, name, pageIDs)` – creates zine record and page ordering
  - `AddPageToZine(zineID, pageID, position)` – insert page at position
  - `ReorderZinePages(zineID, pageIDs)` – update page order
- [ ] 3.6 Add REST endpoints in `pkg/serve/server.go`
  - `/api/projects/{id}/page-templates` (list, create)
  - `/api/page-templates/{id}` (get, update, delete)
  - `/api/projects/{id}/laid-out-pages` (list, create)
  - `/api/laid-out-pages/{id}` (get, update, delete)
  - `/api/laid-out-pages/{id}/inputs` (get, set)
  - `/api/laid-out-pages/{id}/preview` – render and return image
  - `/api/projects/{id}/zines` (list, create)
  - `/api/zines/{id}` (get, update, delete)
  - `/api/zines/{id}/pages` (get, set order)

**CLI:**
- [ ] 3.7 Add Glazed commands
  - `page-templates/*` (list, get, create, update, delete)
  - `laid-out-pages/*` (list, get, create, delete, preview)
  - `zines/*` (list, get, create, update, delete, add-page, reorder-pages)

**Frontend:**
- [ ] 3.8 Extend `web/src/api.ts`
  - Add types: `PageTemplate`, `LaidOutPage`, `Zine`
  - Add all CRUD endpoints
- [ ] 3.9 Create `web/src/views/PageComposer.tsx`
  - Select page template (grid layout)
  - Drag laid-out images into template slots
  - Preview final page composition
  - Save as laid-out page
- [ ] 3.10 Create `web/src/views/ZineBuilder.tsx`
  - Create new zine
  - Add laid-out pages to zine
  - Drag-and-drop page ordering
  - Preview zine as page sequence
  - Export individual pages or entire zine

**Testing:**
- [ ] 3.11 Write service tests for page composition
- [ ] 3.12 Write integration test: create page template → add images → render page → assemble zine
- [ ] 3.13 CLI test for complete workflow

**Validation:**
- [ ] 3.14 Test workflow: create multi-image page → add to zine → preview zine

---

### Phase 4: Zine Export & Imposition Templates

**Goal:** Implement final imposition schemes for print-ready output (folding, booklet layouts, etc.).

**Backend:**
- [ ] 4.1 Add table to schema
  - `zine_layout_templates` (imposition rules)
- [ ] 4.2 Add entity type to `pkg/repo/types.go`
  - `ZineLayoutTemplate`
- [ ] 4.3 Implement repository
  - `pkg/repo/sqlite/zine_layout_templates.go` – global templates only
- [ ] 4.4 Create service layer in `pkg/services/imposition.go`
  - `ApplyZineLayout(zineID, templateID)` – takes zine pages and reorders/rotates per template
  - `RenderZineForPrint(zineID, templateID, outputDir)` – generates final print-ready PDFs/images
  - Support common layouts: 8-page folded zine, 16-page booklet, spreads
- [ ] 4.5 Seed default templates
  - Port examples from `data/presets/10_8_sheet_zine.yaml` and `11_16_sheet_zine.yaml` into template records
  - Create `pkg/services/seed_templates.go` for initialization
- [ ] 4.6 Add REST endpoints in `pkg/serve/server.go`
  - `/api/zine-layout-templates` (list, create)
  - `/api/zine-layout-templates/{id}` (get, update, delete)
  - `/api/zines/{id}/apply-layout` (POST with template_id) – generates final output
  - `/api/zines/{id}/export` (GET with template_id, format) – downloads ZIP of print files

**CLI:**
- [ ] 4.7 Add Glazed commands
  - `zine-layout-templates/*` (list, get, create, delete)
  - `zine-apply-layout` – apply imposition template to zine
  - `zine-export` – export zine to print-ready files
- [ ] 4.8 Update `render` command to work with zines
  - `zine-layout render --zine-id <id> --template <template-id> --output-dir <dir>`

**Frontend:**
- [ ] 4.9 Extend `web/src/api.ts`
  - Add `ZineLayoutTemplate` type
  - Add endpoints for templates and export
- [ ] 4.10 Create `web/src/views/ZineExporter.tsx`
  - Select zine
  - Choose imposition template
  - Preview final layout (page ordering)
  - Export options (PDF, PNG, with/without crop marks)
  - Download button
- [ ] 4.11 Add export panel to `ZineBuilder.tsx`
  - Quick export button with default template
  - Link to full exporter view

**Testing:**
- [ ] 4.12 Write imposition service tests
- [ ] 4.13 Integration test: create 8-page zine → apply folding template → verify page order
- [ ] 4.14 Visual regression tests for known templates

**Validation:**
- [ ] 4.15 Test complete workflow: project → images → sequence → templates → laid-out images → pages → zine → export for print
- [ ] 4.16 Print test: export 8-page zine, fold, verify page order matches expected

**Documentation & Polish:**
- [ ] 4.17 Add help topics for all new CLI commands
- [ ] 4.18 Update README with complete workflow examples
- [ ] 4.19 Create video/GIF walkthrough of UI workflow
- [ ] 4.20 Performance optimization (lazy loading, caching previews)
- [ ] 4.21 Deploy and dogfood on real photobook project

---

## Key Files Reference

### Backend
- `pkg/repo/types.go` – all entity struct definitions
- `pkg/repo/sqlite/migrations.go` – database schema
- `pkg/repo/sqlite/*.go` – repository implementations
- `pkg/services/*.go` – business logic orchestration
- `pkg/serve/server.go` – REST API handlers
- `pkg/projects/projects.go` – file storage helpers (gutted)

### CLI
- `cmd/zine-layout/main.go` – CLI entry point
- `cmd/zine-layout/cmds/api/*.go` – Glazed command implementations
- `cmd/zine-layout/cmds/serve.go` – server command
- `cmd/zine-layout/cmds/render.go` – render command

### Frontend
- `web/src/api.ts` – RTK Query definitions
- `web/src/views/*.tsx` – page components
- `web/src/components/*.tsx` – reusable UI components
- `web/src/store/*.ts` – Redux state management

### Tests
- `pkg/serve/integration_test.go` – HTTP API tests
- `pkg/repo/sqlite/*_test.go` – repository unit tests
- `cmd/zine-layout/cmds/api/integration_test.sh` – CLI smoke tests

---

## Next Steps for an Intern

1. **Start with database schema**: Copy the SQL from section 1.1 into `migrations.go` and test it loads correctly.
2. **Add one entity end-to-end**: Pick `ImageSequence`, implement repository → API → CLI command → test.
3. **Follow the pattern**: Once one entity works, the rest follow the same structure.
4. **Ask questions early**: If foreign keys fail or JSON serialization breaks, don't guess—check existing patterns in `pages.go` and `assets.go`.
5. **Test incrementally**: Don't write all repositories before testing. Get one working, then move to the next.

Good luck! 🚀
