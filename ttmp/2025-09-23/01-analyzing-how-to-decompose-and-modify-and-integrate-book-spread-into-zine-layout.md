### Purpose

Analyze how to decompose the prototype `PhotobookSpreadDesigner` (from `zine-layout/ttmp/2025-09-23/book-spread.tsx`) and integrate it into the existing `zine-layout/web/src` React application driven by `main.tsx` and `routes/App.tsx`, client-side only.

### Current App Overview

- React + Vite style entry at `web/src/main.tsx`, rendering `<App />` with Redux provider.
- Routing and shell handled in `web/src/routes/App.tsx` via `react-router-dom` with an `AppShell` layout. Existing routes:
  - `/` → `views/Home`
  - `/projects` → `views/Projects`
  - `/projects/:id` → `views/ProjectDetail`
  - `/projects/:id/yaml` → `views/ProjectYamlPage`
- UI components available in `web/src/components/ui` (Button, Card, Input) and other app-specific components.
- Tailwind-style utility classes in use (e.g., `bg-primary-600`).

### Prototype Component Overview (`book-spread.tsx`)

- Single-file page component: `PhotobookSpreadDesigner` with purely client-side state and canvas export.
- Key state (all local):
  - `image: HTMLImageElement | null` with attached metadata (`fileSize`, `fileName`)
  - `paperSize: '4x6'|'5x7'|'8x10'|'11x14'|'12x12'|'8.5x11'|'A4'`
  - `isSpread: boolean`
  - `margins: { top,right,bottom,left }` (inches)
  - `orientation: 'portrait'|'landscape'`
  - `cropRatio: 'original'|'1:1'|'2:3'|...`
  - `cropToFill: boolean`
  - `imageScale: number`, `imagePosition: { x,y }`
  - `dpi: 150|300|600`
  - refs: `canvasRef`, `fileInputRef`
- Rendering flow:
  - "Controls Panel" left; "Preview" and information right, built with Tailwind classes.
  - `renderPreview()` computes content area and optional gutter, crops/scales an `<img>` preview for on-screen rendering (not canvas).
  - `exportImage()` draws to an off-screen `<canvas>` at true pixel dimensions (`inches * dpi`) with clipping and optional gutter, then downloads as PNG.
- Data model assumptions: inches as primary unit, DPI determines export pixels.

### Issues/Adjustments Noticed

- Missing state: `gutterMargin` and `setGutterMargin` are used but never declared. Needs a `useState<number>` (e.g., default `0.25`).
- Types:
  - `image` is assigned ad-hoc props (`fileSize`, `fileName`). Prefer a type augmentation: `type LoadedImage = HTMLImageElement & { fileSize?: number; fileName?: string }`.
  - Event handlers should be typed (`React.ChangeEvent<HTMLInputElement>`, `React.DragEvent<HTMLDivElement>`).
  - Refs should be typed (`useRef<HTMLInputElement | null>(null)`).
- Layout wrapper conflict: Component contains a top-level `min-h-screen bg-gray-50` wrapper which duplicates `AppShell` responsibilities. For integration, remove/adjust outer wrapper and rely on the shell’s `main` container.
- Styling: Could optionally replace raw containers with existing `Card`, `Button`, `Input` from `components/ui` for visual consistency (not required for a minimal port).
- Performance: Calculations happen inline in render; acceptable for now. Later we can extract computations into `useMemo` and draw-related operations into a hook if necessary.

### Integration Goals (Client-side Only)

- Add a new route/page that hosts the designer without backend dependencies.
- Keep image selection via file input and drag-and-drop (no project storage yet).
- Avoid Redux for now; keep all state local to the page.
- Minimal invasive changes to the existing app shell and navigation.

### Two Integration Paths

1) Minimal Drop-in (Fastest)
- Create `web/src/views/SpreadDesigner.tsx` by moving the prototype content into a page component.
- Fix TypeScript types and add missing `gutterMargin` state.
- Remove the outer `min-h-screen` wrapper to fit inside `AppShell`.
- Add a route: `/designer/spread` in `routes/App.tsx`.
- Optional: Add a nav link “Designer”.

2) Structured Refactor (Clean Separation)
- Create a small feature module under `web/src/features/spread/`:
  - `components/SpreadPreview.tsx` — presentational preview receiving computed props.
  - `components/SpreadControls.tsx` — upload + controls panel.
  - `hooks/useSpreadDesigner.ts` — encapsulate state, calculations, and export.
  - `views/SpreadDesignerPage.tsx` — page composition.
- Same route `/designer/spread`. UI consistency via `components/ui` replacements.

For now, proceed with Path 1 to ship quickly.

### Minimal Drop-in: Concrete Changes

- New file: `web/src/views/SpreadDesigner.tsx`
  - Import React, use state hooks with proper types.
  - Add missing `gutterMargin` state.
  - Type `image` as `LoadedImage | null`.
  - Type refs and events.
  - Remove/adjust outer layout wrapper; keep inner structure.
  - Optionally use existing `Button` for the Export and keep current Tailwind classes otherwise.

- Update router: `web/src/routes/App.tsx`
  - Add new route mapping: `<Route path="/designer/spread" element={<SpreadDesigner />} />`.
  - Optionally add a nav `<Link>` to `/designer/spread`.

#### Example: new page skeleton

