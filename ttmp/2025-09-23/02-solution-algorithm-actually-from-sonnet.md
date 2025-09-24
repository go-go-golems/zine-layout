Below is a precise, implementation‑ready geometry algorithm that matches your spec. It computes a single “virtual spread” placement for the image (ignoring the gutter), then derives per‑panel positions so the left/right canvases clip their own halves. You can feed the output directly to `CanvasRenderingContext2D.drawImage` (or equivalent) to render two separate pages.

---

## Coordinate system & units

* Work in **one** render unit everywhere (pixels recommended).
  If your paper/margins/gutter are in inches, convert once with `px = inches * dpi` before calling the algorithm.
* Origin `(0,0)` is the **top-left of the full spread**.
* The “virtual single image” is placed **only** in the **effective content area** (which excludes margins and gutter).

---

## Inputs (all lengths in the same unit)

```ts
src: { width: number, height: number }              // source image pixels
paper: { totalWidth: number, totalHeight: number }  // full spread
margins: { top: number, right: number, bottom: number, left: number }
gutterWidth: number
transform: {
  cropRatio: 'original' | number | string, // e.g. 1.5 or "3:2" or "original"
  cropToFill: boolean,                      // fill effective area (aspect match via crop)
  imageScale: number,                       // 0.1 .. 3.0 (ignored when cropToFill = true)
  imagePosition: { x: number, y: number }   // -100 .. 100, see semantics below
}
```

**ImagePosition semantics**

* When `cropToFill = true`: `x/y` choose which part of the image survives the aspect-ratio crop
  (`x=-100` = keep far left, `x=+100` = keep far right; analogous for `y`).
* When `cropToFill = false`: `x/y` pan within the overhang when the scaled image is **larger** than the content area
  (`+100` shows the right/bottommost part).

---

## Outputs

For each panel (left/right) you’ll get:

```ts
type Rect = { x:number, y:number, width:number, height:number };

{
  effectiveRect: Rect,                        // content area (shared)
  sourceCrop: Rect,                           // final crop on the source image (shared)
  imgDisplaySize: { width:number, height:number }, // same for both panels
  left: {
    panelRect: Rect,                          // absolute position/dimensions of left panel
    imgPositionInPanel: { x:number, y:number } // where to draw the (same) image inside this panel
  },
  right: {
    panelRect: Rect,
    imgPositionInPanel: { x:number, y:number }
  }
}
```

Use with canvas:

```js
ctx.drawImage(
  img,
  sourceCrop.x, sourceCrop.y, sourceCrop.width, sourceCrop.height,
  imgPositionInPanel.x, imgPositionInPanel.y, imgDisplaySize.width, imgDisplaySize.height
);
// The canvas or a clip path provides clipping to panelRect.size.
```

---

## Algorithm (step-by-step math)

### 1) Effective content area (virtual spread without gutter)

```text
effectiveWidth  = totalWidth  - margins.left - margins.right - gutterWidth
effectiveHeight = totalHeight - margins.top  - margins.bottom

contentX0 = margins.left
contentY0 = margins.top
```

Panel widths (sum exactly to `effectiveWidth`):

```text
leftPanelWidth  = floor(effectiveWidth / 2)
rightPanelWidth = effectiveWidth - leftPanelWidth
```

Panel rectangles (absolute, in spread coordinates):

```text
left.panelRect  = (x: contentX0,
                   y: contentY0,
                   width: leftPanelWidth,
                   height: effectiveHeight)

right.panelRect = (x: contentX0 + leftPanelWidth + gutterWidth,
                   y: contentY0,
                   width: rightPanelWidth,
                   height: effectiveHeight)
```

### 2) Source crop ratio (optional pre-crop)

Parse `cropRatio` to a positive number `targetAR` (width/height) or `null` (`'original'`).

If `targetAR != null`, center‑crop the **source image** to `targetAR`:

Let `srcAR = src.width / src.height`.

* If `srcAR > targetAR` (too wide):
  `cropW = round(src.height * targetAR)`
  `cropH = src.height`
  `cropX = round((src.width - cropW)/2)`
  `cropY = 0`
