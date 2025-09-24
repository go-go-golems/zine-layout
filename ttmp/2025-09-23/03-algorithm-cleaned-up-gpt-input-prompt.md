## Clean, deterministic algorithm for single page or double‑page spread with inner gutter

This document defines a precise, re‑implementable algorithm to place and render an image on a single page or split it across a double‑page spread with a centered gutter. It is written to be easy to port to JavaScript/TypeScript.

### Coordinate spaces

- Source space S: original image pixels, top‑left origin.
- Layout space L: pixels inside the content area (after margins). The content rectangle origin is (0, 0).

### Inputs

- Source image: `srcW`, `srcH` (px)
- Paper and layout:
  - `paperWIn`, `paperHIn` (in)
  - `orientation`: `'portrait' | 'landscape'`
  - margins (in): `marginTopIn`, `marginRightIn`, `marginBottomIn`, `marginLeftIn`
  - `dpi` (px/in)
- Spread / gutter:
  - `is_spread`: boolean
  - `gutterIn`: inches (non‑negative)
- Crop & scale:
  - `crop_ratio`: `{ w, h } | null` (null means "original")
  - `crop_to_fill`: boolean (true = cover)
  - `user_scale`: number ≥ 0
- Position:
  - `image_position`: `{ x, y }`
  - `position_units`: `'px' | 'normalized'`
    - normalized ∈ [−1, 1]
    - in fill mode: moves the crop window in S (along the free axis)
    - in fit mode: translates in L

### Outputs

Rectangles (floats) and sizes (ints):
- `src_rect_global: { x, y, w, h }` in S
- `dst_rect_global: { x, y, w, h }` in L (content‑relative)
- `content_rect: { x, y, w, h }` in L
- `gutter_px: number`
- `effective_spread_w: number` (visible width across both pages, excluding the gutter)
- For spreads only:
  - `pageW: number` (page canvas width in px, equals `content_rect.w/2`)
  - `left_panel: { x, y, w, h }` (visible area on left page in L)
  - `right_panel: { x, y, w, h }` (visible area on right page in L)
  - `dst_rect_left: { x, y, w, h }` (destination rect relative to left page canvas)
  - `dst_rect_right: { x, y, w, h }` (destination rect relative to right page canvas)
- Export sizes:
  - single page: `{ w, h } = { content_rect.w, content_rect.h }`
  - spread: two sizes, both `{ w, h } = { pageW, content_rect.h }`

---

### Step‑by‑step algorithm

1) Paper and content rectangle in L (px)

```
if orientation === 'portrait':
  pageWpx = paperWIn * dpi
  pageHpx = paperHIn * dpi
else: // landscape swaps
  pageWpx = paperHIn * dpi
  pageHpx = paperWIn * dpi

spreadWpx = is_spread ? 2 * pageWpx : pageWpx

Mt = marginTopIn    * dpi
Mr = marginRightIn  * dpi
Mb = marginBottomIn * dpi
Ml = marginLeftIn   * dpi

content_w = spreadWpx - (Ml + Mr)
content_h = pageHpx  - (Mt + Mb)
content_rect = { x: 0, y: 0, w: content_w, h: content_h }

gutter_px = is_spread ? max(0, gutterIn * dpi) : 0
effective_spread_w = is_spread ? max(0, content_w - gutter_px) : content_w

// Target box for global placement (image covers visible width only; gutter is blank)
target_w = effective_spread_w
target_h = content_h
target_ratio = (target_w > 0 && target_h > 0) ? (target_w / target_h) : 0
```

2) Source crop window in S

Let `source_ratio = srcW / srcH`. Choose requested crop ratio:

```
if crop_ratio != null: req_ratio = crop_ratio.w / crop_ratio.h
else if crop_to_fill == true: req_ratio = target_ratio
else: req_ratio = source_ratio // no crop
```

Compute `[sx, sy, sw, sh]` in S:

- If `source_ratio > req_ratio` crop width:

```
sh = srcH
sw = srcH * req_ratio
range_x = max(0, srcW - sw)
if position_units == 'normalized':
  t = clamp((image_position.x + 1)/2, 0, 1)
  sx = t * range_x
else: // px
  sx = clamp(range_x/2 + image_position.x, 0, range_x)
sy = 0
```

- Else if `source_ratio < req_ratio` crop height:

```
sw = srcW
sh = srcW / req_ratio
range_y = max(0, srcH - sh)
if position_units == 'normalized':
  t = clamp((image_position.y + 1)/2, 0, 1)
  sy = t * range_y
else:
  sy = clamp(range_y/2 + image_position.y, 0, range_y)
sx = 0
```

- Else (ratios equal): no crop → `[sx, sy, sw, sh] = [0, 0, srcW, srcH]`.

3) Scale and position in L

