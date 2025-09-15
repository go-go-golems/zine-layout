package cmds

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/settings"
)

type ServeCommand struct {
    *cmds.CommandDescription
}

var _ cmds.BareCommand = (*ServeCommand)(nil)

func NewServeCommand() (*ServeCommand, error) {
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }

    return &ServeCommand{
        CommandDescription: cmds.NewCommandDescription(
            "serve",
            cmds.WithShort("Serve the Zine Layout web UI and API"),
            cmds.WithFlags(
                parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithDefault("./cmd/zine-layout/dist"), parameters.WithHelp("Path to built web assets (dist)")),
                parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to server data (projects, presets)")),
                parameters.NewParameterDefinition("addr", parameters.ParameterTypeString, parameters.WithDefault(":8088"), parameters.WithHelp("Listen address")),
            ),
            cmds.WithLayersList(glazedLayer),
        ),
    }, nil
}

type ServeSettings struct {
    Root string `glazed.parameter:"root"`
    DataRoot string `glazed.parameter:"data-root"`
    Addr string `glazed.parameter:"addr"`
}

func (c *ServeCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    s := &ServeSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }

    // Ensure data directories exist
    projectsRoot := filepath.Join(s.DataRoot, "projects")
    if err := os.MkdirAll(projectsRoot, 0o755); err != nil {
        return fmt.Errorf("create projects root: %w", err)
    }
    presetsRoot := filepath.Join(s.DataRoot, "presets")
    if err := os.MkdirAll(presetsRoot, 0o755); err != nil {
        return fmt.Errorf("create presets root: %w", err)
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, http.StatusOK, map[string]any{"ok": true})
    })

    // Projects collection
    mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            projects, err := listProjects(projectsRoot)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
        case http.MethodPost:
            var req struct{ Name string `json:"name"` }
            _ = json.NewDecoder(r.Body).Decode(&req)
            p, err := createProject(projectsRoot, req.Name)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            writeJSON(w, http.StatusCreated, map[string]any{"project": p})
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    })

    // Project item
    mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
        // path: /api/projects/{id}
        rest := strings.TrimPrefix(r.URL.Path, "/api/projects/")
        if rest == "" || strings.Contains(rest, "/") {
            http.NotFound(w, r)
            return
        }
        id := rest
        switch r.Method {
        case http.MethodGet:
            p, err := readProject(projectsRoot, id)
            if err != nil {
                status := http.StatusInternalServerError
                if os.IsNotExist(err) {
                    status = http.StatusNotFound
                }
                http.Error(w, err.Error(), status)
                return
            }
            writeJSON(w, http.StatusOK, map[string]any{"project": p})
        case http.MethodPut:
            var req struct{ Name string `json:"name"` }
            if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "invalid json", http.StatusBadRequest)
                return
            }
            p, err := readProject(projectsRoot, id)
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
            if err := writeProject(projectsRoot, p); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            writeJSON(w, http.StatusOK, map[string]any{"project": p})
        case http.MethodDelete:
            if err := deleteProject(projectsRoot, id); err != nil {
                status := http.StatusInternalServerError
                if os.IsNotExist(err) {
                    status = http.StatusNotFound
                }
                http.Error(w, err.Error(), status)
                return
            }
            writeJSON(w, http.StatusOK, map[string]any{"ok": true})
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    })

    abs, err := filepath.Abs(s.Root)
    if err != nil {
        return fmt.Errorf("resolve root: %w", err)
    }
    if _, err := os.Stat(abs); err != nil {
        log.Printf("warning: web dist not found at %s", abs)
    }
    mux.Handle("/", http.FileServer(http.Dir(abs)))

    log.Printf("serving on %s (web from %s)", s.Addr, abs)
    return http.ListenAndServe(s.Addr, mux)
}

// ===== Helpers and data model =====

type Project struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    Images    []string  `json:"images"`
    Order     []string  `json:"order"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func listProjects(projectsRoot string) ([]Project, error) {
    entries, err := os.ReadDir(projectsRoot)
    if err != nil {
        return nil, err
    }
    res := make([]Project, 0, len(entries))
    for _, e := range entries {
        if !e.IsDir() {
            continue
        }
        p, err := readProject(projectsRoot, e.Name())
        if err != nil {
            continue
        }
        res = append(res, *p)
    }
    return res, nil
}

func createProject(projectsRoot, name string) (*Project, error) {
    if name == "" {
        name = "Untitled"
    }
    id := newProjectID()
    dir := filepath.Join(projectsRoot, id)
    if err := os.MkdirAll(filepath.Join(dir, "images"), 0o755); err != nil {
        return nil, err
    }
    now := time.Now().UTC()
    p := &Project{ID: id, Name: name, CreatedAt: now, UpdatedAt: now, Images: []string{}, Order: []string{}}
    if err := writeProject(projectsRoot, p); err != nil {
        return nil, err
    }
    return p, nil
}

func readProject(projectsRoot, id string) (*Project, error) {
    fn := filepath.Join(projectsRoot, id, "project.json")
    f, err := os.Open(fn)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    var p Project
    if err := json.NewDecoder(f).Decode(&p); err != nil {
        return nil, err
    }
    return &p, nil
}

func writeProject(projectsRoot string, p *Project) error {
    fn := filepath.Join(projectsRoot, p.ID, "project.json")
    tmp := fn + ".tmp"
    f, err := os.Create(tmp)
    if err != nil {
        return err
    }
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    if err := enc.Encode(p); err != nil {
        f.Close()
        _ = os.Remove(tmp)
        return err
    }
    if err := f.Close(); err != nil {
        _ = os.Remove(tmp)
        return err
    }
    return os.Rename(tmp, fn)
}

func deleteProject(projectsRoot, id string) error {
    return os.RemoveAll(filepath.Join(projectsRoot, id))
}

func newProjectID() string {
    // Simple sortable-ish ID with timestamp and random suffix
    ts := time.Now().UTC().Format("20060102T150405Z")
    const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, 6)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return fmt.Sprintf("prj-%s-%s", ts, string(b))
}
