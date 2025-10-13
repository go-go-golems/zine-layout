package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/go-go-golems/zine-layout/pkg/app"
	"github.com/go-go-golems/zine-layout/pkg/zinelayout"
)

// scenario describes a page layout configuration exercised by the harness.
type scenario struct {
	ID          string
	Name        string
	Description string
	Spec        string
	SpecPath    string
	SpecDisplay string
}

type sizeSpec struct {
	ID     string
	Width  int
	Height int
}

type runResult struct {
	Scenario    scenario
	Size        sizeSpec
	Command     string
	InputImages []imageInfo
	PageResults []pageResult
	Stdout      string
	Stderr      string
	Error       error
}

type imageInfo struct {
	Path        string
	DisplayPath string
	Width       int
	Height      int
}

type pageResult struct {
	PageID         string
	OutputPath     string
	DisplayPath    string
	ExpectedWidth  int
	ExpectedHeight int
	ActualWidth    int
	ActualHeight   int
	Diagnostics    pageDiagnostics
	DiagnosticsRaw string
	Status         string
	Error          string
}

type pageDiagnostics struct {
	CanvasWidth  int            `json:"canvasWidth"`
	CanvasHeight int            `json:"canvasHeight"`
	FinalWidth   int            `json:"finalWidth"`
	FinalHeight  int            `json:"finalHeight"`
	Cells        []cellSummary  `json:"cells"`
	Margins      marginSnapshot `json:"margins"`
}

type marginSnapshot struct {
	PageSetup zinelayout.Margin `json:"pageSetup"`
	Page      zinelayout.Margin `json:"page"`
}

type cellSummary struct {
	InputIndex  int `json:"inputIndex"`
	Row         int `json:"row"`
	Column      int `json:"column"`
	Rotation    int `json:"rotation"`
	CellX       int `json:"cellX"`
	CellY       int `json:"cellY"`
	CellWidth   int `json:"cellWidth"`
	CellHeight  int `json:"cellHeight"`
	ImageX      int `json:"imageX"`
	ImageY      int `json:"imageY"`
	ImageWidth  int `json:"imageWidth"`
	ImageHeight int `json:"imageHeight"`
	Margins     struct {
		Top    int `json:"top"`
		Right  int `json:"right"`
		Bottom int `json:"bottom"`
		Left   int `json:"left"`
	} `json:"margins"`
}

