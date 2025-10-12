Below is a precise, deterministic algorithm (with a clean reference implementation) that computes all rectangles/transforms you need for a single page **or** a double‑page spread where the same transformed image is split by a centered gutter.

---

## Coordinate spaces (recap)

* **S (Source space):** pixels of the original image.
* **L (Layout space):** pixels **inside the content area** (i.e., after margins). In L, the content rectangle’s origin is `(0,0)`.
  This makes all downstream math and clipping panel‑relative and avoids double‑counting margins.

---

## Algorithm (math-first)

### 0) Inputs

```
Image S:          (W, H)  // px
Paper (in):       (Pw_in, Ph_in), orientation ∈ {portrait, landscape}
Margins (in):     (Mt_in, Mr_in, Mb_in, Ml_in)
DPI:              dpi  // px/in
Spread flag:      is_spread ∈ {false, true}
Gutter (in):      G_in (used only if is_spread)
Crop settings:    crop_ratio r ∈ {null, w:h}, crop_to_fill ∈ {false, true}
Scale:            user_scale s ≥ 0
Position:         image_position (x, y) with units ∈ {'px','normalized'} (normalized ∈ [-1,1])
```

### 1) Content area in pixels

1. Apply orientation:

   ```
   if portrait:   Pw = Pw_in * dpi;  Ph = Ph_in * dpi
   if landscape:  Pw = Ph_in * dpi;  Ph = Pw_in * dpi   // swap
   ```

2. If spread, double width **before** margins:

   ```
   Pw_spread = (is_spread ? 2*Pw : Pw)
   ```

3. Margins in px: `Mt, Mr, Mb, Ml = margins_in * dpi`

4. Content rectangle in L (origin at top-left of content):

   ```
   content_w = Pw_spread - (Ml + Mr)
   content_h = Ph        - (Mt + Mb)
   content_rect = [0, 0, content_w, content_h]
   ```

5. Gutter in px:

   ```
   gutter_px = (is_spread ? max(0, G_in * dpi) : 0)
   effective_spread_w = (is_spread ? max(0, content_w - gutter_px) : content_w)
   ```

   If `effective_spread_w == 0`, nothing is drawable.

6. Target box for the global placement:

   ```
   target_w = effective_spread_w
   target_h = content_h
   target_ratio = target_w / target_h
   ```

### 2) Source crop window in S

Let `source_ratio = W / H`.

Define the **requested crop ratio**:

```
if r != null:
    req_ratio = r.w / r.h
else if crop_to_fill == true:
    req_ratio = target_ratio
else:
    req_ratio = source_ratio       // no crop
```

Compute crop window `[sx, sy, sw, sh]` in S:

* If `source_ratio > req_ratio`: crop **width**.

  ```
  sh = H
  sw = H * req_ratio
  range_x = W - sw
  // position.x chooses the horizontal crop window center along the free axis
  if units=='normalized': t = clamp((x + 1)/2, 0, 1); sx = t * range_x
  else (px):              sx = clamp(range_x/2 + x, 0, range_x)
  sy = 0
  ```
* Else if `source_ratio < req_ratio`: crop **height**.

  ```
  sw = W
  sh = W / req_ratio
  range_y = H - sh
  if units=='normalized': t = clamp((y + 1)/2, 0, 1); sy = t * range_y
  else (px):              sy = clamp(range_y/2 + y, 0, range_y)
  sx = 0
  ```
* Else (ratios equal): no crop → `[sx, sy, sw, sh] = [0, 0, W, H]`.

### 3) Scale (S → L) and position in L

Base scales to map the crop window to the target box:

```
scale_x = target_w / sw
scale_y = target_h / sh
```

* **Fill** (`crop_to_fill == true`): must cover

  ```
  coverage = max(scale_x, scale_y)
  final_scale = coverage * max(1, s)  // allow zoom-in, never below coverage
  translate_x = 0
  translate_y = 0
  ```

  (In fill mode you already used `image_position` to move the crop window in S.)

* **Fit** (`crop_to_fill == false`): preserve entire (possibly cropped) image

  ```
  final_scale = min(scale_x, scale_y) * s
  // In fit, image_position is a translation in L.
  if units=='normalized':
      translate_x = x * (target_w - sw*final_scale)/2
      translate_y = y * (target_h - sh*final_scale)/2
  else (px):
      translate_x = x
      translate_y = y
  ```

Destination size in L:

```
dw = sw * final_scale
dh = sh * final_scale
```

Place it centered in the target box, plus any fit-translation:

