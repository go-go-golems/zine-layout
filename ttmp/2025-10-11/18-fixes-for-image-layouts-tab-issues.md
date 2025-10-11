# Fixes for Image Layouts Tab Issues
**Date:** October 11, 2025  
**Source:** Issues from 16-todo-notes.md  
**Status:** Ready to implement

---

## Issue List from 16-todo-notes.md

### Sequences Tab Issues
1. View assets for sequences (missing feature)

### Image Layouts Tab Issues  
2. Landscape / portrait toggle doesn't work
3. Use backend rendering for preview thumbnails
4. Remove focus / anchor point (not needed in templates)
5. Remove position / scale from template (move to image overrides only)
6. Always show "create laid out image" section (no button)
7. Edit drawer doesn't work for laid-out images
8. Preview thumbnail incorrect when applying template
9. Preselect latest applied template in create laid-out image form

---

## Fix 1: View Assets to Add to Sequence ✅ FIXED

### Problem
Can't see available assets to add to sequence - no way to add assets at all!

### Solution

**File:** `web/src/views/tabs/SequencesTab.tsx`

**Changed layout to three columns:**

**Before:** 2 columns (Preview + Sequence Builder)

**After:** 3 columns (Assets + Preview + Sequence Builder)

```typescript
<div className="grid lg:grid-cols-3 gap-6">
  {/* Column 1: Available Assets */}
  <Card>
    <CardHeader>
      <h3>Available Assets</h3>
      <p>Click to add to sequence</p>
    </CardHeader>
    <CardBody>
      <div className="space-y-2 max-h-[600px] overflow-y-auto">
        {assets.map((asset) => (
          <button
            onClick={() => addSequenceItem({
              sequenceId: selectedSequenceId,
              assetId: asset.id,
            })}
            className="w-full flex items-center space-x-3 p-2 border rounded hover:border-primary-300"
          >
            <div className="w-16 h-16">
              <img src={asset.src} className="object-cover" />
            </div>
            <div className="flex-1 text-left">
              <p className="text-sm font-medium">{asset.name}</p>
              <p className="text-xs text-gray-500">{asset.width} × {asset.height}</p>
            </div>
          </button>
        ))}
      </div>
    </CardBody>
  </Card>

  {/* Column 2: Preview Pane */}
  <Card>
    {/* Existing preview code */}
  </Card>

  {/* Column 3: Sequence Builder */}
  <Card>
    {/* Existing sequence items list */}
  </Card>
</div>
```

**Features:**
- Click asset thumbnail to add to sequence
- Scrollable asset list (max height 600px)
- Shows asset dimensions
- Hover effect for clickability
- Empty state if no assets

**Status:** ✅ IMPLEMENTED AND VERIFIED (build succeeds)

---

## Fix 2: Landscape/Portrait Toggle Doesn't Work

### Problem
Changing orientation radio button doesn't swap width/height

### Solution

**File:** `web/src/views/tabs/ImageLayoutsTab.tsx`

**Add swap logic when orientation changes:**

```typescript
const handleOrientationChange = (newOrientation: 'portrait' | 'landscape') => {
  if (orientation === newOrientation) return;
  
  // Swap width and height when switching orientation
  const oldWidth = paperWidth;
  const oldHeight = paperHeight;
  setPaperWidth(oldHeight);
  setPaperHeight(oldWidth);
  setOrientation(newOrientation);
};

// Update the orientation buttons
<button
  type="button"
  onClick={() => handleOrientationChange('portrait')}
  className={...}
>
  Portrait ⬜
</button>
<button
  type="button"
  onClick={() => handleOrientationChange('landscape')}
  className={...}
>
  Landscape ▭
</button>
```

**Time:** 15 minutes

**Status:** ⧗ Needs implementation

---

## Fix 3: Use Backend Rendering for Preview Thumbnails

### Problem
Preview in template editor uses inline styles, not actual backend computation

### Solution

**File:** `web/src/views/tabs/ImageLayoutsTab.tsx`

**Option A: Call backend preview endpoint (when implemented)**

