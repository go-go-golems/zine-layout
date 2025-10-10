package sqlite

const schemaSQL = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

DROP TABLE IF EXISTS image_sequence_items;
DROP TABLE IF EXISTS image_sequences;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS spreads;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS projects;

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
`
