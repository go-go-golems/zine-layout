# Phase 3 & 4 Implementation Guide
**Complete Step-by-Step Breakdown**  
**Date:** October 11, 2025  
**Updated:** October 11, 2025 (API Discovery)  
**Status:** REST APIs Already Done, Rendering Service Needed

---

## 🎉 IMPORTANT UPDATE

**Major Discovery:** Phase 3 REST APIs (tasks 3C, 3D, 3F) are **already fully implemented**!

**What This Means:**
- ✅ All page template endpoints work
- ✅ All laid-out page endpoints work (create with single image!)
- ✅ All zine endpoints work
- ✅ Frontend hooks are ready to use
- ⏰ **28 hours of work saved!**
- 📅 **Timeline reduced from 5 weeks to 3-4 weeks**

**What You Can Do RIGHT NOW:**
- Create page templates via API ✅
- Create print pages (single image per page) ✅
- Create and manage zines ✅
- All CRUD operations work ✅

**What Still Needs Work:**
- Page rendering service (preview/export return 501)
- Connect frontend tabs to APIs
- Imposition algorithms
- PDF generation

**See `15-phase3-api-status-and-discrepancies.md` for complete analysis.**

---

## Table of Contents

1. [Overview & Prerequisites](#overview--prerequisites)
2. [Phase 3A: Page Layout Settings & Types](#phase-3a-page-layout-settings--types)
3. [Phase 3B: Page Rendering Service](#phase-3b-page-rendering-service)
4. [Phase 3C: Page Templates REST API](#phase-3c-page-templates-rest-api)
5. [Phase 3D: Laid-Out Pages REST API](#phase-3d-laid-out-pages-rest-api)
6. [Phase 3E: PageLayoutsTab Frontend](#phase-3e-pagelayoutstab-frontend)
7. [Phase 3F: Zine REST API](#phase-3f-zine-rest-api)
8. [Phase 3G: Testing & Integration](#phase-3g-testing--integration)
9. [Phase 4A: Imposition Algorithms](#phase-4a-imposition-algorithms)
10. [Phase 4B: PDF Generation](#phase-4b-pdf-generation)
11. [Phase 4C: Export Service](#phase-4c-export-service)
12. [Phase 4D: ZineTab Frontend](#phase-4d-zinetab-frontend)
13. [Phase 4E: Final Testing & Launch](#phase-4e-final-testing--launch)

---

## Overview & Prerequisites

### What's Already Done

✅ **Phase 1 & 2 Complete:**
- Database schema for all entities
- Repositories for CRUD operations
- Image layout engine with computation
- REST API for projects, assets, sequences, templates, laid-out images
- CLI commands for all Phase 1 & 2 entities
- Complete tabbed UI with visual controls

✅ **Phase 3 MOSTLY DONE:** (Discovered October 11, 2025)
- ✅ Database tables: `page_templates`, `laid_out_pages` (corrected), `zines`, `zine_pages`
- ✅ Repository layer: All CRUD operations (updated for single image per page)
- ✅ Service layer: `PagesService` (rendering stubbed), `ZinesService`
- ✅ **REST API: ALL ENDPOINTS IMPLEMENTED!** (page templates, laid-out pages, zines)
- ✅ **Frontend API types and RTK Query hooks: COMPLETE!**
- ✅ CLI workflow commands (direct DB access)
- ✅ Dummy frontend tabs showing target UX

**🎉 SURPRISE DISCOVERY:** 
REST APIs for Phase 3 (page templates, laid-out pages, zines) are already fully implemented! They correctly use the single-image-per-page model. Preview and export endpoints exist but return 501 (renderer not implemented yet). See `15-phase3-api-status-and-discrepancies.md` for detailed analysis.

### What Needs Implementation

⧗ **Phase 3 Remaining (MUCH LESS THAN EXPECTED):**
- `PageLayoutSettings` struct definition (Go)
- Page rendering service (single pages + spreads with gutter)
- ~~REST API endpoints~~ ✅ DONE! (Already implemented)
- Connect PageLayoutsTab to real API (hooks ready, just remove dummy data)
- Connect ZineTab to real API (hooks ready, just remove dummy data)
- Implement preview/export rendering

○ **Phase 4 Complete:**
- Imposition algorithms (8-page fold, 16-page booklet, etc.)
- PDF generation with proper imposition
- Crop marks and bleed rendering
- Zine export service (create export endpoint)

**Revised Estimate:** ~92 hours total (down from 120 hours - REST APIs already done!)

### Tools & Dependencies

**Already Available:**
- Go 1.21+
- SQLite with existing schema
- Vite + React + RTK Query
- Tailwind CSS
- Canvas API (browser)

**Will Need:**
- Go PDF library (recommend: `github.com/jung-kurt/gofpdf` or `github.com/pdfcpu/pdfcpu`)
- Image manipulation: `github.com/disintegration/imaging` (already may be in use)
- Canvas rendering for server-side (or client-side only initially)

---

## ⚡ Quick Start: Test Existing APIs

**Good news!** You can test Phase 3 APIs right now (they're already implemented):

```bash
# Start server
cd zine-layout
./zine-layout serve --addr :8088 --data-root ./data

# In another terminal, test the APIs:

# 1. Create page template
curl -X POST http://localhost:8088/api/page-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "8x10 Test",
    "template": {
      "page_width_in": 8,
      "page_height_in": 10,
      "dpi": 300,
      "margin_top_in": 0.5,
      "margin_right_in": 0.5,
      "margin_bottom_in": 0.5,
      "margin_left_in": 0.5,
      "is_spread": false,
      "positioning_mode": "fill"
    }
  }' | jq .

# 2. List templates
curl http://localhost:8088/api/page-templates | jq .

# 3. Create zine (with page IDs from your database)
curl -X POST http://localhost:8088/api/projects/YOUR_PROJECT_ID/zines \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Zine",
    "laid_out_page_ids": []
  }' | jq .

# All of these work TODAY!
```

**See `15-phase3-api-status-and-discrepancies.md` for complete API documentation.**

---

## Phase 3A: Page Layout Settings & Types

### Step 3A.1: Define PageLayoutSettings Struct

**File:** `pkg/pagelayout/types.go` (NEW)

**Create new package for page layout logic:**

```go
package pagelayout

// PageLayoutSettings defines how a laid-out image is positioned on a physical print page.
type PageLayoutSettings struct {
	// Page dimensions
	PageWidthIn  float64 `json:"page_width_in"`
	PageHeightIn float64 `json:"page_height_in"`
	DPI          float64 `json:"dpi"`

	// Margins
	MarginTopIn    float64 `json:"margin_top_in"`
	MarginRightIn  float64 `json:"margin_right_in"`
	MarginBottomIn float64 `json:"margin_bottom_in"`
	MarginLeftIn   float64 `json:"margin_left_in"`

	// Spread mode
	IsSpread         bool    `json:"is_spread"`
	GutterWidthIn    float64 `json:"gutter_width_in,omitempty"`     // Only for spreads
	GutterOverlapIn  float64 `json:"gutter_overlap_in,omitempty"`   // How much each page extends into gutter

	// Image positioning
	PositioningMode string  `json:"positioning_mode"` // "fill" | "absolute" | "snap"
	
	// For absolute mode
	ImageXIn      float64 `json:"image_x_in,omitempty"`
	ImageYIn      float64 `json:"image_y_in,omitempty"`
	ImageWidthIn  float64 `json:"image_width_in,omitempty"`
	ImageHeightIn float64 `json:"image_height_in,omitempty"`
	
	// For snap mode
	SnapAnchor string `json:"snap_anchor,omitempty"` // "top-left", "middle-center", etc.

	// Export settings
	IncludeBleed     bool    `json:"include_bleed"`
	BleedIn          float64 `json:"bleed_in"`
	IncludeCropMarks bool    `json:"include_crop_marks"`
	CropMarkOffsetIn float64 `json:"crop_mark_offset_in"`
}

// PageLayoutResult contains the computed placement for rendering.
type PageLayoutResult struct {
	// Output dimensions
	OutputWidthPx  int `json:"output_width_px"`
	OutputHeightPx int `json:"output_height_px"`

	// For spreads
	IsSpread            bool `json:"is_spread"`
	LeftPageWidthPx     int  `json:"left_page_width_px,omitempty"`
	RightPageWidthPx    int  `json:"right_page_width_px,omitempty"`
	CombinedWidthPx     int  `json:"combined_width_px,omitempty"`
	GutterPositionPx    int  `json:"gutter_position_px,omitempty"`
	GutterWidthPx       int  `json:"gutter_width_px,omitempty"`      // For visualization
	GutterOverlapPx     int  `json:"gutter_overlap_px,omitempty"`    // For visualization

	// Image placement on page (or spread)
	ImageXPx      int `json:"image_x_px"`
	ImageYPx      int `json:"image_y_px"`
	ImageWidthPx  int `json:"image_width_px"`
	ImageHeightPx int `json:"image_height_px"`

	// File paths (when rendered)
	RenderPath      string `json:"render_path,omitempty"`
	LeftPagePath    string `json:"left_page_path,omitempty"`      // Spreads: left page PNG
	RightPagePath   string `json:"right_page_path,omitempty"`     // Spreads: right page PNG
	CombinedPath    string `json:"combined_path,omitempty"`       // Spreads: both pages with gap
	FullSpreadPath  string `json:"full_spread_path,omitempty"`    // Spreads: uncut wide image
}

// RenderOptions controls what gets rendered and how.
type RenderOptions struct {
	// What to render (for spreads)
	RenderLeft     bool `json:"render_left"`      // Render left page
	RenderRight    bool `json:"render_right"`     // Render right page
	RenderCombined bool `json:"render_combined"`  // Render both pages with binding gap
	RenderFull     bool `json:"render_full"`      // Render uncut spread
	
	// Visualization options
	ShowGutter      bool `json:"show_gutter"`       // Overlay gutter area in red (debug)
	ShowOverlap     bool `json:"show_overlap"`      // Show overlap zones with dashed lines
	ShowCropMarks   bool `json:"show_crop_marks"`   // Add crop marks
	ShowBleed       bool `json:"show_bleed"`        // Extend with bleed area
	
	// Output format
	Format string `json:"format"` // "png" | "pdf"
}

// PageLayoutComputation wraps settings and result.
type PageLayoutComputation struct {
	Settings PageLayoutSettings `json:"settings"`
	Result   PageLayoutResult   `json:"result"`
}
```

**Testing:**
```bash
go build ./pkg/pagelayout
```

---

## Phase 3B: Page Rendering Service

### Step 3B.1: Implement Page Layout Computation

**File:** `pkg/pagelayout/engine.go` (NEW)

**Purpose:** Calculate where laid-out image sits on physical page

```go
package pagelayout

import "fmt"

// ComputePageLayout calculates how a laid-out image fits on a physical page.
func ComputePageLayout(settings PageLayoutSettings, laidOutImageWidthPx, laidOutImageHeightPx int) (*PageLayoutResult, error) {
	if settings.PageWidthIn <= 0 || settings.PageHeightIn <= 0 {
		return nil, fmt.Errorf("page dimensions must be positive")
	}
	if settings.DPI <= 0 {
		return nil, fmt.Errorf("DPI must be positive")
	}

	result := &PageLayoutResult{
		IsSpread: settings.IsSpread,
	}

	// Calculate page dimensions in pixels
	pageWidthPx := int(settings.PageWidthIn * settings.DPI)
	pageHeightPx := int(settings.PageHeightIn * settings.DPI)

	if !settings.IsSpread {
		// Single page mode
		result.OutputWidthPx = pageWidthPx
		result.OutputHeightPx = pageHeightPx

		// Calculate content area (page minus margins)
		contentXPx := int(settings.MarginLeftIn * settings.DPI)
		contentYPx := int(settings.MarginTopIn * settings.DPI)
		contentWidthPx := pageWidthPx - int((settings.MarginLeftIn+settings.MarginRightIn)*settings.DPI)
		contentHeightPx := pageHeightPx - int((settings.MarginTopIn+settings.MarginBottomIn)*settings.DPI)

		switch settings.PositioningMode {
		case "fill":
			// Image fills content area
			result.ImageXPx = contentXPx
			result.ImageYPx = contentYPx
			result.ImageWidthPx = contentWidthPx
			result.ImageHeightPx = contentHeightPx

		case "absolute":
			// User-specified position and size
			result.ImageXPx = int(settings.ImageXIn * settings.DPI)
			result.ImageYPx = int(settings.ImageYIn * settings.DPI)
			result.ImageWidthPx = int(settings.ImageWidthIn * settings.DPI)
			result.ImageHeightPx = int(settings.ImageHeightIn * settings.DPI)

		case "snap":
			// Snap to anchor within content area
			result.ImageXPx = contentXPx
			result.ImageYPx = contentYPx
			result.ImageWidthPx = contentWidthPx
			result.ImageHeightPx = contentHeightPx
			// TODO: Apply snap_anchor offset

		default:
			return nil, fmt.Errorf("unknown positioning mode: %s", settings.PositioningMode)
		}

	} else {
		// Spread mode: wide image split into left/right pages
		totalSpreadWidth := pageWidthPx * 2
		result.OutputWidthPx = totalSpreadWidth
		result.OutputHeightPx = pageHeightPx
		result.CombinedWidthPx = totalSpreadWidth

		gutterPx := int(settings.GutterWidthIn * settings.DPI)
		overlapPx := int(settings.GutterOverlapIn * settings.DPI)

		// Gutter at center
		gutterCenterPx := totalSpreadWidth / 2
		result.GutterPositionPx = gutterCenterPx

		// Each page extends into gutter by overlap amount
		result.LeftPageWidthPx = pageWidthPx + overlapPx
		result.RightPageWidthPx = pageWidthPx + overlapPx

		// Image fills entire spread content area
		contentXPx := int(settings.MarginLeftIn * settings.DPI)
		contentYPx := int(settings.MarginTopIn * settings.DPI)
		contentWidthPx := totalSpreadWidth - int((settings.MarginLeftIn+settings.MarginRightIn)*settings.DPI)
		contentHeightPx := pageHeightPx - int((settings.MarginTopIn+settings.MarginBottomIn)*settings.DPI)

		result.ImageXPx = contentXPx
		result.ImageYPx = contentYPx
		result.ImageWidthPx = contentWidthPx
		result.ImageHeightPx = contentHeightPx
	}

	return result, nil
}
```

**Testing:**
```bash
go test ./pkg/pagelayout
```

### Step 3B.2: Implement Page Renderer

**File:** `pkg/pagelayout/renderer.go` (NEW)

**Purpose:** Render laid-out image onto physical page

```go
package pagelayout

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

// RenderPageToFile renders a laid-out image onto a physical page and saves to disk.
func RenderPageToFile(
	settings PageLayoutSettings,
	laidOutImagePath string,
	outputPath string,
) error {
	// Load the laid-out image
	srcImg, err := imaging.Open(laidOutImagePath)
	if err != nil {
		return fmt.Errorf("open laid-out image: %w", err)
	}

	// Compute page layout
	result, err := ComputePageLayout(settings, srcImg.Bounds().Dx(), srcImg.Bounds().Dy())
	if err != nil {
		return fmt.Errorf("compute page layout: %w", err)
	}

	if !settings.IsSpread {
		// Render single page
		return renderSinglePage(settings, result, srcImg, outputPath)
	}

	// Render spread (left, right, combined)
	dir := filepath.Dir(outputPath)
	base := filepath.Base(outputPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	leftPath := filepath.Join(dir, name+"-left"+ext)
	rightPath := filepath.Join(dir, name+"-right"+ext)
	combinedPath := filepath.Join(dir, name+"-combined"+ext)

	if err := renderSpreadPages(settings, result, srcImg, leftPath, rightPath, combinedPath); err != nil {
		return err
	}

	result.RenderPath = outputPath
	result.LeftPagePath = leftPath
	result.RightPagePath = rightPath

	return nil
}

func renderSinglePage(settings PageLayoutSettings, result *PageLayoutResult, srcImg image.Image, outputPath string) error {
	// Create output canvas
	canvas := image.NewRGBA(image.Rect(0, 0, result.OutputWidthPx, result.OutputHeightPx))
	
	// Fill background (white for print)
	draw.Draw(canvas, canvas.Bounds(), image.White, image.Point{}, draw.Src)

	// Resize laid-out image to target size on page
	resized := imaging.Resize(srcImg, result.ImageWidthPx, result.ImageHeightPx, imaging.Lanczos)

	// Draw image at calculated position
	draw.Draw(canvas,
		image.Rect(result.ImageXPx, result.ImageYPx,
			result.ImageXPx+result.ImageWidthPx,
			result.ImageYPx+result.ImageHeightPx),
		resized,
		image.Point{},
		draw.Over,
	)

	// TODO: Add bleed area if settings.IncludeBleed
	// TODO: Add crop marks if settings.IncludeCropMarks

	// Save to file
	return savePNG(canvas, outputPath)
}

func renderSpreadPages(settings PageLayoutSettings, result *PageLayoutResult, srcImg image.Image, leftPath, rightPath, combinedPath string, opts RenderOptions) error {
	// 1. Create full spread canvas
	spreadCanvas := image.NewRGBA(image.Rect(0, 0, result.CombinedWidthPx, result.OutputHeightPx))
	draw.Draw(spreadCanvas, spreadCanvas.Bounds(), image.White, image.Point{}, draw.Src)

	// 2. Resize laid-out image to spread dimensions
	resized := imaging.Resize(srcImg, result.ImageWidthPx, result.ImageHeightPx, imaging.Lanczos)

	// 3. Draw image on spread at calculated position
	draw.Draw(spreadCanvas,
		image.Rect(result.ImageXPx, result.ImageYPx,
			result.ImageXPx+result.ImageWidthPx,
			result.ImageYPx+result.ImageHeightPx),
		resized,
		image.Point{},
		draw.Over,
	)

	// 4. Split at gutter position
	gutterX := result.GutterPositionPx
	overlapPx := result.GutterOverlapPx

	// 5. Extract left page (with overlap into gutter)
	if opts.RenderLeft {
		leftCanvas := image.NewRGBA(image.Rect(0, 0, result.LeftPageWidthPx, result.OutputHeightPx))
		draw.Draw(leftCanvas, leftCanvas.Bounds(), image.White, image.Point{}, draw.Src)
		
		// Copy from spread: 0 to (gutterX + overlapPx)
		sourceRect := image.Rect(0, 0, gutterX+overlapPx, result.OutputHeightPx)
		draw.Draw(leftCanvas, leftCanvas.Bounds(), spreadCanvas, sourceRect.Min, draw.Src)

		// Optional: Show overlap zone with dashed line
		if opts.ShowOverlap {
			drawDashedLine(leftCanvas, gutterX-gutterX+overlapPx, 0, gutterX-gutterX+overlapPx, result.OutputHeightPx, color.RGBA{255, 0, 0, 180})
		}

		// Optional: Draw border
		drawBorder(leftCanvas, result.LeftPageWidthPx, result.OutputHeightPx)
		
		if err := savePNG(leftCanvas, leftPath); err != nil {
			return fmt.Errorf("save left page: %w", err)
		}
	}

	// 6. Extract right page (with overlap into gutter)
	if opts.RenderRight {
		rightCanvas := image.NewRGBA(image.Rect(0, 0, result.RightPageWidthPx, result.OutputHeightPx))
		draw.Draw(rightCanvas, rightCanvas.Bounds(), image.White, image.Point{}, draw.Src)
		
		// Copy from spread: (gutterX - overlapPx) to end
		sourceRect := image.Rect(gutterX-overlapPx, 0, result.CombinedWidthPx, result.OutputHeightPx)
		draw.Draw(rightCanvas, rightCanvas.Bounds(), spreadCanvas, sourceRect.Min, draw.Src)

		// Optional: Show overlap zone with dashed line
		if opts.ShowOverlap {
			drawDashedLine(rightCanvas, overlapPx, 0, overlapPx, result.OutputHeightPx, color.RGBA{255, 0, 0, 180})
		}

		// Optional: Draw border
		drawBorder(rightCanvas, result.RightPageWidthPx, result.OutputHeightPx)
		
		if err := savePNG(rightCanvas, rightPath); err != nil {
			return fmt.Errorf("save right page: %w", err)
		}
	}

	// 7. Render combined view (with binding gap for visualization)
	if opts.RenderCombined {
		gapPx := 4 // Small gap to show binding
		combinedCanvas := image.NewRGBA(image.Rect(0, 0, result.LeftPageWidthPx+result.RightPageWidthPx+gapPx, result.OutputHeightPx))
		
		// Background (gap shows through as dark binding area)
		draw.Draw(combinedCanvas, combinedCanvas.Bounds(), &image.Uniform{color.RGBA{51, 51, 51, 255}}, image.Point{}, draw.Src)

		// Draw left page
		leftImg, _ := imaging.Open(leftPath)
		draw.Draw(combinedCanvas, image.Rect(0, 0, result.LeftPageWidthPx, result.OutputHeightPx), leftImg, image.Point{}, draw.Src)

		// Draw right page (offset by left width + gap)
		rightImg, _ := imaging.Open(rightPath)
		draw.Draw(combinedCanvas, image.Rect(result.LeftPageWidthPx+gapPx, 0, result.LeftPageWidthPx+gapPx+result.RightPageWidthPx, result.OutputHeightPx), rightImg, image.Point{}, draw.Src)

		if err := savePNG(combinedCanvas, combinedPath); err != nil {
			return fmt.Errorf("save combined: %w", err)
		}
	}

	// 8. Optional: Render full uncut spread with gutter visualization
	if opts.RenderFull || opts.ShowGutter {
		fullCanvas := imaging.Clone(spreadCanvas)
		
		if opts.ShowGutter {
			// Draw gutter overlay (semi-transparent red)
			gutterStart := gutterX - result.GutterWidthPx/2
			gutterEnd := gutterX + result.GutterWidthPx/2
			drawFilledRect(fullCanvas, gutterStart, 0, gutterEnd, result.OutputHeightPx, color.RGBA{255, 0, 0, 38}) // 15% opacity
			
			// Draw center line (dashed)
			drawDashedLine(fullCanvas, gutterX, 0, gutterX, result.OutputHeightPx, color.RGBA{255, 0, 0, 204})
			
			// Add "GUTTER" label
			drawText(fullCanvas, "GUTTER", gutterX-30, 30, color.RGBA{255, 0, 0, 204})
		}
		
		fullPath := strings.TrimSuffix(combinedPath, filepath.Ext(combinedPath)) + "-full.png"
		if err := savePNG(fullCanvas, fullPath); err != nil {
			return fmt.Errorf("save full spread: %w", err)
		}
		result.FullSpreadPath = fullPath
	}

	return nil
}

// Helper functions for visualization
func drawDashedLine(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	// TODO: Implement dashed line drawing
	// For now, draw solid line
	// Production: use image/draw or external library for dashed lines
}

func drawBorder(img *image.RGBA, width, height int) {
	// Draw 2px border around image
	borderColor := color.RGBA{51, 51, 51, 255}
	// Top and bottom
	for x := 0; x < width; x++ {
		img.Set(x, 0, borderColor)
		img.Set(x, 1, borderColor)
		img.Set(x, height-2, borderColor)
		img.Set(x, height-1, borderColor)
	}
	// Left and right
	for y := 0; y < height; y++ {
		img.Set(0, y, borderColor)
		img.Set(1, y, borderColor)
		img.Set(width-2, y, borderColor)
		img.Set(width-1, y, borderColor)
	}
}

func drawFilledRect(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	rect := image.Rect(x1, y1, x2, y2)
	draw.Draw(img, rect, &image.Uniform{col}, image.Point{}, draw.Over)
}

func drawText(img *image.RGBA, text string, x, y int, col color.Color) {
	// TODO: Implement text drawing (requires font library)
	// For now, skip or use basic pixel text
}

func savePNG(img image.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()
	return png.Encode(f, img)
}
```

### Step 3B.3: Update PagesService to Use Renderer

**File:** `pkg/services/pages.go`

**Add rendering capability with options:**

```go
// RenderPage renders a laid-out page to disk using the page template settings.
// For spreads, can render multiple variants based on options.
func (s *PagesService) RenderPage(pageID, outputDir string, opts *pagelayout.RenderOptions) (*repo.LaidOutPage, error) {
	if opts == nil {
		// Default: render everything
		opts = &pagelayout.RenderOptions{
			RenderLeft:     true,
			RenderRight:    true,
			RenderCombined: true,
			RenderFull:     false,
			ShowGutter:     false,
			ShowOverlap:    false,
			ShowCropMarks:  false,
			ShowBleed:      false,
			Format:         "png",
		}
	}
	page, err := s.repos.LaidOutPages.Get(pageID)
	if err != nil {
		return nil, fmt.Errorf("fetch laid-out page: %w", err)
	}

	// Fetch page template
	template, err := s.repos.PageTemplates.Get(page.PageTemplateID)
	if err != nil {
		return nil, fmt.Errorf("fetch page template: %w", err)
	}

	// Parse template settings
	var settings pagelayout.PageLayoutSettings
	if err := json.Unmarshal([]byte(template.TemplateJSON), &settings); err != nil {
		return nil, fmt.Errorf("parse template settings: %w", err)
	}

	// Fetch laid-out image
	laidOutImage, err := s.repos.LaidOutImages.Get(page.LaidOutImageID)
	if err != nil {
		return nil, fmt.Errorf("fetch laid-out image: %w", err)
	}

	// Get the rendered image file path for the laid-out image
	// TODO: This assumes laid-out images have been rendered to disk
	// For now, use the original asset as placeholder
	asset, err := s.repos.Assets.Get(laidOutImage.AssetID)
	if err != nil {
		return nil, fmt.Errorf("fetch asset: %w", err)
	}

	laidOutImagePath := filepath.Join(outputDir, "temp", asset.Filename) // TODO: Use actual laid-out image render

	// Render the page
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.png", page.ID))
	if err := pagelayout.RenderPageToFile(settings, laidOutImagePath, outputPath); err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}

	// Update result JSON
	result := pagelayout.PageLayoutComputation{
		Settings: settings,
		Result:   *computedResult, // From renderer
	}
	resultJSON, _ := json.Marshal(result)
	resultStr := string(resultJSON)
	page.ResultJSON = &resultStr
	page.UpdatedAt = time.Now().UTC()

	if err := s.repos.LaidOutPages.Update(page); err != nil {
		return nil, fmt.Errorf("update page with result: %w", err)
	}

	return page, nil
}
```

**Testing:**
```bash
go build ./pkg/services
```

---

## Phase 3C: Page Templates REST API ✅ COMPLETE

**Status:** ✅ Already Implemented (238 lines)

**File:** `pkg/serve/page_templates_routes.go` (EXISTS)

```go
package serve

import (
	"encoding/json"
	"net/http"

	"github.com/go-go-golems/zine-layout/pkg/pagelayout"
	"github.com/go-go-golems/zine-layout/pkg/repo"
)

// GET /api/page-templates (global templates)
func (s *Server) handleListGlobalPageTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := s.repos.PageTemplates.ListGlobal()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"templates": adaptPageTemplates(templates)})
}

// GET /api/projects/{id}/page-templates (global + project)
func (s *Server) handleListProjectPageTemplates(w http.ResponseWriter, r *http.Request, projectID string) {
	all, err := s.repos.PageTemplates.ListByProject(&projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"templates": adaptPageTemplates(all)})
}

// POST /api/projects/{id}/page-templates
func (s *Server) handleCreatePageTemplate(w http.ResponseWriter, r *http.Request, projectID string) {
	var req struct {
		Name        string                           `json:"name"`
		Description string                           `json:"description"`
		Settings    pagelayout.PageLayoutSettings    `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	settingsJSON, err := json.Marshal(req.Settings)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid settings")
		return
	}

	template := &repo.PageTemplate{
		ProjectID:    &projectID,
		Name:         req.Name,
		Description:  req.Description,
		TemplateJSON: string(settingsJSON),
	}

	if err := s.repos.PageTemplates.Create(template); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"template": adaptPageTemplate(template, req.Settings)})
}

// Similar handlers for: GET /api/page-templates/{id}, PUT, DELETE
```

### Step 3C.2: Register Routes

**File:** `pkg/serve/server.go`

**Add to `Routes()` method:**

```go
// Page Templates
mux.HandleFunc("/api/page-templates", s.handlePageTemplatesGlobal)
mux.HandleFunc("/api/page-templates/", s.handlePageTemplate)
mux.HandleFunc("/api/projects/{id}/page-templates", s.handleProjectPageTemplates)
```

**Testing:**
```bash
# Start server
./zine-layout serve --addr :8088

# Test endpoints
curl http://localhost:8088/api/page-templates
curl -X POST http://localhost:8088/api/page-templates \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","settings":{...}}'
```

---

## Phase 3D: Laid-Out Pages REST API ✅ MOSTLY COMPLETE

**Status:** ✅ CRUD endpoints implemented, ⧗ Preview/Export stubbed (208 lines)

**File:** `pkg/serve/laid_out_pages_routes.go` (EXISTS)

```go
// GET /api/projects/{id}/laid-out-pages
func (s *Server) handleListLaidOutPages(w http.ResponseWriter, r *http.Request, projectID string) {
	pages, err := s.repos.LaidOutPages.ListByProject(projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"pages": pages})
}

// POST /api/projects/{id}/laid-out-pages
func (s *Server) handleCreateLaidOutPage(w http.ResponseWriter, r *http.Request, projectID string) {
	var req struct {
		PageTemplateID string `json:"page_template_id"`
		LaidOutImageID string `json:"laid_out_image_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	page, err := s.pagesService.CreatePage(projectID, req.PageTemplateID, req.LaidOutImageID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"page": page})
}

// GET /api/laid-out-pages/{id}/preview
func (s *Server) handlePreviewLaidOutPage(w http.ResponseWriter, r *http.Request, pageID string) {
	page, err := s.pagesService.GetPage(pageID)
	if err != nil {
		respondError(w, http.StatusNotFound, "page not found")
		return
	}

	// Return preview payload (for client-side rendering initially)
	// Later: server-side render and return image
	respondJSON(w, http.StatusOK, map[string]any{
		"page": page,
		// Include template settings, laid-out image URL, etc.
	})
}
```

**Testing:**
```bash
curl http://localhost:8088/api/projects/prj-.../laid-out-pages
curl -X POST http://localhost:8088/api/projects/prj-.../laid-out-pages \
  -d '{"page_template_id":"ptpl-...","laid_out_image_id":"loi-..."}'
```

---

## Phase 3E: PageLayoutsTab Frontend

### Step 3E.1: Update API Types

**File:** `web/src/api.ts`

**Add types:**

```typescript
export interface PageLayoutSettings {
  page_width_in: number;
  page_height_in: number;
  dpi: number;
  margin_top_in: number;
  margin_right_in: number;
  margin_bottom_in: number;
  margin_left_in: number;
  is_spread: boolean;
  gutter_width_in?: number;
  gutter_overlap_in?: number;
  positioning_mode: 'fill' | 'absolute' | 'snap';
  image_x_in?: number;
  image_y_in?: number;
  image_width_in?: number;
  image_height_in?: number;
  snap_anchor?: string;
  include_bleed: boolean;
  bleed_in: number;
  include_crop_marks: boolean;
  crop_mark_offset_in: number;
}

export interface PageTemplate {
  id: string;
  project_id?: string;
  scope: 'global' | 'project';
  name: string;
  description?: string;
  settings: PageLayoutSettings;
  created_at: string;
  updated_at: string;
}

export interface LaidOutPage {
  id: string;
  project_id: string;
  page_template_id: string;
  laid_out_image_id: string;
  result?: PageLayoutResult;
  created_at: string;
  updated_at: string;
}

// Add RTK Query endpoints...
```

### Step 3E.2: Rewrite PageLayoutsTab

**File:** `web/src/views/tabs/PageLayoutsTab.tsx`

**Follow design from `12-page-layout-tab-design.md`:**

**Key sections:**
1. Template library (similar to ImageLayoutsTab)
2. Template editor with spread mode toggle
3. Gutter settings (when spread enabled)
4. Positioning mode selector
5. Print pages grid
6. Quick create form

**Pattern to follow:** Same as ImageLayoutsTab but with page-specific settings

---

## Phase 3F: Zine REST API ✅ COMPLETE

**Status:** ✅ Already Implemented (209 lines)

**File:** `pkg/serve/zines_routes.go` (EXISTS)

```go
// GET /api/projects/{id}/zines
func (s *Server) handleListZines(w http.ResponseWriter, r *http.Request, projectID string) {
	zines, err := s.repos.Zines.ListByProject(projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"zines": zines})
}

// POST /api/projects/{id}/zines
func (s *Server) handleCreateZine(w http.ResponseWriter, r *http.Request, projectID string) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		PageIDs     []string `json:"page_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	zine, err := s.zinesService.CreateZine(projectID, req.Name, req.PageIDs)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"zine": zine})
}

// GET /api/zines/{id}/pages
func (s *Server) handleGetZinePages(w http.ResponseWriter, r *http.Request, zineID string) {
	pages, err := s.repos.Zines.GetPages(zineID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"pages": pages})
}

// PUT /api/zines/{id}/pages (reorder)
```

---

## Phase 3G: Testing & Integration

### Test Checklist

```bash
# 1. Create page template via API
curl -X POST http://localhost:8088/api/page-templates \
  -d '{"name":"8x10 Portrait","settings":{...}}'

# 2. Create print page
curl -X POST http://localhost:8088/api/projects/prj-.../laid-out-pages \
  -d '{"page_template_id":"ptpl-...","laid_out_image_id":"loi-..."}'

# 3. List pages
curl http://localhost:8088/api/projects/prj-.../laid-out-pages

# 4. Create zine
curl -X POST http://localhost:8088/api/projects/prj-.../zines \
  -d '{"name":"My Book","page_ids":["lpg-1","lpg-2"]}'

# 5. Verify in UI
# - Navigate to Page Layouts tab
# - Create template with visual controls
# - Create print page
# - Navigate to Zine tab
# - Add pages to zine
```

---

## Phase 4A: Imposition Algorithms

### Step 4A.1: Define Imposition Types

**File:** `pkg/imposition/types.go` (NEW)

```go
package imposition

// ImpositionTemplate defines how pages are arranged for printing and folding.
type ImpositionTemplate struct {
	ID          string
	Name        string
	Description string
	PageCount   int    // 0 = any, 8 = must be 8 pages, etc.
	Algorithm   string // "8-page-fold", "16-page-booklet", "stack"
}

// ImpositionResult contains the sheet layout for printing.
type ImpositionResult struct {
	Sheets []Sheet `json:"sheets"`
}

// Sheet represents one physical sheet of paper with page placements.
type Sheet struct {
	Front []PagePlacement `json:"front"` // Pages on front side
	Back  []PagePlacement `json:"back"`  // Pages on back side
}

// PagePlacement defines where a zine page appears on a print sheet.
type PagePlacement struct {
	PageIndex int     `json:"page_index"` // Index in zine page list
	X         float64 `json:"x"`          // Position on sheet (normalized 0-1)
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Rotation  int     `json:"rotation"`   // 0, 90, 180, 270 degrees
}
```

### Step 4A.2: Implement 8-Page Fold

**File:** `pkg/imposition/eight_page_fold.go` (NEW)

```go
package imposition

// Compute8PageFold generates imposition for standard 8-page folded zine.
// Page order when folded: 8-1 on outside, 2-7 inside, fold creates 6-3, 4-5
func Compute8PageFold(pageCount int) (*ImpositionResult, error) {
	if pageCount != 8 {
		return nil, fmt.Errorf("8-page fold requires exactly 8 pages, got %d", pageCount)
	}

	result := &ImpositionResult{
		Sheets: []Sheet{
			{
				// Front side (when looking at paper)
				Front: []PagePlacement{
					{PageIndex: 7, X: 0.0, Y: 0, Width: 0.25, Height: 1, Rotation: 0},   // Page 8
					{PageIndex: 0, X: 0.25, Y: 0, Width: 0.25, Height: 1, Rotation: 0},  // Page 1
					{PageIndex: 1, X: 0.5, Y: 0, Width: 0.25, Height: 1, Rotation: 0},   // Page 2
					{PageIndex: 6, X: 0.75, Y: 0, Width: 0.25, Height: 1, Rotation: 0},  // Page 7
				},
				// Back side (flip paper over)
				Back: []PagePlacement{
					{PageIndex: 5, X: 0.0, Y: 0, Width: 0.25, Height: 1, Rotation: 0},   // Page 6
					{PageIndex: 2, X: 0.25, Y: 0, Width: 0.25, Height: 1, Rotation: 0},  // Page 3
					{PageIndex: 3, X: 0.5, Y: 0, Width: 0.25, Height: 1, Rotation: 0},   // Page 4
					{PageIndex: 4, X: 0.75, Y: 0, Width: 0.25, Height: 1, Rotation: 0},  // Page 5
				},
			},
		},
	}

	return result, nil
}
```

**Similar for:**
- `sixteen_page_booklet.go` - Two sheets, stapled
- `simple_stack.go` - No imposition, sequential

---

## Phase 4B: PDF Generation

### Step 4B.1: Add PDF Dependency

**File:** `go.mod`

```bash
go get github.com/jung-kurt/gofpdf
```

### Step 4B.2: Implement PDF Export

**File:** `pkg/export/pdf.go` (NEW)

```go
package export

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"github.com/go-go-golems/zine-layout/pkg/imposition"
)

// ExportZineToPDF creates a print-ready PDF with imposition.
func ExportZineToPDF(
	zinePages []string,      // File paths to rendered pages
	impositionTemplate string,
	outputPath string,
	includeCropMarks bool,
) error {
	// 1. Load imposition template
	imposition, err := loadImpositionTemplate(impositionTemplate)
	if err != nil {
		return err
	}

	// 2. Calculate imposition
	layout, err := imposition.Compute(len(zinePages))
	if err != nil {
		return err
	}

	// 3. Create PDF
	pdf := gofpdf.New("L", "in", "Letter", "")
	
	// 4. For each sheet
	for _, sheet := range layout.Sheets {
		// Add front page
		pdf.AddPage()
		for _, placement := range sheet.Front {
			// Load page image
			// Position on sheet according to placement
			// Apply rotation
		}

		// Add back page
		pdf.AddPage()
		for _, placement := range sheet.Back {
			// Similar
		}

		// Add crop marks if requested
		if includeCropMarks {
			// Draw crop marks at corners
		}
	}

	// 5. Save PDF
	return pdf.OutputFileAndClose(outputPath)
}
```

---

## Phase 4C: Export Service

### Step 4C.1: Create Export Service

**File:** `pkg/services/export.go` (NEW)

```go
package services

type ExportService struct {
	repos        *repo.Repositories
	pagesService *PagesService
}

// ExportZine renders and exports a complete zine.
func (s *ExportService) ExportZine(
	zineID string,
	impositionTemplateID string,
	format string, // "pdf" | "png" | "zip"
	outputDir string,
) (string, error) {
	// 1. Fetch zine
	zine, err := s.repos.Zines.Get(zineID)
	if err != nil {
		return "", fmt.Errorf("fetch zine: %w", err)
	}

	// 2. Fetch zine pages
	zinePages, err := s.repos.Zines.GetPages(zineID)
	if err != nil {
		return "", fmt.Errorf("fetch zine pages: %w", err)
	}

	// 3. Render each page (if not already rendered)
	pagePaths := make([]string, len(zinePages))
	for i, zinePage := range zinePages {
		page, err := s.pagesService.RenderPage(zinePage.LaidOutPageID, outputDir)
		if err != nil {
			return "", fmt.Errorf("render page %s: %w", zinePage.LaidOutPageID, err)
		}
		// Extract path from result JSON
		pagePaths[i] = extractRenderPath(page.ResultJSON)
	}

	// 4. Apply imposition and export
	switch format {
	case "pdf":
		return exportToPDF(pagePaths, impositionTemplateID, outputDir)
	case "png":
		return exportToPNGSequence(pagePaths, outputDir)
	case "zip":
		return exportToZIP(pagePaths, outputDir)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
```

### Step 4C.2: Add Export REST Endpoint

**File:** `pkg/serve/zines_routes.go`

```go
// POST /api/zines/{id}/export
func (s *Server) handleExportZine(w http.ResponseWriter, r *http.Request, zineID string) {
	var req struct {
		ImpositionTemplateID string `json:"imposition_template_id"`
		Format               string `json:"format"` // pdf, png, zip
		IncludeCropMarks     bool   `json:"include_crop_marks"`
		IncludeBleed         bool   `json:"include_bleed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Export in background or synchronously
	outputPath, err := s.exportService.ExportZine(
		zineID,
		req.ImpositionTemplateID,
		req.Format,
		s.dataRoot,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return download URL or file
	http.ServeFile(w, r, outputPath)
}
```

---

## Phase 4D: ZineTab Frontend

### Step 4D.1: Update ZineTab with Real API

**File:** `web/src/views/tabs/ZineTab.tsx`

**Replace dummy data with RTK Query hooks:**

```typescript
export const ZineTab: React.FC<ZineTabProps> = ({ projectId }) => {
  const { data: zines } = useGetZinesQuery({ projectId });
  const { data: laidOutPages } = useGetLaidOutPagesQuery({ projectId });
  const { data: impositionTemplates } = useGetImpositionTemplatesQuery();
  
  const [createZine] = useCreateZineMutation();
  const [updateZinePages] = useUpdateZinePagesMutation();
  const [exportZine] = useExportZineMutation();
  
  // ... rest of implementation
};
```

**Add export with progress:**

```typescript
const handleExport = async () => {
  try {
    setIsExporting(true);
    const result = await exportZine({
      zineId: selectedZine,
      impositionTemplateId: selectedImposition,
      format: exportFormat,
      includeCropMarks,
      includeBleed,
    }).unwrap();
    
    // Download file
    window.location.href = result.download_url;
  } catch (err) {
    alert('Export failed: ' + err.message);
  } finally {
    setIsExporting(false);
  }
};
```

---

## Phase 4E: Final Testing & Launch

### Integration Test Script

**File:** `test/integration/complete_workflow_test.sh`

```bash
#!/bin/bash
set -e

# 1. Create project
PROJECT_ID=$(curl -X POST http://localhost:8088/api/projects \
  -d '{"name":"Integration Test"}' | jq -r '.project.id')

# 2. Upload images
# (multipart upload)

# 3. Create image sequence
SEQ_ID=$(curl -X POST http://localhost:8088/api/projects/$PROJECT_ID/image-sequences \
  -d '{"name":"Test Sequence"}' | jq -r '.sequence.id')

# 4. Create image layout template
TMPL_ID=$(curl -X POST http://localhost:8088/api/image-layout-templates \
  -d '{"name":"8x10","settings":{...}}' | jq -r '.template.id')

# 5. Apply template to create laid-out images
LOI_ID=$(curl -X POST http://localhost:8088/api/projects/$PROJECT_ID/laid-out-images \
  -d '{"asset_id":"...","template_id":"'$TMPL_ID'"}' | jq -r '.laid_out_image.id')

# 6. Create page template
PAGE_TMPL_ID=$(curl -X POST http://localhost:8088/api/page-templates \
  -d '{"name":"Letter","settings":{...}}' | jq -r '.template.id')

# 7. Create print page
PAGE_ID=$(curl -X POST http://localhost:8088/api/projects/$PROJECT_ID/laid-out-pages \
  -d '{"page_template_id":"'$PAGE_TMPL_ID'","laid_out_image_id":"'$LOI_ID'"}' | jq -r '.page.id')

# 8. Create zine
ZINE_ID=$(curl -X POST http://localhost:8088/api/projects/$PROJECT_ID/zines \
  -d '{"name":"Test Zine","page_ids":["'$PAGE_ID'"]}' | jq -r '.zine.id')

# 9. Export zine
curl -X POST http://localhost:8088/api/zines/$ZINE_ID/export \
  -d '{"format":"pdf","imposition_template_id":"8-page-fold"}' \
  -o test-zine.pdf

echo "✓ Complete workflow succeeded!"
```

### UI Testing Checklist

**Manual Test Steps:**

1. **Assets Tab**
   - [ ] Upload 10 images
   - [ ] View asset details
   - [ ] Delete one asset

2. **Sequences Tab**
   - [ ] Create sequence "Test Sequence"
   - [ ] Add all images
   - [ ] Reorder with drag-and-drop
   - [ ] Insert gap
   - [ ] Play slideshow

3. **Image Layouts Tab**
   - [ ] Create template with visual controls
   - [ ] Adjust all sliders and see live preview
   - [ ] Save template
   - [ ] Create single laid-out image
   - [ ] Batch apply to sequence
   - [ ] Verify grid shows all images

4. **Page Layouts Tab**
   - [ ] Create page template (8×10")
   - [ ] Create spread template (16×10" with gutter)
   - [ ] Apply template to laid-out image
   - [ ] View print page preview
   - [ ] Verify margins shown correctly
   - [ ] For spread: verify left/right split

5. **Zine Tab**
   - [ ] Create zine
   - [ ] Add print pages
   - [ ] Reorder pages
   - [ ] Select imposition template
   - [ ] Preview fold diagram
   - [ ] Export as PDF
   - [ ] Verify PDF opens and pages are in correct order

---

## Implementation Order (Recommended)

### Sprint 1: Page Rendering Core (Week 1)
- [ ] Day 1: Define `PageLayoutSettings` struct in Go + TypeScript
- [ ] Day 2: Implement `ComputePageLayout()` function with tests
- [ ] Day 3: Implement single page renderer
- [ ] Day 4: Implement spread renderer with gutter
- [ ] Day 5: Wire up preview/export endpoints to use renderer

### Sprint 2: Frontend Connection (Week 2)
- [ ] Day 1-2: Rewrite PageLayoutsTab (remove dummy, use real hooks)
  - ✅ Hooks already available: `useGetPageTemplatesQuery`, `useCreatePageTemplateMutation`, etc.
  - Build template editor with visual controls
  - Build print pages grid
- [ ] Day 3-4: Update ZineTab (remove dummy, use real hooks)
  - ✅ Hooks already available: `useGetZinesQuery`, `useCreateZineMutation`, etc.
  - Connect zine CRUD
  - Connect page ordering
- [ ] Day 5: End-to-end testing of Phase 3 workflows

### Sprint 3: Imposition & Export (Week 3)
- [ ] Day 1-2: Imposition algorithms (8-page fold, 16-page booklet)
- [ ] Day 3-4: PDF generation with imposition
- [ ] Day 5: Add export endpoint and test

### Sprint 4: Polish & Testing (Week 4)
- [ ] Day 1-2: Crop marks and bleed rendering
- [ ] Day 3: Physical print tests (fold and verify)
- [ ] Day 4-5: Bug fixes and performance optimization

### Sprint 5: Launch (Optional - if needed)
- [ ] Day 1-2: Final bug fixes
- [ ] Day 3: Documentation updates
- [ ] Day 4: User testing
- [ ] Day 5: Production deployment

**Revised Timeline:** 3-4 weeks (down from 5 weeks)

---

## Detailed Task Breakdown

### Task 3.1: PageLayoutSettings Struct ⏱️ 2 hours

**What:** Define Go struct for page layout settings

**Steps:**
1. Create `pkg/pagelayout/types.go`
2. Define `PageLayoutSettings` with all fields
3. Define `PageLayoutResult` for computation output
4. Define `PageLayoutComputation` wrapper
5. Add JSON tags
6. Add validation function
7. Write struct tests

**Done when:** `go build ./pkg/pagelayout` succeeds

---

### Task 3.2: Page Layout Engine ⏱️ 8 hours

**What:** Calculate image placement on page

**Steps:**
1. Create `pkg/pagelayout/engine.go`
2. Implement `ComputePageLayout()`:
   - Convert inches to pixels
   - Calculate content area (page - margins)
   - Handle fill mode (default)
   - Handle absolute mode (x, y, w, h)
   - Handle snap mode (anchor within content)
3. Implement spread mode:
   - Calculate combined spread dimensions
   - Calculate gutter position and overlap
   - Calculate left/right page dimensions
4. Write unit tests:
   - Single page, fill mode
   - Single page, absolute mode
   - Single page, snap mode
   - Spread with center gutter
   - Spread with left/right gutter
5. Edge cases:
   - Zero margins
   - Margins larger than page
   - Negative overlap

**Done when:** All tests pass, coverage > 80%

---

### Task 3.3: Single Page Renderer ⏱️ 6 hours

**What:** Render laid-out image onto page canvas

**Steps:**
1. Create `pkg/pagelayout/renderer.go`
2. Implement `RenderSinglePage()`:
   - Create canvas (page dimensions)
   - Fill background (white or custom)
   - Load laid-out image
   - Resize image to target dimensions
   - Draw image at calculated position
   - Add bleed area if enabled
   - Add crop marks if enabled
3. Write tests with sample images
4. Test output quality at 300 DPI

**Dependencies:** `github.com/disintegration/imaging`

**Done when:** Can render test image on 8×10" page with margins

---

### Task 3.4: Spread Renderer ⏱️ 10 hours

**What:** Split wide image into left/right pages with gutter

**Steps:**
1. Implement `RenderSpreadPages()`:
   - Create full spread canvas
   - Resize wide laid-out image to spread dimensions
   - Draw image on spread
   - Calculate gutter split position
   - Extract left page (0 to gutterPos + overlap)
   - Extract right page (gutterPos - overlap to end)
   - Save three files: left, right, combined
2. Handle overlap properly:
   - Left page extends `overlap` past gutter
   - Right page extends `overlap` before gutter
   - When bound, overlap disappears into binding
3. Add gutter visualization (optional debug mode)
4. Test with panoramic images
5. Verify page dimensions match expectations

**Complex part:** Gutter math must account for binding loss

**Reference:** 
- See `02-image-resizer-code.tsx` lines 128-611 for complete client-side implementation
- See `16-spread-rendering-visualization-guide.md` for detailed rendering guide with diagrams
- Code examples in guide show all 4 render variants (left, right, combined, full)

**Render Variants to Implement:**
1. Left page (with overlap) - For printing
2. Right page (with overlap) - For printing
3. Combined (with gap) - For preview
4. Full with gutter overlay - For debugging

**Visualization Features:**
- Semi-transparent red overlay on gutter area
- Dashed red lines showing overlap zones
- Page borders (2px, dark gray)
- "GUTTER" text label
- Configurable via `RenderOptions` struct

**Done when:** 
- Can split 16×10" image into two 8.125×10" pages
- Can export all 4 variants via API
- Debug mode shows gutter overlay correctly

---

### Task 3.5: Page Template REST API ⏱️ 4 hours

**What:** CRUD endpoints for page templates

**Steps:**
1. Create `pkg/serve/page_templates_routes.go`
2. Implement handlers:
   - `GET /api/page-templates` (global)
   - `GET /api/projects/{id}/page-templates` (global + project)
   - `POST /api/projects/{id}/page-templates` (create project template)
   - `POST /api/page-templates` (create global template)
   - `GET /api/page-templates/{id}`
   - `PUT /api/page-templates/{id}`
   - `DELETE /api/page-templates/{id}`
3. Add response adapters in `pkg/serve/types.go`
4. Register routes in `server.go`
5. Test with curl

**Pattern:** Follow `image_layout_templates_routes.go` exactly

**Done when:** Can CRUD page templates via API

---

### Task 3.6: Laid-Out Pages REST API ⏱️ 6 hours

**What:** CRUD + preview/export for print pages

**Steps:**
1. Create `pkg/serve/laid_out_pages_routes.go`
2. Implement handlers:
   - `GET /api/projects/{id}/laid-out-pages`
   - `POST /api/projects/{id}/laid-out-pages`
   - `GET /api/laid-out-pages/{id}`
   - `PUT /api/laid-out-pages/{id}` (update image)
   - `DELETE /api/laid-out-pages/{id}`
   - `GET /api/laid-out-pages/{id}/preview` (JSON payload initially)
   - `GET /api/laid-out-pages/{id}/export` (PNG file)
3. Wire up `PagesService`
4. Handle spread vs single page in preview
5. Test all endpoints

**Done when:** Can create page, preview, and export via API

**Enhancement - Spread Export Variants:**

For spreads, support exporting different variants via query parameters:

**API Enhancement:**
```
GET /api/laid-out-pages/{id}/export?variant={variant}&debug={bool}

Query params:
- variant: "single" | "left" | "right" | "combined" | "full"
- debug: "true" | "false" (adds gutter/overlap visualization)
```

**Render Options:**

Based on `02-image-resizer-code.tsx` (lines 128-247), support rendering:
1. **Left page only** - With overlap into gutter, optional dashed line showing overlap zone
2. **Right page only** - With overlap into gutter, optional dashed line showing overlap zone
3. **Combined spread** - Both pages with small gap showing binding area
4. **Full spread** - Uncut wide image with gutter overlay (for debugging)

**Implementation Note:**

```go
// In handleLaidOutPageExport
opts := &pagelayout.RenderOptions{
	RenderLeft:     variant == "left" || variant == "combined",
	RenderRight:    variant == "right" || variant == "combined",
	RenderCombined: variant == "combined",
	RenderFull:     variant == "full" || debug,
	ShowGutter:     debug,
	ShowOverlap:    debug,
}
```

**Usage:**
```bash
# Get left page of spread
curl "http://localhost:8088/api/laid-out-pages/lpg-.../export?variant=left" -o left.png

# Get combined with debug visualization
curl "http://localhost:8088/api/laid-out-pages/lpg-.../export?variant=combined&debug=true" -o debug.png
```

---

### Task 3.7: Update PageLayoutsTab ⏱️ 12 hours → 8 hours (APIs exist)

**What:** Replace dummy with real implementation

**Steps:**
1. Add RTK Query hooks for page templates
2. Add RTK Query hooks for laid-out pages
3. Build template editor (follow `12-page-layout-tab-design.md`):
   - Page size controls (reuse from ImageLayoutsTab)
   - Spread mode checkbox
   - Gutter width/overlap sliders (conditional on spread)
   - Positioning mode radio buttons
   - Conditional sections for absolute/snap
   - Live preview (load sample laid-out image)
4. Build print pages grid:
   - Quick create form
   - Cards with previews
   - Edit/delete actions
5. Build page detail view:
   - Large preview
   - Template info
   - Export button
6. Test complete flow

**Reference:** Use ImageLayoutsTab.tsx as template, modify for page settings

**Done when:** Can create page template, apply to image, see preview in UI

---

### Task 3.8: Zine REST API ⏱️ 4 hours

**What:** CRUD endpoints for zines

**Steps:**
1. Create `pkg/serve/zines_routes.go`
2. Implement handlers:
   - `GET /api/projects/{id}/zines`
   - `POST /api/projects/{id}/zines`
   - `GET /api/zines/{id}`
   - `PUT /api/zines/{id}`
   - `DELETE /api/zines/{id}`
   - `GET /api/zines/{id}/pages`
   - `PUT /api/zines/{id}/pages` (reorder)
3. Wire up `ZinesService`
4. Test page ordering

**Done when:** Can CRUD zines and manage page order via API

---

### Task 4.1: Imposition ~~16 hours~~ → 4 hours! ✅ REUSE EXISTING

**🎉 DISCOVERY:** Imposition system already exists in `pkg/zinelayout`!

**What:** Use existing zinelayout package instead of building from scratch

**Existing Assets:**
- ✅ `pkg/zinelayout/layout.go` - Complete imposition engine
- ✅ `data/presets/10_8_sheet_zine.yaml` - 8-page fold layout
- ✅ `data/presets/11_16_sheet_zine.yaml` - 16-page booklet layout
- ✅ Grid-based placement with rotation support
- ✅ YAML parsing and rendering already implemented

**Steps:**
1. Create `zine_layout_templates` repository (1 hour)
2. Seed database with existing YAML presets (30 minutes)
   - Load `10_8_sheet_zine.yaml` as "8-Page Fold"
   - Load `11_16_sheet_zine.yaml` as "16-Page Booklet"
   - Create simple stack template
3. Create service wrapper (2 hours):
   - Convert laid-out pages to input images
   - Call `zinelayout.CreateOutputImage()`
   - Return output sheets
4. Test with physical printing (30 minutes)

**Reference:** See `17-zinelayout-imposition-integration.md` for complete analysis

**Advantages:**
- ✅ Proven code (already tested)
- ✅ YAML-based (easy to add new layouts)
- ✅ Rotation support built-in
- ✅ Grid placement solved
- ✅ Saves 12 hours of development!

**Done when:** 
- Can load YAML imposition template
- Can generate print sheets using zinelayout
- Physical folding test succeeds

---

### Task 4.2: PDF Generation ⏱️ 12 hours

**What:** Export zine as print-ready PDF

**Steps:**
1. Add PDF library dependency
2. Create `pkg/export/pdf.go`
3. Implement basic PDF export:
   - Load page images
   - Apply imposition layout
   - Position pages on sheets
   - Handle rotation if needed
4. Add crop marks:
   - Draw marks at corners
   - Offset from page edge
5. Add bleed:
   - Extend page boundaries
   - Clip content appropriately
6. Test PDF output:
   - Open in Acrobat/Preview
   - Verify page sizes
   - Check imposition
   - Print test copy

**Validation:** Print actual zine, fold it, verify it works

**Done when:** PDF prints correctly and folds into proper zine

---

### Task 4.3: Export Service Integration ⏱️ 8 hours

**What:** Wire export into REST API

**Steps:**
1. Create `pkg/services/export.go`
2. Implement `ExportZine()`:
   - Fetch zine and pages
   - Render each page if needed
   - Apply imposition
   - Generate PDF/PNG/ZIP
   - Return download path
3. Add export endpoint to `zines_routes.go`
4. Handle async export (for large zines):
   - Background job tracking
   - Progress updates
   - Notification when complete
5. Test with UI

**Done when:** Can click "Export" in UI and download PDF

---

### Task 4.4: Update ZineTab ⏱️ 8 hours

**What:** Connect to real APIs

**Steps:**
1. Add API hooks for zines
2. Remove all dummy data
3. Wire up zine CRUD
4. Wire up page reordering
5. Add imposition template selector (from API)
6. Add export with progress indicator
7. Test complete workflow

**Done when:** Can create zine, order pages, export PDF entirely from UI

---

## Milestone Checklist

### Milestone 3.1: Page Layouts Work End-to-End
- [ ] Can create page template in UI
- [ ] Can apply to laid-out image
- [ ] Preview shows image on page correctly
- [ ] Can export single page as PNG
- [ ] Spread mode splits image correctly
- [ ] Gutter overlap renders properly

### Milestone 3.2: Zines Work End-to-End
- [ ] Can create zine in UI
- [ ] Can add pages in order
- [ ] Can reorder pages with drag-and-drop
- [ ] Preview shows correct page sequence
- [ ] Can export zine pages as PNG sequence

### Milestone 4.1: Imposition Works
- [ ] 8-page fold algorithm correct
- [ ] 16-page booklet algorithm correct
- [ ] Can print and physically fold test zine
- [ ] Pages are in correct reading order when folded

### Milestone 4.2: PDF Export Works
- [ ] PDF generates without errors
- [ ] Imposition applied correctly
- [ ] Crop marks render properly
- [ ] Bleed areas included
- [ ] PDF is print-ready (CMYK, proper resolution)

### Final Milestone: Production Ready
- [ ] All Phase 3 & 4 features functional
- [ ] Complete workflow tested (upload → zine → PDF)
- [ ] No critical bugs
- [ ] Performance acceptable (< 30s for 20-page zine)
- [ ] Documentation updated
- [ ] User guide created

---

## Time Estimates (REVISED)

| Phase | Task | Hours | Status | Cumulative |
|-------|------|-------|--------|------------|
| 3A | Page layout types (Go + TS) | 4 | To Do | 4h |
| 3B | Page rendering service + spreads | 24 | To Do | 28h |
| ~~3C~~ | ~~Page template API~~ | ~~4~~ | ✅ Done | ~~32h~~ |
| ~~3D~~ | ~~Laid-out pages API~~ | ~~6~~ | ✅ Done | ~~38h~~ |
| 3E | PageLayoutsTab (connect + fixes) | 8 | To Do | 36h |
| ~~3F~~ | ~~Zine API~~ | ~~4~~ | ✅ Done | ~~40h~~ |
| 3G | Phase 3 testing | 6 | To Do | **42h** |
| ~~4A~~ | ~~Imposition~~ (reuse zinelayout!) | ~~16~~ **4** | To Do | 46h |
| 4B | PDF generation (simplified) | ~~12~~ **8** | To Do | 54h |
| 4C | Export service (simplified) | ~~8~~ **6** | To Do | 60h |
| 4D | ZineTab (connect) | 4 | To Do | 64h |
| 4E | UI fixes from TODO notes | 4 | To Do | 68h |
| 4F | Final testing | 6 | To Do | **74h** |

**FINAL Total: ~74 hours (under 2 weeks for one developer!)**

**Time Saved:**
- REST APIs already done: -28 hours
- Imposition reuse (pkg/zinelayout): -12 hours  
- PDF simplified (using zinelayout): -4 hours
- Frontend hooks ready: -4 hours
- **Total saved: 48 hours (40% reduction!)**

**Original Estimate:** 120 hours (5 weeks)  
**Actual Remaining:** 74 hours (under 2 weeks)  
**Efficiency Gain:** Nearly half the work eliminated!

---

## Dependencies & Blockers

### External Dependencies

**Required:**
- Go PDF library (choose one):
  - `github.com/jung-kurt/gofpdf` (simple, good for basic PDFs)
  - `github.com/pdfcpu/pdfcpu` (more features, steeper learning curve)
- Image library: `github.com/disintegration/imaging` (likely already in use)

**Optional:**
- `github.com/google/uuid` (if not using existing ID generation)

### Potential Blockers

1. **Spread rendering complexity**: Gutter math can be tricky
   - **Mitigation**: Use reference code from `02-image-resizer-code.tsx`
   - **Fallback**: Start with center gutter only, add left/right later

2. **PDF generation learning curve**: First time using PDF library
   - **Mitigation**: Start with simple single-page PDF
   - **Fallback**: Export PNG sequence initially, PDF later

3. **Imposition correctness**: Easy to get page order wrong
   - **Mitigation**: Print and fold physical test zines
   - **Fallback**: Start with simple stack (no folding)

4. **Performance**: Rendering large zines might be slow
   - **Mitigation**: Background job queue for export
   - **Fallback**: Show progress indicator, allow cancel

---

## Testing Strategy

### Unit Tests

**Page Layout Engine:**
```go
func TestComputePageLayout_SinglePage(t *testing.T) {
	settings := pagelayout.PageLayoutSettings{
		PageWidthIn: 8, PageHeightIn: 10, DPI: 300,
		MarginTopIn: 0.5, MarginRightIn: 0.5,
		MarginBottomIn: 0.5, MarginLeftIn: 0.5,
		PositioningMode: "fill",
	}
	result, err := pagelayout.ComputePageLayout(settings, 2400, 3000)
	assert.NoError(t, err)
	assert.Equal(t, 2400, result.OutputWidthPx)
	assert.Equal(t, 3000, result.OutputHeightPx)
	// Verify content area dimensions
}

func TestComputePageLayout_Spread(t *testing.T) {
	settings := pagelayout.PageLayoutSettings{
		PageWidthIn: 8, PageHeightIn: 10, DPI: 300,
		IsSpread: true, GutterWidthIn: 0.25, GutterOverlapIn: 0.125,
		PositioningMode: "fill",
	}
	result, err := pagelayout.ComputePageLayout(settings, 4800, 3000)
	assert.NoError(t, err)
	assert.True(t, result.IsSpread)
	assert.Equal(t, 4800, result.CombinedWidthPx)
	// Verify left/right page dimensions include overlap
}
```

### Integration Tests

```go
func TestPageLayoutWorkflow(t *testing.T) {
	// 1. Create page template
	// 2. Create laid-out image
	// 3. Create print page
	// 4. Verify result JSON populated
	// 5. Render page
	// 6. Verify output file exists
}

func TestZineExportWorkflow(t *testing.T) {
	// 1. Create 8 print pages
	// 2. Create zine
	// 3. Add pages in order
	// 4. Apply 8-page fold imposition
	// 5. Export as PDF
	// 6. Verify PDF has 2 pages (front/back of sheet)
}
```

### Manual Testing

**Physical Print Test:**
1. Create 8-page test zine with numbered pages (1-8)
2. Export with 8-page fold imposition
3. Print PDF on letter-size paper
4. Fold twice (hamburger style)
5. Verify pages read 1, 2, 3, 4, 5, 6, 7, 8 in order when flipped

**If pages are out of order:** Fix imposition algorithm!

---

## Common Pitfalls & Solutions

### Pitfall 1: Gutter Overlap Confusion

**Problem:** Pages too wide or too narrow when bound

**Solution:** 
```
Left page width = (spread_width / 2) + overlap
Right page width = (spread_width / 2) + overlap
When bound, overlap disappears into binding
```

### Pitfall 2: Imposition Page Order

**Problem:** Pages in wrong order when folded

**Solution:** Print test sheet with page numbers, fold, verify, adjust algorithm

### Pitfall 3: DPI Mismatches

**Problem:** Images blurry or too small

**Solution:** Ensure all calculations use same DPI, verify output dimensions

### Pitfall 4: Memory Issues with Large Images

**Problem:** Rendering 300 DPI images consumes too much RAM

**Solution:** 
- Stream rendering instead of loading all at once
- Downsample for preview, full res for export
- Use disk-based temp files for large operations

---

## Success Criteria

### Phase 3 Complete When:
- ✅ Page templates CRUD works in UI
- ✅ Can create print pages from laid-out images
- ✅ Single pages render correctly
- ✅ Spreads render with correct gutter
- ✅ Zines can be created and pages ordered
- ✅ All REST endpoints functional
- ✅ UI tabs connected to real APIs

### Phase 4 Complete When:
- ✅ Can export 8-page folded zine as PDF
- ✅ Can export 16-page booklet as PDF
- ✅ Physical print test succeeds
- ✅ Crop marks render correctly
- ✅ Bleed areas included
- ✅ Complete workflow: upload → zine → print → fold → SUCCESS!

---

## Reference Documents

**For Implementation (Priority Order):**
1. **`15-phase3-api-status-and-discrepancies.md`** - ⭐ START HERE! What APIs are already done
2. **`17-zinelayout-imposition-integration.md`** - ⭐ How to reuse existing imposition
3. **`18-fixes-for-image-layouts-tab-issues.md`** - ⭐ Quick UI fixes (4 hours)
4. **`16-spread-rendering-visualization-guide.md`** - Complete spread rendering guide
5. **`12-page-layout-tab-design.md`** - UI design for PageLayoutsTab
6. **`10-ui-design-for-the-zine-photo-layout-software.md`** - Complete UI spec
7. **`02-image-resizer-code.tsx`** - Spread rendering reference (client-side)

**For Context:**
- `09-system-specification-after-phase1-and-phase2.md` - System architecture
- `05-expansion-plan-for-zine-layout-platform.md` - Overall plan
- `11-changelog-and-things-we-learned.md` - UI refactor learnings

**Quick Start Guide:**
1. Read `15-phase3-api-status-and-discrepancies.md` → Understand what's done (REST APIs!)
2. Read `17-zinelayout-imposition-integration.md` → Don't rebuild imposition!
3. Read `18-fixes-for-image-layouts-tab-issues.md` → Quick UI wins (40 min)
4. Follow this guide starting with Task 3A.1 (Page layout types)

---

**END OF GUIDE**


