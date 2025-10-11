# UI/UX Design Specification for Zine Layout Platform
**Version:** 1.0  
**Date:** October 11, 2025  
**Status:** Design Specification (No Implementation)

---

## Executive Summary

This document provides a comprehensive UI/UX design for the Zine Layout Platform, addressing the current implementation where all components are stacked vertically on a single long page. The new design introduces a **tabbed workflow structure** with clear visual hierarchy, progressive disclosure, and task-oriented interfaces.

### Key Problems with Current UI

1. **Everything on one page**: Assets, sequences, templates, laid-out images, and layout sequences all stacked vertically
2. **No clear workflow**: Users don't know the order of operations
3. **JSON everywhere**: Forms require raw JSON input instead of proper controls
4. **Poor visual hierarchy**: No distinction between primary and secondary actions
5. **No preview integration**: Template editing happens separately from asset preview
6. **Context switching**: Difficult to connect assets → templates → laid-out images → sequences

### Design Principles

1. **Progressive Disclosure**: Show only what's needed at each step
2. **Workflow-Oriented**: Guide users through: Upload → Organize → Template → Layout → Sequence → Export
3. **Visual First**: Preview prominently, settings secondary
4. **Form over JSON**: Proper UI controls instead of textarea JSON editing
5. **Contextual Actions**: Actions available where they make sense

---

## Overall Application Structure

### Global Navigation (Unchanged)

```
┌────────────────────────────────────────────────────────────────────────┐
│  🅩 Zine Layout        Home   Projects   Health                      👤 │
└────────────────────────────────────────────────────────────────────────┘
```

---

## Project Detail Page Architecture

### NEW: Tabbed Workflow Navigation

The project detail page now uses a **horizontal tab bar** to organize the workflow into discrete stages:

```
┌────────────────────────────────────────────────────────────────────────┐
│  ← Back to Projects                                                     │
│                                                                         │
│  Summer Photobook 2025                                                 │
│  Created Oct 10, 2025 • Updated Oct 11, 2025                          │
│  Vacation photos from Greece trip                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────┐ ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐           │
│  │ 📁  │ │  🔢     │ │  🖼️      │ │  📄      │ │  📚     │           │
│  │Assets│ │Sequences│ │  Image   │ │   Page   │ │  Zine   │           │
│  │     │ │         │ │ Layouts  │ │ Layouts  │ │         │           │
│  └─────┘ └─────────┘ └──────────┘ └──────────┘ └─────────┘           │
│    (1)       (2)         (3)          (4)          (5)                │
│                                                                         │
│  [Tab content area below]                                              │
└─────────────────────────────────────────────────────────────────────────┘
```

**Tab Workflow:**

1. **Assets** (📁): Upload and manage raw images
2. **Sequences** (🔢): Organize images into ordered collections
3. **Image Layouts** (🖼️): Create layout templates and apply them to assets
4. **Page Layouts** (📄): Multi-image page composition *(Phase 3 - Placeholder)*
5. **Zine** (📚): Zine assembly and export *(Phase 3 - Placeholder)*

---

## Tab 1: Assets Tab

### Purpose
Upload, browse, and manage raw image assets. Simple gallery view with upload.

### Layout

```
┌────────────────────────────────────────────────────────────────────────┐
│  Assets (23 images)                                    [Upload Images] │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─ Upload Area (if empty or drag-over) ─────────────────────────┐   │
│  │                                                                  │   │
│  │              📤  Drag images here or click to upload            │   │
│  │                                                                  │   │
│  │                Supported: PNG files, max 64MB each              │   │
│  │                                                                  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  Sort: [Most Recent ▾]    View: [Grid] [List]    Filter: [____] 🔍   │
│                                                                         │
│  ┌─ Asset Gallery ───────────────────────────────────────────────┐   │
│  │                                                                 │   │
│  │  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐     │   │
│  │  │████│  │████│  │████│  │████│  │████│  │████│  │████│     │   │
│  │  │████│  │████│  │████│  │████│  │████│  │████│  │████│     │   │
│  │  └────┘  └────┘  └────┘  └────┘  └────┘  └────┘  └────┘     │   │
│  │  img01  img02  img03  img04  img05  img06  img07          │   │
│  │  ☑       ☐      ☑       ☐      ☐      ☐      ☐           │   │
│  │                                                                 │   │
│  │  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐     │   │
│  │  │████│  │████│  │████│  │████│  │████│  │████│  │████│     │   │
│  │  │████│  │████│  │████│  │████│  │████│  │████│  │████│     │   │
│  │  └────┘  └────┘  └────┘  └────┘  └────┘  └────┘  └────┘     │   │
│  │  img08  img09  img10  img11  img12  img13  img14          │   │
│  │  ☐       ☐      ☐       ☐      ☐      ☐      ☐           │   │
│  │                                                                 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  [3 selected]  [Delete Selected]  [Add to Sequence ▸]                 │
│                                                                         │
│  ┌─ Selected Asset Details (when one asset clicked) ────────────┐    │
│  │                                                                 │    │
│  │  ┌─────────────┐  img-20251011-k2p7wq.png                     │    │
│  │  │             │  4032 × 3024 px (12.2 MP)                     │    │
│  │  │   Preview   │  2.4 MB • Uploaded Oct 11, 2025 3:15 PM     │    │
│  │  │             │                                                │    │
│  │  └─────────────┘  [View Full Size]  [Delete Asset]  [✕ Close] │    │
│  │                                                                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Features

- **Drag-and-drop upload** with visual feedback
- **Grid view** with thumbnails (default) or **List view** with details
- **Batch selection** with checkboxes
- **Quick actions**: Add to sequence, delete
- **Details panel** slides up when single asset clicked (non-modal)
- **Search/filter** by filename
- **Sort options**: Most recent, oldest, filename, size, dimensions

### Behavior

- Click asset thumbnail → shows details panel at bottom
- Check multiple assets → bulk action bar appears
- Drag asset → can drop in Sequences tab
- Empty state shows upload area prominently
- Loading state shows skeleton thumbnails

---

## Tab 2: Sequences Tab

### Purpose
Organize assets into named, ordered collections. Drag-and-drop interface for sequencing with live preview.

### Layout

```
┌────────────────────────────────────────────────────────────────────────┐
│  Image Sequences                                      [+ New Sequence] │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─ Sequence Selector ──────────────────────────────────────────────┐ │
│  │                                                                    │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  [+]      │ │
│  │  │● Best of     │  │  Contact     │  │  For Book    │           │ │
│  │  │  Summer      │  │  Sheet       │  │  Chapter 1   │           │ │
│  │  │  (20 items)  │  │  (50 items)  │  │  (15 items)  │           │ │
│  │  └──────────────┘  └──────────────┘  └──────────────┘           │ │
│  │  Selected                                                         │ │
│  │                                                                    │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Best of Summer                          [✎ Edit] [🗑️ Delete]   │  │
│  │  Top 20 photos for photobook                                     │  │
│  │  Updated Oct 11, 2025 4:30 PM                                    │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─── Two-Column Layout ──────────────────────────────────────────┐   │
│  │                          │                                       │   │
│  │  ┌─ Preview ─────────┐  │  ┌─ Sequence Builder ──────────────┐ │   │
│  │  │                   │  │  │                                   │ │   │
│  │  │                   │  │  │  [1] ┌────┐ img01.png    [↑][↓][✕]│ │   │
│  │  │    Main Image     │  │  │      │████│ 3024×4032            │ │   │
│  │  │    Display        │  │  │      └────┘                      │ │   │
│  │  │                   │  │  │                                   │ │   │
│  │  │                   │  │  │  [2] ┌────┐ img07.png    [↑][↓][✕]│ │   │
│  │  │                   │  │  │      │████│ 4032×3024            │ │   │
│  │  │                   │  │  │      └────┘                      │ │   │
│  │  └───────────────────┘  │  │                                   │ │   │
│  │                          │  │  [3] [ GAP ]          [↑][↓][✕]  │ │   │
│  │  Item 2 of 20           │  │                                   │ │   │
│  │  [◀ Prev] [▶ Play] [Next▶]│  │  [4] ┌────┐ img12.png   [↑][↓][✕]│ │   │
│  │                          │  │      │████│ 3024×4032            │ │   │
│  │  img07.png              │  │      └────┘                      │ │   │
│  │  4032 × 3024 px         │  │                                   │ │   │
│  │                          │  │  ⋮                                │ │   │
│  │  [Export Sequence]      │  │                                   │ │   │
│  │                          │  │  Drop assets here or add below   │ │   │
│  │                          │  │                                   │ │   │
│  │                          │  │  [Add Selected Asset] [Insert Gap]│ │   │
│  │                          │  └───────────────────────────────────┘ │   │
│  │                          │                                       │   │
│  └──────────────────────────┴───────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Features