func main() {
	scenarios := []scenario{
		{
			ID:          "single-portrait",
			Name:        "Single Portrait Page",
			Description: "One-up portrait page with generous global and cell margins.",
			Spec: strings.TrimSpace(`
---
global:
  ppi: 300
  border:
    enabled: true
    color: "#0f172a"
    type: plain
page_setup:
  grid_size:
    rows: 1
    columns: 1
  margin:
    top: 0.35in
    bottom: 0.35in
    left: 0.25in
    right: 0.25in
  border:
    enabled: true
    color: "#1d4ed8"
    type: plain
output_pages:
  - id: portrait-full
    margin:
      top: 18px
      bottom: 18px
      left: 24px
      right: 24px
    border:
      enabled: true
      color: "#22c55e"
      type: dotted
    layout:
      - input_index: 1
        position:
          row: 0
          column: 0
        rotation: 0
        margin:
          top: 0.1in
          bottom: 0.1in
          left: 0.2in
          right: 0.2in
`),
		},
		{
			ID:          "spread",
			Name:        "Two-Page Spread",
			Description: "Landscape spread with gutter and mirrored cells, highlighting spread splitting.",
			Spec: strings.TrimSpace(`
---
global:
  ppi: 300
page_setup:
  grid_size:
    rows: 1
    columns: 2
  margin:
    top: 0.25in
    bottom: 0.25in
    left: 0.3in
    right: 0.3in
output_pages:
  - id: spread-combined
    margin:
      top: 12px
      bottom: 12px
      left: 12px
      right: 12px
    layout:
      - input_index: 1
        position:
          row: 0
          column: 0
        rotation: 0
        margin:
          top: 8px
          bottom: 8px
          left: 12px
          right: 8px
      - input_index: 2
        position:
          row: 0
          column: 1
        rotation: 180
        margin:
          top: 8px
          bottom: 8px
          left: 8px
          right: 12px
`),
		},
		{
			ID:          "grid",
			Name:        "Four-Up Grid",
			Description: "2x2 grid with alternating borders and cell-level padding to stress layout math.",
			Spec: strings.TrimSpace(`
---
global:
  ppi: 200
page_setup:
  grid_size:
    rows: 2
    columns: 2
  margin:
    top: 0.2in
    bottom: 0.2in
    left: 0.2in
    right: 0.2in
  border:
    enabled: true
    color: "#f97316"
    type: dashed
output_pages:
  - id: grid-quad
    margin:
      top: 10px
      bottom: 10px
      left: 10px
      right: 10px
    border:
      enabled: true
      color: "#be123c"
      type: plain
    layout:
      - input_index: 1
        position:
          row: 0
          column: 0
        rotation: 0
        margin:
          top: 6px
          bottom: 6px
          left: 6px
          right: 6px
      - input_index: 2
        position:
          row: 0
          column: 1
        rotation: 0
        margin:
          top: 0.15in
          bottom: 0.1in
          left: 6px
          right: 12px
      - input_index: 3
        position:
          row: 1
          column: 0
        rotation: 0
        margin:
          top: 0.08in
          bottom: 0.08in
          left: 10px
          right: 10px
      - input_index: 4
        position:
          row: 1
          column: 1
        rotation: 0
        margin:
          top: 5px
          bottom: 5px
          left: 5px
          right: 5px
`),
		},
	}

	sizes := []sizeSpec{
		{ID: "landscape", Width: 3600, Height: 2400},
		{ID: "portrait", Width: 2400, Height: 3600},
		{ID: "square", Width: 2800, Height: 2800},
	}

	outputRoot, err := prepareOutputRoot()
	if err != nil {
		panic(err)
	}

	cliPath, err := buildCLI(outputRoot)
	if err != nil {
		panic(err)
	}

	// Persist scenario specs once.
	for i := range scenarios {
		path := filepath.Join(outputRoot, "specs", scenarios[i].ID+".yaml")
		if err := os.WriteFile(path, []byte(scenarios[i].Spec+"\n"), 0o644); err != nil {
			panic(err)
		}
		scenarios[i].SpecPath = path
		scenarios[i].SpecDisplay = relPath(path)
	}

	var results []runResult
	for _, sc := range scenarios {
		for _, sz := range sizes {
			res := executeScenario(cliPath, sc, sz, outputRoot)
			results = append(results, res)
			if res.Error != nil {
				fmt.Fprintf(os.Stderr, "scenario %s/%s failed: %v\n", sc.ID, sz.ID, res.Error)
			}
		}
	}

	if err := writeReport(outputRoot, results); err != nil {
		panic(err)
	}

	fmt.Printf("Report generated at %s\n", filepath.Join(outputRoot, "index.html"))
}

func prepareOutputRoot() (string, error) {
	today := time.Now().Format("2006-01-02")
	root := filepath.Join("ttmp", today, "pagelayout-validation")
	for _, sub := range []string{"assets", "renders", "specs", "diagnostics"} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
			return "", err
		}
	}
	return root, nil
}

func buildCLI(outputRoot string) (string, error) {
	binPath := filepath.Join(outputRoot, "zine-layout-cli")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/zine-layout")
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build cli: %w (output: %s)", err, combined.String())
	}
	return binPath, nil
}

func executeScenario(cliPath string, sc scenario, sz sizeSpec, root string) runResult {
	res := runResult{Scenario: sc, Size: sz}
	layouts, err := app.LoadLayoutsFromSpec(sc.SpecPath, map[string]interface{}{})
	if err != nil {
		res.Error = fmt.Errorf("parse spec: %w", err)
		return res
	}
	if len(layouts) == 0 {
		res.Error = fmt.Errorf("no layout documents in %s", sc.SpecPath)
		return res
	}
	zl := layouts[0]

	maxInput := 0
	for _, page := range zl.OutputPages {
		for _, cell := range page.Layout {
			if cell.InputIndex > maxInput {
				maxInput = cell.InputIndex
			}
		}
	}
	if maxInput == 0 {
		res.Error = fmt.Errorf("spec %s has no layout cells", sc.ID)
		return res
	}

	images, err := synthesizeInputs(sc.ID, sz, maxInput, root)
	if err != nil {
		res.Error = err
		return res
	}
	res.InputImages = images

	renderDir := filepath.Join(root, "renders", fmt.Sprintf("%s_%s", sc.ID, sz.ID))
	if err := os.RemoveAll(renderDir); err != nil {
		res.Error = err
		return res
	}
	if err := os.MkdirAll(renderDir, 0o755); err != nil {
		res.Error = err
		return res
	}

	absSpec, _ := filepath.Abs(sc.SpecPath)
	absRender, _ := filepath.Abs(renderDir)

	args := []string{
		"render",
		"--spec", absSpec,
		"--output-dir", absRender,
		"--verbose",
	}
	for _, img := range images {
		absImg, _ := filepath.Abs(img.Path)
		args = append(args, absImg)
	}

	res.Command = cliPath + " " + strings.Join(args, " ")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, cliPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()
	if err != nil {
		res.Error = fmt.Errorf("command failed: %w", err)
		return res
	}

	pageResults, err := validateOutputs(sc.SpecPath, images, renderDir)
	if err != nil {
		res.Error = err
		return res
	}
	res.PageResults = pageResults
	return res
}