```
scale_x = (sw > 0) ? target_w / sw : 0
scale_y = (sh > 0) ? target_h / sh : 0

if crop_to_fill:
  coverage    = max(scale_x, scale_y)
  final_scale = coverage * max(1, user_scale)  // allow zoom‑in, never below coverage
  translate_x = 0
  translate_y = 0
else: // fit: preserve entire (possibly cropped) image
  final_scale = min(scale_x, scale_y) * user_scale
  dw_tmp = sw * final_scale
  dh_tmp = sh * final_scale
  if position_units == 'normalized':
    translate_x = image_position.x * (target_w - dw_tmp)/2
    translate_y = image_position.y * (target_h - dh_tmp)/2
  else:
    translate_x = image_position.x
    translate_y = image_position.y

dw = sw * final_scale
dh = sh * final_scale

center_x = target_w / 2
center_y = target_h / 2
dx = center_x - dw/2 + translate_x
dy = center_y - dh/2 + translate_y

src_rect_global = { x: sx, y: sy, w: sw, h: sh }
dst_rect_global = { x: dx, y: dy, w: dw, h: dh }
```

4) Spread pages and inner gutter (only if `is_spread == true`)

Pages are rendered as two canvases of equal width `pageW = content_w / 2` and height `content_h`.

- Let `m = gutter_px / 2`. The visible areas (clip rectangles) in L are:

```
left_panel  = { x: 0,           y: 0, w: pageW - m,     h: content_h }
right_panel = { x: pageW + m,   y: 0, w: pageW - m,     h: content_h }
```

This creates a blank inner margin of width `m` on the right of the left page and `m` on the left of the right page. The gap between `left_panel` and `right_panel` is exactly `gutter_px`.

Per‑page destination rectangles, page‑relative:

```
pageLeftOriginX = 0
pageRightOriginX = pageW

dst_rect_left  = { x: dx - pageLeftOriginX,  y: dy, w: dw, h: dh }
dst_rect_right = { x: dx - pageRightOriginX, y: dy, w: dw, h: dh }
```

5) Export sizes

```
single page: export_single = { w: round(content_w), h: round(content_h) }
spread pages: export_spread = {
  left:  { w: round(pageW), h: round(content_h) },
  right: { w: round(pageW), h: round(content_h) }
}
```

---

### Rendering contract (canvas drawing)

Single page
1. Create a canvas of size `export_single`.
2. Draw `src_rect_global → dst_rect_global` onto the canvas. No additional clipping required.

Double‑page spread
Render twice, once per page:
1. Left page canvas: size `pageW × content_h`.
   - Clip to `clipLeft = [0, 0, pageW - m, content_h]` (inner margin on the right).
   - Draw mapping `src_rect_global → dst_rect_left`.
2. Right page canvas: size `pageW × content_h`.
   - Clip to `clipRight = [m, 0, pageW, content_h]` (inner margin on the left).
   - Draw mapping `src_rect_global → dst_rect_right`.

Notes
- The global placement uses `target_w = effective_spread_w` so the image covers only the visible width across both pages (excluding the gutter). This ensures the gutter area remains blank on both inner edges.
- For robust and fast implementations, intersect the per‑page destination rectangles with the page clip rectangles first and derive the corresponding sub‑rectangle in S before scaling (clip‑aware scaling).

---

### TypeScript reference interfaces

```ts
type Orientation = 'portrait' | 'landscape';
type PositionUnits = 'px' | 'normalized';

export interface CropRatio { w: number; h: number; }

export interface Inputs {
  srcW: number; srcH: number;            // px
  paperWIn: number; paperHIn: number;    // inches
  orientation: Orientation;
  marginTopIn: number; marginRightIn: number; marginBottomIn: number; marginLeftIn: number;
  dpi: number; is_spread: boolean; gutterIn?: number; // inches
  crop_ratio?: CropRatio | null; crop_to_fill: boolean; user_scale: number;
  image_position: { x: number; y: number }; position_units?: PositionUnits;
}

export interface Rect { x: number; y: number; w: number; h: number; }
export interface Size { w: number; h: number; }

export interface Result {
  src_rect_global: Rect; // in S
  dst_rect_global: Rect; // in L
  content_rect: Rect; gutter_px: number; effective_spread_w: number;
  left_panel: Rect | null; right_panel: Rect | null; // in L
  dst_rect_left: Rect | null; dst_rect_right: Rect | null; // page‑relative
  export_single: Size | null;
  export_spread: { left: Size; right: Size } | null;
}
```

---

### Validation rules

- `dpi > 0`, margins ≥ 0, gutter ≥ 0.
- If `is_spread == true`, `gutterIn` may be 0.
- `user_scale ≥ 0`.
- `crop_ratio`: null or positive w:h.
- `position_units`: `'px' | 'normalized'` only. Normalized must be in [−1, 1].

### Edge cases

- If `target_w <= 0` or `target_h <= 0`, nothing is drawable; return sizes 0 and skip rendering.
- If `srcW <= 0` or `srcH <= 0`, treat as invalid source.

---

### Clip‑aware scaling (optional performance optimization)

Before scaling on each page:
1. Compute the page destination rectangle `D` (left or right), and intersect with its clip rectangle `C`.
2. If `I = D ∩ C` is empty, skip.
3. Convert `I` back into a proportional sub‑rectangle of the source crop `S` (using linear mapping along X/Y).
4. Scale only `S' → I` instead of `S → D`.

This avoids scaling pixels that are outside the page’s visible area, greatly improving performance on heavy zoom or large gutter scenarios.

---

This specification is complete and deterministic. A JavaScript implementation should be able to reproduce identical rectangles and export sizes given the same inputs.


