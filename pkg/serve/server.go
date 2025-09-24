package serve

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/presets"
	"github.com/go-go-golems/zine-layout/pkg/projects"
	"github.com/go-go-golems/zine-layout/pkg/render"
	"github.com/go-go-golems/zine-layout/pkg/repo"
	sqliterepo "github.com/go-go-golems/zine-layout/pkg/repo/sqlite"
	"github.com/go-go-golems/zine-layout/pkg/spread"
	simple "github.com/go-go-golems/zine-layout/pkg/spread/simple"
	"github.com/go-go-golems/zine-layout/pkg/validation"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type Settings struct {
	Root     string
	DataRoot string
	Addr     string
}

type Server struct {
	settings     Settings
	httpServer   *http.Server
	projectsRoot string
	presetsRoot  string
	uploadsRoot  string
	db           *sql.DB
	repos        *repo.Repositories
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	copy := trimmed
	return &copy
}

func (s *Server) upsertProjectRecord(p *projects.Project) {
	if s.repos == nil || p == nil {
		return
	}
	rp := &repo.Project{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if val := optionalString(p.PresetID); val != nil {
		rp.PresetID = val
	}
	existing, err := s.repos.Projects.Get(p.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if rp.CreatedAt.IsZero() {
				rp.CreatedAt = time.Now().UTC()
			}
			if rp.UpdatedAt.IsZero() {
				rp.UpdatedAt = rp.CreatedAt
			}
			if err := s.repos.Projects.Create(rp); err != nil {
				log.Printf("warn: failed to create project record: %v", err)
			}
			return
		}
		log.Printf("warn: failed to read project record: %v", err)
		return
	}
	if rp.CreatedAt.IsZero() {
		rp.CreatedAt = existing.CreatedAt
	}
	if rp.UpdatedAt.IsZero() {
		rp.UpdatedAt = existing.UpdatedAt
	}
	if rp.PresetID == nil {
		rp.PresetID = existing.PresetID
	}
	rp.CoverAssetID = existing.CoverAssetID
	if err := s.repos.Projects.Update(rp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := s.repos.Projects.Create(rp); err != nil {
				log.Printf("warn: failed to recreate project record: %v", err)
			}
			return
		}
		log.Printf("warn: failed to update project record: %v", err)
	}
}

func (s *Server) ensureProjectAssets(projectID string) {
	if s.repos == nil {
		return
	}
	assets, err := s.repos.Assets.ListByProject(projectID)
	if err != nil {
		log.Printf("warn: list assets for seeding failed: %v", err)
		return
	}
	if len(assets) > 0 {
		return
	}
	imgs, order, err := projects.ListProjectImages(s.projectsRoot, projectID)
	if err != nil {
		log.Printf("warn: seed assets failed: %v", err)
		return
	}
	orderIndex := make(map[string]int)
	for idx, id := range order {
		orderIndex[id] = idx
	}
	for _, img := range imgs {
		statPath := filepath.Join(projects.ProjectImagesDir(s.projectsRoot, projectID), img.ID)
		info, err := os.Stat(statPath)
		if err != nil {
			continue
		}
		created := info.ModTime().UTC()
		asset := &repo.Asset{
			ProjectID:   projectID,
			ID:          img.ID,
			Filename:    img.Name,
			RelPath:     filepath.ToSlash(filepath.Join("projects", projectID, "images", img.ID)),
			ContentType: "image/png",
			Bytes:       info.Size(),
			Width:       img.Width,
			Height:      img.Height,
			SortIndex:   orderIndex[img.ID],
			CreatedAt:   created,
		}
		if err := s.repos.Assets.Create(asset); err != nil {
			log.Printf("warn: failed to seed asset %s: %v", img.ID, err)
		}
	}
}

func (s *Server) fetchAssets(projectID string) ([]*repo.Asset, error) {
	if s.repos == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	assets, err := s.repos.Assets.ListByProject(projectID)
	if err == nil && len(assets) > 0 {
		return assets, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("warn: list assets failed: %v", err)
	}
	s.ensureProjectAssets(projectID)
	return s.repos.Assets.ListByProject(projectID)
}

func (s *Server) respondWithAssets(w http.ResponseWriter, projectID string) {
	assets, err := s.fetchAssets(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	images := make([]map[string]any, 0, len(assets))
	order := make([]string, 0, len(assets))
	for _, asset := range assets {
		images = append(images, map[string]any{
			"id":     asset.ID,
			"name":   asset.Filename,
			"width":  asset.Width,
			"height": asset.Height,
		})
		order = append(order, asset.ID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"images": images, "order": order})
}

func (s *Server) handleUploadImages(w http.ResponseWriter, r *http.Request, projectID string) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, "multipart parse error", http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["images[]"]
	if len(files) == 0 {
		if fh := r.MultipartForm.File["file"]; len(fh) > 0 {
			files = fh
		}
	}
	if len(files) == 0 {
		http.Error(w, "no files provided", http.StatusBadRequest)
		return
	}
	s.ensureProjectAssets(projectID)
	var baseIndex int
	if s.repos != nil {
		if current, err := s.repos.Assets.ListByProject(projectID); err == nil {
			baseIndex = len(current)
		}
	}
	newImages := make([]map[string]any, 0, len(files))
	for idx, fh := range files {
		if fh.Size == 0 {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(fh.Filename), ".png") {
			http.Error(w, "only PNG files supported", http.StatusBadRequest)
			return
		}
		item, err := projects.SavePNGImage(s.projectsRoot, projectID, fh)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if s.repos != nil {
			asset := &repo.Asset{
				ProjectID:   projectID,
				ID:          item.ID,
				Filename:    item.Name,
				RelPath:     filepath.ToSlash(filepath.Join("projects", projectID, "images", item.ID)),
				ContentType: "image/png",
				Bytes:       fh.Size,
				Width:       item.Width,
				Height:      item.Height,
				SortIndex:   baseIndex + idx,
				CreatedAt:   time.Now().UTC(),
			}
			if err := s.repos.Assets.Create(asset); err != nil {
				log.Printf("warn: failed to create asset record: %v", err)
			}
		}
		newImages = append(newImages, map[string]any{
			"id":     item.ID,
			"name":   item.Name,
			"width":  item.Width,
			"height": item.Height,
		})
	}
	s.bumpProjectUpdatedAt(projectID)
	writeJSON(w, http.StatusCreated, map[string]any{"images": newImages})
}

