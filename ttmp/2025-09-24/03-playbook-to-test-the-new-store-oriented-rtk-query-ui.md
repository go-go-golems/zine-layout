# Playbook: Testing the SQL-backed RTK Query UI

The goal is to confirm that the React app persists through the new SQLite-backed REST API for assets, pages, and spreads. Use this Q&A flow while running the server (`tmux attach -t zine_rest`) and the web dev server.

## Project + Asset Lifecycle

**Q: How do I verify existing projects seed correctly into SQLite?**
A: Open the project list view. On first load the UI calls `GET /api/projects`; watch `/tmp/zine-serve.log` for `upsertProjectRecord` debug lines. Optionally query the database with `sqlite3 data/zine-layout.db "SELECT id FROM projects;"` in another terminal.

**Q: How can I test image upload with SQL persistence?**
A: Inside a project, upload two PNGs via the carousel. The UI sends `POST /api/projects/{id}/images`; the response should show both in order. Confirm by refreshing—ordering should survive. For backend validation, run `sqlite3 data/zine-layout.db "SELECT id, sort_index FROM assets WHERE project_id='…';"`.

**Q: How do I confirm drag-and-drop reorder hits the new endpoint?**
A: Reorder the images. The UI issues `POST /api/projects/{id}/images/reorder`. After a refresh, the carousel order should match. Check logs for the reorder handler and verify the `sort_index` values using the SQL query above.

**Q: How do I know the new project asset sidebar is synced with Redux state?**
A: Select a thumbnail in the “Project Assets” panel and watch the preview + controls update instantly. Network panel should stay idle (selection is client-side), while the Redux devtools should show a `bookSpread/setImage` action populated with the asset dimensions and `uploadedPath`.

**Q: How do I validate uploads through `ProjectAssetsPanel` trigger previews?**
A: Drop a PNG onto the upload zone. The UI issues `POST /api/projects/{id}/images` and immediately selects the latest image. Confirm the preview renders the new asset and that Redux contains the new `uploadedPath`. SQL check: `SELECT COUNT(*) FROM assets WHERE project_id='…'` should increment.

## Page Editing

**Q: What steps prove page edits persist?**
A: Select a page, change layout settings (e.g., margins), and save/apply. This should call `PUT /api/projects/{id}/pages/{pageNumber}`. Reload the browser or open a new tab—the settings should repopulate. Backend check: `sqlite3 … "SELECT settings_json FROM pages WHERE project_id='…' AND page_number=1;"`.

**Q: How can I test deleting a page?**
A: Use the UI delete action. It sends `DELETE /api/projects/{id}/pages/{pageNumber}`. Confirm the page disappears without a refresh and that reloading the project keeps it gone. Validate with `sqlite3 … "SELECT COUNT(*) FROM pages WHERE project_id='…';"`.

**Q: How do I confirm loading a stored page back into the designer works?**
A: Save a page layout via the “Persisted Layouts → Pages → Save” button, clear some controls (e.g., reset margins), then click “Load”. Redux should receive `bookSpread/loadSettings`, the panels should jump back to the saved values, and the selected asset should update if `asset_id` was stored.

## Spread Editing

**Q: How do I ensure spreads reference pages correctly?**
A: Create a spread linking left/right page numbers and save. Watch for `PUT /api/projects/{id}/spreads/{spreadNumber}` in network tools. Refresh; associations should persist. Check SQL: `SELECT left_page_number, right_page_number FROM spreads …;`.

**Q: How can I test removing a spread?**
A: Trigger the UI delete. Expect `DELETE /api/projects/{id}/spreads/{spreadNumber}`. After a reload, confirm it’s gone. SQL check: `SELECT COUNT(*) FROM spreads …;`.

**Q: How do I validate spread loads set the page selectors?**
A: Save a spread with left/right page assignments. Change the numeric inputs and click “Load”. The inputs should reset to the persisted numbers and Redux settings should match the saved spread configuration.

## Preview & Render Integration

**Q: What indicates previews respect persisted settings?**
A: After saving a page or spread, click the preview button. The UI should use the stored settings (or fetch them first) before posting to `/api/v1/preview`. Compare the preview image with expected layout, and ensure logs show matching JSON payloads.

**Q: How do I verify Book YAML export still works with SQL data?**
A: Use the export action. The UI calls `GET /api/projects/{id}/yaml/book`. Inspect the YAML: it should include spreads you configured via SQL-backed forms. For deeper validation, run the CLI `go run ./cmd/zine-layout render --spec <downloaded.yaml> --output-dir /tmp/out` and inspect the generated files.

**Q: How can I ensure previews reflect the selected project asset?**
A: Toggle between thumbnails in the Project Assets panel. Each selection should dispatch `bookSpread/setImage`, update the preview request payload (check DevTools Network for `/api/v1/preview`), and regenerate the preview image without a page reload.

## Failure Handling

**Q: How does the UI behave if the DB entry is missing?**
A: Manually delete a row (e.g., `DELETE FROM pages …`) while the server runs. Reload the project—Pages panel should fall back gracefully (empty list), and logs should show 404s only on stale requests. Use this to confirm error toasts or retries appear.

**Q: What happens if the server rejects a save (e.g., unplug SQLite or send bad JSON)?**
A: Temporarily chmod the DB to readonly or force a body validation error (remove required fields via DevTools). Attempt to save in the persistence panel—the UI should surface the failure message and the Redux state should remain unchanged.

## Cleanup

**Q: How do I reset the environment between runs?**
A: Stop the tmux session (`tmux kill-session -t zine_rest`), remove `data/zine-layout.db*`, restart the server, and refresh the UI. Projects will reseed from disk, letting you rerun the playbook.