func synthesizeInputs(scenarioID string, sz sizeSpec, count int, root string) ([]imageInfo, error) {
	var infos []imageInfo
	for i := 0; i < count; i++ {
		img := synthesizeImage(fmt.Sprintf("%s-%d", scenarioID, i), sz.Width, sz.Height)
		name := fmt.Sprintf("%s_%s_input%02d.png", scenarioID, sz.ID, i+1)
		path := filepath.Join(root, "assets", name)
		if err := savePNG(path, img); err != nil {
			return nil, fmt.Errorf("save input %d: %w", i, err)
		}
		infos = append(infos, imageInfo{
			Path:        path,
			DisplayPath: relPath(path),
			Width:       sz.Width,
			Height:      sz.Height,
		})
	}
	return infos, nil
}

func synthesizeImage(seed string, w, h int) image.Image {
	if w <= 0 {
		w = 1
	}
	if h <= 0 {
		h = 1
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	base := colorFromString(seed)
	draw.Draw(img, img.Bounds(), &image.Uniform{C: base}, image.Point{}, draw.Src)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			mix := float64(x+y) / float64(w+h)
			r := lerpColor(base.R, 255, mix)
			g := lerpColor(base.G, 180, mix)
			b := lerpColor(base.B, 90, mix)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	stripe := color.RGBA{255 - base.R/2, 255 - base.G/2, 255 - base.B/2, 255}
	step := h / 8
	if step <= 0 {
		step = 1
	}
	for y := 0; y < h; y += step {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, stripe)
		}
	}
	return img
}

func colorFromString(s string) color.RGBA {
	hash := uint32(0)
	for _, c := range s {
		hash = hash*33 + uint32(c)
	}
	return color.RGBA{
		R: uint8((hash >> 16) & 0xFF),
		G: uint8((hash >> 8) & 0xFF),
		B: uint8(hash & 0xFF),
		A: 255,
	}
}

func lerpColor(a, b uint8, t float64) uint8 {
	av := float64(a)
	bv := float64(b)
	v := av + (bv-av)*t
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return uint8(v + 0.5)
}

func savePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func validateOutputs(specPath string, inputs []imageInfo, renderDir string) ([]pageResult, error) {
	var results []pageResult

	baseLayouts, err := app.LoadLayoutsFromSpec(specPath, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("parse spec for validation: %w", err)
	}
	if len(baseLayouts) == 0 {
		return nil, fmt.Errorf("no layouts found in %s", specPath)
	}
	baseLayout := baseLayouts[0]

	images := make([]image.Image, len(inputs))
	for i, info := range inputs {
		f, err := os.Open(info.Path)
		if err != nil {
			return nil, fmt.Errorf("open input %s: %w", info.Path, err)
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("decode input %s: %w", info.Path, err)
		}
		images[i] = img
	}

	for _, basePage := range baseLayout.OutputPages {
		pr := pageResult{PageID: basePage.ID}
		outputPath := filepath.Join(renderDir, ensurePNGExtension(basePage.ID))
		pr.OutputPath = outputPath
		pr.DisplayPath = relPath(outputPath)

		layoutCopies, err := app.LoadLayoutsFromSpec(specPath, map[string]interface{}{})
		if err != nil {
			return nil, fmt.Errorf("re-parse spec for page %s: %w", basePage.ID, err)
		}
		if len(layoutCopies) == 0 {
			return nil, fmt.Errorf("no layouts available when validating page %s", basePage.ID)
		}
		current := layoutCopies[0]

		var targetPage *zinelayout.OutputPage
		for _, p := range current.OutputPages {
			if p.ID == basePage.ID {
				targetPage = p
				break
			}
		}
		if targetPage == nil {
			return nil, fmt.Errorf("unable to locate page %s in spec", basePage.ID)
		}

		if err := current.ComputeAllMargins(); err != nil {
			return nil, fmt.Errorf("compute margins: %w", err)
		}

		diag, err := computeDiagnostics(&current, targetPage, images)
		if err != nil {
			return nil, fmt.Errorf("diagnostics for %s: %w", basePage.ID, err)
		}
		pr.Diagnostics = diag
		diagJSON, _ := json.MarshalIndent(diag, "", "  ")
		pr.DiagnosticsRaw = string(diagJSON)

		expectedImage, err := current.CreateOutputImage(targetPage, images)
		if err != nil {
			return nil, fmt.Errorf("render expected image for %s: %w", basePage.ID, err)
		}
		pr.ExpectedWidth = expectedImage.Bounds().Dx()
		pr.ExpectedHeight = expectedImage.Bounds().Dy()

		actualW, actualH, err := imageDimensions(outputPath)
		if err != nil {
			pr.Status = "error"
			pr.Error = fmt.Sprintf("read output: %v", err)
			results = append(results, pr)
			continue
		}
		pr.ActualWidth = actualW
		pr.ActualHeight = actualH

		if pr.ExpectedWidth == actualW && pr.ExpectedHeight == actualH {
			pr.Status = "ok"
		} else {
			pr.Status = "mismatch"
			pr.Error = fmt.Sprintf("expected %dx%d got %dx%d", pr.ExpectedWidth, pr.ExpectedHeight, actualW, actualH)
		}

		results = append(results, pr)
	}

	return results, nil
}

func ensurePNGExtension(id string) string {
	if strings.HasSuffix(strings.ToLower(id), ".png") {
		return id
	}
	return id + ".png"
}

func imageDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	img, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return img.Width, img.Height, nil
}

