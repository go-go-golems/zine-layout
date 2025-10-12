# Image Crop YAML DSL

A declarative YAML format for specifying image resizing and cropping operations.

---

## Schema Definition

### Root Structure

```yaml
version: "1.0"
mode: page | crop | fit | spread
source:
  path: string          # Path to source image
  width: number         # Original width (optional, for validation)
  height: number        # Original height (optional, for validation)
output:
  path: string          # Output file path
  format: png | jpg | webp | tiff
  quality: number       # 1-100 for lossy formats
transform:
  # Mode-specific configuration (see below)
```

---

## Mode A: Page + Margins

```yaml
version: "1.0"
mode: page

source:
  path: "input/photo.jpg"

output:
  path: "output/framed-photo.png"
  format: png

transform:
  page:
    width: 1920         # Page dimensions in pixels
    height: 1080
  
  margins:
    left: 100
    right: 100
    top: 100
    bottom: 100
    # OR for uniform margins:
    # all: 100
  
  fit:
    mode: cover | contain     # cover = fill/crop, contain = fit inside
  
  position:
    method: anchor | drag | none
    # Method: anchor
    anchor:
      horizontal: left | center | right    # or 0.0-1.0
      vertical: top | middle | bottom      # or 0.0-1.0
    # Method: drag
    drag:
      x: 0                # Pixel offset from center
      y: 0
    # Method: none - just centered with optional manual drag
  
  adjustments:
    zoom: 1.0             # Scale multiplier
    drag:                 # Additional manual drag (added to anchor)
      x: 0
      y: 0
```

### Complete Example: Page Mode

```yaml
version: "1.0"
mode: page

source:
  path: "photos/landscape.jpg"
  width: 4000
  height: 3000

output:
  path: "prints/framed-landscape.png"
  format: png

transform:
  page:
    width: 2400
    height: 1600
  
  margins:
    all: 120
  
  fit:
    mode: cover          # Fill the visible area, crop if needed
  
  position:
    method: anchor
    anchor:
      horizontal: center
      vertical: middle
  
  adjustments:
    zoom: 1.2
    drag:
      x: -50
      y: 20
```

---

## Mode B: Fixed Crop

```yaml
version: "1.0"
mode: crop

source:
  path: "input/photo.jpg"

output:
  path: "output/cropped.jpg"
  format: jpg
  quality: 90

transform:
  crop:
    width: 1000
    height: 1000
    # OR use aspect ratio
    aspect: "1:1" | "3:2" | "2:3" | "4:3" | "3:4" | "5:4" | "4:5" | "16:9" | "9:16"
    # If aspect is set, only width or height is required
  
  fit:
    mode: cover | contain
  
  position:
    method: anchor | focus
    # Method: anchor
    anchor:
      horizontal: left | center | right
      vertical: top | middle | bottom
    # Method: focus (align specific image point to crop center)
    focus:
      image:              # Point in source image
        x: 1200
        y: 800
      target:             # Point in crop area (often center)
        x: 500            # In crop coordinates
        y: 500
  
  adjustments:
    zoom: 1.0
    drag:
      x: 0
      y: 0
```

### Complete Example: Fixed Crop Mode

```yaml
version: "1.0"
mode: crop

source:
  path: "photos/portrait.jpg"
  width: 3000
  height: 4000

output:
  path: "social/instagram-post.jpg"
  format: jpg
  quality: 95

transform:
  crop:
    aspect: "1:1"
    width: 1080           # Instagram standard
  
  fit:
    mode: cover
  
  position:
    method: focus
    focus:
      image:
        x: 1500           # Focus on subject's face
        y: 1200
      target:
        x: 540            # Center of crop (1080/2)
        y: 540
  
  adjustments:
    zoom: 1.1
    drag:
      x: 0
      y: -30
```

---

## Mode C: Fit to Width/Height

```yaml
version: "1.0"
mode: fit

source:
  path: "input/photo.jpg"

output:
  path: "output/resized.png"
  format: png

transform:
  fit:
    target: width | height
    width: 1920           # If target is width
    height: 1080          # If target is height (or both for crop/fill)
    
    # Crop/fill mode
    crop_fill: true | false
    # If false: simple resize, aspect preserved, one dimension determines size
    # If true: both dimensions enforced, uses cover mode
  
  position:
    # Only used when crop_fill is true
    method: anchor
    anchor:
      horizontal: center
      vertical: middle
  
  adjustments:
    zoom: 1.0
    drag:
      x: 0
      y: 0
```

### Complete Example: Fit Mode (No Crop)

```yaml
version: "1.0"
mode: fit

source:
  path: "photos/hero-image.jpg"

output:
  path: "web/hero-1920.jpg"
  format: jpg
  quality: 85

transform:
  fit:
    target: width
    width: 1920
    crop_fill: false      # Preserve aspect, height auto-calculated
  
  adjustments:
    zoom: 1.0
```

### Complete Example: Fit Mode (With Crop)

```yaml
version: "1.0"
mode: fit

source:
  path: "photos/banner.jpg"

output:
  path: "web/banner-exact.jpg"
  format: jpg
  quality: 85

transform:
  fit:
    target: width         # Target can be width or height when crop_fill is true
    width: 1920
    height: 600
    crop_fill: true       # Force exact dimensions
  
  position:
    method: anchor
    anchor:
      horizontal: center
      vertical: top       # Keep top of image
  
  adjustments:
    zoom: 1.0
    drag:
      x: 0
      y: 0
```

