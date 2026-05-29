package services

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/go-go-golems/zine-layout/pkg/repo"
	"github.com/go-go-golems/zine-layout/pkg/zinelayout"
	"gopkg.in/yaml.v3"
)

// ImpositionService assembles rendered laid-out pages into print sheets based on YAML presets.
type ImpositionService struct {
	repos    *repo.Repositories
	dataRoot string
}

// NewImpositionService constructs a new imposition service.
func NewImpositionService(repos *repo.Repositories) *ImpositionService {
	return &ImpositionService{repos: repos}
}

// SetDataRoot configures where project files and presets are stored on disk.
func (s *ImpositionService) SetDataRoot(root string) { s.dataRoot = root }

// SheetResult contains an imposed output sheet and associated metadata.
type SheetResult struct {
	Index  int         `json:"index"`
	Image  image.Image `json:"-"`
	Width  int         `json:"width"`
	Height int         `json:"height"`
}

// ImposeZine loads a preset by ID and imposes the given zine's laid-out pages into output sheets.
// The preset ID maps to a YAML file under {dataRoot}/presets/{presetID}.yaml.
// This returns in-memory images ready for downstream export (e.g., PDF generation).
func (s *ImpositionService) ImposeZine(zineID string, presetID string) ([]*SheetResult, error) {
	if s == nil || s.repos == nil {
		return nil, fmt.Errorf("imposition service not initialized")
	}
	if s.dataRoot == "" {
		return nil, fmt.Errorf("imposition service dataRoot not configured")
	}

	// Load zine and pages
	zine, err := s.repos.Zines.Get(zineID)
	if err != nil {
		return nil, fmt.Errorf("fetch zine: %w", err)
	}
	pages, err := s.repos.Zines.GetPages(zineID)
	if err != nil {
		return nil, fmt.Errorf("fetch zine pages: %w", err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("zine has no pages")
	}

	// Ensure page renders exist and collect input page images (combined > full > thumbnail)
	ps := NewPagesService(s.repos)
	ps.SetDataRoot(s.dataRoot)
	var inputs []image.Image
	for _, zp := range pages {
		page, err := s.repos.LaidOutPages.Get(zp.LaidOutPageID)
		if err != nil {
			return nil, fmt.Errorf("fetch laid-out page %s: %w", zp.LaidOutPageID, err)
		}
		rendered, err := ps.RenderPage(page.ID)
		if err != nil {
			return nil, fmt.Errorf("render page %s: %w", page.ID, err)
		}
		if rendered.ResultJSON == nil {
			return nil, fmt.Errorf("page %s missing render metadata", page.ID)
		}
		var meta struct {
			Variants map[string]string `json:"variants"`
		}
		if err := json.Unmarshal([]byte(*rendered.ResultJSON), &meta); err != nil {
			return nil, fmt.Errorf("decode render metadata for %s: %w", page.ID, err)
		}
		rel := firstNonEmpty(meta.Variants["combined"], meta.Variants["full"], meta.Variants["thumbnail"])
		if rel == "" {
			return nil, fmt.Errorf("no usable variant for page %s", page.ID)
		}
		abs := filepath.Join(s.dataRoot, filepath.FromSlash(rel))
		f, err := os.Open(abs)
		if err != nil {
			return nil, fmt.Errorf("open page image %s: %w", abs, err)
		}
		img, _, err := image.Decode(f)
		_ = f.Close()
		if err != nil {
			return nil, fmt.Errorf("decode page image %s: %w", abs, err)
		}
		inputs = append(inputs, img)
	}

	// Load preset YAML
	presetPath := filepath.Join(s.dataRoot, "presets", filepath.Base(presetID)+".yaml")
	data, err := os.ReadFile(presetPath)
	if err != nil {
		return nil, fmt.Errorf("read preset %s: %w", presetID, err)
	}
	var layout zinelayout.ZineLayout
	if err := yaml.Unmarshal(data, &layout); err != nil {
		return nil, fmt.Errorf("decode preset yaml: %w", err)
	}

	// Generate output sheets
	var results []*SheetResult
	for idx, op := range layout.OutputPages {
		if op == nil {
			continue
		}
		out, err := layout.CreateOutputImage(op, inputs)
		if err != nil {
			return nil, fmt.Errorf("render sheet %d: %w", idx, err)
		}
		b := out.Bounds()
		results = append(results, &SheetResult{Index: idx, Image: out, Width: b.Dx(), Height: b.Dy()})
	}

	// Optionally, write debug PNGs under project dir for visibility (off by default)
	_ = zine // reserved for per-project output in future

	return results, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// SaveSheetsAsPNGs is a utility to dump imposed sheets to the project directory for manual inspection.
// This is optional and not used by HTTP; useful in CLI/testing workflows.
func (s *ImpositionService) SaveSheetsAsPNGs(projectID, zineID string, sheets []*SheetResult) ([]string, error) {
	if s.dataRoot == "" {
		return nil, fmt.Errorf("dataRoot not configured")
	}
	outDir := filepath.Join(s.dataRoot, "projects", projectID, "zines", zineID, "imposed")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	var rels []string
	for _, sheet := range sheets {
		name := fmt.Sprintf("sheet-%02d.png", sheet.Index+1)
		abs := filepath.Join(outDir, name)
		fp, err := os.Create(abs)
		if err != nil {
			return nil, err
		}
		if err := png.Encode(fp, sheet.Image); err != nil {
			_ = fp.Close()
			return nil, err
		}
		_ = fp.Close()
		rels = append(rels, filepath.ToSlash(filepath.Join("projects", projectID, "zines", zineID, "imposed", name)))
	}
	return rels, nil
}
