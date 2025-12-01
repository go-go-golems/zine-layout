---
Title: Frontend Developer Guide
Slug: frontend-developer-guide
Short: A friendly guide to working with React, RTK Query, and Tailwind in the Zine Layout web app.
Topics:
- react
- web
- rtk-query
- tailwind
- dnd
- docs
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Frontend Developer Guide

## 1. Welcome to the Zine Layout Web Frontend

This guide will help you get productive quickly when working on the Zine Layout web application. Whether you're adding a new feature, fixing a bug, or just trying to understand how things work, this document covers the essential patterns, conventions, and "gotchas" you'll encounter.

The frontend is built with React, TypeScript, and RTK Query, and follows a philosophy of **fast, optimistic UI updates** combined with **reliable error handling**. You'll see this pattern throughout the codebase, especially in the sequencing editor where drag-and-drop operations feel instant because we update the UI immediately and sync with the server in the background.

## 2. Tech Stack and Why We Chose It

Understanding the "why" behind our technology choices helps you work with the grain of the system:

**React 18** — We use functional components and hooks exclusively. The codebase is fully modern; you won't find class components here. Hooks like `useMemo`, `useCallback`, and `useEffect` are used extensively to optimize rendering and handle side effects.

**TypeScript** — Strong typing catches bugs before they reach production. All components have typed props (see `interface ComponentNameProps`), and RTK Query generates types for API responses automatically. When you see `type ImageSequenceItem` or `interface Asset`, those types come from `src/api.ts` and match the backend's JSON structure.

**Redux Toolkit Query (RTK Query)** — This handles all data fetching, caching, and state management. It's not just a fetch wrapper—it provides optimistic updates, automatic cache invalidation, and loading/error states. You'll work with hooks like `useGetImageSequencesQuery` and `useReorderImageSequenceItemsMutation` throughout the app.

**React Router** — Standard routing library. Routes are defined in `src/routes/App.tsx`. We use URL parameters for IDs (e.g., `/projects/:projectId/sequencing`) and query params for tabs (e.g., `?tab=sequences`).

**Tailwind CSS v4** — Utility-first CSS framework. We compose styles directly in JSX using className strings. Note that v4 changed the opacity syntax from `bg-opacity-50` to `bg-black/50` (slash notation).

**Vite** — Lightning-fast dev server with hot module replacement (HMR). The config lives at `zine-layout/web/vite.config.ts` and includes a proxy that forwards `/api` requests to the Go backend running on port 8088.

## 3. Project Structure (Where Everything Lives)

Here's where to find things when you're navigating the codebase:

```
zine-layout/web/src/
├── routes/
│   └── App.tsx                    # React Router setup, app shell with header/nav
├── api.ts                         # RTK Query API: all endpoints and hooks defined here
├── store.ts                       # Redux store config + uiSlice for toasts
├── components/
│   └── ui/                        # Shared primitives: Button, Card, Input, Tabs
│       ├── Button.tsx             # Reusable button with variants (primary/secondary/danger)
│       ├── Card.tsx               # Card wrapper, CardHeader, CardBody, CardFooter
│       ├── Input.tsx              # Form input with error states
│       └── Tabs.tsx               # Tab navigation component
├── views/
│   ├── ProjectDetail.tsx          # Project detail page with tabs
│   ├── tabs/
│   │   ├── SequencesTab.tsx       # Legacy sequencing UI (still used)
│   │   └── SequencesTabWrapper.tsx # Wrapper with toggle between v2/legacy
│   └── v2/                        # New sequencing implementation (isolated)
│       ├── SequencingPage.tsx     # Main sequencing page
│       └── components/
│           ├── SequenceList.tsx   # Sidebar list of sequences
│           ├── SequenceEditor.tsx # Main editor with drag-drop grid
│           ├── SequenceItem.tsx   # Individual item (image or gap)
│           ├── AssetPicker.tsx    # Modal for selecting images
│           └── SequenceSlideshow.tsx # Fullscreen preview (single/spread)
└── index.css                      # Tailwind imports + custom CSS variables
```

