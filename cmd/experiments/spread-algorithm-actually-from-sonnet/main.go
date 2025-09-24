package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

    "github.com/go-go-golems/zine-layout/pkg/spread/sonnet/config"
    "github.com/go-go-golems/zine-layout/pkg/spread/sonnet/engine"
    "github.com/go-go-golems/zine-layout/pkg/spread/sonnet/gallery"
    "github.com/go-go-golems/zine-layout/pkg/spread/sonnet/media"
    "github.com/go-go-golems/zine-layout/pkg/spread/sonnet/namer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type filterSpec struct {
	Key   string
	Value string
}

type computeReport struct {
	Index  int           `json:"index"`
	Asset  string        `json:"asset,omitempty"`
	Result engine.Result `json:"result"`
}

type renderReport struct {
	Index   int                 `json:"index"`
	Asset   string              `json:"asset,omitempty"`
	Result  engine.Result       `json:"result"`
	Outputs []engine.OutputFile `json:"outputs"`
	Trace   []engine.TraceEntry `json:"trace"`
}

func init() {
	console := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	log.Logger = log.Output(console)
}

func main() {
	root := &cobra.Command{
		Use:   "spread-cli",
		Short: "Compute and render image spreads from YAML configs",
	}

	root.AddCommand(newComputeCommand())
	root.AddCommand(newRenderCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newComputeCommand() *cobra.Command {
	var printMode string
	var only []string
	cmd := &cobra.Command{
		Use:   "compute <config.yaml>",
		Args:  cobra.ExactArgs(1),
		Short: "Compute spread geometry and emit JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			baseDir := filepath.Dir(configPath)
			specs, err := cfg.Resolve(baseDir)
			if err != nil {
				return err
			}
			filters, err := parseFilters(only)
			if err != nil {
				return err
			}
			specs = applyFilters(specs, filters)
			reports := make([]computeReport, 0, len(specs))
			for i, spec := range specs {
				meta, err := media.Metadata(spec.ImagePath)
				if err != nil {
					return fmt.Errorf("meta(%s): %w", spec.ImagePath, err)
				}
				result, err := engine.Compute(spec, meta)
				if err != nil {
					return fmt.Errorf("compute(%s): %w", spec.Name, err)
				}
				reports = append(reports, computeReport{
					Index:  i + 1,
					Asset:  spec.AssetName,
					Result: result,
				})
			}
			switch strings.ToLower(printMode) {
			case "", "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(reports)
			default:
				return fmt.Errorf("unknown print mode %q", printMode)
			}
		},
	}
	cmd.Flags().StringVar(&printMode, "print", "json", "Output format: json")
	cmd.Flags().StringSliceVar(&only, "only", nil, "Filter spreads (name=<value>, asset=<value>)")
	return cmd
}

func newRenderCommand() *cobra.Command {
	var printMode string
	var only []string
	var dryRun bool
	var htmlPath string
	cmd := &cobra.Command{
		Use:   "render <config.yaml>",
		Args:  cobra.ExactArgs(1),
		Short: "Render spreads to images and build an HTML gallery",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			baseDir := filepath.Dir(configPath)
			specs, err := cfg.Resolve(baseDir)
			if err != nil {
				return err
			}
			filters, err := parseFilters(only)
			if err != nil {
				return err
			}
			specs = applyFilters(specs, filters)
			reports := make([]renderReport, 0, len(specs))
			galleryEntries := make([]gallery.Entry, 0)
			index := 1
			for _, spec := range specs {
				meta, err := media.Metadata(spec.ImagePath)
				if err != nil {
					return fmt.Errorf("meta(%s): %w", spec.ImagePath, err)
				}
				result, err := engine.Compute(spec, meta)
				if err != nil {
					return fmt.Errorf("compute(%s): %w", spec.Name, err)
				}
				outPaths, err := planOutputs(result, spec, index, baseDir)
				if err != nil {
					return err
				}
				combinedTrace := append([]engine.TraceEntry{}, result.Trace...)
				renderTrace := make([]engine.TraceEntry, 0)
				renderTracer := engine.NewTracer(spec.Name, "render", &renderTrace)
				var outputs []engine.OutputFile
				if !dryRun {
					img, err := media.Load(spec.ImagePath)
					if err != nil {
						return fmt.Errorf("load(%s): %w", spec.ImagePath, err)
					}
					bounds := img.Bounds()
					renderTracer.Log("decode", "decoded image %dx%d px", bounds.Dx(), bounds.Dy())
					renderedOutputs, engineTrace, err := engine.Render(result, img, outPaths)
					if err != nil {
						return err
					}
					outputs = renderedOutputs
					renderTrace = append(renderTrace, engineTrace...)
				} else {
					keys := make([]string, 0, len(outPaths))
					for k := range outPaths {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						path := outPaths[k]
						outputs = append(outputs, engine.OutputFile{Panel: k, Path: path})
						renderTracer.Log("dry-run", "would write %s panel to %s", k, path)
					}
					renderTracer.Log("dry-run", "no files written")
				}
				combinedTrace = append(combinedTrace, renderTrace...)
				if !dryRun {
					formattedTrace := formatTrace(combinedTrace)
					for _, out := range outputs {
						rel, _ := filepath.Rel(filepath.Dir(htmlPathOrDefault(htmlPath, baseDir)), out.Path)
						galleryEntries = append(galleryEntries, gallery.Entry{
							Index:    index,
							Spread:   spec.Name,
							Panel:    out.Panel,
							RelPath:  filepath.ToSlash(rel),
							FileName: filepath.Base(out.Path),
							Trace:    formattedTrace,
						})
					}
				}
				reports = append(reports, renderReport{
					Index:   index,
					Asset:   spec.AssetName,
					Result:  result,
					Outputs: outputs,
					Trace:   combinedTrace,
				})
				index++
			}

			switch strings.ToLower(printMode) {
			case "":
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(reports); err != nil {
					return err
				}
			case "filenames":
				for _, report := range reports {
					for _, out := range report.Outputs {
						fmt.Println(out.Path)
					}
				}
			case "none":
				// no-op
			default:
				return fmt.Errorf("unknown print mode %q", printMode)
			}

			if !dryRun {
				galleryPath := htmlPathOrDefault(htmlPath, baseDir)
				if len(galleryEntries) > 0 {
					if err := gallery.Write(galleryPath, "Spread Gallery", galleryEntries); err != nil {
						return err
					}
					if strings.ToLower(printMode) != "none" {
						fmt.Fprintf(os.Stderr, "gallery written to %s\n", galleryPath)
					}
				}
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Compute outputs without writing files")
	cmd.Flags().StringVar(&printMode, "print", "filenames", "Output mode: filenames|json|none")
	cmd.Flags().StringSliceVar(&only, "only", nil, "Filter spreads (name=<value>, asset=<value>)")
	cmd.Flags().StringVar(&htmlPath, "html", "", "Optional gallery HTML output path")
	return cmd
}

func parseFilters(raw []string) ([]filterSpec, error) {
	filters := make([]filterSpec, 0, len(raw))
	for _, item := range raw {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid filter %q", item)
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		if key != "name" && key != "asset" {
			return nil, fmt.Errorf("unsupported filter key %q", key)
		}
		filters = append(filters, filterSpec{Key: key, Value: strings.TrimSpace(parts[1])})
	}
	return filters, nil
}

func applyFilters(specs []config.SpreadSpec, filters []filterSpec) []config.SpreadSpec {
	if len(filters) == 0 {
		return specs
	}
	out := make([]config.SpreadSpec, 0, len(specs))
	for _, spec := range specs {
		if matchesFilters(spec, filters) {
			out = append(out, spec)
		}
	}
	return out
}

func matchesFilters(spec config.SpreadSpec, filters []filterSpec) bool {
	for _, f := range filters {
		switch f.Key {
		case "name":
			if spec.Name != f.Value {
				return false
			}
		case "asset":
			if spec.AssetName != f.Value {
				return false
			}
		}
	}
	return true
}

func planOutputs(result engine.Result, spec config.SpreadSpec, index int, baseDir string) (map[string]string, error) {
	export := spec.Settings.Export
	outDir := export.OutDir
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(baseDir, outDir)
	}
	imageBase := strings.TrimSuffix(filepath.Base(spec.ImagePath), filepath.Ext(spec.ImagePath))
	ext := export.Format
	data := namer.TemplateData{
		Index:         index,
		Name:          spec.Name,
		ImageBasename: imageBase,
		Extension:     ext,
	}
	paths := map[string]string{}
	if !spec.Settings.IsSpread {
		data.Panel = "single"
		filename, err := namer.Build(export.FilenameTemplate, data)
		if err != nil {
			return nil, err
		}
		paths["single"] = filepath.Join(outDir, filename)
		return paths, nil
	}
	panels := []string{"left", "right"}
	for _, panel := range panels {
		data.Panel = panel
		filename, err := namer.Build(export.FilenameTemplate, data)
		if err != nil {
			return nil, err
		}
		paths[panel] = filepath.Join(outDir, filename)
	}
	return paths, nil
}

func htmlPathOrDefault(path string, baseDir string) string {
	if strings.TrimSpace(path) == "" {
		return filepath.Join(baseDir, "out", "gallery.html")
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}

func formatTrace(entries []engine.TraceEntry) []string {
	formatted := make([]string, len(entries))
	for i, entry := range entries {
		if entry.Stage != "" {
			formatted[i] = fmt.Sprintf("[%s] %s", entry.Stage, entry.Message)
		} else {
			formatted[i] = entry.Message
		}
	}
	return formatted
}
