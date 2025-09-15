package app

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/go-emrichen/pkg/emrichen"
	"github.com/go-go-golems/zine-layout/pkg/zinelayout"
	"github.com/go-go-golems/zine-layout/pkg/zinelayout/parser"
	"gopkg.in/yaml.v3"
)

// Overrides holds optional overrides applied on top of a parsed layout
type Overrides struct {
	GlobalBorder bool
	PageBorder   bool
	LayoutBorder bool
	InnerBorder  bool
	BorderColor  string
	BorderType   string
	PPI          int
}

// LoadLayoutsFromSpec loads one or more ZineLayout documents from a YAML file,
// processing Go-Emrichen templates.
func LoadLayoutsFromSpec(specPath string, env map[string]interface{}) ([]zinelayout.ZineLayout, error) {
	var layouts []zinelayout.ZineLayout

	yamlFile, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("reading YAML file: %w", err)
	}

	_ = yamlFile // available for future debug hooks

	interpreter, err := emrichen.NewInterpreter(
		emrichen.WithVars(env),
		emrichen.WithFuncMap(sprig.TxtFuncMap()),
	)
	if err != nil {
		return nil, fmt.Errorf("creating Emrichen interpreter: %w", err)
	}

	f, err := os.Open(specPath)
	if err != nil {
		return nil, fmt.Errorf("opening spec: %w", err)
	}
	defer func() { _ = f.Close() }()

	decoder := yaml.NewDecoder(f)
	for {
		var document interface{}
		err = decoder.Decode(interpreter.CreateDecoder(&document))
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("processing YAML with Emrichen: %w", err)
		}
		if document == nil {
			continue
		}
		processedYAMLBytes, err := yaml.Marshal(document)
		if err != nil {
			return nil, fmt.Errorf("marshaling processed YAML: %w", err)
		}
		var zl zinelayout.ZineLayout
		if err := yaml.Unmarshal(processedYAMLBytes, &zl); err != nil {
			return nil, fmt.Errorf("parsing processed YAML: %w", err)
		}
		layouts = append(layouts, zl)
	}
	return layouts, nil
}

// ApplyOverrides applies override settings onto a parsed layout.
func ApplyOverrides(zl *zinelayout.ZineLayout, ov Overrides) error {
    if zl.Global == nil {
        zl.Global = &zinelayout.Global{}
    }
    if zl.PageSetup == nil {
        zl.PageSetup = &zinelayout.PageSetup{}
    }
    if ov.PPI > 0 {
        zl.Global.PPI = float64(ov.PPI)
    }
	if ov.GlobalBorder {
		if zl.Global.Border == nil {
			zl.Global.Border = &zinelayout.Border{}
		}
		zl.Global.Border.Enabled = true
	}
	if ov.PageBorder {
		if zl.PageSetup.PageBorder == nil {
			zl.PageSetup.PageBorder = &zinelayout.Border{}
		}
		zl.PageSetup.PageBorder.Enabled = true
	}
	if ov.LayoutBorder {
		for i := range zl.OutputPages {
			if zl.OutputPages[i].LayoutBorder == nil {
				zl.OutputPages[i].LayoutBorder = &zinelayout.Border{}
			}
			zl.OutputPages[i].LayoutBorder.Enabled = true
		}
	}
	if ov.InnerBorder {
		for i := range zl.OutputPages {
			for j := range zl.OutputPages[i].Layout {
				if zl.OutputPages[i].Layout[j].InnerLayoutBorder == nil {
					zl.OutputPages[i].Layout[j].InnerLayoutBorder = &zinelayout.Border{}
				}
				zl.OutputPages[i].Layout[j].InnerLayoutBorder.Enabled = true
			}
		}
	}
	if ov.BorderColor != "" {
		c, err := ParseBorderColor(ov.BorderColor)
		if err != nil {
			return err
		}
		if zl.Global.Border == nil {
			zl.Global.Border = &zinelayout.Border{}
		}
		zl.Global.Border.Color = zinelayout.CustomColor{RGBA: c}
	}
	if ov.BorderType != "" {
		bt, err := zinelayout.ParseBorderType(ov.BorderType)
		if err != nil {
			return err
		}
		if zl.Global.Border == nil {
			zl.Global.Border = &zinelayout.Border{}
		}
		zl.Global.Border.Type = bt
	}
	return nil
}

// ParseTestDimensions parses WIDTH,HEIGHT using zinelayout units at given PPI.
func ParseTestDimensions(dimensions string, ppi float64) (int, int, error) {
	dimensions = strings.TrimSpace(dimensions)
	if dimensions == "" {
		return 600, 800, nil
	}
	parts := strings.Split(dimensions, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid test dimensions; expected WIDTH,HEIGHT")
	}
	ep := &parser.ExpressionParser{PPI: ppi}
	wv, err := ep.Parse(parts[0])
	if err != nil {
		return 0, 0, err
	}
	hv, err := ep.Parse(parts[1])
	if err != nil {
		return 0, 0, err
	}
	uc := parser.UnitConverter{PPI: ppi}
	w, err := uc.ToPixels(wv.Val, wv.Unit)
	if err != nil {
		return 0, 0, err
	}
	h, err := uc.ToPixels(hv.Val, hv.Unit)
	if err != nil {
		return 0, 0, err
	}
	return int(w), int(h), nil
}