func (s *Server) handleReorderImages(w http.ResponseWriter, r *http.Request, projectID string) {
	var req struct {
		Order []string `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(req.Order) == 0 {
		http.Error(w, "order required", http.StatusBadRequest)
		return
	}
	if s.repos != nil {
		if err := s.repos.Assets.UpdateOrder(projectID, req.Order); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := projects.SetProjectOrder(s.projectsRoot, projectID, req.Order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.bumpProjectUpdatedAt(projectID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteImage(w http.ResponseWriter, projectID, imageID string) {
	if err := projects.DeleteProjectImage(s.projectsRoot, projectID, imageID); err != nil {
		status := http.StatusInternalServerError
		if os.IsNotExist(err) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	if s.repos != nil {
		if err := s.repos.Assets.Delete(projectID, imageID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("warn: failed to delete asset record: %v", err)
		}
	}
	s.bumpProjectUpdatedAt(projectID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) serveProjectImage(w http.ResponseWriter, r *http.Request, projectID, imageID string) {
	fn := filepath.Join(projects.ProjectImagesDir(s.projectsRoot, projectID), filepath.Base(imageID))
	if _, err := os.Stat(fn); err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, fn)
}

func (s *Server) bumpProjectUpdatedAt(projectID string) {
	if p, err := projects.ReadProject(s.projectsRoot, projectID); err == nil {
		p.UpdatedAt = time.Now().UTC()
		if err := projects.WriteProject(s.projectsRoot, p); err != nil {
			log.Printf("warn: failed to write project for updated_at: %v", err)
		}
		s.upsertProjectRecord(p)
	}
}

type pageResponse struct {
	PageNumber int             `json:"page_number"`
	AssetID    *string         `json:"asset_id,omitempty"`
	Settings   spread.Settings `json:"settings"`
	Result     *simple.Result  `json:"result,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type spreadResponse struct {
	SpreadNumber    int             `json:"spread_number"`
	LeftPageNumber  *int            `json:"left_page_number,omitempty"`
	RightPageNumber *int            `json:"right_page_number,omitempty"`
	Settings        spread.Settings `json:"settings"`
	Result          *simple.Result  `json:"result,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func encodeSettingsJSON(settings spread.Settings) (string, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeSettingsJSON(raw string) (spread.Settings, error) {
	var settings spread.Settings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return spread.Settings{}, err
	}
	return settings, nil
}

func encodeResultJSON(result *simple.Result) (*string, error) {
	if result == nil {
		return nil, nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	str := string(data)
	return &str, nil
}

func decodeResultJSON(raw *string) (*simple.Result, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	var result simple.Result
	if err := json.Unmarshal([]byte(*raw), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func pageRecordToResponse(rec *repo.Page) (*pageResponse, error) {
	settings, err := decodeSettingsJSON(rec.SettingsJSON)
	if err != nil {
		return nil, err
	}
	result, err := decodeResultJSON(rec.ResultJSON)
	if err != nil {
		return nil, err
	}
	return &pageResponse{
		PageNumber: rec.PageNumber,
		AssetID:    rec.AssetID,
		Settings:   settings,
		Result:     result,
		CreatedAt:  rec.CreatedAt,
		UpdatedAt:  rec.UpdatedAt,
	}, nil
}

func spreadRecordToResponse(rec *repo.Spread) (*spreadResponse, error) {
	settings, err := decodeSettingsJSON(rec.SettingsJSON)
	if err != nil {
		return nil, err
	}
	result, err := decodeResultJSON(rec.ResultJSON)
	if err != nil {
		return nil, err
	}
	return &spreadResponse{
		SpreadNumber:    rec.SpreadNumber,
		LeftPageNumber:  rec.LeftPageNumber,
		RightPageNumber: rec.RightPageNumber,
		Settings:        settings,
		Result:          result,
		CreatedAt:       rec.CreatedAt,
		UpdatedAt:       rec.UpdatedAt,
	}, nil
}

func (s *Server) handleListPages(w http.ResponseWriter, projectID string) {
	if s.repos == nil {
		http.Error(w, "repository not initialized", http.StatusInternalServerError)
		return
	}
	records, err := s.repos.Pages.List(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pages := make([]*pageResponse, 0, len(records))
	for _, rec := range records {
		resp, err := pageRecordToResponse(rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pages = append(pages, resp)
	}
	writeJSON(w, http.StatusOK, map[string]any{"pages": pages})
}

func (s *Server) handleUpsertPage(w http.ResponseWriter, r *http.Request, projectID string, pageNumber int) {
	if s.repos == nil {
		http.Error(w, "repository not initialized", http.StatusInternalServerError)
		return
	}
	var req struct {
		AssetID  *string          `json:"asset_id"`
		Settings *spread.Settings `json:"settings"`
		Result   *simple.Result   `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Settings == nil {
		http.Error(w, "settings required", http.StatusBadRequest)
		return
	}
	settingsJSON, err := encodeSettingsJSON(*req.Settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resultJSON, err := encodeResultJSON(req.Result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	page := &repo.Page{
		ProjectID:    projectID,
		PageNumber:   pageNumber,
		AssetID:      req.AssetID,
		SettingsJSON: settingsJSON,
		ResultJSON:   resultJSON,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if existing, err := s.repos.Pages.GetByNumber(projectID, pageNumber); err == nil {
		page.CreatedAt = existing.CreatedAt
		if req.AssetID == nil {
			page.AssetID = existing.AssetID
		}
	}
	if err := s.repos.Pages.Upsert(page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := s.repos.Pages.GetByNumber(projectID, pageNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := pageRecordToResponse(updated)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bumpProjectUpdatedAt(projectID)
	writeJSON(w, http.StatusOK, map[string]any{"page": resp})
}

func (s *Server) handleDeletePage(w http.ResponseWriter, projectID string, pageNumber int) {
	if s.repos == nil {
		http.Error(w, "repository not initialized", http.StatusInternalServerError)
		return
	}
	if err := s.repos.Pages.Delete(projectID, pageNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bumpProjectUpdatedAt(projectID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleListSpreads(w http.ResponseWriter, projectID string) {
	if s.repos == nil {
		http.Error(w, "repository not initialized", http.StatusInternalServerError)
		return
	}
	records, err := s.repos.Spreads.List(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	spreads := make([]*spreadResponse, 0, len(records))
	for _, rec := range records {
		resp, err := spreadRecordToResponse(rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		spreads = append(spreads, resp)
	}
	writeJSON(w, http.StatusOK, map[string]any{"spreads": spreads})
}

func (s *Server) handleUpsertSpread(w http.ResponseWriter, r *http.Request, projectID string, spreadNumber int) {
	if s.repos == nil {
		http.Error(w, "repository not initialized", http.StatusInternalServerError)
		return
	}
	var req struct {
		LeftPageNumber  *int             `json:"left_page_number"`
		RightPageNumber *int             `json:"right_page_number"`
		Settings        *spread.Settings `json:"settings"`
		Result          *simple.Result   `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Settings == nil {
		http.Error(w, "settings required", http.StatusBadRequest)
		return
	}
	settingsJSON, err := encodeSettingsJSON(*req.Settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resultJSON, err := encodeResultJSON(req.Result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	spreadRecord := &repo.Spread{
		ProjectID:       projectID,
		SpreadNumber:    spreadNumber,
		LeftPageNumber:  req.LeftPageNumber,
		RightPageNumber: req.RightPageNumber,
		SettingsJSON:    settingsJSON,
		ResultJSON:      resultJSON,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if existing, err := s.repos.Spreads.GetByNumber(projectID, spreadNumber); err == nil {
		spreadRecord.CreatedAt = existing.CreatedAt
	}
	if err := s.repos.Spreads.Upsert(spreadRecord); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := s.repos.Spreads.GetByNumber(projectID, spreadNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := spreadRecordToResponse(updated)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bumpProjectUpdatedAt(projectID)
	writeJSON(w, http.StatusOK, map[string]any{"spread": resp})
}

func (s *Server) handleDeleteSpread(w http.ResponseWriter, projectID string, spreadNumber int) {
	if s.repos == nil {
		http.Error(w, "repository not initialized", http.StatusInternalServerError)
		return
	}
	if err := s.repos.Spreads.Delete(projectID, spreadNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bumpProjectUpdatedAt(projectID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type panelImage struct {
	Panel    string `json:"panel"`
	MimeType string `json:"mime_type"`
	DataURL  string `json:"data_url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type yamlRenderSpread struct {
	Name      string        `json:"name"`
	ImagePath string        `json:"image_path"`
	Result    simple.Result `json:"result"`
	Trace     []string      `json:"trace,omitempty"`
    TraceJSON any           `json:"trace_json,omitempty"`
	Panels    []panelImage  `json:"panels"`
}

type yamlRenderResponse struct {
	Spreads []yamlRenderSpread `json:"spreads"`
    HTML    string             `json:"html,omitempty"`
}

func makePanelPaths(tempDir, prefix, format string, isSpread bool) map[string]string {
	paths := map[string]string{}
	ext := normalizeExtension(format)
	if !isSpread {
		paths["single"] = filepath.Join(tempDir, fmt.Sprintf("%s-single%s", prefix, ext))
		return paths
	}
	paths["left"] = filepath.Join(tempDir, fmt.Sprintf("%s-left%s", prefix, ext))
	paths["right"] = filepath.Join(tempDir, fmt.Sprintf("%s-right%s", prefix, ext))
	return paths
}

func New(settings Settings) *Server {
	return &Server{settings: settings}
}

func (s *Server) prepare() error {
	s.projectsRoot = filepath.Join(s.settings.DataRoot, "projects")
	s.presetsRoot = filepath.Join(s.settings.DataRoot, "presets")
	s.uploadsRoot = filepath.Join(s.settings.DataRoot, "uploads")
	if err := os.MkdirAll(s.projectsRoot, 0o755); err != nil {
		return fmt.Errorf("create projects root: %w", err)
	}
	if err := os.MkdirAll(s.presetsRoot, 0o755); err != nil {
		return fmt.Errorf("create presets root: %w", err)
	}
	if err := os.MkdirAll(s.uploadsRoot, 0o755); err != nil {
		return fmt.Errorf("create uploads root: %w", err)
	}
	if err := s.initDatabase(); err != nil {
		return err
	}
	if err := presets.SeedPresetsIfEmpty(s.presetsRoot); err != nil {
		log.Printf("warning: failed to seed presets: %v", err)
	}
	return nil
}

func (s *Server) databasePath() string {
	return filepath.Join(s.settings.DataRoot, "zine-layout.db")
}

func (s *Server) initDatabase() error {
	if s.repos != nil && s.db != nil {
		return nil
	}
	dbPath := s.databasePath()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("prepare db dir: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL", dbPath)
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

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	// Presets
	mux.HandleFunc("/api/presets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list, err := presets.ListPresets(s.presetsRoot)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"presets": list})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/presets/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/presets/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		fn := filepath.Join(s.presetsRoot, filepath.Base(id)+".yaml")
		if _, err := os.Stat(fn); err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/yaml")
		http.ServeFile(w, r, fn)
	})

	// serve uploaded files statically
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.uploadsRoot))))

	// Projects collection
	mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			prjs, err := projects.ListProjects(s.projectsRoot)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for i := range prjs {
				s.upsertProjectRecord(&prjs[i])
			}
			writeJSON(w, http.StatusOK, map[string]any{"projects": prjs})
		case http.MethodPost:
			var req struct{ Name, PresetID string }
			_ = json.NewDecoder(r.Body).Decode(&req)
			p, err := projects.CreateProject(s.projectsRoot, req.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if req.PresetID != "" {
				if err := presets.ApplyPresetToProject(s.presetsRoot, projects.ProjectDir(s.projectsRoot, p.ID), req.PresetID); err != nil {
					log.Printf("warning: failed to apply preset to new project: %v", err)
				} else {
					p.PresetID = req.PresetID
					_ = projects.WriteProject(s.projectsRoot, p)
				}
			}
			s.upsertProjectRecord(p)
			writeJSON(w, http.StatusCreated, map[string]any{"project": p})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Uploads
	mux.HandleFunc("/api/uploads", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if err := r.ParseMultipartForm(64 << 20); err != nil {
				http.Error(w, "multipart parse error", http.StatusBadRequest)
				return
			}
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "missing file", http.StatusBadRequest)
				return
			}
			defer file.Close()
			name := sanitizeFilename(header.Filename)
			if name == "" {
				name = fmt.Sprintf("upload-%d", time.Now().UnixNano())
			}
			dstName := uniqueName(s.uploadsRoot, name)
			dstPath := filepath.Join(s.uploadsRoot, dstName)
			if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			out, err := os.Create(dstPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			n, copyErr := io.Copy(out, file)
			cerr := out.Close()
			if copyErr != nil {
				http.Error(w, copyErr.Error(), http.StatusInternalServerError)
				return
			}
			if cerr != nil {
				http.Error(w, cerr.Error(), http.StatusInternalServerError)
				return
			}
			url := "/uploads/" + dstName
			writeJSON(w, http.StatusOK, map[string]any{"name": dstName, "url": url, "bytes": n})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Spread API (v1)
	mux.HandleFunc("/api/v1/compute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req spread.ComputeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		meta, err := resolveImageMeta(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		inputs, err := simple.InputsFromSettings(req.Settings, meta)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
    trace := &simple.Trace{UseZerolog: true}
		result := simple.ComputePlacement(inputs, trace)
		resp := map[string]any{"result": result, "trace": trace.Lines}
		if trace.Structured != nil {
			resp["trace_json"] = trace.Structured
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/api/v1/preview", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req spread.ComputeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.ImagePath) == "" {
			http.Error(w, "image_path required", http.StatusBadRequest)
			return
		}
		img, meta, err := loadImageFromPath(req.ImagePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("load image: %v", err), http.StatusBadRequest)
			return
		}
		inputs, err := simple.InputsFromSettings(req.Settings, meta)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		trace := &simple.Trace{UseZerolog: true}
		result := simple.ComputePlacement(inputs, trace)
		tempDir, err := os.MkdirTemp("", "preview-simple-")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.RemoveAll(tempDir)
		imageBase := strings.TrimSuffix(filepath.Base(req.ImagePath), filepath.Ext(req.ImagePath))
		opts := simple.RenderOptions{PNGLevel: "speed", Scaler: "fast", ParallelEncode: true}
		info := simple.RenderInfoFromExport(req.Settings.Export, 1, req.Name, imageBase, opts)
		info.OutputDir = tempDir
		prefix := sanitizeNameForFile(req.Name)
		overrides := makePanelPaths(tempDir, prefix, info.Format, req.Settings.IsSpread)
		info.PathOverrides = overrides
		var target string
		if !req.Settings.IsSpread {
			path, err := simple.RenderSingle(r.Context(), img, result, info)
			if err != nil {
				http.Error(w, fmt.Sprintf("render preview: %v", err), http.StatusInternalServerError)
				return
			}
			target = path
		} else {
			leftPath, _, err := simple.RenderSpread(r.Context(), img, result, info)
			if err != nil {
				http.Error(w, fmt.Sprintf("render preview: %v", err), http.StatusInternalServerError)
				return
			}
			target = leftPath
		}
		http.ServeFile(w, r, target)
	})

	mux.HandleFunc("/api/v1/render", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req spread.ComputeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.ImagePath) == "" {
			http.Error(w, "image_path required", http.StatusBadRequest)
			return
		}
		img, meta, err := loadImageFromPath(req.ImagePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("load image: %v", err), http.StatusBadRequest)
			return
		}
		inputs, err := simple.InputsFromSettings(req.Settings, meta)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		trace := &simple.Trace{UseZerolog: true}
		result := simple.ComputePlacement(inputs, trace)
		outDir := strings.TrimSpace(req.Settings.Export.OutDir)
		if outDir == "" {
			outDir = filepath.Join(os.TempDir(), "spread-render")
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			http.Error(w, fmt.Sprintf("create output dir: %v", err), http.StatusInternalServerError)
			return
		}
		imageBase := strings.TrimSuffix(filepath.Base(req.ImagePath), filepath.Ext(req.ImagePath))
		opts := simple.RenderOptions{}
		info := simple.RenderInfoFromExport(req.Settings.Export, 1, req.Name, imageBase, opts)
		info.Trace = trace
		info.OutputDir = outDir
		overrides := makePanelPaths(outDir, sanitizeNameForFile(req.Name), info.Format, req.Settings.IsSpread)
		info.PathOverrides = overrides

		type outputFile struct {
			Panel string `json:"panel"`
			Path  string `json:"path"`
		}
		var outputs []outputFile
		if !req.Settings.IsSpread {
			path, err := simple.RenderSingle(r.Context(), img, result, info)
			if err != nil {
				http.Error(w, fmt.Sprintf("render: %v", err), http.StatusBadRequest)
				return
			}
			outputs = append(outputs, outputFile{Panel: "single", Path: path})
		} else {
			leftPath, rightPath, err := simple.RenderSpread(r.Context(), img, result, info)
			if err != nil {
				http.Error(w, fmt.Sprintf("render: %v", err), http.StatusBadRequest)
				return
			}
			outputs = append(outputs,
				outputFile{Panel: "left", Path: leftPath},
				outputFile{Panel: "right", Path: rightPath},
			)
		}
		respBody := map[string]any{"outputs": outputs, "trace": trace.Lines}
		if trace.Structured != nil {
			respBody["trace_json"] = trace.Structured
		}
		if trace.RenderSingle != nil {
			respBody["render_single_trace"] = trace.RenderSingle
		}
		if trace.RenderSpread != nil {
			respBody["render_spread_trace"] = trace.RenderSpread
		}
		writeJSON(w, http.StatusOK, respBody)
	})

	mux.HandleFunc("/api/v1/yaml", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req spread.ComputeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		yamlText, err := spread.BuildSimpleYAML(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		_, _ = w.Write([]byte(yamlText))
	})

	mux.HandleFunc("/api/v1/yaml/render", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			YAML    string `json:"yaml"`
			BaseDir string `json:"base_dir"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		yamlText := strings.TrimSpace(req.YAML)
		if yamlText == "" {
			http.Error(w, "yaml is required", http.StatusBadRequest)
			return
		}
		doc, err := spread.ParseSimpleBookYAML(yamlText)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		baseDir := strings.TrimSpace(req.BaseDir)
		if baseDir == "" {
			baseDir = s.settings.DataRoot
		}
		resp := yamlRenderResponse{Spreads: make([]yamlRenderSpread, 0, len(doc.Spreads))}
        var htmlOut simple.SpreadOutput
        for idx, spec := range doc.Spreads {
            spreadResp, out, err := s.renderSimpleSpreadSpec(r.Context(), idx+1, spec, baseDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			resp.Spreads = append(resp.Spreads, spreadResp)
            // Build HTML from the last spread (or accumulate later if needed)
            htmlOut = out
		}
        // Synthesize an HTML index blob with one representative spread (simple for now)
        if (len(htmlOut.PanelFiles) > 0) || (htmlOut.StructuredPlacement != nil) || (len(htmlOut.Logs) > 0) {
            tmp := filepath.Join(os.TempDir(), fmt.Sprintf("yaml-index-%d.html", time.Now().UnixNano()))
            if err := simple.WriteHTMLIndex(tmp, []simple.SpreadOutput{htmlOut}); err == nil {
                if b, err := os.ReadFile(tmp); err == nil {
                    resp.HTML = string(b)
                }
                _ = os.Remove(tmp)
            }
        }
        writeJSON(w, http.StatusOK, resp)
	})

	// Project subtree
	mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/projects/")
		if rest == "" {
			http.NotFound(w, r)
			return
		}
		parts := strings.Split(rest, "/")
		id := parts[0]
		if id == "" {
			http.NotFound(w, r)
			return
		}

		// /api/projects/{id}/yaml/book
		if len(parts) == 3 && parts[1] == "yaml" && parts[2] == "book" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			baseDirParam := strings.TrimSpace(r.URL.Query().Get("base_dir"))
			baseDir := baseDirParam
			if baseDir == "" {
				baseDir = s.settings.DataRoot
			} else if !filepath.IsAbs(baseDir) {
				baseDir = filepath.Join(s.settings.DataRoot, baseDir)
			}
			yamlText, err := s.buildProjectBookYAML(id, baseDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
			_, _ = w.Write([]byte(yamlText))
			return
		}

		// /api/projects/{id}/yaml
		if len(parts) == 2 && parts[1] == "yaml" {
			switch r.Method {
			case http.MethodGet:
				fn := filepath.Join(projects.ProjectDir(s.projectsRoot, id), "spec.yaml")
				if _, err := os.Stat(fn); err != nil {
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "text/yaml")
				http.ServeFile(w, r, fn)
				return
			case http.MethodPut:
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "read body", http.StatusBadRequest)
					return
				}
				fn := filepath.Join(projects.ProjectDir(s.projectsRoot, id), "spec.yaml")
				if err := os.MkdirAll(filepath.Dir(fn), 0o755); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if err := os.WriteFile(fn, body, 0o644); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		if len(parts) >= 2 && parts[1] == "images" {
			switch {
			case len(parts) == 2 && r.Method == http.MethodGet:
				s.respondWithAssets(w, id)
				return
			case len(parts) == 2 && r.Method == http.MethodPost:
				s.handleUploadImages(w, r, id)
				return
			case len(parts) == 3 && parts[2] == "reorder" && r.Method == http.MethodPost:
				s.handleReorderImages(w, r, id)
				return
			case len(parts) == 3 && r.Method == http.MethodDelete:
				s.handleDeleteImage(w, id, parts[2])
				return
			case len(parts) == 3 && r.Method == http.MethodGet:
				s.serveProjectImage(w, r, id, parts[2])
				return
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		if len(parts) >= 2 && parts[1] == "pages" {
			switch {
			case len(parts) == 2 && r.Method == http.MethodGet:
				s.handleListPages(w, id)
				return
			case len(parts) == 3:
				num, err := strconv.Atoi(parts[2])
				if err != nil {
					http.Error(w, "invalid page number", http.StatusBadRequest)
					return
				}
				switch r.Method {
				case http.MethodPut:
					s.handleUpsertPage(w, r, id, num)
				case http.MethodDelete:
					s.handleDeletePage(w, id, num)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
				return
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		if len(parts) >= 2 && parts[1] == "spreads" {
			switch {
			case len(parts) == 2 && r.Method == http.MethodGet:
				s.handleListSpreads(w, id)
				return
			case len(parts) == 3:
				num, err := strconv.Atoi(parts[2])
				if err != nil {
					http.Error(w, "invalid spread number", http.StatusBadRequest)
					return
				}
				switch r.Method {
				case http.MethodPut:
					s.handleUpsertSpread(w, r, id, num)
				case http.MethodDelete:
					s.handleDeleteSpread(w, id, num)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
				return
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		// spec to/from ui placeholders
		if len(parts) == 3 && parts[1] == "spec" && parts[2] == "to-ui" && r.Method == http.MethodPost {
			fn := filepath.Join(projects.ProjectDir(s.projectsRoot, id), "spec.yaml")
			var raw string
			if b, err := os.ReadFile(fn); err == nil {
				raw = string(b)
			}
			writeJSON(w, http.StatusOK, map[string]any{"uiState": map[string]any{"yaml": raw}})
			return
		}
		if len(parts) == 3 && parts[1] == "spec" && parts[2] == "from-ui" && r.Method == http.MethodPost {
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			var yamlStr string
			if v, ok := in["yaml"].(string); ok {
				yamlStr = v
			} else if ui, ok := in["uiState"].(map[string]any); ok {
				if v, ok := ui["yaml"].(string); ok {
					yamlStr = v
				}
			}
			if yamlStr == "" {
				http.Error(w, "missing yaml", http.StatusBadRequest)
				return
			}
			fn := filepath.Join(projects.ProjectDir(s.projectsRoot, id), "spec.yaml")
			if err := os.WriteFile(fn, []byte(yamlStr), 0o644); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"yaml": yamlStr})
			return
		}

		// /api/projects/{id}
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				p, err := projects.ReadProject(s.projectsRoot, id)
				if err != nil {
					status := http.StatusInternalServerError
					if os.IsNotExist(err) {
						status = http.StatusNotFound
					}
					http.Error(w, err.Error(), status)
					return
				}
				s.upsertProjectRecord(p)
				writeJSON(w, http.StatusOK, map[string]any{"project": p})
				return
			case http.MethodPut:
				var req struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "invalid json", http.StatusBadRequest)
					return
				}
				p, err := projects.ReadProject(s.projectsRoot, id)
				if err != nil {
					status := http.StatusInternalServerError
					if os.IsNotExist(err) {
						status = http.StatusNotFound
					}
					http.Error(w, err.Error(), status)
					return
				}
				if req.Name != "" {
					p.Name = req.Name
				}
				p.UpdatedAt = time.Now().UTC()
				if err := projects.WriteProject(s.projectsRoot, p); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				s.upsertProjectRecord(p)
				writeJSON(w, http.StatusOK, map[string]any{"project": p})
				return
			case http.MethodDelete:
				if err := projects.DeleteProject(s.projectsRoot, id); err != nil {
					status := http.StatusInternalServerError
					if os.IsNotExist(err) {
						status = http.StatusNotFound
					}
					http.Error(w, err.Error(), status)
					return
				}
				if s.repos != nil {
					if err := s.repos.Projects.Delete(id); err != nil && !errors.Is(err, sql.ErrNoRows) {
						log.Printf("warn: failed to delete project record: %v", err)
					}
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		// /api/projects/{id}/preset
		if len(parts) == 2 && parts[1] == "preset" {
			switch r.Method {
			case http.MethodPost:
				var req struct {
					PresetID string `json:"presetId"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PresetID == "" {
					http.Error(w, "invalid presetId", http.StatusBadRequest)
					return
				}
				if err := presets.ApplyPresetToProject(s.presetsRoot, projects.ProjectDir(s.projectsRoot, id), req.PresetID); err != nil {
					status := http.StatusInternalServerError
					if os.IsNotExist(err) {
						status = http.StatusNotFound
					}
					http.Error(w, err.Error(), status)
					return
				}
				if p, err := projects.ReadProject(s.projectsRoot, id); err == nil {
					p.PresetID = req.PresetID
					p.UpdatedAt = time.Now().UTC()
					_ = projects.WriteProject(s.projectsRoot, p)
					s.upsertProjectRecord(p)
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		// validate
		if len(parts) == 2 && parts[1] == "validate" && r.Method == http.MethodPost {
			issues, details, ok := validation.ValidateProject(s.projectsRoot, id)
			writeJSON(w, http.StatusOK, map[string]any{"ok": ok, "issues": issues, "details": details})
			return
		}

		// renders
		if len(parts) >= 2 && parts[1] == "renders" {
			if len(parts) == 2 && r.Method == http.MethodGet {
				list, err := render.ListProjectRenders(s.projectsRoot, id)
				if err != nil {
					status := http.StatusInternalServerError
					if os.IsNotExist(err) {
						status = http.StatusNotFound
					}
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"renders": list})
				return
			}
			if len(parts) == 5 && parts[3] == "files" && r.Method == http.MethodGet {
				rid := parts[2]
				name := filepath.Base(parts[4])
				fn := filepath.Join(render.ProjectRenderDir(s.projectsRoot, id, rid), name)
				http.ServeFile(w, r, fn)
				return
			}
			if len(parts) == 4 && parts[3] == "download.zip" && r.Method == http.MethodGet {
				rid := parts[2]
				dir := render.ProjectRenderDir(s.projectsRoot, id, rid)
				if err := streamZipDir(w, dir); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}

		// render
		if len(parts) == 2 && parts[1] == "render" && r.Method == http.MethodPost {
			var req struct {
				Test           bool   `json:"test"`
				TestBW         bool   `json:"test_bw"`
				TestDimensions string `json:"test_dimensions"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			log.Printf("render request project=%s test=%v bw=%v dims=%q", id, req.Test, req.TestBW, req.TestDimensions)
			out, err := render.DoProjectRender(s.projectsRoot, id, req.Test, req.TestBW, req.TestDimensions)
			if err != nil {
				log.Printf("render error project=%s: %v", id, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, out)
			return
		}

		http.NotFound(w, r)
	})

	abs, err := filepath.Abs(s.settings.Root)
	if err != nil {
		log.Printf("resolve root: %v", err)
		abs = s.settings.Root
	}
	mux.HandleFunc("/", spaHandler(abs))

	return mux
}

func (s *Server) renderSimpleSpreadSpec(ctx context.Context, index int, spec spread.SimpleSpread, baseDir string) (yamlRenderSpread, simple.SpreadOutput, error) {
	resolvedPath := resolveImagePath(baseDir, spec.ImagePath)
    resp := yamlRenderSpread{Name: spec.Name, ImagePath: resolvedPath}
    out := simple.SpreadOutput{Name: spec.Name, Timestamp: time.Now()}
	if strings.TrimSpace(resolvedPath) == "" {
        return resp, out, fmt.Errorf("spread %q missing image path", spec.Name)
	}
	img, err := decodeImage(resolvedPath)
	if err != nil {
        return resp, out, fmt.Errorf("load image %s: %w", resolvedPath, err)
	}
	bounds := img.Bounds()
	meta := spread.ImageMeta{Width: bounds.Dx(), Height: bounds.Dy()}
    inputs, err := simple.InputsFromSettings(spec.Settings, meta)
	if err != nil {
        return resp, out, fmt.Errorf("inputs simple(%s): %w", spec.Name, err)
	}
	trace := &simple.Trace{UseZerolog: true}
	result := simple.ComputePlacement(inputs, trace)
	tempDir, err := os.MkdirTemp("", "yaml-simple-")
	if err != nil {
        return resp, out, err
	}
	defer os.RemoveAll(tempDir)
	format := normalizeFormat(spec.Settings.Export.Format)
	if format == "" {
		format = "png"
	}
	prefix := sanitizeNameForFile(spec.Name) + "-simple"
	overrides := makePanelPaths(tempDir, prefix, format, spec.Settings.IsSpread)
	imageBase := strings.TrimSuffix(filepath.Base(resolvedPath), filepath.Ext(resolvedPath))
	opts := simple.RenderOptions{}
	info := simple.RenderInfoFromExport(spec.Settings.Export, index, spec.Name, imageBase, opts)
	info.OutputDir = tempDir
	info.PathOverrides = overrides
	panels := make([]panelImage, 0, len(overrides))
    // attach trace handle to render info so render traces are captured
    info.Trace = trace
    if !spec.Settings.IsSpread {
        path, err := simple.RenderSingle(ctx, img, result, info)
		if err != nil {
            return resp, out, fmt.Errorf("render simple(%s): %w", spec.Name, err)
		}
		panel, err := buildPanelImage("single", path)
		if err != nil {
            return resp, out, err
		}
		panels = append(panels, panel)
        out.PanelFiles = append(out.PanelFiles, panel.DataURL)
        out.PanelLabels = append(out.PanelLabels, "single")
	} else {
        leftPath, rightPath, err := simple.RenderSpread(ctx, img, result, info)
		if err != nil {
            return resp, out, fmt.Errorf("render simple(%s): %w", spec.Name, err)
		}
		leftPanel, err := buildPanelImage("left", leftPath)
		if err != nil {
            return resp, out, err
		}
		rightPanel, err := buildPanelImage("right", rightPath)
		if err != nil {
            return resp, out, err
		}
		panels = append(panels, leftPanel, rightPanel)
        out.PanelFiles = append(out.PanelFiles, leftPanel.DataURL, rightPanel.DataURL)
        out.PanelLabels = append(out.PanelLabels, "left", "right")
	}
    resp.Result = result
    resp.Trace = append([]string(nil), trace.Lines...)
    if trace.Structured != nil {
        // include render traces if any were set by the render functions
        if info.Trace != nil {
            if info.Trace.RenderSingle != nil || info.Trace.RenderSpread != nil {
                // Copy placement reference so HTML can find it later when producing an index
            }
        }
        resp.TraceJSON = trace.Structured
        out.StructuredPlacement = trace.Structured
        out.Logs = append(out.Logs, trace.Lines...)
        // Attach render traces when available
        if info.Trace != nil {
            if info.Trace.RenderSingle != nil {
                out.RenderSingle = info.Trace.RenderSingle
            }
            if info.Trace.RenderSpread != nil {
                out.RenderSpread = info.Trace.RenderSpread
            }
        }
    }
	resp.Panels = panels
    return resp, out, nil
}

func resolveImagePath(baseDir, imagePath string) string {
	trimmed := strings.TrimSpace(imagePath)
	if trimmed == "" {
		return ""
	}
	if filepath.IsAbs(trimmed) || strings.TrimSpace(baseDir) == "" {
		return filepath.Clean(trimmed)
	}
	return filepath.Clean(filepath.Join(baseDir, trimmed))
}

func (s *Server) buildProjectBookYAML(projectID, baseDir string) (string, error) {
	project, err := projects.ReadProject(s.projectsRoot, projectID)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("project not found")
		}
		return "", err
	}

	defaults := spread.DefaultSettings()
	var existingDoc *spread.SimpleBookDocument
	specPath := filepath.Join(projects.ProjectDir(s.projectsRoot, projectID), "spec.yaml")
	if data, err := os.ReadFile(specPath); err == nil {
		if doc, parseErr := spread.ParseSimpleBookYAML(string(data)); parseErr == nil {
			existingDoc = doc
			defaults = doc.Defaults
		}
	}

	imagesDir := projects.ProjectImagesDir(s.projectsRoot, projectID)
	nameOrder := orderedImageList(project)
	if len(nameOrder) == 0 {
		entries, err := os.ReadDir(imagesDir)
		if err != nil {
			return "", fmt.Errorf("list images: %w", err)
		}
		files := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			files = append(files, e.Name())
		}
		sort.Strings(files)
		nameOrder = files
	}

	if len(nameOrder) == 0 {
		return "", fmt.Errorf("project has no images to export")
	}

	var overrides map[string]spread.SimpleSpread
	if existingDoc != nil {
		overrides = make(map[string]spread.SimpleSpread)
		for _, sp := range existingDoc.Spreads {
			key := filepath.Base(strings.TrimSpace(sp.ImagePath))
			if key == "" {
				continue
			}
			overrides[key] = sp
		}
	}

	items := make([]spread.BookSpreadItem, 0, len(nameOrder))
	for idx, name := range nameOrder {
		baseName := filepath.Base(name)
		fullPath := filepath.Join(imagesDir, baseName)
		if _, err := os.Stat(fullPath); err != nil {
			continue
		}
		spreadName := fmt.Sprintf("page-%04d", idx+1)
		if overrides != nil {
			if spec, ok := overrides[baseName]; ok {
				if strings.TrimSpace(spec.Name) != "" {
					spreadName = spec.Name
				}
				overrideSettings := spec.Settings
				items = append(items, spread.BookSpreadItem{
					Name:      spreadName,
					ImagePath: fullPath,
					Settings:  &overrideSettings,
				})
				continue
			}
		}
		items = append(items, spread.BookSpreadItem{
			Name:      spreadName,
			ImagePath: fullPath,
		})
	}

	if len(items) == 0 {
		return "", fmt.Errorf("project has no images to export")
	}

	return spread.BuildBookYAML(defaults, items, baseDir)
}

func orderedImageList(p *projects.Project) []string {
	seen := make(map[string]bool)
	appendList := func(dst *[]string, list []string) {
		for _, name := range list {
			clean := strings.TrimSpace(name)
			if clean == "" {
				continue
			}
			clean = filepath.Base(clean)
			if seen[clean] {
				continue
			}
			seen[clean] = true
			*dst = append(*dst, clean)
		}
	}
	var order []string
	appendList(&order, p.Order)
	appendList(&order, p.Images)
	return order
}

func normalizeExtension(format string) string {
	format = normalizeFormat(format)
	if format == "" {
		format = "png"
	}
	if strings.HasPrefix(format, ".") {
		return format
	}
	return "." + format
}

func normalizeFormat(format string) string {
	fmtStr := strings.ToLower(strings.TrimSpace(format))
	switch fmtStr {
	case "", "png":
		return "png"
	case "jpeg":
		return "jpg"
	default:
		return fmtStr
	}
}

func sanitizeNameForFile(name string) string {
	clean := strings.TrimSpace(name)
	if clean == "" {
		return "spread"
	}
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-", "\t", "-", "\n", "-")
	clean = replacer.Replace(clean)
	return clean
}

func buildPanelImage(panel, filePath string) (panelImage, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return panelImage{}, err
	}
	mt := mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath)))
	if mt == "" {
		mt = "image/png"
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		cfg.Width = 0
		cfg.Height = 0
	}
	return panelImage{
		Panel:    panel,
		MimeType: mt,
		DataURL:  fmt.Sprintf("data:%s;base64,%s", mt, encoded),
		Width:    cfg.Width,
		Height:   cfg.Height,
	}, nil
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	if err := s.prepare(); err != nil {
		return err
	}
	defer s.close()
	handler := s.Routes()
	s.httpServer = &http.Server{Addr: s.settings.Addr, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("serving on %s (web from %s)", s.settings.Addr, s.settings.Root)
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) close() {
	if s.db != nil {
		_ = s.db.Close()
		s.db = nil
	}
	s.repos = nil
}

func readImageMeta(path string) (spread.ImageMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return spread.ImageMeta{}, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return spread.ImageMeta{}, err
	}
	return spread.ImageMeta{Width: cfg.Width, Height: cfg.Height}, nil
}

func resolveImageMeta(req spread.ComputeRequest) (spread.ImageMeta, error) {
	if req.Meta != nil && req.Meta.Width > 0 && req.Meta.Height > 0 {
		return spread.ImageMeta{Width: req.Meta.Width, Height: req.Meta.Height}, nil
	}
	if strings.TrimSpace(req.ImagePath) == "" {
		return spread.ImageMeta{}, fmt.Errorf("image_path required to determine metadata")
	}
	return readImageMeta(req.ImagePath)
}

func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func loadImageFromPath(path string) (image.Image, spread.ImageMeta, error) {
	img, err := decodeImage(path)
	if err != nil {
		return nil, spread.ImageMeta{}, err
	}
	bounds := img.Bounds()
	meta := spread.ImageMeta{Width: bounds.Dx(), Height: bounds.Dy()}
	return img, meta, nil
}

func streamZipDir(w http.ResponseWriter, dir string) error {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=render.zip")
	zw := zip.NewWriter(w)
	defer zw.Close()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		fp := filepath.Join(dir, name)
		f, err := os.Open(fp)
		if err != nil {
			return err
		}
		hdr, _ := f.Stat()
		wtr, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, Modified: time.Now(), UncompressedSize64: uint64(hdr.Size())})
		if err != nil {
			_ = f.Close()
			return err
		}
		if _, err := io.Copy(wtr, f); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()
	}
	return nil
}

func spaHandler(root string) http.HandlerFunc {
	indexPath := filepath.Join(root, "index.html")
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		up := pathClean(r.URL.Path)
		fp := filepath.Join(root, up)
		if st, err := os.Stat(fp); err == nil && !st.IsDir() {
			http.ServeFile(w, r, fp)
			return
		}
		http.ServeFile(w, r, indexPath)
	}
}

func pathClean(p string) string {
	if p == "" || p == "/" {
		return "index.html"
	}
	for len(p) > 0 && p[0] == '/' {
		p = p[1:]
	}
	return p
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "..", "-")
	if name == "" {
		return name
	}
	return name
}

func uniqueName(dir, base string) string {
	candidate := base
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	i := 1
	for {
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d%s", stem, i, ext)
		i++
	}
}
