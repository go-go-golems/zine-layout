package presets

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PresetInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Filename string `json:"filename"`
}

func SeedPresetsIfEmpty(presetsRoot string) error {
	hasAny := false
	entries, _ := os.ReadDir(presetsRoot)
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".yaml") {
			hasAny = true
			break
		}
	}
	if hasAny {
		return nil
	}
	candidates := []string{
		filepath.Join("examples", "tests"),
		filepath.Join("examples", "layouts"),
		filepath.Join("zine-layout", "examples", "tests"),
		filepath.Join("zine-layout", "examples", "layouts"),
	}
	for _, cand := range candidates {
		_ = copyYamlFiles(cand, presetsRoot)
	}
	return nil
}

func copyYamlFiles(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".yaml") && !strings.HasSuffix(lower, ".yml") {
			continue
		}
		src := filepath.Join(srcDir, name)
		dst := filepath.Join(dstDir, name)
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		if err := copyFile(src, dst); err != nil {
			log.Printf("failed to copy preset %s: %v", name, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func ListPresets(presetsRoot string) ([]PresetInfo, error) {
	var out []PresetInfo
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		lower := strings.ToLower(d.Name())
		if !strings.HasSuffix(lower, ".yaml") && !strings.HasSuffix(lower, ".yml") {
			return nil
		}
		base := d.Name()
		id := strings.TrimSuffix(strings.TrimSuffix(base, ".yaml"), ".yml")
		out = append(out, PresetInfo{ID: id, Name: id, Filename: base})
		return nil
	}
	_ = filepath.WalkDir(presetsRoot, walkFn)
	return out, nil
}

func ApplyPresetToProject(presetsRoot, projectDir string, presetID string) error {
	src := filepath.Join(presetsRoot, filepath.Base(presetID)+".yaml")
	if _, err := os.Stat(src); err != nil {
		return err
	}
	dst := filepath.Join(projectDir, "spec.yaml")
	return copyFile(src, dst)
}
