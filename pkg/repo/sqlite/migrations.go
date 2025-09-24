package sqlite

const schemaSQL = `
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    preset_id TEXT,
    cover_asset_id TEXT
);

CREATE TABLE IF NOT EXISTS assets (
    project_id TEXT NOT NULL,
    id TEXT NOT NULL,
    filename TEXT NOT NULL,
    rel_path TEXT NOT NULL,
    content_type TEXT,
    bytes INTEGER NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    sort_index INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (project_id, id),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assets_project_order ON assets(project_id, sort_index, created_at);

CREATE TABLE IF NOT EXISTS pages (
    project_id TEXT NOT NULL,
    page_number INTEGER NOT NULL,
    asset_id TEXT,
    settings_json TEXT NOT NULL,
    result_json TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (project_id, page_number),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id, asset_id) REFERENCES assets(project_id, id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS spreads (
    project_id TEXT NOT NULL,
    spread_number INTEGER NOT NULL,
    left_page_number INTEGER,
    right_page_number INTEGER,
    settings_json TEXT NOT NULL,
    result_json TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (project_id, spread_number),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id, left_page_number) REFERENCES pages(project_id, page_number) ON DELETE SET NULL,
    FOREIGN KEY (project_id, right_page_number) REFERENCES pages(project_id, page_number) ON DELETE SET NULL
);
`
