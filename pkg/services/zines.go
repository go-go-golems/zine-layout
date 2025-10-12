package services

import (
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

// ZinesService orchestrates management of zines and their ordered pages.
type ZinesService struct {
	repos *repo.Repositories
}

// NewZinesService constructs a zine service.
func NewZinesService(repos *repo.Repositories) *ZinesService {
	return &ZinesService{repos: repos}
}

// CreateZine constructs a zine and optionally seeds it with laid-out pages.
func (s *ZinesService) CreateZine(projectID, name, description string, laidOutPageIDs []string) (*repo.Zine, []*repo.ZinePage, error) {
	if s == nil || s.repos == nil {
		return nil, nil, fmt.Errorf("zines service not initialized")
	}
	if projectID == "" {
		return nil, nil, fmt.Errorf("project id is required")
	}
	if name == "" {
		name = "Untitled Zine"
	}

	if _, err := s.repos.Projects.Get(projectID); err != nil {
		return nil, nil, fmt.Errorf("fetch project: %w", err)
	}

	pages := make([]*repo.ZinePage, 0, len(laidOutPageIDs))
	for idx, pageID := range laidOutPageIDs {
		if pageID == "" {
			continue
		}
		page, err := s.repos.LaidOutPages.Get(pageID)
		if err != nil {
			return nil, nil, fmt.Errorf("fetch laid-out page %s: %w", pageID, err)
		}
		if page.ProjectID != projectID {
			return nil, nil, fmt.Errorf("laid-out page %s does not belong to project %s", pageID, projectID)
		}
		pages = append(pages, &repo.ZinePage{
			ZineID:        "",
			Position:      idx,
			LaidOutPageID: pageID,
		})
	}

	zine := &repo.Zine{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := s.repos.Zines.Create(zine); err != nil {
		return nil, nil, fmt.Errorf("create zine: %w", err)
	}

	if len(pages) > 0 {
		for _, page := range pages {
			page.ZineID = zine.ID
		}
		if err := s.repos.Zines.SetPages(zine.ID, pages); err != nil {
			return nil, nil, fmt.Errorf("persist zine pages: %w", err)
		}
	}

	persisted, err := s.repos.Zines.GetPages(zine.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("reload zine pages: %w", err)
	}
	return zine, persisted, nil
}

// UpdateZine updates metadata fields and refreshes the updated timestamp.
func (s *ZinesService) UpdateZine(zine *repo.Zine) error {
	if s == nil || s.repos == nil {
		return fmt.Errorf("zines service not initialized")
	}
	if zine == nil {
		return fmt.Errorf("zine payload is nil")
	}
	zine.UpdatedAt = time.Now().UTC()
	if err := s.repos.Zines.Update(zine); err != nil {
		return fmt.Errorf("update zine: %w", err)
	}
	return nil
}

// UpdateZinePages replaces the ordered laid-out pages in a zine.
func (s *ZinesService) UpdateZinePages(zineID string, laidOutPageIDs []string) error {
	if s == nil || s.repos == nil {
		return fmt.Errorf("zines service not initialized")
	}
	zine, err := s.repos.Zines.Get(zineID)
	if err != nil {
		return fmt.Errorf("fetch zine: %w", err)
	}

	pages := make([]*repo.ZinePage, 0, len(laidOutPageIDs))
	for idx, pageID := range laidOutPageIDs {
		if pageID == "" {
			continue
		}
		page, err := s.repos.LaidOutPages.Get(pageID)
		if err != nil {
			return fmt.Errorf("fetch laid-out page %s: %w", pageID, err)
		}
		if page.ProjectID != zine.ProjectID {
			return fmt.Errorf("laid-out page %s does not belong to project %s", pageID, zine.ProjectID)
		}
		pages = append(pages, &repo.ZinePage{
			ZineID:        zineID,
			Position:      idx,
			LaidOutPageID: pageID,
		})
	}

	if err := s.repos.Zines.SetPages(zineID, pages); err != nil {
		return fmt.Errorf("set zine pages: %w", err)
	}

	zine.UpdatedAt = time.Now().UTC()
	if err := s.repos.Zines.Update(zine); err != nil {
		return fmt.Errorf("touch zine: %w", err)
	}
	return nil
}

// GetZineWithPages retrieves a zine and its ordered pages.
func (s *ZinesService) GetZineWithPages(zineID string) (*repo.Zine, []*repo.ZinePage, error) {
	if s == nil || s.repos == nil {
		return nil, nil, fmt.Errorf("zines service not initialized")
	}
	zine, err := s.repos.Zines.Get(zineID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch zine: %w", err)
	}
	pages, err := s.repos.Zines.GetPages(zineID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch zine pages: %w", err)
	}
	return zine, pages, nil
}

// DeleteZine removes a zine and its associated page ordering.
func (s *ZinesService) DeleteZine(zineID string) error {
	if s == nil || s.repos == nil {
		return fmt.Errorf("zines service not initialized")
	}
	if _, err := s.repos.Zines.Get(zineID); err != nil {
		return fmt.Errorf("fetch zine: %w", err)
	}
	if err := s.repos.Zines.Delete(zineID); err != nil {
		return fmt.Errorf("delete zine: %w", err)
	}
	return nil
}
