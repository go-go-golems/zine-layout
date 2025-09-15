## How to Write React + RTK Query Components for the Zine Layout Web UI

This guide documents the patterns we use to build React components and pages for the Zine Layout web UI. It covers project structure, RTK Query usage, routing, forms (uploads), basic state, and integration with the Go API. The goal is to enable fast iteration while keeping code consistent and easy to extend.

### 1) Project structure and key files

- `web/src/api.ts` — RTK Query API slice. Define endpoints here (projects, images, later presets, YAML, render). Export typed hooks for components.
- `web/src/store.ts` — Redux store configuration. Integrates `api.reducer` and `api.middleware`. Add local slices here as needed.
- `web/src/routes/App.tsx` — Top-level router, header nav, and route definitions.
- `web/src/views/Projects.tsx` — Projects list and create/delete.
- `web/src/views/ProjectDetail.tsx` — ImageTray (upload/list/reorder/delete) for a single project.

Routing is handled with `react-router-dom` using `<BrowserRouter>`. The Go server provides an SPA fallback that serves `index.html` for non-API routes, so visiting `/projects` or `/projects/:id` directly works.

### 2) RTK Query patterns (api.ts)

RTK Query centralizes API definitions and generates hooks for use in components.

- Keep endpoint names consistent with server routes.
- Group endpoints by domain: `projects`, `images`, `presets`, `yaml`, `render`.
- Use tags to manage cache invalidation: e.g., `Project`, `Image`.
- Export typed hooks (e.g., `useGetProjectsQuery`) and use them in components.

Example (snippets from `web/src/api.ts`):

```ts
export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Project', 'Image'],
  endpoints: (b) => ({
    getProjects: b.query<{ projects: Project[] }, void>({
      query: () => '/projects',
      providesTags: ['Project']
    }),
    createProject: b.mutation<{ project: Project }, { name?: string }>({
      query: (body) => ({ url: '/projects', method: 'POST', body }),
      invalidatesTags: ['Project']
    }),
    deleteProject: b.mutation<{ ok: boolean }, { id: string }>({
      query: ({ id }) => ({ url: `/projects/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Project']
    }),
    getImages: b.query<{ images: ImageItem[]; order: string[] }, { id: string }>({
      query: ({ id }) => `/projects/${id}/images`,
      providesTags: ['Image']
    }),
    uploadImages: b.mutation<{ images: ImageItem[] }, { id: string; files: FileList | File[] }>({
      query: ({ id, files }) => {
        const fd = new FormData()
        for (const f of Array.from(files as FileList)) fd.append('images[]', f)
        return { url: `/projects/${id}/images`, method: 'POST', body: fd }
      },
      invalidatesTags: ['Image']
    }),
    deleteImage: b.mutation<{ ok: boolean }, { id: string; imageId: string }>({
      query: ({ id, imageId }) => ({ url: `/projects/${id}/images/${encodeURIComponent(imageId)}`, method: 'DELETE' }),
      invalidatesTags: ['Image']
    }),
    reorderImages: b.mutation<{ ok: boolean }, { id: string; order: string[] }>({
      query: ({ id, order }) => ({ url: `/projects/${id}/images/reorder`, method: 'POST', body: { order } }),
      invalidatesTags: ['Image']
    }),
  })
})
```

### 3) Store configuration (store.ts)

Integrate `api.reducer` and `api.middleware` and add local slices as needed.

```ts
import { configureStore, createSlice } from '@reduxjs/toolkit'
import { api } from './api'

const uiSlice = createSlice({
  name: 'ui',
  initialState: { toasts: [] as { id: string; text: string; type?: 'info'|'error' }[] },
  reducers: {
    addToast: (s, a) => { s.toasts.push(a.payload) },
    removeToast: (s, a) => { s.toasts = s.toasts.filter(t => t.id !== a.payload) }
  }
})

export const store = configureStore({
  reducer: {
    ui: uiSlice.reducer,
    [api.reducerPath]: api.reducer
  },
  middleware: (gDM) => gDM().concat(api.middleware)
})
```

### 4) Routing and navigation (routes/App.tsx)

- Use `<Link>` from `react-router-dom` for in-app navigation (prevents full reloads).
- Define routes for pages and keep the header simple.
- Add `/projects` and `/projects/:id` routes.

```tsx
<BrowserRouter>
  <header>
    <Link to="/">Home</Link>
    <Link to="/projects">Projects</Link>
  </header>
  <Routes>
    <Route path="/" element={<Home />} />
    <Route path="/projects" element={<Projects />} />
    <Route path="/projects/:id" element={<ProjectDetail />} />
  </Routes>
