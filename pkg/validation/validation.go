package validation

import (
    "fmt"
    "os"
    "path/filepath"
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
    imgs, _, err := projects.ListProjectImages(projectsRoot, id)
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

func readSpecGridAndPages(dir string) (rows, cols, pages int) {
    fn := filepath.Join(dir, "spec.yaml")
    b, err := os.ReadFile(fn)
    if err != nil { return 0, 0, 0 }
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


