# Spread Rendering & Visualization Guide
**Date:** October 11, 2025  
**Reference:** `02-image-resizer-code.tsx` (lines 128-611)  
**Purpose:** Document how spread rendering works with multiple output variants and debug visualization

---

## Overview

When a **Page Layout Template has `is_spread: true`**, a single wide laid-out image is split into left and right facing pages for book binding. The system must support rendering multiple variants for different use cases.

---

## Spread Rendering Variants

### Variant 1: Left Page Only

**Use case:** Print left page separately

```
┌─────────────────────────┐
│                         │
│  ┌─────────────────┐   │
│  │                 │   │
│  │  Left Half of   │   │
│  │  Wide Image     │   │
│  │                 │   │
│  │  (Extends into  │   │
│  │   gutter by     │   │
│  │   overlap amt)  │   │
│  │                 │   │
│  └─────────────────┘   │
│         ↑               │
│    overlap zone         │
│  (gets hidden when      │
│   bound)                │
└─────────────────────────┘

Size: 2437×3000 px
(8.125" × 10" @ 300 DPI)
(Base 8" + 0.125" overlap)
```

**With debug=true:**
```
┌─────────────────────────┐
│                         │
│  ┌─────────────────┐   │
│  │                 │   │
│  │  Left Half      │   │
│  │                 │   │
│  │                 │   │
│  │                 ┊   │  ← Red dashed line
│  │                 ┊   │     at overlap boundary
│  │                 │   │
│  └─────────────────┘   │
│                    ↑    │
│              overlap    │
│              0.125"     │
└─────────────────────────┘
```

**API Call:**
```
GET /api/laid-out-pages/{id}/export?variant=left&debug=true
```

---

### Variant 2: Right Page Only

**Use case:** Print right page separately

```
┌─────────────────────────┐
│                         │
│   ┌─────────────────┐  │
│   │                 │  │
│   │  Right Half of  │  │
│   │  Wide Image     │  │
│   │                 │  │
│   │  (Extends into  │  │
│   │   gutter by     │  │
│   │   overlap amt)  │  │
│   │                 │  │
│   └─────────────────┘  │
│   ↑                     │
│ overlap zone            │
│ (gets hidden when bound)│
│                         │
└─────────────────────────┘

Size: 2437×3000 px
(8.125" × 10" @ 300 DPI)
```

**With debug=true:**
```
┌─────────────────────────┐
│   ┊                     │  ← Red dashed line
│   ┊                     │     at overlap boundary
│   ┊──────────────────┐  │
│   │                 │  │
│   │  Right Half     │  │
│   │                 │  │
│   │                 │  │
│   │                 │  │
│   │                 │  │
│   └─────────────────┘  │
│   ↑                     │
│ overlap                 │
│                         │
└─────────────────────────┘
```

**API Call:**
```
GET /api/laid-out-pages/{id}/export?variant=right&debug=true
```

---

### Variant 3: Combined Spread (Print Preview)

**Use case:** Preview how pages look when bound together

```
┌─────────────────┐ ║ ┌─────────────────┐
│                 │ ║ │                 │
│  Left Page      │ ║ │  Right Page     │
│  (with overlap) │ ║ │  (with overlap) │
│                 │ ║ │                 │
│                 │→║←│                 │
│  Overlap extends│ ║ │Overlap extends  │
│  into gap       │ ║ │into gap         │
│                 │ ║ │                 │
└─────────────────┘ ║ └─────────────────┘
                    ↑
              Binding gap
           (4px for visualization)

Size: 4878×3000 px
(2437 + 4 + 2437)
Shows how pages sit when bound
```

**Purpose:**
- Gap represents the physical binding
- Shows how overlap areas meet in the middle
- Verifies gutter calculations are correct

**API Call:**
```
GET /api/laid-out-pages/{id}/export?variant=combined
```

---

### Variant 4: Full Spread with Gutter Visualization

**Use case:** Debug gutter placement, verify calculations

```
┌────────────────────────────────────────────────────┐
│                                                    │
│  ┌──────────────────────┐ ║ ┌──────────────────┐  │
│  │                      │▓║▓│                  │  │
│  │  Wide Image          │▓║▓│  Spanning Both   │  │
│  │  Rendered on         │▓║▓│  Pages          │  │
│  │  Full Spread         │▓║▓│                  │  │
│  │  Canvas              │▓║▓│                  │  │
│  │                      │▓║▓│                  │  │
│  │                      │▓║▓│                  │  │
│  └──────────────────────┘ ║ └──────────────────┘  │
│                           ↑                        │
│                     GUTTER label                   │
│                  (red overlay 15% opacity)         │
│                                                    │
└────────────────────────────────────────────────────┘

Size: 4800×3000 px (full spread, uncut)

Visual elements:
- Red semi-transparent overlay on gutter area
- Red dashed line at gutter center
- "GUTTER" text label
- Shows where the split will happen
```

