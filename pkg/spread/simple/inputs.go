package simple

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/spread"
)

// RenderOptions captures CLI-level toggles that are not part of the canonical export block.
type RenderOptions struct {
	PNGLevel       string
	Scaler         string
	ParallelEncode bool
}

// InputsFromSettings converts spread.Settings into simple algorithm inputs.
func InputsFromSettings(settings spread.Settings, meta spread.ImageMeta) (Inputs, error) {
	if meta.Width <= 0 || meta.Height <= 0 {
		return Inputs{}, fmt.Errorf("invalid source dimensions %dx%d", meta.Width, meta.Height)
	}

	var crop *CropRatio
	if settings.CropRatio != nil {
		ratio := *settings.CropRatio
		if ratio <= 0 {
			return Inputs{}, fmt.Errorf("crop ratio must be > 0")
		}
		crop = &CropRatio{W: ratio, H: 1}
	}

	positionUnits := strings.TrimSpace(settings.Units)
	if positionUnits == "" {
		positionUnits = "normalized"
	}

	inputs := Inputs{
		SrcW:           float64(meta.Width),
		SrcH:           float64(meta.Height),
		PaperWIn:       settings.PaperWidthIn,
		PaperHIn:       settings.PaperHeightIn,
		Orientation:    settings.Orientation,
		MarginTopIn:    settings.MarginTopIn,
		MarginRightIn:  settings.MarginRightIn,
		MarginBottomIn: settings.MarginBottomIn,
		MarginLeftIn:   settings.MarginLeftIn,
		DPI:            settings.DPI,
		IsSpread:       settings.IsSpread,
		GutterIn:       settings.GutterIn,
		CropRatio:      crop,
		CropToFill:     settings.CropToFill,
		UserScale:      settings.UserScale,
		ImagePosition:  Point{X: settings.PositionX, Y: settings.PositionY},
		PositionUnits:  positionUnits,
	}
	return inputs, nil
}

// RenderInfoFromExport prepares render info from spread.Export along with optional overrides.
func RenderInfoFromExport(export spread.Export, index int, spreadName, imageBase string, opts RenderOptions) RenderInfo {
	format := strings.TrimSpace(export.Format)
	if format == "" {
		format = "png"
	}

	background := strings.TrimSpace(export.Background)
	if background == "" {
		background = "white"
	}

	filenameTemplate := export.FilenameTemplate
	if strings.TrimSpace(filenameTemplate) == "" {
		filenameTemplate = "{index:03d}-{name}-{panel}.{ext}"
	}

	outputDir := export.OutDir
	if strings.TrimSpace(outputDir) == "" {
		outputDir = "./out"
	}

	info := RenderInfo{
		Index:          index,
		SpreadName:     spreadName,
		ImageBaseName:  imageBase,
		Format:         format,
		Quality:        export.Quality,
		Background:     background,
		FilenameTmpl:   filenameTemplate,
		OutputDir:      outputDir,
		PNGLevel:       opts.PNGLevel,
		Scaler:         opts.Scaler,
		ParallelEncode: opts.ParallelEncode,
	}

	if info.Quality == 0 {
		info.Quality = 90
	}
	if info.PNGLevel == "" {
		info.PNGLevel = "default"
	}
	if info.Scaler == "" {
		info.Scaler = "quality"
	}

	return info
}