- **Card-based sequence selector** at top
- **Split view**: Preview pane (left) + Sequence builder (right)
- **Drag-and-drop** from Assets tab or within sequence to reorder
- **Live preview** with slideshow controls (prev/play/next)
- **Position indicators** [1], [2], [3] for clear ordering
- **Inline controls** for each item: move up/down, remove
- **Gap insertion** for spreads or spacing
- **Quick edit** sequence name/description via edit button

### Behavior

- Click sequence card → loads that sequence
- Drag assets from Assets tab → automatically adds to end of sequence
- Drag within sequence → reorders items
- Click item in sequence → shows in preview pane
- Play button → auto-advances every 2.5 seconds
- Export sequence → downloads all images in order

---

## Tab 3: Image Layouts Tab

### Purpose
Create reusable layout templates and apply them to assets to generate laid-out images. This combines template management with the production workflow in a single, cohesive interface.

### Layout

This tab has two main sections: **Template Library** (top) and **Laid-Out Images** (bottom), providing a complete workflow from template creation to image production.

```
┌────────────────────────────────────────────────────────────────────────┐
│  Image Layouts                                                          │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ═══ SECTION 1: Layout Templates ═══                                   │
│                                                                         │
│                                                                         │
│  ┌─ Template Library ──────────────────────────────────────────────┐  │
│  │                                                                   │  │
│  │  [All] [Global] [Project]          Sort: [Name ▾]    🔍 Search   │  │
│  │                                                                   │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐ │  │
│  │  │ 8×10       │  │ Square     │  │ Instagram  │  │ Postcard   │ │  │
│  │  │ Portrait   │  │ 1:1        │  │ Story      │  │ 4×6        │ │  │
│  │  │            │  │            │  │            │  │            │ │  │
│  │  │ [   📄]    │  │ [   📄]    │  │ [   📄]    │  │ [   📄]    │ │  │
│  │  │            │  │            │  │            │  │            │ │  │
│  │  │ 8×10", 300│  │ 8×8", 300 │  │ 1080×1920  │  │ 4×6", 300 │ │  │
│  │  │ DPI        │  │ DPI        │  │ 72 DPI     │  │ DPI        │ │  │
│  │  │ Global     │  │ Project    │  │ Global     │  │ Global     │ │  │
│  │  │            │  │            │  │            │  │            │ │  │
│  │  │ [Edit]     │  │ [Edit]     │  │ [Edit]     │  │ [Edit]     │ │  │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘ │  │
│  │                                                                   │  │
│  │  ┌────────────┐  ┌────────────┐                                  │  │
│  │  │ Full Bleed │  │ Classic    │                                  │  │
│  │  │ Landscape  │  │ Portrait   │                                  │  │
│  │  │            │  │            │                                  │  │
│  │  │ [   📄]    │  │ [   📄]    │          [+ Create Template]    │  │
│  │  │            │  │            │                                  │  │
│  │  │ 11×8.5"    │  │ 8×10"      │                                  │  │
│  │  │ 300 DPI    │  │ 300 DPI    │                                  │  │
│  │  │ Project    │  │ Global     │                                  │  │
│  │  │            │  │            │                                  │  │
│  │  │ [Edit]     │  │ [Edit]     │                                  │  │
│  │  └────────────┘  └────────────┘                                  │  │
│  │                                                                   │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─ When "Edit" or "Create" clicked ──────────────────────────────┐   │
│  │                                                                  │   │
│  │  ┌── Template Editor Modal/Drawer ──────────────────────────┐  │   │
│  │  │                                                            │  │   │
│  │  │  Create Layout Template                          [✕ Close]│  │   │
│  │  │  ──────────────────────────────────────────────────────── │  │   │
│  │  │                                                            │  │   │
│  │  │  ┌─ Left: Form Controls ─┐  ┌─ Right: Live Preview ────┐ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ Template Name:         │  │  ┌─────────────────────┐ │ │  │   │
│  │  │  │ [________________]     │  │  │                     │ │ │  │   │
│  │  │  │                        │  │  │  ┌─────────────┐   │ │ │  │   │
│  │  │  │ Description:           │  │  │  │             │   │ │ │  │   │
│  │  │  │ [________________]     │  │  │  │   Sample    │   │ │ │  │   │
│  │  │  │                        │  │  │  │   Image     │   │ │ │  │   │
│  │  │  │ Scope:                 │  │  │  │   Preview   │   │ │ │  │   │
│  │  │  │ ◉ Project  ○ Global    │  │  │  │             │   │ │ │  │   │
│  │  │  │                        │  │  │  └─────────────┘   │ │ │  │   │
│  │  │  │ ═══ Page Setup ═══     │  │  │     ↑ margins      │ │ │  │   │
│  │  │  │                        │  │  └─────────────────────┘ │ │  │   │
│  │  │  │ Paper Size:            │  │                           │ │  │   │
│  │  │  │ [8×10 ▾] or Custom     │  │  Dimensions:             │ │  │   │
│  │  │  │   Width:  [8.0] in     │  │  Canvas: 2400×3000 px    │ │  │   │
│  │  │  │   Height: [10.0] in    │  │  Content: 2250×2850 px   │ │  │   │
│  │  │  │                        │  │  Source: 1500×2000 px    │ │  │   │
│  │  │  │ DPI: [300_______]      │  │  Scale: 1.2×             │ │  │   │
│  │  │  │      100    600         │  │                           │ │  │   │
│  │  │  │                        │  │  [← Select Preview Asset] │ │  │   │
│  │  │  │ Orientation:           │  │                           │ │  │   │
│  │  │  │ [Portrait ▾]           │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ ═══ Margins ═══        │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ [✓] Uniform margins    │  │                           │ │  │   │
│  │  │  │   All: [0.5] in        │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ [ ] Individual:        │  │                           │ │  │   │
│  │  │  │   Top:    [0.5] in     │  │                           │ │  │   │
│  │  │  │   Right:  [0.5] in     │  │                           │ │  │   │
│  │  │  │   Bottom: [0.5] in     │  │                           │ │  │   │
│  │  │  │   Left:   [0.5] in     │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ ═══ Crop Settings ═══  │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ Crop Mode:             │  │                           │ │  │   │
│  │  │  │ ◉ Fill (cover)         │  │                           │ │  │   │
│  │  │  │ ○ Fit (contain)        │  │                           │ │  │   │
│  │  │  │ ○ Custom ratio         │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ Aspect Ratio:          │  │                           │ │  │   │
│  │  │  │ [2:3 (Portrait) ▾]     │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ ═══ Position ═══       │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ Anchor:                │  │                           │ │  │   │
│  │  │  │ [TL][TC][TR]           │  │                           │ │  │   │
│  │  │  │ [ML][MC][MR]           │  │                           │ │  │   │
│  │  │  │ [BL][BC][BR]           │  │                           │ │  │   │
│  │  │  │     ^^ selected        │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ Fine-tune:             │  │                           │ │  │   │
│  │  │  │   X: [____0____] norm  │  │                           │ │  │   │
│  │  │  │       -1    +1         │  │                           │ │  │   │
│  │  │  │   Y: [____0____] norm  │  │                           │ │  │   │
│  │  │  │       -1    +1         │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ User Scale:            │  │                           │ │  │   │
│  │  │  │   [____1.0____] ×      │  │                           │ │  │   │
│  │  │  │   0.5      2.0         │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  │ [< Advanced Settings]  │  │                           │ │  │   │
│  │  │  │                        │  │                           │ │  │   │
│  │  │  └────────────────────────┘  └───────────────────────────┘ │  │   │
│  │  │                                                            │  │   │
│  │  │  [Cancel]  [Save as JSON]  [Save Template]               │  │   │
│  │  │                                                            │  │   │
│  │  └────────────────────────────────────────────────────────────┘  │   │
│  │                                                                  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Features

- **Card-based template library** with visual indicators
- **Filter by scope**: All, Global, Project
- **Visual form controls** instead of JSON:
  - Dropdowns for presets (paper sizes, aspect ratios)
  - Sliders for numeric values (DPI, scale, position)
  - Radio buttons for modes (fill/fit/custom)
  - 9-point grid for anchor selection
  - Toggle for uniform margins
- **Live preview pane** showing sample asset with template applied
- **Real-time updates** as user adjusts sliders/controls
- **Preset paper sizes**: Letter, A4, 8×10, 4×6, etc.
- **Preset aspect ratios**: 1:1, 2:3, 3:2, 4:5, 16:9, etc.
- **Advanced settings** collapsible section for focus points, export options
- **JSON export** option for power users

### Behavior

- Click "Create Template" → opens modal/drawer editor
- Click "Edit" on template card → loads template into editor
- Select preview asset → dropdown of project assets
- Adjust any control → preview updates immediately
- Save template → adds to library, closes editor
- Delete template → confirms, removes from library (if not in use)

### Advanced Settings Section (Collapsed by Default)

```
│  │  │  [v Advanced Settings]  │  │                           │ │  │   │
│  │  │  │                      │  │                           │ │  │   │
│  │  │  │ ═══ Focus Point ═══  │  │                           │ │  │   │
│  │  │  │                      │  │                           │ │  │   │
│  │  │  │ [ ] Enable focus pt  │  │                           │ │  │   │
│  │  │  │   Source X: [___] px │  │                           │ │  │   │
│  │  │  │   Source Y: [___] px │  │                           │ │  │   │
│  │  │  │   Target X: [___]    │  │                           │ │  │   │
│  │  │  │   Target Y: [___]    │  │                           │ │  │   │
│  │  │  │                      │  │                           │ │  │   │
│  │  │  │ ═══ Export ═══       │  │                           │ │  │   │
│  │  │  │                      │  │                           │ │  │   │
│  │  │  │ Format: [PNG ▾]      │  │                           │ │  │   │
│  │  │  │ Quality: [90___] %   │  │                           │ │  │   │
│  │  │  │ Background: [⬜]      │  │                           │ │  │   │
│  │  │  │                      │  │                           │ │  │   │
```

│                                                                         │
│  ═══ SECTION 2: Laid-Out Images ═══                                    │
│                                                                         │
```