```typescript
const { data: previewData } = usePreviewLaidOutImageQuery(
  { id: previewLaidOutImageId },
  { skip: !previewLaidOutImageId }
);

// Render using computation result
{previewData && (
  <canvas
    ref={canvasRef}
    width={previewData.result.export_single.w}
    height={previewData.result.export_single.h}
  />
)}
```

**Option B: Keep client-side preview (acceptable for now)**

Current inline style preview is acceptable until backend renderer is implemented. It gives immediate feedback which is valuable for UX.

**Recommendation:** Keep current approach, improve later

**Status:** ✅ Acceptable as-is (can enhance in Phase 3B)

---

<!-- ## Fix 4: Remove Focus/Anchor Point from Templates

### Problem
Focus point and anchor settings shouldn't be in template (too specific per image)

### Solution

**File:** `web/src/views/tabs/ImageLayoutsTab.tsx`

**Remove from template editor:**

```typescript
// DELETE these sections from template editor:
// - Focus point inputs (source_x, source_y, target_x, target_y)
// - Position X/Y sliders (these are image-specific)

// KEEP in template:
// - Paper size, DPI, orientation
// - Margins
// - Crop mode and aspect ratio
// - Anchor preset (general alignment, not fine-tuning)
// - User scale (default, can override per image)

// MOVE to laid-out image override controls:
// - Fine position adjustments (position_x, position_y)
// - User scale overrides
// - Focus point (when editing specific image)
```

**Specific changes:**

