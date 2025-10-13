# Developer Navigation Index

This index highlights the primary entry points for working on image layout behaviour, rendering, and validation utilities. Each section links to the canonical Go packages, CLI verbs, and support scripts so a new contributor can quickly explore the codebase.

## Image Layout CLI (`cmd/zine-layout`)
- **`cmd/zine-layout/main.go`** – Registers the Cobra root command and wires sub-commands, including `imagelayout`, `render`, and project workflows.
- **`cmd/zine-layout/cmds/imagelayout/command.go`** – Declares the `imagelayout` verb group.
- **`cmd/zine-layout/cmds/imagelayout/compute.go`** – Implements `zine-layout imagelayout compute`, mapping CLI flags to `imagelayout.ViewportSettings`, invoking the engine, and printing a JSON payload containing the settings, result, and trace data.

## Image Layout Engine (`pkg/imagelayout`)
- **`pkg/imagelayout/types.go`** – Shared data structures for viewport settings, computation results, and trace steps used by both CLI and services.
- **`pkg/imagelayout/defaults.go`** – Provides `DefaultSettings()` used whenever a CLI call or preset omits fields.
- **`pkg/imagelayout/engine/engine.go`** – Core placement logic. `InputsFromSettings` normalises units, DPI, orientation, crop/fit constraints, and focus data; `ComputeViewport` returns the source/target rectangles and a diagnostic trace.
- **`pkg/imagelayout/engine/engine_test.go`** – Algorithm tests that exercise contain/cover, crop ratios, fit modes, and focus positioning.

## Page Composition (`pkg/pagelayout`)
- **`pkg/pagelayout/settings.go`** – Page-level sizing helpers (content rectangles, margins, spreads, border metadata).
- **`pkg/pagelayout/renderer/renderer.go`** – Renders laid-out pages. Accepts an optional `imagelayout.ViewportResult` to crop images before scaling into the page content area and produces multiple variants (full, thumbnail, left/right spreads).

## Validation & QA Utilities (`scripts`)
- **`scripts/imagelayout_validation/main.go`** – Generates synthetic test images, invokes `zine-layout imagelayout compute` across page, crop, and fit templates, validates geometry against the derived engine inputs, renders the resulting placements, and emits a rich HTML report (stored under `ttmp/<date>/imagelayout-validation/index.html`). Useful for manual inspection of algorithm behaviour.
- **`scripts/pagelayout_validation/main.go`** – Builds representative zine layout specs, synthesises multi-image inputs across portrait, landscape, and square ratios, shells out to `zine-layout render`, verifies the generated page dimensions against library computations, and assembles an HTML dashboard (under `ttmp/<date>/pagelayout-validation/index.html`) with CLI traces, diagnostics, and rendered pages.

## Temporary Artifacts (`ttmp/<date>`)
- Daily folders collect generated reports, screenshots, and debugging notes. The validation scripts write inputs (`assets/`), overlays or diagnostics, render outputs (`renders/`), and `index.html` into `ttmp/<date>/imagelayout-validation/` and `ttmp/<date>/pagelayout-validation/`.

## How to Explore Quickly
1. Start with the CLI command to understand accepted flags (`cmd/zine-layout/cmds/imagelayout/compute.go`).
2. Follow the call into the engine (`pkg/imagelayout/engine/engine.go`) for the maths behind mode-specific behaviour.
3. Use the validation scripts to reproduce scenarios or extend coverage (`go run ./scripts/imagelayout_validation` for viewport maths, `go run ./scripts/pagelayout_validation` for full pages).
4. For end-to-end rendering on pages, review `pkg/pagelayout/renderer/renderer.go` alongside the page validation harness outputs.

## Helpful Commands
```bash
# Run the validation harness and produce the HTML report
go run ./scripts/imagelayout_validation

# Exercise page layout rendering scenarios and build the dashboard
go run ./scripts/pagelayout_validation

# Execute the existing Go unit tests
go test ./...
```