* Else if `srcAR < targetAR` (too tall):
  `cropW = src.width`
  `cropH = round(src.width / targetAR)`
  `cropX = 0`
  `cropY = round((src.height - cropH)/2)`
* Else (already matches): use full image.

Call the result `crop0 = (cropX, cropY, cropW, cropH)`.
If `targetAR == null`, then `crop0 = (0,0, src.width, src.height)`.

### 3) Crop-to-fill (if enabled) **inside** `crop0`

Target aspect is the effective area:
`effAR = effectiveWidth / effectiveHeight`.

Evaluate `r0 = crop0.width / crop0.height`.

* If `r0 > effAR` (still too wide): keep full height, crop width.

  ```
  targetW  = round(crop0.height * effAR)
  targetH  = crop0.height
  maxShift = crop0.width - targetW        // pixels of permissible horizontal slide
  px       = clamp(imagePosition.x, -100, 100) / 100
  offsetX  = round(((px + 1) / 2) * maxShift)   // -100→0 (left), 0→center, +100→max (right)
  offsetY  = 0
  ```
* Else if `r0 < effAR` (too tall): keep full width, crop height.

  ```
  targetW  = crop0.width
  targetH  = round(crop0.width / effAR)
  maxShift = crop0.height - targetH
  py       = clamp(imagePosition.y, -100, 100) / 100
  offsetX  = 0
  offsetY  = round(((py + 1) / 2) * maxShift)   // -100→top, +100→bottom
  ```
* Else: use `crop0` as-is.

Final **source crop** rectangle (used for both panels):

```text
sourceCrop = (x: crop0.x + offsetX,
              y: crop0.y + offsetY,
              width: targetW, height: targetH)
```

### 4) Determine display size and placement of the virtual image

**A. If `cropToFill = true`**

* The destination **must** exactly fill the effective area:

  ```
  imgDisplayWidth  = effectiveWidth
  imgDisplayHeight = effectiveHeight
  virtualX = contentX0
  virtualY = contentY0
  ```

  (User `imageScale` and destination panning are ignored in fill mode; their effect was applied as crop offsets in step 3.)

**B. If `cropToFill = false`**

* Fit the **cropped source** into the effective area, then apply user scale and pan.

  ```
  baseScale = min(effectiveWidth  / sourceCrop.width,
                  effectiveHeight / sourceCrop.height)
  scale     = baseScale * clamp(imageScale, 0.1, 3.0)

  imgDisplayWidth  = sourceCrop.width  * scale
  imgDisplayHeight = sourceCrop.height * scale
  ```
* Center in the effective area:

  ```
  cx = contentX0 + (effectiveWidth  - imgDisplayWidth)  / 2
  cy = contentY0 + (effectiveHeight - imgDisplayHeight) / 2
  ```
* Pan only when the image overhangs (no pan if it’s smaller than the content area):

  ```
  panRangeX = max(0, (imgDisplayWidth  - effectiveWidth)  / 2)
  panRangeY = max(0, (imgDisplayHeight - effectiveHeight) / 2)

  px = clamp(imagePosition.x, -100, 100) / 100
  py = clamp(imagePosition.y, -100, 100) / 100

  // Positive px/py reveals right/bottommost parts: shift image left/up respectively.
  virtualX = cx - px * panRangeX
  virtualY = cy - py * panRangeY
  ```

### 5) Per‑panel image positions (clipping does the split)

Compute once for the left panel; the right panel is a fixed additional shift by `(leftPanelWidth + gutterWidth)`:

```text
left.imgPositionInPanel.x  = virtualX - left.panelRect.x
left.imgPositionInPanel.y  = virtualY - left.panelRect.y

right.imgPositionInPanel.x = (virtualX - right.panelRect.x)   // equivalent to:
                           = (virtualX - left.panelRect.x) - (leftPanelWidth + gutterWidth)
right.imgPositionInPanel.y = virtualY - right.panelRect.y

// imgDisplaySize is identical for both panels.
```

**Why this works:** Both panels draw the **same** (cropped + scaled) image. The right panel shifts the image left by `(leftPanelWidth + gutterWidth)` relative to the left panel, so the guttered seam aligns and the two halves butt seamlessly when viewed side‑by‑side.