**Purpose:**
- Verify gutter is in correct position
- Check overlap calculations
- Debug before printing

**API Call:**
```
GET /api/laid-out-pages/{id}/export?variant=full&debug=true
```

---

## Code Reference from 02-image-resizer-code.tsx

### Gutter Visualization (Lines 464-478)

```typescript
// Draw gutter visualization
ctx.fillStyle = 'rgba(255, 0, 0, 0.15)';  // Semi-transparent red
ctx.fillRect(gutterX - gutterWidth / 2, 0, gutterWidth, Ht);

ctx.strokeStyle = 'rgba(255, 0, 0, 0.8)';  // Bright red
ctx.lineWidth = 2;
ctx.setLineDash([10, 5]);  // Dashed line
ctx.beginPath();
ctx.moveTo(gutterX, 0);
ctx.lineTo(gutterX, Ht);
ctx.stroke();
ctx.setLineDash([]);

// Add label
ctx.fillStyle = 'rgba(255, 0, 0, 0.8)';
ctx.font = 'bold 14px sans-serif';
ctx.fillText('GUTTER', gutterX - 30, 30);
```

### Overlap Indicator Lines (Lines 180-189, 213-221)

```typescript
// On left page - show where overlap ends
const gutterLineX = leftPageWidth - gutterOverlap;
ctx.strokeStyle = 'rgba(255, 0, 0, 0.6)';
ctx.lineWidth = 2;
ctx.setLineDash([10, 5]);
ctx.beginPath();
ctx.moveTo(gutterLineX, 0);
ctx.lineTo(gutterLineX, Ht);
ctx.stroke();
ctx.setLineDash([]);
```

### Page Borders (Lines 192-194)

```typescript
ctx.strokeStyle = '#333';
ctx.lineWidth = 2;
ctx.strokeRect(0, 0, leftPageWidth, Ht);
```

---

## Export API Design

### Endpoint

```
GET /api/laid-out-pages/{id}/export?variant={variant}&debug={debug}
```

### Query Parameters

| Parameter | Values | Default | Purpose |
|-----------|--------|---------|---------|
| `variant` | `single`, `left`, `right`, `combined`, `full` | Auto (based on is_spread) | Which render to return |
| `debug` | `true`, `false` | `false` | Add visualization overlays |

### Behavior Matrix

| Is Spread? | Variant | Returns | Debug Adds |
|------------|---------|---------|------------|
| No | (any) | Single page PNG | Crop marks, margins overlay |
| Yes | `left` | Left page PNG (2437×3000) | Overlap dashed line |
| Yes | `right` | Right page PNG (2437×3000) | Overlap dashed line |
| Yes | `combined` | Both pages with gap | Shows binding area |
| Yes | `full` | Uncut spread (4800×3000) | Gutter overlay + label |

### Response Headers

```
Content-Type: image/png
Content-Disposition: attachment; filename="page-{id}-{variant}.png"
Content-Length: {size}
```

---

## Implementation Checklist

### Required for Basic Spreads
- [ ] Compute gutter position (center of spread)
- [ ] Calculate overlap amount in pixels
- [ ] Extract left page (0 to gutterX + overlap)
- [ ] Extract right page (gutterX - overlap to end)
- [ ] Save both pages as separate PNGs

### Required for Visualization
- [ ] Implement `drawDashedLine()` helper
- [ ] Implement `drawBorder()` helper
- [ ] Implement `drawFilledRect()` helper
- [ ] Add gutter overlay to full spread
- [ ] Add overlap indicators to split pages
- [ ] Add "GUTTER" text label (optional, needs font library)

### Optional Enhancements
- [ ] Support different gutter positions (left, right, not just center)
- [ ] Configurable gap width for combined view
- [ ] Color-coded overlays (red = gutter, blue = margins, green = bleed)
- [ ] Measurement annotations (show dimensions on image)

---

## Testing Strategy

### Visual Verification

