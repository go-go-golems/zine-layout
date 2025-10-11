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

// CreatePage instantiates a laid-out page from a template and optional laid-out images.
func (s *PagesService) CreatePage(projectID, pageTemplateID string, laidOutImageIDs []string) (*repo.LaidOutPage, []*repo.LaidOutPageInput, error) {
	if s == nil || s.repos == nil {
		return nil, nil, fmt.Errorf("pages service not initialized")
	}
	if projectID == "" {
		return nil, nil, fmt.Errorf("project id is required")
	}
	if pageTemplateID == "" {
		return nil, nil, fmt.Errorf("page template id is required")
	}

	tpl, err := s.repos.PageTemplates.Get(pageTemplateID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch page template: %w", err)
	}
	if tpl.ProjectID != nil && *tpl.ProjectID != projectID {
		return nil, nil, fmt.Errorf("template %s not available for project %s", pageTemplateID, projectID)
	}

	inputs := make([]*repo.LaidOutPageInput, 0, len(laidOutImageIDs))
	for idx, imageID := range laidOutImageIDs {
		if imageID == "" {
			continue
		}
		image, err := s.repos.LaidOutImages.Get(imageID)
		if err != nil {
			return nil, nil, fmt.Errorf("fetch laid-out image %s: %w", imageID, err)
		}
		if image.ProjectID != projectID {
			return nil, nil, fmt.Errorf("laid-out image %s does not belong to project %s", imageID, projectID)
		}
		inputs = append(inputs, &repo.LaidOutPageInput{
			PageID:         "",
			InputIndex:     idx,
			LaidOutImageID: imageID,
		})
	}

	page := &repo.LaidOutPage{
		ProjectID:      projectID,
		PageTemplateID: pageTemplateID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := s.repos.LaidOutPages.Create(page); err != nil {
		return nil, nil, fmt.Errorf("create laid-out page: %w", err)
	}

	if len(inputs) > 0 {
		for _, input := range inputs {
			input.PageID = page.ID
		}
		if err := s.repos.LaidOutPages.SetInputs(page.ID, inputs); err != nil {
			return nil, nil, fmt.Errorf("persist page inputs: %w", err)
		}
	}

	persistedInputs, err := s.repos.LaidOutPages.GetInputs(page.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("reload page inputs: %w", err)
	}
	return page, persistedInputs, nil
}

// UpdatePageInputs replaces the laid-out image inputs for an existing page.
func (s *PagesService) UpdatePageInputs(pageID string, laidOutImageIDs []string) error {
	if s == nil || s.repos == nil {
		return fmt.Errorf("pages service not initialized")
	}
	page, err := s.repos.LaidOutPages.Get(pageID)
	if err != nil {
		return fmt.Errorf("fetch laid-out page: %w", err)
	}

	inputs := make([]*repo.LaidOutPageInput, 0, len(laidOutImageIDs))
	for idx, imageID := range laidOutImageIDs {
		if imageID == "" {
			continue
		}
		image, err := s.repos.LaidOutImages.Get(imageID)
		if err != nil {
			return fmt.Errorf("fetch laid-out image %s: %w", imageID, err)
		}
		if image.ProjectID != page.ProjectID {
			return fmt.Errorf("laid-out image %s does not belong to project %s", imageID, page.ProjectID)
		}
		inputs = append(inputs, &repo.LaidOutPageInput{
			PageID:         pageID,
			InputIndex:     idx,
			LaidOutImageID: imageID,
		})
	}

	if err := s.repos.LaidOutPages.SetInputs(pageID, inputs); err != nil {
		return fmt.Errorf("set laid-out page inputs: %w", err)
	}

	page.UpdatedAt = time.Now().UTC()
	if err := s.repos.LaidOutPages.Update(page); err != nil {
		return fmt.Errorf("touch laid-out page: %w", err)
	}
	return nil
}

// GetPageWithInputs returns a laid-out page and its ordered inputs.
func (s *PagesService) GetPageWithInputs(pageID string) (*repo.LaidOutPage, []*repo.LaidOutPageInput, error) {
	if s == nil || s.repos == nil {
		return nil, nil, fmt.Errorf("pages service not initialized")
	}
	page, err := s.repos.LaidOutPages.Get(pageID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch laid-out page: %w", err)
	}
	inputs, err := s.repos.LaidOutPages.GetInputs(pageID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch page inputs: %w", err)
	}
	return page, inputs, nil
}

// DeletePage removes a laid-out page and its associated inputs.
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
