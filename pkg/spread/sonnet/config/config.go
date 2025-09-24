package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
)

// Config represents the top-level YAML document.
type Config struct {
	Version  string         `yaml:"version"`
	Defaults SpreadSettings `yaml:"defaults"`
	Assets   []AssetConfig  `yaml:"assets"`
	Spreads  []SpreadEntry  `yaml:"spreads"`
}

// SpreadSettings captures values shared between defaults and per-spread overrides.
type SpreadSettings struct {
	Paper    *PaperConfig    `yaml:"paper"`
	Margins  *MarginsConfig  `yaml:"margins"`
	Spread   *SpreadConfig   `yaml:"spread"`
	Crop     *CropConfig     `yaml:"crop"`
	Scale    *ScaleConfig    `yaml:"scale"`
	Position *PositionConfig `yaml:"position"`
	Export   *ExportConfig   `yaml:"export"`
}

// PaperConfig describes paper geometry and resolution.
type PaperConfig struct {
	Size        string   `yaml:"size"`
	WidthIn     *float64 `yaml:"width_in"`
	HeightIn    *float64 `yaml:"height_in"`
	Orientation string   `yaml:"orientation"`
	DPI         *float64 `yaml:"dpi"`
}

// MarginsConfig stores page margins in inches.
type MarginsConfig struct {
	AllIn    *float64 `yaml:"all_in"`
	TopIn    *float64 `yaml:"top_in"`
	RightIn  *float64 `yaml:"right_in"`
	BottomIn *float64 `yaml:"bottom_in"`
	LeftIn   *float64 `yaml:"left_in"`
}

// SpreadConfig toggles spread rendering and gutter.
type SpreadConfig struct {
	IsSpread *bool    `yaml:"is_spread"`
	GutterIn *float64 `yaml:"gutter_in"`
}

// CropConfig handles pre-crop ratio and fill behaviour.
type CropConfig struct {
	Ratio  any   `yaml:"ratio"`
	ToFill *bool `yaml:"to_fill"`
}

// ScaleConfig stores user scale factor.
type ScaleConfig struct {
	UserScale *float64 `yaml:"user_scale"`
}

// PositionConfig holds placement values and units.
type PositionConfig struct {
	X     *float64 `yaml:"x"`
	Y     *float64 `yaml:"y"`
	Units string   `yaml:"units"`
}

// ExportConfig defines output parameters.
type ExportConfig struct {
	Format           string `yaml:"format"`
	Quality          *int   `yaml:"quality"`
	Background       string `yaml:"background"`
	OutDir           string `yaml:"out_dir"`
	FilenameTemplate string `yaml:"filename_template"`
}

// AssetConfig lists named bundles of images.
type AssetConfig struct {
	Name    string   `yaml:"name"`
	Path    string   `yaml:"path"`
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// SpreadEntry describes one spread entry in the YAML.
type SpreadEntry struct {
	Name     string          `yaml:"name"`
	Image    ImageSelector   `yaml:"image"`
	ForEach  *ForEachConfig  `yaml:"for_each"`
	Paper    *PaperConfig    `yaml:"paper"`
	Margins  *MarginsConfig  `yaml:"margins"`
	Spread   *SpreadConfig   `yaml:"spread"`
	Crop     *CropConfig     `yaml:"crop"`
	Scale    *ScaleConfig    `yaml:"scale"`
	Position *PositionConfig `yaml:"position"`
	Export   *ExportConfig   `yaml:"export"`
}

// ForEachConfig expands a spread into multiple outputs.
type ForEachConfig struct {
	Images       ImageSelector `yaml:"images"`
	NameTemplate string        `yaml:"name_template"`
}

// ImageSelector references image paths or asset names.
type ImageSelector struct {
	Paths []string
	Asset string
}

// UnmarshalYAML customises parsing for ImageSelector.
func (s *ImageSelector) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var v string
		if err := value.Decode(&v); err != nil {
			return err
		}
		if strings.TrimSpace(v) == "" {
			return errors.New("image selector cannot be empty")
		}
		s.Paths = []string{v}
	case yaml.SequenceNode:
		var arr []string
		if err := value.Decode(&arr); err != nil {
			return err
		}
		if len(arr) == 0 {
			return errors.New("image selector sequence cannot be empty")
		}
		s.Paths = append([]string{}, arr...)
	case yaml.MappingNode:
		type raw struct {
			Asset string `yaml:"asset"`
		}
		var r raw
		if err := value.Decode(&r); err != nil {
			return err
		}
		if r.Asset == "" {
			return errors.New("image selector mapping must include asset")
		}
		s.Asset = r.Asset
	case 0:
		// allow missing node -> leave zero value
		return nil
	default:
		return fmt.Errorf("unsupported image selector kind: %v", value.Kind)
	}
	return nil
}

