package sqlite

const schemaSQL = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_projects_updated ON projects(updated_at DESC);

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

CREATE TABLE IF NOT EXISTS image_layout_templates (
    id TEXT PRIMARY KEY,
    project_id TEXT,
    name TEXT NOT NULL,
    description TEXT,
    settings_json TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_layout_templates_project ON image_layout_templates(project_id, name);

CREATE TABLE IF NOT EXISTS laid_out_images (
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

CREATE INDEX IF NOT EXISTS idx_laid_out_images_project ON laid_out_images(project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_laid_out_images_asset ON laid_out_images(asset_id, updated_at DESC);

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

CREATE TABLE IF NOT EXISTS image_sequence_items (
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

CREATE TABLE IF NOT EXISTS layout_sequences (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_layout_sequences_project ON layout_sequences(project_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS layout_sequence_items (
    sequence_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    laid_out_image_id TEXT NOT NULL,
    PRIMARY KEY (sequence_id, position),
    FOREIGN KEY (sequence_id) REFERENCES layout_sequences(id) ON DELETE CASCADE,
    FOREIGN KEY (laid_out_image_id) REFERENCES laid_out_images(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS page_templates (
    id TEXT PRIMARY KEY,
    project_id TEXT,
    name TEXT NOT NULL,
    description TEXT,
    template_json TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_page_templates_project ON page_templates(project_id, name);

CREATE TABLE IF NOT EXISTS laid_out_pages (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    page_template_id TEXT NOT NULL,
    laid_out_image_id TEXT NOT NULL,
    result_json TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (page_template_id) REFERENCES page_templates(id) ON DELETE RESTRICT,
    FOREIGN KEY (laid_out_image_id) REFERENCES laid_out_images(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_laid_out_pages_project ON laid_out_pages(project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_laid_out_pages_image ON laid_out_pages(laid_out_image_id);

CREATE TABLE IF NOT EXISTS zines (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_zines_project ON zines(project_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS zine_pages (
    zine_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    laid_out_page_id TEXT NOT NULL,
    PRIMARY KEY (zine_id, position),
    FOREIGN KEY (zine_id) REFERENCES zines(id) ON DELETE CASCADE,
    FOREIGN KEY (laid_out_page_id) REFERENCES laid_out_pages(id) ON DELETE CASCADE
);
`
