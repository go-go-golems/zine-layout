package serve

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/projects"
)

// Asset routes ----------------------------------------------------------------

func (s *Server) handleAssetRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/assets/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(path, "/")
	assetID := parts[0]
	if assetID == "" {
		http.NotFound(w, r)
		return
	}
	if len(parts) > 1 {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		asset, err := s.repos.Assets.Get(assetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"asset": assetToResponse(asset, s.projectsRoot)})
	case http.MethodDelete:
		asset, err := s.repos.Assets.Get(assetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := s.repos.Assets.Delete(assetID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		imagePath := filepath.Join(projects.ProjectImagesDir(s.projectsRoot, asset.ProjectID), asset.Filename)
		_ = os.Remove(imagePath)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
