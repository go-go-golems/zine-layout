# Image Resizer Specification (Updated with UI Improvements)

Complete, implementation-ready spec with all enhancements for the resizing modes.

---

## Controls (minimal but complete)

### Global

* **Target size mode**
  * **A) Page + margins** (outputs a page-sized bitmap with margins and the image placed inside the usable area)
  * **B) Fixed crop format** (outputs only the crop rectangle)
  * **C) Fit to width/height** (outputs the resized image, optionally cropped)
  * **D) Spread with gutter** (outputs two overlapping pages for book/magazine spreads)
* **Export size units**: px (recommended); optionally mm/in + PPI if you prefer deriving pixels.
* **Portrait/Landscape toggle**: Quick swap for Page, Fixed Crop, and Spread modes

---

## A) Page + margins

* **Page size**: (W_p × H_p) (px)
  * **Portrait/Landscape swap**: Button to swap width ↔ height
* **Margins**: (m_l, m_r, m_t, m_b) (px; link toggle for uniform margins)
* **Visible/Content area**: 
  * **Computed and displayed prominently**: W_u = W_p - (m_l + m_r), H_u = H_p - (m_t + m_b)
  * This is the primordial element - all positioning happens within this area
* **Crop / Fill** (checkbox):
  * **On** = **Cover** (image fills content area; parts may be cropped)
  * **Off** = **Contain** (no crop; image fits inside content area; empty margin remains)
* **Centering mode**:
  * **Checkbox: "Use anchor centering"** (default: off)
  * **When OFF**: Image is centered in visible area by default; use manual drag to position
  * **When ON**: Use anchor positioning
    * Anchor: Left/Center/Right × Top/Middle/Bottom
    * or **Focus point** (click in image to align to a target point, e.g., center)
* **Manual adjustments**:
  * **Zoom** (multiplier): Scale the image beyond base fit
  * **Drag** (translation): Move the **image** within the visible area (not the area itself)
    * Intelligent limits: Only allows dragging where there's room
    * Quick anchor buttons: Snap to Left/Center/Right and Top/Middle/Bottom
    * Auto-disable on axes where image fits exactly (no dragging possible)
* **Preview display**: 
  * Shows page with margins
  * Image parts outside visible area rendered at 30% opacity
  * Visible area content at full opacity with blue outline

---

## B) Fixed crop format

* **Aspect presets**: 1:1, 3:2, 2:3, 4:3, 3:4, 5:4, 4:5, 16:9, 9:16, custom (W:H)
  * Portrait and landscape variants available in dropdown
* **Portrait/Landscape swap**: Button to swap dimensions and aspect ratio
* **Crop rectangle**: locked aspect; user can size/position (or "max inside page")
* **Fit mode**: Cover (default for true cropping) or Contain (letterbox inside crop)
* **Centering**: anchor or focus point
* **Manual adjustments**:
  * **Zoom** and **Drag** with intelligent limits (same as Page mode)
  * Quick anchor buttons for positioning
  * Auto-disable when no dragging possible

---

## C) Fit to width/height (image-only output)

* **Target**: exact width **or** height (px); other dimension implied
* **Crop / Fill** checkbox:
  * **On** = "fill" to meet both target width and height → **Cover** (crop as needed)
  * **Off** = **Contain** to the constraining dimension; other dimension determined by aspect
* **Centering** (only relevant when cropping is on): anchor or focus point
* **Manual adjustments**: Zoom and drag with intelligent limits
* **Optional max dimension** clamp (keep it minimal—only if you need it)

---

## D) Spread with gutter

For creating book spreads, magazine layouts, or any printed material where an image spans across facing pages.

* **Spread size**: (W_s × H_s) (px) - total dimensions of the full spread
  * **Portrait/Landscape swap**: Button to swap width ↔ height
* **Gutter configuration**:
  * **Width**: (W_g) (px) - width of the gutter/binding area
  * **Position**: Left | Center | Right
    * **Center**: Gutter at exact center (symmetric pages)
    * **Left**: Gutter at left edge (small left page, large right page)
    * **Right**: Gutter at right edge (large left page, small right page)
  * **Overlap**: (O_g) (px) - how much each page extends into gutter
