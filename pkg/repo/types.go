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

// Repositories aggregates all persistence adapters.
type Repositories struct {
	Projects       ProjectRepository
	Assets         AssetRepository
	ImageSequences ImageSequenceRepository
}