### Section 1: Template Library

**Purpose:** Manage reusable layout templates with visual form controls.

**Features:**
- Card-based template library with visual indicators
- Filter by scope: All, Global, Project
- Visual form controls instead of JSON
- Live preview pane showing sample asset with template applied
- Preset paper sizes and aspect ratios
- Advanced settings collapsible section

See Template Editor Modal design in the full layout below.

### Section 2: Laid-Out Images

**Purpose:** Apply templates to assets to create laid-out images. Shows asset-template pairings with visual previews.

```
┌────────────────────────────────────────────────────────────────────────┐
│  Laid-Out Images (35)                        [+ Create Laid-Out Image] │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─ Quick Actions ─────────────────────────────────────────────────┐  │
│  │                                                                   │  │
│  │  Batch Apply Template:                                           │  │
│  │    Source: [Best of Summer (sequence) ▾]                         │  │
│  │    Template: [8×10 Portrait ▾]                                   │  │
│  │    [Apply to All 20 Images]                                      │  │
│  │                                                                   │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  Filter: [All] [By Asset] [By Template]     Sort: [Recent ▾]  🔍      │
│                                                                         │
│  ┌─ Laid-Out Images Grid ────────────────────────────────────────┐    │
│  │                                                                 │    │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────┐ │    │
│  │  │┌──────────┐│  │┌──────────┐│  │┌──────────┐│  │┌────────┐│ │    │
│  │  ││          ││  ││          ││  ││          ││  ││        ││ │    │
│  │  ││  Preview ││  ││  Preview ││  ││  Preview ││  ││ Preview││ │    │
│  │  ││  Image   ││  ││  Image   ││  ││  Image   ││  ││ Image  ││ │    │
│  │  ││          ││  ││          ││  ││          ││  ││        ││ │    │
│  │  │└──────────┘│  │└──────────┘│  │└──────────┘│  │└────────┘│ │    │
│  │  │            │  │            │  │            │  │          │ │    │
│  │  │ img01.png  │  │ img02.png  │  │ img03.png  │  │img04.png │ │    │
│  │  │ 8×10 Port  │  │ Square 1:1 │  │ 8×10 Port  │  │Instagram │ │    │
│  │  │ Oct 11 3:15│  │ Oct 11 3:16│  │ Oct 11 3:17│  │Oct 11 3:18│ │    │
│  │  │            │  │            │  │            │  │          │ │    │
│  │  │ [Edit] [✕] │  │ [Edit] [✕] │  │ [Edit] [✕] │  │[Edit][✕]│ │    │
│  │  └────────────┘  └────────────┘  └────────────┘  └──────────┘ │    │
│  │                                                                 │    │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐               │    │
│  │  │┌──────────┐│  │┌──────────┐│  │┌──────────┐│               │    │
│  │  ││          ││  ││          ││  ││          ││               │    │
│  │  ││  Preview ││  ││  Preview ││  ││  Preview ││               │    │
│  │  ││  Image   ││  ││  Image   ││  ││  Image   ││               │    │
│  │  ││          ││  ││          ││  ││          ││               │    │
│  │  │└──────────┘│  │└──────────┘│  │└──────────┘│               │    │
│  │  │            │  │            │  │            │               │    │
│  │  │ img05.png  │  │ img06.png  │  │ img07.png  │               │    │
│  │  │ Postcard   │  │ 8×10 Port  │  │ Full Bleed │               │    │
│  │  │ Oct 11 3:19│  │ Oct 11 3:20│  │ Oct 11 3:21│               │    │
│  │  │            │  │            │  │            │               │    │
│  │  │ [Edit] [✕] │  │ [Edit] [✕] │  │ [Edit] [✕] │               │    │
│  │  └────────────┘  └────────────┘  └────────────┘               │    │
│  │                                                                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─ When "Edit" clicked on a laid-out image ─────────────────────┐    │
│  │                                                                 │    │
│  │  ┌── Edit Laid-Out Image Drawer ──────────────────────────┐   │    │
│  │  │                                                          │   │    │
│  │  │  Edit Layout: img01.png                        [✕ Close]│   │    │
│  │  │  ───────────────────────────────────────────────────────│   │    │
│  │  │                                                          │   │    │
│  │  │  ┌─ Left: Preview ─────┐  ┌─ Right: Adjustments ─────┐ │   │    │
│  │  │  │                      │  │                           │ │   │    │
│  │  │  │  ┌────────────────┐ │  │ Source Asset:             │ │   │    │
│  │  │  │  │                │ │  │ img01.png (4032×3024)     │ │   │    │
│  │  │  │  │                │ │  │                           │ │   │    │
│  │  │  │  │  Final Layout  │ │  │ Base Template:            │ │   │    │
│  │  │  │  │  Preview       │ │  │ [8×10 Portrait ▾]         │ │   │    │
│  │  │  │  │                │ │  │ [Change Template]         │ │   │    │
│  │  │  │  │                │ │  │                           │ │   │    │
│  │  │  │  │  (Shows crop   │ │  │ ═══ Overrides ═══        │ │   │    │
│  │  │  │  │   window and   │ │  │                           │ │   │    │
│  │  │  │  │   positioning) │ │  │ User Scale:               │ │   │    │
│  │  │  │  │                │ │  │ [____1.2____] ×           │ │   │    │
│  │  │  │  │                │ │  │ 0.5      2.0              │ │   │    │
│  │  │  │  └────────────────┘ │  │                           │ │   │    │
│  │  │  │                      │  │ Position Offset:          │ │   │    │
│  │  │  │  Canvas: 2400×3000  │  │   X: [____0____] norm     │ │   │    │
│  │  │  │  Content: 2250×2850 │  │       -1    +1            │ │   │    │
│  │  │  │  Crop: 1800×2400    │  │   Y: [____0____] norm     │ │   │    │
│  │  │  │  Scale: 1.2×        │  │       -1    +1            │ │   │    │
│  │  │  │                      │  │                           │ │   │    │
│  │  │  │  [Download Preview]  │  │ [Reset Overrides]         │ │   │    │
│  │  │  │                      │  │                           │ │   │    │
│  │  │  │                      │  │ ═══ Computation ═══       │ │   │    │
│  │  │  │                      │  │                           │ │   │    │
│  │  │  │                      │  │ Source Rect:              │ │   │    │
│  │  │  │                      │  │   x:516 y:0 w:3000 h:3024│ │   │    │
│  │  │  │                      │  │                           │ │   │    │
│  │  │  │                      │  │ Target Rect:              │ │   │    │
│  │  │  │                      │  │   x:0 y:0 w:2250 h:2850  │ │   │    │
│  │  │  │                      │  │                           │ │   │    │
│  │  │  │                      │  │ [< View Full Trace]       │ │   │    │
│  │  │  │                      │  │                           │ │   │    │
│  │  │  └──────────────────────┘  └───────────────────────────┘ │   │    │
│  │  │                                                          │   │    │
│  │  │  [Recompute]  [Save Changes]  [Delete]                 │   │    │
│  │  │                                                          │   │    │
│  │  └──────────────────────────────────────────────────────────┘   │    │
│  │                                                                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Features

- **Batch apply** template to entire sequence at top
- **Visual grid** of laid-out images with thumbnail previews
- **Filter and sort** options
- **Quick actions** on each card: Edit, Delete
- **Edit drawer** with split view:
  - Left: Large preview showing final layout
  - Right: Template selection + override controls
- **Override controls** are subset of template settings (scale, position)
- **Computation display** shows technical details (rects, scale) in collapsed section
- **Real-time preview** updates as overrides change
- **Change template** button to switch base template
- **Reset overrides** returns to base template settings

### Behavior

- Click "Create" → modal with asset selector + template selector
- Click "Edit" on card → opens drawer editor
- Adjust sliders in editor → preview updates immediately
- "Recompute" → forces fresh calculation (after template change)
- "Save Changes" → persists overrides, updates preview thumbnail
- Batch apply → creates multiple laid-out images in background, shows progress

### Create Modal (When [+ Create Laid-Out Image] clicked)

```
┌─ Create Laid-Out Image ──────────────────────────────┐
│                                                       │
│  Select Source Asset:                                │
│  ┌──────────────────────────────────────────────┐   │
│  │ ◉ From Assets                                 │   │
│  │   Asset: [img01.png ▾]                       │   │
│  │                                               │   │
│  │ ○ From Sequence                               │   │
│  │   Sequence: [Best of Summer ▾]               │   │
│  │   [Apply template to entire sequence]        │   │
│  └──────────────────────────────────────────────┘   │
│                                                       │
│  Select Template:                                    │
│  [8×10 Portrait ▾]                                   │
│                                                       │
│  Optional Overrides (JSON):                          │
│  [ ] Use custom overrides                            │
│  ┌─────────────────────────────────────────────┐    │
│  │ { "user_scale": 1.2 }                        │    │
│  └─────────────────────────────────────────────┘    │
│                                                       │
│  [Cancel]  [Create]  [Create & Edit]                │
└───────────────────────────────────────────────────────┘
```

---

## Tab 4: Page Layouts Tab *(Phase 3 - Placeholder)*

### Purpose
Compose multi-image pages using page templates. Select a page template (e.g., 2-up, 4-up grid) and assign laid-out images to slots.

### Placeholder Design

```
┌────────────────────────────────────────────────────────────────────────┐
│  Page Layouts                                        [+ Create Page]    │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─ Page Templates ──────────────────────────────────────────────────┐ │
│  │                                                                     │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │ │
│  │  │ Single   │  │ 2-Up     │  │ 4-Up     │  │ Custom   │          │ │
│  │  │ Image    │  │ Spread   │  │ Grid     │  │ Grid     │          │ │
│  │  │ [  📄  ] │  │ [ 📄📄 ] │  │ [📄📄] │  │ [ ⊞ ]    │          │ │
│  │  │          │  │          │  │ [📄📄] │  │          │          │ │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘          │ │
│  │                                                                     │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─ Composed Pages ──────────────────────────────────────────────────┐  │
│  │                                                                     │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐                         │  │
│  │  │┌────────┐│  │┌────────┐│  │┌────────┐│                         │  │
│  │  ││Page    ││  ││Page    ││  ││Page    ││                         │  │
│  │  ││Preview ││  ││Preview ││  ││Preview ││                         │  │
│  │  │└────────┘│  │└────────┘│  │└────────┘│                         │  │
│  │  │ Page 1   │  │ Page 2   │  │ Page 3   │                         │  │
│  │  │ 2-Up     │  │ Single   │  │ 4-Up     │                         │  │
│  │  │ [Edit]   │  │ [Edit]   │  │ [Edit]   │                         │  │
│  │  └──────────┘  └──────────┘  └──────────┘                         │  │
│  │                                                                     │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  Coming in Phase 3                                                     │
│  • Select page template (1-up, 2-up, 4-up, custom grids)             │
│  • Drag laid-out images into page slots                               │
│  • Preview complete page composition                                   │
│  • Save as laid-out page                                               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Features (Planned)

