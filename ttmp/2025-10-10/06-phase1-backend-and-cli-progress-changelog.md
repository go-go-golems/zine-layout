# Changelog – Phase 1 Backend & CLI Progress

## What I Did
- Replaced the legacy pages/spreads schema with fresh `projects`, `assets`, `image_sequences`, and `image_sequence_items` tables (`pkg/repo/sqlite/migrations.go`).
- Updated repository interfaces and implementations to match the new entities, including ID generation and sequence item helpers.
- Simplified the filesystem helper layer to just handle image persistence and dimensions.
- Rebuilt the HTTP server to operate solely on the SQL repositories and introduced REST routes for projects, assets, and image sequences/items.
- Reorganised the CLI under an `image-sequences` verb group with individual subcommands (`list`, `get`, `create`, `update`, `delete`, `add-item`, `reorder`, `delete-item`).

## What Worked
- Database migrations run cleanly against an empty DB and align with the new repository methods.
- `go build ./...` passes using the rewritten server, repositories, and CLI layout.
- The new CLI command structure mirrors the verb grouping the spec called for, making registration and future extensions straightforward.
- Manual smoke runs with `dist/zine-layout api …` confirmed project creation, image upload, sequence CRUD, item ordering, and deletion work end-to-end on a fresh SQLite database.
- Direct REST calls with `curl` verified the `/api/projects`, `/api/projects/{id}/images`, and `/api/image-sequences/*` routes behave as expected.

## What Didn’t Work (Yet)
- No automated CLI smoke test script exists for the new workflow (`cmd/zine-layout/cmds/api/phase1_test.sh` is still a TODO).
- The frontend and RTK Query layers remain untouched and still target the old endpoints.
- There is no API-level test suite; current confidence comes from ad-hoc `curl` sessions.

## What I Learned
- Removing the legacy manifest layer simplified the server substantially, but required careful alignment of helper utilities (e.g., keeping `projects.SavePNGImage` purely for disk I/O).
- Grouping CLI verbs under a dedicated namespace reduces churn when adding new subcommands—factory registration becomes obvious and discoverability improves.
- Dropping old tables from the migration upfront avoids confusion when iterating in a clean-slate schema, but requires discipline to recreate data during testing.

## Attention Points For Next Steps
- Add scripted CLI validation that spins up the server, exercises project/asset/sequence flows, and tears down temporary data.
- Verify REST responses with curl/httpie (or Go tests) once CLI coverage is in place.
- Ensure the frontend contracts are updated or clearly flagged as broken before moving to Phase 2.
- Consider seeding helper fixtures (e.g., sample PNGs) to accelerate manual testing while the UI lags behind.
