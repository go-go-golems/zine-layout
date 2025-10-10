package serve

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/projects"
	"github.com/go-go-golems/zine-layout/pkg/repo"
	sqliterepo "github.com/go-go-golems/zine-layout/pkg/repo/sqlite"
)

// Settings controls how the HTTP server is configured.
type Settings struct {
	Root     string
	DataRoot string
	Addr     string
}

// Server exposes the HTTP API for the zine layout platform.
type Server struct {
	settings     Settings
	httpServer   *http.Server
	projectsRoot string
	uploadsRoot  string
	db           *sql.DB
	repos        *repo.Repositories
}

// New constructs a Server with the provided settings.
func New(settings Settings) *Server {
	return &Server{settings: settings}
}

// ListenAndServe boots the HTTP server and blocks until the context is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	if err := s.prepare(); err != nil {
		return err
	}
	defer s.close()

	handler := s.Routes()
	s.httpServer = &http.Server{
		Addr:    s.settings.Addr,
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("serving API on %s (projects in %s)", s.settings.Addr, s.projectsRoot)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (s *Server) prepare() error {
	s.projectsRoot = filepath.Join(s.settings.DataRoot, "projects")
	s.uploadsRoot = filepath.Join(s.settings.DataRoot, "uploads")

	if err := os.MkdirAll(s.projectsRoot, 0o755); err != nil {
		return fmt.Errorf("create projects root: %w", err)
	}
	if err := os.MkdirAll(s.uploadsRoot, 0o755); err != nil {
		return fmt.Errorf("create uploads root: %w", err)
	}
	return s.initDatabase()
}

func (s *Server) databasePath() string {
	return filepath.Join(s.settings.DataRoot, "zine-layout.db")
}

func (s *Server) initDatabase() error {
	if s.repos != nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.databasePath()), 0o755); err != nil {
		return fmt.Errorf("prepare database dir: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL", s.databasePath())
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	repos, err := sqliterepo.NewRepositories(db)
	if err != nil {
		_ = db.Close()
		return err
	}
	s.db = db
	s.repos = repos
	return nil
}

// Routes configures all HTTP endpoints exposed by the server.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/projects", s.handleProjects)
	mux.HandleFunc("/api/projects/", s.handleProjectRoutes)

	mux.HandleFunc("/api/assets/", s.handleAssetRoutes)

	mux.HandleFunc("/api/image-sequences", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/image-sequences/", s.handleSequenceRoutes)

	// Serve project files (images) under /projects/{id}/...
	mux.Handle("/projects/", http.StripPrefix("/projects/", http.FileServer(http.Dir(s.projectsRoot))))

	if s.settings.Root != "" {
		staticFS := http.FileServer(http.Dir(s.settings.Root))
		mux.Handle("/", staticFS)
	}

	return mux
}

func (s *Server) close() {
	if s.db != nil {
		_ = s.db.Close()
		s.db = nil
	}
	s.repos = nil
}

// HTTP handlers ----------------------------------------------------------------

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

func (s *Server) handleProjectSequences(w http.ResponseWriter, r *http.Request, projectID string) {
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
		sequences, err := s.repos.ImageSequences.ListByProject(projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]imageSequenceResponse, 0, len(sequences))
		for _, seq := range sequences {
			resp = append(resp, sequenceToResponse(seq))
		}
		writeJSON(w, http.StatusOK, map[string]any{"sequences": resp})
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
			name = "Untitled Sequence"
		}
		sequence := &repo.ImageSequence{
			ProjectID:   projectID,
			Name:        name,
			Description: strings.TrimSpace(req.Description),
		}
		if err := s.repos.ImageSequences.Create(sequence); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"sequence": sequenceToResponse(sequence)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

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

func (s *Server) handleSequenceRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/image-sequences/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(path, "/")
	sequenceID := parts[0]
	if sequenceID == "" {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 {
		s.handleSequenceItem(w, r, sequenceID)
		return
	}
	if parts[1] != "items" {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 2 {
		s.handleSequenceItems(w, r, sequenceID)
		return
	}
	position, err := strconv.Atoi(parts[2])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item position")
		return
	}
	s.handleSequenceItemDeletion(w, r, sequenceID, position)
}