- **Page template selector** with common presets
- **Drag-and-drop** laid-out images into page slots
- **Visual page composer** showing final page layout
- **Grid customization** for custom slot arrangements
- **Preview rendering** of complete pages

---

## Tab 5: Zine Tab *(Phase 3 - Placeholder)*

### Purpose
Assemble pages into complete zines, manage page order, and export for print. Final production step.

### Layout

```
┌────────────────────────────────────────────────────────────────────────┐
│  Zines                                              [+ Create Zine]     │
├────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─ Zine Selector ────────────────────────────────────────────────────┐ │
│  │                                                                     │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │ │
│  │  │● Summer      │  │  Draft       │  │  Alternate   │            │ │
│  │  │  Photobook   │  │  v2          │  │  Layout      │            │ │
│  │  │  (24 pages)  │  │  (18 pages)  │  │  (20 pages)  │            │ │
│  │  └──────────────┘  └──────────────┘  └──────────────┘            │ │
│  │  Selected                                                          │ │
│  │                                                                     │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─ Zine Pages ───────────────────────────────────────────────────────┐ │
│  │                                                                     │ │
│  │  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐                 │ │
│  │  │ P1 │  │ P2 │  │ P3 │  │ P4 │  │ P5 │  │ P6 │  ...            │ │
│  │  │ ▭  │  │ ▭  │  │ ▭▭ │  │ ▭▭ │  │ ▭  │  │ ▭  │                 │ │
│  │  └────┘  └────┘  └────┘  └────┘  └────┘  └────┘                 │ │
│  │  Cover   Title   Spread  Spread  Photo  Photo                     │ │
│  │                                                                     │ │
│  │  Drag to reorder • Click to preview                               │ │
│  │                                                                     │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─ Imposition & Export ─────────────────────────────────────────────┐  │
│  │                                                                     │  │
│  │  Imposition Template: [8-Page Fold ▾]                             │  │
│  │                                                                     │  │
│  │  ┌─────────────────────────────────────────────────────────────┐  │  │
│  │  │  [ Preview of folded layout ]                                │  │  │
│  │  │                                                               │  │  │
│  │  │    8    1                  Shows how pages will be           │  │  │
│  │  │            →  fold  →      arranged when printed             │  │  │
│  │  │    2    7                  and folded                         │  │  │
│  │  │                                                               │  │  │
│  │  └─────────────────────────────────────────────────────────────┘  │  │
│  │                                                                     │  │
│  │  Export Format: ◉ PDF  ○ PNG Sequence  ○ Print-Ready ZIP         │  │
│  │                                                                     │  │
│  │  [✓] Include crop marks    [✓] Include bleed                      │  │
│  │  [✓] Embed color profile   [ ] Flatten layers                     │  │
│  │                                                                     │  │
│  │  [Preview Export]  [Download Zine]                                │  │
│  │                                                                     │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  Coming in Phase 4                                                     │
│  • Create zine from laid-out pages                                    │
│  • Reorder pages for final book sequence                              │
│  • Apply imposition templates (8-page fold, 16-page booklet)         │
│  • Export print-ready PDF or image sequence                           │
│  • Preview folded/bound result                                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Features (Planned)

- **Zine selector** for managing multiple books
- **Page sequence editor** with drag-and-drop reordering
- **Imposition template selector** (8-page fold, 16-page booklet, etc.)
- **Visual imposition preview** showing how pages map to print sheets
- **Export options**:
  - PDF with crop marks and bleed
  - PNG sequence for online viewing
  - Print-ready ZIP with multiple sheets
- **Page thumbnail preview** in sequence order
- **Print preview mode** showing spreads as they'll appear when bound

---

## Responsive Design Considerations

### Desktop (>1200px)
- Full three-column layouts in Sequences and Output tabs
- Side-by-side template editor (form + preview)
- Large preview panes

### Tablet (768px - 1200px)
- Two-column layouts (preview below controls)
- Template editor stacked vertically
- Smaller thumbnails in grids

### Mobile (<768px)
- Single-column layouts
- Preview full-width
- Controls in collapsible sections
- Template editor full-screen modal
- Simplified navigation (tabs become dropdown selector)

---

## Component States & Feedback

### Loading States

```
┌─ Loading Skeleton Example ──────────────────┐
│                                              │
│  ┌────┐  ┌────┐  ┌────┐  ┌────┐            │
│  │░░░░│  │░░░░│  │░░░░│  │░░░░│            │
│  │░░░░│  │░░░░│  │░░░░│  │░░░░│            │
│  └────┘  └────┘  └────┘  └────┘            │
│  ░░░░░   ░░░░░   ░░░░░   ░░░░░             │
│                                              │
└──────────────────────────────────────────────┘
```

### Empty States

Each tab has a helpful empty state:

```
┌─ Empty State Example (Assets Tab) ──────────┐
│                                              │
│              📤                              │
│                                              │
│      No images yet. Get started!            │
│                                              │
│      Drag images here or click below        │
│      to upload your first photos.           │
│                                              │
│      [Upload Images]                         │
│                                              │
└──────────────────────────────────────────────┘
```

### Error States

```
┌─ Error Message Example ──────────────────────┐
│                                              │
│  ⚠️  Failed to create template               │
│                                              │
│  Invalid paper dimensions: width must be    │
│  greater than 0.                             │
│                                              │
│  [Dismiss]  [Try Again]                      │
│                                              │
└──────────────────────────────────────────────┘
```

### Success Feedback

```
┌─ Success Toast (Top-Right) ──────────┐
│                                       │
│  ✓  Template "8×10 Portrait" saved!  │
│                                       │
└───────────────────────────────────────┘
  (auto-dismisses after 3 seconds)