---

## TypeScript reference implementation

```ts
type Vec2 = { x:number, y:number };
type Size = { width:number, height:number };
type Rect = Vec2 & Size;

type Inputs = {
  src: Size;
  paper: Size;
  margins: { top:number, right:number, bottom:number, left:number };
  gutterWidth: number;
  transform: {
    cropRatio: 'original' | number | string;
    cropToFill: boolean;
    imageScale: number;                // ignored in fill mode
    imagePosition: { x:number, y:number };
  };
};

type PanelOut = {
  panelRect: Rect;
  imgPositionInPanel: Vec2;
};

type Output = {
  effectiveRect: Rect;
  sourceCrop: Rect;
  imgDisplaySize: Size;
  left: PanelOut;
  right: PanelOut;
};

const clamp = (v:number, lo:number, hi:number) => Math.min(Math.max(v, lo), hi);

function parseCropRatio(r: Inputs['transform']['cropRatio']): number | null {
  if (r === 'original' || r == null) return null;
  if (typeof r === 'number' && isFinite(r) && r > 0) return r;
  if (typeof r === 'string') {
    const m = r.trim().match(/^(\d+(?:\.\d+)?)[\s]*[:\/][\s]*(\d+(?:\.\d+)?)$/);
    if (m) return parseFloat(m[1]) / parseFloat(m[2]);
    const n = parseFloat(r);
    if (isFinite(n) && n > 0) return n;
  }
  return null;
}

function centerCropToAspect(srcW:number, srcH:number, targetAR:number): Rect {
  const srcAR = srcW / srcH;
  if (Math.abs(srcAR - targetAR) < 1e-9) {
    return { x:0, y:0, width:srcW, height:srcH };
  }
  if (srcAR > targetAR) {
    const width  = Math.round(srcH * targetAR);
    const height = srcH;
    const x = Math.round((srcW - width) / 2);
    return { x, y:0, width, height };
  } else {
    const width  = srcW;
    const height = Math.round(srcW / targetAR);
    const y = Math.round((srcH - height) / 2);
    return { x:0, y, width, height };
  }
}

function cropToAspectWithin(rect: Rect, targetAR:number, pos: Vec2): Rect {
  const r0 = rect.width / rect.height;
  let { x, y, width:w, height:h } = rect;

  if (Math.abs(r0 - targetAR) < 1e-9) return rect;

  if (r0 > targetAR) {
    const targetW = Math.round(h * targetAR);
    const maxShift = w - targetW;
    const px = clamp(pos.x, -100, 100) / 100;          // -1..1
    const offsetX = Math.round(((px + 1) / 2) * maxShift);
    return { x: x + offsetX, y, width: targetW, height: h };
  } else {
    const targetH = Math.round(w / targetAR);
    const maxShift = h - targetH;
    const py = clamp(pos.y, -100, 100) / 100;
    const offsetY = Math.round(((py + 1) / 2) * maxShift);
    return { x, y: y + offsetY, width: w, height: targetH };
  }
}

export function computeSpreadLayout(i: Inputs): Output {
  const { src, paper, margins, gutterWidth, transform } = i;

  const effectiveWidth  = paper.totalWidth  - margins.left - margins.right - gutterWidth;
  const effectiveHeight = paper.totalHeight - margins.top  - margins.bottom;

  const leftPanelWidth  = Math.floor(effectiveWidth / 2);
  const rightPanelWidth = effectiveWidth - leftPanelWidth;

  const contentX0 = margins.left;
  const contentY0 = margins.top;

  const effAR = effectiveWidth / effectiveHeight;

  // Step 2: initial crop (optional ratio)
  const ratio = parseCropRatio(transform.cropRatio);
  const crop0 = ratio
    ? centerCropToAspect(src.width, src.height, ratio)
    : { x:0, y:0, width: src.width, height: src.height };

  // Step 3: crop-to-fill inside crop0 (if enabled)
  const sourceCrop = transform.cropToFill
    ? cropToAspectWithin(crop0, effAR, transform.imagePosition)
    : crop0;

  // Step 4: size + placement of the virtual image
  let imgDisplayWidth: number, imgDisplayHeight: number, virtualX: number, virtualY: number;

  if (transform.cropToFill) {
    imgDisplayWidth  = effectiveWidth;
    imgDisplayHeight = effectiveHeight;
    virtualX = contentX0;
    virtualY = contentY0;
  } else {
    const baseScale = Math.min(
      effectiveWidth  / sourceCrop.width,
      effectiveHeight / sourceCrop.height
    );
    const scale = baseScale * clamp(transform.imageScale, 0.1, 3.0);

    imgDisplayWidth  = sourceCrop.width  * scale;
    imgDisplayHeight = sourceCrop.height * scale;

    const cx = contentX0 + (effectiveWidth  - imgDisplayWidth)  / 2;
    const cy = contentY0 + (effectiveHeight - imgDisplayHeight) / 2;

    const panRangeX = Math.max(0, (imgDisplayWidth  - effectiveWidth)  / 2);
    const panRangeY = Math.max(0, (imgDisplayHeight - effectiveHeight) / 2);

    const px = clamp(transform.imagePosition.x, -100, 100) / 100;
    const py = clamp(transform.imagePosition.y, -100, 100) / 100;

    // Positive px/py show right/bottom: move image left/up.
    virtualX = cx - px * panRangeX;
    virtualY = cy - py * panRangeY;
  }

  // Step 5: per-panel positions
  const leftPanelRect: Rect = {
    x: contentX0, y: contentY0, width: leftPanelWidth, height: effectiveHeight
  };
  const rightPanelRect: Rect = {
    x: contentX0 + leftPanelWidth + gutterWidth,
    y: contentY0,
    width: rightPanelWidth,
    height: effectiveHeight
  };

  const left_imgPos: Vec2 = {
    x: virtualX - leftPanelRect.x,
    y: virtualY - leftPanelRect.y
  };
  const right_imgPos: Vec2 = {
    x: (virtualX - leftPanelRect.x) - (leftPanelWidth + gutterWidth),
    y: virtualY - rightPanelRect.y
  };

  return {
    effectiveRect: { x: contentX0, y: contentY0, width: effectiveWidth, height: effectiveHeight },
    sourceCrop,
    imgDisplaySize: { width: imgDisplayWidth, height: imgDisplayHeight },
    left:  { panelRect: leftPanelRect,  imgPositionInPanel: left_imgPos },
    right: { panelRect: rightPanelRect, imgPositionInPanel: right_imgPos }
  };
}
```