**Why the `v2/` directory?** We're building the new sequencing UI alongside the old one without breaking existing functionality. This lets us iterate quickly while keeping a stable fallback. The wrapper (`SequencesTabWrapper.tsx`) provides a toggle button to switch between versions.

## 4. Getting Started Locally

**Starting the backend:**
```bash
cd zine-layout
make serve-tmux  # Starts Go server in tmux on port 8088
# OR
go run ./cmd/zine-layout serve --addr :8088 --root ./web/dist --data-root ./data
```

**Starting the web dev server:**
```bash
cd zine-layout/web
pnpm install      # First time only
pnpm dev          # Starts Vite on http://localhost:5173
```

**Vite proxy configuration** (`web/vite.config.ts`):
The dev server proxies `/api/*` requests to `http://localhost:8088` (the Go backend). If your backend runs on a different port, create a `.env` file:
```bash
# zine-layout/web/.env
VITE_API_PROXY=http://localhost:8088
```

**Common startup issue:** If you see "500 Internal Server Error" on API calls, the proxy might be pointing to the wrong port. Check `vite.config.ts` line 6 and ensure it matches your backend port.

## 5. Working with Data: RTK Query Patterns

RTK Query is the heart of our data layer. All API interactions happen through it, and it provides powerful features like automatic caching, optimistic updates, and error handling.

### Where API Definitions Live

Open `src/api.ts` to see all endpoints. Here's the structure:

```typescript
// src/api.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

const baseQuery = fetchBaseQuery({ baseUrl: '/api' });

export const api = createApi({
  baseQuery,
  tagTypes: ['Projects', 'Assets', 'ImageSequences', 'ImageSequenceItems'],
  endpoints: (builder) => ({
    // Queries (GET requests)
    getProjects: builder.query<{ projects: Project[] }, void>({...}),
    getImageSequenceDetail: builder.query<SequenceDetail, { sequenceId: string }>({...}),
    
    // Mutations (POST/PUT/DELETE requests)
    createImageSequence: builder.mutation<ImageSequence, CreateSequenceRequest>({...}),
    reorderImageSequenceItems: builder.mutation<...>({...}),
  }),
});

// Export hooks for use in components
export const {
  useGetProjectsQuery,
  useGetImageSequenceDetailQuery,
  useCreateImageSequenceMutation,
  useReorderImageSequenceItemsMutation,
  // ... many more
} = api;
```

### Using Queries in Components

Queries are for fetching data. They automatically handle loading, caching, and refetching:

```typescript
// In SequenceList.tsx
import { useGetImageSequencesQuery } from '../../../api';

export const SequenceList: React.FC<Props> = ({ projectId }) => {
  // Hook provides data, isLoading, error states
  const { data: sequences = [], isLoading, error } = useGetImageSequencesQuery(
    { projectId },
    { skip: !projectId }  // Don't fetch if no projectId
  );

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading sequences</div>;
  
  return (
    <ul>
      {sequences.map(seq => <li key={seq.id}>{seq.name}</li>)}
    </ul>
  );
};
```

**Why this pattern works:** The query hook automatically subscribes to cache updates. If another component invalidates the `ImageSequences` tag, this component automatically refetches. No manual state management needed.

### Using Mutations for Updates

Mutations are for creating, updating, or deleting data:

```typescript
// In SequenceEditor.tsx
import { useAddImageSequenceItemMutation } from '../../../api';

const [addItem, { isLoading }] = useAddImageSequenceItemMutation();

const handleAddGap = async () => {
  try {
    await addItem({ sequenceId, assetId: undefined }).unwrap();
    // Success! Cache is automatically updated
  } catch (error) {
    console.error('Failed to add gap:', error);
  }
};
```

### Optimistic Updates: The Secret to Snappy UX

This is where RTK Query shines. When a user drags an item to reorder the sequence, we don't want them to wait for the server response—that feels slow. Instead, we update the UI immediately (optimistically) and sync with the server in the background.

Here's the pattern from `src/api.ts` for `reorderImageSequenceItems`:

