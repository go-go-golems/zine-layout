# zine-layout

Compose multi-page zines from input images using a simple YAML specification. This tool arranges images into a grid per page, with precise control over margins, borders, and rotation.

Features
- Grid-based composition with per-cell margins
- Page, layout, and inner borders (plain, dotted, dashed, corner)
- 0° and 180° rotation for inputs
- Global margins and PPI configuration
- Multi-document YAML via Emrichen with Sprig functions

Install
- Build: `go build -o ./dist/zine-layout ./cmd/zine-layout`
- Run: `go run ./zine-layout/cmd/zine-layout --help`
- Snapshot release (dev): `PATH=$(pwd)/.bin:$PATH goreleaser release --skip=sign --snapshot --clean`

Quick Start
- Use an example layout: `zine-layout --spec examples/layouts/two_pages_two_inputs.yaml --output-dir out/ img1.png img2.png`
- Or try a test spec: see `examples/tests/*.yaml` for more patterns.

CLI Flags
- `--spec` Path to YAML spec (default `layout.yaml`)
- `--output-dir` Output directory for generated pages
- `--log-level` debug | info | warn | error
- `--ppi` Override Pixels Per Inch specified in the layout
- `--global-border`, `--page-border`, `--layout-border`, `--inner-border` Toggle specific borders
- `--border-type` plain | dotted | dashed | corner
- `--border-color` R,G,B,A (0–255 each)
- `--test` Generate built-in test images instead of reading inputs
- `--test-bw` Use black/white test images
- `--test-dimensions` Specify test image size (e.g., `600px,800px`)

Spec Example
```yaml
global:
  ppi: 300
page_setup:
  margin: { top: 10px, right: 10px, bottom: 10px, left: 10px }
  grid_size: { rows: 1, columns: 2 }
  border: { enabled: true, type: plain, color: "#000000" }
output_pages:
  - id: page-1
    margin: { top: 0px, right: 0px, bottom: 0px, left: 0px }
    border: { enabled: true, type: dotted, color: [180,180,180,255] }
    layout:
      - input_index: 1
        position: { row: 0, column: 0 }
        margin: { top: 5px, right: 5px, bottom: 5px, left: 5px }
      - input_index: 2
        position: { row: 0, column: 1 }
        margin: { top: 5px, right: 5px, bottom: 5px, left: 5px }
```

Examples and DSL
- Examples: `examples/layouts/` and `examples/tests/`
- DSL overview: `doc/dsl.md`
- Units and expressions reference: `pkg/zinelayout/parser/units_doc.md`

Notes
- Input images are indexed starting at 1 in the layout spec.
- Only 0 and 180 rotation are allowed.
- Color accepts hex (e.g., `#000000`), names (e.g., `black`), or `[R,G,B,A]`.
