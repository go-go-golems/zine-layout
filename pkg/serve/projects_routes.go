package serve

import (
	"database/sql"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/projects"
	"github.com/go-go-golems/zine-layout/pkg/repo"
)

// Project collection routes ---------------------------------------------------

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		projectsList, err := s.repos.Projects.List()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]projectResponse, 0, len(projectsList))
		for _, p := range projectsList {
			resp = append(resp, projectToResponse(p))
		}
		writeJSON(w, http.StatusOK, map[string]any{"projects": resp})
	case http.MethodPost:
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = "Untitled Project"
		}
		project := &repo.Project{
			Name:        name,
			Description: strings.TrimSpace(req.Description),
		}
		if err := s.repos.Projects.Create(project); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := projects.EnsureProjectDirs(s.projectsRoot, project.ID); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"project": projectToResponse(project)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Project scoped routes -------------------------------------------------------

func (s *Server) handleProjectRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(path, "/")
	projectID := parts[0]
	if projectID == "" {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 {
		s.handleProjectItem(w, r, projectID)
		return
	}

	switch parts[1] {
	case "assets":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectAssets(w, r, projectID)
	case "images":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectImageUpload(w, r, projectID)
	case "image-sequences":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectSequences(w, r, projectID)
	case "image-layout-templates":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectLayoutTemplates(w, r, projectID)
	case "laid-out-images":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectLaidOutImages(w, r, projectID)
	case "layout-sequences":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectLayoutSequences(w, r, projectID)
	case "page-templates":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectPageTemplates(w, r, projectID)
	case "laid-out-pages":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectLaidOutPages(w, r, projectID)
	case "zines":
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		s.handleProjectZines(w, r, projectID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleProjectItem(w http.ResponseWriter, r *http.Request, projectID string) {
	switch r.Method {
	case http.MethodGet:
		project, err := s.repos.Projects.Get(projectID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"project": projectToResponse(project)})
	case http.MethodPatch, http.MethodPut:
		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		project, err := s.repos.Projects.Get(projectID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				name = "Untitled Project"
			}
			project.Name = name
		}
		if req.Description != nil {
			project.Description = strings.TrimSpace(*req.Description)
		}
		project.UpdatedAt = time.Now().UTC()
		if err := s.repos.Projects.Update(project); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"project": projectToResponse(project)})
	case http.MethodDelete:
		if err := s.repos.Projects.Delete(projectID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		_ = os.RemoveAll(projects.ProjectDir(s.projectsRoot, projectID))
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PATCH, PUT, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Project assets --------------------------------------------------------------

func (s *Server) handleProjectAssets(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, err := s.repos.Projects.Get(projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	assets, err := s.repos.Assets.ListByProject(projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]assetResponse, 0, len(assets))
	for _, asset := range assets {
		resp = append(resp, assetToResponse(asset, s.projectsRoot))
	}
	writeJSON(w, http.StatusOK, map[string]any{"assets": resp})
}

func (s *Server) handleProjectImageUpload(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, err := s.repos.Projects.Get(projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	files := r.MultipartForm.File["images[]"]
	if len(files) == 0 {
		files = r.MultipartForm.File["file"]
	}
	if len(files) == 0 {
		respondError(w, http.StatusBadRequest, "no files uploaded")
		return
	}
	contentType := mime.TypeByExtension(".png")
	if contentType == "" {
		contentType = "image/png"
	}
	now := time.Now().UTC()
	created := make([]assetResponse, 0, len(files))
	for _, fh := range files {
		saved, err := projects.SavePNGImage(s.projectsRoot, projectID, fh)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		asset := &repo.Asset{
			ProjectID:   projectID,
			Filename:    saved.Filename,
			RelPath:     filepath.ToSlash(filepath.Join("projects", projectID, "images", saved.Filename)),
			ContentType: contentType,
			Bytes:       saved.Bytes,
			Width:       saved.Width,
			Height:      saved.Height,
			UploadedAt:  now,
		}
		meta := map[string]any{
			"width":  saved.Width,
			"height": saved.Height,
		}
		if data, err := json.Marshal(meta); err == nil {
			asset.MetadataJSON = string(data)
		}
		if err := s.repos.Assets.Create(asset); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		created = append(created, assetToResponse(asset, s.projectsRoot))
	}
	writeJSON(w, http.StatusCreated, map[string]any{"assets": created})
}