* **Crop / Fill** (checkbox):
  * **On** = **Cover** (image fills spread; parts may be cropped)
  * **Off** = **Contain** (image fits inside spread)
* **Centering**: Anchor positioning (Left/Center/Right × Top/Middle/Bottom)
* **Manual adjustments**: Zoom and drag with intelligent limits
* **Preview outputs**:
  * Full spread with gutter visualization
  * Combined spread (both pages with binding gap)
  * Individual left and right pages with gutter indicators
* **Export options**:
  * Combined spread (realistic preview)
  * Left page (W_left × H_s)
  * Right page (W_right × H_s)
  * Full uncut spread (W_s × H_s)

### Page dimension formulas

Given gutter position G_pos and gutter center line at x = G_x:

**Center gutter** (G_pos = 'center'):
```
G_x = W_s / 2
W_left = G_x + O_g
W_right = G_x + O_g
```

**Left gutter** (G_pos = 'left'):
```
G_x = W_g
W_left = W_g + O_g
W_right = W_s - W_g + O_g
```

**Right gutter** (G_pos = 'right'):
```
G_x = W_s - W_g
W_left = W_s - W_g + O_g
W_right = W_g + O_g
```

### Extraction coordinates

For extracting pages from the full spread image:

**Left page source**:
```
source_x = G_x - W_left
source_y = 0
source_width = W_left
source_height = H_s
```

**Right page source**:
```
source_x = G_x - O_g
source_y = 0
source_width = W_right
source_height = H_s
```

---

## Coordinate model

* **Image** (I): pixels (W_i, H_i)
* **Page** (P): pixels (W_p, H_p)
* **Usable/content area** (U) (for Page + margins):
  ```
  W_u = W_p - (m_l + m_r),  H_u = H_p - (m_t + m_b)
  ```
  origin (c_u) at the center of (U) (use center for simpler math)
* **Crop rect** (C) (for Fixed crop format): size (W_c, H_c), aspect locked, center (c_c)

We render by transforming the image with scale (s) and translation (u=(u_x, u_y)) in the target rect's local coordinates (either (U) or (C)).

---

## Fit / scale formulas

Given a target rect with size (W_t, H_t) (this is (U) for Page + margins, or (C) for Fixed crop, or the requested output for Fit to width/height):

* **Cover** (fill target, allow cropping):
  ```
  s_cover = max(W_t/W_i, H_t/H_i)
  ```
* **Contain** (no crop, may letterbox inside target):
  ```
  s_contain = min(W_t/W_i, H_t/H_i)
  ```
* **Fit to width**:
  ```
  s_w = W_t/W_i
  ```
* **Fit to height**:
  ```
  s_h = H_t/H_i
  ```

Final scale:
```
s = s_base · z
```
with (z) the user's zoom multiplier (start at 1).

---

## Translation (centering, focus, clamping)

### Drag Limits Calculation

For intelligent UI controls, calculate the maximum drag range:

```javascript
xRange = (s·W_i - W_t) / 2
yRange = (s·H_i - H_t) / 2

// In Cover mode (or when clamping is needed):
canDragX = (s·W_i > W_t)
canDragY = (s·H_i > H_t)

if (Cover mode) {
  dragLimits = {
    minX: canDragX ? -xRange : 0,
    maxX: canDragX ? xRange : 0,
    minY: canDragY ? -yRange : 0,
    maxY: canDragY ? yRange : 0
  }
} else {
  // Contain mode: allow more freedom
  dragLimits = {
    minX: -max(xRange, fallbackValue),
    maxX: max(xRange, fallbackValue),
    minY: -max(yRange, fallbackValue),
    maxY: max(yRange, fallbackValue)
  }
}
```

### Clamping (Cover mode)

When **Cover** (or any mode that must not reveal empty space), clamp translation so the target rect is fully covered:

```javascript
function clampCover(ux, uy, s, Wi, Hi, Wt, Ht) {
  const xMin = -(s*Wi - Wt)/2;
  const xMax = +(s*Wi - Wt)/2;
  const yMin = -(s*Hi - Ht)/2;
  const yMax = +(s*Hi - Ht)/2;
  
  // Only clamp if there's room to move (image larger than target)
  // Otherwise force to 0 (centered)
  const uxClamped = (s*Wi > Wt) ? clamp(ux, xMin, xMax) : 0;
  const uyClamped = (s*Hi > Ht) ? clamp(uy, yMin, yMax) : 0;
  
  return { ux: uxClamped, uy: uyClamped };
}
```

### Anchor-based translation

With α_x, α_y ∈ {0, 0.5, 1} (or continuous [0,1]):

```
u_x = (2α_x - 1)·(s·W_i - W_t)/2
u_y = (2α_y - 1)·(s·H_i - H_t)/2
```

Quick anchor values:
* Left/Top: 0
* Center/Middle: 0.5
* Right/Bottom: 1

**Focus point**: given image point (x_f, y_f) in pixels and chosen target point (x_t, y_t) in target-local coords (often (0,0) for center), solve:

```
c_t + u + s·(x_f - W_i/2, y_f - H_i/2) = c_t + (x_t, y_t)
⇒ u = (x_t, y_t) - s·(x_f - W_i/2, y_f - H_i/2)
```

then clamp (u) if Cover.

### For Contain mode

You can skip clamping (letterbox allowed). If you prefer, clamp to keep the image entirely visible:

```
u_x ∈ [-(W_t - s·W_i)/2, +(W_t - s·W_i)/2]
u_y ∈ [-(H_t - s·H_i)/2, +(H_t - s·H_i)/2]
```

(useful only if you want to prevent the image from drifting outside the target).

---

## Algorithms by mode

### A) Page + margins

**Inputs**: W_p, H_p, m_l, m_r, m_t, m_b; checkbox **Crop/Fill**; checkbox **Use Anchor Centering**; Anchor (if centering on) or manual drag; optional zoom (z).

1. Compute visible area (U): W_u, H_u. **Display these dimensions prominently**. Set target rect T ≡ U.

2. Base fit:
   * If **Crop/Fill = On** → **Cover**, else **Contain**.
   ```
   s = (Cover? max : min)(W_u/W_i, H_u/H_i)·z
   ```

3. Calculate drag limits based on s, W_i, H_i, W_u, H_u (see Drag Limits section)

4. Positioning:
   * If **Use Anchor Centering = OFF**: 
     * u = (dragX, dragY) (manual drag only; starts at center)
   * If **Use Anchor Centering = ON**:
     * If Anchor → compute u from (α_x, α_y), then add manual drag
     * If Focus → compute u from focus mapping
   * Apply clamping if Crop/Fill is On

5. **Output**: a page-sized bitmap (W_p × H_p)
   * Draw blank page background
   * Draw image at 30% opacity (shows what gets cut off)
   * Clip to visible area and draw image at full opacity
   * Draw blue outline around visible area
   * This mode **keeps margins** as empty space around the image

### B) Fixed crop format

**Inputs**: aspect (r=W:H); crop rect (C) (user-sized/positioned or "max inside page"). Fit = Cover (default) or Contain; Anchor/Focus; zoom (z).

1. Determine crop rect size (W_c, H_c) by aspect (r) and user sizing. (T ≡ C)

2. Base fit: s = Cover/Contain(W_c, H_c; W_i, H_i)·z

3. Calculate drag limits

4. Centering: Anchor or Focus → compute u; add manual drag; clamp if Cover

5. **Output**: bitmap **exactly** (W_c × H_c)
   * Render only the crop rectangle

### C) Fit to width/height (image-only)

**Inputs**: target W_t or H_t (px), **Crop/Fill**; Anchor/Focus (only if Crop/Fill On); zoom (z).