```typescript
reorderImageSequenceItems: builder.mutation<
  ImageSequenceItem[],
  { sequenceId: string; items: { assetId?: string; isGap: boolean }[] }
>({
  query: ({ sequenceId, items }) => ({
    url: `/image-sequences/${encodeURIComponent(sequenceId)}/items`,
    method: 'PUT',
    body: {
      items: items.map((item) => ({
        asset_id: item.assetId,
        is_gap: item.isGap ?? !item.assetId,
      })),
    },
  }),
  async onQueryStarted({ sequenceId, items }, { dispatch, queryFulfilled }) {
    // 1. Optimistically update the cache immediately
    const patchResult = dispatch(
      api.util.updateQueryData('getImageSequenceDetail', { sequenceId }, (draft) => {
        draft.items = items.map((item, idx) => ({
          sequence_id: sequenceId,
          position: idx,
          asset_id: item.assetId,
          is_gap: item.isGap ?? !item.assetId,
        }));
      })
    );
    
    try {
      // 2. Wait for server confirmation
      await queryFulfilled;
      // 3. If successful, server response automatically replaces optimistic update
    } catch (error) {
      // 4. If failed, undo the optimistic update
      patchResult.undo();
      // 5. Show error to user
      dispatch(uiSlice.actions.addToast({
        id: Date.now().toString(),
        text: 'Failed to reorder items. Please try again.',
        type: 'error',
      }));
    }
  },
  invalidatesTags: (_result, _error, { sequenceId }) => [
    { type: 'ImageSequenceItems', id: sequenceId },
  ],
}),
```

**Why this is powerful:** The user sees their change instantly. If the network is slow or fails, we automatically rollback and notify them. This pattern is used throughout the app for add, delete, and reorder operations.

## 6. Styling with Tailwind CSS v4

We use Tailwind's utility classes directly in JSX. This keeps styles co-located with components and makes it easy to build responsive, consistent UIs.

### Important: Tailwind v4 Changes

If you're used to Tailwind v3, note that **opacity utilities changed**:

**OLD (v3):** `bg-black bg-opacity-50`  
**NEW (v4):** `bg-black/50`

The slash notation is more concise and works for all color utilities:
- `bg-white/10` = white background at 10% opacity
- `text-gray-500/80` = gray text at 80% opacity
- `border-red-500/30` = red border at 30% opacity

**Example from SequenceSlideshow.tsx:**
```tsx
<div className="fixed inset-0 z-50 bg-black/50">  {/* 50% opacity black overlay */}
  <Button className="bg-white/10 hover:bg-white/20 text-white">
    Close
  </Button>
</div>
```

### Responsive Design

Tailwind's breakpoint prefixes make responsive layouts straightforward:

```tsx
{/* 2 columns on mobile, 3 on tablet, 4 on desktop, 5 on xl screens */}
<div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
  {items.map(item => <SequenceItem key={item.id} item={item} />)}
</div>
```

### UI Component Primitives

Our shared components live in `src/components/ui/`. They're intentionally minimal and forward DOM props:

**Button** (`components/ui/Button.tsx`):
```tsx
<Button onClick={handleClick} variant="primary" size="sm">
  Save
</Button>
// Variants: primary, secondary, danger
// Sizes: sm, default, lg
```

**Card** (`components/ui/Card.tsx`):
```tsx
<Card onClick={handleSelect}>  {/* onClick forwarded to div */}
  <CardHeader>Title</CardHeader>
  <CardBody>Content here</CardBody>
  <CardFooter>Actions</CardFooter>
</Card>
```

**Important:** Card components extend `React.HTMLAttributes<HTMLDivElement>`, so you can pass any div props like `onClick`, `onMouseEnter`, etc. This is crucial for interactive cards (see `SequenceList.tsx` where cards are clickable).

## 7. Routing and Navigation

Routes are defined in `src/routes/App.tsx`:

