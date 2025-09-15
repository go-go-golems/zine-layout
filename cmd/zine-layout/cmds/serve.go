package cmds

import (
    "context"
    "encoding/json"
    "fmt"
    "image"
    _ "image/png"
    "io"
    "log"
    "math/rand"
    "net/http"
    "mime/multipart"
    "os"
    "os/signal"
    "path/filepath"
    "strings"
    "syscall"
    "time"
    "io/fs"
    yaml "gopkg.in/yaml.v3"

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
    // Seed presets from examples if directory is empty
    if err := seedPresetsIfEmpty(presetsRoot); err != nil {
        log.Printf("warning: failed to seed presets: %v", err)
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, http.StatusOK, map[string]any{"ok": true})
    })

    // Presets
    mux.HandleFunc("/api/presets", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            presets, err := listPresets(presetsRoot)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            writeJSON(w, http.StatusOK, map[string]any{"presets": presets})
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
        fn := filepath.Join(presetsRoot, filepath.Base(id)+".yaml")
        if _, err := os.Stat(fn); err != nil {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        // Serve as text/yaml
        w.Header().Set("Content-Type", "text/yaml")
        http.ServeFile(w, r, fn)
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
            var req struct{
                Name string `json:"name"`
                PresetID string `json:"presetId"`
            }
            _ = json.NewDecoder(r.Body).Decode(&req)
            p, err := createProject(projectsRoot, req.Name)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            if req.PresetID != "" {
                if err := applyPresetToProject(presetsRoot, projectsRoot, p.ID, req.PresetID); err != nil {
                    log.Printf("warning: failed to apply preset to new project: %v", err)
                } else {
                    // Update project metadata with preset
                    p.PresetID = req.PresetID
                    _ = writeProject(projectsRoot, p)
                }
            }
            writeJSON(w, http.StatusCreated, map[string]any{"project": p})
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    })

    // Project subtree router
    mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
        // path: /api/projects/{id}[...]
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

        // /api/projects/{id}/yaml
        if len(parts) == 2 && parts[1] == "yaml" {
            switch r.Method {
            case http.MethodGet:
                fn := filepath.Join(projectDir(projectsRoot, id), "spec.yaml")
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
                fn := filepath.Join(projectDir(projectsRoot, id), "spec.yaml")
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

        // /api/projects/{id}/spec/to-ui and /from-ui (simple placeholder)
        if len(parts) == 3 && parts[1] == "spec" && parts[2] == "to-ui" && r.Method == http.MethodPost {
            fn := filepath.Join(projectDir(projectsRoot, id), "spec.yaml")
            var raw string
            if b, err := os.ReadFile(fn); err == nil { raw = string(b) }
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
                if v, ok := ui["yaml"].(string); ok { yamlStr = v }
            }
            if yamlStr == "" {
                http.Error(w, "missing yaml", http.StatusBadRequest)
                return
            }
            fn := filepath.Join(projectDir(projectsRoot, id), "spec.yaml")
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
                return
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
                return
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
                if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PresetID == "" {
                    http.Error(w, "invalid presetId", http.StatusBadRequest)
                    return
                }
                if err := applyPresetToProject(presetsRoot, projectsRoot, id, req.PresetID); err != nil {
                    status := http.StatusInternalServerError
                    if os.IsNotExist(err) { status = http.StatusNotFound }
                    http.Error(w, err.Error(), status)
                    return
                }
                // Update project metadata with preset
                if p, err := readProject(projectsRoot, id); err == nil {
                    p.PresetID = req.PresetID
                    _ = writeProject(projectsRoot, p)
                }
                writeJSON(w, http.StatusOK, map[string]any{"ok": true})
                return
            default:
                http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
                return
            }
        }

        // /api/projects/{id}/validate
        if len(parts) == 2 && parts[1] == "validate" && r.Method == http.MethodPost {
            issues, details, ok := validateProject(projectsRoot, id)
            writeJSON(w, http.StatusOK, map[string]any{"ok": ok, "issues": issues, "details": details})
            return
        }

        // /api/projects/{id}/images[...]
        if len(parts) >= 2 && parts[1] == "images" {
            switch {
            // List images
            case len(parts) == 2 && r.Method == http.MethodGet:
                imgs, order, err := listProjectImages(projectsRoot, id)
                if err != nil {
                    status := http.StatusInternalServerError
                    if os.IsNotExist(err) {
                        status = http.StatusNotFound
                    }
                    http.Error(w, err.Error(), status)
                    return
                }
                writeJSON(w, http.StatusOK, map[string]any{"images": imgs, "order": order})
                return
            // Upload images (multipart form)
            case len(parts) == 2 && r.Method == http.MethodPost:
                if err := r.ParseMultipartForm(64 << 20); err != nil { // 64MB
                    http.Error(w, "invalid multipart form", http.StatusBadRequest)
                    return
                }
                files := r.MultipartForm.File["images[]"]
                if len(files) == 0 {
                    http.Error(w, "no images[] files provided", http.StatusBadRequest)
                    return
                }
                saved := make([]ImageItem, 0, len(files))
                for _, fh := range files {
                    it, err := savePngImage(projectsRoot, id, fh)
                    if err != nil {
                        http.Error(w, err.Error(), http.StatusBadRequest)
                        return
                    }
                    saved = append(saved, *it)
                }
                writeJSON(w, http.StatusCreated, map[string]any{"images": saved})
                return
            // Reorder images
            case len(parts) == 3 && parts[2] == "reorder" && r.Method == http.MethodPost:
                var req struct{ Order []string `json:"order"` }
                if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Order) == 0 {
                    http.Error(w, "invalid order", http.StatusBadRequest)
                    return
                }
                if err := setProjectOrder(projectsRoot, id, req.Order); err != nil {
                    http.Error(w, err.Error(), http.StatusBadRequest)
                    return
                }
                writeJSON(w, http.StatusOK, map[string]any{"ok": true})
                return
            // Serve or delete single image
            case len(parts) == 3 && r.Method == http.MethodGet:
                imageID := parts[2]
                fn := filepath.Join(projectImagesDir(projectsRoot, id), filepath.Base(imageID))
                http.ServeFile(w, r, fn)
                return
            case len(parts) == 3 && r.Method == http.MethodDelete:
                imageID := parts[2]
                if err := deleteProjectImage(projectsRoot, id, imageID); err != nil {
                    status := http.StatusInternalServerError
                    if os.IsNotExist(err) {
                        status = http.StatusNotFound
                    }
                    http.Error(w, err.Error(), status)
                    return
                }
                writeJSON(w, http.StatusOK, map[string]any{"ok": true})
                return
            }
        }

        http.NotFound(w, r)
    })

    abs, err := filepath.Abs(s.Root)
    if err != nil {
        return fmt.Errorf("resolve root: %w", err)
    }
    if _, err := os.Stat(abs); err != nil {
        log.Printf("warning: web dist not found at %s", abs)
    }
    // SPA file server with index.html fallback for client-side routes
    mux.HandleFunc("/", spaHandler(abs))

    srv := &http.Server{ Addr: s.Addr, Handler: mux }

    // Start server
    errCh := make(chan error, 1)
    go func() {
        log.Printf("serving on %s (web from %s)", s.Addr, abs)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errCh <- err
        }
        close(errCh)
    }()

    // Wait for Ctrl-C (SIGINT) or SIGTERM
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    select {
    case sig := <-sigCh:
        log.Printf("received signal %s, shutting down...", sig)
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = srv.Shutdown(ctx)
        return nil
    case err := <-errCh:
        return err
    }
}