// Validate performs high-level config validation.
func (c *Config) Validate() error {
	if c.Version == "" {
		return errors.New("version is required")
	}
	if len(c.Spreads) == 0 {
		return errors.New("at least one spread is required")
	}
	seen := map[string]bool{}
	for _, asset := range c.Assets {
		if asset.Name == "" {
			return errors.New("asset name is required")
		}
		if seen[asset.Name] {
			return fmt.Errorf("duplicate asset name: %s", asset.Name)
		}
		seen[asset.Name] = true
	}
	for i, sp := range c.Spreads {
		if strings.TrimSpace(sp.Name) == "" {
			return fmt.Errorf("spread %d is missing a name", i)
		}
		if sp.ForEach == nil && sp.Image.IsEmpty() {
			return fmt.Errorf("spread %s must provide image or for_each", sp.Name)
		}
		if sp.ForEach != nil && sp.ForEach.Images.IsEmpty() {
			return fmt.Errorf("spread %s for_each.images cannot be empty", sp.Name)
		}
	}
	return nil
}

// IsEmpty reports whether the selector has no content.
func (s ImageSelector) IsEmpty() bool {
	return len(s.Paths) == 0 && s.Asset == ""
}

// Merge applies overrides from b atop a copy of a.
func (base SpreadSettings) Merge(override SpreadSettings) SpreadSettings {
	out := base
	if override.Paper != nil {
		if out.Paper == nil {
			out.Paper = &PaperConfig{}
		}
		mergePaper(out.Paper, override.Paper)
	}
	if override.Margins != nil {
		if out.Margins == nil {
			out.Margins = &MarginsConfig{}
		}
		mergeMargins(out.Margins, override.Margins)
	}
	if override.Spread != nil {
		if out.Spread == nil {
			out.Spread = &SpreadConfig{}
		}
		mergeSpread(out.Spread, override.Spread)
	}
	if override.Crop != nil {
		if out.Crop == nil {
			out.Crop = &CropConfig{}
		}
		mergeCrop(out.Crop, override.Crop)
	}
	if override.Scale != nil {
		if out.Scale == nil {
			out.Scale = &ScaleConfig{}
		}
		mergeScale(out.Scale, override.Scale)
	}
	if override.Position != nil {
		if out.Position == nil {
			out.Position = &PositionConfig{}
		}
		mergePosition(out.Position, override.Position)
	}
	if override.Export != nil {
		if out.Export == nil {
			out.Export = &ExportConfig{}
		}
		mergeExport(out.Export, override.Export)
	}
	return out
}

func mergePaper(dst, src *PaperConfig) {
	if src.Size != "" {
		dst.Size = src.Size
	}
	if src.WidthIn != nil {
		v := *src.WidthIn
		dst.WidthIn = &v
	}
	if src.HeightIn != nil {
		v := *src.HeightIn
		dst.HeightIn = &v
	}
	if src.Orientation != "" {
		dst.Orientation = src.Orientation
	}
	if src.DPI != nil {
		v := *src.DPI
		dst.DPI = &v
	}
}

func mergeMargins(dst, src *MarginsConfig) {
	if src.AllIn != nil {
		v := *src.AllIn
		dst.AllIn = &v
	}
	if src.TopIn != nil {
		v := *src.TopIn
		dst.TopIn = &v
	}
	if src.RightIn != nil {
		v := *src.RightIn
		dst.RightIn = &v
	}
	if src.BottomIn != nil {
		v := *src.BottomIn
		dst.BottomIn = &v
	}
	if src.LeftIn != nil {
		v := *src.LeftIn
		dst.LeftIn = &v
	}
}

func mergeSpread(dst, src *SpreadConfig) {
	if src.IsSpread != nil {
		v := *src.IsSpread
		dst.IsSpread = &v
	}
	if src.GutterIn != nil {
		v := *src.GutterIn
		dst.GutterIn = &v
	}
}

func mergeCrop(dst, src *CropConfig) {
	if src.Ratio != nil {
		dst.Ratio = src.Ratio
	}
	if src.ToFill != nil {
		v := *src.ToFill
		dst.ToFill = &v
	}
}

func mergeScale(dst, src *ScaleConfig) {
	if src.UserScale != nil {
		v := *src.UserScale
		dst.UserScale = &v
	}
}

func mergePosition(dst, src *PositionConfig) {
	if src.X != nil {
		v := *src.X
		dst.X = &v
	}
	if src.Y != nil {
		v := *src.Y
		dst.Y = &v
	}
	if src.Units != "" {
		dst.Units = src.Units
	}
}
func (s SpreadSettings) Clone() SpreadSettings {
	out := SpreadSettings{}
	if s.Paper != nil {
		c := *s.Paper
		out.Paper = &c
	}
	if s.Margins != nil {
		c := *s.Margins
		out.Margins = &c
	}
	if s.Spread != nil {
		c := *s.Spread
		out.Spread = &c
	}
	if s.Crop != nil {
		c := *s.Crop
		out.Crop = &c
	}
	if s.Scale != nil {
		c := *s.Scale
		out.Scale = &c
	}
	if s.Position != nil {
		c := *s.Position
		out.Position = &c
	}
	if s.Export != nil {
		c := *s.Export
		out.Export = &c
	}
	return out
}

