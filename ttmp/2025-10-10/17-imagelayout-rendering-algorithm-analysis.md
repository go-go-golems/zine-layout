# Image Layout Thumbnail Rendering – Algorithm Parity & Backend Mapping

## 1. Purpose
Generate backend thumbnails for laid-out images that match the interactive behaviour exposed in the React prototype (`web/src/views/tabs/ImageLayoutsTab.tsx`). This note documents the current TypeScript logic, the Go engine that already computes viewport geometry, and whether the Go implementation can serve as the source of truth for server-side rendering.

## 2. Inputs & Terminology
- **Template settings** → `imagelayout.ViewportSettings` (JSON persisted with each template).  
- **Asset metadata** → intrinsic width/height coming from the project asset library.  
- **Computation output** → `imagelayout.ViewportResult` (`SourceRect`, `TargetRect`, `CanvasRect`, `Scale`, `Mode`) plus a debug `Trace`.  
- **Rendering goal** → draw the source image into a new canvas of size `CanvasRect`, cropping to `SourceRect` and painting the scaled image into `TargetRect`, then downscale to thumbnail resolution.

## 3. Frontend Behaviour (TypeScript Reference)
The interactive resizer in `ttmp/2025-10-10/02-image-resizer-code.tsx` drives the math that designers validated:
- **Scaling modes** – `baseScale` chooses `cover`, `contain`, `fitWidth`, or `fitHeight` based on target and source dimensions (`02-image-resizer-code.tsx:97-103`).
- **Anchors & drag** – `anchorToTranslation` converts an anchor preset (`0…1` normalized) into x/y offsets, while `clampCover` keeps the image within bounds when using cover mode (`02-image-resizer-code.tsx:105-126`).
- **Page mode** – subtracts margins, converts to content box, applies `cover` when `crop_to_fill` is true, and centers/anchors before drawing (`02-image-resizer-code.tsx:259-320`).
- **Crop mode** – treats requested crop width/height as the viewport and applies the same anchor + clamp logic (`02-image-resizer-code.tsx:320-345`).
- **Fit mode** – either pure scaling (no crop) or cover-style crop based on flags (`02-image-resizer-code.tsx:346-395`).
- **Focus points** – the React proof-of-concept does not yet expose focus coordinates, but earlier TS prototypes (see `ttmp/2025-09-23/book-spread.tsx:157-177`) place the focus target before clamping.

Within `ImageLayoutsTab`, the “Preview” panel simply scales an `<img>` inside a margin box and is therefore **not** the canonical algorithm (`web/src/views/tabs/ImageLayoutsTab.tsx:720-764`). The authoritative math lives in the resizer prototype and DSL spec documents.

## 4. Backend Engine (Go Reference)
`pkg/imagelayout/engine` mirrors the TS behaviour:
- `InputsFromSettings` normalizes DPI, orientation, margins, crop ratio, fit dimensions, positioning units, and focus points (`pkg/imagelayout/engine/engine.go:28-238`).  
- `ComputeViewport` derives the same rectangles:
  - Computes target ratio vs. requested crop ratio and trims the source (`engine.go:260-321`).
  - Applies focus override when provided, otherwise uses anchor/position offsets (`engine.go:301-321` & `engine.go:323-336`).
  - Chooses `cover` vs `contain` scaling, multiplies by `UserScale`, and resolves translation offsets (`engine.go:338-370`).
  - Emits `ViewportResult` for downstream rendering (`engine.go:372-392`).
- Helper functions replicate TS utilities: `computeOffset` ≈ anchor translation (`engine.go:408-426`), `positionOffsets` aligns with normalized vs pixel offsets (`engine.go:428-439`), and `clampFloat`/`resolveFocusTarget` stabilise bounds (`engine.go:441-456`).

Unit tests (`pkg/imagelayout/engine/engine_test.go`) cover contain, crop ratio, cover mode, explicit crop dimensions, fit width/height, and focus positioning, confirming parity with the prototype scenarios.

