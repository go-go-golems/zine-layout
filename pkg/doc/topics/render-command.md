---
Title: Render Command
Slug: render
Short: Render output pages from a layout spec and input images.
Topics:
- zine-layout
Commands:
- render
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Render Command

The `render` command composes one or more output pages from input images according to a YAML layout specification. It supports per-cell margins, page/layout/inner borders, and unit-aware sizing. Use this to generate printable sheets for zines and booklets.

## Usage

```bash
zine-layout render --spec layout.yaml --output-dir out/ img1.png img2.png
```

Common flags:
- `--spec` Path to YAML spec (default `layout.yaml`)
- `--output-dir` Output directory for PNG files
- `--ppi` Override Pixels Per Inch from the spec
- `--global-border`, `--page-border`, `--layout-border`, `--inner-border` Toggle borders
- `--border-type` plain | dotted | dashed | corner
- `--border-color` R,G,B,A or `#hex` or color name
- `--test`, `--test-bw`, `--test-dimensions` Generate synthetic inputs

## Examples

```bash
# Two inputs on a single page
zine-layout render \
  --spec examples/layouts/two_pages_two_inputs.yaml \
  --output-dir out/ \
  img-1.png img-2.png

# Test images with borders
zine-layout render --spec layout.yaml --layout-border --test --test-dimensions 600px,800px
```

For the layout specification format, see:

```
glaze help zine-layout-dsl
```