// ===== Helpers and data model =====

type Project struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    Images    []string  `json:"images"`
    Order     []string  `json:"order"`
    PresetID  string    `json:"presetId,omitempty"`
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

// ===== Images helpers =====

type ImageItem struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Width  int    `json:"width"`
    Height int    `json:"height"`
}

func projectDir(projectsRoot, id string) string {
    return filepath.Join(projectsRoot, id)
}

func projectImagesDir(projectsRoot, id string) string {
    return filepath.Join(projectDir(projectsRoot, id), "images")
}

func listProjectImages(projectsRoot, id string) ([]ImageItem, []string, error) {
    p, err := readProject(projectsRoot, id)
    if err != nil {
        return nil, nil, err
    }
    dir := projectImagesDir(projectsRoot, id)
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, nil, err
    }
    images := make([]ImageItem, 0, len(entries))
    for _, e := range entries {
        if e.IsDir() { continue }
        name := e.Name()
        fp := filepath.Join(dir, name)
        w, h, err := readImageSize(fp)
        if err != nil { continue }
        images = append(images, ImageItem{ID: name, Name: name, Width: w, Height: h})
    }
    return images, p.Order, nil
}

func readImageSize(fp string) (int, int, error) {
    f, err := os.Open(fp)
    if err != nil { return 0, 0, err }
    defer f.Close()
    cfg, _, err := image.DecodeConfig(f)
    if err != nil { return 0, 0, err }
    return cfg.Width, cfg.Height, nil
}