func (s SpreadEntry) AsSettings() SpreadSettings {
	out := SpreadSettings{}
	if s.Paper != nil {
		c := *s.Paper
		out.Paper = &c
	}
	if s.Margins != nil {
		c := *s.Margins
		out.Margins = &c
	}
	if s.Spread != nil {
		c := *s.Spread
		out.Spread = &c
	}
	if s.Crop != nil {
		c := *s.Crop
		out.Crop = &c
	}
	if s.Scale != nil {
		c := *s.Scale
		out.Scale = &c
	}
	if s.Position != nil {
		c := *s.Position
		out.Position = &c
	}
	if s.Export != nil {
		c := *s.Export
		out.Export = &c
	}
	return out
}

func mergeExport(dst, src *ExportConfig) {
	if src.Format != "" {
		dst.Format = strings.ToLower(src.Format)
	}
	if src.Quality != nil {
		v := *src.Quality
		dst.Quality = &v
	}
	if src.Background != "" {
		dst.Background = src.Background
	}
	if src.OutDir != "" {
		dst.OutDir = src.OutDir
	}
	if src.FilenameTemplate != "" {
		dst.FilenameTemplate = src.FilenameTemplate
	}
}

// RuntimeAssets holds resolved asset file lists.
type RuntimeAssets map[string][]string

// ResolveAssets expands asset definitions on disk.
func (c *Config) ResolveAssets(baseDir string) (RuntimeAssets, error) {
	result := make(RuntimeAssets, len(c.Assets))
	for _, asset := range c.Assets {
		files, err := resolveAsset(baseDir, asset)
		if err != nil {
			return nil, err
		}
		result[asset.Name] = files
	}
	return result, nil
}

func resolveAsset(baseDir string, asset AssetConfig) ([]string, error) {
	root := filepath.Clean(filepath.Join(baseDir, asset.Path))
	matches, err := globPaths(root)
	if err != nil {
		return nil, fmt.Errorf("asset %s: %w", asset.Name, err)
	}
	var files []string
	for _, m := range matches {
		info, err := fileInfo(m)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err = filepath.WalkDir(m, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(m, path)
				if !shouldInclude(rel, asset.Include, asset.Exclude) {
					return nil
				}
				files = append(files, path)
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		if !shouldInclude(filepath.Base(m), asset.Include, asset.Exclude) {
			continue
		}
		files = append(files, m)
	}
	sort.Strings(files)
	return files, nil
}

// Additional helpers using fs package

func globPaths(path string) ([]string, error) {
	if hasMeta(path) {
		matches, err := doublestar.FilepathGlob(path)
		if err != nil {
			return nil, err
		}
		return matches, nil
	}
	return []string{path}, nil
}

func hasMeta(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func fileInfo(path string) (fs.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func shouldInclude(rel string, include, exclude []string) bool {
	rel = filepath.ToSlash(rel)
	if len(include) > 0 {
		ok := false
		for _, pattern := range include {
			if pattern == "" {
				continue
			}
			match, err := doublestar.Match(pattern, rel)
			if err == nil && match {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	for _, pattern := range exclude {
		if pattern == "" {
			continue
		}
		match, err := doublestar.Match(pattern, rel)
		if err == nil && match {
			return false
		}
	}
	return true
}

// ImageItem is a resolved image reference.
type ImageItem struct {
	Path  string
	Asset string
}

// ResolveSelector expands an image selector into file paths.
func (c *Config) ResolveSelector(baseDir string, selector ImageSelector, assets RuntimeAssets) ([]ImageItem, error) {
	if selector.Asset != "" {
		files, ok := assets[selector.Asset]
		if !ok {
			return nil, fmt.Errorf("unknown asset %s", selector.Asset)
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("asset %s has no files", selector.Asset)
		}
		items := make([]ImageItem, len(files))
		for i, path := range files {
			items[i] = ImageItem{Path: path, Asset: selector.Asset}
		}
		return items, nil
	}
	var results []ImageItem
	for _, p := range selector.Paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		abs := filepath.Clean(filepath.Join(baseDir, p))
		matches, err := globPaths(abs)
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			results = append(results, ImageItem{Path: m})
		}
	}
	if len(results) == 0 {
		return nil, errors.New("selector resolved to no files")
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Path < results[j].Path })
	return results, nil
}
