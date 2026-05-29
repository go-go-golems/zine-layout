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
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	xdraw "golang.org/x/image/draw"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout/engine"
)

type scenario struct {
	ID          string
	Name        string
	Description string
	Args        []string
}

type sizeSpec struct {
	ID     string
	Width  int
	Height int
}

type validationResult struct {
	Messages []string
	Errors   []error
}

type runResult struct {
	Scenario      scenario
	Size          sizeSpec
	Command       string
	RawOutput     []byte
	Computation   *imagelayout.Computation
	Inputs        engine.Inputs
	Validation    validationResult
	SourceImage   string
	RenderedImage string
	OverlayImage  string
	TraceJSON     string
	SettingsJSON  string
	ResultJSON    string
	Error         error
}

func main() {
	scenarios := []scenario{
		{
			ID:          "page-contain",
			Name:        "Page / Contain",
			Description: "Default page mode with contain behaviour inside the printable area.",
			Args:        []string{"--paper-width-in", "8.5", "--paper-height-in", "11", "--dpi", "300"},
		},
		{
			ID:          "page-cover",
			Name:        "Page / Cover",
			Description: "Page mode forcing cover scaling inside the printable area.",
			Args:        []string{"--paper-width-in", "8.5", "--paper-height-in", "11", "--dpi", "300", "--crop-to-fill"},
		},
		{
			ID:          "page-landscape-anchor",
			Name:        "Page / Landscape Anchor",
			Description: "Landscape orientation using the top-left anchor preset.",
			Args: []string{
				"--paper-width-in", "8.5",
				"--paper-height-in", "11",
				"--dpi", "200",
				"--orientation", "landscape",
				"--anchor-preset", "top-left",
			},
		},
		{
			ID:          "crop-ratio",
			Name:        "Crop / Ratio",
			Description: "Crop mode constrained to a 3:2 ratio.",
			Args: []string{
				"--mode", "crop",
				"--crop-ratio", "1.5",
				"--crop-to-fill",
			},
		},
		{
			ID:          "crop-dimensions",
			Name:        "Crop / Fixed Dimensions",
			Description: "Crop mode with explicit pixel dimensions.",
			Args: []string{
				"--mode", "crop",
				"--crop-width", "1600",
				"--crop-height", "900",
			},
		},
		{
			ID:          "fit-width",
			Name:        "Fit / Width",
			Description: "Fit mode driven by a target width.",
			Args: []string{
				"--mode", "fit",
				"--fit-mode", "width",
				"--fit-width", "1200",
			},
		},
		{
			ID:          "fit-height",
			Name:        "Fit / Height",
			Description: "Fit mode driven by a target height.",
			Args: []string{
				"--mode", "fit",
				"--fit-mode", "height",
				"--fit-height", "900",
			},
		},
		{
			ID:          "fit-auto",
			Name:        "Fit / Auto",
			Description: "Fit mode with both target dimensions provided.",
			Args: []string{
				"--mode", "fit",
				"--fit-width", "1400",
				"--fit-height", "700",
			},
		},
	}

	sizes := []sizeSpec{
		{ID: "landscape", Width: 4000, Height: 3000},
		{ID: "portrait", Width: 3000, Height: 4000},
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
	root := filepath.Join("ttmp", today, "imagelayout-validation")
	if err := os.MkdirAll(filepath.Join(root, "assets"), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(root, "renders"), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(root, "overlays"), 0o755); err != nil {
		return "", err
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

	img := synthesizeImage(sc.ID, sz)
	assetName := fmt.Sprintf("%s_%s.png", sc.ID, sz.ID)
	assetPath := filepath.Join(root, "assets", assetName)
	if err := savePNG(assetPath, img); err != nil {
		res.Error = fmt.Errorf("save source image: %w", err)
		return res
	}
	res.SourceImage = relPath(assetPath)

	cmdArgs := append([]string{"imagelayout", "compute", "--source-width", fmt.Sprint(sz.Width), "--source-height", fmt.Sprint(sz.Height)}, sc.Args...)
	res.Command = cliPath + " " + strings.Join(cmdArgs, " ")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	fmt.Printf("Running %s/%s\n", sc.ID, sz.ID)
	cmd := exec.CommandContext(ctx, cliPath, cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res.RawOutput = stdout.Bytes()
	if err != nil {
		res.Error = fmt.Errorf("command failed: %w (stderr: %s)", err, stderr.String())
		return res
	}

	var comp imagelayout.Computation
	if err := json.Unmarshal(res.RawOutput, &comp); err != nil {
		res.Error = fmt.Errorf("decode output: %w", err)
		return res
	}
	res.Computation = &comp

	inputs, err := engine.InputsFromSettings(comp.Settings, imagelayout.ImageMeta{Width: sz.Width, Height: sz.Height})
	if err != nil {
		res.Error = fmt.Errorf("derive inputs: %w", err)
		return res
	}
	res.Inputs = inputs

	res.Validation = validateComputation(comp, inputs)

	renderPath, overlayPath, err := renderOutputs(root, sc.ID, sz.ID, img, comp, inputs)
	if err != nil {
		res.Error = fmt.Errorf("render outputs: %w", err)
		return res
	}
	res.RenderedImage = relPath(renderPath)
	res.OverlayImage = relPath(overlayPath)

	res.TraceJSON = marshalPretty(comp.Trace)
	res.SettingsJSON = marshalPretty(comp.Settings)
	res.ResultJSON = marshalPretty(comp.Result)
	return res
}

func relPath(path string) string {
	rel, err := filepath.Rel(".", path)
	if err != nil {
		return path
	}
	return rel
}

func savePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

func synthesizeImage(seed string, sz sizeSpec) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, sz.Width, sz.Height))
	baseColor := colorFromString(seed)
	draw.Draw(img, img.Bounds(), &image.Uniform{C: baseColor}, image.Point{}, draw.Src)

	// Add diagonal gradient overlay for visual cues
	for y := 0; y < sz.Height; y++ {
		for x := 0; x < sz.Width; x++ {
			alpha := float64(x+y) / float64(sz.Width+sz.Height)
			r := clampColor(float64(baseColor.R)*(1-alpha) + 255*alpha)
			g := clampColor(float64(baseColor.G)*(1-alpha) + 128*alpha)
			b := clampColor(float64(baseColor.B)*(1-alpha) + 64*alpha)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func colorFromString(s string) color.RGBA {
	hash := uint32(0)
	for _, c := range s {
		hash = hash*33 + uint32(c)
	}
	r := uint8((hash >> 16) & 0xFF)
	g := uint8((hash >> 8) & 0xFF)
	b := uint8(hash & 0xFF)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func clampColor(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}

func validateComputation(comp imagelayout.Computation, inputs engine.Inputs) validationResult {
	vr := validationResult{}
	addMsg := func(msg string) { vr.Messages = append(vr.Messages, msg) }
	addErr := func(format string, args ...interface{}) {
		vr.Errors = append(vr.Errors, fmt.Errorf(format, args...))
	}

	if comp.Result.CanvasRect.W <= 0 || comp.Result.CanvasRect.H <= 0 {
		addErr("canvas rect has non-positive dimensions: %+v", comp.Result.CanvasRect)
	} else {
		addMsg(fmt.Sprintf("canvas %.2fx%.2f", comp.Result.CanvasRect.W, comp.Result.CanvasRect.H))
	}

	if diff := math.Abs(comp.Result.CanvasRect.W-inputs.ContentW) + math.Abs(comp.Result.CanvasRect.H-inputs.ContentH); diff > 1e-4 {
		addErr("canvas rect mismatch derived content size: diff=%.4f", diff)
	} else {
		addMsg("canvas rect matches derived content size")
	}

	expectedW := comp.Result.SourceRect.W * comp.Result.Scale
	expectedH := comp.Result.SourceRect.H * comp.Result.Scale
	if math.Abs(comp.Result.TargetRect.W-expectedW) > 1e-3 || math.Abs(comp.Result.TargetRect.H-expectedH) > 1e-3 {
		addErr("target rect does not match scale: expected %.3fx%.3f got %.3fx%.3f", expectedW, expectedH, comp.Result.TargetRect.W, comp.Result.TargetRect.H)
	} else {
		addMsg("target rect matches scale factor")
	}

	if comp.Result.Scale <= 0 {
		addErr("scale must be positive, got %.3f", comp.Result.Scale)
	} else {
		addMsg(fmt.Sprintf("scale %.3f", comp.Result.Scale))
	}

	if comp.Trace == nil || len(comp.Trace.Steps) == 0 {
		addErr("trace missing steps")
	} else {
		addMsg(fmt.Sprintf("trace contains %d steps", len(comp.Trace.Steps)))
	}

	return vr
}

func renderOutputs(root, scenarioID, sizeID string, src image.Image, comp imagelayout.Computation, inputs engine.Inputs) (string, string, error) {
	canvasW := int(math.Round(inputs.CanvasW))
	canvasH := int(math.Round(inputs.CanvasH))
	if canvasW <= 0 {
		canvasW = int(math.Round(comp.Result.CanvasRect.W + comp.Result.CanvasRect.X))
	}
	if canvasH <= 0 {
		canvasH = int(math.Round(comp.Result.CanvasRect.H + comp.Result.CanvasRect.Y))
	}
	if canvasW <= 0 {
		canvasW = src.Bounds().Dx()
	}
	if canvasH <= 0 {
		canvasH = src.Bounds().Dy()
	}

	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{235, 235, 240, 255}}, image.Point{}, draw.Src)

	// Highlight printable/content area
	contentRect := rectToImageRect(comp.Result.CanvasRect)
	draw.Draw(canvas, contentRect.Intersect(canvas.Bounds()), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	cropped := cropImage(src, comp.Result.SourceRect)
	targetRect := rectToImageRect(comp.Result.TargetRect)
	targetRect = targetRect.Intersect(canvas.Bounds())
	if targetRect.Empty() {
		return "", "", fmt.Errorf("target rect empty after clipping")
	}
	xdraw.CatmullRom.Scale(canvas, targetRect, cropped, cropped.Bounds(), draw.Over, nil)

	renderPath := filepath.Join(root, "renders", fmt.Sprintf("%s_%s.png", scenarioID, sizeID))
	if err := savePNG(renderPath, canvas); err != nil {
		return "", "", err
	}

	overlay := annotateSource(src, comp.Result.SourceRect)
	overlayPath := filepath.Join(root, "overlays", fmt.Sprintf("%s_%s_overlay.png", scenarioID, sizeID))
	if err := savePNG(overlayPath, overlay); err != nil {
		return "", "", err
	}

	return renderPath, overlayPath, nil
}

func rectToImageRect(r imagelayout.Rect) image.Rectangle {
	x0 := int(math.Round(r.X))
	y0 := int(math.Round(r.Y))
	x1 := int(math.Round(r.X + r.W))
	y1 := int(math.Round(r.Y + r.H))
	if x1 < x0 {
		x1 = x0
	}
	if y1 < y0 {
		y1 = y0
	}
	return image.Rect(x0, y0, x1, y1)
}

func cropImage(src image.Image, rect imagelayout.Rect) image.Image {
	if src == nil {
		return nil
	}
	srcB := src.Bounds()
	cropRect := rectToImageRect(rect).Intersect(srcB)
	if cropRect.Empty() {
		return src
	}
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if si, ok := src.(subImager); ok {
		return si.SubImage(cropRect)
	}
	out := image.NewRGBA(image.Rect(0, 0, cropRect.Dx(), cropRect.Dy()))
	draw.Draw(out, out.Bounds(), src, cropRect.Min, draw.Src)
	return out
}

func annotateSource(src image.Image, rect imagelayout.Rect) image.Image {
	bounds := src.Bounds()
	annotated := image.NewRGBA(bounds)
	draw.Draw(annotated, bounds, src, bounds.Min, draw.Src)

	overlay := rectToImageRect(rect)
	if overlay.Empty() {
		return annotated
	}

	border := color.RGBA{255, 0, 0, 255}
	drawRectBorder(annotated, overlay, border)
	return annotated
}

func drawRectBorder(img *image.RGBA, rect image.Rectangle, c color.RGBA) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		setRGBA(img, x, rect.Min.Y, c)
		setRGBA(img, x, rect.Max.Y-1, c)
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		setRGBA(img, rect.Min.X, y, c)
		setRGBA(img, rect.Max.X-1, y, c)
	}
}

func setRGBA(img *image.RGBA, x, y int, c color.RGBA) {
	if !image.Pt(x, y).In(img.Bounds()) {
		return
	}
	img.SetRGBA(x, y, c)
}

func marshalPretty(v interface{}) string {
	if v == nil {
		return ""
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r)
}

func writeReport(root string, results []runResult) error {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Scenario.ID == results[j].Scenario.ID {
			return results[i].Size.ID < results[j].Size.ID
		}
		return results[i].Scenario.ID < results[j].Scenario.ID
	})

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>Image Layout Validation Report</title>
<style>
body { font-family: Arial, sans-serif; margin: 2rem; background: #f6f7fb; color: #1f2333; }
section { background: #fff; border-radius: 12px; padding: 1.5rem; margin-bottom: 2rem; box-shadow: 0 10px 40px rgba(15, 23, 42, 0.08); }
h1 { font-size: 2.2rem; margin-bottom: 1rem; }
h2 { font-size: 1.6rem; margin-top: 0; color: #30364a; }
h3 { font-size: 1.2rem; margin-top: 1.5rem; color: #3a4258; }
code { background: #1f2933; color: #f8fafc; padding: 0.2rem 0.4rem; border-radius: 4px; }
pre { background: #0f172a; color: #e2e8f0; padding: 1rem; border-radius: 8px; overflow-x: auto; font-size: 0.85rem; }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); gap: 1rem; }
.badge { display: inline-block; padding: 0.2rem 0.6rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; }
.badge-ok { background: rgba(16, 185, 129, 0.12); color: #047857; }
.badge-err { background: rgba(239, 68, 68, 0.12); color: #b91c1c; }
details { margin-top: 1rem; }
details summary { cursor: pointer; font-weight: 600; }
figure { margin: 0; }
figcaption { margin-top: 0.4rem; font-size: 0.85rem; color: #475569; }
img { max-width: 100%; border-radius: 8px; border: 1px solid rgba(15, 23, 42, 0.1); background: #fff; }
ul { padding-left: 1.2rem; }
</style>
</head>
<body>
<h1>Image Layout Validation Report</h1>
<p>Generated at {{.GeneratedAt}}. Each section exercises the <code>zine-layout imagelayout compute</code> CLI across multiple source image ratios. The rendered output uses the computed viewport to crop and place images, while overlays highlight the crop on the original asset.</p>
{{range .Items}}
<section>
  <h2>{{.Scenario.Name}}</h2>
  <p>{{.Scenario.Description}}</p>
  <div class="grid">
    {{range .Runs}}
    <div>
      <h3>{{.Size.ID | title}}</h3>
      <p><code>{{.Command}}</code></p>
      {{if .Error}}
        <span class="badge badge-err">failed</span>
        <p>{{.Error}}</p>
      {{else}}
        {{if .Validation.Errors}}
          <span class="badge badge-err">validation issues</span>
        {{else}}
          <span class="badge badge-ok">validated</span>
        {{end}}
        <details>
          <summary>Validation</summary>
          <ul>
            {{range .Validation.Messages}}<li>{{.}}</li>{{end}}
            {{range .Validation.Errors}}<li style="color:#b91c1c">{{.}}</li>{{end}}
          </ul>
        </details>
        <details>
          <summary>Trace</summary>
          <pre>{{.TraceJSON}}</pre>
        </details>
        <details>
          <summary>Settings</summary>
          <pre>{{.SettingsJSON}}</pre>
        </details>
        <details>
          <summary>Result</summary>
          <pre>{{.ResultJSON}}</pre>
        </details>
        <figure>
          <img src="{{.SourceImage}}" alt="Source image {{.Scenario.ID}} {{.Size.ID}}" />
          <figcaption>Source ({{.Size.Width}}×{{.Size.Height}} px)</figcaption>
        </figure>
        <figure>
          <img src="{{.OverlayImage}}" alt="Crop overlay {{.Scenario.ID}} {{.Size.ID}}" />
          <figcaption>Crop overlay</figcaption>
        </figure>
        <figure>
          <img src="{{.RenderedImage}}" alt="Rendered output {{.Scenario.ID}} {{.Size.ID}}" />
          <figcaption>Rendered output</figcaption>
        </figure>
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
		Scenario scenario
		Size     sizeSpec
	}

	type templateItem struct {
		Scenario scenario
		Runs     []templateRun
	}

	data := struct {
		GeneratedAt string
		Items       []templateItem
	}{
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	grouped := map[string][]templateRun{}
	for _, res := range results {
		grouped[res.Scenario.ID] = append(grouped[res.Scenario.ID], templateRun{runResult: res, Scenario: res.Scenario, Size: res.Size})
	}

	for _, scRuns := range grouped {
		if len(scRuns) == 0 {
			continue
		}
		sc := scRuns[0].Scenario
		data.Items = append(data.Items, templateItem{Scenario: sc, Runs: scRuns})
	}

	sort.Slice(data.Items, func(i, j int) bool {
		return data.Items[i].Scenario.ID < data.Items[j].Scenario.ID
	})

	funcMap := map[string]interface{}{
		"title": titleCase,
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
	defer func() { _ = f.Close() }()

	return t.Execute(f, data)
}
