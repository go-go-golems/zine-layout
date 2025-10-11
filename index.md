# Implementation Marks

Date: 2025-10-11

## Phase 3 — Stage A (Page-Level Rendering)

- [x] Add `pkg/pagelayout/settings.go` with `PageLayoutSettings` and helpers
- [x] Add `pkg/pagelayout/renderer/renderer.go` with variants: thumbnail, full, combined, left, right
- [x] HTTP preview endpoint: `GET /api/projects/{id}/page-preview?variant=thumbnail|full|left|right|combined`
- [x] CLI command: `pages-render` (renders a single image to a page)
- [x] Register `pages-render` in `cmd/zine-layout/main.go`

### How to run

- Server and web (serves API + static web if built):
```bash
go run ./cmd/zine-layout serve --data-root tmp-phase3-dev --addr :8090
```

- Preview first project image as PNG:
```bash
# thumbnail (default)
curl -sS "http://localhost:8090/api/projects/<PROJECT_ID>/page-preview" -o thumb.png
# or specific variant
curl -sS "http://localhost:8090/api/projects/<PROJECT_ID>/page-preview?variant=full" -o full.png
```

- CLI page render (single image):
```bash
go run ./cmd/zine-layout pages-render \
  --input ./some.png \
  --output ./page.png \
  --page-width-in 8.5 --page-height-in 11 --dpi 300 \
  --margin-top-in 0.5 --margin-right-in 0.5 --margin-bottom-in 0.5 --margin-left-in 0.5 \
  --positioning-mode fill --variant thumbnail
```

### Key files
- `pkg/pagelayout/settings.go`
- `pkg/pagelayout/renderer/renderer.go`
- `cmd/zine-layout/cmds/serve.go` (preview route)
- `cmd/zine-layout/cmds/pages/render.go` (CLI)
- `cmd/zine-layout/main.go` (command registration)

Notes:
- Current preview renders the first project image into a standard 8.5x11 page at 300 DPI with 0.5in margins. Spread splitting and absolute placement are supported by the renderer; UI wiring and persisted render metadata can be layered later.