**1. Create test spread:**
```bash
# Create wide image (16×10") with obvious left/right difference
# (e.g., gradient from blue on left to red on right)

# Create page template with spread mode
curl -X POST http://localhost:8088/api/page-templates \
  -d '{
    "name": "Test Spread",
    "template": {
      "page_width_in": 8,
      "page_height_in": 10,
      "dpi": 300,
      "is_spread": true,
      "gutter_width_in": 0.25,
      "gutter_overlap_in": 0.125,
      "positioning_mode": "fill"
    }
  }'

# Create print page
curl -X POST http://localhost:8088/api/projects/prj-.../laid-out-pages \
  -d '{"page_template_id":"ptpl-...","laid_out_image_id":"loi-..."}'

# Export all variants
curl "http://localhost:8088/api/laid-out-pages/lpg-.../export?variant=left" -o left.png
curl "http://localhost:8088/api/laid-out-pages/lpg-.../export?variant=right" -o right.png
curl "http://localhost:8088/api/laid-out-pages/lpg-.../export?variant=combined" -o combined.png
curl "http://localhost:8088/api/laid-out-pages/lpg-.../export?variant=full&debug=true" -o debug.png
```

**2. Verify outputs:**
- Left page should show left half + overlap on right edge
- Right page should show right half + overlap on left edge
- Combined should show both pages with visible gap
- Debug should show red gutter overlay on full spread

**3. Measure dimensions:**
```bash
# Should be exactly:
file left.png    # 2437×3000 px (8.125" × 10" @ 300 DPI)
file right.png   # 2437×3000 px
file combined.png # ~4878×3000 px (with gap)
file debug.png   # 4800×3000 px (full spread)
```

### Physical Print Test

**1. Print test pages:**
```
Print left.png and right.png on separate sheets
```

**2. Align and tape together:**
```
Place pages side-by-side
The overlap zones should align perfectly
Tape on back to simulate binding
```

**3. Verify:**
```
✓ No gaps visible at binding
✓ Image continuous across both pages
✓ Overlap creates seamless join
```

**If there's a gap:** Overlap amount is too small  
**If there's overlap showing:** Overlap amount is too large  
**Typical value:** 0.125" (1/8 inch) works for most binding types

---

## UI Preview Requirements

### In PageLayoutsTab

When user edits a spread template, show preview with:

**For spread mode:**
```
┌─ Preview ──────────────────────────────────────┐
│                                                │
│  ┌─ Left ──────┐  ║  ┌─ Right ──────┐        │
│  │              │  ║  │              │        │
│  │  Preview of  │  ║  │  Preview of  │        │
│  │  left page   │  ║  │  right page  │        │
│  │              │  ║  │              │        │
│  │           ┊  │  ║  │  ┊           │        │
│  │           ┊  │  ║  │  ┊           │        │
│  └──────────────┘  ║  └──────────────┘        │
│             ↑      ║      ↑                    │
│          overlap   ║   overlap                │
│                                                │
│  [ ] Show gutter visualization                │
│  [ ] Show overlap indicators                  │
│                                                │
│  Download: [Left] [Right] [Combined] [Debug]  │
│                                                │
└────────────────────────────────────────────────┘
```

**Toggle options:**
- Show gutter visualization (red overlay on full spread)
- Show overlap indicators (dashed lines on split pages)
- Download specific variants

---

## Data Flow

### Creating a Spread Print Page

```
1. User uploads wide panorama photo (6000×3000 px)
   ↓
2. Creates image layout template:
   - paper_width_in: 16 (for 16" wide spread)
   - paper_height_in: 10
   - dpi: 300
   - crop/scale settings
   ↓
3. Applies template → Laid-Out Image
   - Result: 4800×3000 px cropped/scaled image
   ↓
4. Creates page layout template:
   - page_width_in: 8 (each page is 8" wide)
   - page_height_in: 10
   - dpi: 300
   - is_spread: true
   - gutter_width_in: 0.25
   - gutter_overlap_in: 0.125
   ↓
5. Applies page template to laid-out image → Print Page
   - Computes split position
   - Calculates overlap zones
   - Stores result JSON with all dimensions
   ↓
6. Renders print page:
   - Loads 4800×3000 laid-out image
   - Creates full spread canvas
   - Draws image on canvas
   - Splits at x=2400 (center)
   - Extracts left (0 to 2437)
   - Extracts right (2363 to 4800)
   - Saves: left.png, right.png, combined.png, debug.png
   ↓
7. User downloads variants:
   - left.png → Send to printer for left page
   - right.png → Send to printer for right page
   - combined.png → Preview how it looks bound
   - debug.png → Verify gutter placement
```

---

## Gutter Math Explained

### The Problem

When binding two pages together, some of the image gets "lost" in the binding area. We compensate with overlap.

