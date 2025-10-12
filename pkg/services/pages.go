package services

import (
    "encoding/json"
    "errors"
    "fmt"
    "image"
    _ "image/gif"
    _ "image/jpeg"
    "image/png"
    "os"
    "path/filepath"
    "time"

    "github.com/go-go-golems/zine-layout/pkg/pagelayout"
    "github.com/go-go-golems/zine-layout/pkg/pagelayout/renderer"
    "github.com/go-go-golems/zine-layout/pkg/repo"
)

// ErrPageRendererNotImplemented signals that actual rendering is pending implementation.
var ErrPageRendererNotImplemented = errors.New("page rendering not implemented")

// PagesService orchestrates creation and management of laid-out pages.
type PagesService struct {
	repos *repo.Repositories
    dataRoot string
}

// NewPagesService constructs a new pages service.
func NewPagesService(repos *repo.Repositories) *PagesService {
	return &PagesService{repos: repos}
}

// SetDataRoot configures where project files are stored on disk.
func (s *PagesService) SetDataRoot(root string) {
    s.dataRoot = root
}

// CreatePage creates a print-ready page by placing one laid-out image on a physical page.
// The page template defines page size, margins, spread mode, gutter, and image positioning.
func (s *PagesService) CreatePage(projectID, pageTemplateID, laidOutImageID string) (*repo.LaidOutPage, error) {
	if s == nil || s.repos == nil {
		return nil, fmt.Errorf("pages service not initialized")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project id is required")
	}
	if pageTemplateID == "" {
		return nil, fmt.Errorf("page template id is required")
	}
	if laidOutImageID == "" {
		return nil, fmt.Errorf("laid-out image id is required")
	}

	// Validate page template exists and is accessible
	tpl, err := s.repos.PageTemplates.Get(pageTemplateID)
	if err != nil {
		return nil, fmt.Errorf("fetch page template: %w", err)
	}
	if tpl.ProjectID != nil && *tpl.ProjectID != projectID {
		return nil, fmt.Errorf("template %s not available for project %s", pageTemplateID, projectID)
	}

	// Validate laid-out image exists and belongs to project
	image, err := s.repos.LaidOutImages.Get(laidOutImageID)
	if err != nil {
		return nil, fmt.Errorf("fetch laid-out image: %w", err)
	}
	if image.ProjectID != projectID {
		return nil, fmt.Errorf("laid-out image %s does not belong to project %s", laidOutImageID, projectID)
	}

	// Create the print page
	page := &repo.LaidOutPage{
		ProjectID:      projectID,
		PageTemplateID: pageTemplateID,
		LaidOutImageID: laidOutImageID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := s.repos.LaidOutPages.Create(page); err != nil {
		return nil, fmt.Errorf("create laid-out page: %w", err)
	}

	return page, nil
}

// UpdatePageImage changes which laid-out image is used for an existing print page.
func (s *PagesService) UpdatePageImage(pageID, laidOutImageID string) error {
	if s == nil || s.repos == nil {
		return fmt.Errorf("pages service not initialized")
	}
	if laidOutImageID == "" {
		return fmt.Errorf("laid-out image id is required")
	}

	page, err := s.repos.LaidOutPages.Get(pageID)
	if err != nil {
		return fmt.Errorf("fetch laid-out page: %w", err)
	}

	// Validate new laid-out image
	image, err := s.repos.LaidOutImages.Get(laidOutImageID)
	if err != nil {
		return fmt.Errorf("fetch laid-out image: %w", err)
	}
	if image.ProjectID != page.ProjectID {
		return fmt.Errorf("laid-out image %s does not belong to project %s", laidOutImageID, page.ProjectID)
	}

	// Update the page
	page.LaidOutImageID = laidOutImageID
	page.UpdatedAt = time.Now().UTC()
	// Clear result JSON since image changed (will need re-render)
	page.ResultJSON = nil

	if err := s.repos.LaidOutPages.Update(page); err != nil {
		return fmt.Errorf("update laid-out page: %w", err)
	}
	return nil
}

// GetPage returns a laid-out page by ID.
func (s *PagesService) GetPage(pageID string) (*repo.LaidOutPage, error) {
	if s == nil || s.repos == nil {
		return nil, fmt.Errorf("pages service not initialized")
	}
	page, err := s.repos.LaidOutPages.Get(pageID)
	if err != nil {
		return nil, fmt.Errorf("fetch laid-out page: %w", err)
	}
	return page, nil
}

// DeletePage removes a laid-out page.
func (s *PagesService) DeletePage(pageID string) error {
	if s == nil || s.repos == nil {
		return fmt.Errorf("pages service not initialized")
	}
	if _, err := s.repos.LaidOutPages.Get(pageID); err != nil {
		return fmt.Errorf("fetch laid-out page: %w", err)
	}
	if err := s.repos.LaidOutPages.Delete(pageID); err != nil {
		return fmt.Errorf("delete laid-out page: %w", err)
	}
	return nil
}

// PageRenderMetadata summarizes generated variant files for a laid-out page.
type PageRenderMetadata struct {
    Width    int               `json:"width"`
    Height   int               `json:"height"`
    Variants map[string]string `json:"variants"` // variant name -> rel path (projects/...)
}

// RenderPage renders the laid-out page to files and persists metadata on the record.
func (s *PagesService) RenderPage(pageID string) (*repo.LaidOutPage, error) {
    if s == nil || s.repos == nil {
        return nil, fmt.Errorf("pages service not initialized")
    }
    page, err := s.repos.LaidOutPages.Get(pageID)
    if err != nil {
        return nil, fmt.Errorf("fetch laid-out page: %w", err)
    }
    // Load dependencies
    tpl, err := s.repos.PageTemplates.Get(page.PageTemplateID)
    if err != nil {
        return nil, fmt.Errorf("fetch page template: %w", err)
    }
    laid, err := s.repos.LaidOutImages.Get(page.LaidOutImageID)
    if err != nil {
        return nil, fmt.Errorf("fetch laid-out image: %w", err)
    }
    asset, err := s.repos.Assets.Get(laid.AssetID)
    if err != nil {
        return nil, fmt.Errorf("fetch asset: %w", err)
    }

    // Decode page settings
    var settings pagelayout.PageLayoutSettings
    if err := json.Unmarshal([]byte(tpl.TemplateJSON), &settings); err != nil {
        return nil, fmt.Errorf("decode page template settings: %w", err)
    }
    if err := settings.Canonicalize(); err != nil {
        return nil, fmt.Errorf("invalid page layout settings: %w", err)
    }

    // Decode layout computation (for crop geometry)
    var comp LayoutComputation
    if err := json.Unmarshal([]byte(laid.ResultJSON), &comp); err != nil {
        return nil, fmt.Errorf("decode laid-out image result: %w", err)
    }

    // Load source image from disk
    if s.dataRoot == "" {
        return nil, fmt.Errorf("pages service dataRoot not configured")
    }
    absPath := filepath.Join(s.dataRoot, filepath.FromSlash(asset.RelPath))
    f, err := os.Open(absPath)
    if err != nil {
        return nil, fmt.Errorf("open asset image: %w", err)
    }
    defer f.Close()
    srcImg, _, err := image.Decode(f)
    if err != nil {
        return nil, fmt.Errorf("decode image: %w", err)
    }

    // Render
    ctx := renderer.RenderContext{
        Settings:     settings,
        Source:       srcImg,
        LayoutResult: &comp.Result,
    }
    result, err := renderer.RenderPage(ctx)
    if err != nil {
        return nil, fmt.Errorf("render page: %w", err)
    }

    // Write files under dataRoot/projects/{projectID}/pages/{pageID}/
    outDir := filepath.Join(s.dataRoot, "projects", page.ProjectID, "pages", page.ID)
    if err := os.MkdirAll(outDir, 0o755); err != nil {
        return nil, fmt.Errorf("prepare output dir: %w", err)
    }
    // always write thumbnail and full/combined; left/right only if present
    relBase := filepath.ToSlash(filepath.Join("projects", page.ProjectID, "pages", page.ID))
    variants := map[string]string{}

    // Helper to write a PNG
    writePNG := func(name string, img image.Image) error {
        if img == nil { return nil }
        outPath := filepath.Join(outDir, name+".png")
        fp, err := os.Create(outPath)
        if err != nil { return err }
        defer fp.Close()
        // Always encode as PNG for previews
        if err := png.Encode(fp, img); err != nil { return err }
        variants[name] = filepath.ToSlash(filepath.Join(relBase, name+".png"))
        return nil
    }

    // Collect and write
    if thumb, ok := result.Variants["thumbnail"]; ok { _ = writePNG("thumbnail", thumb) }
    if full, ok := result.Variants["full"]; ok { _ = writePNG("full", full) }
    if combined, ok := result.Variants["combined"]; ok { _ = writePNG("combined", combined) }
    if left, ok := result.Variants["left"]; ok { _ = writePNG("left", left) }
    if right, ok := result.Variants["right"]; ok { _ = writePNG("right", right) }

    meta := PageRenderMetadata{
        Width:    result.Full.Bounds().Dx(),
        Height:   result.Full.Bounds().Dy(),
        Variants: variants,
    }
    metaBytes, err := json.Marshal(meta)
    if err != nil {
        return nil, fmt.Errorf("encode render metadata: %w", err)
    }
    metaStr := string(metaBytes)
    page.ResultJSON = &metaStr
    page.UpdatedAt = time.Now().UTC()
    if err := s.repos.LaidOutPages.Update(page); err != nil {
        return nil, fmt.Errorf("persist laid-out page: %w", err)
    }
    return page, nil
}
