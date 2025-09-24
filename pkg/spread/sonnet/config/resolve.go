package config

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type SpreadSpec struct {
	Name      string
	ImagePath string
	AssetName string
	Settings  ResolvedSettings
}

type ResolvedSettings struct {
	PaperWidthIn  float64
	PaperHeightIn float64
	DPI           float64
	Orientation   string
	Margins       MarginValues
	IsSpread      bool
	GutterIn      float64
	CropRatio     *float64
	CropToFill    bool
	UserScale     float64
	Position      PositionValues
	Export        ExportValues
}

type MarginValues struct {
	TopIn    float64
	RightIn  float64
	BottomIn float64
	LeftIn   float64
}

type PositionValues struct {
	X     float64
	Y     float64
	Units string
}

type ExportValues struct {
	Format           string
	Quality          int
	Background       string
	OutDir           string
	FilenameTemplate string
}

var (
	paperSizes = map[string][2]float64{
		"a4":      {8.27, 11.69},
		"letter":  {8.5, 11},
		"legal":   {8.5, 14},
		"tabloid": {11, 17},
		"a3":      {11.69, 16.54},
	}
	filenameTemplateDefault = "{index:03d}-{name}-{panel}.{ext}"
)

// Resolve expands spreads into concrete specs for downstream processing.
func (c *Config) Resolve(baseDir string) ([]SpreadSpec, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	assets, err := c.ResolveAssets(baseDir)
	if err != nil {
		return nil, err
	}

	defaults := c.Defaults.Clone()
	var specs []SpreadSpec

	for _, spreadCfg := range c.Spreads {
		merged := defaults.Clone().Merge(spreadCfg.AsSettings())
		resolved, err := compileSettings(merged)
		if err != nil {
			return nil, fmt.Errorf("spread %s: %w", spreadCfg.Name, err)
		}

		if spreadCfg.ForEach != nil {
			items, err := c.ResolveSelector(baseDir, spreadCfg.ForEach.Images, assets)
			if err != nil {
				return nil, fmt.Errorf("spread %s: %w", spreadCfg.Name, err)
			}
			nameTemplate := strings.TrimSpace(spreadCfg.ForEach.NameTemplate)
			if nameTemplate == "" {
				nameTemplate = spreadCfg.Name + "-{index}"
			}
			for idx, item := range items {
				name := renderNameTemplate(nameTemplate, spreadCfg.Name, item, idx)
				specs = append(specs, SpreadSpec{
					Name:      name,
					ImagePath: item.Path,
					AssetName: item.Asset,
					Settings:  resolved,
				})
			}
			continue
		}

		if spreadCfg.Image.IsEmpty() {
			return nil, fmt.Errorf("spread %s is missing image", spreadCfg.Name)
		}
		items, err := c.ResolveSelector(baseDir, spreadCfg.Image, assets)
		if err != nil {
			return nil, fmt.Errorf("spread %s: %w", spreadCfg.Name, err)
		}
		if len(items) == 1 {
			specs = append(specs, SpreadSpec{
				Name:      spreadCfg.Name,
				ImagePath: items[0].Path,
				AssetName: items[0].Asset,
				Settings:  resolved,
			})
			continue
		}
		for idx, item := range items {
			name := fmt.Sprintf("%s-%d", spreadCfg.Name, idx+1)
			specs = append(specs, SpreadSpec{
				Name:      name,
				ImagePath: item.Path,
				AssetName: item.Asset,
				Settings:  resolved,
			})
		}
	}

	return specs, nil
}

type paperResolved struct {
	widthIn     float64
	heightIn    float64
	dpi         float64
	orientation string
}

func compileSettings(settings SpreadSettings) (ResolvedSettings, error) {
	paperVals, err := resolvePaper(settings.Paper)
	if err != nil {
		return ResolvedSettings{}, err
	}
	margins, err := resolveMargins(settings.Margins)
	if err != nil {
		return ResolvedSettings{}, err
	}
	spreadVals, err := resolveSpread(settings.Spread)
	if err != nil {
		return ResolvedSettings{}, err
	}
	cropVals, err := resolveCrop(settings.Crop)
	if err != nil {
		return ResolvedSettings{}, err
	}
	scaleVal, err := resolveScale(settings.Scale)
	if err != nil {
		return ResolvedSettings{}, err
	}
	positionVals, err := resolvePosition(settings.Position)
	if err != nil {
		return ResolvedSettings{}, err
	}
	exportVals, err := resolveExport(settings.Export)
	if err != nil {
		return ResolvedSettings{}, err
	}

	return ResolvedSettings{
		PaperWidthIn:  paperVals.widthIn,
		PaperHeightIn: paperVals.heightIn,
		DPI:           paperVals.dpi,
		Orientation:   paperVals.orientation,
		Margins:       margins,
		IsSpread:      spreadVals.isSpread,
		GutterIn:      spreadVals.gutterIn,
		CropRatio:     cropVals.ratio,
		CropToFill:    cropVals.toFill,
		UserScale:     scaleVal,
		Position:      positionVals,
		Export:        exportVals,
	}, nil
}

