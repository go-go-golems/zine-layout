package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/zine-layout/pkg/spread"
	simplepkg "github.com/go-go-golems/zine-layout/pkg/spread/simple"
	"github.com/go-go-golems/zine-layout/pkg/spread/sonnet/config"
	"github.com/go-go-golems/zine-layout/pkg/spread/sonnet/namer"
)

func init() {
	console := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	log.Logger = log.Output(console)
}

func main() {
	root := &cobra.Command{
		Use:   "spread-cli",
		Short: "Compute and render image spreads from Sonnet YAML configs using the simple algorithm",
	}

	root.AddCommand(newComputeCommand())
	root.AddCommand(newRenderCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// --- compute command ---

type computeReport struct {
	Index  int              `json:"index"`
	Name   string           `json:"name"`
	Asset  string           `json:"asset,omitempty"`
	Result simplepkg.Result `json:"result"`
	Trace  []string         `json:"trace,omitempty"`
}

func newComputeCommand() *cobra.Command {
	var printMode string
	var only []string
	cmd := &cobra.Command{
		Use:   "compute <config.yaml>",
		Args:  cobra.ExactArgs(1),
		Short: "Compute spread geometry using the simple algorithm",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			specs, err := loadSpecs(configPath)
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
				meta, err := readImageMeta(spec.ImagePath)
				if err != nil {
					return fmt.Errorf("meta(%s): %w", spec.ImagePath, err)
				}
				inputs, err := simplepkg.InputsFromResolved(spec.Settings, meta)
				if err != nil {
					return fmt.Errorf("inputs(%s): %w", spec.Name, err)
				}
				trace := &simplepkg.Trace{UseZerolog: true}
				result := simplepkg.ComputePlacement(inputs, trace)
				reports = append(reports, computeReport{
					Index:  i + 1,
					Name:   spec.Name,
					Asset:  spec.AssetName,
					Result: result,
					Trace:  append([]string(nil), trace.Lines...),
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

// --- render command ---

type simpleOutput struct {
	Panel string `json:"panel"`
	Path  string `json:"path"`
}

type renderReport struct {
	Index   int              `json:"index"`
	Name    string           `json:"name"`
	Asset   string           `json:"asset,omitempty"`
	Result  simplepkg.Result `json:"result"`
	Outputs []simpleOutput   `json:"outputs"`
	Trace   []string         `json:"trace,omitempty"`
}

func newRenderCommand() *cobra.Command {
	var printMode string
	var only []string
	var dryRun bool
	var fast bool
	cmd := &cobra.Command{
		Use:   "render <config.yaml>",
		Args:  cobra.ExactArgs(1),
		Short: "Render spreads to disk using the simple algorithm",
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

			ctx := context.Background()
			reports := make([]renderReport, 0, len(specs))
			filenameSet := map[string]struct{}{}

			for i, spec := range specs {
				meta, err := readImageMeta(spec.ImagePath)
				if err != nil {
					return fmt.Errorf("meta(%s): %w", spec.ImagePath, err)
				}
				inputs, err := simplepkg.InputsFromResolved(spec.Settings, meta)
				if err != nil {
					return fmt.Errorf("inputs(%s): %w", spec.Name, err)
				}
				trace := &simplepkg.Trace{UseZerolog: true}
				result := simplepkg.ComputePlacement(inputs, trace)

				outPaths, absOutDir, err := planOutputs(spec, i+1, baseDir)
				if err != nil {
					return err
				}

				rep := renderReport{
					Index:  i + 1,
					Name:   spec.Name,
					Asset:  spec.AssetName,
					Result: result,
					Trace:  append([]string(nil), trace.Lines...),
				}

				for _, p := range outPaths {
					filenameSet[p] = struct{}{}
				}

				if dryRun {
					panels := sortedPanels(outPaths)
					for _, panel := range panels {
						rep.Outputs = append(rep.Outputs, simpleOutput{Panel: panel, Path: outPaths[panel]})
					}
					reports = append(reports, rep)
					continue
				}

				img, err := loadImage(spec.ImagePath)
				if err != nil {
					return fmt.Errorf("load(%s): %w", spec.ImagePath, err)
				}

				opts := simplepkg.RenderOptions{}
				if fast {
					opts.PNGLevel = "speed"
					opts.Scaler = "fast"
					opts.ParallelEncode = true
				}

				imageBase := strings.TrimSuffix(filepath.Base(spec.ImagePath), filepath.Ext(spec.ImagePath))
				info := simplepkg.RenderInfoFromExport(spec.Settings.Export, i+1, spec.Name, imageBase, opts)
				info.OutputDir = absOutDir
				info.PathOverrides = outPaths

				if !spec.Settings.IsSpread {
					path, err := simplepkg.RenderSingle(ctx, img, result, info)
					if err != nil {
						return fmt.Errorf("render single(%s): %w", spec.Name, err)
					}
					rep.Outputs = append(rep.Outputs, simpleOutput{Panel: "single", Path: path})
				} else {
					leftPath, rightPath, err := simplepkg.RenderSpread(ctx, img, result, info)
					if err != nil {
						return fmt.Errorf("render spread(%s): %w", spec.Name, err)
					}
					rep.Outputs = append(rep.Outputs,
						simpleOutput{Panel: "left", Path: leftPath},
						simpleOutput{Panel: "right", Path: rightPath},
					)
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
	cmd.Flags().BoolVar(&fast, "fast", false, "Use faster render settings (speed PNG, fast scaler, parallel encode)")
	cmd.Flags().StringVar(&printMode, "print", "json", "Output mode: json|filenames|none")
	cmd.Flags().StringSliceVar(&only, "only", nil, "Filter spreads (name=<value>, asset=<value>)")
	return cmd
}

// --- shared helpers ---

type filterSpec struct {
	Key   string
	Value string
}

func loadSpecs(path string) ([]config.SpreadSpec, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Dir(path)
	return cfg.Resolve(baseDir)
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

func readImageMeta(path string) (spread.ImageMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return spread.ImageMeta{}, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return spread.ImageMeta{}, err
	}
	return spread.ImageMeta{Width: cfg.Width, Height: cfg.Height}, nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func planOutputs(spec config.SpreadSpec, index int, baseDir string) (map[string]string, string, error) {
	export := spec.Settings.Export
	outDir := strings.TrimSpace(export.OutDir)
	if outDir == "" {
		outDir = "./out"
	}
	absOutDir := outDir
	if !filepath.IsAbs(absOutDir) {
		absOutDir = filepath.Join(baseDir, absOutDir)
	}
	if err := os.MkdirAll(absOutDir, 0o755); err != nil {
		return nil, "", err
	}

	format := normalizeFormat(export.Format)
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
			return nil, "", err
		}
		paths["single"] = filepath.Join(absOutDir, filename)
		return paths, absOutDir, nil
	}

	for _, panel := range []string{"left", "right"} {
		data.Panel = panel
		filename, err := namer.Build(export.FilenameTemplate, data)
		if err != nil {
			return nil, "", err
		}
		paths[panel] = filepath.Join(absOutDir, filename)
	}
	return paths, absOutDir, nil
}

func sortedPanels(paths map[string]string) []string {
	panels := make([]string, 0, len(paths))
	for panel := range paths {
		panels = append(panels, panel)
	}
	sort.Strings(panels)
	return panels
}

func normalizeFormat(format string) string {
	return strings.ToLower(strings.TrimSpace(format))
}
