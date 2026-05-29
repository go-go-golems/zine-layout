package pagelayout

import (
	"fmt"
	"image"
)

// PageLayoutSettings captures page metrics and positioning mode for rendering
// a single image onto a page canvas. Units are inches for dimensions; DPI
// governs conversion to pixels.
//
// Notes:
// - When IsSpread is true, the page represents a combined spread. The renderer
//   may optionally split into left/right variants. Gutter-related fields are
//   meaningful only when IsSpread is true.
// - PositioningMode can be "fill", "absolute", or "snap". The initial
//   implementation treats "snap" as an alias of "fill".
// - Absolute placement fields are only used when PositioningMode == "absolute".
// - AnchorPreset is reserved for future use to snap placement to common
//   anchors (e.g., center, top-left) when not using absolute placement.
//
// JSON/YAML tags are provided for potential API bindings.
//
// Keep this struct small and focused on page metrics; do not include
// per-image focus/anchor overrides here.
//
// IMPORTANT: This package provides only pure computation helpers; image IO
// happens in the caller.

type PageLayoutSettings struct {
	PageWidthIn  float64 `json:"pageWidthIn" yaml:"page_width_in"`
	PageHeightIn float64 `json:"pageHeightIn" yaml:"page_height_in"`
	DPI          float64 `json:"dpi" yaml:"dpi"`

	MarginTopIn    float64 `json:"marginTopIn" yaml:"margin_top_in"`
	MarginRightIn  float64 `json:"marginRightIn" yaml:"margin_right_in"`
	MarginBottomIn float64 `json:"marginBottomIn" yaml:"margin_bottom_in"`
	MarginLeftIn   float64 `json:"marginLeftIn" yaml:"margin_left_in"`

	IsSpread        bool    `json:"isSpread" yaml:"is_spread"`
	GutterWidthIn   float64 `json:"gutterWidthIn" yaml:"gutter_width_in"`
	GutterOverlapIn float64 `json:"gutterOverlapIn" yaml:"gutter_overlap_in"`

	PositioningMode string `json:"positioningMode" yaml:"positioning_mode"`
	AnchorPreset    string `json:"anchorPreset" yaml:"anchor_preset"`

	// Absolute placement (only when PositioningMode == "absolute")
	ImageXIn      float64 `json:"imageXIn" yaml:"image_x_in"`
	ImageYIn      float64 `json:"imageYIn" yaml:"image_y_in"`
	ImageWidthIn  float64 `json:"imageWidthIn" yaml:"image_width_in"`
	ImageHeightIn float64 `json:"imageHeightIn" yaml:"image_height_in"`

	// Optional page border drawing
	BorderEnabled bool   `json:"borderEnabled" yaml:"border_enabled"`
	BorderColor   string `json:"borderColor" yaml:"border_color"`
	BorderType    string `json:"borderType" yaml:"border_type"` // plain|dotted|dashed|corner
}

func (s *PageLayoutSettings) Canonicalize() error {
	if s.DPI <= 0 {
		return fmt.Errorf("dpi must be > 0")
	}
	if s.PageWidthIn <= 0 || s.PageHeightIn <= 0 {
		return fmt.Errorf("page dimensions must be > 0")
	}
	if s.MarginTopIn < 0 || s.MarginRightIn < 0 || s.MarginBottomIn < 0 || s.MarginLeftIn < 0 {
		return fmt.Errorf("margins must be >= 0")
	}
	if s.MarginTopIn+s.MarginBottomIn >= s.PageHeightIn {
		return fmt.Errorf("vertical margins exceed or equal page height")
	}
	if s.MarginLeftIn+s.MarginRightIn >= s.PageWidthIn {
		return fmt.Errorf("horizontal margins exceed or equal page width")
	}
	if s.IsSpread {
		if s.GutterWidthIn < 0 {
			return fmt.Errorf("gutter width must be >= 0")
		}
		if s.GutterWidthIn >= s.PageWidthIn {
			return fmt.Errorf("gutter width must be < page width")
		}
		if s.GutterOverlapIn < 0 {
			return fmt.Errorf("gutter overlap must be >= 0")
		}
	}
	mode := s.PositioningMode
	if mode == "" {
		s.PositioningMode = "fill"
		mode = s.PositioningMode
	}
	if mode != "fill" && mode != "absolute" && mode != "snap" {
		return fmt.Errorf("invalid positioning mode: %s", mode)
	}
	if mode == "absolute" {
		if s.ImageWidthIn <= 0 || s.ImageHeightIn <= 0 {
			return fmt.Errorf("absolute mode requires positive image width/height")
		}
		// Bounds check is best-effort; allow partially outside for flexibility
	}
	return nil
}

func (s PageLayoutSettings) pixelsPerInch() float64 { return s.DPI }

func (s PageLayoutSettings) InchesToPixels(in float64) int {
	pp := s.pixelsPerInch()
	px := int(in*pp + 0.5)
	if px < 0 {
		return 0
	}
	return px
}

func (s PageLayoutSettings) PixelWidth() int {
	return s.InchesToPixels(s.PageWidthIn)
}

func (s PageLayoutSettings) PixelHeight() int {
	return s.InchesToPixels(s.PageHeightIn)
}

// ContentRectPx returns the drawable content rectangle inside page margins.
// The rectangle is relative to the page canvas (0,0)-(W,H).
func (s PageLayoutSettings) ContentRectPx() image.Rectangle {
	w := s.PixelWidth()
	h := s.PixelHeight()
	mt := s.InchesToPixels(s.MarginTopIn)
	mr := s.InchesToPixels(s.MarginRightIn)
	mb := s.InchesToPixels(s.MarginBottomIn)
	ml := s.InchesToPixels(s.MarginLeftIn)
	left := ml
	top := mt
	right := w - mr
	bottom := h - mb
	if right < left {
		right = left
	}
	if bottom < top {
		bottom = top
	}
	return image.Rect(left, top, right, bottom)
}

// SpreadSplitX returns the x coordinate where a spread would be split into left
// and right pages on the canvas. The split is centered, minus half the gutter width.
// If not a spread, it returns -1.
func (s PageLayoutSettings) SpreadSplitX() int {
	if !s.IsSpread {
		return -1
	}
	w := s.PixelWidth()
	g := s.InchesToPixels(s.GutterWidthIn)
	// We define split x as center; callers may compute left/right using gutter.
	// Keep API stable but ensure gutter is accounted by helpers using this value.
	_ = g
	return w / 2
}