func resolvePaper(p *PaperConfig) (paperResolved, error) {
	if p == nil {
		return paperResolved{}, fmt.Errorf("paper configuration missing")
	}

	orientation := strings.ToLower(strings.TrimSpace(p.Orientation))
	if orientation == "" {
		orientation = "portrait"
	}

	var width, height float64
	if size := strings.ToLower(strings.TrimSpace(p.Size)); size != "" {
		if dims, ok := paperSizes[size]; ok {
			width, height = dims[0], dims[1]
		} else {
			return paperResolved{}, fmt.Errorf("unknown paper size %q", p.Size)
		}
	}
	if p.WidthIn != nil {
		width = *p.WidthIn
	}
	if p.HeightIn != nil {
		height = *p.HeightIn
	}
	if width <= 0 || height <= 0 {
		return paperResolved{}, fmt.Errorf("paper width_in and height_in must be positive")
	}

	if orientation == "landscape" {
		if height > width {
			width, height = height, width
		}
	} else if orientation == "portrait" {
		if width > height {
			width, height = height, width
		}
	} else {
		return paperResolved{}, fmt.Errorf("invalid orientation %q", p.Orientation)
	}

	if p.DPI == nil || *p.DPI <= 0 {
		return paperResolved{}, fmt.Errorf("paper.dpi must be > 0")
	}

	return paperResolved{
		widthIn:     width,
		heightIn:    height,
		dpi:         *p.DPI,
		orientation: orientation,
	}, nil
}

func resolveMargins(m *MarginsConfig) (MarginValues, error) {
	out := MarginValues{}
	if m == nil {
		return out, nil
	}

	if m.AllIn != nil {
		if *m.AllIn < 0 {
			return MarginValues{}, fmt.Errorf("margins.all_in must be >= 0")
		}
		out.TopIn, out.RightIn, out.BottomIn, out.LeftIn = *m.AllIn, *m.AllIn, *m.AllIn, *m.AllIn
	}
	if m.TopIn != nil {
		if *m.TopIn < 0 {
			return MarginValues{}, fmt.Errorf("margins.top_in must be >= 0")
		}
		out.TopIn = *m.TopIn
	}
	if m.RightIn != nil {
		if *m.RightIn < 0 {
			return MarginValues{}, fmt.Errorf("margins.right_in must be >= 0")
		}
		out.RightIn = *m.RightIn
	}
	if m.BottomIn != nil {
		if *m.BottomIn < 0 {
			return MarginValues{}, fmt.Errorf("margins.bottom_in must be >= 0")
		}
		out.BottomIn = *m.BottomIn
	}
	if m.LeftIn != nil {
		if *m.LeftIn < 0 {
			return MarginValues{}, fmt.Errorf("margins.left_in must be >= 0")
		}
		out.LeftIn = *m.LeftIn
	}
	return out, nil
}

func resolveSpread(s *SpreadConfig) (struct {
	isSpread bool
	gutterIn float64
}, error) {
	isSpread := false
	if s != nil && s.IsSpread != nil {
		isSpread = *s.IsSpread
	}
	gutter := 0.0
	if s != nil && s.GutterIn != nil {
		gutter = *s.GutterIn
		if gutter < 0 {
			return struct {
				isSpread bool
				gutterIn float64
			}{}, fmt.Errorf("spread.gutter_in must be >= 0")
		}
	}
	return struct {
		isSpread bool
		gutterIn float64
	}{isSpread: isSpread, gutterIn: gutter}, nil
}

type cropResolved struct {
	ratio  *float64
	toFill bool
}

var ratioPattern = regexp.MustCompile(`^(\d+(?:\.\d+)?)[\s]*[:\/][\s]*(\d+(?:\.\d+)?)$`)

