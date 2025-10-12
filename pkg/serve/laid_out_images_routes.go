package serve

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Project laid-out images -----------------------------------------------------

func (s *Server) handleProjectLaidOutImages(w http.ResponseWriter, r *http.Request, projectID string) {
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
		items, err := s.repos.LaidOutImages.ListByProject(projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]*laidOutImageResponse, 0, len(items))
		for _, record := range items {
			out, convErr := laidOutImageToResponse(record)
			if convErr != nil {
				respondError(w, http.StatusInternalServerError, convErr.Error())
				return
			}
			resp = append(resp, out)
		}
		writeJSON(w, http.StatusOK, map[string]any{"laid_out_images": resp})
	case http.MethodPost:
		if s.layout == nil {
			respondError(w, http.StatusInternalServerError, "layout service is not initialized")
			return
		}
		var req struct {
			AssetID    string         `json:"asset_id"`
			TemplateID string         `json:"template_id"`
			Overrides  map[string]any `json:"overrides"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.TrimSpace(req.AssetID) == "" {
			respondError(w, http.StatusBadRequest, "asset_id is required")
			return
		}
		if strings.TrimSpace(req.TemplateID) == "" {
			respondError(w, http.StatusBadRequest, "template_id is required")
			return
		}
		var overridesJSON *string
		if req.Overrides != nil {
			bytes, err := json.Marshal(req.Overrides)
			if err != nil {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid overrides: %v", err))
				return
			}
			text := strings.TrimSpace(string(bytes))
			if text != "" && text != "{}" {
				overridesJSON = &text
			}
		}
		record, err := s.layout.CreateLaidOutImage(projectID, req.AssetID, req.TemplateID, overridesJSON)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, convErr := laidOutImageToResponse(record)
		if convErr != nil {
			respondError(w, http.StatusInternalServerError, convErr.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"laid_out_image": resp})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Laid-out image resource -----------------------------------------------------

func (s *Server) handleLaidOutImageRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/laid-out-images/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	imageID := parts[0]
	if imageID == "" {
		http.NotFound(w, r)
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "preview":
			s.handleLaidOutImagePreview(w, r, imageID)
			return
		case "export":
			s.handleLaidOutImageExport(w, r, imageID)
			return
		default:
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		record, err := s.repos.LaidOutImages.Get(imageID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp, convErr := laidOutImageToResponse(record)
		if convErr != nil {
			respondError(w, http.StatusInternalServerError, convErr.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"laid_out_image": resp})
	case http.MethodPatch, http.MethodPut:
		if s.layout == nil {
			respondError(w, http.StatusInternalServerError, "layout service is not initialized")
			return
		}
		record, err := s.repos.LaidOutImages.Get(imageID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var req struct {
			TemplateID *string         `json:"template_id"`
			Overrides  *map[string]any `json:"overrides"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.TemplateID != nil {
			template := strings.TrimSpace(*req.TemplateID)
			if template == "" {
				respondError(w, http.StatusBadRequest, "template_id cannot be empty")
				return
			}
			record.TemplateID = template
		}
		if req.Overrides != nil {
			if len(*req.Overrides) == 0 {
				record.OverridesJSON = nil
			} else {
				bytes, err := json.Marshal(req.Overrides)
				if err != nil {
					respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid overrides: %v", err))
					return
				}
				text := strings.TrimSpace(string(bytes))
				record.OverridesJSON = &text
			}
		}
		if err := s.layout.RecomputeLaidOutImage(record); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, convErr := laidOutImageToResponse(record)
		if convErr != nil {
			respondError(w, http.StatusInternalServerError, convErr.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"laid_out_image": resp})
	case http.MethodDelete:
		if err := s.repos.LaidOutImages.Delete(imageID); err != nil {
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

func (s *Server) handleLaidOutImagePreview(w http.ResponseWriter, r *http.Request, imageID string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	record, err := s.repos.LaidOutImages.Get(imageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp, convErr := laidOutImageToResponse(record)
	if convErr != nil {
		respondError(w, http.StatusInternalServerError, convErr.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": resp.Result})
}

func (s *Server) handleLaidOutImageExport(w http.ResponseWriter, r *http.Request, imageID string) {
	respondError(w, http.StatusNotImplemented, "export endpoint not implemented yet")
}
