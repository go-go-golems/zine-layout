# End-to-End Test Plan – Workflow CLI Only

Date: 2025-10-12 (revised 2025-10-12 to cover new workflow verbs)

This plan validates the entire pipeline from synthetic image generation → image layout templates → laid-out images → page layout templates → laid-out pages → zine imposition → PDF export using the **serverless workflow CLI verbs**. Optional callouts show how to cross-check with the REST API. Each step includes programmatic verification ideas, and all JSON responses are saved alongside artifacts for reuse.

---

## 0. Prerequisites
- Go ≥ 1.22
- ImageMagick (`convert`) available for generating test images
- Fresh data root (disposable) and working scratch dir: `./tmp-e2e`
- jq available in the shell environment

Directory convention (persist outputs for reuse):
```bash
mkdir -p ./tmp-e2e
# All JSON state is written here for reuse between steps, e.g.:
#   ./tmp-e2e/project.json
#   ./tmp-e2e/assets.json
#   ./tmp-e2e/image_layout_template.json
#   ./tmp-e2e/laid_out_image-01.json ...
#   ./tmp-e2e/page_template.json
#   ./tmp-e2e/laid_out_page-01.json ...
#   ./tmp-e2e/zine.json
```

---

## 1. Generate Test Images

```bash
# Create data root
mkdir -p ./tmp-e2e

# Generate three synthetic PNGs
convert -size 1200x800  xc:lightblue -fill navy  -draw "rectangle 100,100 1100,700" ./tmp-e2e/img1.png
convert -size 800x1200  xc:mistyrose -fill maroon -draw "rectangle 50,50 750,1150" ./tmp-e2e/img2.png
convert -size 1600x1600 xc:ivory     -fill green  -draw "rectangle 200,200 1400,1400" ./tmp-e2e/img3.png
```

Verify programmatically:
- `file ./tmp-e2e/img*.png | sed 's/, /\n  /g'`
- Optional: `identify -format '%f %w %h\n' ./tmp-e2e/img*.png`

---

## 2. Create Project (workflow)

```bash
go run ./cmd/zine-layout workflow projects create \
  --data-root ./tmp-e2e \
  --name "E2E Test" \
  --description "Serverless end-to-end run" \
  --output json | tee ./tmp-e2e/project.json >/dev/null

PROJECT_ID=$(jq -r '.[0].project_id' ./tmp-e2e/project.json)
printf 'Project ID: %s\n' "$PROJECT_ID"
```

Verify:
- `test -n "$PROJECT_ID"`
- `go run ./cmd/zine-layout workflow projects list --data-root ./tmp-e2e --output json | jq -r '.[].project_id' | grep "$PROJECT_ID"`

---

## 3. Import Assets (workflow)

```bash
go run ./cmd/zine-layout workflow assets create \
  --data-root ./tmp-e2e \
  --project-id "$PROJECT_ID" \
  --file ./tmp-e2e/img1.png \
  --file ./tmp-e2e/img2.png \
  --file ./tmp-e2e/img3.png \
  --output json | tee ./tmp-e2e/assets.json >/dev/null

readarray -t ASSET_IDS < <(jq -r '.[].asset_id' ./tmp-e2e/assets.json)
printf 'Asset IDs:\n%s\n' "${ASSET_IDS[@]}"
```

Verify:
- `test "${#ASSET_IDS[@]}" -eq 3`
- `ls ./tmp-e2e/projects/$PROJECT_ID/images`
- Optional: `identify ./tmp-e2e/projects/$PROJECT_ID/images/*.png`

---

## 4. Create Image Layout Template (workflow)

Create settings file:
```bash
cat > ./tmp-e2e/image_layout_template.json <<'JSON'
{
  "mode": "page",
  "paper_width_in": 8.5,
  "paper_height_in": 11,
  "dpi": 300,
  "orientation": "portrait",
  "margin_top_in": 0.5,
  "margin_right_in": 0.5,
  "margin_bottom_in": 0.5,
  "margin_left_in": 0.5,
  "crop_to_fill": true,
  "user_scale": 1.0,
  "position_x": 0,
  "position_y": 0,
  "units": "normalized",
  "anchor_preset": "center"
}
JSON
```

