package serve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
	"github.com/go-go-golems/zine-layout/pkg/services"
)

// Project API responses -------------------------------------------------------

type projectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type assetResponse struct {
	ID          string         `json:"id"`
	ProjectID   string         `json:"project_id"`
	Filename    string         `json:"filename"`
	RelPath     string         `json:"rel_path"`
	ContentType string         `json:"content_type"`
	Bytes       int64          `json:"bytes"`
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	UploadedAt  time.Time      `json:"uploaded_at"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	URL         string         `json:"url"`
}

// Image sequence responses ----------------------------------------------------

type imageSequenceResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type imageSequenceItemResponse struct {
	Position int     `json:"position"`
	AssetID  *string `json:"asset_id,omitempty"`
	IsGap    bool    `json:"is_gap"`
}

// Layout template responses ---------------------------------------------------

type imageLayoutTemplateResponse struct {
	ID          string         `json:"id"`
	ProjectID   *string        `json:"project_id,omitempty"`
	Scope       string         `json:"scope"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Settings    map[string]any `json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Laid-out image responses ----------------------------------------------------

type laidOutImageResponse struct {
	ID         string                      `json:"id"`
	ProjectID  string                      `json:"project_id"`
	AssetID    string                      `json:"asset_id"`
	TemplateID string                      `json:"template_id"`
	Overrides  map[string]any              `json:"overrides,omitempty"`
	Result     *services.LayoutComputation `json:"result"`
	CreatedAt  time.Time                   `json:"created_at"`
	UpdatedAt  time.Time                   `json:"updated_at"`
}

// Layout sequence responses ---------------------------------------------------

type layoutSequenceResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type layoutSequenceItemResponse struct {
	Position       int    `json:"position"`
	LaidOutImageID string `json:"laid_out_image_id"`
}

// Page template responses ----------------------------------------------------

type pageTemplateResponse struct {
	ID          string         `json:"id"`
	ProjectID   *string        `json:"project_id,omitempty"`
	Scope       string         `json:"scope"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Template    map[string]any `json:"template"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Laid-out page responses ----------------------------------------------------

type laidOutPageResponse struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	PageTemplateID string         `json:"page_template_id"`
	LaidOutImageID string         `json:"laid_out_image_id"`
	Result         map[string]any `json:"result,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// Zine responses -------------------------------------------------------------

type zineResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type zinePageResponse struct {
	Position      int    `json:"position"`
	LaidOutPageID string `json:"laid_out_page_id"`
}

// Response helpers ------------------------------------------------------------

func projectToResponse(p *repo.Project) projectResponse {
	return projectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func assetToResponse(a *repo.Asset, projectsRoot string) assetResponse {
	metadata := map[string]any{}
	if strings.TrimSpace(a.MetadataJSON) != "" {
		_ = json.Unmarshal([]byte(a.MetadataJSON), &metadata)
	}
	url := fmt.Sprintf("/projects/%s/images/%s", a.ProjectID, a.Filename)
	if _, err := os.Stat(filepath.Join(projectsRoot, a.ProjectID, "images", a.Filename)); err != nil {
		url = ""
	}
	return assetResponse{
		ID:          a.ID,
		ProjectID:   a.ProjectID,
		Filename:    a.Filename,
		RelPath:     a.RelPath,
		ContentType: a.ContentType,
		Bytes:       a.Bytes,
		Width:       a.Width,
		Height:      a.Height,
		UploadedAt:  a.UploadedAt,
		Metadata:    metadata,
		URL:         url,
	}
}

func sequenceToResponse(seq *repo.ImageSequence) imageSequenceResponse {
	return imageSequenceResponse{
		ID:          seq.ID,
		ProjectID:   seq.ProjectID,
		Name:        seq.Name,
		Description: seq.Description,
		CreatedAt:   seq.CreatedAt,
		UpdatedAt:   seq.UpdatedAt,
	}
}

func sequenceItemsToResponse(items []*repo.ImageSequenceItem) []imageSequenceItemResponse {
	resp := make([]imageSequenceItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, imageSequenceItemResponse{
			Position: item.Position,
			AssetID:  item.AssetID,
			IsGap:    item.IsGap,
		})
	}
	return resp
}

func layoutTemplateToResponse(tpl *repo.ImageLayoutTemplate) imageLayoutTemplateResponse {
	scope := "project"
	if tpl.ProjectID == nil || *tpl.ProjectID == "" {
		scope = "global"
	}
	settings := map[string]any{}
	if strings.TrimSpace(tpl.SettingsJSON) != "" {
		_ = json.Unmarshal([]byte(tpl.SettingsJSON), &settings)
	}
	return imageLayoutTemplateResponse{
		ID:          tpl.ID,
		ProjectID:   tpl.ProjectID,
		Scope:       scope,
		Name:        tpl.Name,
		Description: tpl.Description,
		Settings:    settings,
		CreatedAt:   tpl.CreatedAt,
		UpdatedAt:   tpl.UpdatedAt,
	}
}

func laidOutImageToResponse(record *repo.LaidOutImage) (*laidOutImageResponse, error) {
	var overrides map[string]any
	if record.OverridesJSON != nil && strings.TrimSpace(*record.OverridesJSON) != "" {
		if err := json.Unmarshal([]byte(*record.OverridesJSON), &overrides); err != nil {
			return nil, fmt.Errorf("decode overrides: %w", err)
		}
	}
	var computation services.LayoutComputation
	if strings.TrimSpace(record.ResultJSON) != "" {
		if err := json.Unmarshal([]byte(record.ResultJSON), &computation); err != nil {
			return nil, fmt.Errorf("decode layout result: %w", err)
		}
	}
	return &laidOutImageResponse{
		ID:         record.ID,
		ProjectID:  record.ProjectID,
		AssetID:    record.AssetID,
		TemplateID: record.TemplateID,
		Overrides:  overrides,
		Result:     &computation,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}, nil
}

func layoutSequenceToResponse(seq *repo.LayoutSequence) layoutSequenceResponse {
	return layoutSequenceResponse{
		ID:          seq.ID,
		ProjectID:   seq.ProjectID,
		Name:        seq.Name,
		Description: seq.Description,
		CreatedAt:   seq.CreatedAt,
		UpdatedAt:   seq.UpdatedAt,
	}
}

func layoutSequenceItemsToResponse(items []*repo.LayoutSequenceItem) []layoutSequenceItemResponse {
	resp := make([]layoutSequenceItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, layoutSequenceItemResponse{
			Position:       item.Position,
			LaidOutImageID: item.LaidOutImageID,
		})
	}
	return resp
}

