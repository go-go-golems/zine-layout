# End-to-End Test Plan – CLI-Only Workflow

Date: 2025-10-12

This plan validates the entire pipeline from synthetic image generation → image layout templates → laid-out images → page layout templates → laid-out pages → zine imposition → PDF export, using CLI verbs wherever possible. Each step includes programmatic verification strategies. JSON responses are saved alongside artifacts in the same directory for reuse.

## 0. Prerequisites
- Go ≥ 1.22
- ImageMagick (`convert`) available for generating test images
- Fresh data root (disposable): `./tmp-e2e`
- A jq-compatible shell environment

Directory convention (persist outputs for reuse):
```bash
mkdir -p ./tmp-e2e
# All JSON state is written here for reuse between steps
# Examples used below:
#   ./tmp-e2e/project.json
#   ./tmp-e2e/assets.json
#   ./tmp-e2e/image_layout_template.json
#   ./tmp-e2e/laid_out_image-01.json ...
#   ./tmp-e2e/page_template.json
#   ./tmp-e2e/laid_out_page-01.json ...
#   ./tmp-e2e/zine.json
```

## 1. Generate Test Images

Commands:
```bash
# Create data root
mkdir -p ./tmp-e2e

# Generate three synthetic PNGs
convert -size 1200x800  xc:lightblue -fill navy -draw "rectangle 100,100 1100,700" ./tmp-e2e/img1.png
convert -size 800x1200  xc:mistyrose -fill maroon -draw "rectangle 50,50 750,1150" ./tmp-e2e/img2.png
convert -size 1600x1600 xc:ivory     -fill green  -draw "rectangle 200,200 1400,1400" ./tmp-e2e/img3.png
```

Verify programmatically:
- Check existence and dimensions:
```bash
file ./tmp-e2e/img*.png | sed 's/, /\n  /g'
```
- Optional: parse width/height via `identify`:
```bash
identify -format '%f %w %h\n' ./tmp-e2e/img*.png
```

## 2. Start API Server

Commands:
```bash
go run ./cmd/zine-layout serve --data-root ./tmp-e2e --addr :8095
```
(Or run in tmux as per prior instructions.)

Health check:
```bash
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8095/api/health
```
Expect: `200`.

## 3. Create Project

```bash
go run ./cmd/zine-layout api projects-create \
  --name "E2E Test" \
  --server http://localhost:8095 \
  --output json | tee ./tmp-e2e/project.json >/dev/null

PROJECT_ID=$(jq -r '.[0].project_id' ./tmp-e2e/project.json)
echo "$PROJECT_ID"
```

Verify programmatically:
- `test -n "$PROJECT_ID"`
- List projects and grep for `PROJECT_ID`.

## 4. Upload Assets

```bash
go run ./cmd/zine-layout api images-upload \
  --project-id "$PROJECT_ID" \
  --files ./tmp-e2e/img1.png ./tmp-e2e/img2.png ./tmp-e2e/img3.png \
  --server http://localhost:8095 --output json | tee ./tmp-e2e/assets.json >/dev/null

readarray -t ASSET_IDS < <(jq -r '.[].asset_id' ./tmp-e2e/assets.json)
printf '%s\n' "${ASSET_IDS[@]}"
```

Verify programmatically:
- Ensure array length == 3.
- Check files exist under `./tmp-e2e/uploads/`.

## 5. Create Image Layout Template (Cover Fill)

```bash
go run ./cmd/zine-layout api image-layout-templates create \\
  --server http://localhost:8095 --project-id "$PROJECT_ID" \\
  --name "Cover Fill" \\
  --settings-json '{"mode":"page","paper_width_in":8.5,"paper_height_in":11,"dpi":300,"orientation":"portrait","margin_top_in":0.5,"margin_right_in":0.5,"margin_bottom_in":0.5,"margin_left_in":0.5,"crop_to_fill":true,"user_scale":1.0,"position_x":0,"position_y":0,"units":"normalized","anchor_preset":"center"}' \\
  --output json | tee ./tmp-e2e/image_layout_template.json >/dev/null

ILT_ID=$(jq -r '.[0].image_layout_template_id' ./tmp-e2e/image_layout_template.json)
```

Verify programmatically:
- `test -n "$ILT_ID"`
- Retrieve template and assert fields.

## 6. Create Laid-Out Images (apply layout to assets)

```bash
LOI_IDS=()
idx=0
for ASSET in "${ASSET_IDS[@]}"; do
  idx=$((idx+1))
  go run ./cmd/zine-layout api laid-out-images create \
    --server http://localhost:8095 --project-id "$PROJECT_ID" \
    --asset-id "$ASSET" --template-id "$ILT_ID" --output json | tee ./tmp-e2e/laid_out_image-$(printf '%02d' "$idx").json >/dev/null
  LOI_ID=$(jq -r '.[0].laid_out_image_id' ./tmp-e2e/laid_out_image-$(printf '%02d' "$idx").json)
  LOI_IDS+=("$LOI_ID")
done
printf '%s\n' "${LOI_IDS[@]}"
```

Verify programmatically:
- Ensure LOI count matches assets.
- Inspect one LOI via GET and check it references the correct template and asset.

## 7. Create Page Layout Template (A4 Portrait 300dpi)