1. Determine target rect (T):
   * If **Crop/Fill OFF**:
     * If width given → s = W_t/W_i, H_out = ⌊s·H_i⌋
     * If height given → s = H_t/H_i, W_out = ⌊s·W_i⌋
     * **Output** size is the scaled image; **no cropping**; centering irrelevant
   
   * If **Crop/Fill ON** (force both dims):
     * Must have both W_t and H_t (either provided or infer one from chosen aspect)
     * Set T=(W_t, H_t), s=max(W_t/W_i, H_t/H_i)·z
     * Calculate drag limits
     * Center via Anchor/Focus; add manual drag; clamp
     * **Output** bitmap size is exactly W_t × H_t (cropped)

2. Render accordingly

### D) Spread with gutter

**Inputs**: W_s, H_s (spread dimensions); W_g, O_g (gutter width and overlap); G_pos (gutter position: left/center/right); **Crop/Fill**; Anchor; zoom (z).

1. Set target rect T = (W_s, H_s) - the full spread

2. Base fit:
   * If **Crop/Fill = On** → **Cover**, else **Contain**
   ```
   s = (Cover? max : min)(W_s/W_i, H_s/H_i)·z
   ```

3. Calculate drag limits based on s, W_i, H_i, W_s, H_s

4. Positioning:
   * Anchor → compute u from (α_x, α_y), then add manual drag
   * Apply clamping if Crop/Fill is On

5. Calculate gutter center line G_x and page dimensions:
   * If G_pos = 'center':
     * G_x = W_s / 2
     * W_left = G_x + O_g
     * W_right = G_x + O_g
   * If G_pos = 'left':
     * G_x = W_g
     * W_left = W_g + O_g
     * W_right = W_s - W_g + O_g
   * If G_pos = 'right':
     * G_x = W_s - W_g
     * W_left = W_s - W_g + O_g
     * W_right = W_g + O_g

6. **Render full spread** (W_s × H_s):
   * Draw image with transform (scale s, translation u)
   * Visualize gutter: draw translucent red rectangle and dashed line at G_x

7. **Extract and render left page** (W_left × H_s):
   * Source rect: (G_x - W_left, 0, W_left, H_s) from full spread
   * Draw gutter line at x = (W_left - O_g) as red dashed line
   * Draw page border

8. **Extract and render right page** (W_right × H_s):
   * Source rect: (G_x - O_g, 0, W_right, H_s) from full spread
   * Draw gutter line at x = O_g as red dashed line
   * Draw page border

9. **Render combined spread preview**:
   * Place left page at x = 0
   * Add small gap (4px dark background to represent binding)
   * Place right page at x = W_left + gap
   * Total preview size: (W_left + W_right + gap) × H_s

10. **Output options**:
    * Combined spread preview (with binding gap)
    * Left page only (W_left × H_s)
    * Right page only (W_right × H_s)
    * Full uncut spread (W_s × H_s)

---

## Minimal draw pseudocode (TypeScript-ish)

