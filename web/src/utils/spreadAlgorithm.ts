/**
 * Clean, deterministic algorithm for single page or double‑page spread with inner gutter
 * Ported from the Go implementation and specification in 03-algorithm-cleaned-up-gpt-input-prompt.md
 */

// Type definitions
export type Orientation = 'portrait' | 'landscape';
export type PositionUnits = 'px' | 'normalized';

export interface CropRatio { 
  w: number; 
  h: number; 
}

export interface Point {
  x: number;
  y: number;
}

export interface Rect { 
  x: number; 
  y: number; 
  w: number; 
  h: number; 
}

export interface Size { 
  w: number; 
  h: number; 
}

export interface SpreadSizes {
  left: Size;
  right: Size;
}

export interface Inputs {
  // Source image (pixels)
  srcW: number;
  srcH: number;

  // Paper and layout (inches)
  paperWIn: number;
  paperHIn: number;
  orientation: Orientation;
  marginTopIn: number;
  marginRightIn: number;
  marginBottomIn: number;
  marginLeftIn: number;
  dpi: number;

  // Spread / gutter
  isSpread: boolean;
  gutterIn?: number; // inches

  // Crop & scale
  cropRatio?: CropRatio | null; // null means "original"
  cropToFill: boolean;
  userScale: number; // >= 0

  // Position
  imagePosition: Point;
  positionUnits?: PositionUnits; // default 'px'
}

export interface Result {
  // Global rectangles
  srcRectGlobal: Rect; // in S (source space)
  dstRectGlobal: Rect; // in L (layout space)

  // Layout info
  contentRect: Rect;
  effectiveSpreadW: number;
  gutterPx: number;

  // Spread panels (null for single page)
  leftPanel: Rect | null; // visible area on left page in L
  rightPanel: Rect | null; // visible area on right page in L
  dstRectLeft: Rect | null; // destination rect relative to left page canvas
  dstRectRight: Rect | null; // destination rect relative to right page canvas

  // Export sizes
  exportSingle: Size | null;
  exportSpread: SpreadSizes | null;

  // Additional info for rendering
  pageW?: number; // page canvas width (for spreads)
}

// Utility functions
function clamp(v: number, lo: number, hi: number): number {
  return Math.max(lo, Math.min(hi, v));
}

function safeDiv(a: number, b: number): number {
  return b === 0 ? 0 : a / b;
}

/**
 * Main algorithm implementation
 */