func (s *Server) handleSequenceItem(w http.ResponseWriter, r *http.Request, sequenceID string) {
	switch r.Method {
	case http.MethodGet:
		sequence, err := s.repos.ImageSequences.Get(sequenceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.ImageSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"sequence": sequenceToResponse(sequence),
			"items":    sequenceItemsToResponse(items),
		})
	case http.MethodPatch, http.MethodPut:
		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		sequence, err := s.repos.ImageSequences.Get(sequenceID)
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
				name = "Untitled Sequence"
			}
			sequence.Name = name
		}
		if req.Description != nil {
			sequence.Description = strings.TrimSpace(*req.Description)
		}
		sequence.UpdatedAt = time.Now().UTC()
		if err := s.repos.ImageSequences.Update(sequence); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sequence": sequenceToResponse(sequence)})
	case http.MethodDelete:
		if err := s.repos.ImageSequences.Delete(sequenceID); err != nil {
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

func (s *Server) handleSequenceItems(w http.ResponseWriter, r *http.Request, sequenceID string) {
	sequence, err := s.repos.ImageSequences.Get(sequenceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := s.repos.ImageSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": sequenceItemsToResponse(items)})
	case http.MethodPost:
		var req struct {
			AssetID *string `json:"asset_id"`
			IsGap   bool    `json:"is_gap"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if !req.IsGap {
			if req.AssetID == nil || *req.AssetID == "" {
				respondError(w, http.StatusBadRequest, "asset_id required when is_gap is false")
				return
			}
			asset, err := s.repos.Assets.Get(*req.AssetID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					respondError(w, http.StatusBadRequest, "asset not found")
					return
				}
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if asset.ProjectID != sequence.ProjectID {
				respondError(w, http.StatusBadRequest, "asset does not belong to sequence project")
				return
			}
		}
		item := &repo.ImageSequenceItem{
			SequenceID: sequenceID,
			AssetID:    req.AssetID,
			IsGap:      req.IsGap,
		}
		if err := s.repos.ImageSequences.AddItem(item); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.ImageSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"items": sequenceItemsToResponse(items)})
	case http.MethodPut:
		var req struct {
			Items []struct {
				AssetID *string `json:"asset_id"`
				IsGap   bool    `json:"is_gap"`
			} `json:"items"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		replacements := make([]*repo.ImageSequenceItem, 0, len(req.Items))
		for idx, item := range req.Items {
			if !item.IsGap {
				if item.AssetID == nil || *item.AssetID == "" {
					respondError(w, http.StatusBadRequest, fmt.Sprintf("asset_id required for item %d", idx))
					return
				}
				asset, err := s.repos.Assets.Get(*item.AssetID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						respondError(w, http.StatusBadRequest, fmt.Sprintf("asset %s not found", *item.AssetID))
						return
					}
					respondError(w, http.StatusInternalServerError, err.Error())
					return
				}
				if asset.ProjectID != sequence.ProjectID {
					respondError(w, http.StatusBadRequest, "asset does not belong to sequence project")
					return
				}
			}
			replacements = append(replacements, &repo.ImageSequenceItem{
				SequenceID: sequenceID,
				AssetID:    item.AssetID,
				IsGap:      item.IsGap,
			})
		}
		if err := s.repos.ImageSequences.ReplaceItems(sequenceID, replacements); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.ImageSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": sequenceItemsToResponse(items)})
	default:
		w.Header().Set("Allow", "GET, POST, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSequenceItemDeletion(w http.ResponseWriter, r *http.Request, sequenceID string, position int) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.repos.ImageSequences.DeleteItem(sequenceID, position); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Helpers ---------------------------------------------------------------------

type projectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type assetResponse struct {
	ID          string         `json:"id"`
	ProjectID   string         `json:"project_id"`
	Filename    string         `json:"filename"`
	RelPath     string         `json:"rel_path"`
	ContentType string         `json:"content_type"`
	Bytes       int64          `json:"bytes"`
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	UploadedAt  time.Time      `json:"uploaded_at"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	URL         string         `json:"url"`
}

type imageSequenceResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type imageSequenceItemResponse struct {
	Position int     `json:"position"`
	AssetID  *string `json:"asset_id,omitempty"`
	IsGap    bool    `json:"is_gap"`
}

func projectToResponse(p *repo.Project) projectResponse {
	return projectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func assetToResponse(a *repo.Asset, projectsRoot string) assetResponse {
	metadata := map[string]any{}
	if strings.TrimSpace(a.MetadataJSON) != "" {
		_ = json.Unmarshal([]byte(a.MetadataJSON), &metadata)
	}
	url := fmt.Sprintf("/projects/%s/images/%s", a.ProjectID, a.Filename)
	if _, err := os.Stat(filepath.Join(projectsRoot, a.ProjectID, "images", a.Filename)); err != nil {
		url = ""
	}
	return assetResponse{
		ID:          a.ID,
		ProjectID:   a.ProjectID,
		Filename:    a.Filename,
		RelPath:     a.RelPath,
		ContentType: a.ContentType,
		Bytes:       a.Bytes,
		Width:       a.Width,
		Height:      a.Height,
		UploadedAt:  a.UploadedAt,
		Metadata:    metadata,
		URL:         url,
	}
}

func sequenceToResponse(seq *repo.ImageSequence) imageSequenceResponse {
	return imageSequenceResponse{
		ID:          seq.ID,
		ProjectID:   seq.ProjectID,
		Name:        seq.Name,
		Description: seq.Description,
		CreatedAt:   seq.CreatedAt,
		UpdatedAt:   seq.UpdatedAt,
	}
}

func sequenceItemsToResponse(items []*repo.ImageSequenceItem) []imageSequenceItemResponse {
	resp := make([]imageSequenceItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, imageSequenceItemResponse{
			Position: item.Position,
			AssetID:  item.AssetID,
			IsGap:    item.IsGap,
		})
	}
	return resp
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

func decodeJSON(r *http.Request, dest any) error {
	defer r.Body.Close()
	if r.Body == nil {
		return nil
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}