```
center_x = target_w / 2
center_y = target_h / 2
dx = center_x - dw/2 + translate_x
dy = center_y - dh/2 + translate_y
```

So the **global rectangles** are:

```
src_rect_global = [sx, sy, sw, sh]   // in S
dst_rect_global = [dx, dy, dw, dh]   // in L (content-relative)
```

### 4) Panels & clipping (only when `is_spread == true`)

Panels in **L**:

```
left_panel  = [0,                      0, effective_spread_w/2, content_h]
right_panel = [effective_spread_w/2 + gutter_px, 0, effective_spread_w/2, content_h]
clip_rect_left  = left_panel
clip_rect_right = right_panel
```

Per-panel destination rectangles (panel‑relative):

```
dst_rect_left  = [dx - left_panel.x,  dy - left_panel.y,  dw, dh]
dst_rect_right = [dx - right_panel.x, dy - right_panel.y, dw, dh]
```

**Render rule:** draw `src_rect_global → dst_rect_global` **clipped** by each panel’s clip rect. This shows each half without recomputing transforms.

### 5) Export sizes (px)

* Single page: `export_size = [content_w, content_h]`.
* Spread: `export_sizes = [[effective_spread_w/2, content_h], [effective_spread_w/2, content_h]]`.

All calculations are in pixels; only the paper & margin inputs are converted from inches at the very start.

---

## Reference implementation (TypeScript)