export function computePlacement(inputs: Inputs): Result {
  const {
    srcW,
    srcH,
    paperWIn,
    paperHIn,
    orientation,
    marginTopIn,
    marginRightIn,
    marginBottomIn,
    marginLeftIn,
    dpi,
    isSpread,
    gutterIn = 0,
    cropRatio = null,
    cropToFill,
    userScale,
    imagePosition,
    positionUnits = 'px'
  } = inputs;

  // 1) Paper → px (apply orientation), content rect in L
  let pageWpx: number, pageHpx: number;
  if (orientation === 'landscape') {
    pageWpx = paperHIn * dpi;
    pageHpx = paperWIn * dpi;
  } else {
    pageWpx = paperWIn * dpi;
    pageHpx = paperHIn * dpi;
  }

  const spreadWpx = isSpread ? 2 * pageWpx : pageWpx;

  const Mt = marginTopIn * dpi;
  const Mr = marginRightIn * dpi;
  const Mb = marginBottomIn * dpi;
  const Ml = marginLeftIn * dpi;

  const contentW = spreadWpx - (Ml + Mr);
  const contentH = pageHpx - (Mt + Mb);
  const contentRect: Rect = { x: 0, y: 0, w: contentW, h: contentH };

  const gutterPx = isSpread ? Math.max(0, gutterIn * dpi) : 0;
  const effectiveSpreadW = isSpread ? Math.max(0, contentW - gutterPx) : contentW;

  // Target box for global placement (image covers visible width only; gutter is blank)
  const targetW = effectiveSpreadW;
  const targetH = contentH;
  const targetRatio = (targetW > 0 && targetH > 0) ? (targetW / targetH) : 0;

  // 2) Source crop window in S
  const W = srcW;
  const H = srcH;
  const sourceRatio = safeDiv(W, H);
  
  let reqRatio = sourceRatio;
  if (cropRatio != null) {
    reqRatio = safeDiv(cropRatio.w, cropRatio.h);
  } else if (cropToFill) {
    reqRatio = targetRatio;
  }

  let sx = 0, sy = 0, sw = W, sh = H;
  if (W > 0 && H > 0 && reqRatio > 0) {
    if (sourceRatio > reqRatio) {
      // crop width
      sh = H;
      sw = H * reqRatio;
      const rangeX = Math.max(0, W - sw);
      if (positionUnits === 'normalized') {
        const t = clamp((imagePosition.x + 1) / 2, 0, 1);
        sx = t * rangeX;
      } else { // px
        sx = clamp(rangeX / 2 + imagePosition.x, 0, rangeX);
      }
      sy = 0;
    } else if (sourceRatio < reqRatio) {
      // crop height
      sw = W;
      sh = safeDiv(W, reqRatio);
      const rangeY = Math.max(0, H - sh);
      if (positionUnits === 'normalized') {
        const t = clamp((imagePosition.y + 1) / 2, 0, 1);
        sy = t * rangeY;
      } else {
        sy = clamp(rangeY / 2 + imagePosition.y, 0, rangeY);
      }
      sx = 0;
    } else {
      // ratios equal: no crop
      sx = 0; sy = 0; sw = W; sh = H;
    }
  }

  // 3) Scale and position in L
  const scaleX = sw > 0 ? targetW / sw : 0;
  const scaleY = sh > 0 ? targetH / sh : 0;

  let finalScale = 0;
  let tx = 0, ty = 0;

  if (cropToFill) {
    const coverage = Math.max(scaleX, scaleY);
    finalScale = coverage * Math.max(1, userScale); // allow zoom‑in, never below coverage
    tx = 0;
    ty = 0;
  } else {
    // fit: preserve entire (possibly cropped) image
    finalScale = Math.min(scaleX, scaleY) * userScale;
    const dwTmp = sw * finalScale;
    const dhTmp = sh * finalScale;
    if (positionUnits === 'normalized') {
      tx = imagePosition.x * (targetW - dwTmp) / 2;
      ty = imagePosition.y * (targetH - dhTmp) / 2;
    } else {
      tx = imagePosition.x;
      ty = imagePosition.y;
    }
  }

  const dw = sw * finalScale;
  const dh = sh * finalScale;

  const centerX = targetW / 2;
  const centerY = targetH / 2;
  const dx = centerX - dw / 2 + tx;
  const dy = centerY - dh / 2 + ty;

  const srcRectGlobal: Rect = { x: sx, y: sy, w: sw, h: sh };
  const dstRectGlobal: Rect = { x: dx, y: dy, w: dw, h: dh };

  // 4) Spread pages and inner gutter (only if isSpread === true)
  let leftPanel: Rect | null = null;
  let rightPanel: Rect | null = null;
  let dstRectLeft: Rect | null = null;
  let dstRectRight: Rect | null = null;
  let pageW: number | undefined;

  if (isSpread) {
    // Page canvas width is half of content width
    pageW = contentW / 2;
    const m = gutterPx / 2;
    
    // Visible areas (clip rectangles) in L are:
    leftPanel = { x: 0, y: 0, w: pageW - m, h: contentH };
    rightPanel = { x: pageW + m, y: 0, w: pageW - m, h: contentH };

    // Per‑page destination rectangles, page‑relative:
    const pageLeftOriginX = 0;
    const pageRightOriginX = pageW;

    dstRectLeft = { x: dx - pageLeftOriginX, y: dy, w: dw, h: dh };
    dstRectRight = { x: dx - pageRightOriginX, y: dy, w: dw, h: dh };
  }

  // 5) Export sizes
  let exportSingle: Size | null = null;
  let exportSpread: SpreadSizes | null = null;

  if (!isSpread) {
    exportSingle = { w: Math.round(contentW), h: Math.round(contentH) };
  } else {
    // Export each page at page canvas width (includes inner gutter margins)
    const lw = Math.round(contentW / 2);
    const rw = lw;
    exportSpread = {
      left: { w: lw, h: Math.round(contentH) },
      right: { w: rw, h: Math.round(contentH) }
    };
  }

  return {
    srcRectGlobal,
    dstRectGlobal,
    contentRect,
    effectiveSpreadW,
    gutterPx,
    leftPanel,
    rightPanel,
    dstRectLeft,
    dstRectRight,
    exportSingle,
    exportSpread,
    pageW
  };
}

/**
 * Helper function to convert preview dimensions to actual dimensions
 * This is useful for the preview canvas where we scale down for display
 */
export function scaleResultForPreview(result: Result, previewScale: number): Result {
  const scaleRect = (rect: Rect | null): Rect | null => {
    if (!rect) return null;
    return {
      x: rect.x * previewScale,
      y: rect.y * previewScale,
      w: rect.w * previewScale,
      h: rect.h * previewScale
    };
  };

  const scaleSize = (size: Size | null): Size | null => {
    if (!size) return null;
    return {
      w: Math.round(size.w * previewScale),
      h: Math.round(size.h * previewScale)
    };
  };

  return {
    ...result,
    srcRectGlobal: scaleRect(result.srcRectGlobal)!,
    dstRectGlobal: scaleRect(result.dstRectGlobal)!,
    contentRect: scaleRect(result.contentRect)!,
    effectiveSpreadW: result.effectiveSpreadW * previewScale,
    gutterPx: result.gutterPx * previewScale,
    leftPanel: scaleRect(result.leftPanel),
    rightPanel: scaleRect(result.rightPanel),
    dstRectLeft: scaleRect(result.dstRectLeft),
    dstRectRight: scaleRect(result.dstRectRight),
    exportSingle: scaleSize(result.exportSingle),
    exportSpread: result.exportSpread ? {
      left: scaleSize(result.exportSpread.left)!,
      right: scaleSize(result.exportSpread.right)!
    } : null,
    pageW: result.pageW ? result.pageW * previewScale : undefined
  };
}