func pageTemplateToResponse(tpl *repo.PageTemplate) pageTemplateResponse {
	scope := "project"
	if tpl.ProjectID == nil || *tpl.ProjectID == "" {
		scope = "global"
	}
	template := map[string]any{}
	if strings.TrimSpace(tpl.TemplateJSON) != "" {
		_ = json.Unmarshal([]byte(tpl.TemplateJSON), &template)
	}
	return pageTemplateResponse{
		ID:          tpl.ID,
		ProjectID:   tpl.ProjectID,
		Scope:       scope,
		Name:        tpl.Name,
		Description: tpl.Description,
		Template:    template,
		CreatedAt:   tpl.CreatedAt,
		UpdatedAt:   tpl.UpdatedAt,
	}
}

func laidOutPageToResponse(page *repo.LaidOutPage) (*laidOutPageResponse, error) {
	var result map[string]any
	if page.ResultJSON != nil {
		if text := strings.TrimSpace(*page.ResultJSON); text != "" {
			if err := json.Unmarshal([]byte(text), &result); err != nil {
				return nil, fmt.Errorf("decode laid-out page result: %w", err)
			}
		}
	}
	return &laidOutPageResponse{
		ID:             page.ID,
		ProjectID:      page.ProjectID,
		PageTemplateID: page.PageTemplateID,
		LaidOutImageID: page.LaidOutImageID,
		Result:         result,
		CreatedAt:      page.CreatedAt,
		UpdatedAt:      page.UpdatedAt,
	}, nil
}

func zineToResponse(z *repo.Zine) zineResponse {
	return zineResponse{
		ID:          z.ID,
		ProjectID:   z.ProjectID,
		Name:        z.Name,
		Description: z.Description,
		CreatedAt:   z.CreatedAt,
		UpdatedAt:   z.UpdatedAt,
	}
}

func zinePagesToResponse(pages []*repo.ZinePage) []zinePageResponse {
	resp := make([]zinePageResponse, 0, len(pages))
	for _, page := range pages {
		resp = append(resp, zinePageResponse{
			Position:      page.Position,
			LaidOutPageID: page.LaidOutPageID,
		})
	}
	return resp
}
