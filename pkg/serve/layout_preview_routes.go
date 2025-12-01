package serve

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
)

// handleProjectLayoutPreview computes layout geometry without persisting records.
func (s *Server) handleProjectLayoutPreview(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.layout == nil {
		respondError(w, http.StatusInternalServerError, "layout service is not initialized")
		return
	}

	var req struct {
		Layout  imagelayout.LayoutRequest `json:"layout"`
		AssetID string                    `json:"asset_id"`
		Image   *struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"image,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var meta imagelayout.ImageMeta
	if assetID := strings.TrimSpace(req.AssetID); assetID != "" {
		asset, err := s.repos.Assets.Get(assetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if asset.ProjectID != projectID {
			respondError(w, http.StatusBadRequest, "asset does not belong to project")
			return
		}
		meta = imagelayout.ImageMeta{Width: asset.Width, Height: asset.Height}
	} else if req.Image != nil {
		meta = imagelayout.ImageMeta{Width: req.Image.Width, Height: req.Image.Height}
	}

	if meta.Width <= 0 || meta.Height <= 0 {
		respondError(w, http.StatusBadRequest, "image dimensions are required")
		return
	}

	result, err := s.layout.PreviewLayout(req.Layout, meta)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"result": result})
}
