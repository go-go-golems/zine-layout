package serve

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

// Project-scoped layout templates --------------------------------------------

func (s *Server) handleProjectLayoutTemplates(w http.ResponseWriter, r *http.Request, projectID string) {
	switch r.Method {
	case http.MethodGet:
		if _, err := s.repos.Projects.Get(projectID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		projectTemplates, err := s.repos.ImageLayoutTemplates.ListByProject(projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		globalTemplates, err := s.repos.ImageLayoutTemplates.ListGlobal()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]imageLayoutTemplateResponse, 0, len(projectTemplates)+len(globalTemplates))
		for _, tpl := range projectTemplates {
			resp = append(resp, layoutTemplateToResponse(tpl))
		}
		for _, tpl := range globalTemplates {
			resp = append(resp, layoutTemplateToResponse(tpl))
		}
		writeJSON(w, http.StatusOK, map[string]any{"templates": resp})
	case http.MethodPost:
		var req struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Settings    map[string]any `json:"settings"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Settings == nil {
			respondError(w, http.StatusBadRequest, "settings are required")
			return
		}
		settingsBytes, err := json.Marshal(req.Settings)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid settings: %v", err))
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = "Untitled Template"
		}
		projectIDCopy := projectID
		template := &repo.ImageLayoutTemplate{
			ProjectID:    &projectIDCopy,
			Name:         name,
			Description:  strings.TrimSpace(req.Description),
			SettingsJSON: string(settingsBytes),
		}
		if err := s.repos.ImageLayoutTemplates.Create(template); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"template": layoutTemplateToResponse(template)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Global layout template routes ----------------------------------------------

func (s *Server) handleImageLayoutTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		templates, err := s.repos.ImageLayoutTemplates.ListGlobal()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]imageLayoutTemplateResponse, 0, len(templates))
		for _, tpl := range templates {
			resp = append(resp, layoutTemplateToResponse(tpl))
		}
		writeJSON(w, http.StatusOK, map[string]any{"templates": resp})
	case http.MethodPost:
		var req struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Settings    map[string]any `json:"settings"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Settings == nil {
			respondError(w, http.StatusBadRequest, "settings are required")
			return
		}
		settingsBytes, err := json.Marshal(req.Settings)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid settings: %v", err))
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = "Untitled Template"
		}
		template := &repo.ImageLayoutTemplate{
			Name:         name,
			Description:  strings.TrimSpace(req.Description),
			SettingsJSON: string(settingsBytes),
		}
		if err := s.repos.ImageLayoutTemplates.Create(template); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"template": layoutTemplateToResponse(template)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleImageLayoutTemplateRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/image-layout-templates/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	templateID := strings.Trim(path, "/")
	if templateID == "" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		template, err := s.repos.ImageLayoutTemplates.Get(templateID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"template": layoutTemplateToResponse(template)})
	case http.MethodPatch, http.MethodPut:
		template, err := s.repos.ImageLayoutTemplates.Get(templateID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var req struct {
			Name        *string        `json:"name"`
			Description *string        `json:"description"`
			Settings    map[string]any `json:"settings"`
			ProjectID   *string        `json:"project_id"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				name = "Untitled Template"
			}
			template.Name = name
		}
		if req.Description != nil {
			template.Description = strings.TrimSpace(*req.Description)
		}
		if req.ProjectID != nil {
			project := strings.TrimSpace(*req.ProjectID)
			if project == "" {
				template.ProjectID = nil
			} else {
				template.ProjectID = &project
			}
		}
		if req.Settings != nil {
			settingsBytes, err := json.Marshal(req.Settings)
			if err != nil {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid settings: %v", err))
				return
			}
			template.SettingsJSON = string(settingsBytes)
		}
		template.UpdatedAt = time.Now().UTC()
		if err := s.repos.ImageLayoutTemplates.Update(template); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"template": layoutTemplateToResponse(template)})
	case http.MethodDelete:
		if err := s.repos.ImageLayoutTemplates.Delete(templateID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PATCH, PUT, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