func savePngImage(projectsRoot, id string, fh *multipart.FileHeader) (*ImageItem, error) {
    if fh.Size == 0 { return nil, fmt.Errorf("empty file") }
    // Simple extension check
    name := fh.Filename
    lower := strings.ToLower(name)
    if !strings.HasSuffix(lower, ".png") {
        return nil, fmt.Errorf("only .png allowed: %s", name)
    }
    dir := projectImagesDir(projectsRoot, id)
    if err := os.MkdirAll(dir, 0o755); err != nil { return nil, err }
    // Determine next index filename: 0001.png, 0002.png, ...
    next := nextImageNumber(dir)
    outName := fmt.Sprintf("%04d.png", next)
    dstPath := filepath.Join(dir, outName)

    src, err := fh.Open()
    if err != nil { return nil, err }
    defer src.Close()
    dst, err := os.Create(dstPath)
    if err != nil { return nil, err }
    if _, err := io.Copy(dst, src); err != nil {
        _ = dst.Close(); _ = os.Remove(dstPath)
        return nil, err
    }
    if err := dst.Close(); err != nil { return nil, err }

    // Update project metadata
    p, err := readProject(projectsRoot, id)
    if err != nil { return nil, err }
    p.Images = append(p.Images, outName)
    p.Order = append(p.Order, outName)
    p.UpdatedAt = time.Now().UTC()
    if err := writeProject(projectsRoot, p); err != nil { return nil, err }

    w, h, err := readImageSize(dstPath)
    if err != nil { return nil, err }
    item := &ImageItem{ID: outName, Name: outName, Width: w, Height: h}
    return item, nil
}

func nextImageNumber(dir string) int {
    max := 0
    entries, _ := os.ReadDir(dir)
    for _, e := range entries {
        if e.IsDir() { continue }
        name := e.Name()
        if len(name) != 8 || !strings.HasSuffix(name, ".png") { continue }
        nStr := name[:4]
        var n int
        _, err := fmt.Sscanf(nStr, "%04d", &n)
        if err == nil && n > max { max = n }
    }
    return max + 1
}

func setProjectOrder(projectsRoot, id string, order []string) error {
    p, err := readProject(projectsRoot, id)
    if err != nil { return err }
    // validate that order is a permutation of p.Images
    if len(order) != len(p.Images) { return fmt.Errorf("order length mismatch") }
    seen := map[string]bool{}
    for _, im := range p.Images { seen[im] = true }
    for _, o := range order { if !seen[o] { return fmt.Errorf("unknown image in order: %s", o) } }
    p.Order = order
    p.UpdatedAt = time.Now().UTC()
    return writeProject(projectsRoot, p)
}

func deleteProjectImage(projectsRoot, id, imageID string) error {
    p, err := readProject(projectsRoot, id)
    if err != nil { return err }
    // Remove file
    fp := filepath.Join(projectImagesDir(projectsRoot, id), filepath.Base(imageID))
    if err := os.Remove(fp); err != nil { return err }
    // Filter metadata
    filter := func(xs []string) []string {
        out := make([]string, 0, len(xs))
        for _, x := range xs { if x != imageID { out = append(out, x) } }
        return out
    }
    p.Images = filter(p.Images)
    p.Order = filter(p.Order)
    p.UpdatedAt = time.Now().UTC()
    return writeProject(projectsRoot, p)
}

// end images helpers
// ===== Presets helpers =====

type PresetInfo struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Filename string `json:"filename"`
}

func seedPresetsIfEmpty(presetsRoot string) error {
    // if directory has any .yaml, do nothing
    hasAny := false
    entries, _ := os.ReadDir(presetsRoot)
    for _, e := range entries {
        if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".yaml") {
            hasAny = true
            break
        }
    }
    if hasAny { return nil }
    // Try to copy from known example dirs relative to repo structure
    candidates := []string{
        filepath.Join("examples", "tests"),
        filepath.Join("examples", "layouts"),
        filepath.Join("zine-layout", "examples", "tests"),
        filepath.Join("zine-layout", "examples", "layouts"),
    }
    for _, cand := range candidates {
        _ = copyYamlFiles(cand, presetsRoot)
    }
    // Best-effort, no error even if none found
    return nil
}

func copyYamlFiles(srcDir, dstDir string) error {
    entries, err := os.ReadDir(srcDir)
    if err != nil { return err }
    for _, e := range entries {
        if e.IsDir() { continue }
        name := e.Name()
        lower := strings.ToLower(name)
        if !strings.HasSuffix(lower, ".yaml") && !strings.HasSuffix(lower, ".yml") { continue }
        src := filepath.Join(srcDir, name)
        dst := filepath.Join(dstDir, name)
        if _, err := os.Stat(dst); err == nil { continue }
        if err := copyFile(src, dst); err != nil {
            log.Printf("failed to copy preset %s: %v", name, err)
        }
    }
    return nil
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil { return err }
    defer in.Close()
    if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
    out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
    if err != nil { return err }
    if _, err := io.Copy(out, in); err != nil { _ = out.Close(); return err }
    return out.Close()
}

