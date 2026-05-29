package serve

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/presets"
	"github.com/go-go-golems/zine-layout/pkg/repo"
	sqliterepo "github.com/go-go-golems/zine-layout/pkg/repo/sqlite"
	"github.com/go-go-golems/zine-layout/pkg/services"
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
	layout       *services.LayoutService
	pages        *services.PagesService
	zines        *services.ZinesService
	impose       *services.ImpositionService
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
	presetsRoot := filepath.Join(s.settings.DataRoot, "presets")

	if err := os.MkdirAll(s.projectsRoot, 0o755); err != nil {
		return fmt.Errorf("create projects root: %w", err)
	}
	if err := os.MkdirAll(s.uploadsRoot, 0o755); err != nil {
		return fmt.Errorf("create uploads root: %w", err)
	}
	if err := os.MkdirAll(presetsRoot, 0o755); err != nil {
		return fmt.Errorf("create presets root: %w", err)
	}
	_ = presets.SeedPresetsIfEmpty(presetsRoot)
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
	s.layout = services.NewLayoutService(repos)
	s.pages = services.NewPagesService(repos)
	// Provide data root to services that need filesystem access
	if s.pages != nil {
		s.pages.SetDataRoot(s.settings.DataRoot)
	}
	s.zines = services.NewZinesService(repos)
	s.impose = services.NewImpositionService(repos)
	if s.impose != nil {
		s.impose.SetDataRoot(s.settings.DataRoot)
	}
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

	mux.HandleFunc("/api/image-layout-templates", s.handleImageLayoutTemplates)
	mux.HandleFunc("/api/image-layout-templates/", s.handleImageLayoutTemplateRoutes)

	mux.HandleFunc("/api/laid-out-images/", s.handleLaidOutImageRoutes)

	mux.HandleFunc("/api/layout-sequences", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/layout-sequences/", s.handleLayoutSequenceRoutes)

	mux.HandleFunc("/api/page-templates", s.handlePageTemplates)
	mux.HandleFunc("/api/page-templates/", s.handlePageTemplateRoutes)

	mux.HandleFunc("/api/laid-out-pages", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/laid-out-pages/", s.handleLaidOutPageRoutes)

	mux.HandleFunc("/api/zines", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/zines/", s.handleZineRoutes)

	// Serve project files (images) under /projects/{id}/...
	projectFiles := http.StripPrefix("/projects/", http.FileServer(http.Dir(s.projectsRoot)))
	mux.Handle("/projects/", projectFiles)

	if s.settings.Root != "" {
		distDir := s.settings.Root
		fileServer := http.FileServer(http.Dir(distDir))
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "" || path == "/" {
				fileServer.ServeHTTP(w, r)
				return
			}

			fullPath := filepath.Join(distDir, filepath.Clean(path))
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}

			indexPath := filepath.Join(distDir, "index.html")
			if _, err := os.Stat(indexPath); err != nil {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, indexPath)
		}))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.RequestURI())
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) close() {
	if s.db != nil {
		_ = s.db.Close()
		s.db = nil
	}
	s.repos = nil
	s.layout = nil
	s.pages = nil
	s.zines = nil
}

// HTTP handlers ----------------------------------------------------------------

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