Create template:
```bash
go run ./cmd/zine-layout workflow image-layout-templates create \
  --data-root ./tmp-e2e \
  --project-id "$PROJECT_ID" \
  --name "Cover Fill" \
  --file ./tmp-e2e/image_layout_template.json \
  --output json | tee ./tmp-e2e/image_layout_template_create.json >/dev/null

ILT_ID=$(jq -r '.[0].template_id' ./tmp-e2e/image_layout_template_create.json)
printf 'Template ID: %s\n' "$ILT_ID"
```

Verify:
- `test -n "$ILT_ID"`
- `go run ./cmd/zine-layout workflow image-layout-templates get --data-root ./tmp-e2e --template-id "$ILT_ID" --output json | jq`

---

## 5. Create Laid-Out Images (workflow)

```bash
LOI_IDS=()
idx=0
for ASSET in "${ASSET_IDS[@]}"; do
  idx=$((idx+1))
  go run ./cmd/zine-layout workflow laid-out-images create \
    --data-root ./tmp-e2e \
    --project-id "$PROJECT_ID" \
    --asset-id "$ASSET" \
    --template-id "$ILT_ID" \
    --output json | tee ./tmp-e2e/laid_out_image-$(printf '%02d' "$idx").json >/dev/null
  LOI_ID=$(jq -r '.[0].laid_out_image_id' ./tmp-e2e/laid_out_image-$(printf '%02d' "$idx").json)
  LOI_IDS+=("$LOI_ID")
done
printf 'Laid-out image IDs:\n%s\n' "${LOI_IDS[@]}"
```

Verify:
- `test "${#LOI_IDS[@]}" -eq "${#ASSET_IDS[@]}"`
- `go run ./cmd/zine-layout workflow laid-out-images get --data-root ./tmp-e2e --id "${LOI_IDS[0]}" --output json | jq`

---

## 6. Create Page Layout Template (workflow)

```bash
go run ./cmd/zine-layout workflow page-templates create \
  --data-root ./tmp-e2e \
  --project-id "$PROJECT_ID" \
  --name "A4 Portrait 300dpi" \
  --template-json '{"pageWidthIn":8.27,"pageHeightIn":11.69,"dpi":300,"marginTopIn":0.5,"marginRightIn":0.5,"marginBottomIn":0.5,"marginLeftIn":0.5,"isSpread":false,"gutterWidthIn":0,"gutterOverlapIn":0,"positioningMode":"fill","anchorPreset":"center","borderEnabled":false,"borderColor":"#000000","borderType":"plain"}' \
  --output json | tee ./tmp-e2e/page_template.json >/dev/null

PTPL_ID=$(jq -r '.[0].page_template_id' ./tmp-e2e/page_template.json)
printf 'Page template ID: %s\n' "$PTPL_ID"
```

Verify:
- `test -n "$PTPL_ID"`
- `go run ./cmd/zine-layout workflow page-templates get --data-root ./tmp-e2e --page-template-id "$PTPL_ID" --output json | jq`

---

## 7. Create Laid-Out Pages (workflow)

```bash
LPG_IDS=()
idx=0
for LOI in "${LOI_IDS[@]}"; do
  idx=$((idx+1))
  go run ./cmd/zine-layout workflow laid-out-pages create \
    --data-root ./tmp-e2e \
    --project-id "$PROJECT_ID" \
    --template-id "$PTPL_ID" \
    --laid-out-image-id "$LOI" \
    --output json | tee ./tmp-e2e/laid_out_page-$(printf '%02d' "$idx").json >/dev/null
  LPG_ID=$(jq -r '.[0].page_id' ./tmp-e2e/laid_out_page-$(printf '%02d' "$idx").json)
  LPG_IDS+=("$LPG_ID")
done
printf 'Laid-out page IDs:\n%s\n' "${LPG_IDS[@]}"
```

Verify:
- `test "${#LPG_IDS[@]}" -eq "${#LOI_IDS[@]}"`,
- `go run ./cmd/zine-layout workflow laid-out-pages get --data-root ./tmp-e2e --page-id "${LPG_IDS[0]}" --output json | jq`

---

## 8. Render Pages (workflow)

```bash
for LPG in "${LPG_IDS[@]}"; do
  go run ./cmd/zine-layout workflow laid-out-pages render \
    --data-root ./tmp-e2e \
    --page-id "$LPG" \
    --output json | tee "./tmp-e2e/render-${LPG}.json" >/dev/null
done
```

