package imagelayoutcmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"github.com/go-go-golems/zine-layout/pkg/imagelayout/engine"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func TestLayoutFrameCommand(t *testing.T) {
	specPath := writeTestSpec(t, imagelayout.DefaultLayoutRequest(), 4000, 3000)
	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand: %v", err)
	}
	cmd.SetArgs([]string{"layout", "frame", "--spec", specPath})
	output := executeAndCapture(t, cmd)
	var analysis engine.FrameAnalysis
	if err := json.Unmarshal(bytes_trimSpaceBytes(output), &analysis); err != nil {
		t.Fatalf("decode frame json: %v", err)
	}
	if analysis.TargetRatio <= 0 {
		t.Fatalf("expected positive target ratio, got %f", analysis.TargetRatio)
	}
}

func TestLayoutCropCommand(t *testing.T) {
	layout := imagelayout.DefaultLayoutRequest()
	ratio := 1.0
	layout.Frame.Mode = "ratio"
	layout.Frame.Ratio = &ratio
	layout.Crop.Strategy = "auto"
	specPath := writeTestSpec(t, layout, 3200, 2400)

	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand: %v", err)
	}
	cmd.SetArgs([]string{"layout", "crop", "--spec", specPath})
	output := executeAndCapture(t, cmd)
	var analysis engine.CropAnalysis
	if err := json.Unmarshal(bytes_trimSpaceBytes(output), &analysis); err != nil {
		t.Fatalf("decode crop json: %v", err)
	}
	if analysis.SourceRect.W <= 0 || analysis.SourceRect.H <= 0 {
		t.Fatalf("invalid source rect: %+v", analysis.SourceRect)
	}
}

func TestLayoutFrameWithFlags(t *testing.T) {
	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand: %v", err)
	}
	cmd.SetArgs([]string{
		"layout", "frame",
		"--source-width", "4200",
		"--source-height", "2800",
		"--frame-mode", "ratio",
		"--frame-ratio", "1.5",
	})
	output := executeAndCapture(t, cmd)
	var analysis engine.FrameAnalysis
	if err := json.Unmarshal(bytes_trimSpaceBytes(output), &analysis); err != nil {
		t.Fatalf("decode frame json: %v", err)
	}
	if analysis.TargetRatio <= 0 {
		t.Fatalf("expected positive target ratio")
	}
}

func TestComputeCommandWithFlags(t *testing.T) {
	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand: %v", err)
	}
	cmd.SetArgs([]string{
		"compute",
		"--source-width", "3600",
		"--source-height", "2400",
		"--frame-mode", "page",
		"--frame-page-width-in", "8.5",
		"--frame-page-height-in", "11",
		"--frame-page-dpi", "300",
		"--crop-strategy", "auto",
	})
	output := executeAndCapture(t, cmd)
	var comp imagelayout.Computation
	if err := json.Unmarshal(bytes_trimSpaceBytes(output), &comp); err != nil {
		t.Fatalf("decode computation: %v", err)
	}
	if comp.Result.CanvasRect.W <= 0 || comp.Result.CanvasRect.H <= 0 {
		t.Fatalf("invalid canvas rect: %+v", comp.Result.CanvasRect)
	}
}

func writeTestSpec(t *testing.T, layout imagelayout.LayoutRequest, width, height int) string {
	t.Helper()
	doc := layoutSpecDocument{
		Layout: layout,
	}
	doc.Image.Width = width
	doc.Image.Height = height
	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(specPath, data, 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	return specPath
}

func executeAndCapture(t *testing.T, cmd *cobra.Command) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = w
	os.Stderr = w
	execErr := cmd.Execute()
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy output: %v", err)
	}
	if execErr != nil {
		t.Fatalf("command failed: %v (output: %s)", execErr, buf.String())
	}
	return buf.String()
}

func bytes_trimSpaceBytes(s string) []byte {
	return bytes.TrimSpace([]byte(s))
}
