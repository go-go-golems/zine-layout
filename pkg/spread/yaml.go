package spread

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const simpleYAMLVersion = "0.2"

type simpleYAMLDocument struct {
	Version  string       `yaml:"version"`
	Defaults yamlDefaults `yaml:"defaults"`
	Spreads  []yamlSpread `yaml:"spreads"`
}

type yamlDefaults struct {
	Paper    yamlPaper     `yaml:"paper"`
	Margins  yamlMargins   `yaml:"margins"`
	Spread   yamlSpreadOpt `yaml:"spread"`
	Crop     yamlCrop      `yaml:"crop"`
	Scale    yamlScale     `yaml:"scale"`
	Position yamlPosition  `yaml:"position"`
	Export   yamlExport    `yaml:"export"`
}

type yamlPaper struct {
	WidthIn     float64 `yaml:"width_in"`
	HeightIn    float64 `yaml:"height_in"`
	Orientation string  `yaml:"orientation"`
	DPI         float64 `yaml:"dpi"`
}

type yamlMargins struct {
	TopIn    float64 `yaml:"top_in"`
	RightIn  float64 `yaml:"right_in"`
	BottomIn float64 `yaml:"bottom_in"`
	LeftIn   float64 `yaml:"left_in"`
}

type yamlSpreadOpt struct {
	IsSpread bool    `yaml:"is_spread"`
	GutterIn float64 `yaml:"gutter_in"`
}

type yamlCrop struct {
	Ratio  any  `yaml:"ratio"`
	ToFill bool `yaml:"to_fill"`
}

type yamlScale struct {
	UserScale float64 `yaml:"user_scale"`
}

type yamlPosition struct {
	X     float64 `yaml:"x"`
	Y     float64 `yaml:"y"`
	Units string  `yaml:"units"`
}

type yamlExport struct {
	Format           string `yaml:"format"`
	Quality          int    `yaml:"quality"`
	Background       string `yaml:"background"`
	OutDir           string `yaml:"out_dir"`
	FilenameTemplate string `yaml:"filename_template"`
}

type yamlSpread struct {
	Name     string         `yaml:"name"`
	Image    string         `yaml:"image"`
	Paper    *yamlPaper     `yaml:"paper,omitempty"`
	Margins  *yamlMargins   `yaml:"margins,omitempty"`
	Spread   *yamlSpreadOpt `yaml:"spread,omitempty"`
	Crop     *yamlCrop      `yaml:"crop,omitempty"`
	Scale    *yamlScale     `yaml:"scale,omitempty"`
	Position *yamlPosition  `yaml:"position,omitempty"`
	Export   *yamlExport    `yaml:"export,omitempty"`
}

// SimpleSpread describes a single spread entry resolved from YAML.
type SimpleSpread struct {
	Name      string
	ImagePath string
	Settings  Settings
}

// SimpleBookDocument represents a parsed Simple YAML document.
type SimpleBookDocument struct {
	Version  string
	Defaults Settings
	Spreads  []SimpleSpread
}

// BookSpreadItem describes a spread to emit when building a book YAML document.
type BookSpreadItem struct {
	Name      string
	ImagePath string
	Settings  *Settings
}

