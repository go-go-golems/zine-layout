package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
	"github.com/go-go-golems/zine-layout/pkg/spread"
	"github.com/go-go-golems/zine-layout/pkg/spread/simple"
)

// LayoutService orchestrates template application and laid-out image persistence.
type LayoutService struct {
	repos *repo.Repositories
}

// NewLayoutService constructs a layout service using the provided repositories.
func NewLayoutService(repos *repo.Repositories) *LayoutService {
	return &LayoutService{repos: repos}
}

// LayoutComputation captures the canonical persisted payload for a laid-out image.
type LayoutComputation struct {
	Settings       spread.Settings        `json:"settings"`
	Result         simple.Result          `json:"result"`
	PlacementTrace *simple.PlacementTrace `json:"placement_trace,omitempty"`
}

// CreateLaidOutImage renders placement metadata for an asset/template pair and persists the record.
func (s *LayoutService) CreateLaidOutImage(projectID, assetID, templateID string, overridesJSON *string) (*repo.LaidOutImage, error) {
	if s == nil || s.repos == nil {
		return nil, fmt.Errorf("layout service not initialized")
	}

	asset, err := s.repos.Assets.Get(assetID)
	if err != nil {
		return nil, fmt.Errorf("fetch asset: %w", err)
	}
	if asset.ProjectID != projectID {
		return nil, fmt.Errorf("asset %s does not belong to project %s", assetID, projectID)
	}

	template, err := s.repos.ImageLayoutTemplates.Get(templateID)
	if err != nil {
		return nil, fmt.Errorf("fetch template: %w", err)
	}
	if template.ProjectID != nil && *template.ProjectID != projectID {
		return nil, fmt.Errorf("template %s not available for project %s", templateID, projectID)
	}

	settings, normalizedOverrides, err := mergeTemplateSettings(template.SettingsJSON, overridesJSON)
	if err != nil {
		return nil, fmt.Errorf("merge settings: %w", err)
	}

	meta := spread.ImageMeta{Width: asset.Width, Height: asset.Height}
	inputs, err := simple.InputsFromSettings(settings, meta)
	if err != nil {
		return nil, fmt.Errorf("build inputs: %w", err)
	}

	trace := &simple.Trace{}
	result := simple.ComputePlacement(inputs, trace)

	payload := LayoutComputation{
		Settings:       settings,
		Result:         result,
		PlacementTrace: trace.Structured,
	}
	resultBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal layout computation: %w", err)
	}
	resultJSON := string(resultBytes)

	record := &repo.LaidOutImage{
		ProjectID:     projectID,
		AssetID:       assetID,
		TemplateID:    templateID,
		OverridesJSON: normalizedOverrides,
		ResultJSON:    resultJSON,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.repos.LaidOutImages.Create(record); err != nil {
		return nil, fmt.Errorf("persist laid-out image: %w", err)
	}
	return record, nil
}

// RecomputeLaidOutImage recalculates layout metadata for an existing record and persists updates.
func (s *LayoutService) RecomputeLaidOutImage(record *repo.LaidOutImage) error {
	if s == nil || s.repos == nil {
		return fmt.Errorf("layout service not initialized")
	}
	if record == nil {
		return fmt.Errorf("laid-out image record is nil")
	}

	asset, err := s.repos.Assets.Get(record.AssetID)
	if err != nil {
		return fmt.Errorf("fetch asset: %w", err)
	}
	if asset.ProjectID != record.ProjectID {
		return fmt.Errorf("asset %s does not belong to project %s", record.AssetID, record.ProjectID)
	}

	template, err := s.repos.ImageLayoutTemplates.Get(record.TemplateID)
	if err != nil {
		return fmt.Errorf("fetch template: %w", err)
	}
	if template.ProjectID != nil && *template.ProjectID != record.ProjectID {
		return fmt.Errorf("template %s not available for project %s", record.TemplateID, record.ProjectID)
	}

	settings, normalizedOverrides, err := mergeTemplateSettings(template.SettingsJSON, record.OverridesJSON)
	if err != nil {
		return fmt.Errorf("merge settings: %w", err)
	}

	meta := spread.ImageMeta{Width: asset.Width, Height: asset.Height}
	inputs, err := simple.InputsFromSettings(settings, meta)
	if err != nil {
		return fmt.Errorf("build inputs: %w", err)
	}

	trace := &simple.Trace{}
	result := simple.ComputePlacement(inputs, trace)
	payload := LayoutComputation{
		Settings:       settings,
		Result:         result,
		PlacementTrace: trace.Structured,
	}
	resultBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal layout computation: %w", err)
	}
	record.OverridesJSON = normalizedOverrides
	record.ResultJSON = string(resultBytes)
	record.UpdatedAt = time.Now().UTC()
	if err := s.repos.LaidOutImages.Update(record); err != nil {
		return fmt.Errorf("update laid-out image: %w", err)
	}
	return nil
}

