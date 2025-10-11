package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

// ErrPageRendererNotImplemented signals that actual rendering is pending implementation.
var ErrPageRendererNotImplemented = errors.New("page rendering not implemented")

// PagesService orchestrates creation and management of laid-out pages.
type PagesService struct {
	repos *repo.Repositories
}

// NewPagesService constructs a new pages service.
func NewPagesService(repos *repo.Repositories) *PagesService {
	return &PagesService{repos: repos}
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

// RenderPage is a placeholder for the future rendering pipeline.
func (s *PagesService) RenderPage(pageID string) (*repo.LaidOutPage, error) {
	_, err := s.repos.LaidOutPages.Get(pageID)
	if err != nil {
		return nil, fmt.Errorf("fetch laid-out page: %w", err)
	}
	return nil, ErrPageRendererNotImplemented
}
