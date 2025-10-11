# Page Layout Tab - ASCII UI Design
**Date:** October 11, 2025  
**Purpose:** Define how a single laid-out image is positioned on a physical print page

---

## Concept Clarification

**Page Layout Templates** specify how to place ONE laid-out image onto a physical page for printing. This is the final step before exporting print-ready files.

**Key differences from Image Layouts:**
- **Image Layout**: Asset + crop/scale settings → Laid-out image (digital)
- **Page Layout**: Laid-out image + page settings → Print-ready page (physical)

**Supports:**
- Single pages (portrait/landscape)
- Spreads (two facing pages from one wide layout)
- Gutters (for binding area in spreads)
- Precise positioning on page
- Margins and bleed areas

---

## Tab Layout: Page Layouts

```
┌────────────────────────────────────────────────────────────────────────┐
│  Page Layouts                                                           │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ═══ SECTION 1: Page Layout Templates ═══                              │
│                                                                         │
│  Page Templates (5)                              [+ Create Template]   │
│                                                                         │
│  ┌─ Template Library ──────────────────────────────────────────────┐  │
│  │                                                                   │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │  │
│  │  │ 8×10"        │  │ Spread       │  │ Full Bleed   │          │  │
│  │  │ Portrait     │  │ 16×10"       │  │ Letter       │          │  │
│  │  │              │  │              │  │              │          │  │
│  │  │ [   📄   ]   │  │ [  📄📄  ]   │  │ [   📄   ]   │          │  │
│  │  │              │  │              │  │              │          │  │
│  │  │ Single page  │  │ Spread mode  │  │ Single page  │          │  │
│  │  │ 0.5" margins │  │ 0.25" gutter │  │ 0" margins   │          │  │
│  │  │ 300 DPI      │  │ 300 DPI      │  │ 300 DPI      │          │  │
│  │  │              │  │              │  │              │          │  │
│  │  │ [Edit]       │  │ [Edit]       │  │ [Edit]       │          │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘          │  │
│  │                                                                   │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─ When "Create" or "Edit" clicked ──────────────────────────────┐   │
│  │                                                                  │   │
│  │  ┌── Page Layout Template Editor ────────────────────────────┐ │   │
│  │  │                                                            │ │   │
│  │  │  Create Page Layout Template                   [✕ Close]  │ │   │
│  │  │  ──────────────────────────────────────────────────────  │ │   │
│  │  │                                                            │ │   │
│  │  │  ┌─ Left: Settings ────┐  ┌─ Right: Preview ──────────┐ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ Template Name:       │  │  ┌──────────────────────┐ │ │ │   │
│  │  │  │ [______________]     │  │  │                      │ │ │ │   │
│  │  │  │                      │  │  │  ┌────────────────┐ │ │ │ │   │
│  │  │  │ Description:         │  │  │  │                │ │ │ │ │   │
│  │  │  │ [______________]     │  │  │  │  Laid-Out     │ │ │ │ │   │
│  │  │  │                      │  │  │  │  Image        │ │ │ │ │   │
│  │  │  │ ═══ Page Setup ═══   │  │  │  │  Preview      │ │ │ │ │   │
│  │  │  │                      │  │  │  │                │ │ │ │ │   │
│  │  │  │ Page Size:           │  │  │  │  (positioned  │ │ │ │ │   │
│  │  │  │ [8×10" Portrait ▾]   │  │  │  │   on page)    │ │ │ │ │   │
│  │  │  │   or Custom:         │  │  │  │                │ │ │ │ │   │
│  │  │  │   W: [8.0] in        │  │  │  └────────────────┘ │ │ │ │   │
│  │  │  │   H: [10.0] in       │  │  │    ↑ page margins   │ │ │ │ │   │
│  │  │  │                      │  │  └──────────────────────┘ │ │ │   │
│  │  │  │ DPI: [300_____]      │  │                            │ │ │   │
│  │  │  │      72    600       │  │  Page: 2400×3000 px       │ │ │ │   │
│  │  │  │                      │  │  Image: 2250×2700 px      │ │ │ │   │
│  │  │  │ [ ] Spread Mode      │  │                            │ │ │   │
│  │  │  │                      │  │  [← Select Preview]        │ │ │   │
│  │  │  │ ═══ Margins ═══      │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ [✓] Uniform margins  │  │                            │ │ │   │
│  │  │  │   All: [0.5] in      │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ [ ] Individual:      │  │                            │ │ │   │
│  │  │  │   Top: [0.5] in      │  │                            │ │ │   │
│  │  │  │   Right: [0.5] in    │  │                            │ │ │   │
│  │  │  │   Bottom: [0.5] in   │  │                            │ │ │   │
│  │  │  │   Left: [0.5] in     │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ ═══ Spread Settings  │  │                            │ │ │   │
│  │  │  │     (if spread) ═══  │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ Gutter Width:        │  │                            │ │ │   │
│  │  │  │   [0.25] in          │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ Overlap into gutter: │  │                            │ │ │   │
│  │  │  │   [0.125] in         │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ ═══ Image Position   │  │                            │ │ │   │
│  │  │  │        ═══           │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ Positioning:         │  │                            │ │ │   │
│  │  │  │ ◉ Fill content area  │  │                            │ │ │   │
│  │  │  │ ○ Absolute position  │  │                            │ │ │   │
│  │  │  │ ○ Snap to margins    │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ [if Absolute:]       │  │                            │ │ │   │
│  │  │  │   X: [4.0] in        │  │                            │ │ │   │
│  │  │  │   Y: [5.0] in        │  │                            │ │ │   │
│  │  │  │   W: [7.0] in        │  │                            │ │ │   │
│  │  │  │   H: [9.0] in        │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  │ [if Snap to margins:]│  │                            │ │ │   │
│  │  │  │   [TL][TC][TR]       │  │                            │ │ │   │
│  │  │  │   [ML][MC][MR]       │  │                            │ │ │   │
│  │  │  │   [BL][BC][BR]       │  │                            │ │ │   │
│  │  │  │     ^^ fills space   │  │                            │ │ │   │
│  │  │  │                      │  │                            │ │ │   │
│  │  │  └──────────────────────┘  └────────────────────────────┘ │ │   │
│  │  │                                                            │ │   │
│  │  │  [Cancel]  [Save Template]                               │ │   │
│  │  │                                                            │ │   │
│  │  └────────────────────────────────────────────────────────────┘ │   │
│  │                                                                  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ═══ SECTION 2: Apply Template to Create Print Pages ═══               │
│                                                                         │
│  Print-Ready Pages (12)                           [+ Create Page]      │
│                                                                         │
│  ┌─ Quick Create ──────────────────────────────────────────────────┐  │
│  │                                                                   │  │
│  │  Laid-Out Image: [img01-8x10-portrait ▾]                         │  │
│  │  Page Template:  [8×10" Portrait with margins ▾]                 │  │
│  │                                                                   │  │
│  │  [Create Print Page]                                             │  │
│  │                                                                   │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─ Print Pages Grid ──────────────────────────────────────────────┐  │
│  │                                                                   │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │  │
│  │  │┌──────────┐│  │┌──────────┐│  │┌──────────┐│                 │  │
│  │  ││          ││  ││          ││  ││          ││                 │  │
│  │  ││  Page    ││  ││  Page    ││  ││  Spread  ││                 │  │
│  │  ││  Preview ││  ││  Preview ││  ││  Preview ││                 │  │
│  │  ││          ││  ││          ││  ││          ││                 │  │
│  │  │└──────────┘│  │└──────────┘│  │└──────────┘│                 │  │
│  │  │            │  │            │  │            │                 │  │
│  │  │ Page 1     │  │ Page 2     │  │ Spread 1   │                 │  │
│  │  │ img01.png  │  │ img02.png  │  │ img03.png  │                 │  │
│  │  │ 8×10"      │  │ 8×10"      │  │ 16×10"     │                 │  │
│  │  │            │  │            │  │            │                 │  │
│  │  │ [View][✕]  │  │ [View][✕]  │  │ [View][✕]  │                 │  │
│  │  └────────────┘  └────────────┘  └────────────┘                 │  │
│  │                                                                   │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─ When "View" clicked on a print page ─────────────────────────┐    │
│  │                                                                 │    │
│  │  ┌── Print Page Detail ──────────────────────────────────────┐ │   │
│  │  │                                                            │ │   │
│  │  │  Page 1: img01.png on 8×10" Portrait      [Export] [✕]   │ │   │
│  │  │  ────────────────────────────────────────────────────────│ │   │
│  │  │                                                            │ │   │
│  │  │  ┌─ Preview ──────────┐  ┌─ Details ──────────────────┐ │ │   │
│  │  │  │                     │  │                             │ │ │   │
│  │  │  │  ┌───────────────┐ │  │ Source Laid-Out Image:      │ │ │   │
│  │  │  │  │:::::::::::::::│ │  │ img01.png (8×10 Portrait)   │ │ │   │
│  │  │  │  │:::::::::::::::│ │  │                             │ │ │   │
│  │  │  │  │::: Image :::::│ │  │ Page Template:              │ │ │   │
│  │  │  │  │::: on Page :::│ │  │ 8×10" Portrait with margins │ │ │   │
│  │  │  │  │:::::::::::::::│ │  │                             │ │ │   │
│  │  │  │  │:::::::::::::::│ │  │ Page Dimensions:            │ │ │   │
│  │  │  │  └───────────────┘ │  │ 2400 × 3000 px @ 300 DPI   │ │ │   │
│  │  │  │   ↑ margins shown   │  │                             │ │ │   │
│  │  │  │                     │  │ Content Area:               │ │ │   │
│  │  │  │  Shows crop marks,  │  │ 2250 × 2850 px             │ │ │   │
│  │  │  │  bleed, page border │  │ (after margins)             │ │ │   │
│  │  │  │                     │  │                             │ │ │   │
│  │  │  │  [Download PNG]     │  │ Spread: No                 │ │ │   │
│  │  │  │  [Download PDF]     │  │ Gutter: N/A                │ │ │   │
│  │  │  │                     │  │                             │ │ │   │
│  │  │  └─────────────────────┘  └─────────────────────────────┘ │ │   │
│  │  │                                                            │ │   │
│  │  │  [< Back to Grid]  [Add to Zine]                         │ │   │
│  │  │                                                            │ │   │
│  │  └────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Template Editor Detail: Spread Mode

When "Spread Mode" checkbox is enabled:

```
┌─ Spread Mode Settings ──────────────────────────────────────────────┐
│                                                                      │
│  [✓] Spread Mode (two facing pages from one wide image)            │
│                                                                      │
│  Page Size (each page):                                             │
│    Width:  [8.0] in    Height: [10.0] in                           │
│                                                                      │
│  → Total spread width: 16.0" (2 pages × 8.0")                      │
│                                                                      │
│  ═══ Gutter Settings ═══                                            │
│                                                                      │
│  Gutter Width: [0.25___] in                                         │
│                0      1                                              │
│                                                                      │
│  Overlap into Gutter: [0.125___] in                                 │
│                       0      0.5                                     │
│  (How much each page extends into binding area)                     │
│                                                                      │
│  ═══ Output Pages ═══                                               │
│                                                                      │
│  Left Page:  2400 px + 37 px (overlap) = 2437 × 3000 px            │
│  Right Page: 2400 px + 37 px (overlap) = 2437 × 3000 px            │
│  Combined:   16.0" × 10.0" @ 300 DPI = 4800 × 3000 px              │
│                                                                      │
│  ┌─ Visual Diagram ───────────────────────────────────────────────┐ │
│  │                                                                  │ │
│  │     ┌─────────────┐ ║ ┌─────────────┐                          │ │
│  │     │             │ ║ │             │                          │ │
│  │     │  Left Page  │ ║ │ Right Page  │                          │ │
│  │     │             │→║←│             │  ← overlap into gutter   │ │
│  │     └─────────────┘ ║ └─────────────┘                          │ │
│  │                     ↑                                            │ │
│  │                  gutter                                          │ │
│  │                                                                  │ │
│  └──────────────────────────────────────────────────────────────────┘ │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Image Positioning Options

