package render

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	apppkg "github.com/go-go-golems/zine-layout/pkg/app"
	"github.com/go-go-golems/zine-layout/pkg/projects"
)

type Result struct {
	RenderID string   `json:"renderId"`
	Files    []string `json:"files"`
}

type ListItem struct {
	ID    string   `json:"id"`
	Files []string `json:"files"`
}

func DoProjectRender(projectsRoot, id string, test, testBW bool, testDimensions string) (*Result, error) {
	projDir := projects.ProjectDir(projectsRoot, id)
	specPath := filepath.Join(projDir, "spec.yaml")
	layouts, err := apppkg.LoadLayoutsFromSpec(specPath, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	if len(layouts) == 0 {
		return nil, fmt.Errorf("spec.yaml did not produce any layouts")
	}
	zl := layouts[0]

	var inputs []image.Image
	if test {
		ppi := 300.0
		if zl.Global != nil && zl.Global.PPI != 0 {
			ppi = zl.Global.PPI
		}
		w, h, err := apppkg.ParseTestDimensions(testDimensions, ppi)
		if err != nil {
			return nil, err
		}
		rows, cols := 1, 1
		if zl.PageSetup != nil {
			if zl.PageSetup.GridSize.Rows > 0 {
				rows = zl.PageSetup.GridSize.Rows
			}
			if zl.PageSetup.GridSize.Columns > 0 {
				cols = zl.PageSetup.GridSize.Columns
			}
		}
		pages := len(zl.OutputPages)
		if pages <= 0 {
			pages = 1
		}
		n := rows * cols * pages
		if n <= 0 {
			n = 1
		}
		inputs, err = apppkg.GenerateTestImages(n, w, h, testBW)
		if err != nil {
			return nil, err
		}
	} else {
		files, err := projectImageFiles(projectsRoot, id)
		if err != nil {
			return nil, err
		}
		inputs, err = apppkg.ReadInputImages(files)
		if err != nil {
			return nil, err
		}
	}

	rid := time.Now().UTC().Format("20060102-150405")
	outDir := ProjectRenderDir(projectsRoot, id, rid)
	files, err := apppkg.RenderOutputs(&zl, inputs, outDir)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = filepath.Base(f)
	}
	return &Result{RenderID: rid, Files: names}, nil
}

func ProjectRendersRoot(projectsRoot, id string) string {
	return filepath.Join(projects.ProjectDir(projectsRoot, id), "renders")
}

func ProjectRenderDir(projectsRoot, id, rid string) string {
	return filepath.Join(ProjectRendersRoot(projectsRoot, id), rid)
}

func ListProjectRenders(projectsRoot, id string) ([]ListItem, error) {
	root := ProjectRendersRoot(projectsRoot, id)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []ListItem{}, nil
		}
		return nil, err
	}
	var out []ListItem
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rid := e.Name()
		files, _ := os.ReadDir(filepath.Join(root, rid))
		names := []string{}
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".png") {
				names = append(names, f.Name())
			}
		}
		out = append(out, ListItem{ID: rid, Files: names})
	}
	return out, nil
}

func projectImageFiles(projectsRoot, id string) ([]string, error) {
	dir := projects.ProjectImagesDir(projectsRoot, id)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".png") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	files := make([]string, 0, len(names))
	for _, name := range names {
		files = append(files, filepath.Join(dir, name))
	}
	return files, nil
}