// BuildSimpleYAML converts a compute request into a Simple YAML snippet.
func BuildSimpleYAML(req ComputeRequest) (string, error) {
	imagePath := strings.TrimSpace(req.ImagePath)
	if imagePath == "" {
		return "", fmt.Errorf("image_path required to build YAML")
	}
	doc := simpleYAMLDocument{
		Version:  simpleYAMLVersion,
		Defaults: settingsToYAML(req.Settings),
		Spreads: []yamlSpread{
			{
				Name:  defaultPreviewName(req.Name),
				Image: filepath.ToSlash(imagePath),
			},
		},
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// BuildBookYAML emits a Simple YAML document describing an entire book.
func BuildBookYAML(defaults Settings, spreads []BookSpreadItem, baseDir string) (string, error) {
	doc := simpleYAMLDocument{
		Version:  simpleYAMLVersion,
		Defaults: settingsToYAML(defaults),
		Spreads:  make([]yamlSpread, 0, len(spreads)),
	}
	for idx, item := range spreads {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = fmt.Sprintf("page-%04d", idx+1)
		}
		entry := yamlSpread{
			Name:  name,
			Image: normalizeOutputImagePath(item.ImagePath, baseDir),
		}
		if item.Settings != nil {
			entry.Paper, entry.Margins, entry.Spread, entry.Crop, entry.Scale, entry.Position, entry.Export = buildOverrides(defaults, *item.Settings)
		}
		doc.Spreads = append(doc.Spreads, entry)
	}
	if len(doc.Spreads) == 0 {
		return "", fmt.Errorf("no spreads to export")
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ParseSimpleBookYAML parses Simple YAML and resolves spreads relative to defaults.
func ParseSimpleBookYAML(raw string) (*SimpleBookDocument, error) {
	payload := strings.TrimSpace(raw)
	if payload == "" {
		return nil, fmt.Errorf("yaml is empty")
	}
	var doc simpleYAMLDocument
	if err := yaml.Unmarshal([]byte(payload), &doc); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if strings.TrimSpace(doc.Version) == "" {
		doc.Version = simpleYAMLVersion
	}
	defaults, err := yamlToSettings(doc.Defaults)
	if err != nil {
		return nil, fmt.Errorf("defaults: %w", err)
	}
	spreads := make([]SimpleSpread, 0, len(doc.Spreads))
	for idx, entry := range doc.Spreads {
		imagePath := strings.TrimSpace(entry.Image)
		if imagePath == "" {
			return nil, fmt.Errorf("spread %d missing image path", idx+1)
		}
		settings, err := applyOverrides(defaults, entry)
		if err != nil {
			return nil, fmt.Errorf("spread %d: %w", idx+1, err)
		}
		name := strings.TrimSpace(entry.Name)
		if name == "" {
			name = fmt.Sprintf("spread-%03d", idx+1)
		}
		spreads = append(spreads, SimpleSpread{
			Name:      name,
			ImagePath: entry.Image,
			Settings:  settings,
		})
	}
	return &SimpleBookDocument{Version: doc.Version, Defaults: defaults, Spreads: spreads}, nil
}

func settingsToYAML(settings Settings) yamlDefaults {
	return yamlDefaults{
		Paper: yamlPaper{
			WidthIn:     settings.PaperWidthIn,
			HeightIn:    settings.PaperHeightIn,
			Orientation: settings.Orientation,
			DPI:         settings.DPI,
		},
		Margins: yamlMargins{
			TopIn:    settings.MarginTopIn,
			RightIn:  settings.MarginRightIn,
			BottomIn: settings.MarginBottomIn,
			LeftIn:   settings.MarginLeftIn,
		},
		Spread: yamlSpreadOpt{
			IsSpread: settings.IsSpread,
			GutterIn: settings.GutterIn,
		},
		Crop: yamlCrop{
			Ratio:  cropRatioValue(settings.CropRatio),
			ToFill: settings.CropToFill,
		},
		Scale: yamlScale{
			UserScale: settings.UserScale,
		},
		Position: yamlPosition{
			X:     settings.PositionX,
			Y:     settings.PositionY,
			Units: settings.Units,
		},
		Export: yamlExport{
			Format:           settings.Export.Format,
			Quality:          settings.Export.Quality,
			Background:       settings.Export.Background,
			OutDir:           settings.Export.OutDir,
			FilenameTemplate: settings.Export.FilenameTemplate,
		},
	}
}

func yamlToSettings(cfg yamlDefaults) (Settings, error) {
	ratio, err := parseCropRatio(cfg.Crop.Ratio)
	if err != nil {
		return Settings{}, err
	}
	settings := Settings{
		PaperWidthIn:   cfg.Paper.WidthIn,
		PaperHeightIn:  cfg.Paper.HeightIn,
		Orientation:    cfg.Paper.Orientation,
		DPI:            cfg.Paper.DPI,
		MarginTopIn:    cfg.Margins.TopIn,
		MarginRightIn:  cfg.Margins.RightIn,
		MarginBottomIn: cfg.Margins.BottomIn,
		MarginLeftIn:   cfg.Margins.LeftIn,
		IsSpread:       cfg.Spread.IsSpread,
		GutterIn:       cfg.Spread.GutterIn,
		CropRatio:      ratio,
		CropToFill:     cfg.Crop.ToFill,
		UserScale:      cfg.Scale.UserScale,
		PositionX:      cfg.Position.X,
		PositionY:      cfg.Position.Y,
		Units:          cfg.Position.Units,
		Export: Export{
			Format:           cfg.Export.Format,
			Quality:          cfg.Export.Quality,
			Background:       cfg.Export.Background,
			OutDir:           cfg.Export.OutDir,
			FilenameTemplate: cfg.Export.FilenameTemplate,
		},
	}
	return settings, nil
}

func buildOverrides(defaults Settings, override Settings) (*yamlPaper, *yamlMargins, *yamlSpreadOpt, *yamlCrop, *yamlScale, *yamlPosition, *yamlExport) {
	var paper *yamlPaper
	if !floatsEqual(defaults.PaperWidthIn, override.PaperWidthIn) || !floatsEqual(defaults.PaperHeightIn, override.PaperHeightIn) || !strings.EqualFold(strings.TrimSpace(defaults.Orientation), strings.TrimSpace(override.Orientation)) || !floatsEqual(defaults.DPI, override.DPI) {
		paper = &yamlPaper{WidthIn: override.PaperWidthIn, HeightIn: override.PaperHeightIn, Orientation: override.Orientation, DPI: override.DPI}
	}

	var margins *yamlMargins
	if !floatsEqual(defaults.MarginTopIn, override.MarginTopIn) || !floatsEqual(defaults.MarginRightIn, override.MarginRightIn) || !floatsEqual(defaults.MarginBottomIn, override.MarginBottomIn) || !floatsEqual(defaults.MarginLeftIn, override.MarginLeftIn) {
		margins = &yamlMargins{TopIn: override.MarginTopIn, RightIn: override.MarginRightIn, BottomIn: override.MarginBottomIn, LeftIn: override.MarginLeftIn}
	}

	var spreadOpt *yamlSpreadOpt
	if defaults.IsSpread != override.IsSpread || !floatsEqual(defaults.GutterIn, override.GutterIn) {
		spreadOpt = &yamlSpreadOpt{IsSpread: override.IsSpread, GutterIn: override.GutterIn}
	}

	var crop *yamlCrop
	if !cropEqual(defaults.CropRatio, override.CropRatio) || defaults.CropToFill != override.CropToFill {
		crop = &yamlCrop{Ratio: cropRatioValue(override.CropRatio), ToFill: override.CropToFill}
	}

	var scale *yamlScale
	if !floatsEqual(defaults.UserScale, override.UserScale) {
		scale = &yamlScale{UserScale: override.UserScale}
	}

	var position *yamlPosition
	if !floatsEqual(defaults.PositionX, override.PositionX) || !floatsEqual(defaults.PositionY, override.PositionY) || strings.TrimSpace(defaults.Units) != strings.TrimSpace(override.Units) {
		position = &yamlPosition{X: override.PositionX, Y: override.PositionY, Units: override.Units}
	}

	var export *yamlExport
	if exportDiffers(defaults.Export, override.Export) {
		export = &yamlExport{
			Format:           override.Export.Format,
			Quality:          override.Export.Quality,
			Background:       override.Export.Background,
			OutDir:           override.Export.OutDir,
			FilenameTemplate: override.Export.FilenameTemplate,
		}
	}
	return paper, margins, spreadOpt, crop, scale, position, export
}

func applyOverrides(defaults Settings, entry yamlSpread) (Settings, error) {
	out := cloneSettings(defaults)
	if entry.Paper != nil {
		out.PaperWidthIn = entry.Paper.WidthIn
		out.PaperHeightIn = entry.Paper.HeightIn
		out.Orientation = entry.Paper.Orientation
		out.DPI = entry.Paper.DPI
	}
	if entry.Margins != nil {
		out.MarginTopIn = entry.Margins.TopIn
		out.MarginRightIn = entry.Margins.RightIn
		out.MarginBottomIn = entry.Margins.BottomIn
		out.MarginLeftIn = entry.Margins.LeftIn
	}
	if entry.Spread != nil {
		out.IsSpread = entry.Spread.IsSpread
		out.GutterIn = entry.Spread.GutterIn
	}
	if entry.Crop != nil {
		ratio, err := parseCropRatio(entry.Crop.Ratio)
		if err != nil {
			return out, err
		}
		out.CropRatio = ratio
		out.CropToFill = entry.Crop.ToFill
	}
	if entry.Scale != nil {
		out.UserScale = entry.Scale.UserScale
	}
	if entry.Position != nil {
		out.PositionX = entry.Position.X
		out.PositionY = entry.Position.Y
		if strings.TrimSpace(entry.Position.Units) != "" {
			out.Units = entry.Position.Units
		}
	}
	if entry.Export != nil {
		out.Export.Format = entry.Export.Format
		out.Export.Quality = entry.Export.Quality
		out.Export.Background = entry.Export.Background
		out.Export.OutDir = entry.Export.OutDir
		out.Export.FilenameTemplate = entry.Export.FilenameTemplate
	}
	return out, nil
}

func cloneSettings(src Settings) Settings {
	dst := src
	if src.CropRatio != nil {
		v := *src.CropRatio
		dst.CropRatio = &v
	}
	return dst
}

func parseCropRatio(v any) (*float64, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case float64:
		if t <= 0 {
			return nil, fmt.Errorf("crop ratio must be > 0")
		}
		val := t
		return &val, nil
	case int:
		if t <= 0 {
			return nil, fmt.Errorf("crop ratio must be > 0")
		}
		val := float64(t)
		return &val, nil
	case int64:
		if t <= 0 {
			return nil, fmt.Errorf("crop ratio must be > 0")
		}
		val := float64(t)
		return &val, nil
	case string:
		trimmed := strings.TrimSpace(strings.ToLower(t))
		if trimmed == "" || trimmed == "original" {
			return nil, nil
		}
		parsed, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid crop ratio %q", t)
		}
		if parsed <= 0 {
			return nil, fmt.Errorf("crop ratio must be > 0")
		}
		return &parsed, nil
	default:
		return nil, fmt.Errorf("unsupported crop ratio type %T", v)
	}
}

func cropRatioValue(ratio *float64) any {
	if ratio == nil {
		return "original"
	}
	if *ratio == 0 {
		return "original"
	}
	return *ratio
}

func cropEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return floatsEqual(*a, *b)
}

func floatsEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func exportDiffers(a, b Export) bool {
	if strings.TrimSpace(a.Format) != strings.TrimSpace(b.Format) {
		return true
	}
	if a.Quality != b.Quality {
		return true
	}
	if strings.TrimSpace(a.Background) != strings.TrimSpace(b.Background) {
		return true
	}
	if strings.TrimSpace(a.OutDir) != strings.TrimSpace(b.OutDir) {
		return true
	}
	if strings.TrimSpace(a.FilenameTemplate) != strings.TrimSpace(b.FilenameTemplate) {
		return true
	}
	return false
}

func defaultPreviewName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "preview"
	}
	return name
}

func normalizeOutputImagePath(path string, baseDir string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return trimmed
	}
	if baseDir != "" && filepath.IsAbs(trimmed) {
		if rel, err := filepath.Rel(baseDir, trimmed); err == nil {
			trimmed = rel
		}
	}
	return filepath.ToSlash(trimmed)
}