```typescript
<Routes>
  <Route path="/" element={<Home />} />
  <Route path="/projects" element={<Projects />} />
  <Route path="/projects/:id" element={<ProjectDetail />} />
  <Route path="/v2/projects/:projectId/sequencing" element={<SequencingPage />} />
</Routes>
```

**Adding a new route:**
1. Create your page component in `src/views/` or `src/views/v2/`
2. Import it in `App.tsx`
3. Add a `<Route>` element with your path and component

**Accessing route params:**
```typescript
import { useParams } from 'react-router-dom';

const MyPage = () => {
  const { projectId } = useParams<{ projectId: string }>();
  // Use projectId in queries, etc.
};
```

## 8. Drag-and-Drop Sequencing (The Core Pattern)

The sequence editor is the heart of the app, so understanding drag-and-drop is essential. We use native HTML5 drag-and-drop APIs.

### The Mental Model

When a user drags an item from position A to position B:
1. Remove the item from position A
2. Insert it at position B (adjusting for the removal)
3. Update all positions to be sequential (0, 1, 2, ...)
4. Show this change immediately in the UI
5. Send the new order to the server in the background

### Implementation in SequenceEditor.tsx

```typescript
// State to track drag operation
const [dragSourceIndex, setDragSourceIndex] = useState<number | null>(null);
const [dragTargetIndex, setDragTargetIndex] = useState<number | null>(null);

// Handlers
const handleDragStart = (index: number) => {
  setDragSourceIndex(index);
};

const handleDragEnter = (index: number) => {
  if (dragSourceIndex !== null && dragSourceIndex !== index) {
    setDragTargetIndex(index);
  }
};

const handleDrop = (event: React.DragEvent, targetIndex: number | null) => {
  event.preventDefault();
  if (dragSourceIndex === null || targetIndex === null) return;
  
  // Calculate new order
  const newItems = [...sortedItems];
  const [removed] = newItems.splice(dragSourceIndex, 1);  // Remove from source
  newItems.splice(targetIndex, 0, removed);                // Insert at target
  
  // Update positions
  const reorderedItems = newItems.map((item, idx) => ({
    ...item,
    position: idx,
  }));
  
  // Trigger debounced API call
  debouncedReorder(reorderedItems);
};
```

**Critical insight:** When dragging forward (e.g., position 0 → position 3), after removing the source item, the indices shift down. The target index naturally points to the correct insertion spot without adjustment. This is why we just use `targetIndex` directly.

### Debouncing for Performance

If a user rapidly drags items, we don't want to send 10 API calls. Instead, we batch them:

```typescript
// SequenceEditor.tsx
const reorderRef = useRef<ReturnType<typeof debounce>>();
if (!reorderRef.current) {
  reorderRef.current = debounce((items: ImageSequenceItem[]) => {
    const payload = items.map((item) => ({
      assetId: item.asset_id ?? undefined,
      isGap: item.is_gap,
    }));
    reorderItems({ sequenceId, items: payload }).catch((err) => {
      console.error('Failed to reorder items:', err);
    });
  }, 300);  // 300ms debounce
}
```

**How it works:** Each drag calls `debouncedReorder()`, which resets a 300ms timer. Only when dragging stops for 300ms does the API call fire. This reduces server load and feels just as responsive to the user.

### Visual Feedback

Users need to see where they're dropping:

```tsx
{/* Drop indicator - blue pulsing line */}
{showDropIndicator && (
  <div
    className={`absolute top-0 bottom-0 w-1 bg-primary-500 rounded-full z-10 shadow-lg ${
      dropIndicatorOnRight ? '-right-2.5' : '-left-2.5'
    }`}
    style={{ animation: 'pulse 1s cubic-bezier(0.4, 0, 0.6, 1) infinite' }}
  />
)}

{/* Dragged item appears faded */}
<div className={isDragging ? 'opacity-30 scale-95' : ''}>
  <SequenceItem ... />
</div>
```

**For a deep dive on implementing drag-and-drop from scratch**, see the playbook:
```
Reference: vibes/.../reference/playbook-draggable-sequence-editor.md
```

## 9. Slideshow and Book-Spread Preview