func computeDiagnostics(zl *zinelayout.ZineLayout, page *zinelayout.OutputPage, inputs []image.Image) (pageDiagnostics, error) {
	diag := pageDiagnostics{}
	if len(inputs) == 0 {
		return diag, fmt.Errorf("no inputs provided")
	}
	inputSize := inputs[0].Bounds().Size()

	rows := zl.PageSetup.GridSize.Rows
	cols := zl.PageSetup.GridSize.Columns
	maxRow, maxCol := 0, 0
	for _, l := range page.Layout {
		if l.Position.Row > maxRow {
			maxRow = l.Position.Row
		}
		if l.Position.Column > maxCol {
			maxCol = l.Position.Column
		}
	}
	if rows <= 0 || rows <= maxRow {
		rows = maxRow + 1
	}
	if cols <= 0 || cols <= maxCol {
		cols = maxCol + 1
	}

	type cell struct {
		margin *zinelayout.Margin
		width  int
		height int
		x      int
		y      int
	}

	cells := make([][]cell, rows)
	for r := range cells {
		cells[r] = make([]cell, cols)
		for c := range cells[r] {
			cells[r][c].margin = &zinelayout.Margin{}
		}
	}

	for _, layout := range page.Layout {
		r := layout.Position.Row
		c := layout.Position.Column
		if r < 0 || c < 0 || r >= rows || c >= cols {
			return diag, fmt.Errorf("layout cell out of bounds row=%d col=%d", r, c)
		}
		margin := layout.Margin
		if margin == nil {
			margin = &zinelayout.Margin{}
		}
		cells[r][c].margin = margin
		cells[r][c].width = inputSize.X + margin.Left.Pixels + margin.Right.Pixels
		cells[r][c].height = inputSize.Y + margin.Top.Pixels + margin.Bottom.Pixels
	}

	width := 0
	height := 0
	totalWidth := 0
	totalHeight := 0
	for r := range cells {
		maxH := 0
		for c := range cells[r] {
			cells[r][c].x = width
			cells[r][c].y = height
			width += cells[r][c].width
			if cells[r][c].height > maxH {
				maxH = cells[r][c].height
			}
		}
		height += maxH
		if width > totalWidth {
			totalWidth = width
		}
		totalHeight += maxH
		width = 0
	}

	canvasWidth := totalWidth
	canvasHeight := totalHeight
	diag.CanvasWidth = canvasWidth
	diag.CanvasHeight = canvasHeight

	psMargin := ensureMargin(zl.PageSetup.Margin)
	pageMargin := ensureMargin(page.Margin)

	finalWidth := canvasWidth + psMargin.Left.Pixels + psMargin.Right.Pixels + pageMargin.Left.Pixels + pageMargin.Right.Pixels
	finalHeight := canvasHeight + psMargin.Top.Pixels + psMargin.Bottom.Pixels + pageMargin.Top.Pixels + pageMargin.Bottom.Pixels
	diag.FinalWidth = finalWidth
	diag.FinalHeight = finalHeight
	diag.Margins = marginSnapshot{PageSetup: *psMargin, Page: *pageMargin}

	for _, layout := range page.Layout {
		r := layout.Position.Row
		c := layout.Position.Column
		cell := cells[r][c]
		if layout.InputIndex <= 0 || layout.InputIndex > len(inputs) {
			return diag, fmt.Errorf("layout input index %d out of range", layout.InputIndex)
		}
		img := inputs[layout.InputIndex-1]
		bounds := img.Bounds().Size()
		imgW := bounds.X
		imgH := bounds.Y
		destX := cell.x + cell.margin.Left.Pixels
		destY := cell.y + cell.margin.Top.Pixels

		summary := cellSummary{
			InputIndex:  layout.InputIndex,
			Row:         r,
			Column:      c,
			Rotation:    layout.Rotation,
			CellX:       cell.x,
			CellY:       cell.y,
			CellWidth:   cell.width,
			CellHeight:  cell.height,
			ImageX:      destX,
			ImageY:      destY,
			ImageWidth:  imgW,
			ImageHeight: imgH,
		}
		summary.Margins.Top = cell.margin.Top.Pixels
		summary.Margins.Bottom = cell.margin.Bottom.Pixels
		summary.Margins.Left = cell.margin.Left.Pixels
		summary.Margins.Right = cell.margin.Right.Pixels
		diag.Cells = append(diag.Cells, summary)
	}

	return diag, nil
}

func ensureMargin(m *zinelayout.Margin) *zinelayout.Margin {
	if m == nil {
		return &zinelayout.Margin{}
	}
	return m
}

