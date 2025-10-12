package serve

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/services"
)

// Project laid-out pages -----------------------------------------------------

func (s *Server) handleProjectLaidOutPages(w http.ResponseWriter, r *http.Request, projectID string) {
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
		records, err := s.repos.LaidOutPages.ListByProject(projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]*laidOutPageResponse, 0, len(records))
		for _, record := range records {
			converted, convErr := laidOutPageToResponse(record)
			if convErr != nil {
				respondError(w, http.StatusInternalServerError, convErr.Error())
				return
			}
			resp = append(resp, converted)
		}
		writeJSON(w, http.StatusOK, map[string]any{"laid_out_pages": resp})
	case http.MethodPost:
		if s.pages == nil {
			respondError(w, http.StatusInternalServerError, "pages service not initialized")
			return
		}
		var req struct {
			PageTemplateID string `json:"page_template_id"`
			LaidOutImageID string `json:"laid_out_image_id"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.TrimSpace(req.PageTemplateID) == "" {
			respondError(w, http.StatusBadRequest, "page_template_id is required")
			return
		}
		if strings.TrimSpace(req.LaidOutImageID) == "" {
			respondError(w, http.StatusBadRequest, "laid_out_image_id is required")
			return
		}
		record, err := s.pages.CreatePage(projectID, req.PageTemplateID, req.LaidOutImageID)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, convErr := laidOutPageToResponse(record)
		if convErr != nil {
			respondError(w, http.StatusInternalServerError, convErr.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"laid_out_page": resp})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Laid-out page resource -----------------------------------------------------

func (s *Server) handleLaidOutPageRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/laid-out-pages/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	pageID := parts[0]
	if pageID == "" {
		http.NotFound(w, r)
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "preview":
			s.handleLaidOutPagePreview(w, r, pageID)
			return
		case "export":
			s.handleLaidOutPageExport(w, r, pageID)
			return
		default:
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		record, err := s.repos.LaidOutPages.Get(pageID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp, convErr := laidOutPageToResponse(record)
		if convErr != nil {
			respondError(w, http.StatusInternalServerError, convErr.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"laid_out_page": resp})
	case http.MethodPatch, http.MethodPut:
		if s.pages == nil {
			respondError(w, http.StatusInternalServerError, "pages service not initialized")
			return
		}
		var req struct {
			LaidOutImageID *string `json:"laid_out_image_id"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.LaidOutImageID == nil || strings.TrimSpace(*req.LaidOutImageID) == "" {
			respondError(w, http.StatusBadRequest, "laid_out_image_id is required")
			return
		}
		if err := s.pages.UpdatePageImage(pageID, strings.TrimSpace(*req.LaidOutImageID)); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := s.repos.LaidOutPages.Get(pageID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp, convErr := laidOutPageToResponse(record)
		if convErr != nil {
			respondError(w, http.StatusInternalServerError, convErr.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"laid_out_page": resp})
	case http.MethodDelete:
		if err := s.repos.LaidOutPages.Delete(pageID); err != nil {
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

func (s *Server) handleLaidOutPagePreview(w http.ResponseWriter, r *http.Request, pageID string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.pages == nil {
		respondError(w, http.StatusInternalServerError, "pages service not initialized")
		return
	}
	_, err := s.pages.RenderPage(pageID)
	if err != nil {
		if errors.Is(err, services.ErrPageRendererNotImplemented) {
			respondError(w, http.StatusNotImplemented, err.Error())
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondError(w, http.StatusNotImplemented, "page preview renderer not implemented yet")
}

func (s *Server) handleLaidOutPageExport(w http.ResponseWriter, r *http.Request, pageID string) {
	respondError(w, http.StatusNotImplemented, "page export endpoint not implemented yet")
}