---

## Canvas usage sketch

Left page canvas (`W = left.panelRect.width`, `H = left.panelRect.height`):

```js
ctxLeft.save(); // canvas sized exactly to W×H (acts as clip)
ctxLeft.drawImage(
  img,
  sourceCrop.x, sourceCrop.y, sourceCrop.width, sourceCrop.height,
  left.imgPositionInPanel.x, left.imgPositionInPanel.y,
  imgDisplaySize.width, imgDisplaySize.height
);
ctxLeft.restore();
```

Right page canvas (`W = right.panelRect.width`, `H = right.panelRect.height`):

```js
ctxRight.save();
ctxRight.drawImage(
  img,
  sourceCrop.x, sourceCrop.y, sourceCrop.width, sourceCrop.height,
  right.imgPositionInPanel.x, right.imgPositionInPanel.y,
  imgDisplaySize.width, imgDisplaySize.height
);
ctxRight.restore();
```

---

## Sanity checks you should observe

* Increasing `gutterWidth` reduces `effectiveWidth` and thus both `leftPanelWidth/rightPanelWidth` (while the image remains continuous across the seam).
* With `cropToFill = true`, `imgDisplaySize == effectiveRect.size` and panning is via **crop offsets** only.
* With `cropToFill = false`, `imageScale` changes `imgDisplaySize`, and `imagePosition` only pans when the image overhangs the effective area.
* The right panel’s `imgPositionInPanel.x` is always the left panel’s minus `(leftPanelWidth + gutterWidth)`, ensuring a perfect split.

That’s it—you have all container rectangles, final source crop, shared display size, and per‑panel image offsets to produce a seamless two‑page export that matches the preview.
