package serve

import (
    "archive/zip"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/go-go-golems/zine-layout/pkg/presets"
    "github.com/go-go-golems/zine-layout/pkg/projects"
    "github.com/go-go-golems/zine-layout/pkg/render"
    "github.com/go-go-golems/zine-layout/pkg/validation"
)

type Settings struct {
    Root     string
    DataRoot string
    Addr     string
}

type Server struct {
    settings    Settings
    httpServer  *http.Server
    projectsRoot string
    presetsRoot  string
}

func New(settings Settings) *Server {
    return &Server{ settings: settings }
}

func (s *Server) prepare() error {
    s.projectsRoot = filepath.Join(s.settings.DataRoot, "projects")
    s.presetsRoot = filepath.Join(s.settings.DataRoot, "presets")
    if err := os.MkdirAll(s.projectsRoot, 0o755); err != nil { return fmt.Errorf("create projects root: %w", err) }
    if err := os.MkdirAll(s.presetsRoot, 0o755); err != nil { return fmt.Errorf("create presets root: %w", err) }
    if err := presets.SeedPresetsIfEmpty(s.presetsRoot); err != nil {
        log.Printf("warning: failed to seed presets: %v", err)
    }
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
            if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
            writeJSON(w, http.StatusOK, map[string]any{"presets": list})
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    })
    mux.HandleFunc("/api/presets/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
        id := strings.TrimPrefix(r.URL.Path, "/api/presets/")
        if id == "" { http.NotFound(w, r); return }
        fn := filepath.Join(s.presetsRoot, filepath.Base(id)+".yaml")
        if _, err := os.Stat(fn); err != nil { http.Error(w, "not found", http.StatusNotFound); return }
        w.Header().Set("Content-Type", "text/yaml")
        http.ServeFile(w, r, fn)
    })

    // Projects collection
    mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            prjs, err := projects.ListProjects(s.projectsRoot)
            if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
            writeJSON(w, http.StatusOK, map[string]any{"projects": prjs})
        case http.MethodPost:
            var req struct{ Name, PresetID string }
            _ = json.NewDecoder(r.Body).Decode(&req)
            p, err := projects.CreateProject(s.projectsRoot, req.Name)
            if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
            if req.PresetID != "" {
                if err := presets.ApplyPresetToProject(s.presetsRoot, projects.ProjectDir(s.projectsRoot, p.ID), req.PresetID); err != nil {
                    log.Printf("warning: failed to apply preset to new project: %v", err)
                } else {
                    p.PresetID = req.PresetID
                    _ = projects.WriteProject(s.projectsRoot, p)
                }
            }
            writeJSON(w, http.StatusCreated, map[string]any{"project": p})
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    })

    // Project subtree
    mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
        rest := strings.TrimPrefix(r.URL.Path, "/api/projects/")
        if rest == "" { http.NotFound(w, r); return }
        parts := strings.Split(rest, "/")
        id := parts[0]
        if id == "" { http.NotFound(w, r); return }

        // /api/projects/{id}/yaml
        if len(parts) == 2 && parts[1] == "yaml" {
            switch r.Method {
            case http.MethodGet:
                fn := filepath.Join(projects.ProjectDir(s.projectsRoot, id), "spec.yaml")
                if _, err := os.Stat(fn); err != nil { http.Error(w, "not found", http.StatusNotFound); return }
                w.Header().Set("Content-Type", "text/yaml")
                http.ServeFile(w, r, fn)
                return
            case http.MethodPut:
                body, err := io.ReadAll(r.Body)
                if err != nil { http.Error(w, "read body", http.StatusBadRequest); return }
                fn := filepath.Join(projects.ProjectDir(s.projectsRoot, id), "spec.yaml")
                if err := os.MkdirAll(filepath.Dir(fn), 0o755); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
                if err := os.WriteFile(fn, body, 0o644); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
                writeJSON(w, http.StatusOK, map[string]any{"ok": true})
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
            if b, err := os.ReadFile(fn); err == nil { raw = string(b) }
            writeJSON(w, http.StatusOK, map[string]any{"uiState": map[string]any{"yaml": raw}})
            return
        }
        if len(parts) == 3 && parts[1] == "spec" && parts[2] == "from-ui" && r.Method == http.MethodPost {
            var in map[string]any
            if err := json.NewDecoder(r.Body).Decode(&in); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
            var yamlStr string
            if v, ok := in["yaml"].(string); ok { yamlStr = v } else if ui, ok := in["uiState"].(map[string]any); ok { if v, ok := ui["yaml"].(string); ok { yamlStr = v } }
            if yamlStr == "" { http.Error(w, "missing yaml", http.StatusBadRequest); return }
            fn := filepath.Join(projects.ProjectDir(s.projectsRoot, id), "spec.yaml")
            if err := os.WriteFile(fn, []byte(yamlStr), 0o644); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
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
                    if os.IsNotExist(err) { status = http.StatusNotFound }
                    http.Error(w, err.Error(), status)
                    return
                }
                writeJSON(w, http.StatusOK, map[string]any{"project": p})
                return
            case http.MethodPut:
                var req struct{ Name string `json:"name"` }
                if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
                p, err := projects.ReadProject(s.projectsRoot, id)
                if err != nil {
                    status := http.StatusInternalServerError
                    if os.IsNotExist(err) { status = http.StatusNotFound }
                    http.Error(w, err.Error(), status)
                    return
                }
                if req.Name != "" { p.Name = req.Name }
                p.UpdatedAt = time.Now().UTC()
                if err := projects.WriteProject(s.projectsRoot, p); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
                writeJSON(w, http.StatusOK, map[string]any{"project": p})
                return
            case http.MethodDelete:
                if err := projects.DeleteProject(s.projectsRoot, id); err != nil {
                    status := http.StatusInternalServerError
                    if os.IsNotExist(err) { status = http.StatusNotFound }
                    http.Error(w, err.Error(), status)
                    return
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
                var req struct{ PresetID string `json:"presetId"` }
                if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PresetID == "" { http.Error(w, "invalid presetId", http.StatusBadRequest); return }
                if err := presets.ApplyPresetToProject(s.presetsRoot, projects.ProjectDir(s.projectsRoot, id), req.PresetID); err != nil {
                    status := http.StatusInternalServerError
                    if os.IsNotExist(err) { status = http.StatusNotFound }
                    http.Error(w, err.Error(), status)
                    return
                }
                if p, err := projects.ReadProject(s.projectsRoot, id); err == nil { p.PresetID = req.PresetID; _ = projects.WriteProject(s.projectsRoot, p) }
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
                    if os.IsNotExist(err) { status = http.StatusNotFound }
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
                if err := streamZipDir(w, dir); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError) }
                return
            }
        }

        // render
        if len(parts) == 2 && parts[1] == "render" && r.Method == http.MethodPost {
            var req struct{ Test bool `json:"test"`; TestBW bool `json:"test_bw"`; TestDimensions string `json:"test_dimensions"` }
            _ = json.NewDecoder(r.Body).Decode(&req)
            log.Printf("render request project=%s test=%v bw=%v dims=%q", id, req.Test, req.TestBW, req.TestDimensions)
            out, err := render.DoProjectRender(s.projectsRoot, id, req.Test, req.TestBW, req.TestDimensions)
            if err != nil { log.Printf("render error project=%s: %v", id, err); http.Error(w, err.Error(), http.StatusBadRequest); return }
            writeJSON(w, http.StatusOK, out)
            return
        }

        http.NotFound(w, r)
    })

    abs, err := filepath.Abs(s.settings.Root)
    if err != nil { log.Printf("resolve root: %v", err); abs = s.settings.Root }
    mux.HandleFunc("/", spaHandler(abs))

    return mux
}