### Option 1: Fill Content Area (Default)

```
┌─ Fill Content Area ─────────────────────────────────────┐
│                                                          │
│  ◉ Fill content area (default)                         │
│                                                          │
│  The laid-out image will be scaled to fill the         │
│  content area (page minus margins) exactly.             │
│                                                          │
│  ┌─ Page ─────────────────────────────────────────┐    │
│  │ ← margin →                                      │    │
│  │ ↑         ┌─────────────────────────────────┐  │    │
│  │ m         │                                 │  │    │
│  │ a         │  Laid-Out Image                │  │    │
│  │ r         │  fills this area                │  │    │
│  │ g         │  (respects aspect ratio)        │  │    │
│  │ i         │                                 │  │    │
│  │ n         └─────────────────────────────────┘  │    │
│  │ ↓                                               │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Option 2: Absolute Position

```
┌─ Absolute Position ─────────────────────────────────────┐
│                                                          │
│  ○ Absolute position (precise placement)                │
│                                                          │
│  Position from top-left corner:                         │
│    X: [1.0] in    Y: [1.5] in                          │
│                                                          │
│  Image dimensions on page:                              │
│    Width: [6.0] in    Height: [7.0] in                 │
│                                                          │
│  ┌─ Page ─────────────────────────────────────────┐    │
│  │                                                 │    │
│  │  ↓ Y=1.5"                                       │    │
│  │  → X=1.0"                                       │    │
│  │     ┌──────────────────┐                        │    │
│  │     │                  │ ← W=6.0"              │    │
│  │     │  Laid-Out Image  │                        │    │
│  │     │  at exact size   │ ↕ H=7.0"              │    │
│  │     │  and position    │                        │    │
│  │     └──────────────────┘                        │    │
│  │                                                 │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Option 3: Snap to Margins