```

### Progress Indicators

```
┌─ Batch Operation Progress ───────────────────┐
│                                              │
│  Applying template to 20 images...          │
│                                              │
│  ████████████░░░░░░░░  12 of 20 complete    │
│                                              │
│  Current: img-12.png                         │
│                                              │
│  [Cancel]                                    │
│                                              │
└──────────────────────────────────────────────┘
```

---

## Color & Visual Design Language

### Color Palette (Conceptual)

- **Primary**: Blue (#3B82F6) - Actions, selected states, links
- **Success**: Green (#10B981) - Confirmations, checkmarks
- **Warning**: Yellow (#F59E0B) - Cautions, alerts
- **Danger**: Red (#EF4444) - Delete, errors, destructive actions
- **Neutral Grays**: 
  - 50: Background (#F9FAFB)
  - 200: Borders (#E5E7EB)
  - 500: Secondary text (#6B7280)
  - 900: Primary text (#111827)

### Typography Hierarchy

```
H1 (Page Title):        32px, Bold, Gray-900
H2 (Section Title):     24px, Semibold, Gray-900
H3 (Subsection):        18px, Semibold, Gray-800
Body:                   16px, Regular, Gray-700
Small/Meta:             14px, Regular, Gray-500
Tiny/Label:             12px, Medium, Gray-600
Code/Technical:         14px, Mono, Gray-800
```

### Spacing Scale

```
xs:   4px
sm:   8px
md:   16px
lg:   24px
xl:   32px
2xl:  48px
```

### Card/Panel Elevation

```
Level 0 (Flat):     border-gray-200, bg-white
Level 1 (Card):     shadow-sm, border-gray-200
Level 2 (Raised):   shadow-md, border-gray-300
Level 3 (Modal):    shadow-lg, border-gray-300
```

---

## Interaction Patterns

### Drag-and-Drop

**Visual Feedback:**
```
┌─ Dragging Asset ────────────────────────────┐
│                                              │
│      ┌────┐  ← Drag source becomes          │
│      │░░░░│     semi-transparent (50%)       │
│      └────┘                                  │
│                                              │
│      ┌────────────────────┐                 │
│      │  Drop Zone         │  ← Highlight    │
│      │  (blue border)     │     valid drops  │
│      └────────────────────┘                 │
│                                              │
│      ┌────┐  ← Drag preview follows cursor  │
│      │████│                                  │
│      └────┘                                  │
│      img01.png                               │
│                                              │
└──────────────────────────────────────────────┘
```

### Keyboard Navigation

- **Tab**: Navigate between focusable elements
- **Enter/Space**: Activate buttons, toggle selections
- **Arrow keys**: Navigate within lists/grids
- **Esc**: Close modals, cancel operations
- **Cmd/Ctrl + S**: Quick save (in editors)
- **Cmd/Ctrl + Z**: Undo (in editors)

### Context Menus (Right-Click)

```
┌─ Asset Context Menu ─────┐
│                           │
│  View Full Size           │
│  Add to Sequence  ▸       │
│  Apply Template   ▸       │
│  ───────────────          │
│  Download Original        │
│  Copy Asset ID            │
│  ───────────────          │
│  Delete                   │
│                           │
└───────────────────────────┘
```

---

## Mobile-Specific Considerations

### Mobile Tab Navigation

On mobile, the tab bar becomes a dropdown selector:

```
┌────────────────────────────────────┐
│  ← Summer Photobook 2025           │
│                                    │
│  Current: [📁 Assets ▾]           │
│                                    │
│  ┌──────────────────────────────┐ │
│  │ 📁 Assets                    │ │
│  │ 🔢 Sequences                 │ │
│  │ 📐 Templates                 │ │
│  │ 🖼️  Layouts                   │ │
│  │ 📚 Output                    │ │
│  └──────────────────────────────┘ │
│                                    │
└────────────────────────────────────┘
```

### Mobile Cards

Cards stack vertically, actions move to bottom:

```
┌────────────────────────────────────┐
│  ┌──────────────────────────────┐ │
│  │                              │ │
│  │       Image Preview          │ │
│  │                              │ │
│  └──────────────────────────────┘ │
│                                    │
│  img01.png                         │
│  4032 × 3024 px                   │
│  Uploaded Oct 11, 2025            │
│                                    │
│  [View] [Add to Sequence] [Delete]│
│                                    │
└────────────────────────────────────┘
```

### Mobile Gestures

- **Swipe left/right**: Navigate between items in preview
- **Pinch to zoom**: In preview panes
- **Long press**: Show context menu
- **Pull to refresh**: Reload current tab data

---

## Accessibility Features

### Screen Reader Support

- All interactive elements have proper `aria-label` attributes
- Image thumbnails have descriptive alt text
- Form inputs have associated labels
- Status messages use `aria-live` regions
- Modals trap focus and announce on open

### Keyboard-Only Navigation

- All functionality accessible via keyboard
- Visible focus indicators (blue outline)
- Skip links for quick navigation
- Logical tab order throughout application

### Color Contrast

- All text meets WCAG AA standards (4.5:1 for normal text)
- Interactive elements have sufficient contrast
- Error states don't rely solely on color (include icons)

### Visual Aids

- Icon + text for important actions
- Loading spinners with text labels
- Progress indicators show percentage
- Tooltips on hover for additional context

---

## User Workflows (Step-by-Step)

### Workflow 1: Create Photo Book from Scratch

1. **Upload Images** (Assets Tab)
   - User clicks "Upload Images" or drags files
   - System shows upload progress
   - Thumbnails appear in gallery

2. **Organize Sequence** (Sequences Tab)
   - User clicks "+ New Sequence"
   - Names it "Summer Book Final"
   - Drags images from Assets tab (or uses batch select)
   - Reorders items via drag-and-drop
   - Inserts gaps for spreads

3. **Create Template & Apply** (Image Layouts Tab)
   - User scrolls to Template Library section
   - Clicks "+ Create Template"
   - Names it "8×10 Portrait Crop"
   - Adjusts settings with visual controls:
     - Paper: 8×10"
     - DPI: 300
     - Margins: 0.5" all sides
     - Crop: Fill, 2:3 ratio
   - Previews on sample image
   - Saves template
   - Scrolls to Laid-Out Images section
   - Selects "Summer Book Final" sequence
   - Selects "8×10 Portrait Crop" template
   - Clicks "Apply to All 20 Images"
   - System processes batch (shows progress)
   - Grid populates with laid-out images

4. **Fine-Tune Layouts** (Image Layouts Tab)
   - User clicks "Edit" on specific laid-out image
   - Adjusts scale slider to 1.2×
   - Nudges position slightly
   - Sees live preview update
   - Saves changes

5. **Compose Pages** (Page Layouts Tab - *Phase 3*)
   - User selects page template (e.g., 2-up spread)
   - Drags laid-out images into page slots
   - Previews complete page
   - Saves as laid-out page
   - Repeats for all pages

6. **Assemble Zine** (Zine Tab - *Phase 3/4*)
   - User creates new zine
   - Adds laid-out pages in order
   - Applies imposition template (8-page fold)
   - Previews folding/binding
   - Exports print-ready PDF

**Total time:** ~15 minutes for 20 images (after initial setup)

### Workflow 2: Quick Template Test

1. **Upload Test Image** (Assets Tab)
2. **Create & Apply Template** (Image Layouts Tab) - adjust settings, preview, apply to single asset
3. **Review Preview** - check computed layout
4. **Download** - export single image *(when export is implemented)*

**Total time:** ~2 minutes

### Workflow 3: Apply Existing Template to New Photos

1. **Upload Images** (Assets Tab)
2. **Create Sequence** (Sequences Tab) - optional, can apply directly
3. **Batch Apply** (Image Layouts Tab) - select existing template, apply to all
4. **Review & Export** - preview results, export as needed

**Total time:** ~5 minutes for 10 images (template already exists)

---

## Implementation Priorities

### Phase 1: Core Tabbed Structure ✅ **COMPLETED**
- ✅ Implement tab navigation component
- ✅ Create Assets tab with upload and gallery
- ✅ Create Sequences tab with split-view (preview + builder)
- ✅ Migrate existing API hooks to new tab components
- ✅ URL-based tab state with search params

### Phase 2: Image Layouts Tab (Week 2)
- Build unified Image Layouts tab with two sections:
  - **Section 1:** Template Library with visual form editor
  - **Section 2:** Laid-Out Images with batch apply
- Replace JSON textareas with proper form controls:
  - Dropdowns for presets (paper sizes, aspect ratios)
  - Sliders for numeric values (DPI, scale, position)
  - 9-point anchor grid selector
  - Toggle for uniform margins
- Implement live preview integration in template editor
- Build edit drawer for laid-out image overrides
- Add batch apply progress indicators

### Phase 3: Page Layouts & Zine Tabs (Week 3-4)
- Build Page Layouts tab (Phase 3 backend work)
  - Page template selector
  - Visual page composer with drag-and-drop slots
  - Preview rendering of composed pages
- Build Zine tab (Phase 3/4 backend work)
  - Zine page sequence editor
  - Imposition template selector
  - Export options form
  - Print preview mode

### Phase 4: Polish & Mobile (Week 5)
- Responsive design implementation
- Mobile tab dropdown selector
- Cross-tab drag-and-drop
- Keyboard shortcuts
- Accessibility audit and fixes
- Performance optimization (virtual scrolling)

---

## Technical Notes for Implementation

### State Management

- **Tab state**: Use React Router query params (`?tab=assets`)
- **Selected items**: Component-local state (don't pollute global store)
- **API data**: RTK Query cache (already implemented)
- **Form state**: Controlled components with local state
- **Drag state**: Context provider for drag-and-drop operations

### Component Architecture

```
<ProjectDetail>
  <ProjectHeader />
  
  <Tabs value={activeTab} onValueChange={setActiveTab}>
    <TabsList>
      <TabsTrigger value="assets">📁 Assets</TabsTrigger>
      <TabsTrigger value="sequences">🔢 Sequences</TabsTrigger>
      <TabsTrigger value="image-layouts">🖼️ Image Layouts</TabsTrigger>
      <TabsTrigger value="page-layouts">📄 Page Layouts</TabsTrigger>
      <TabsTrigger value="zine">📚 Zine</TabsTrigger>
    </TabsList>
    
    <TabsContent value="assets">
      <AssetsTab projectId={id} />
    </TabsContent>
    
    <TabsContent value="sequences">
      <SequencesTab projectId={id} />
    </TabsContent>
    
    <TabsContent value="image-layouts">
      <ImageLayoutsTab projectId={id} />
    </TabsContent>
    
    <TabsContent value="page-layouts">
      <PageLayoutsTab projectId={id} /> {/* Phase 3 */}
    </TabsContent>
    
    <TabsContent value="zine">
      <ZineTab projectId={id} /> {/* Phase 3/4 */}
    </TabsContent>
  </Tabs>
