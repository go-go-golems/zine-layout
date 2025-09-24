# How to Write React Code for the New Store-Oriented Designer

This guide is a hand-off for frontend engineers joining the project after the SQLite persistence work (see `ttmp/2025-09-24/04-session-log.md`). It explains how the React app talks to the new store layer, where the key files live, and what you need to touch when extending the Book Spread Designer or the project detail pages.

## Architectural context

- Backend persistence moved to SQLite via the repository layer documented in `ttmp/2025-09-24/01-analysis-and-architecture-on-how-to-add-proper-sql-storage-backend-to-the-zine-layout-project.md`.
- All image, page, and spread traffic goes through REST routes implemented in `pkg/serve/server.go` (look at `handleProjectImages` around line 174, `handleUpsertPage` ~344, and `handleUpsertSpread` ~522).
- The frontend consumes those endpoints through RTK Query definitions in `web/src/api.ts` (images around line 120, pages/spreads around line 210).
- State orchestration lives in the Redux slice `web/src/store/bookSpreadSlice.ts` which tracks the currently selected asset, paper settings, margins, etc.

## Core data flow in the designer

1. **Project selection** (`web/src/views/BookSpreadDesigner.tsx:43`): loads `/api/projects` and drives the rest of the screen through the chosen project id.
2. **Asset loading & selection**:
   - `useGetImagesQuery` fetches `{ images, order }`.
   - `ProjectAssetsPanel` (`web/src/components/bookSpread/ProjectAssetsPanel.tsx:10`) turns that into a clickable gallery and wraps `useUploadImagesMutation` for drag/drop or file input uploads.
   - Selecting an asset dispatches `setImage` from the slice so preview controls and the canvas share dimensions.
3. **Layout persistence**:
   - `LayoutPersistencePanel` (`web/src/components/bookSpread/LayoutPersistencePanel.tsx:20`) reads persisted pages/spreads via `useGetPagesQuery` / `useGetSpreadsQuery` and serializes Redux state with `buildSpreadSettingsFromState` before saving.
   - Loading a record calls `dispatch(loadSettings(persisted.settings))`, which rehydrates paper size, margins, crop ratio, etc.
4. **Preview generation**:
   - `usePreviewRequest` (defined in `web/src/utils/spreadRequestBuilder.ts`) builds the payload for `/api/v1/preview` using the current Redux state and selected asset.
   - Preview responses power the framed `<img>` elements rendered inside `renderFramedPreview` in the main view.
5. **YAML export** still calls `useLazyExportBookYamlQuery`, ensuring the designer output stays compatible with the CLI export pipeline.

## Key files and symbols

- `web/src/views/BookSpreadDesigner.tsx` — primary orchestrator; ties project selection, asset panel, settings panels, and persistence panel together.
- `web/src/components/bookSpread/ProjectAssetsPanel.tsx` — handles uploading PNGs and selecting from the SQL-backed asset list using `useUploadImagesMutation`.
- `web/src/components/bookSpread/LayoutPersistencePanel.tsx` — provides Save/Load/Delete for pages and spreads using `usePutPageMutation`, `usePutSpreadMutation`, etc.
- `web/src/store/bookSpreadSlice.ts` — Redux slice exposing actions like `setImage`, `loadSettings`, `setPaperSize`, and shape constants (`PAPER_SIZES`, `CROP_RATIOS`).
- `web/src/utils/spreadRequestBuilder.ts` — converts slice state into `SpreadSettings` before hitting compute or persistence endpoints.
- `web/src/components/ImageTray.tsx` — legacy image manager in the project detail view; still references the same RTK Query hooks but needs an upgrade to share the new asset UI.
- `ttmp/2025-09-24/03-playbook-to-test-the-new-store-oriented-rtk-query-ui.md` — testing walkthrough you should follow after any change touching persistence or previews.

## Extending the designer: workflow checklist

When you implement new features (e.g., storing additional metadata, surfacing existing pages, or wiring new buttons), use this sequence:

1. **Model the data in RTK Query**
   - Add or update endpoint definitions in `web/src/api.ts` and export the typed hooks.
   - Keep request/response shapes aligned with the handlers in `pkg/serve/server.go`.

2. **Update Redux state if needed**
   - Extend `BookSpreadState` for any new fields.
   - Provide reducer actions in `bookSpreadSlice.ts` so components remain declarative.

3. **Compose UI panels**
   - Add new components under `web/src/components/bookSpread/` to keep the view modular.
   - Pass computed props from `BookSpreadDesigner.tsx` rather than re-querying inside child components.

4. **Persist and hydrate**
   - Use `buildSpreadSettingsFromState` as a single source for saving settings; update it if the state shape changes.
   - Ensure `loadSettings` handles new fields so loading persisted rows repopulates the UI.

5. **Test**
   - Run `pnpm run typecheck` for TypeScript coverage.
   - Execute `go test ./...` to catch backend regressions.
   - Work through the playbook Q&A to confirm uploads, save/load cycles, and previews behave as expected.

## Upcoming tasks for the next engineer

- **Project detail parity**: Replace the bespoke `ImageTray` UI in `web/src/views/ProjectDetail.tsx` with `ProjectAssetsPanel` so uploads and ordering match the designer experience.
- **Page-first workflow**: Expose persisted pages/spreads inside the project view (e.g., list from `useGetPagesQuery`) to let users jump into the designer with a stored layout.
- **Error handling polish**: Surface mutations errors from RTK Query using toast notifications or inline states, especially for PUT/DELETE actions in `LayoutPersistencePanel`.
- **Asset metadata**: Decide how to show asset timestamps or file sizes; the API already returns `bytes` and `content_type` if you want to expose them.

## How to validate your changes

1. Start the server (see `make serve-tmux` or attach to `tmux attach -t zine_rest`).
2. Run the frontend dev server with `pnpm dev` from the `web/` directory.
3. Follow the Q&A steps in `ttmp/2025-09-24/03-playbook-to-test-the-new-store-oriented-rtk-query-ui.md`:
   - Upload assets, reorder, and verify via SQL queries if needed.
   - Save a page layout, reload the browser, and confirm `loadSettings` hydrates the state.
   - Save spreads and ensure left/right page bindings persist.
   - Export Book YAML and, optionally, run the CLI render command.
4. Check `/tmp/zine-serve.log` for DEBUG lines showing repository writes.
5. If you touch the RTK Query layer, inspect the Redux DevTools trace to ensure queries/mutations are tagged and cached correctly.

## Reference materials

- Backend deep dive: `pkg/repo/sqlite/sqlite.go`, `pkg/serve/server.go`.
- Frontend data access: `web/src/api.ts` (search for `usePutPageMutation`, `usePutSpreadMutation`).
- State helpers: `web/src/utils/` for preview request builders and dimension calculations.
- Session context: `ttmp/2025-09-24/04-session-log.md` for a narrative of everything implemented before this hand-off.

Keep this document close while you work; update it whenever you add a new store-backed feature so future teammates can ramp quickly.
