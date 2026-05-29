package validation

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/go-go-golems/zine-layout/pkg/projects"
)

type Details struct {
	Count    int `json:"count"`
	Width    int `json:"width"`
	Height   int `json:"height"`
	Rows     int `json:"rows"`
	Columns  int `json:"columns"`
	Pages    int `json:"pages"`
	Multiple int `json:"multiple"`
}

func ValidateProject(projectsRoot, id string) ([]string, *Details, bool) {
	issues := []string{}
	imgs, err := listImages(projectsRoot, id)
	if err != nil {
		issues = append(issues, fmt.Sprintf("read images: %v", err))
		return issues, nil, false
	}
	var w0, h0 int
	for i, im := range imgs {
		if i == 0 {
			w0, h0 = im.Width, im.Height
		}
		if im.Width != w0 || im.Height != h0 {
			issues = append(issues, fmt.Sprintf("image %s has size %dx%d, expected %dx%d", im.Name, im.Width, im.Height, w0, h0))
		}
	}
	rows, cols, pages := readSpecGridAndPages(projects.ProjectDir(projectsRoot, id))
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
	det := &Details{Count: len(imgs), Width: w0, Height: h0, Rows: rows, Columns: cols, Pages: pages, Multiple: mult}
	return issues, det, ok
}

type imageInfo struct {
	Name   string
	Width  int
	Height int
}

func listImages(projectsRoot, id string) ([]imageInfo, error) {
	dir := projects.ProjectImagesDir(projectsRoot, id)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	imgs := make([]imageInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(strings.ToLower(name)) != ".png" {
			continue
		}
		width, height, err := imageDimensions(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		imgs = append(imgs, imageInfo{Name: name, Width: width, Height: height})
	}
	return imgs, nil
}

func imageDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = f.Close() }()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func readSpecGridAndPages(dir string) (int, int, int) {
	fn := filepath.Join(dir, "spec.yaml")
	b, err := os.ReadFile(fn)
	if err != nil {
		return 0, 0, 0
	}
	var doc struct {
		PageSetup struct {
			GridSize struct {
				Rows    int `yaml:"rows"`
				Columns int `yaml:"columns"`
			} `yaml:"grid_size"`
		} `yaml:"page_setup"`
		OutputPages []any `yaml:"output_pages"`
	}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return 0, 0, 0
	}
	pageCount := len(doc.OutputPages)
	return doc.PageSetup.GridSize.Rows, doc.PageSetup.GridSize.Columns, pageCount
}