func resolveCrop(c *CropConfig) (cropResolved, error) {
	out := cropResolved{}
	if c == nil {
		return out, nil
	}
	if c.ToFill != nil {
		out.toFill = *c.ToFill
	}
	if c.Ratio == nil {
		return out, nil
	}
	switch v := c.Ratio.(type) {
	case string:
		val := strings.TrimSpace(strings.ToLower(v))
		if val == "" || val == "original" || val == "null" {
			return out, nil
		}
		if m := ratioPattern.FindStringSubmatch(val); m != nil {
			w, _ := strconv.ParseFloat(m[1], 64)
			h, _ := strconv.ParseFloat(m[2], 64)
			if w <= 0 || h <= 0 {
				return cropResolved{}, fmt.Errorf("crop.ratio components must be positive")
			}
			r := w / h
			out.ratio = &r
			return out, nil
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil || !isFinitePositive(f) {
			return cropResolved{}, fmt.Errorf("invalid crop.ratio %q", v)
		}
		r := f
		out.ratio = &r
		return out, nil
	case float64:
		if !isFinitePositive(v) {
			return cropResolved{}, fmt.Errorf("crop.ratio must be > 0")
		}
		r := v
		out.ratio = &r
		return out, nil
	case int:
		if v <= 0 {
			return cropResolved{}, fmt.Errorf("crop.ratio must be > 0")
		}
		r := float64(v)
		out.ratio = &r
		return out, nil
	case int64:
		if v <= 0 {
			return cropResolved{}, fmt.Errorf("crop.ratio must be > 0")
		}
		r := float64(v)
		out.ratio = &r
		return out, nil
	default:
		return cropResolved{}, fmt.Errorf("unsupported crop.ratio type %T", v)
	}
}

func resolveScale(s *ScaleConfig) (float64, error) {
	if s == nil || s.UserScale == nil {
		return 1.0, nil
	}
	if *s.UserScale < 0 {
		return 0, fmt.Errorf("scale.user_scale must be >= 0")
	}
	return *s.UserScale, nil
}

func resolvePosition(p *PositionConfig) (PositionValues, error) {
	units := "normalized"
	if p != nil && strings.TrimSpace(p.Units) != "" {
		units = strings.ToLower(strings.TrimSpace(p.Units))
	}
	if units != "normalized" && units != "px" {
		return PositionValues{}, fmt.Errorf("position.units must be 'px' or 'normalized'")
	}
	x := 0.0
	y := 0.0
	if p != nil && p.X != nil {
		x = *p.X
	}
	if p != nil && p.Y != nil {
		y = *p.Y
	}
	return PositionValues{X: x, Y: y, Units: units}, nil
}

func resolveExport(e *ExportConfig) (ExportValues, error) {
	format := "png"
	if e != nil && strings.TrimSpace(e.Format) != "" {
		format = strings.ToLower(strings.TrimSpace(e.Format))
	}
	switch format {
	case "png", "jpg", "jpeg", "pdf":
	default:
		return ExportValues{}, fmt.Errorf("unsupported export.format %q", format)
	}

	quality := 90
	if e != nil && e.Quality != nil {
		quality = *e.Quality
	}
	if (format == "jpg" || format == "jpeg") && (quality < 1 || quality > 100) {
		return ExportValues{}, fmt.Errorf("export.quality must be between 1 and 100 for jpg")
	}
	if format == "png" || format == "pdf" {
		quality = clamp(quality, 1, 100)
	}

	background := "transparent"
	if e != nil && strings.TrimSpace(e.Background) != "" {
		background = strings.TrimSpace(e.Background)
	}

	outDir := "./out"
	if e != nil && strings.TrimSpace(e.OutDir) != "" {
		outDir = e.OutDir
	}
	outDir = filepath.Clean(outDir)

	filenameTemplate := filenameTemplateDefault
	if e != nil && strings.TrimSpace(e.FilenameTemplate) != "" {
		filenameTemplate = e.FilenameTemplate
	}

	canonicalFormat := format
	if canonicalFormat == "jpeg" {
		canonicalFormat = "jpg"
	}

	return ExportValues{
		Format:           canonicalFormat,
		Quality:          quality,
		Background:       background,
		OutDir:           outDir,
		FilenameTemplate: filenameTemplate,
	}, nil
}

func renderNameTemplate(template, defaultName string, item ImageItem, index int) string {
	base := strings.TrimSpace(template)
	if base == "" {
		return fmt.Sprintf("%s-%d", defaultName, index+1)
	}
	basename := filepath.Base(item.Path)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))

	replacements := map[string]string{
		"{image_basename}": basename,
		"{asset}":          item.Asset,
		"{index}":          fmt.Sprintf("%d", index+1),
		"{name}":           defaultName,
	}

	keys := make([]string, 0, len(replacements))
	for k := range replacements {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := base
	for _, k := range keys {
		result = strings.ReplaceAll(result, k, replacements[k])
	}
	return strings.TrimSpace(result)
}

func isFinitePositive(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > 0
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
