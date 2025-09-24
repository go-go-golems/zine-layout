package serve

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/presets"
	"github.com/go-go-golems/zine-layout/pkg/projects"
	"github.com/go-go-golems/zine-layout/pkg/render"
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
	Panels    []panelImage  `json:"panels"`
}

type yamlRenderResponse struct {
	Spreads []yamlRenderSpread `json:"spreads"`
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
		writeJSON(w, http.StatusOK, map[string]any{"result": result, "trace": trace.Lines})
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
		writeJSON(w, http.StatusOK, map[string]any{"outputs": outputs, "trace": trace.Lines})
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
		for idx, spec := range doc.Spreads {
			spreadResp, err := s.renderSimpleSpreadSpec(r.Context(), idx+1, spec, baseDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			resp.Spreads = append(resp.Spreads, spreadResp)
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
					_ = projects.WriteProject(s.projectsRoot, p)
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

func (s *Server) renderSimpleSpreadSpec(ctx context.Context, index int, spec spread.SimpleSpread, baseDir string) (yamlRenderSpread, error) {
	resolvedPath := resolveImagePath(baseDir, spec.ImagePath)
	resp := yamlRenderSpread{Name: spec.Name, ImagePath: resolvedPath}
	if strings.TrimSpace(resolvedPath) == "" {
		return resp, fmt.Errorf("spread %q missing image path", spec.Name)
	}
	img, err := decodeImage(resolvedPath)
	if err != nil {
		return resp, fmt.Errorf("load image %s: %w", resolvedPath, err)
	}
	bounds := img.Bounds()
	meta := spread.ImageMeta{Width: bounds.Dx(), Height: bounds.Dy()}
	inputs, err := simple.InputsFromSettings(spec.Settings, meta)
	if err != nil {
		return resp, fmt.Errorf("inputs simple(%s): %w", spec.Name, err)
	}
	trace := &simple.Trace{UseZerolog: true}
	result := simple.ComputePlacement(inputs, trace)
	tempDir, err := os.MkdirTemp("", "yaml-simple-")
	if err != nil {
		return resp, err
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
	if !spec.Settings.IsSpread {
		path, err := simple.RenderSingle(ctx, img, result, info)
		if err != nil {
			return resp, fmt.Errorf("render simple(%s): %w", spec.Name, err)
		}
		panel, err := buildPanelImage("single", path)
		if err != nil {
			return resp, err
		}
		panels = append(panels, panel)
	} else {
		leftPath, rightPath, err := simple.RenderSpread(ctx, img, result, info)
		if err != nil {
			return resp, fmt.Errorf("render simple(%s): %w", spec.Name, err)
		}
		leftPanel, err := buildPanelImage("left", leftPath)
		if err != nil {
			return resp, err
		}
		rightPanel, err := buildPanelImage("right", rightPath)
		if err != nil {
			return resp, err
		}
		panels = append(panels, leftPanel, rightPanel)
	}
	resp.Result = result
	resp.Trace = append([]string(nil), trace.Lines...)
	resp.Panels = panels
	return resp, nil
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