func relPath(path string) string {
	rel, err := filepath.Rel(".", path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

func writeReport(root string, results []runResult) error {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>Page Layout Validation</title>
<style>
body { font-family: "Inter", system-ui, -apple-system, sans-serif; margin: 0; padding: 2rem; background: #f8fafc; color: #0f172a; }
h1 { margin-bottom: 1rem; }
section { margin-bottom: 3rem; background: #fff; border-radius: 16px; padding: 1.5rem; box-shadow: 0 20px 40px rgba(15, 23, 42, 0.08); }
.grid { display: grid; gap: 1.5rem; }
@media (min-width: 1024px) { .grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
.badge { display: inline-flex; align-items: center; padding: 0.15rem 0.5rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; }
.badge-ok { background: rgba(16, 185, 129, 0.15); color: #047857; }
.badge-err { background: rgba(239, 68, 68, 0.15); color: #b91c1c; }
.badge-warn { background: rgba(249, 115, 22, 0.15); color: #c2410c; }
pre { background: #0f172a; color: #e2e8f0; padding: 1rem; border-radius: 12px; overflow: auto; font-size: 0.85rem; }
details { background: #f1f5f9; border-radius: 12px; padding: 0.75rem 1rem; }
details + details { margin-top: 0.75rem; }
details summary { cursor: pointer; font-weight: 600; }
figure { margin: 0; }
figcaption { margin-top: 0.35rem; font-size: 0.8rem; color: #475569; }
img { max-width: 100%; border-radius: 12px; border: 1px solid rgba(15, 23, 42, 0.1); background: #fff; }
table { width: 100%; border-collapse: collapse; font-size: 0.85rem; margin-top: 0.75rem; }
th, td { padding: 0.45rem 0.6rem; text-align: left; border-bottom: 1px solid rgba(148, 163, 184, 0.35); }
th { background: rgba(15, 23, 42, 0.05); font-weight: 600; }
code { background: rgba(15, 23, 42, 0.08); padding: 0.1rem 0.35rem; border-radius: 6px; font-size: 0.85rem; }
</style>
</head>
<body>
<h1>Page Layout Validation Report</h1>
<p>Generated at {{.GeneratedAt}}. Each section exercises the <code>zine-layout render</code> CLI using synthetic inputs across multiple aspect ratios. Render outputs are validated against library computations to ensure consistent canvas sizing and margin math.</p>
{{range .Items}}
<section>
  <h2>{{.Scenario.Name}}</h2>
  <p>{{.Scenario.Description}}</p>
  <details open>
    <summary>Layout Specification ({{.Scenario.SpecDisplay}})</summary>
    <pre>{{.Scenario.Spec}}</pre>
  </details>
  <div class="grid">
    {{range .Runs}}
    <div>
      <h3>{{.Size.ID | title}}</h3>
      <p><code>{{.Command}}</code></p>
      {{if .Error}}
        <span class="badge badge-err">failed</span>
        <p>{{.Error}}</p>
      {{else}}
        {{if .HasMismatch}}
          <span class="badge badge-warn">validation issues</span>
        {{else}}
          <span class="badge badge-ok">validated</span>
        {{end}}
        <details>
          <summary>CLI Stdout</summary>
          <pre>{{.Stdout}}</pre>
        </details>
        {{if .Stderr}}
        <details>
          <summary>CLI Stderr</summary>
          <pre>{{.Stderr}}</pre>
        </details>
        {{end}}
        <details>
          <summary>Inputs</summary>
          <table>
            <thead><tr><th>#</th><th>Path</th><th>Size</th></tr></thead>
            <tbody>
              {{range $idx, $img := .InputImages}}
              <tr><td>{{$idx | inc}}</td><td>{{$img.DisplayPath}}</td><td>{{$img.Width}}×{{$img.Height}}</td></tr>
              {{end}}
            </tbody>
          </table>
        </details>
        {{range .PageResults}}
          <details open>
            <summary>Page {{.PageID}} – {{.ActualWidth}}×{{.ActualHeight}} px (expected {{.ExpectedWidth}}×{{.ExpectedHeight}})</summary>
            {{if eq .Status "ok"}}
              <span class="badge badge-ok">match</span>
            {{else if eq .Status "mismatch"}}
              <span class="badge badge-warn">mismatch</span>
              <p>{{.Error}}</p>
            {{else}}
              <span class="badge badge-err">error</span>
              <p>{{.Error}}</p>
            {{end}}
            <figure>
              <img src="{{.DisplayPath}}" alt="Rendered page {{.PageID}}" />
              <figcaption>Rendered page output</figcaption>
            </figure>
            <details>
              <summary>Layout diagnostics</summary>
              <pre>{{.DiagnosticsRaw}}</pre>
            </details>
          </details>
        {{end}}
      {{end}}
    </div>
    {{end}}
  </div>
</section>
{{end}}
</body>
</html>`

	type templateRun struct {
		runResult
		HasMismatch bool
	}
	type templateScenario struct {
		Scenario scenario
		Runs     []templateRun
	}

	data := struct {
		GeneratedAt string
		Items       []templateScenario
	}{
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	grouped := map[string][]templateRun{}
	for _, res := range results {
		tr := templateRun{runResult: res}
		for _, pr := range res.PageResults {
			if pr.Status != "ok" {
				tr.HasMismatch = true
				break
			}
		}
		grouped[res.Scenario.ID] = append(grouped[res.Scenario.ID], tr)
	}

	for _, runs := range grouped {
		if len(runs) == 0 {
			continue
		}
		sc := runs[0].Scenario
		data.Items = append(data.Items, templateScenario{Scenario: sc, Runs: runs})
	}

	sort.Slice(data.Items, func(i, j int) bool {
		return data.Items[i].Scenario.ID < data.Items[j].Scenario.ID
	})

	funcMap := template.FuncMap{
		"title": titleCase,
		"inc":   func(i int) int { return i + 1 },
	}

	t, err := template.New("report").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	outPath := filepath.Join(root, "index.html")
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, data)
}

func titleCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)
	for i, word := range words {
		runes := []rune(strings.ToLower(word))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