func listPresets(presetsRoot string) ([]PresetInfo, error) {
    var out []PresetInfo
    walkFn := func(path string, d fs.DirEntry, err error) error {
        if err != nil { return nil }
        if d.IsDir() { return nil }
        lower := strings.ToLower(d.Name())
        if !strings.HasSuffix(lower, ".yaml") && !strings.HasSuffix(lower, ".yml") { return nil }
        base := d.Name()
        id := strings.TrimSuffix(strings.TrimSuffix(base, ".yaml"), ".yml")
        out = append(out, PresetInfo{ID: id, Name: id, Filename: base})
        return nil
    }
    _ = filepath.WalkDir(presetsRoot, walkFn)
    return out, nil
}

func applyPresetToProject(presetsRoot, projectsRoot, projectID, presetID string) error {
    // locate preset file
    src := filepath.Join(presetsRoot, filepath.Base(presetID)+".yaml")
    if _, err := os.Stat(src); err != nil {
        return err
    }
    dst := filepath.Join(projectDir(projectsRoot, projectID), "spec.yaml")
    return copyFile(src, dst)
}

// ===== Validation =====
type validationDetails struct {
    Count    int `json:"count"`
    Width    int `json:"width"`
    Height   int `json:"height"`
    Rows     int `json:"rows"`
    Columns  int `json:"columns"`
    Pages    int `json:"pages"`
    Multiple int `json:"multiple"`
}

func validateProject(projectsRoot, id string) ([]string, *validationDetails, bool) {
    issues := []string{}
    // Check image sizes all equal
    imgs, _, err := listProjectImages(projectsRoot, id)
    if err != nil {
        issues = append(issues, fmt.Sprintf("read images: %v", err))
        return issues, nil, false
    }
    var w0, h0 int
    for i, im := range imgs {
        if i == 0 { w0, h0 = im.Width, im.Height }
        if im.Width != w0 || im.Height != h0 {
            issues = append(issues, fmt.Sprintf("image %s has size %dx%d, expected %dx%d", im.Name, im.Width, im.Height, w0, h0))
        }
    }
    rows, cols, pages := readSpecGridAndPages(projectDir(projectsRoot, id))
    mult := 0
    if rows > 0 && cols > 0 && pages > 0 {
        mult = rows * cols * pages
        if len(imgs) > 0 && (len(imgs)%mult != 0) {
            issues = append(issues, fmt.Sprintf("image count %d is not a multiple of rows*columns*pages=%d", len(imgs), mult))
        }
    } else {
        issues = append(issues, "spec.yaml missing or incomplete grid/pages; skipping multiple check")
    }
    ok := len(issues) == 0
    det := &validationDetails{Count: len(imgs), Width: w0, Height: h0, Rows: rows, Columns: cols, Pages: pages, Multiple: mult}
    return issues, det, ok
}

func readSpecGridAndPages(dir string) (rows, cols, pages int) {
    fn := filepath.Join(dir, "spec.yaml")
    b, err := os.ReadFile(fn)
    if err != nil { return 0, 0, 0 }
    // minimal yaml mapping
    var doc struct {
        PageSetup struct {
            GridSize struct {
                Rows    int `yaml:"rows"`
                Columns int `yaml:"columns"`
            } `yaml:"grid_size"`
        } `yaml:"page_setup"`
        OutputPages []any `yaml:"output_pages"`
    }
    if err := yaml.Unmarshal(b, &doc); err != nil { return 0, 0, 0 }
    pages = len(doc.OutputPages)
    return doc.PageSetup.GridSize.Rows, doc.PageSetup.GridSize.Columns, pages
}

// spaHandler serves static files from root, and falls back to index.html
// for any non-API request whose file does not exist. This allows BrowserRouter
// to handle client-side routes like /projects or /projects/:id when navigated
// to directly.
func spaHandler(root string) http.HandlerFunc {
    indexPath := filepath.Join(root, "index.html")
    return func(w http.ResponseWriter, r *http.Request) {
        // Never intercept API
        if strings.HasPrefix(r.URL.Path, "/api/") {
            http.NotFound(w, r)
            return
        }

        // Sanitize path and check if a static file exists
        up := pathClean(r.URL.Path)
        fp := filepath.Join(root, up)
        if st, err := os.Stat(fp); err == nil && !st.IsDir() {
            http.ServeFile(w, r, fp)
            return
        }
        // Fallback to index.html for SPA routes
        http.ServeFile(w, r, indexPath)
    }
}

// pathClean keeps leading slash semantics similar to http.FileServer
func pathClean(p string) string {
    if p == "" || p == "/" {
        return "index.html" // let caller join with root/index.html
    }
    // trim leading '/'
    for len(p) > 0 && p[0] == '/' {
        p = p[1:]
    }
    return p
}