---

## Batch Processing

Process multiple images with the same settings:

```yaml
version: "1.0"
mode: crop
batch: true

sources:
  - path: "photos/img001.jpg"
  - path: "photos/img002.jpg"
  - path: "photos/img003.jpg"

output:
  directory: "output/thumbnails/"
  format: jpg
  quality: 90
  naming: "{filename}-thumb.{ext}"

transform:
  crop:
    aspect: "1:1"
    width: 400
  
  fit:
    mode: cover
  
  position:
    method: anchor
    anchor:
      horizontal: center
      vertical: middle
```

---

## Presets

Define reusable presets:

```yaml
version: "1.0"

# Define presets
presets:
  instagram_square:
    mode: crop
    transform:
      crop:
        aspect: "1:1"
        width: 1080
      fit:
        mode: cover
      position:
        method: anchor
        anchor:
          horizontal: center
          vertical: middle
  
  instagram_story:
    mode: crop
    transform:
      crop:
        aspect: "9:16"
        width: 1080
      fit:
        mode: cover
      position:
        method: anchor
        anchor:
          horizontal: center
          vertical: middle
  
  print_4x6:
    mode: page
    transform:
      page:
        width: 1800
        height: 1200
      margins:
        all: 72          # 0.25 inch at 300 DPI
      fit:
        mode: contain
      position:
        method: anchor
        anchor:
          horizontal: center
          vertical: middle
  
  book_spread_8x10:
    mode: spread
    transform:
      spread:
        width: 4800      # 16" at 300 DPI (8" per page)
        height: 3000     # 10" at 300 DPI
      gutter:
        width: 120       # 0.4" binding
        position: center
        overlap: 60      # 0.2" bleed
      fit:
        mode: cover
      position:
        method: anchor
        anchor:
          horizontal: center
          vertical: middle

# Use a preset
operations:
  - source:
      path: "photo1.jpg"
    output:
      path: "social/post1.jpg"
    preset: instagram_square
    
  - source:
      path: "photo2.jpg"
    output:
      path: "social/story1.jpg"
    preset: instagram_story
    adjustments:
      zoom: 1.2        # Override preset settings
```

---

## Advanced Features

### Conditional Processing

```yaml
version: "1.0"
mode: crop

source:
  path: "input.jpg"

output:
  path: "output.jpg"
  format: jpg
  quality: 90

transform:
  crop:
    # Use different sizes based on source dimensions
    width: "{{ source.width > 2000 ? 1080 : 800 }}"
    aspect: "{{ source.width > source.height ? '16:9' : '9:16' }}"
  
  fit:
    mode: cover
  
  position:
    method: anchor
    anchor:
      horizontal: center
      vertical: middle
```

### Variables and Templates

```yaml
version: "1.0"

variables:
  social_size: 1080
  standard_zoom: 1.1
  output_quality: 95

mode: crop

source:
  path: "input.jpg"

output:
  path: "output.jpg"
  format: jpg
  quality: "{{ output_quality }}"

transform:
  crop:
    aspect: "1:1"
    width: "{{ social_size }}"
  
  fit:
    mode: cover
  
  position:
    method: anchor
    anchor:
      horizontal: center
      vertical: middle
  
  adjustments:
    zoom: "{{ standard_zoom }}"
```

---

## Schema Validation

### Anchor Values
- **String**: `left`, `center`, `right`, `top`, `middle`, `bottom`
- **Numeric**: 0.0 to 1.0 (0 = left/top, 0.5 = center/middle, 1.0 = right/bottom)

### Fit Modes
- **cover**: Fill target area completely (may crop)
- **contain**: Fit entirely inside target area (may letterbox)

### Position Methods
- **none**: No anchoring, just centered (manual drag only)
- **anchor**: Use 9-point grid or continuous positioning
- **drag**: Manual pixel offset only
- **focus**: Align specific image point to target point

### Gutter Positions (Spread Mode)
- **center**: Gutter at center of spread (symmetric pages)
- **left**: Gutter at left edge (asymmetric, small left page)
- **right**: Gutter at right edge (asymmetric, small right page)

### Units
All dimensions in pixels unless otherwise specified.

---

## Implementation Notes

This DSL can be:
1. **Parsed** by image processing tools to automate resizing operations
2. **Generated** by UI tools (like our web app) to export configurations
3. **Version controlled** alongside images for reproducible builds
4. **Shared** between team members for consistent processing
5. **Extended** with custom fields for specific use cases

Example parser usage:
```bash
image-resize --config crop-config.yml
image-resize --batch batch-process.yml
image-resize --preset instagram_square input.jpg -o output.jpg
image-resize --mode spread --config book-spread.yml input.jpg
```

### Spread Mode Special Handling

When processing spread mode:
- Parser generates multiple output files (combined, left, right, optional full)
- Gutter visualization should be shown in preview but not in final output pages
- Left and right pages include overlap area for binding bleed
- Combined preview includes a small gap to represent the physical binding