```bash
go run ./cmd/zine-layout workflow page-templates create \
  --data-root ./tmp-e2e --project-id "$PROJECT_ID" \
  --name "A4 Portrait 300dpi" \
  --template-json '{"pageWidthIn":8.27,"pageHeightIn":11.69,"dpi":300,"marginTopIn":0.5,"marginRightIn":0.5,"marginBottomIn":0.5,"marginLeftIn":0.5,"isSpread":false,"gutterWidthIn":0,"gutterOverlapIn":0,"positioningMode":"fill","anchorPreset":"center","borderEnabled":false,"borderColor":"#000000","borderType":"plain"}' \
  --output json | tee ./tmp-e2e/page_template.json >/dev/null

PTPL_ID=$(jq -r '.[0].page_template_id' ./tmp-e2e/page_template.json)
```

Verify programmatically:
- `test -n "$PTPL_ID"`
- GET the page template and confirm dimensions and DPI.

## 8. Create Laid-Out Pages (renderable pages)

```bash
LPG_IDS=()
idx=0
for LOI in "${LOI_IDS[@]}"; do
  idx=$((idx+1))
  go run ./cmd/zine-layout workflow laid-out-pages create \
    --data-root ./tmp-e2e --project-id "$PROJECT_ID" \
    --template-id "$PTPL_ID" --laid-out-image-id "$LOI" --output json | tee ./tmp-e2e/laid_out_page-$(printf '%02d' "$idx").json >/dev/null
  LPG_ID=$(jq -r '.[0].laid_out_page_id' ./tmp-e2e/laid_out_page-$(printf '%02d' "$idx").json)
  LPG_IDS+=("$LPG_ID")
done
printf '%s\n' "${LPG_IDS[@]}"
```

Verify programmatically:
- Ensure count matches LOIs.
- GET `/api/laid-out-pages/{id}` and check template/id wiring.

## 9. Trigger Page Renders and Previews

```bash
for LPG in "${LPG_IDS[@]}"; do
  curl -sS -o /dev/null -w "%{http_code}\n" "http://localhost:8095/api/laid-out-pages/${LPG}/preview?variant=thumbnail"
done
```
Expect: `200` for each.

Verify programmatically:
- Check files under `./tmp-e2e/projects/${PROJECT_ID}/pages/${LPG}/` include `thumbnail.png`, `full.png` (and `combined.png` for spreads).
- Read metadata via GET `/api/laid-out-pages/{id}` and assert `ResultJSON` contains `variants.thumbnail`.

## 10. Create Zine and Set Pages

```bash
go run ./cmd/zine-layout workflow zines create --data-root ./tmp-e2e \
  --project-id "$PROJECT_ID" --name "E2E Zine" --pages "${LPG_IDS[0]}" --output json | tee ./tmp-e2e/zine.json >/dev/null
ZINE_ID=$(jq -r '.[0].zine_id' ./tmp-e2e/zine.json)

# Set pages to first 8 pages if available (for 8-up preset), otherwise cycle
PAGES_ARG=$(printf '%s,' "${LPG_IDS[@]:0:8}")
PAGES_ARG=${PAGES_ARG%,}
go run ./cmd/zine-layout workflow zines set-pages --data-root ./tmp-e2e --zine-id "$ZINE_ID" --pages "$PAGES_ARG" --output json | tee ./tmp-e2e/zine_set_pages.json >/dev/null
```

Verify programmatically:
- GET `/api/zines/{id}` and check `pages` length.

## 11. Export Zine to PDF (HTTP)

```bash
curl -sS -o ./tmp-e2e/zine.pdf "http://localhost:8095/api/zines/${ZINE_ID}/export?preset=10_8_sheet_zine&dpi=300" -w "%{http_code}\n"
file ./tmp-e2e/zine.pdf | tee ./tmp-e2e/zine_pdf_info.txt >/dev/null
```
Expect: `200`, and `PDF document` in `file` output.

Programmatic verification ideas:
- Validate non-zero size: `test -s ./tmp-e2e/zine.pdf`
- Count pages via `pdfinfo` if available: `pdfinfo ./tmp-e2e/zine.pdf | grep Pages:`
- Hash snapshot for reproducibility: `sha256sum ./tmp-e2e/zine.pdf`

## 12. Export Zine to PDF (CLI)

```bash
go run ./cmd/zine-layout workflow zines export --data-root ./tmp-e2e --zine-id "$ZINE_ID" --preset 10_8_sheet_zine --out ./tmp-e2e/zine_cli.pdf --dpi 300
file ./tmp-e2e/zine_cli.pdf | tee ./tmp-e2e/zine_cli_pdf_info.txt >/dev/null
```

Verify programmatically:
- Compare sizes / hashes between HTTP and CLI exports (should match when inputs match):
```bash
sha256sum ./tmp-e2e/zine.pdf ./tmp-e2e/zine_cli.pdf
```

---

## Automated Verification Hooks (Ideas)
- Wrap each step in a shell script and assert via `set -euo pipefail` and inline checks.
- Use `jq` to assert presence of IDs/fields in JSON.
- Verify image existence/dimensions with `identify`.
- PDF checks with `pdfinfo` and `sha256sum`.
- Optionally collect timing (`time`), and write a JUnit XML via a simple helper to integrate with CI.

---

## Which CLI verbs bypass the API server?

- Bypass REST (operate directly on the database/services via `--data-root`):
  - `workflow page-templates *`
  - `workflow laid-out-pages *`
  - `workflow zines *` (including `export`)

- Require REST API server (hit HTTP endpoints):
  - `api projects-*`, `api images-*`, `api image-layout-templates *`, `api laid-out-images *`

Notes:
- This plan mixes both: uses `api` for projects/uploads/image-layout templates/laid-out images, and uses `workflow` for page templates/pages/zines. If you want a fully serverless CLI workflow, corresponding `workflow` verbs would be needed for projects/uploads/image-layout templates/laid-out images.

## Cleanup
```bash
rm -rf ./tmp-e2e
```