// ApplyTemplateToSequence generates laid-out images for every asset in a sequence (skipping gaps).
func (s *LayoutService) ApplyTemplateToSequence(projectID, sequenceID, templateID string) ([]*repo.LaidOutImage, error) {
	if s == nil || s.repos == nil {
		return nil, fmt.Errorf("layout service not initialized")
	}

	sequence, err := s.repos.ImageSequences.Get(sequenceID)
	if err != nil {
		return nil, fmt.Errorf("fetch image sequence: %w", err)
	}
	if sequence.ProjectID != projectID {
		return nil, fmt.Errorf("sequence %s not in project %s", sequenceID, projectID)
	}

	items, err := s.repos.ImageSequences.ListItems(sequenceID)
	if err != nil {
		return nil, fmt.Errorf("list sequence items: %w", err)
	}

	var results []*repo.LaidOutImage
	for _, item := range items {
		if item.IsGap || item.AssetID == nil || *item.AssetID == "" {
			continue
		}
		record, err := s.CreateLaidOutImage(projectID, *item.AssetID, templateID, nil)
		if err != nil {
			return nil, err
		}
		results = append(results, record)
	}
	return results, nil
}

// mergeTemplateSettings applies override JSON onto base template settings and returns canonical override JSON.
func mergeTemplateSettings(baseJSON string, overridesJSON *string) (spread.Settings, *string, error) {
	var baseMap map[string]any
	if err := json.Unmarshal([]byte(baseJSON), &baseMap); err != nil {
		return spread.Settings{}, nil, fmt.Errorf("decode template settings: %w", err)
	}

	var normalizedOverrides *string
	if overridesJSON != nil && *overridesJSON != "" {
		var overridesMap map[string]any
		if err := json.Unmarshal([]byte(*overridesJSON), &overridesMap); err != nil {
			return spread.Settings{}, nil, fmt.Errorf("decode overrides: %w", err)
		}
		baseMap = deepMerge(baseMap, overridesMap)
		if len(overridesMap) > 0 {
			buf, err := json.Marshal(overridesMap)
			if err != nil {
				return spread.Settings{}, nil, fmt.Errorf("encode overrides: %w", err)
			}
			str := strings.TrimSpace(string(buf))
			if str != "" && str != "null" && str != "{}" {
				normalizedOverrides = new(string)
				*normalizedOverrides = str
			}
		}
	}

	mergedBytes, err := json.Marshal(baseMap)
	if err != nil {
		return spread.Settings{}, nil, fmt.Errorf("encode merged settings: %w", err)
	}

	var settings spread.Settings
	if err := json.Unmarshal(mergedBytes, &settings); err != nil {
		return spread.Settings{}, nil, fmt.Errorf("decode merged settings: %w", err)
	}
	return settings, normalizedOverrides, nil
}

func deepMerge(base, overrides map[string]any) map[string]any {
	for k, v := range overrides {
		if v == nil {
			base[k] = nil
			continue
		}
		if overrideMap, ok := v.(map[string]any); ok {
			if existing, exists := base[k]; exists {
				if existingMap, okExisting := existing.(map[string]any); okExisting {
					base[k] = deepMerge(existingMap, overrideMap)
					continue
				}
			}
			base[k] = deepMerge(make(map[string]any), overrideMap)
			continue
		}
		base[k] = v
	}
	return base
}