Verify:
- `ls ./tmp-e2e/projects/$PROJECT_ID/pages/$LPG`
- Inspect `render-<page_id>.json` (produced above) for variant paths.
- Optional: open the SQLite database (`sqlite3 ./tmp-e2e/zine-layout.db "SELECT result_json FROM laid_out_pages LIMIT 1;"`) to confirm metadata persistence.

---

## 9. Create Zine and Set Pages (workflow)

```bash
go run ./cmd/zine-layout workflow zines create \
  --data-root ./tmp-e2e \
  --project-id "$PROJECT_ID" \
  --name "E2E Zine" \
  --pages "${LPG_IDS[0]}" \
  --output json | tee ./tmp-e2e/zine.json >/dev/null

ZINE_ID=$(jq -r '.[0].zine_id' ./tmp-e2e/zine.json)
printf 'Zine ID: %s\n' "$ZINE_ID"

# Build 8 entries (10_8_sheet_zine expects 8 pages); cycle when fewer source pages exist
PAGES=()
for i in {0..7}; do
  idx=$((i % ${#LPG_IDS[@]}))
  PAGES+=("${LPG_IDS[$idx]}")
done
PAGES_ARG=$(IFS=,; echo "${PAGES[*]}")

go run ./cmd/zine-layout workflow zines set-pages \
  --data-root ./tmp-e2e \
  --zine-id "$ZINE_ID" \
  --pages "$PAGES_ARG" \
  --output json | tee ./tmp-e2e/zine_set_pages.json >/dev/null
```

Verify:
- `go run ./cmd/zine-layout workflow zines get --data-root ./tmp-e2e --zine-id "$ZINE_ID" --output json | jq`

---

## 10. Export Zine to PDF (workflow)

Ensure the preset exists (workflow commands do not seed it automatically in a fresh data root):
```bash
mkdir -p ./tmp-e2e/presets
cp ./examples/tests/10_8_sheet_zine.yaml ./tmp-e2e/presets/
```

Export:
```bash
go run ./cmd/zine-layout workflow zines export \
  --data-root ./tmp-e2e \
  --zine-id "$ZINE_ID" \
  --preset 10_8_sheet_zine \
  --out ./tmp-e2e/zine_cli.pdf \
  --dpi 300

file ./tmp-e2e/zine_cli.pdf | tee ./tmp-e2e/zine_cli_pdf_info.txt >/dev/null
```

> Note: The export can take several seconds depending on image size; the CLI prints verbose layout information during processing.

Verify:
- `test -s ./tmp-e2e/zine_cli.pdf`
- Optional: `pdfinfo ./tmp-e2e/zine_cli.pdf | grep Pages:`
- Optional hash: `sha256sum ./tmp-e2e/zine_cli.pdf`

---

## Optional HTTP Cross-Checks
- Start server: `go run ./cmd/zine-layout serve --data-root ./tmp-e2e --addr :8095`
- Hit REST endpoints:
  - `go run ./cmd/zine-layout api projects-list --server http://localhost:8095`
  - `curl -sS -o ./tmp-e2e/thumb.png "http://localhost:8095/api/laid-out-pages/${LPG_IDS[0]}/preview?variant=thumbnail"`
  - `curl -sS -o ./tmp-e2e/zine_http.pdf "http://localhost:8095/api/zines/${ZINE_ID}/export?preset=10_8_sheet_zine&dpi=300"`
  - Compare CLI vs HTTP PDF with `sha256sum`.

---

## Automated Verification Hooks (Ideas)
- Wrap commands in `set -euo pipefail` scripts
- Use jq assertions for required fields
- Validate image dimensions with `identify`
- Check PDF metadata with `pdfinfo`
- Emit timing/metrics or a JUnit XML for CI integration

---

## Workflow vs API Quick Reference
- Workflow coverage used above: projects, assets, image layout templates, laid-out images, page templates, laid-out pages, image sequences, layout sequences, zines (including export).
- API verbs remain useful when exercising the HTTP layer, streaming previews, or integrating with remote deployments.

---

## Cleanup
```bash
rm -rf ./tmp-e2e
```
