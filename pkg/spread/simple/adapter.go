package simple

import (
	"fmt"

	sonnetcfg "github.com/go-go-golems/zine-layout/pkg/spread/sonnet/config"
	sonneteng "github.com/go-go-golems/zine-layout/pkg/spread/sonnet/engine"
)

// InputsFromResolved converts sonnet resolved settings into simple algorithm inputs.
func InputsFromResolved(settings sonnetcfg.ResolvedSettings, meta sonneteng.SourceMeta) (Inputs, error) {
	if meta.Width <= 0 || meta.Height <= 0 {
		return Inputs{}, fmt.Errorf("invalid source dimensions %dx%d", meta.Width, meta.Height)
	}

	var crop *CropRatio
	if settings.CropRatio != nil {
		r := *settings.CropRatio
		if r <= 0 {
			return Inputs{}, fmt.Errorf("crop ratio must be > 0")
		}
		crop = &CropRatio{W: r, H: 1}
	}

	inp := Inputs{
		SrcW:           float64(meta.Width),
		SrcH:           float64(meta.Height),
		PaperWIn:       settings.PaperWidthIn,
		PaperHIn:       settings.PaperHeightIn,
		Orientation:    settings.Orientation,
		MarginTopIn:    settings.Margins.TopIn,
		MarginRightIn:  settings.Margins.RightIn,
		MarginBottomIn: settings.Margins.BottomIn,
		MarginLeftIn:   settings.Margins.LeftIn,
		DPI:            settings.DPI,
		IsSpread:       settings.IsSpread,
		GutterIn:       settings.GutterIn,
		CropRatio:      crop,
		CropToFill:     settings.CropToFill,
		UserScale:      settings.UserScale,
		ImagePosition:  Point{X: settings.Position.X, Y: settings.Position.Y},
		PositionUnits:  settings.Position.Units,
	}
	return inp, nil
}

// RenderOptions captures CLI-level toggles that are not part of the canonical YAML export block.
type RenderOptions struct {
	PNGLevel       string
	Scaler         string
	ParallelEncode bool
}

// RenderInfoFromExport prepares render info from resolved export settings and CLI overrides.
func RenderInfoFromExport(export sonnetcfg.ExportValues, index int, spreadName, imageBase string, opts RenderOptions) RenderInfo {
	format := export.Format
	if format == "" {
		format = "png"
	}
	info := RenderInfo{
		Index:          index,
		SpreadName:     spreadName,
		ImageBaseName:  imageBase,
		Format:         format,
		Quality:        export.Quality,
		Background:     export.Background,
		FilenameTmpl:   export.FilenameTemplate,
		OutputDir:      export.OutDir,
		PNGLevel:       opts.PNGLevel,
		Scaler:         opts.Scaler,
		ParallelEncode: opts.ParallelEncode,
	}
	if info.Quality == 0 {
		info.Quality = 90
	}
	if info.Background == "" {
		info.Background = "transparent"
	}
	if info.PNGLevel == "" {
		info.PNGLevel = "default"
	}
	if info.Scaler == "" {
		info.Scaler = "quality"
	}
	if info.OutputDir == "" {
		info.OutputDir = "./out"
	}
	if info.FilenameTmpl == "" {
		info.FilenameTmpl = "{index:03d}-{name}-{panel}.{ext}"
	}
	return info
}