```
Without overlap:
┌─────────┐   ┌─────────┐
│  Left   │ X │  Right  │  ← Gap appears at binding!
│  8.0"   │   │  8.0"   │
└─────────┘   └─────────┘
         ↑ Bad! ↑

With overlap:
┌─────────────┐ ┌─────────────┐
│  Left       │ │       Right │
│  8.125"     │ │     8.125"  │
│          ┊──┼─┼──┊          │
└─────────────┘ └─────────────┘
         ↑ Good! ↑
    Overlaps hide in binding
```

### The Calculation

```go
// For 16×10" spread @ 300 DPI with 0.125" overlap:

totalSpreadWidthPx := 16 * 300 = 4800
pageWidthPx := 8 * 300 = 2400
gutterOverlapPx := 0.125 * 300 = 37 (rounded)

// Gutter at center
gutterPositionPx := 4800 / 2 = 2400

// Each page extends into gutter
leftPageWidthPx := 2400 + 37 = 2437
rightPageWidthPx := 2400 + 37 = 2437

// Extract regions from full spread
leftRegion := image.Rect(0, 0, 2437, 3000)          // Start to (center + overlap)
rightRegion := image.Rect(2363, 0, 4800, 3000)      // (center - overlap) to end
```

**Verification:**
```
Left start: 0
Left end: 2437
Left width: 2437

Right start: 2363 (= 2400 - 37)
Right end: 4800
Right width: 4800 - 2363 = 2437 ✓

Overlap zone: 2363 to 2437 = 74px total
Each page overlaps by: 74 / 2 = 37px ✓
```

---

## Common Issues & Solutions

### Issue 1: Gap at Binding

**Symptom:** White gap visible between pages when bound

**Cause:** Overlap too small or extraction regions wrong

**Fix:**
```go
// Verify extraction includes overlap
leftEnd := gutterX + overlapPx    // Should be 2437
rightStart := gutterX - overlapPx  // Should be 2363
```

### Issue 2: Duplicate Content in Middle

**Symptom:** Image appears "doubled" at binding area

**Cause:** Overlap too large or gutter position wrong

**Fix:**
```go
// Verify gutter at exact center
gutterX := totalSpreadWidthPx / 2  // For 4800, should be 2400
```

### Issue 3: Pages Don't Align

**Symptom:** When placed together, image doesn't match

**Cause:** Different DPI or rounding errors

**Fix:**
```go
// Ensure all calculations use integer pixels
gutterX := int(totalSpreadWidthIn * dpi / 2)
overlapPx := int(gutterOverlapIn * dpi)
```

### Issue 4: Visualization Obscures Content

**Symptom:** Can't see image under red overlay

**Fix:**
```go
// Use low opacity for overlays
gutterOverlayColor := color.RGBA{255, 0, 0, 38}  // 15% opacity
// OR: Render visualization on separate layer
```

---

## Frontend Download UI

### Suggested UI for Spread Export

```
┌─ Export Print Page ────────────────────────────┐
│                                                │
│  Page: img01-spread.png (16×10" spread)       │
│                                                │
│  Export Options:                               │
│                                                │
│  ┌─ Spread Variants ────────────────────────┐ │
│  │                                           │ │
│  │  ◉ Both pages separately (for printing)  │ │
│  │    [Download Left] [Download Right]      │ │
│  │                                           │ │
│  │  ○ Combined preview (visualization)      │ │
│  │    [Download Combined]                   │ │
│  │                                           │ │
│  │  ○ Full spread (debug)                   │ │
│  │    [Download Full Spread]                │ │
│  │                                           │ │
│  └───────────────────────────────────────────┘ │
│                                                │
│  Debug Options:                                │
│  [✓] Show gutter overlay                      │
│  [✓] Show overlap indicators                  │
│  [ ] Show crop marks                          │
│  [ ] Include bleed                            │
│                                                │
│  [Download All Variants as ZIP]               │
│                                                │
└────────────────────────────────────────────────┘
```

---

## Summary

**For single pages:** Simple - one render output

**For spreads:** Four possible outputs
1. **Left page** (2437×3000) - For printing
2. **Right page** (2437×3000) - For printing  
3. **Combined** (4878×3000) - For preview
4. **Full+debug** (4800×3000) - For debugging

**Visualization helps:**
- Users understand what overlap means
- Designers verify gutter placement
- Printers see exact page boundaries
- Debugging layout issues

**Reference:** Study `02-image-resizer-code.tsx` for complete client-side implementation. Server implementation follows same pattern using Go image libraries.

---

**END OF GUIDE**