```tsx
// web/src/views/SpreadDesigner.tsx
import React, { useState, useRef, useCallback } from 'react';

// Optional: reuse app UI components later
// import { Button, Card, Input } from '../components/ui';

type LoadedImage = HTMLImageElement & { fileSize?: number; fileName?: string };

type Margins = { top: number; right: number; bottom: number; left: number };

type Orientation = 'portrait' | 'landscape';

type PaperSizeKey = '4x6'|'5x7'|'8x10'|'11x14'|'12x12'|'8.5x11'|'A4';

type CropKey = 'original'|'1:1'|'2:3'|'3:4'|'4:5'|'5:7'|'16:9';

export const SpreadDesigner: React.FC = () => {
  const [image, setImage] = useState<LoadedImage | null>(null);
  const [paperSize, setPaperSize] = useState<PaperSizeKey>('8x10');
  const [isSpread, setIsSpread] = useState<boolean>(false);
  const [margins, setMargins] = useState<Margins>({ top: 0.5, right: 0.5, bottom: 0.5, left: 0.5 });
  const [orientation, setOrientation] = useState<Orientation>('portrait');
  const [cropRatio, setCropRatio] = useState<CropKey>('original');
  const [cropToFill, setCropToFill] = useState<boolean>(false);
  const [imageScale, setImageScale] = useState<number>(1);
  const [imagePosition, setImagePosition] = useState<{ x: number; y: number }>({ x: 0, y: 0 });
  const [dpi, setDpi] = useState<number>(300);
  const [gutterMargin, setGutterMargin] = useState<number>(0.25);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  // ... port logic from prototype here, adjust event/ref types ...

  return (
    <div className="space-y-6">{/* Let AppShell handle page padding/background */}
      {/* Controls + Preview layout, copied from prototype with minor adjustments */}
    </div>
  );
};

export default SpreadDesigner;
```

#### Example: router update

```tsx
// web/src/routes/App.tsx
import { SpreadDesigner } from '../views/SpreadDesigner';

// inside <Routes>
<Route path="/designer/spread" element={<SpreadDesigner />} />

// optional in nav
<Link
  to="/designer/spread"
  className={`px-3 py-2 rounded-md text-sm font-medium transition-colors duration-200 ${
    location.pathname.startsWith('/designer')
      ? 'bg-primary-100 text-primary-700'
      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
  }`}
>
  Designer
</Link>
```

### Decomposition (Later Refactor)

If we proceed to Path 2, propose the following responsibilities and props:

- `SpreadPreview` (pure):
  - Props: `image`, derived `contentWidth/Height/X/Y`, `displayWidth/Height`, `positions`, `hasGutter`, `left/right panel dims`, `gutterWidth`, `styles`, `orientation`, `isSpread`.
  - Renders on-screen `<img>` preview only.

- `SpreadControls` (stateful UI):
  - Props: set-state callbacks, current values for size/margins/crop/scale/position/dpi/gutter.
  - Handles upload via file input and drag-and-drop.

- `useSpreadDesigner` (logic):
  - Holds state + derives preview/export numbers with `useMemo`.
  - Exposes `exportImage()` drawing to canvas.

- `SpreadDesignerPage` composes the hook + components.

### TypeScript and Event Typing Checklist

- `handleImageDrop(e: React.DragEvent<HTMLDivElement>)`
- `handleImageSelect(e: React.ChangeEvent<HTMLInputElement>)`
- `fileInputRef = useRef<HTMLInputElement | null>(null)`
- `image: LoadedImage | null`
- Avoid `any`; annotate helper functions where inference is unclear.

### UI/UX Adjustments for AppShell

- Remove `min-h-screen bg-gray-50 p-6` wrapper from the prototype. `AppShell` already sets background and spacing. Use inner containers with `Card` where appropriate.
- Optional: replace raw buttons/inputs with `components/ui` versions for consistency.

### Client-Only Note

- All operations remain in-browser. Image never leaves client; export uses `canvas.toDataURL('image/png')` and a temporary `<a>` download.
- No interaction with `store` or `api` slices is necessary for the initial integration.

### Risks and Edge Cases

- Very large images may consume significant memory at high DPI. Consider warning labels or clamping max export pixels later.
- Gutter math must be consistent between preview and export; ensure both use identical ratio calculations.
- Crop-to-fill vs. ratio-crop logic: validate with landscape vs. portrait images and spread vs. single page.
- Accessibility: inputs/sliders should have labels (already present) and aria attributes if needed.

### Acceptance Criteria

- Visiting `/designer/spread` displays the designer.
- Upload via click or drag-and-drop shows preview with margin/gutter guides.
- Adjusting controls updates the preview in real time.
- Export downloads a PNG at `width * dpi` by `height * dpi` pixels, honoring margins and gutter.
- No backend calls. Page fits within existing `AppShell` without visual conflicts.

### Future Enhancements

- Persist last-used settings in `localStorage` or Redux.
- Integrate with `ImageTray` to pick from project images.
- Add presets and keyboard controls for nudging position/scale.
- Lazy-load image and use `createImageBitmap` for performance when available.
- Add unit tests for the crop/scale math (extract math into pure functions).

### Quick Task List (Path 1)

- [ ] Create `web/src/views/SpreadDesigner.tsx` with typed state and missing `gutterMargin`.
- [ ] Port prototype JSX, remove outer page wrapper, keep layout.
- [ ] Wire route `/designer/spread` in `routes/App.tsx`.
- [ ] Optional: Add nav link to Designer.
- [ ] Manual smoke test with a few portrait/landscape images.