The slideshow (`SequenceSlideshow.tsx`) supports two modes:

**Single-page mode:** Shows images one at a time, skipping gaps. This is the default "preview" view.

**Book-spread mode:** Shows pages in pairs (like an open book), with gaps rendered as blank pages. This lets users see how their zine will look when printed and folded.

### Mode Switching

```typescript
const [viewMode, setViewMode] = useState<'single' | 'spread'>('single');

// Single mode: filter out gaps
const singleSlides = items
  .filter((item) => !item.is_gap && item.asset_id)
  .map((item) => ({ item, asset: assets.get(item.asset_id) }));

// Spread mode: pair consecutive items
const spreadSlides = [];
for (let i = 0; i < items.length; i += 2) {
  spreadSlides.push({
    left: items[i] ? { item: items[i], asset: assets.get(items[i].asset_id) } : undefined,
    right: items[i + 1] ? { item: items[i + 1], asset: assets.get(items[i + 1].asset_id) } : undefined,
  });
}
```

### Fullscreen Support

We use the native Fullscreen API:

```typescript
const enterFullscreen = useCallback(async () => {
  const element = document.documentElement;
  try {
    if (element.requestFullscreen) {
      await element.requestFullscreen();
    }
  } catch (err) {
    console.error('Failed to enter fullscreen:', err);
  }
}, []);
```

**Keyboard shortcuts:**
- `←` / `→` : Navigate slides
- `F` : Toggle fullscreen
- `Esc` : Exit fullscreen or close slideshow

### Common Pitfall: Hook Ordering

When setting up keyboard navigation, define your callbacks **before** the `useEffect` that uses them:

```typescript
// ✅ GOOD: Define callbacks first
const goToPrevious = useCallback(() => { ... }, [deps]);
const goToNext = useCallback(() => { ... }, [deps]);

useEffect(() => {
  const handleKeyDown = (e) => {
    if (e.key === 'ArrowLeft') goToPrevious();  // ✅ Already defined
    if (e.key === 'ArrowRight') goToNext();
  };
  window.addEventListener('keydown', handleKeyDown);
  return () => window.removeEventListener('keydown', handleKeyDown);
}, [goToPrevious, goToNext]);

// ❌ BAD: useEffect references callbacks that don't exist yet
useEffect(() => {
  const handleKeyDown = (e) => {
    if (e.key === 'ArrowLeft') goToPrevious();  // ❌ ReferenceError!
  };
  // ...
}, []);

const goToPrevious = useCallback(() => { ... }, []);  // ❌ Too late
```

This is a "temporal dead zone" error and will crash the app. Always define callbacks before effects that reference them.

## 10. Error Handling and User Feedback

We use a toast notification system defined in `src/store.ts`:

```typescript
// store.ts
export const uiSlice = createSlice({
  name: 'ui',
  initialState: { toasts: [] },
  reducers: {
    addToast: (state, action) => {
      state.toasts.push(action.payload);
    },
    removeToast: (state, action) => {
      state.toasts = state.toasts.filter((t) => t.id !== action.payload);
    },
  },
});

export const { addToast, removeToast } = uiSlice.actions;
```

### Showing a Toast

```typescript
import { useDispatch } from 'react-redux';
import { uiSlice } from '../../store';

const dispatch = useDispatch();

// Error toast
dispatch(uiSlice.actions.addToast({
  id: Date.now().toString(),
  text: 'Failed to save changes. Please try again.',
  type: 'error',
}));

// Success toast
dispatch(uiSlice.actions.addToast({
  id: Date.now().toString(),
  text: 'Sequence saved successfully!',
  type: 'success',
}));
```

**Why toasts instead of alerts?** Toasts are non-blocking and don't interrupt the user's flow. They appear briefly and auto-dismiss, allowing the user to continue working.

### Error Boundaries (Future Enhancement)

For production-grade error handling, wrap risky components with React Error Boundaries. This prevents a crash in one part of the UI from breaking the entire app.

## 11. Performance Considerations

### Memoization

Use `useMemo` for expensive computations:

```typescript
// SequenceEditor.tsx
const sortedItems = useMemo(() => {
  if (!sequenceData?.items) return [];
  return [...sequenceData.items].sort((a, b) => a.position - b.position);
}, [sequenceData?.items]);
```

**When to memoize:** If a computation involves sorting, filtering, or transforming large arrays, and the input only changes occasionally, memoize it.

### Callback Stability

Use `useCallback` for callbacks passed to child components to prevent unnecessary re-renders:

```typescript
const handleDelete = useCallback((position: number) => {
  deleteItem({ sequenceId, position });
}, [sequenceId, deleteItem]);

// Pass to child
<SequenceItem onDelete={() => handleDelete(item.position)} />
```

### Virtualization (For Future)

If a sequence has hundreds of items, consider using `react-window` or `react-virtual` to only render visible items. This isn't implemented yet but would be the next optimization step.

## 12. Testing Checklist

When implementing or modifying features, verify:

**Drag-and-Drop:**
- [ ] Drag item forward (e.g., position 0 → 3) lands in correct spot
- [ ] Drag item backward (e.g., position 5 → 2) lands in correct spot
- [ ] Rapid dragging batches into single API call after debounce
- [ ] Visual feedback: drop indicator shows on correct side (left when dragging back, right when dragging forward)

**Optimistic Updates:**
- [ ] UI updates immediately on add/delete/reorder
- [ ] Network failure triggers rollback (UI returns to previous state)
- [ ] Error toast appears on failure

**Slideshow:**
- [ ] Single mode shows only images
- [ ] Spread mode shows pairs, including gaps as blank pages
- [ ] Keyboard navigation works (←, →, F, Esc)
- [ ] Fullscreen toggles correctly

**Responsive:**
- [ ] Grid adjusts columns at breakpoints (2/3/4/5 columns)
- [ ] Modal/slideshow works on mobile sizes

## 13. Code Style Guidelines

**TypeScript:**
- Always type component props: `interface MyComponentProps { ... }`
- Avoid `any`; use `unknown` if truly unknown, then narrow with type guards
- Export types needed by consumers

**Component Structure:**
```tsx
// 1. Imports
import React, { useState, useCallback } from 'react';
import { useGetDataQuery } from '../../api';

// 2. Types/Interfaces
interface MyComponentProps {
  projectId: string;
  onSave?: () => void;
}

// 3. Component
export const MyComponent: React.FC<MyComponentProps> = ({ projectId, onSave }) => {
  // 4. Hooks (top of component)
  const { data, isLoading } = useGetDataQuery({ projectId });
  const [isOpen, setIsOpen] = useState(false);
  
  // 5. Callbacks
  const handleSave = useCallback(() => {
    // ...
    onSave?.();
  }, [onSave]);
  
  // 6. Early returns
  if (isLoading) return <div>Loading...</div>;
  
  // 7. Main render
  return (
    <div>
      {/* ... */}
    </div>
  );
};
```

**Naming:**
- Components: `PascalCase` (e.g., `SequenceEditor`)
- Hooks: `useCamelCase` (e.g., `useDebounce`)
- Handlers: `handleCamelCase` (e.g., `handleDragStart`)
- Boolean props/state: `isSomething`, `hasSomething`, `showSomething`

**Avoid:**
- Deep nesting (max 3 levels if possible)
- Abbreviations (prefer `selectedSequenceId` over `selSeqId`)
- Comments that describe what code does (code should be self-documenting)
- Comments are for *why*, not *what*

## 14. Quick Reference: Common Tasks

### A) Add a New API Endpoint

**File:** `src/api.ts`

```typescript
// 1. Add to endpoints in api definition
getMyData: builder.query<MyDataResponse, { id: string }>({
  query: ({ id }) => `/my-endpoint/${encodeURIComponent(id)}`,
  providesTags: (result, error, { id }) => [{ type: 'MyData', id }],
}),

// 2. Export auto-generated hook (bottom of file)
export const { useGetMyDataQuery } = api;

// 3. Use in component
import { useGetMyDataQuery } from '../../api';

const { data, isLoading, error } = useGetMyDataQuery({ id: '123' });
```