</ProjectDetail>
```

### Key Reusable Components

- `<TabBar>` - Horizontal tab navigation
- `<ImageGrid>` - Responsive thumbnail grid
- `<PreviewPanel>` - Large image preview with controls
- `<SequenceBuilder>` - Drag-and-drop list builder
- `<SliderInput>` - Visual slider with numeric input
- `<AspectRatioSelector>` - Radio buttons for common ratios
- `<AnchorGrid>` - 9-point alignment selector
- `<ProgressBar>` - Batch operation progress
- `<EmptyState>` - Consistent empty state messaging
- `<ConfirmDialog>` - Reusable confirmation modal

### Performance Optimizations

- **Virtualized lists** for sequences with >50 items
- **Lazy load** thumbnails (IntersectionObserver)
- **Debounce** slider inputs before API calls
- **Memoize** expensive computations (sorted/filtered lists)
- **Optimistic updates** for reordering operations
- **Background workers** for batch template application

### Data Fetching Strategy

- **Prefetch** next tab data when user hovers over tab
- **Cache invalidation** only when mutations occur
- **Polling** for batch operation status (every 2 seconds)
- **Stale-while-revalidate** for non-critical data

---

## Future Enhancements (Post-Launch)

### Advanced Features

1. **Undo/Redo** - Global action history
2. **Keyboard shortcuts** - Power user commands
3. **Bulk edit** - Change template for multiple laid-out images
4. **Smart cropping** - AI-based subject detection
5. **Template marketplace** - Share/download community templates
6. **Collaboration** - Share projects with team members
7. **Version history** - Revert to previous sequence states
8. **Custom grid layouts** - Visual page composition tool (Phase 3)
9. **Print preview** - See book spreads before export (Phase 4)
10. **Direct printing** - Integration with print services

### Analytics & Insights

- **Usage stats** - Most popular templates, average session time
- **Template performance** - Which templates are used most
- **Export tracking** - What formats/DPI are most common

### Automation

- **Auto-sequences** - Generate sequences based on criteria (date, size)
- **Batch rename** - Rename assets with patterns
- **Template suggestions** - Recommend templates based on image dimensions
- **Smart export** - Automatically choose best settings based on image analysis

---

## Conclusion

This design specification transforms the current "everything stacked vertically" UI into a **workflow-oriented, tabbed interface** that guides users through the zine layout process step-by-step. The design prioritizes:

1. **Visual clarity** - Clear hierarchy and purpose for each screen
2. **Progressive disclosure** - Show only what's needed at each step
3. **Form over JSON** - Proper UI controls instead of raw JSON editing
4. **Live previews** - See results immediately as settings change
5. **Contextual actions** - Actions available where they make sense

The tab structure mirrors the actual workflow: **Upload (Assets) → Organize (Sequences) → Create Layouts (Image Layouts) → Compose Pages (Page Layouts) → Assemble & Export (Zine)**, making it intuitive for new users while remaining efficient for power users.

**Key Improvement Areas:**

- **From chaos to order**: Clear tabs instead of one long scroll
- **From JSON to forms**: Visual controls instead of text areas
- **From guessing to guided**: Workflow steps are obvious
- **From disconnected to integrated**: Preview + settings side-by-side
- **From basic to professional**: Modern, polished interface

Implementation should follow the phased approach outlined above, starting with core tab structure and progressively adding functionality. The visual design language and component patterns ensure consistency across all tabs.

---

**END OF SPECIFICATION**

