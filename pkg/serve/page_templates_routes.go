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

// Project page template routes ------------------------------------------------

func (s *Server) handleProjectPageTemplates(w http.ResponseWriter, r *http.Request, projectID string) {
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
		projectTemplates, err := s.repos.PageTemplates.ListByProject(projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		globalTemplates, err := s.repos.PageTemplates.ListGlobal()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]pageTemplateResponse, 0, len(projectTemplates)+len(globalTemplates))
		for _, tpl := range projectTemplates {
			resp = append(resp, pageTemplateToResponse(tpl))
		}
		for _, tpl := range globalTemplates {
			resp = append(resp, pageTemplateToResponse(tpl))
		}
		writeJSON(w, http.StatusOK, map[string]any{"page_templates": resp})
	case http.MethodPost:
		var req struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Template    map[string]any `json:"template"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Template == nil {
			respondError(w, http.StatusBadRequest, "template payload is required")
			return
		}
		templateBytes, err := json.Marshal(req.Template)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid template payload: %v", err))
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = "Untitled Page Template"
		}
		projectIDCopy := projectID
		record := &repo.PageTemplate{
			ProjectID:    &projectIDCopy,
			Name:         name,
			Description:  strings.TrimSpace(req.Description),
			TemplateJSON: string(templateBytes),
		}
		if err := s.repos.PageTemplates.Create(record); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"page_template": pageTemplateToResponse(record)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Global page template routes -------------------------------------------------

func (s *Server) handlePageTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		templates, err := s.repos.PageTemplates.ListGlobal()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]pageTemplateResponse, 0, len(templates))
		for _, tpl := range templates {
			resp = append(resp, pageTemplateToResponse(tpl))
		}
		writeJSON(w, http.StatusOK, map[string]any{"page_templates": resp})
	case http.MethodPost:
		var req struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Template    map[string]any `json:"template"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Template == nil {
			respondError(w, http.StatusBadRequest, "template payload is required")
			return
		}
		templateBytes, err := json.Marshal(req.Template)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid template payload: %v", err))
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = "Untitled Page Template"
		}
		record := &repo.PageTemplate{
			Name:         name,
			Description:  strings.TrimSpace(req.Description),
			TemplateJSON: string(templateBytes),
		}
		if err := s.repos.PageTemplates.Create(record); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"page_template": pageTemplateToResponse(record)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePageTemplateRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/page-templates/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	templateID := parts[0]
	if templateID == "" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		tpl, err := s.repos.PageTemplates.Get(templateID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"page_template": pageTemplateToResponse(tpl)})
	case http.MethodPatch, http.MethodPut:
		tpl, err := s.repos.PageTemplates.Get(templateID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var req struct {
			Name        *string         `json:"name"`
			Description *string         `json:"description"`
			Template    *map[string]any `json:"template"`
			ProjectID   *string         `json:"project_id"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				name = "Untitled Page Template"
			}
			tpl.Name = name
		}
		if req.Description != nil {
			tpl.Description = strings.TrimSpace(*req.Description)
		}
		if req.ProjectID != nil {
			projectID := strings.TrimSpace(*req.ProjectID)
			if projectID == "" {
				tpl.ProjectID = nil
			} else {
				projectIDCopy := projectID
				tpl.ProjectID = &projectIDCopy
			}
		}
		if req.Template != nil {
			templateBytes, err := json.Marshal(*req.Template)
			if err != nil {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid template payload: %v", err))
				return
			}
			tpl.TemplateJSON = string(templateBytes)
		}
		tpl.UpdatedAt = time.Now().UTC()
		if err := s.repos.PageTemplates.Update(tpl); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"page_template": pageTemplateToResponse(tpl)})
	case http.MethodDelete:
		if err := s.repos.PageTemplates.Delete(templateID); err != nil {
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