```ts
type Orientation = 'portrait' | 'landscape';
type PositionUnits = 'px' | 'normalized';

export interface CropRatio { w: number; h: number; }

export interface Inputs {
  // Source image
  srcW: number;           // px
  srcH: number;           // px

  // Paper and layout
  paperWIn: number;       // inches
  paperHIn: number;       // inches
  orientation: Orientation;
  marginTopIn: number;    // inches
  marginRightIn: number;  // inches
  marginBottomIn: number; // inches
  marginLeftIn: number;   // inches
  dpi: number;            // px/in

  // Spread / gutter
  is_spread: boolean;
  gutterIn?: number;      // inches

  // Crop & scale
  crop_ratio?: CropRatio | null; // null => "original"
  crop_to_fill: boolean;         // fill vs fit
  user_scale: number;            // s >= 0

  // Position
  image_position: { x: number; y: number };
  position_units?: PositionUnits; // default 'px'
}

export interface Rect { x: number; y: number; w: number; h: number; }

export interface Result {
  // Step 1
  src_rect_global: Rect; // in S
  dst_rect_global: Rect; // in L

  // Content & panels (in L)
  content_rect: Rect;
  effective_spread_w: number;
  gutter_px: number;

  left_panel: Rect | null;
  right_panel: Rect | null;
  dst_rect_left: Rect | null;
  dst_rect_right: Rect | null;

  // Export sizes
  export_single: { w: number; h: number } | null;
  export_spread: { left: { w: number; h: number }, right: { w: number; h: number } } | null;
}

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v));

export function computePlacement(inp: Inputs): Result {
  const {
    srcW: W, srcH: H,
    paperWIn, paperHIn, orientation,
    marginTopIn, marginRightIn, marginBottomIn, marginLeftIn,
    dpi,
    is_spread,
    gutterIn = 0,
    crop_ratio = null,
    crop_to_fill,
    user_scale: s,
    image_position,
    position_units = 'px'
  } = inp;

  // 1) Paper → px (apply orientation), content rect in L
  const pageWpx = (orientation === 'portrait' ? paperWIn : paperHIn) * dpi;
  const pageHpx = (orientation === 'portrait' ? paperHIn : paperWIn) * dpi;

  const spreadWpx = is_spread ? 2 * pageWpx : pageWpx;

  const Mt = marginTopIn * dpi;
  const Mr = marginRightIn * dpi;
  const Mb = marginBottomIn * dpi;
  const Ml = marginLeftIn * dpi;

  const content_w = spreadWpx - (Ml + Mr);
  const content_h = pageHpx - (Mt + Mb);
  const content_rect: Rect = { x: 0, y: 0, w: content_w, h: content_h };

  const gutter_px = is_spread ? Math.max(0, gutterIn * dpi) : 0;
  const effective_spread_w = is_spread ? Math.max(0, content_w - gutter_px) : content_w;

  const target_w = effective_spread_w;
  const target_h = content_h;
  const target_ratio = target_w > 0 && target_h > 0 ? (target_w / target_h) : 0;

  // 2) Source crop window in S
  const source_ratio = W / H;
  const req_ratio = ((): number => {
    if (crop_ratio != null) return (crop_ratio.w / crop_ratio.h);
    if (crop_to_fill)      return target_ratio;
    return source_ratio;
  })();

  let sx = 0, sy = 0, sw = W, sh = H;

  if (W > 0 && H > 0 && req_ratio > 0) {
    if (source_ratio > req_ratio) {
      // crop width
      sh = H;
      sw = H * req_ratio;
      const range_x = Math.max(0, W - sw);
      if (position_units === 'normalized') {
        const t = clamp((image_position.x + 1) / 2, 0, 1);
        sx = t * range_x;
      } else {
        sx = clamp(range_x / 2 + image_position.x, 0, range_x);
      }
      sy = 0;
    } else if (source_ratio < req_ratio) {
      // crop height
      sw = W;
      sh = W / req_ratio;
      const range_y = Math.max(0, H - sh);
      if (position_units === 'normalized') {
        const t = clamp((image_position.y + 1) / 2, 0, 1);
        sy = t * range_y;
      } else {
        sy = clamp(range_y / 2 + image_position.y, 0, range_y);
      }
      sx = 0;
    } else {
      // equal ratios: no crop
      sx = 0; sy = 0; sw = W; sh = H;
    }
  }

  // 3) Scale and position in L
  const scale_x = sw > 0 ? (target_w / sw) : 0;
  const scale_y = sh > 0 ? (target_h / sh) : 0;

  let final_scale = 0;
  let translate_x = 0;
  let translate_y = 0;

  if (crop_to_fill) {
    const coverage = Math.max(scale_x, scale_y);
    final_scale = coverage * Math.max(1, s);
    translate_x = 0;
    translate_y = 0;
  } else {
    final_scale = Math.min(scale_x, scale_y) * s;
    const dw_tmp = sw * final_scale;
    const dh_tmp = sh * final_scale;
    if (position_units === 'normalized') {
      translate_x = image_position.x * (target_w - dw_tmp) / 2;
      translate_y = image_position.y * (target_h - dh_tmp) / 2;
    } else {
      translate_x = image_position.x;
      translate_y = image_position.y;
    }
  }

  const dw = sw * final_scale;
  const dh = sh * final_scale;

  const center_x = target_w / 2;
  const center_y = target_h / 2;

  const dx = center_x - dw / 2 + translate_x;
  const dy = center_y - dh / 2 + translate_y;

  const src_rect_global: Rect = { x: sx, y: sy, w: sw, h: sh };
  const dst_rect_global: Rect = { x: dx, y: dy, w: dw, h: dh };

  // 4) Panels & clipping (spread split)
  let left_panel: Rect | null = null;
  let right_panel: Rect | null = null;
  let dst_rect_left: Rect | null = null;
  let dst_rect_right: Rect | null = null;

  if (is_spread) {
    left_panel  = { x: 0, y: 0, w: effective_spread_w / 2,               h: content_h };
    right_panel = { x: effective_spread_w / 2 + gutter_px, y: 0, w: effective_spread_w / 2, h: content_h };

    dst_rect_left  = { x: dx - left_panel.x,  y: dy - left_panel.y,  w: dw, h: dh };
    dst_rect_right = { x: dx - right_panel.x, y: dy - right_panel.y, w: dw, h: dh };
  }

  // 5) Export sizes
  const export_single = !is_spread ? { w: content_w, h: content_h } : null;
  const export_spread = is_spread ? {
    left:  { w: effective_spread_w / 2, h: content_h },
    right: { w: effective_spread_w / 2, h: content_h }
  } : null;

  return {
    src_rect_global,
    dst_rect_global,
    content_rect,
    effective_spread_w,
    gutter_px,
    left_panel,
    right_panel,
    dst_rect_left,
    dst_rect_right,
    export_single,
    export_spread
  };
}
```

**Rendering contract**

* **Single page:** draw `src_rect_global → dst_rect_global` onto a canvas of size `export_single`.
* **Spread:** render **twice**:

  1. clip to `left_panel`, draw `src_rect_global → dst_rect_global`, export size `export_spread.left`.
  2. clip to `right_panel`, draw the same mapping, export size `export_spread.right`.

This satisfies:

* determinism and parity between preview/export,
* continuous behavior when toggling spread/gutter (only `effective_spread_w` changes),
* correct use of `image_position` in **fill** (crop sliding in S) vs **fit** (translation in L),
* gutter simply removes width from the effective spread and the visible image between panels.
