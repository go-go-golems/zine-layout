package serve

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
    "strconv"
    "time"
    "context"
    "github.com/go-go-golems/zine-layout/pkg/export"
)

// Project zine routes --------------------------------------------------------

func (s *Server) handleProjectZines(w http.ResponseWriter, r *http.Request, projectID string) {
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
		zines, err := s.repos.Zines.ListByProject(projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]zineResponse, 0, len(zines))
		for _, z := range zines {
			resp = append(resp, zineToResponse(z))
		}
		writeJSON(w, http.StatusOK, map[string]any{"zines": resp})
	case http.MethodPost:
		if s.zines == nil {
			respondError(w, http.StatusInternalServerError, "zines service not initialized")
			return
		}
		var req struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			PageIDs     []string `json:"laid_out_page_ids"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		name := strings.TrimSpace(req.Name)
		desc := strings.TrimSpace(req.Description)
		zine, pages, err := s.zines.CreateZine(projectID, name, desc, req.PageIDs)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"zine":  zineToResponse(zine),
			"pages": zinePagesToResponse(pages),
		})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Zine resource --------------------------------------------------------------

func (s *Server) handleZineRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/zines/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	zineID := parts[0]
	if zineID == "" {
		http.NotFound(w, r)
		return
	}

    if len(parts) > 1 {
		switch parts[1] {
		case "pages":
			s.handleZinePages(w, r, zineID)
			return
        case "export":
            s.handleZineExport(w, r, zineID)
            return
		default:
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		if s.zines == nil {
			respondError(w, http.StatusInternalServerError, "zines service not initialized")
			return
		}
		zine, pages, err := s.zines.GetZineWithPages(zineID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"zine":  zineToResponse(zine),
			"pages": zinePagesToResponse(pages),
		})
	case http.MethodPatch, http.MethodPut:
		if s.zines == nil {
			respondError(w, http.StatusInternalServerError, "zines service not initialized")
			return
		}
		zine, err := s.repos.Zines.Get(zineID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Name != nil {
			zine.Name = strings.TrimSpace(*req.Name)
			if zine.Name == "" {
				zine.Name = "Untitled Zine"
			}
		}
		if req.Description != nil {
			zine.Description = strings.TrimSpace(*req.Description)
		}
		if err := s.zines.UpdateZine(zine); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"zine": zineToResponse(zine)})
	case http.MethodDelete:
		if s.zines == nil {
			respondError(w, http.StatusInternalServerError, "zines service not initialized")
			return
		}
		if err := s.zines.DeleteZine(zineID); err != nil {
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

func (s *Server) handleZinePages(w http.ResponseWriter, r *http.Request, zineID string) {
	if s.zines == nil {
		respondError(w, http.StatusInternalServerError, "zines service not initialized")
		return
	}
	switch r.Method {
	case http.MethodGet:
		pages, err := s.repos.Zines.GetPages(zineID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"pages": zinePagesToResponse(pages)})
	case http.MethodPut, http.MethodPatch:
		var req struct {
			PageIDs []string `json:"laid_out_page_ids"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.zines.UpdateZinePages(zineID, req.PageIDs); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		pages, err := s.repos.Zines.GetPages(zineID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"pages": zinePagesToResponse(pages)})
	default:
		w.Header().Set("Allow", "GET, PUT, PATCH")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleZineExport(w http.ResponseWriter, r *http.Request, zineID string) {
    if r.Method != http.MethodGet {
        w.Header().Set("Allow", "GET")
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    if s.impose == nil {
        respondError(w, http.StatusInternalServerError, "imposition service not initialized")
        return
    }
    preset := strings.TrimSpace(r.URL.Query().Get("preset"))
    if preset == "" { preset = "10_8_sheet_zine" }
    dpi := 300.0
    if sdpi := strings.TrimSpace(r.URL.Query().Get("dpi")); sdpi != "" {
        if v, err := strconv.ParseFloat(sdpi, 64); err == nil && v > 0 { dpi = v }
    }

    // Impose sheets
    sheets, err := s.impose.ImposeZine(zineID, preset)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            http.NotFound(w, r)
            return
        }
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Stream PDF directly to response
    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", "attachment; filename=\""+zineID+"-"+preset+".pdf\"")
    w.Header().Set("Cache-Control", "no-store")
    w.Header().Set("X-Content-Type-Options", "nosniff")

    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
    defer cancel()

    if err := export.SheetsToPDF(ctx, sheets, dpi, w); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
}
