package projects

import (
    "encoding/json"
    "fmt"
    "image"
    _ "image/png"
    "io"
    "mime/multipart"
    "os"
    "path/filepath"
    "strings"
    "time"
    "math/rand"
)

// Project represents a zine-layout project metadata stored on disk
type Project struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    Images    []string  `json:"images"`
    Order     []string  `json:"order"`
    PresetID  string    `json:"presetId,omitempty"`
}

// ImageItem summarizes an image file stored in a project
type ImageItem struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Width  int    `json:"width"`
    Height int    `json:"height"`
}

func ProjectDir(projectsRoot, id string) string {
    return filepath.Join(projectsRoot, id)
}

func ProjectImagesDir(projectsRoot, id string) string {
    return filepath.Join(ProjectDir(projectsRoot, id), "images")
}

func newProjectID() string {
    ts := time.Now().UTC().Format("20060102T150405Z")
    const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, 6)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return fmt.Sprintf("prj-%s-%s", ts, string(b))
}

func ListProjects(projectsRoot string) ([]Project, error) {
    entries, err := os.ReadDir(projectsRoot)
    if err != nil {
        return nil, err
    }
    res := make([]Project, 0, len(entries))
    for _, e := range entries {
        if !e.IsDir() {
            continue
        }
        p, err := ReadProject(projectsRoot, e.Name())
        if err != nil {
            continue
        }
        res = append(res, *p)
    }
    return res, nil
}

func CreateProject(projectsRoot, name string) (*Project, error) {
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
    if err := WriteProject(projectsRoot, p); err != nil {
        return nil, err
    }
    return p, nil
}

func ReadProject(projectsRoot, id string) (*Project, error) {
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

func WriteProject(projectsRoot string, p *Project) error {
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

func DeleteProject(projectsRoot, id string) error {
    return os.RemoveAll(filepath.Join(projectsRoot, id))
}

func readImageSize(fp string) (int, int, error) {
    f, err := os.Open(fp)
    if err != nil { return 0, 0, err }
    defer f.Close()
    cfg, _, err := image.DecodeConfig(f)
    if err != nil { return 0, 0, err }
    return cfg.Width, cfg.Height, nil
}

func ListProjectImages(projectsRoot, id string) ([]ImageItem, []string, error) {
    p, err := ReadProject(projectsRoot, id)
    if err != nil {
        return nil, nil, err
    }
    dir := ProjectImagesDir(projectsRoot, id)
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

func SavePNGImage(projectsRoot, id string, fh *multipart.FileHeader) (*ImageItem, error) {
    if fh.Size == 0 { return nil, fmt.Errorf("empty file") }
    name := fh.Filename
    lower := strings.ToLower(name)
    if !strings.HasSuffix(lower, ".png") {
        return nil, fmt.Errorf("only .png allowed: %s", name)
    }
    dir := ProjectImagesDir(projectsRoot, id)
    if err := os.MkdirAll(dir, 0o755); err != nil { return nil, err }
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

    p, err := ReadProject(projectsRoot, id)
    if err != nil { return nil, err }
    p.Images = append(p.Images, outName)
    p.Order = append(p.Order, outName)
    p.UpdatedAt = time.Now().UTC()
    if err := WriteProject(projectsRoot, p); err != nil { return nil, err }

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

func SetProjectOrder(projectsRoot, id string, order []string) error {
    p, err := ReadProject(projectsRoot, id)
    if err != nil { return err }
    if len(order) != len(p.Images) { return fmt.Errorf("order length mismatch") }
    seen := map[string]bool{}
    for _, im := range p.Images { seen[im] = true }
    for _, o := range order { if !seen[o] { return fmt.Errorf("unknown image in order: %s", o) } }
    p.Order = order
    p.UpdatedAt = time.Now().UTC()
    return WriteProject(projectsRoot, p)
}

func DeleteProjectImage(projectsRoot, id, imageID string) error {
    p, err := ReadProject(projectsRoot, id)
    if err != nil { return err }
    fp := filepath.Join(ProjectImagesDir(projectsRoot, id), filepath.Base(imageID))
    if err := os.Remove(fp); err != nil { return err }
    filter := func(xs []string) []string {
        out := make([]string, 0, len(xs))
        for _, x := range xs { if x != imageID { out = append(out, x) } }
        return out
    }
    p.Images = filter(p.Images)
    p.Order = filter(p.Order)
    p.UpdatedAt = time.Now().UTC()
    return WriteProject(projectsRoot, p)
}


