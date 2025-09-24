## Goal

Remove the Sonnet algorithm and configuration pipeline from the codebase, keep only the Simple algorithm, and make Simple the default everywhere (server, CLI, web UI).

## Summary of changes

- Delete all code under `zine-layout/pkg/spread/sonnet/` and references to it
- Simplify adapters so Simple does not depend on Sonnet config types
- Keep YAML endpoints; migrate them to a Simple YAML format and add a whole-book YAML export
- Update REST handlers to accept only Simple settings and compute using Simple
- Adjust CLI verbs (if any experimental ones reference Sonnet)
- Update web API client and UI to remove Sonnet-specific paths and fields

---

## Go code to delete

Delete the entire Sonnet subtree:
- `zine-layout/pkg/spread/sonnet/` including subpackages `config`, `engine`, `media`, `gallery`, `namer`

Files to remove:
- `zine-layout/pkg/spread/sonnet/config/*.go`
- `zine-layout/pkg/spread/sonnet/engine/*.go`
- `zine-layout/pkg/spread/sonnet/media/*.go`
- `zine-layout/pkg/spread/sonnet/gallery/*.go`
- `zine-layout/pkg/spread/sonnet/namer/*.go`

Also remove Sonnet adapters. Keep and refactor YAML helpers for Simple:
- `zine-layout/pkg/spread/simple/adapter.go` (replace with Simple-native inputs builder)
- Keep `zine-layout/pkg/spread/yaml.go` and refactor it to emit Simple YAML and add book export helpers

Adjust remaining Simple files to use Simple-native `Inputs` without Sonnet cfg:
- `zine-layout/pkg/spread/simple/algorithm.go`
- `zine-layout/pkg/spread/simple/render.go`
- `zine-layout/pkg/spread/simple/html.go` (if it uses Sonnet types for docs)

Remove any experimental Sonnet CLI:
- `zine-layout/cmd/experiments/spread-cli/main.go` (references Sonnet config)

## REST APIs to adjust

References in `zine-layout/pkg/serve/server.go`:
- Remove imports of `github.com/go-go-golems/zine-layout/pkg/spread/sonnet/config`
- Remove `toResolved(...)` function and replace with Simple-native input assembly
- Keep YAML endpoints and migrate them to Simple YAML:
  - `POST /api/v1/yaml` -> build a Simple YAML snippet from a single ComputeRequest (rename/refactor `RequestToSonnetYAML`)
  - `POST /api/v1/yaml/render` -> accept a Simple Book YAML (defaults + spreads[].image with optional overrides) and render each entry with Simple
- Remove Sonnet-specific preview helpers (`renderSpreadFromSpec`, `renderSimplePreview`) and related types
- Add project-level export endpoint:
  - `GET /api/projects/{id}/yaml/book` -> emit a Simple Book YAML covering all pages/spreads and input images

Keep and simplify these endpoints to Simple only:
- `POST /api/v1/compute` -> Accept Simple settings (already aligned with current `spread.Settings` shape) and call `simple.Inputs` directly
- `POST /api/v1/preview` -> Render PNG preview using Simple render functions
- `POST /api/v1/render` -> Render outputs using Simple render functions

Implementation details:
- Replace `toResolved(spread.Settings)` with a thin pass-through or Simple-native normalization function (no Sonnet types). Example outline:

```go
// new: build Simple Inputs directly from spread.Settings and ImageMeta
inputs, err := simple.InputsFromSettings(req.Settings, meta)
if err != nil { http.Error(w, err.Error(), 400); return }
trace := &simple.Trace{UseZerolog: true}
result := simple.ComputePlacement(inputs, trace)
```

- Update preview/render code paths to use `simple.RenderInfo` built from `req.Settings.Export` without Sonnet conversions.
- For `/api/v1/yaml/render`, parse the Simple YAML (schema below), iterate `spreads`, and call Simple for each.
- For `/api/projects/{id}/yaml/book`, assemble Book YAML from the current project assets (or pages/spreads if present) and chosen defaults.

### Simple YAML schema

Use a minimal schema mirroring `spread.Settings` with a list of spreads:

```yaml
version: "0.2"
defaults:
  paper:
    width_in: 8.0
    height_in: 10.0
    orientation: portrait
    dpi: 300
  margins:
    top_in: 0.25
    right_in: 0.25
    bottom_in: 0.25
    left_in: 0.25
  spread:
    is_spread: false
    gutter_in: 0.0
  crop:
    ratio: original   # or numeric ratio
    to_fill: false
  scale:
    user_scale: 1.0
  position:
    x: 0.0
    y: 0.0
    units: normalized
  export:
    format: png
    quality: 90
    background: white
    out_dir: ./out
    filename_template: "${name}-${panel}.${ext}"
spreads:
  - name: page-0001
    image: ./projects/PRJ/images/0001.png
    # optional overrides as in defaults: paper/margins/spread/crop/scale/position/export
  - name: page-0002
    image: ./projects/PRJ/images/0002.png
```

Notes:
- `image` must be a single file path per spread entry.
- Per-spread override sections mirror keys in `defaults` and are optional.
- To output facing pages, set `defaults.spread.is_spread: true` and use names like spread-0001.

### Book YAML export mapping

- File-based projects: enumerate images in order (existing project order) to produce `spreads` entries; use UI settings as `defaults`.
- With SQL repository: prefer `spreads` table to define order; fallback to `pages` or ordered `assets`.
- Add query param `base_dir` to influence relative path root in emitted YAML.

### Server helpers

- Rename `RequestToSonnetYAML` to `BuildSimpleYAML(req ComputeRequest)`.
- Add `BuildBookYAML(projectID, defaults, baseDir)` helper that lists images and emits the Book YAML.

## CLI verbs to adjust

Search and remove Sonnet-oriented CLI commands:
- `zine-layout/cmd/experiments/spread-cli/main.go` -> remove or refactor to Simple only if still needed

Main CLI `serve` does not expose algorithm selection; it should continue to work after REST changes.

## Web changes (RTK Query and UI)

`zine-layout/web/src/api.ts`:
- Keep `buildYaml` and `renderYaml` targeting Simple YAML.
- Add `exportBookYaml` to GET `/api/projects/{id}/yaml/book` and return raw text.
- Keep `computeSpread`, `previewSpread`, `getPreviewSpread` unchanged.

`zine-layout/web/src/views/BookSpreadDesigner.tsx`:
- No Sonnet-specific references; continue using `useGetPreviewSpreadQuery`
- Remove any imports or UI based on YAML tools if present in other views

Other web code to check:
- Remove any Sonnet-specific assets or for_each UI. Keep Simple controls only.
- Add an "Export YAML" button in BookSpreadDesigner to download or copy the Book YAML.

## Additional cleanups

- Remove Sonnet adapters and types from Simple:
  - Replace `InputsFromResolved(...)` with `InputsFromSettings(settings spread.Settings, meta spread.ImageMeta)` and update server to use it
  - Replace `RenderInfoFromExport(sonnetcfg.ExportValues, ...)` with `RenderInfoFromExport(spread.ExportSettings, ...)`

- Update imports across the repo to remove all `/sonnet/` references

## Step-by-step checklist

1. Delete directory: `zine-layout/pkg/spread/sonnet/`
2. Remove files: `zine-layout/pkg/spread/simple/adapter.go`, `zine-layout/cmd/experiments/spread-cli/main.go`
3. Keep and refactor `zine-layout/pkg/spread/yaml.go` to Simple YAML and add Book YAML export helpers
4. In `zine-layout/pkg/serve/server.go`:
   - Remove Sonnet imports and types
   - Delete `toResolved` and Sonnet spec rendering helpers
   - Compute inputs with `simple.InputsFromSettings` and render with Simple-only functions
   - Keep `POST /api/v1/yaml`, keep `POST /api/v1/yaml/render` (Simple subset), add `GET /api/projects/{id}/yaml/book`
5. In `zine-layout/web/src/api.ts`:
   - Add `exportBookYaml` and keep YAML endpoints (Simple schema)
   - Ensure compute/preview endpoints are unchanged and only reference Simple
6. Search repo for `sonnet/` and `RequestToSonnetYAML` and remove dead code
7. Run build and fix imports

## Notes

- This keeps the existing Simple algorithm behavior as the default for all previews and renders.
- If any docs reference Sonnet YAML, update them to reflect the Simple-only workflow.