func (s *Server) ListenAndServe(ctx context.Context) error {
    if err := s.prepare(); err != nil { return err }
    handler := s.Routes()
    s.httpServer = &http.Server{ Addr: s.settings.Addr, Handler: handler }

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

func streamZipDir(w http.ResponseWriter, dir string) error {
    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", "attachment; filename=render.zip")
    zw := zip.NewWriter(w)
    defer zw.Close()
    entries, err := os.ReadDir(dir)
    if err != nil { return err }
    for _, e := range entries {
        if e.IsDir() { continue }
        name := e.Name()
        fp := filepath.Join(dir, name)
        f, err := os.Open(fp)
        if err != nil { return err }
        hdr, _ := f.Stat()
        wtr, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, Modified: time.Now(), UncompressedSize64: uint64(hdr.Size())})
        if err != nil { _ = f.Close(); return err }
        if _, err := io.Copy(wtr, f); err != nil { _ = f.Close(); return err }
        _ = f.Close()
    }
    return nil
}

func spaHandler(root string) http.HandlerFunc {
    indexPath := filepath.Join(root, "index.html")
    return func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/api/") { http.NotFound(w, r); return }
        up := pathClean(r.URL.Path)
        fp := filepath.Join(root, up)
        if st, err := os.Stat(fp); err == nil && !st.IsDir() { http.ServeFile(w, r, fp); return }
        http.ServeFile(w, r, indexPath)
    }
}

func pathClean(p string) string {
    if p == "" || p == "/" { return "index.html" }
    for len(p) > 0 && p[0] == '/' { p = p[1:] }
    return p
}