### B) Add a New Page Route

**Files:** `src/views/MyNewPage.tsx`, `src/routes/App.tsx`

```typescript
// 1. Create page component
// src/views/MyNewPage.tsx
export const MyNewPage: React.FC = () => {
  return <div>My New Page</div>;
};

// 2. Add route in App.tsx
import { MyNewPage } from '../views/MyNewPage';

<Routes>
  <Route path="/my-page" element={<MyNewPage />} />
</Routes>

// 3. Link to it
<Link to="/my-page">Go to My Page</Link>
```

### C) Add a Modal Dialog

**Pattern:** Modal is a component that renders `null` when closed, and an overlay when open.

```typescript
// Parent component
const [isModalOpen, setIsModalOpen] = useState(false);

<Button onClick={() => setIsModalOpen(true)}>Open Modal</Button>
<MyModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} />

// MyModal.tsx
interface MyModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const MyModal: React.FC<MyModalProps> = ({ isOpen, onClose }) => {
  if (!isOpen) return null;
  
  return (
    <div 
      className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center"
      onClick={onClose}  // Click overlay to close
    >
      <Card onClick={(e) => e.stopPropagation()}>  {/* Prevent overlay close */}
        <CardHeader>Modal Title</CardHeader>
        <CardBody>Content here</CardBody>
        <CardFooter>
          <Button onClick={onClose}>Close</Button>
        </CardFooter>
      </Card>
    </div>
  );
};
```

### D) Add an Optimistic Mutation

**See full example in Section 5**, or check `src/api.ts` for `reorderImageSequenceItems`, `addImageSequenceItem`, or `deleteImageSequenceItem`.

Key steps:
1. Define mutation with `builder.mutation`
2. Add `onQueryStarted` handler
3. Use `api.util.updateQueryData` to update cache
4. Wrap in try/catch; call `patchResult.undo()` on error
5. Dispatch toast on error

### E) Fix "API calls returning 500"

**Problem:** Vite proxy not pointing to correct backend port.

**Solution:**
1. Check backend port: `lsof -ti:8088` (should show process)
2. Update `web/vite.config.ts` line 6: `const proxyTarget = env.VITE_API_PROXY ?? 'http://localhost:8088';`
3. Or create `web/.env` with `VITE_API_PROXY=http://localhost:8088`
4. Restart `pnpm dev`

### F) Debug "Cannot access X before initialization"

**Problem:** Trying to use a variable/function before it's defined.

**Solution:** In React components, this usually means a `useEffect` or event handler is referencing a callback that's defined later. Move callback definitions (`useCallback`, `const handleX = ...`) **before** any code that uses them.

## 15. Helpful Resources

**Within this project:**
- Sequencing UX walkthrough with ASCII diagrams and API patterns:  
  `vibes/.../reference/sequencing-ux-walkthrough.md`
  
- Debate on optimistic updates, batching, and visual feedback:  
  `vibes/.../reference/debate-round-16-sequencing-ux-api.md`
  
- Playbook for building draggable sequence editors (reusable guide):  
  `vibes/.../reference/playbook-draggable-sequence-editor.md`

**External docs:**
- [RTK Query Optimistic Updates](https://redux-toolkit.js.org/rtk-query/usage/optimistic-updates)
- [Tailwind CSS v4 Docs](https://tailwindcss.com/docs)
- [React Router v6](https://reactrouter.com/)

---

## Final Tips

1. **Read the existing code first.** `SequenceEditor.tsx` and `api.ts` are the best teachers.
2. **Start with small changes.** Get one feature working end-to-end before adding complexity.
3. **Use browser DevTools.** React DevTools shows component state; Network tab shows API calls.
4. **When in doubt, optimistic update it.** Users prefer instant feedback over waiting for the server.
5. **Keep components small.** If a component file is >300 lines, consider splitting it.

Happy coding! If you have questions or want to improve this guide, just update this file and share your knowledge with the team.
