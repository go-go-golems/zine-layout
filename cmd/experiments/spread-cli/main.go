package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	simplepkg "github.com/go-go-golems/zine-layout/pkg/spread/simple"
	"github.com/go-go-golems/zine-layout/pkg/spread/sonnet/config"
	"github.com/go-go-golems/zine-layout/pkg/spread/sonnet/engine"
	"github.com/go-go-golems/zine-layout/pkg/spread/sonnet/media"
	"github.com/go-go-golems/zine-layout/pkg/spread/sonnet/namer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	console := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	log.Logger = log.Output(console)
}

func main() {
	root := &cobra.Command{
		Use:   "spread-cli",
		Short: "Compute and render image spreads from Sonnet YAML configs using different algorithms",
	}

	root.AddCommand(newComputeCommand())
	root.AddCommand(newRenderCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type algorithmMode string

const (
	algoSonnet algorithmMode = "sonnet"
	algoSimple algorithmMode = "simple"
	algoBoth   algorithmMode = "both"
)

func parseAlgorithmMode(raw string) (algorithmMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "both":
		return algoBoth, nil
	case "sonnet":
		return algoSonnet, nil
	case "simple":
		return algoSimple, nil
	default:
		return "", fmt.Errorf("unknown algorithm %q (expected sonnet|simple|both)", raw)
	}
}

func (m algorithmMode) includesSonnet() bool {
	return m == algoSonnet || m == algoBoth
}

func (m algorithmMode) includesSimple() bool {
	return m == algoSimple || m == algoBoth
}

// --- compute command ---

type computeReport struct {
	Index  int                  `json:"index"`
	Name   string               `json:"name"`
	Asset  string               `json:"asset,omitempty"`
	Sonnet *engine.Result       `json:"sonnet,omitempty"`
	Simple *simpleComputeResult `json:"simple,omitempty"`
}

type simpleComputeResult struct {
	Result simplepkg.Result `json:"result"`
	Trace  []string         `json:"trace,omitempty"`
}

func newComputeCommand() *cobra.Command {
	var printMode string
	var only []string
	var algoFlag string
	cmd := &cobra.Command{
		Use:   "compute <config.yaml>",
		Args:  cobra.ExactArgs(1),
		Short: "Compute spread geometry with the selected algorithm(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, err := parseAlgorithmMode(algoFlag)
			if err != nil {
				return err
			}

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

				rep := computeReport{Index: i + 1, Name: spec.Name, Asset: spec.AssetName}

				if mode.includesSonnet() {
					res, err := engine.Compute(spec, meta)
					if err != nil {
						return fmt.Errorf("compute sonnet(%s): %w", spec.Name, err)
					}
					resCopy := res
					rep.Sonnet = &resCopy
				}

				if mode.includesSimple() {
					inputs, err := simplepkg.InputsFromResolved(spec.Settings, meta)
					if err != nil {
						return fmt.Errorf("inputs simple(%s): %w", spec.Name, err)
					}
					trace := &simplepkg.Trace{UseZerolog: true}
					simpleRes := simplepkg.ComputePlacement(inputs, trace)
					rep.Simple = &simpleComputeResult{Result: simpleRes, Trace: append([]string{}, trace.Lines...)}
				}

				reports = append(reports, rep)
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
	cmd.Flags().StringVar(&algoFlag, "algo", "both", "Algorithm to run: sonnet|simple|both")
	return cmd
}

// --- render command ---

type renderReport struct {
	Index  int                 `json:"index"`
	Name   string              `json:"name"`
	Asset  string              `json:"asset,omitempty"`
	Sonnet *sonnetRenderReport `json:"sonnet,omitempty"`
	Simple *simpleRenderReport `json:"simple,omitempty"`
}

type sonnetRenderReport struct {
	Result  engine.Result       `json:"result"`
	Outputs []engine.OutputFile `json:"outputs"`
	Trace   []engine.TraceEntry `json:"trace"`
}

type simpleRenderReport struct {
	Result  simplepkg.Result `json:"result"`
	Outputs []simpleOutput   `json:"outputs"`
	Trace   []string         `json:"trace"`
}

type simpleOutput struct {
	Panel string `json:"panel"`
	Path  string `json:"path"`
}

func newRenderCommand() *cobra.Command {
	var printMode string
	var only []string
	var algoFlag string
	var dryRun bool
	var simpleFast bool
	cmd := &cobra.Command{
		Use:   "render <config.yaml>",
		Args:  cobra.ExactArgs(1),
		Short: "Render spreads to disk using the selected algorithm(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, err := parseAlgorithmMode(algoFlag)
			if err != nil {
				return err
			}

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
			filenameSet := map[string]struct{}{}

			for i, spec := range specs {
				meta, err := media.Metadata(spec.ImagePath)
				if err != nil {
					return fmt.Errorf("meta(%s): %w", spec.ImagePath, err)
				}

				outPaths, absOutDir, imageBase, err := planOutputPaths(spec, i+1, baseDir)
				if err != nil {
					return err
				}

				rep := renderReport{Index: i + 1, Name: spec.Name, Asset: spec.AssetName}
				var img image.Image
				var loadErr error
				needImage := !dryRun && (mode.includesSonnet() || mode.includesSimple())
				if needImage {
					img, loadErr = media.Load(spec.ImagePath)
					if loadErr != nil {
						return fmt.Errorf("load(%s): %w", spec.ImagePath, loadErr)
					}
				}
				ctx := context.Background()

				if mode.includesSonnet() {
					res, err := engine.Compute(spec, meta)
					if err != nil {
						return fmt.Errorf("compute sonnet(%s): %w", spec.Name, err)
					}
					sonnetRep := sonnetRenderReport{Result: res}
					if dryRun {
						panels := sortedPanels(outPaths)
						outputs := make([]engine.OutputFile, 0, len(panels))
						for _, panel := range panels {
							outputs = append(outputs, engine.OutputFile{Panel: panel, Path: outPaths[panel]})
							filenameSet[outPaths[panel]] = struct{}{}
						}
						sonnetRep.Outputs = outputs
						sonnetRep.Trace = append([]engine.TraceEntry{}, res.Trace...)
					} else {
						outputs, renderTrace, err := engine.Render(res, img, outPaths)
						if err != nil {
							return fmt.Errorf("render sonnet(%s): %w", spec.Name, err)
						}
						sonnetRep.Outputs = outputs
						sonnetRep.Trace = append([]engine.TraceEntry{}, res.Trace...)
						sonnetRep.Trace = append(sonnetRep.Trace, renderTrace...)
						for _, out := range outputs {
							filenameSet[out.Path] = struct{}{}
						}
					}
					rep.Sonnet = &sonnetRep
				}

				if mode.includesSimple() {
					inputs, err := simplepkg.InputsFromResolved(spec.Settings, meta)
					if err != nil {
						return fmt.Errorf("inputs simple(%s): %w", spec.Name, err)
					}
					trace := &simplepkg.Trace{UseZerolog: true}
					simpleRes := simplepkg.ComputePlacement(inputs, trace)

					opts := simplepkg.RenderOptions{}
					if simpleFast {
						opts.PNGLevel = "speed"
						opts.Scaler = "fast"
						opts.ParallelEncode = true
					}
					info := simplepkg.RenderInfoFromExport(spec.Settings.Export, i+1, spec.Name, imageBase, opts)
					info.OutputDir = absOutDir
					info.PathOverrides = outPaths

					simpleRep := simpleRenderReport{Result: simpleRes, Trace: append([]string{}, trace.Lines...)}
					if dryRun {
						panels := sortedPanels(outPaths)
						outputs := make([]simpleOutput, 0, len(panels))
						for _, panel := range panels {
							outputs = append(outputs, simpleOutput{Panel: panel, Path: outPaths[panel]})
							filenameSet[outPaths[panel]] = struct{}{}
						}
						simpleRep.Outputs = outputs
					} else {
						if !spec.Settings.IsSpread {
							path, err := simplepkg.RenderSingle(ctx, img, simpleRes, info)
							if err != nil {
								return fmt.Errorf("render simple(%s): %w", spec.Name, err)
							}
							simpleRep.Outputs = append(simpleRep.Outputs, simpleOutput{Panel: "single", Path: path})
							filenameSet[path] = struct{}{}
						} else {
							left, right, err := simplepkg.RenderSpread(ctx, img, simpleRes, info)
							if err != nil {
								return fmt.Errorf("render simple(%s): %w", spec.Name, err)
							}
							simpleRep.Outputs = append(simpleRep.Outputs,
								simpleOutput{Panel: "left", Path: left},
								simpleOutput{Panel: "right", Path: right},
							)
							filenameSet[left] = struct{}{}
							filenameSet[right] = struct{}{}
						}
					}
					rep.Simple = &simpleRep
				}

				reports = append(reports, rep)
			}

			switch strings.ToLower(printMode) {
			case "", "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(reports)
			case "filenames":
				names := make([]string, 0, len(filenameSet))
				for name := range filenameSet {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					fmt.Println(name)
				}
				return nil
			case "none":
				return nil
			default:
				return fmt.Errorf("unknown print mode %q", printMode)
			}
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Plan outputs without writing files")
	cmd.Flags().StringVar(&printMode, "print", "json", "Output mode: json|filenames|none")
	cmd.Flags().StringSliceVar(&only, "only", nil, "Filter spreads (name=<value>, asset=<value>)")
	cmd.Flags().StringVar(&algoFlag, "algo", "both", "Algorithm to run: sonnet|simple|both")
	cmd.Flags().BoolVar(&simpleFast, "simple-fast", false, "Use fast render settings for the simple algorithm (PNG speed, fast scaler, parallel encode)")
	return cmd
}

// --- shared helpers ---

type filterSpec struct {
	Key   string
	Value string
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

func planOutputPaths(spec config.SpreadSpec, index int, baseDir string) (map[string]string, string, string, error) {
	export := spec.Settings.Export
	outDir := strings.TrimSpace(export.OutDir)
	if outDir == "" {
		outDir = "./out"
	}
	absOutDir := outDir
	if !filepath.IsAbs(absOutDir) {
		absOutDir = filepath.Join(baseDir, absOutDir)
	}

	format := strings.ToLower(strings.TrimSpace(export.Format))
	if format == "" {
		format = "png"
	}
	if format == "jpeg" {
		format = "jpg"
	}

	imageBase := strings.TrimSuffix(filepath.Base(spec.ImagePath), filepath.Ext(spec.ImagePath))
	data := namer.TemplateData{
		Index:         index,
		Name:          spec.Name,
		ImageBasename: imageBase,
		Extension:     format,
	}

	paths := map[string]string{}
	if !spec.Settings.IsSpread {
		data.Panel = "single"
		filename, err := namer.Build(export.FilenameTemplate, data)
		if err != nil {
			return nil, "", "", err
		}
		paths["single"] = filepath.Join(absOutDir, filename)
		return paths, absOutDir, imageBase, nil
	}
	panels := []string{"left", "right"}
	for _, panel := range panels {
		data.Panel = panel
		filename, err := namer.Build(export.FilenameTemplate, data)
		if err != nil {
			return nil, "", "", err
		}
		paths[panel] = filepath.Join(absOutDir, filename)
	}
	return paths, absOutDir, imageBase, nil
}

func sortedPanels(paths map[string]string) []string {
	panels := make([]string, 0, len(paths))
	for panel := range paths {
		panels = append(panels, panel)
	}
	sort.Strings(panels)
	return panels
}
