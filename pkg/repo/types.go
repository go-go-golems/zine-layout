package repo

import "time"

// Project represents a zine project persisted in SQL.
type Project struct {
	ID           string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PresetID     *string
	CoverAssetID *string
}

// Asset represents a project image/asset stored on disk.
type Asset struct {
	ID          string
	ProjectID   string
	Filename    string
	RelPath     string
	ContentType string
	Bytes       int64
	Width       int
	Height      int
	SortIndex   int
	CreatedAt   time.Time
}

// Page captures layout settings/result for a single page.
type Page struct {
	ProjectID    string
	PageNumber   int
	AssetID      *string
	SettingsJSON string
	ResultJSON   *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Spread captures layout settings/result for a spread (pair of facing pages).
type Spread struct {
	ProjectID       string
	SpreadNumber    int
	LeftPageNumber  *int
	RightPageNumber *int
	SettingsJSON    string
	ResultJSON      *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ProjectRepository manages project metadata persistence.
type ProjectRepository interface {
	Create(project *Project) error
	Update(project *Project) error
	Get(id string) (*Project, error)
	List() ([]*Project, error)
	Delete(id string) error
}

// AssetRepository manages project assets.
type AssetRepository interface {
	Create(asset *Asset) error
	Get(projectID, assetID string) (*Asset, error)
	ListByProject(projectID string) ([]*Asset, error)
	UpdateOrder(projectID string, orderedIDs []string) error
	Delete(projectID, assetID string) error
}

// PageRepository manages project pages.
type PageRepository interface {
	Upsert(page *Page) error
	GetByNumber(projectID string, pageNumber int) (*Page, error)
	List(projectID string) ([]*Page, error)
	Delete(projectID string, pageNumber int) error
}

// SpreadRepository manages project spreads.
type SpreadRepository interface {
	Upsert(spread *Spread) error
	GetByNumber(projectID string, spreadNumber int) (*Spread, error)
	List(projectID string) ([]*Spread, error)
	Delete(projectID string, spreadNumber int) error
}

// Repositories bundles individual repositories together.
type Repositories struct {
	Projects ProjectRepository
	Assets   AssetRepository
	Pages    PageRepository
	Spreads  SpreadRepository
}