// ReadInputImages decodes PNG images from file paths.
func ReadInputImages(files []string) ([]image.Image, error) {
	var images_ []image.Image
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		img, _, err := image.Decode(f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}
		images_ = append(images_, img)
	}
	return images_, nil
}

// RenderOutputs renders all output pages and writes PNG files to outDir.
// Returns the written file paths.
func RenderOutputs(zl *zinelayout.ZineLayout, inputs []image.Image, outDir string) ([]string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	var written []string
	for _, outputPage := range zl.OutputPages {
		img, err := zl.CreateOutputImage(outputPage, inputs)
		if err != nil {
			return nil, err
		}
		filePath := filepath.Join(outDir, outputPage.ID)
		if !strings.HasSuffix(strings.ToLower(filePath), ".png") {
			filePath += ".png"
		}
		if err := writePNG(img, filePath); err != nil {
			return nil, err
		}
		written = append(written, filePath)
	}
	return written, nil
}

func writePNG(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

// ParseBorderColor accepts #hex, color names, or R,G,B,A
func ParseBorderColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") || s == "black" || s == "white" {
		var c zinelayout.CustomColor
		if err := c.UnmarshalYAML(&yaml.Node{Kind: yaml.ScalarNode, Value: s}); err != nil {
			return color.RGBA{}, err
		}
		return c.RGBA, nil
	}
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return color.RGBA{}, fmt.Errorf("invalid color format; expected R,G,B,A or #hex or name")
	}
	var rgba [4]uint8
	for i, part := range parts {
		v, err := strconv.ParseUint(strings.TrimSpace(part), 10, 8)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("invalid color component: %s", part)
		}
		rgba[i] = uint8(v)
	}
	return color.RGBA{R: rgba[0], G: rgba[1], B: rgba[2], A: rgba[3]}, nil
}

// GenerateTestImages generates N images with optional BW pattern at size.
func GenerateTestImages(n, width, height int, bw bool) ([]image.Image, error) {
	if bw {
		return zinelayout.GenerateTestImagesBW(n, width, height)
	}
	return zinelayout.GenerateTestImages(n, width, height)
}

// DebugPrintZineLayout prints human-readable layout details.
func DebugPrintZineLayout(zl zinelayout.ZineLayout) {
	fmt.Printf("PageSetup:\n")
	fmt.Printf("  GridSize: Rows: %d, Columns: %d\n", zl.PageSetup.GridSize.Rows, zl.PageSetup.GridSize.Columns)
	fmt.Printf("  Margin: %+v\n", zl.PageSetup.Margin)
	if zl.PageSetup.PageBorder != nil {
		fmt.Printf("  PageBorder: Enabled: %v, Color: R:%d G:%d B:%d A:%d, Type: %s\n", zl.PageSetup.PageBorder.Enabled, zl.PageSetup.PageBorder.Color.R, zl.PageSetup.PageBorder.Color.G, zl.PageSetup.PageBorder.Color.B, zl.PageSetup.PageBorder.Color.A, zl.PageSetup.PageBorder.Type)
	}
	fmt.Printf("  PPI: %.0f\n", zl.Global.PPI)
	fmt.Printf("OutputPages:\n")
	for i, page := range zl.OutputPages {
		fmt.Printf("  Page %d:\n", i+1)
		fmt.Printf("    ID: %s\n", page.ID)
		fmt.Printf("    Margin: %+v\n", page.Margin)
		if page.LayoutBorder != nil {
			fmt.Printf("    LayoutBorder: Enabled: %v, Color: R:%d G:%d B:%d A:%d, Type: %s\n", page.LayoutBorder.Enabled, page.LayoutBorder.Color.R, page.LayoutBorder.Color.G, page.LayoutBorder.Color.B, page.LayoutBorder.Color.A, page.LayoutBorder.Type)
		}
		fmt.Printf("    Layout:\n")
		for j, layout := range page.Layout {
			fmt.Printf("      Layout %d:\n", j+1)
			fmt.Printf("        InputIndex: %d\n", layout.InputIndex)
			fmt.Printf("        Position: Row: %d, Column: %d\n", layout.Position.Row, layout.Position.Column)
			fmt.Printf("        Rotation: %d\n", layout.Rotation)
			fmt.Printf("        Margin: %+v\n", layout.Margin)
			if layout.InnerLayoutBorder != nil {
				fmt.Printf("        InnerLayoutBorder: Enabled: %v, Color: R:%d G:%d B:%d A:%d, Type: %s\n", layout.InnerLayoutBorder.Enabled, layout.InnerLayoutBorder.Color.R, layout.InnerLayoutBorder.Color.G, layout.InnerLayoutBorder.Color.B, layout.InnerLayoutBorder.Color.A, layout.InnerLayoutBorder.Type)
			}
		}
	}
	if zl.Global.Border != nil {
		fmt.Printf("GlobalBorder:\n")
		fmt.Printf("  Enabled: %v\n", zl.Global.Border.Enabled)
		fmt.Printf("  Color: R:%d G:%d B:%d A:%d\n", zl.Global.Border.Color.R, zl.Global.Border.Color.G, zl.Global.Border.Color.B, zl.Global.Border.Color.A)
		fmt.Printf("  Type: %s\n", zl.Global.Border.Type)
	}
}
