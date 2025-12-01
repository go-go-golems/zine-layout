package serve

import (
	"database/sql"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"github.com/go-go-golems/zine-layout/pkg/projects"
	"golang.org/x/image/draw"
)

// handleProjectLayoutRender renders a layout + asset into an image (PNG) without persisting.
func (s *Server) handleProjectLayoutRender(w http.ResponseWriter, r *http.Request, projectID string) {
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
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.AssetID) == "" {
		respondError(w, http.StatusBadRequest, "asset_id is required for rendering")
		return
	}

	asset, err := s.repos.Assets.Get(req.AssetID)
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

	assetPath := filepath.Join(projects.ProjectImagesDir(s.projectsRoot, asset.ProjectID), asset.Filename)
	file, err := os.Open(assetPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "open asset: "+err.Error())
		return
	}
	defer file.Close()

	srcImg, _, err := image.Decode(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "decode asset: "+err.Error())
		return
	}

	meta := imagelayout.ImageMeta{Width: srcImg.Bounds().Dx(), Height: srcImg.Bounds().Dy()}
	computation, err := s.layout.PreviewLayout(req.Layout, meta)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	rendered, err := renderViewport(srcImg, computation.Result)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/png")
	// Optionally return computation metadata alongside image via header
	if compJSON, err := json.Marshal(computation.Result); err == nil {
		w.Header().Set("X-Viewport-Result", string(compJSON))
	}
	if err := png.Encode(w, rendered); err != nil {
		respondError(w, http.StatusInternalServerError, "encode render: "+err.Error())
		return
	}
}

// renderViewport draws the source image into a canvas according to the viewport result.
func renderViewport(src image.Image, result imagelayout.ViewportResult) (image.Image, error) {
	canvasW := int(math.Max(1, math.Round(result.CanvasRect.W)))
	canvasH := int(math.Max(1, math.Round(result.CanvasRect.H)))
	dst := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	// Fill background white
	for y := 0; y < canvasH; y++ {
		for x := 0; x < canvasW; x++ {
			dst.Set(x, y, color.White)
		}
	}

	srcRect := image.Rect(
		int(math.Round(result.SourceRect.X)),
		int(math.Round(result.SourceRect.Y)),
		int(math.Round(result.SourceRect.X+result.SourceRect.W)),
		int(math.Round(result.SourceRect.Y+result.SourceRect.H)),
	).Intersect(src.Bounds())

	if srcRect.Empty() {
		return dst, nil
	}

	dstRect := image.Rect(
		int(math.Round(result.TargetRect.X)),
		int(math.Round(result.TargetRect.Y)),
		int(math.Round(result.TargetRect.X+result.TargetRect.W)),
		int(math.Round(result.TargetRect.Y+result.TargetRect.H)),
	).Intersect(dst.Bounds())

	if dstRect.Empty() {
		return dst, nil
	}

	draw.NearestNeighbor.Scale(dst, dstRect, src, srcRect, draw.Over, nil)
	return dst, nil
}