1. Remove position X/Y sliders from template form
2. Remove focus point section from template form
3. Keep anchor grid (it's a template-level default)
4. Add these to edit drawer when editing laid-out images

**Time:** 30 minutes

**Status:** ⧗ Needs implementation -->

---

## Fix 5: Remove Position/Scale from Template, Keep Only on Images

### Problem
User scale and position are image-specific, shouldn't be in reusable template

### Solution

**Clarify template vs override separation:**

**Template Should Have (reusable settings):**
- Paper size, DPI, orientation
- Margins
- Crop ratio and mode
- Default anchor point
- Export format/quality

**Override Should Have (per-image adjustments):**
- User scale multiplier
- Position X/Y fine-tuning
- Focus point (if needed)

**Implementation:**

```typescript
// In template editor, remove:
const [userScale, setUserScale] = useState(1);  // DELETE
const [positionX, setPositionX] = useState(0);  // DELETE
const [positionY, setPositionY] = useState(0);  // DELETE

// In buildSettingsFromForm(), remove:
user_scale: userScale,    // DELETE from template
position_x: positionX,    // DELETE from template
position_y: positionY,    // DELETE from template

// These should only appear in laid-out image edit drawer
```

**Time:** 20 minutes

**Status:** ⧗ Needs implementation

---

## Fix 6: Always Show "Create Laid-Out Image" Section

### Problem
Create form is hidden behind button, extra click needed

### Solution

**File:** `web/src/views/tabs/ImageLayoutsTab.tsx`

**Remove toggle, always show form:**

```typescript
// DELETE this state:
const [isCreatingLaidOut, setIsCreatingLaidOut] = useState(false);

// DELETE this button:
<Button
  onClick={() => setIsCreatingLaidOut(!isCreatingLaidOut)}
  variant={isCreatingLaidOut ? 'secondary' : 'primary'}
>
  {isCreatingLaidOut ? 'Cancel' : '+ Create Laid-Out Image'}
</Button>

// ALWAYS render the create form:
<Card className="mb-6">
  <CardHeader>
    <h3 className="text-lg font-semibold text-gray-900">
      Create Laid-Out Image
    </h3>
  </CardHeader>
  <CardBody>
    <form onSubmit={handleCreate} className="grid md:grid-cols-3 gap-4">
      {/* Form always visible */}
    </form>
  </CardBody>
</Card>
```

**Time:** 10 minutes

**Status:** ⧗ Needs implementation

---

## Fix 7: Edit Drawer Doesn't Work

### Problem
Clicking "Edit" on laid-out image shows alert instead of drawer

### Solution

**File:** `web/src/views/tabs/ImageLayoutsTab.tsx`

**Implement edit drawer with override controls:**

```typescript
// Add state for edit drawer
const [editingLaidOutId, setEditingLaidOutId] = useState<string | null>(null);
const [editUserScale, setEditUserScale] = useState(1);
const [editPositionX, setEditPositionX] = useState(0);
const [editPositionY, setEditPositionY] = useState(0);

// Edit handler
const handleEditLaidOut = (image: LaidOutImage) => {
  setEditingLaidOutId(image.id);
  const overrides = image.overrides as Partial<ImageLayoutViewportSettings>;
  setEditUserScale(overrides?.user_scale ?? 1);
  setEditPositionX(overrides?.position_x ?? 0);
  setEditPositionY(overrides?.position_y ?? 0);
};

// Update save handler
const handleSaveEdit = async () => {
  if (!editingLaidOutId) return;
  await updateLaidOutImage({
    id: editingLaidOutId,
    overrides: {
      user_scale: editUserScale,
      position_x: editPositionX,
      position_y: editPositionY,
    },
  }).unwrap();
  setEditingLaidOutId(null);
};

// Render drawer (at bottom of component)
{editingLaidOutId && (
  <Card className="fixed bottom-0 left-0 right-0 max-h-[50vh] overflow-y-auto shadow-2xl border-t-4 border-primary-500">
    <CardHeader>
      <div className="flex items-center justify-between">
        <h3 className="text-xl font-semibold">Edit Laid-Out Image</h3>
        <Button variant="secondary" size="sm" onClick={() => setEditingLaidOutId(null)}>
          ✕ Close
        </Button>
      </div>
    </CardHeader>
    <CardBody>
      <div className="grid lg:grid-cols-2 gap-6">
        {/* Left: Preview */}
        <div className="bg-gray-50 p-4 rounded">
          {/* Show preview with current overrides */}
        </div>
        
        {/* Right: Override Controls */}
        <div className="space-y-4">
          <SliderInput
            label="User Scale"
            value={editUserScale}
            onChange={setEditUserScale}
            min={0.5}
            max={2}
            step={0.05}
            unit="×"
          />
          <SliderInput
            label="Position X"
            value={editPositionX}
            onChange={setEditPositionX}
            min={-1}
            max={1}
            step={0.01}
            unit="norm"
          />
          <SliderInput
            label="Position Y"
            value={editPositionY}
            onChange={setEditPositionY}
            min={-1}
            max={1}
            step={0.01}
            unit="norm"
          />
          
          <div className="flex justify-end space-x-3 pt-4">
            <Button variant="secondary" onClick={() => setEditingLaidOutId(null)}>
              Cancel
            </Button>
            <Button onClick={handleSaveEdit}>
              Save Changes
            </Button>
          </div>
        </div>
      </div>
    </CardBody>
  </Card>
)}
```

**Time:** 2 hours

**Status:** ⧗ Needs implementation

---

## Fix 8: Preview Thumbnail Incorrect

### Problem
Grid thumbnails don't show the actual cropped/scaled result

### Solution

**Two approaches:**

**Option A: Server-side rendering (better)**
```typescript
// When backend render is ready, load actual rendered image
<img src={`/api/laid-out-images/${image.id}/preview-thumbnail`} />
```

**Option B: Client-side canvas rendering (immediate)**
```typescript
// Use computation result to render on canvas
useEffect(() => {
  if (!image.result) return;
  const canvas = canvasRef.current;
  const ctx = canvas.getContext('2d');
  const result = image.result as ImageLayoutComputation;
  
  // Draw using source_rect and target_rect
  const img = new Image();
  img.onload = () => {
    ctx.drawImage(
      img,
      result.result.source_rect.x,
      result.result.source_rect.y,
      result.result.source_rect.w,
      result.result.source_rect.h,
      0, 0,
      canvas.width,
      canvas.height
    );
  };
  img.src = assetURL;
}, [image]);
```

**Recommendation:** Use Option B now, switch to Option A when renderer ready

**Time:** 1.5 hours

**Status:** ⧗ Needs implementation

---

## Fix 9: Preselect Latest Applied Template

### Problem
When creating new laid-out image, have to select template every time

### Solution

**File:** `web/src/views/tabs/ImageLayoutsTab.tsx`

**Remember last used template:**

```typescript
// Add state for last used template
const [lastUsedTemplateId, setLastUsedTemplateId] = useState<string>('');

// Auto-select when available
useEffect(() => {
  if (!createTemplateId && lastUsedTemplateId && templates.length > 0) {
    const exists = templates.find(t => t.id === lastUsedTemplateId);
    if (exists) {
      setCreateTemplateId(lastUsedTemplateId);
    } else if (templates.length > 0) {
      // Fallback to first template
      setCreateTemplateId(templates[0]!.id);
    }
  }
}, [templates, lastUsedTemplateId, createTemplateId]);

// Save on successful create
const handleCreate = async () => {
  // ... existing create logic
  await createLaidOutImage({...}).unwrap();
  setLastUsedTemplateId(createTemplateId);  // Remember this template
  // ...
};
```

**Or simpler: just default to first template in list:**

```typescript
useEffect(() => {
  if (!createTemplateId && templates.length > 0) {
    setCreateTemplateId(templates[0]!.id);
  }
}, [templates, createTemplateId]);
```

**Time:** 15 minutes

**Status:** ⧗ Needs implementation

---

## Implementation Priority

### ✅ Already Fixed

1. **Fix 1:** View assets in sequences ✅ DONE
   - Added 3-column layout
   - Assets list shows all available images
   - Click to add to sequence

### High Priority (Do Next - 40 minutes)

2. **Fix 6:** Always show create form (10 min)
3. **Fix 9:** Preselect template (15 min)
4. **Fix 2:** Landscape/portrait toggle (15 min)

**Total:** 40 minutes

### Medium Priority (Do Soon - 3 hours)

5. **Fix 4 & 5:** Clean up template/override separation (50 min)
6. **Fix 7:** Implement edit drawer (2 hours)

**Total:** 2 hours 50 minutes

### Low Priority (Do Later)

7. **Fix 8:** Better preview thumbnails (1.5 hours)
8. **Fix 3:** Backend rendering integration (Phase 3B)

---

## Quick Wins Implementation

**File: `web/src/views/tabs/ImageLayoutsTab.tsx`**

### Change 1: Remove Create Button Toggle

```typescript
// Line ~780-789: DELETE this button and state
// const [isCreatingLaidOut, setIsCreatingLaidOut] = useState(false);

// Just always render the create form
```

### Change 2: Auto-select First Template

```typescript
// After templates load, auto-select first one
useEffect(() => {
  if (templates.length > 0 && !createTemplateId) {
    setCreateTemplateId(templates[0].id);
  }
}, [templates]);
```

### Change 3: Fix Orientation Toggle

```typescript
// Add handler that swaps dimensions
const handleOrientationChange = (newOr: 'portrait' | 'landscape') => {
  if (newOr === orientation) return;
  // Swap width and height
  const temp = paperWidth;
  setPaperWidth(paperHeight);
  setPaperHeight(temp);
  setOrientation(newOr);
};

// Update button onClick
onClick={() => handleOrientationChange('portrait')}
onClick={() => handleOrientationChange('landscape')}
```

---

## Testing After Fixes

```bash
cd zine-layout/web
npm run typecheck
npm run build

# Manual testing:
# 1. Navigate to Image Layouts tab
# 2. Verify create form is always visible ✓
# 3. Verify template is pre-selected ✓
# 4. Click Portrait → Landscape, verify dimensions swap ✓
# 5. Create laid-out image ✓
```

---

## Summary

**Quick fixes (40 min):**
- Always show create form
- Auto-select template
- Fix orientation toggle

**Medium fixes (2h 50min):**
- Clean up template vs override fields
- Add edit drawer

**Later:**
- Better thumbnails
- Backend preview integration

**All fixable within 4 hours total!**

---

**END OF FIXES**