```ts
// image: Wi,Hi; target rect T: size Wt,Ht, center ct (in output coords)
// mode: 'cover' | 'contain' | 'fitWidth' | 'fitHeight'
// z: zoom multiplier
// anchor ax,ay in [0,1] OR focus point xf,yf (image px) with target point xt,yt (target-local)
// manualDrag: (dragX, dragY) user's manual adjustment

function baseScale(mode: string, Wi:number, Hi:number, Wt:number, Ht:number) {
  if (mode === 'cover')   return Math.max(Wt/Wi, Ht/Hi);
  if (mode === 'contain') return Math.min(Wt/Wi, Ht/Hi);
  if (mode === 'fitWidth')  return Wt/Wi;
  if (mode === 'fitHeight') return Ht/Hi;
  throw new Error('bad mode');
}

function anchorToTranslation(ax:number, ay:number, s:number, Wi:number, Hi:number, Wt:number, Ht:number) {
  const ux = (2*ax - 1) * (s*Wi - Wt)/2;
  const uy = (2*ay - 1) * (s*Hi - Ht)/2;
  return {ux, uy};
}

function clampCover(ux:number, uy:number, s:number, Wi:number, Hi:number, Wt:number, Ht:number) {
  const xMin = -(s*Wi - Wt)/2, xMax = +(s*Wi - Wt)/2;
  const yMin = -(s*Hi - Ht)/2, yMax = +(s*Hi - Ht)/2;
  
  // Only clamp axes where image is larger than target
  return {
    ux: (s*Wi > Wt) ? Math.min(xMax, Math.max(xMin, ux)) : 0,
    uy: (s*Hi > Ht) ? Math.min(yMax, Math.max(yMin, uy)) : 0,
  };
}

function calculateDragLimits(s:number, Wi:number, Hi:number, Wt:number, Ht:number, isCoverMode:boolean) {
  const xRange = (s*Wi - Wt)/2;
  const yRange = (s*Hi - Ht)/2;
  
  if (isCoverMode) {
    return {
      minX: s*Wi > Wt ? -xRange : 0,
      maxX: s*Wi > Wt ? xRange : 0,
      minY: s*Hi > Ht ? -yRange : 0,
      maxY: s*Hi > Ht ? yRange : 0,
      canDragX: s*Wi > Wt,
      canDragY: s*Hi > Ht
    };
  } else {
    // Contain mode - more freedom
    return {
      minX: -Math.max(xRange, 500),
      maxX: Math.max(xRange, 500),
      minY: -Math.max(yRange, 500),
      maxY: Math.max(yRange, 500),
      canDragX: true,
      canDragY: true
    };
  }
}

function focusToTranslation(xf:number, yf:number, xt:number, yt:number, s:number, Wi:number, Hi:number) {
  const imgLocalX = xf - Wi/2;
  const imgLocalY = yf - Hi/2;
  return {
    ux: xt - s*imgLocalX,
    uy: yt - s*imgLocalY,
  };
}

// DRAW (canvas 2D):
// For Page+Margins mode:
// 1. Draw full page background
// 2. Draw image at 30% opacity (full transform)
// 3. Clip to visible area and draw at 100% opacity
// 4. Draw visible area outline

// For other modes:
ctx.save();
ctx.translate(ct.x, ct.y);        // target center
ctx.translate(ux + dragX, uy + dragY);  // anchor translation + manual drag
ctx.scale(s, s);
ctx.drawImage(img, -Wi/2, -Hi/2);
ctx.restore();
```

---

## UI Enhancements

### Quick Anchor Buttons
For both horizontal and vertical positioning:
* Buttons: Left/Center/Right and Top/Middle/Bottom
* Clicking sets drag to the exact limit value:
  * Left/Top: dragLimits.min
  * Center/Middle: 0
  * Right/Bottom: dragLimits.max

### Smart Drag Controls
* Display current position in pixels
* Slider ranges automatically set to calculated drag limits
* Disable sliders on axes where dragging is impossible
* Show helpful message: "Image fits [horizontally|vertically] - no dragging needed"

### Portrait/Landscape Toggle
* Single button swaps width ↔ height
* For fixed crop: also updates aspect ratio (3:2 ↔ 2:3, etc.)

---

## What to export (per mode)

* **Page + margins** → output bitmap (W_p × H_p) (image rendered only inside (U))
* **Fixed crop format** → output bitmap (W_c × H_c)
* **Fit to width/height**
  * **Crop/Fill Off** → output the scaled image size; no crop
  * **Crop/Fill On** → output exactly (W_t × H_t) (cropped)
* **Spread with gutter** → multiple outputs:
  * **Combined spread** → (W_left + W_right + gap) × H_s with binding visualization
  * **Left page** → W_left × H_s
  * **Right page** → W_right × H_s
  * **Full uncut spread** → W_s × H_s

This keeps the surface area deterministic, with **page size + margins + optional centering + crop/fill** as a first-class path, and **spread with gutter** for print-ready book layouts with intelligent UI controls that prevent invalid configurations.