## 5. Feature Parity Assessment
| Behaviour | TypeScript status | Go engine status | Notes |
|-----------|------------------|------------------|-------|
| Cover vs. contain scaling | `baseScale` (`02-image-resizer-code.tsx:97-103`) | `scale := max/min` (`engine.go:338-355`) | Matching; Go applies `UserScale` after base scale, same as TS zoom. |
| Margin box (page mode) | Page mode math (`02-image-resizer-code.tsx:259-320`) | Margins baked into `CanvasRect` (`engine.go:171-208`, `engine.go:245-259`) | Go handles orientation + DPI; frontend preview only approximates. |
| Crop ratio overrides | Optional ratio detection (`02-image-resizer-code.tsx:329-339`) | `CropRatio` with width/height fallback (`engine.go:121-170`) | Equivalent; Go auto-derives missing crop dimensions. |
| Fit width/height mode | Fit branch (`02-image-resizer-code.tsx:346-395`) | `mode == "fit"` with `FitMode`, `FitWidthPx`, `FitHeightPx` (`engine.go:170-205`) | Parity; Go normalizes unspecified dimension like TS. |
| Anchor presets | `anchorToTranslation` (`02-image-resizer-code.tsx:105-109`) | `resolveAnchor` + `computeOffset` (`engine.go:201-230`, `engine.go:408-422`) | Presets map to identical normalized coordinates. |
| Cover clamping | `clampCover` (`02-image-resizer-code.tsx:111-125`) | Range calculations & `computeOffset`/clamp (`engine.go:287-323`) | Go achieves same effect by constraining offsets. |
| Focus point override | Early TS prototype only | `resolveFocusTarget` + clamp (`engine.go:323-336`, `engine.go:441-456`) | Go already supports focus; UI still needs inputs. |
| Result trace | N/A | `Trace` captures steps (`engine.go:244-397`) | Backend-only aid for debugging. |

Conclusion: the Go engine fully supersedes the TS math for viewport computation. Server-side thumbnail rendering can rely on `ComputeViewport` without re-implementing the geometry in TypeScript.

## 6. Thumbnail Rendering Recipe (Backend)
1. **Load computation** – Read the stored `LayoutComputation.Result` from SQLite (`pkg/services/layout.go:64-106`).  
2. **Prepare canvas** – Create an RGBA canvas of size `ceil(CanvasRect.W)` × `ceil(CanvasRect.H)`. Fill with template background (from `ViewportSettings.Export.Background`, default white).  
3. **Load source image** – Decode the asset image (e.g., via `image/jpeg`, `png`, or `bimg`).  
4. **Crop** – Extract `SourceRect` (`result.SourceRect`) from the decoded image (convert to integers with rounding).  
5. **Scale** – Resize the cropped region to `TargetRect.W/H` (`Scale` already accounts for `UserScale`).  
6. **Composite** – Draw the scaled image at `(CanvasRect.X, CanvasRect.Y)` plus `TargetRect.X/Y`. Because `CanvasRect` already includes margins for page mode, the resulting bitmap visually matches the UI.  
7. **Post-process** – Optional downscale to thumbnail max dimensions, apply background flattening for formats without alpha, then encode as PNG/JPEG using `Export.Format` and `Export.Quality`.  
8. **Cache** – Store thumbnails alongside the laid-out image record (filesystem path recorded in DB) so future requests hit the cached file.

## 7. Integration Checklist
- [ ] Implement `pkg/imagelayout/renderer` with helpers `RenderToImage(result, settings, assetPath)` returning `image.Image` plus trace overlays for debugging.  
- [ ] Extend `LayoutService.CreateLaidOutImage` to trigger thumbnail rendering after computing the result (persist file path & metadata).  
- [ ] Wire `/api/laid-out-images/{id}/preview` to stream the cached thumbnail, recomputing on demand if missing.  
- [ ] Update CLI playbook to include a thumbnail verification step.  
- [ ] Add unit tests comparing known TS scenarios against Go-rendered thumbnails (e.g., hash or pixel diff).  
- [ ] Once UI consumes backend thumbnails, remove the placeholder `<img>` preview (`web/src/views/tabs/ImageLayoutsTab.tsx:930-966`).

## 8. Outstanding Gaps
- React preview currently ignores crop/anchor logic; after backend thumbnails land, update the UI to display the server-rendered preview to avoid duplicate math.  
- Need deterministic rounding rules (floor vs. round half up) when converting float rectangles to integer pixels to prevent off-by-one seams—align Go renderer with the TS prototype expectations.  
- Export endpoint (`pkg/serve/laid_out_images_routes.go`) still returns `501`; once thumbnails exist, extend the same renderer to support full-resolution exports.

## 9. Next Steps
1. Build the renderer module and smoke-test via `zine-layout imagelayout compute` outputs.  
2. Provide fixtures comparing TS prototype renders to Go-generated thumbnails for regression testing.  
3. Update documentation (`05-expansion-plan`, `07-phase2...`, `09-system-spec`) after integrating rendering to capture the new pipeline.