</BrowserRouter>
```

Server-side: the Go server implements an SPA fallback that serves `index.html` for non-API paths; this enables direct navigation to client routes.

### 5) Projects list page (views/Projects.tsx)

This component demonstrates a list+create+delete CRUD pattern.

- Fetch: `useGetProjectsQuery()` renders a table of projects
- Create: form posts with `useCreateProjectMutation()` then refetch
- Delete: `useDeleteProjectMutation()` then refetch
- Link to detail: `<Link to={/projects/${id}}>`

Key principles:

- Keep local form state with `useState`.
- Prefer refetch over manual cache edits until patterns stabilize.
- Handle loading/error states (basic for now; toast later).

### 6) Project detail: ImageTray (views/ProjectDetail.tsx)

Shows how to implement file uploads, list items, and reorder/delete using RTK Query hooks.

Patterns applied:

- Upload via `<input type="file" multiple accept="image/png">`, building `FormData` for `images[]` fields.
- Drag-and-drop upload via a dropzone that accepts PNGs.
- Maintain a temporary `order` state for reordering (Up/Down buttons and drag-and-drop), then persist via `reorderImages` mutation.
- After mutations, `refetch()` to refresh consistent state from server. Later we can optimize with tag invalidation only.
- Use small, focused presentational subcomponents (`ImgCell`) to render thumbnails and metadata.

Basic drag-and-drop reordering pattern:

```tsx
const [order, setOrder] = useState<string[] | null>(null)
const [dragIndex, setDragIndex] = useState<number | null>(null)

const onItemDragStart = (idx: number) => (e: React.DragEvent) => {
  setDragIndex(idx)
  e.dataTransfer.effectAllowed = 'move'
}
const onItemDragOver = (e: React.DragEvent) => {
  e.preventDefault()
  e.dataTransfer.dropEffect = 'move'
}
const onItemDrop = (targetIndex: number) => (e: React.DragEvent) => {
  e.preventDefault()
  if (dragIndex === null || dragIndex === targetIndex) return
  const next = [...(order ?? data.order)]
  const [moved] = next.splice(dragIndex, 1)
  next.splice(targetIndex, 0, moved)
  setOrder(next)
  setDragIndex(null)
}

// usage on each image container:
<div draggable onDragStart={onItemDragStart(i)} onDragOver={onItemDragOver} onDrop={onItemDrop(i)}>
  ...
</div>
```

### 7) Component style and conventions

- Keep components small and single-purpose.
- Co-locate minor helper components with their parent (e.g., `ImgCell`).
- Prefer hooks-based logic; avoid class components.
- Use semantic HTML where possible (forms, buttons, tables).
- Defer complex styling until UX stabilizes; keep inline styles or small CSS modules.

### 8) Adding a new page or feature

Steps you can follow:

1. Server API (Go): implement route(s) in `serve.go` under `/api/...`, return JSON with minimal, stable types.
2. RTK Query: add endpoints in `web/src/api.ts`, exporting the generated hooks.
3. Store: ensure the `api` slice is integrated (already in `store.ts`). Add any small UI slices if needed.
4. Component: create a new `views/<Feature>.tsx` using the hooks. Handle loading and errors.
5. Routing: add a route in `routes/App.tsx` and a `<Link>` in the header or parent page.
6. Build + test: `make serve` (or `make web-build` then run `go run ...`).

### 9) Common pitfalls and tips

- Use `<Link>` rather than `<a>` to preserve SPA navigation and state.
- Make sure the Go server's SPA fallback is in place; otherwise, direct loads to nested routes 404.
- For uploads, confirm server constraints (currently PNG only). Validate early in the UI if needed.
- When in doubt, prefer refetch after mutations; optimize later with cache updates.
- Keep endpoints consistent and predictable; model server types in TypeScript.

### 10) Roadmap for more components/pages

- Presets browser page: list built-in presets and load YAML.
- YAML editor page: GET/PUT YAML with validation; show computed spec from UI.
- Render page: trigger render, list outputs, show previews, and allow ZIP download.
- Grid editor: drag-and-drop images into cells, rotate, and edit margins.

Each new page should follow the steps in section 8: server route → RTK endpoint → component → route.
