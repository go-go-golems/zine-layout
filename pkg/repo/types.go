package repo

import "time"

// Project is the top-level workspace for assets and layout artifacts.
type Project struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Asset represents a raw uploaded image stored on disk.
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
	MetadataJSON string
}

// ImageSequence captures a named ordering of assets within a project.
type ImageSequence struct {
	ID          string
	ProjectID   string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ImageSequenceItem defines a single position inside an image sequence.
type ImageSequenceItem struct {
	SequenceID string
	Position   int
	AssetID    *string
	IsGap      bool
}

// ImageLayoutTemplate defines reusable layout settings for assets.
type ImageLayoutTemplate struct {
	ID           string
	ProjectID    *string
	Name         string
	Description  string
	SettingsJSON string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// LaidOutImage stores the rendered placement metadata for an asset and template pair.
type LaidOutImage struct {
	ID            string
	ProjectID     string
	AssetID       string
	TemplateID    string
	OverridesJSON *string
	ResultJSON    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// LayoutSequence orders laid-out images for downstream page/zine composition.
type LayoutSequence struct {
	ID          string
	ProjectID   string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// LayoutSequenceItem defines a single position within a layout sequence.
type LayoutSequenceItem struct {
	SequenceID     string
	Position       int
	LaidOutImageID string
}

// PageTemplate defines how laid-out images should be composed on a page.
type PageTemplate struct {
	ID           string
	ProjectID    *string
	Name         string
	Description  string
	TemplateJSON string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// LaidOutPage represents a concrete page instantiated from a template and inputs.
type LaidOutPage struct {
	ID             string
	ProjectID      string
	PageTemplateID string
	ResultJSON     *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// LaidOutPageInput associates a laid-out image with a page input slot.
type LaidOutPageInput struct {
	PageID         string
	InputIndex     int
	LaidOutImageID string
}

// Zine represents an ordered collection of laid-out pages.
type Zine struct {
	ID          string
	ProjectID   string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ZinePage connects a laid-out page to its position within a zine.
type ZinePage struct {
	ZineID        string
	Position      int
	LaidOutPageID string
}

// ProjectRepository manages project metadata persistence.
type ProjectRepository interface {
	Create(project *Project) error
	Update(project *Project) error
	Get(id string) (*Project, error)
	List() ([]*Project, error)
	Delete(id string) error
}

// AssetRepository manages raw asset storage metadata.
type AssetRepository interface {
	Create(asset *Asset) error
	Update(asset *Asset) error
	Get(id string) (*Asset, error)
	ListByProject(projectID string) ([]*Asset, error)
	Delete(id string) error
}

// ImageSequenceRepository manages named asset orderings.
type ImageSequenceRepository interface {
	Create(sequence *ImageSequence) error
	Update(sequence *ImageSequence) error
	Get(id string) (*ImageSequence, error)
	ListByProject(projectID string) ([]*ImageSequence, error)
	Delete(id string) error

	AddItem(item *ImageSequenceItem) error
	ListItems(sequenceID string) ([]*ImageSequenceItem, error)
	ReplaceItems(sequenceID string, items []*ImageSequenceItem) error
	DeleteItem(sequenceID string, position int) error
}

// ImageLayoutTemplateRepository manages reusable layout templates.
type ImageLayoutTemplateRepository interface {
	Create(tpl *ImageLayoutTemplate) error
	Update(tpl *ImageLayoutTemplate) error
	Get(id string) (*ImageLayoutTemplate, error)
	ListGlobal() ([]*ImageLayoutTemplate, error)
	ListByProject(projectID string) ([]*ImageLayoutTemplate, error)
	Delete(id string) error
}

// LaidOutImageRepository persists computed layout results.
type LaidOutImageRepository interface {
	Create(image *LaidOutImage) error
	Update(image *LaidOutImage) error
	Get(id string) (*LaidOutImage, error)
	ListByProject(projectID string) ([]*LaidOutImage, error)
	ListByAsset(assetID string) ([]*LaidOutImage, error)
	Delete(id string) error
}

// LayoutSequenceRepository manages ordering of laid-out images.
type LayoutSequenceRepository interface {
	Create(sequence *LayoutSequence) error
	Update(sequence *LayoutSequence) error
	Get(id string) (*LayoutSequence, error)
	ListByProject(projectID string) ([]*LayoutSequence, error)
	Delete(id string) error

	AddItem(item *LayoutSequenceItem) error
	ListItems(sequenceID string) ([]*LayoutSequenceItem, error)
	ReplaceItems(sequenceID string, items []*LayoutSequenceItem) error
	DeleteItem(sequenceID string, position int) error
}

// PageTemplateRepository manages reusable page composition templates.
type PageTemplateRepository interface {
	Create(tpl *PageTemplate) error
	Update(tpl *PageTemplate) error
	Get(id string) (*PageTemplate, error)
	ListGlobal() ([]*PageTemplate, error)
	ListByProject(projectID string) ([]*PageTemplate, error)
	Delete(id string) error
}

// LaidOutPageRepository persists composed pages and their input mapping.
type LaidOutPageRepository interface {
	Create(page *LaidOutPage) error
	Update(page *LaidOutPage) error
	Get(id string) (*LaidOutPage, error)
	ListByProject(projectID string) ([]*LaidOutPage, error)
	Delete(id string) error

	SetInputs(pageID string, inputs []*LaidOutPageInput) error
	GetInputs(pageID string) ([]*LaidOutPageInput, error)
}

// ZineRepository manages ordered collections of laid-out pages.
type ZineRepository interface {
	Create(zine *Zine) error
	Update(zine *Zine) error
	Get(id string) (*Zine, error)
	ListByProject(projectID string) ([]*Zine, error)
	Delete(id string) error

	SetPages(zineID string, pages []*ZinePage) error
	GetPages(zineID string) ([]*ZinePage, error)
}

// Repositories aggregates all persistence adapters.
type Repositories struct {
	Projects             ProjectRepository
	Assets               AssetRepository
	ImageSequences       ImageSequenceRepository
	ImageLayoutTemplates ImageLayoutTemplateRepository
	LaidOutImages        LaidOutImageRepository
	LayoutSequences      LayoutSequenceRepository
	PageTemplates        PageTemplateRepository
	LaidOutPages         LaidOutPageRepository
	Zines                ZineRepository
}