```
┌─ Snap to Margins ───────────────────────────────────────┐
│                                                          │
│  ○ Snap to margins (9-point grid)                       │
│                                                          │
│  Select alignment within content area:                  │
│                                                          │
│    [TL][TC][TR]    TL = Top-Left, etc.                 │
│    [ML][MC][MR]    Image fills content area            │
│    [BL][BC][BR]    and aligns to selected point        │
│           ^^                                             │
│        selected                                          │
│                                                          │
│  ┌─ Page ─────────────────────────────────────────┐    │
│  │                                                 │    │
│  │  margin → ┌─────────────────────────────┐      │    │
│  │           │                             │      │    │
│  │           │  Image snapped to           │      │    │
│  │           │  bottom-right of            │      │    │
│  │           │  content area               │      │    │
│  │           │                             │      │    │
│  │           └─────────────────────────────┘      │    │
│  │                                     ↑ margin    │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## Spread Preview Examples

### Single Page Preview

```
┌─ Single Page (8×10" @ 300 DPI) ─────────────────────────┐
│                                                          │
│  ┌─ Crop Marks ─────────────────────────────────────┐   │
│  │                                                   │   │
│  │  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  │   │
│  │  ┃ ← bleed area                                ┃  │   │
│  │  ┃                                             ┃  │   │
│  │  ┃  ┌─────────────────────────────────────┐   ┃  │   │
│  │  ┃  │ ← margin                            │   ┃  │   │
│  │  ┃  │                                     │   ┃  │   │
│  │  ┃  │  ╔═══════════════════════════════╗ │   ┃  │   │
│  │  ┃  │  ║                               ║ │   ┃  │   │
│  │  ┃  │  ║    Laid-Out Image            ║ │   ┃  │   │
│  │  ┃  │  ║    (from Image Layouts)       ║ │   ┃  │   │
│  │  ┃  │  ║                               ║ │   ┃  │   │
│  │  ┃  │  ╚═══════════════════════════════╝ │   ┃  │   │
│  │  ┃  │                                     │   ┃  │   │
│  │  ┃  └─────────────────────────────────────┘   ┃  │   │
│  │  ┃                                             ┃  │   │
│  │  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  │   │
│  │                                                   │   │
│  └───────────────────────────────────────────────────┘   │
│                                                          │
│  Legend:                                                 │
│  ┏━┓ = Page boundary with bleed                         │
│  ┌─┐ = Content area (page - margins)                    │
│  ╔═╗ = Laid-out image placement                         │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Spread Preview (with Gutter)

```
┌─ Spread (16×10" @ 300 DPI, split with 0.25" gutter) ────┐
│                                                          │
│  ┌─ Left Page ──────┐ GUTTER ┌─ Right Page ──────┐     │
│  │                  │   ║     │                   │     │
│  │  ┌─────────────┐ │   ║     │ ┌─────────────┐  │     │
│  │  │             │ │   ║     │ │             │  │     │
│  │  │  Left half  │→│   ║     │←│ Right half  │  │     │
│  │  │  of wide    │ │   ║     │ │ of wide     │  │     │
│  │  │  image      │ │   ║     │ │ image       │  │     │
│  │  │             │ │   ║     │ │             │  │     │
│  │  └─────────────┘ │   ║     │ └─────────────┘  │     │
│  │        ↑         │   ║     │         ↑        │     │
│  │     overlap      │   ║     │      overlap     │     │
│  └──────────────────┘   ║     └──────────────────┘     │
│                                                          │
│  Wide laid-out image spans entire 16" width             │
│  Split in center with 0.25" gutter                      │
│  Each page overlaps 0.125" into gutter (for binding)    │
│                                                          │
│  Export options:                                         │
│  • Left page only  (2437 × 3000 px)                     │
│  • Right page only (2437 × 3000 px)                     │
│  • Both pages      (4800 × 3000 px unsplit)             │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## Workflow: Creating a Print Page

```
User Flow:
──────────

1. Navigate to "Page Layouts" tab

2. Create page template (if needed):
   ┌────────────────────────────────────────┐
   │ Click [+ Create Template]              │
   │                                        │
   │ → Set page size (8×10")                │
   │ → Set DPI (300)                        │
   │ → Set margins (0.5" all)               │
   │ → Choose positioning: Fill content     │
   │ → Preview with sample laid-out image   │
   │ → Save template                        │
   └────────────────────────────────────────┘

3. Create print page:
   ┌────────────────────────────────────────┐
   │ Select laid-out image from dropdown    │
   │ (from Image Layouts tab)               │
   │                                        │
   │ Select page template                   │
   │                                        │
   │ Click [Create Print Page]              │
   │                                        │
   │ → System applies template to image     │
   │ → Generates print-ready page           │
   │ → Shows in grid                        │
   └────────────────────────────────────────┘

4. Review print page:
   ┌────────────────────────────────────────┐
   │ Click [View] on print page card        │
   │                                        │
   │ → See full preview with margins        │
   │ → Check crop marks and bleed           │
   │ → Export as PNG or PDF                 │
   │ → Or add to Zine                       │
   └────────────────────────────────────────┘
```

---

## Data Model

### Page Layout Template

```typescript
{
  id: "ptpl-20251011-abc123",
  project_id: "prj-...",  // null for global
  name: "8×10 Portrait with margins",
  description: "Standard book page with 0.5\" margins",
  
  settings: {
    // Page dimensions
    page_width_in: 8.0,
    page_height_in: 10.0,
    dpi: 300,
    
    // Margins
    margin_top_in: 0.5,
    margin_right_in: 0.5,
    margin_bottom_in: 0.5,
    margin_left_in: 0.5,
    
    // Spread settings
    is_spread: false,
    gutter_width_in: 0.25,      // only for spreads
    gutter_overlap_in: 0.125,   // only for spreads
    
    // Image positioning
    positioning_mode: "fill",  // "fill" | "absolute" | "snap"
    
    // For absolute mode:
    image_x_in: 1.0,
    image_y_in: 1.5,
    image_width_in: 6.0,
    image_height_in: 7.0,
    
    // For snap mode:
    snap_anchor: "middle-center",  // TL, TC, TR, ML, MC, MR, BL, BC, BR
    
    // Export settings
    include_bleed: true,
    bleed_in: 0.125,
    include_crop_marks: true,
    crop_mark_offset_in: 0.25,
  }
}
```

### Print Page (Laid-Out Page)

```typescript
{
  id: "pp-20251011-def456",
  project_id: "prj-...",
  page_template_id: "ptpl-...",
  laid_out_image_id: "loi-...",
  
  result: {
    // Computed output dimensions
    output_width_px: 2400,
    output_height_px: 3000,
    
    // For spreads:
    left_page_width_px: 2437,
    right_page_width_px: 2437,
    combined_width_px: 4800,
    
    // File paths (when rendered)
    render_path: "./renders/pp-.../.png",
    left_page_path: "./renders/pp-...-left.png",   // spreads only
    right_page_path: "./renders/pp-...-right.png", // spreads only
  },
  
  created_at: "2025-10-11T...",
  updated_at: "2025-10-11T..."
}
```

---

## Use Cases

### Use Case 1: Simple Book Page

**Goal:** Place a portrait photo on an 8×10" page with margins

```
Steps:
1. Template: "8×10 Portrait with 0.5\" margins"
2. Laid-out image: "img01-portrait-crop" (already cropped 2:3)
3. Positioning: Fill content area
4. Result: Image fills 7×9" content area on 8×10" page
```

### Use Case 2: Full Bleed Page

**Goal:** Image extends to page edge (no margins)

```
Steps:
1. Template: "8×10 Full Bleed"
   - Margins: 0"
   - Bleed: 0.125"
   - Include crop marks
2. Laid-out image: "img02-landscape"
3. Positioning: Fill content area (entire page)
4. Result: Image fills 8×10" with 0.125" bleed extension
```

### Use Case 3: Spread Across Two Pages

**Goal:** Wide panorama photo spans two facing pages

```
Steps:
1. Template: "16×10 Spread with gutter"
   - Page size: 8×10" each
   - Spread mode: Yes
   - Gutter: 0.25"
   - Overlap: 0.125"
2. Laid-out image: "panorama-wide" (16:10 ratio)
3. Positioning: Fill content area
4. Result:
   - Left page: 2437 × 3000 px
   - Right page: 2437 × 3000 px
   - Both pages have overlap into gutter for binding
```

### Use Case 4: Inset Photo

**Goal:** Small photo in corner of page (like a polaroid)

```
Steps:
1. Template: "Letter with inset photo"
   - Page: 8.5×11"
   - Margins: 0.5"
   - Positioning: Absolute
   - X: 5.5", Y: 1.0"
   - W: 2.0", H: 2.5"
2. Laid-out image: "portrait-small"
3. Result: Small image in top-right area of page
```

---

## Key Differences from Image Layouts

| Image Layouts Tab | Page Layouts Tab |
|-------------------|------------------|
| Crop/scale source image | Place laid-out image on page |
| Output: Laid-out image | Output: Print-ready page |
| Focus: Image composition | Focus: Page composition |
| Settings: Crop, scale, position | Settings: Page size, margins, placement |
| Input: Asset (raw image) | Input: Laid-out image |
| Use case: Prepare images | Use case: Prepare for print |

---

## Implementation Notes

### Template Form Controls Needed

- Page size dropdown (same as Image Layouts)
- DPI slider
- Spread mode checkbox
- Gutter width/overlap sliders (if spread)
- Uniform margins toggle + sliders
- Positioning mode radio buttons (Fill / Absolute / Snap)
- Conditional fields based on positioning mode
- Bleed and crop marks settings

### Preview Rendering

For live preview in template editor:
1. Load a sample laid-out image (user selects from dropdown)
2. Apply page template settings
3. Render:
   - Page boundary
   - Margins (as overlay lines)
   - Laid-out image positioned per settings
   - Bleed area (if enabled)
   - Crop marks (if enabled)
   - Gutter visualization (if spread)

For spread previews:
- Show all three: combined, left page, right page
- Visualize gutter with overlay
- Show overlap areas

### Backend Integration

When user creates print page:
```
POST /api/projects/{id}/print-pages
{
  "page_template_id": "ptpl-...",
  "laid_out_image_id": "loi-..."
}

Response:
{
  "print_page": {
    "id": "pp-...",
    "result": {
      "output_width_px": 2400,
      "output_height_px": 3000,
      ...
    }
  }
}
```

Backend should:
1. Fetch page template settings
2. Fetch laid-out image (with its computation result)
3. Calculate final placement on page
4. Generate render (or store rendering instructions)
5. Return print page record

---

## Comparison with Current Design

**What I got wrong before:**
- ❌ Multiple images per page (that's for Phase 4+ if needed)
- ❌ Complex grid layouts (not the current requirement)
- ❌ Page slots and drag-and-drop multiple images

**What it should be:**
- ✅ ONE laid-out image per page
- ✅ Page dimensions and margins
- ✅ Spread support (wide image → split into left/right pages)
- ✅ Simple positioning: fill, absolute, or snap
- ✅ Print settings: bleed, crop marks

**This makes much more sense** because:
- Each laid-out image has already been cropped/scaled
- Page layout just adds the print wrapper (page size, margins, bleed)
- Spreads let one wide image span two pages
- Simpler workflow: Image Layout → Page Layout → Zine

---

## Revised Workflow

```
Complete Workflow:
─────────────────

1. Assets Tab
   Upload raw images

2. Sequences Tab  
   Organize images into order

3. Image Layouts Tab
   Create templates (crop, scale, position)
   Apply to assets → Laid-Out Images

4. Page Layouts Tab  ← THIS TAB
   Create page templates (size, margins, spread settings)
   Apply to laid-out images → Print Pages
   
5. Zine Tab
   Collect print pages into zine
   Apply imposition template
   Export for printing
```

**Each print page = One laid-out image + Page template**

---

**END OF DESIGN**

