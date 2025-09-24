package spread

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// yamlConfig mirrors the Sonnet configuration schema for emitting copyable YAML documents.
type yamlConfig struct {
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
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

// RequestToSonnetYAML converts a compute request into a canonical Sonnet YAML snippet.
func RequestToSonnetYAML(req ComputeRequest) (string, error) {
	if req.ImagePath == "" {
		return "", fmt.Errorf("image_path required to build YAML")
	}

	cfg := yamlConfig{
		Version: "0.1",
		Defaults: yamlDefaults{
			Paper: yamlPaper{
				WidthIn:     req.Settings.PaperWidthIn,
				HeightIn:    req.Settings.PaperHeightIn,
				Orientation: req.Settings.Orientation,
				DPI:         req.Settings.DPI,
			},
			Margins: yamlMargins{
				TopIn:    req.Settings.MarginTopIn,
				RightIn:  req.Settings.MarginRightIn,
				BottomIn: req.Settings.MarginBottomIn,
				LeftIn:   req.Settings.MarginLeftIn,
			},
			Spread: yamlSpreadOpt{
				IsSpread: req.Settings.IsSpread,
				GutterIn: req.Settings.GutterIn,
			},
			Crop: yamlCrop{
				Ratio:  cropRatioValue(req.Settings.CropRatio),
				ToFill: req.Settings.CropToFill,
			},
			Scale: yamlScale{
				UserScale: req.Settings.UserScale,
			},
			Position: yamlPosition{
				X:     req.Settings.PositionX,
				Y:     req.Settings.PositionY,
				Units: req.Settings.Units,
			},
			Export: yamlExport{
				Format:           req.Settings.Export.Format,
				Quality:          req.Settings.Export.Quality,
				Background:       req.Settings.Export.Background,
				OutDir:           req.Settings.Export.OutDir,
				FilenameTemplate: req.Settings.Export.FilenameTemplate,
			},
		},
		Spreads: []yamlSpread{
			{
				Name:  defaultName(req.Name),
				Image: req.ImagePath,
			},
		},
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(out), nil
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

func defaultName(name string) string {
	if name == "" {
		return "preview"
	}
	return name
}
