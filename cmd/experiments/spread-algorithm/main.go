package main

import (
    "context"
    "flag"
    "fmt"
    "image"
    _ "image/jpeg"
    _ "image/png"
    "os"
    "path/filepath"
    "time"

    "github.com/pkg/errors"
    "github.com/rs/zerolog/log"
)

func main() {
    cfgPath := flag.String("config", "spreads.yaml", "Path to YAML config (no image field)")
    outDir := flag.String("out", "out", "Output directory for renders and index.html")
    verbose := flag.Bool("v", false, "Verbose logging to stdout")
    dpiOverride := flag.Float64("dpi", 0, "Override DPI (0 keeps config)")
    fast := flag.Bool("fast", true, "Use faster scaler and PNG speed for testing")
    flag.Parse()

    if flag.NArg() < 1 {
        fmt.Fprintf(os.Stderr, "usage: spread-algorithm [flags] <image-path>\n")
        flag.PrintDefaults()
        os.Exit(2)
    }

    imagePath := flag.Arg(0)

    logger := initLogger(*verbose)
    _ = logger
    log.Info().Str("config", *cfgPath).Str("out", *outDir).Str("image", imagePath).Msg("startup")

    if err := os.MkdirAll(*outDir, 0o755); err != nil {
        panic(errors.Wrap(err, "failed to create output directory"))
    }

    // Load config
    cfg, err := LoadConfig(*cfgPath)
    if err != nil {
        panic(errors.Wrap(err, "failed to load config"))
    }
    log.Info().Int("spreads", len(cfg.Spreads)).Str("version", cfg.Version).Msg("config loaded")

    // Open image
    f, err := os.Open(imagePath)
    if err != nil {
        panic(errors.Wrap(err, "failed to open input image"))
    }
    defer f.Close()

    img, _, err := image.Decode(f)
    if err != nil {
        panic(errors.Wrap(err, "failed to decode input image"))
    }
    bounds := img.Bounds()
    srcW := bounds.Dx()
    srcH := bounds.Dy()
    log.Info().Int("w", srcW).Int("h", srcH).Msg("decoded image")

    // Derive image basename for filenames
    base := filepath.Base(imagePath)
    imageBase := base
    if ext := filepath.Ext(base); ext != "" {
        imageBase = base[:len(base)-len(ext)]
    }

    ctx := context.Background()

    var results []SpreadOutput
    for idx, sp := range cfg.Spreads {
        merged := cfg.MergeWithDefaults(sp)
        if *dpiOverride > 0 {
            merged.Paper.DPI = *dpiOverride
        }

        trace := &Trace{EnableStdout: *verbose, UseZerolog: true}

        trace.logf("[spread %d] name=%s is_spread=%v gutter_in=%.3f paper=%.2fx%.2f in dpi=%.1f",
            idx+1, merged.Name, merged.Spread.IsSpread, merged.Spread.GutterIn, merged.Paper.WidthIn, merged.Paper.HeightIn, merged.Paper.DPI)

        inputs := merged.ToInputs(float64(srcW), float64(srcH))
        trace.logf("[spread %d] computing placement...", idx+1)
        res := ComputePlacement(inputs, trace)

        // Render
        exp := merged.Export
        if exp.Format == "" {
            exp.Format = "png"
        }
        // Force output directory to CLI-provided path so HTML links work reliably
        exp.OutDir = *outDir
        if *fast {
            if exp.PNGLevel == "" { exp.PNGLevel = "speed" }
            if exp.Scaler == "" { exp.Scaler = "fast" }
            exp.ParallelEncode = true
        }

        // common info
        info := RenderInfo{
            Index:          idx + 1,
            SpreadName:     merged.Name,
            ImageBaseName:  imageBase,
            Format:         exp.Format,
            Quality:        exp.Quality,
            Background:     exp.Background,
            FilenameTmpl:   exp.FilenameTemplate,
            OutputDir:      exp.OutDir,
            PNGLevel:       exp.PNGLevel,
            Scaler:         exp.Scaler,
            ParallelEncode: exp.ParallelEncode,
        }

        if !merged.Spread.IsSpread {
            trace.logf("[spread %d] rendering single ...", idx+1)
            filePath, err := RenderSingle(ctx, img, res, info)
            if err != nil {
                panic(errors.Wrapf(err, "render single failed for spread %q", merged.Name))
            }
            log.Info().Int("spread_index", idx+1).Str("file", filePath).Msg("wrote single")
            trace.logf("[spread %d] wrote %s", idx+1, filePath)
            results = append(results, SpreadOutput{
                Name:      merged.Name,
                PanelFiles: []string{filePath},
                PanelLabels: []string{"single"},
                Logs:      trace.Lines,
                Timestamp: time.Now(),
            })
        } else {
            trace.logf("[spread %d] rendering spread (left/right) ...", idx+1)
            leftPath, rightPath, err := RenderSpread(ctx, img, res, info)
            if err != nil {
                panic(errors.Wrapf(err, "render spread failed for spread %q", merged.Name))
            }
            log.Info().Int("spread_index", idx+1).Str("file", leftPath).Msg("wrote left")
            log.Info().Int("spread_index", idx+1).Str("file", rightPath).Msg("wrote right")
            trace.logf("[spread %d] wrote %s", idx+1, leftPath)
            trace.logf("[spread %d] wrote %s", idx+1, rightPath)
            results = append(results, SpreadOutput{
                Name:       merged.Name,
                PanelFiles: []string{leftPath, rightPath},
                PanelLabels: []string{"left", "right"},
                Logs:       trace.Lines,
                Timestamp:  time.Now(),
            })
        }
    }

    // Write HTML index
    htmlPath := filepath.Join(*outDir, "index.html")
    log.Info().Str("file", htmlPath).Msg("writing HTML index")
    if err := WriteHTMLIndex(htmlPath, results); err != nil {
        panic(errors.Wrap(err, "failed to write HTML index"))
    }

    log.Info().Int("spreads", len(results)).Str("index", htmlPath).Msg("done")
}


