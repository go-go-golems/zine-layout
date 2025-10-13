# CLI Verbs Inventory and Mapping

Date: 2025-10-12

## 1. Purpose
- Capture every CLI verb exposed by `zine-layout` so backend contributors can script flows without chasing files.
- Map verbs to their Go implementations, HTTP routes, and services for the Stage A–C work tracked in `ttmp/2025-10-11/19-phase3-focused-roadmap.md`.
- Highlight which verbs already work offline (workflow commands) versus which still depend on the API, feeding the gaps noted in `ttmp/2025-10-11/21-end-to-end-cli-test-plan.md`.

## 2. Command Surface at a Glance
| Entry point | Description | CLI source | Backend touchpoints |
|-------------|-------------|------------|---------------------|
| `zine-layout serve` | Launch HTTP API + static frontend bundle. | `cmd/zine-layout/cmds/serve/command.go` | `pkg/serve/server.go`, `pkg/serve/*_routes.go` |
| `zine-layout render` | Render imposition YAML directly to PNG sheets. | `cmd/zine-layout/cmds/render.go` | `pkg/app`, `pkg/zinelayout`, `pkg/export/pdf.go` |
| `zine-layout imagelayout compute` | Run viewport math for a single asset/template pair. | `cmd/zine-layout/cmds/imagelayout/*.go` | `pkg/imagelayout/engine`, `pkg/imagelayout` |
| `zine-layout api ...` | REST client verbs that talk to a running server. | `cmd/zine-layout/cmds/api/**/*` | `pkg/serve/*`, `pkg/services/*` (via HTTP) |
| `zine-layout workflow ...` | Direct repository/service verbs (no server). | `cmd/zine-layout/cmds/workflow/**/*` | `pkg/repo`, `pkg/services` |

## 3. Common Mechanics
- Glazed defaults: pass `--output table|json|csv|yaml`, `--fields`, `--sort-by`, `--glazed-limit`, or set `GLAZED_OUTPUT=json` for script-friendly output.
- API verbs accept `--server` (default `http://localhost:8088`) and share helpers in `cmd/zine-layout/cmds/api/common.go` for JSON/multipart transport.
- Workflow verbs accept `--data-root` (default `./data`); `workflow/shared.OpenRepositories` bootstraps `zine-layout.db` and on-disk folders on demand.
- Resource identifiers (`--project-id`, `--template-id`, `--page-id`, etc.) are validated client-side and again by the repositories/services listed below.
- Combine these verbs with `--output json | jq` when following the scripted flow under `ttmp/2025-10-11/21-end-to-end-cli-test-plan.md`.

## 4. API Command Group (requires running server)

### 4.1 Shared behaviour
- Start the server with `zine-layout serve --data-root <dir>` before invoking these verbs.
- HTTP routes live in `pkg/serve/server.go` and the `*_routes.go` files referenced in the tables; errors bubble straight out of those handlers.
- For automation, capture JSON to disk (`./tmp-e2e/*.json`) per the convention in `ttmp/2025-10-11/21-end-to-end-cli-test-plan.md`.

### 4.2 Projects & Assets
| CLI verb | Request | Server / Repo | CLI entry | Notes |
|----------|---------|---------------|-----------|-------|
| `zine-layout api projects-list` | `GET /api/projects` | `pkg/serve/projects_routes.go` → `repos.Projects.List` | `cmd/zine-layout/cmds/api/projects-list.go` | Lists every project; supports Glazed field/limit filters. |
| `zine-layout api projects-get --id <project>` | `GET /api/projects/{id}` | `pkg/serve/projects_routes.go` → `repos.Projects.Get` | `cmd/zine-layout/cmds/api/projects-get.go` | Returns project metadata and counters; 404 if missing. |
| `zine-layout api projects-create --name` | `POST /api/projects` | `pkg/serve/projects_routes.go` → `repos.Projects.Create` | `cmd/zine-layout/cmds/api/projects-create.go` | Accepts optional `--description`; server also creates the data directories. |
| `zine-layout api projects-delete --id` | `DELETE /api/projects/{id}` | `pkg/serve/projects_routes.go` → `repos.Projects.Delete` | `cmd/zine-layout/cmds/api/projects-delete.go` | Removes DB row and project folder; irreversible. |
| `zine-layout api images-list --project-id` | `GET /api/projects/{project}/assets` | `pkg/serve/projects_routes.go` → `repos.Assets.ListByProject` | `cmd/zine-layout/cmds/api/images-list.go` | Emits filenames, dimensions, URLs; requires `--project-id`. |
| `zine-layout api images-upload --project-id --files` | `POST /api/projects/{project}/images` | `pkg/serve/projects_routes.go` → `projects.SavePNGImage`, `repos.Assets.Create` | `cmd/zine-layout/cmds/api/images-upload.go` | Multipart upload for explicit file lists; skips non-PNGs client-side. |
| `zine-layout api images-upload-dir --project-id --directory` | `POST /api/projects/{project}/images` | `pkg/serve/projects_routes.go` → `projects.SavePNGImage`, `repos.Assets.Create` | `cmd/zine-layout/cmds/api/images-upload-dir.go` | Scans a directory (optionally `--recursive`) and uploads in batches of 10. |

> Note: The historical CLI README still references an `images-sync` verb; it has been removed and no longer exists in the codebase.

### 4.3 Image Layout Templates
| CLI verb | Request | Server / Repo | CLI entry | Notes |
|----------|---------|---------------|-----------|-------|
| `zine-layout api image-layout-templates list [--project-id]` | `GET /api/image-layout-templates` (global) or `GET /api/projects/{project}/image-layout-templates` | `pkg/serve/layout_templates_routes.go` → `repos.ImageLayoutTemplates.ListGlobal/ListByProject` | `cmd/zine-layout/cmds/api/image_layout_templates/list.go` | Add `--project-id` to merge project + global templates. |
| `zine-layout api image-layout-templates get --template-id` | `GET /api/image-layout-templates/{id}` | `pkg/serve/layout_templates_routes.go` → `repos.ImageLayoutTemplates.Get` | `cmd/zine-layout/cmds/api/image_layout_templates/get.go` | Dumps full settings JSON for a single template. |
| `zine-layout api image-layout-templates create [--project-id]` | `POST /api/image-layout-templates` or `POST /api/projects/{project}/image-layout-templates` | `pkg/serve/layout_templates_routes.go` → `repos.ImageLayoutTemplates.Create` | `cmd/zine-layout/cmds/api/image_layout_templates/create.go` | Provide settings via `--settings-json` or `--settings-file`; defaults scope to global. |
| `zine-layout api image-layout-templates update --template-id` | `PATCH /api/image-layout-templates/{id}` | `pkg/serve/layout_templates_routes.go` → `repos.ImageLayoutTemplates.Update` | `cmd/zine-layout/cmds/api/image_layout_templates/update.go` | Allows renaming, re-scoping, and settings replacement. |
| `zine-layout api image-layout-templates delete --template-id` | `DELETE /api/image-layout-templates/{id}` | `pkg/serve/layout_templates_routes.go` → `repos.ImageLayoutTemplates.Delete` | `cmd/zine-layout/cmds/api/image_layout_templates/delete.go` | Deletes template; ensure no laid-out images depend on it. |

### 4.4 Laid-Out Images
| CLI verb | Request | Server / Repo | CLI entry | Notes |
|----------|---------|---------------|-----------|-------|
| `zine-layout api laid-out-images list --project-id` | `GET /api/projects/{project}/laid-out-images` | `pkg/serve/laid_out_images_routes.go` → `repos.LaidOutImages.ListByProject` | `cmd/zine-layout/cmds/api/laid_out_images/list.go` | Requires `--project-id`; includes template + asset metadata. |
| `zine-layout api laid-out-images get --id` | `GET /api/laid-out-images/{id}` | `pkg/serve/laid_out_images_routes.go` → `repos.LaidOutImages.Get` | `cmd/zine-layout/cmds/api/laid_out_images/get.go` | Returns overrides JSON and latest layout result. |
| `zine-layout api laid-out-images create --project-id --asset-id --template-id` | `POST /api/projects/{project}/laid-out-images` | `pkg/serve/laid_out_images_routes.go` → `services.LayoutService.CreateLaidOutImage` | `cmd/zine-layout/cmds/api/laid_out_images/create.go` | Optional `--overrides-json` to store per-image adjustments. |
| `zine-layout api laid-out-images update --id` | `PATCH /api/laid-out-images/{id}` | `pkg/serve/laid_out_images_routes.go` → `services.LayoutService.RecomputeLaidOutImage` | `cmd/zine-layout/cmds/api/laid_out_images/update.go` | Swap templates or overrides; server recomputes the viewport. |
| `zine-layout api laid-out-images delete --id` | `DELETE /api/laid-out-images/{id}` | `pkg/serve/laid_out_images_routes.go` → `repos.LaidOutImages.Delete` | `cmd/zine-layout/cmds/api/laid_out_images/delete.go` | Removes the record; downstream pages must be recalculated manually. |
| `zine-layout api laid-out-images preview --id` | `GET /api/laid-out-images/{id}/preview` | `pkg/serve/laid_out_images_routes.go` → preview/export helpers | `cmd/zine-layout/cmds/api/laid_out_images/preview.go` | Returns stored layout geometry as JSON; no image bytes yet. |

### 4.5 Image Sequences
| CLI verb | Request | Server / Repo | CLI entry | Notes |
|----------|---------|---------------|-----------|-------|
| `zine-layout api image-sequences list --project-id` | `GET /api/projects/{project}/image-sequences` | `pkg/serve/image_sequences_routes.go` → `repos.ImageSequences.ListByProject` | `cmd/zine-layout/cmds/api/image_sequences/list.go` | Lists sequence headers for a project. |
| `zine-layout api image-sequences get --sequence-id` | `GET /api/image-sequences/{id}` | `pkg/serve/image_sequences_routes.go` → `repos.ImageSequences.Get` | `cmd/zine-layout/cmds/api/image_sequences/get.go` | Returns sequence info plus ordered items. |
| `zine-layout api image-sequences create --project-id` | `POST /api/projects/{project}/image-sequences` | `pkg/serve/image_sequences_routes.go` → `repos.ImageSequences.Create` | `cmd/zine-layout/cmds/api/image_sequences/create.go` | Optional `--name`, `--description`. |
| `zine-layout api image-sequences update --sequence-id` | `PATCH /api/image-sequences/{id}` | `pkg/serve/image_sequences_routes.go` → `repos.ImageSequences.Update` | `cmd/zine-layout/cmds/api/image_sequences/update.go` | Rename or edit description. |
| `zine-layout api image-sequences delete --sequence-id` | `DELETE /api/image-sequences/{id}` | `pkg/serve/image_sequences_routes.go` → `repos.ImageSequences.Delete` | `cmd/zine-layout/cmds/api/image_sequences/delete.go` | Deletes the sequence and its items. |
| `zine-layout api image-sequences add-item --sequence-id --asset-id` | `POST /api/image-sequences/{id}/items` | `pkg/serve/image_sequences_routes.go` → `repos.ImageSequences.AddItem` | `cmd/zine-layout/cmds/api/image_sequences/add_item.go` | Pass `--asset-id` or `--gap` to insert blanks. |
| `zine-layout api image-sequences delete-item --sequence-id --position` | `DELETE /api/image-sequences/{id}/items/{position}` | `pkg/serve/image_sequences_routes.go` → `repos.ImageSequences.DeleteItem` | `cmd/zine-layout/cmds/api/image_sequences/delete_item.go` | Positions are zero-based in the CLI. |
| `zine-layout api image-sequences reorder --sequence-id --items` | `PUT /api/image-sequences/{id}/items` | `pkg/serve/image_sequences_routes.go` → `repos.ImageSequences.ReplaceItems` | `cmd/zine-layout/cmds/api/image_sequences/reorder.go` | Provide comma-separated tokens; use `gap` for empty slots. |

### 4.6 Layout Sequences
| CLI verb | Request | Server / Repo | CLI entry | Notes |
|----------|---------|---------------|-----------|-------|
| `zine-layout api layout-sequences list --project-id` | `GET /api/projects/{project}/layout-sequences` | `pkg/serve/layout_sequences_routes.go` → `repos.LayoutSequences.ListByProject` | `cmd/zine-layout/cmds/api/layout_sequences/list.go` | Lists layout template sequences (page flow definitions). |
| `zine-layout api layout-sequences get --sequence-id` | `GET /api/layout-sequences/{id}` | `pkg/serve/layout_sequences_routes.go` → `repos.LayoutSequences.Get` | `cmd/zine-layout/cmds/api/layout_sequences/get.go` | Returns sequence metadata plus item list. |
| `zine-layout api layout-sequences create --project-id` | `POST /api/projects/{project}/layout-sequences` | `pkg/serve/layout_sequences_routes.go` → `repos.LayoutSequences.Create` | `cmd/zine-layout/cmds/api/layout_sequences/create.go` | Provide `--name`, `--description`; items list starts empty. |
| `zine-layout api layout-sequences update --sequence-id` | `PATCH /api/layout-sequences/{id}` | `pkg/serve/layout_sequences_routes.go` → `repos.LayoutSequences.Update` | `cmd/zine-layout/cmds/api/layout_sequences/update.go` | Rename or change description. |
| `zine-layout api layout-sequences delete --sequence-id` | `DELETE /api/layout-sequences/{id}` | `pkg/serve/layout_sequences_routes.go` → `repos.LayoutSequences.Delete` | `cmd/zine-layout/cmds/api/layout_sequences/delete.go` | Removes the sequence and ordered items. |
| `zine-layout api layout-sequences add-item --sequence-id --page-template-id` | `POST /api/layout-sequences/{id}/items` | `pkg/serve/layout_sequences_routes.go` → `repos.LayoutSequences.AddItem` | `cmd/zine-layout/cmds/api/layout_sequences/add_item.go` | Append a page template (or gap) to the sequence. |
| `zine-layout api layout-sequences delete-item --sequence-id --position` | `DELETE /api/layout-sequences/{id}/items/{position}` | `pkg/serve/layout_sequences_routes.go` → `repos.LayoutSequences.DeleteItem` | `cmd/zine-layout/cmds/api/layout_sequences/delete_item.go` | Removes an item at the given index. |
| `zine-layout api layout-sequences reorder --sequence-id --items` | `PUT /api/layout-sequences/{id}/items` | `pkg/serve/layout_sequences_routes.go` → `repos.LayoutSequences.ReplaceItems` | `cmd/zine-layout/cmds/api/layout_sequences/reorder.go` | Comma-separated page template IDs; `gap` inserts blanks. |

## 5. Workflow Command Group (no server required)

### 5.1 Shared behaviour
- Operates straight against SQLite via `workflow/shared.OpenRepositories`; the CLI creates `zine-layout.db` and required folders if they do not exist.
- Use dedicated data roots (for example `--data-root ./tmp-workflow`) so you can remove artifacts safely between runs.
- Outputs are the same Glazed rows as the API commands, enabling the same `--output json` scripting patterns.

### 5.2 Projects
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow projects list` | `repos.Projects.List` | `cmd/zine-layout/cmds/workflow/projects/list.go` | Enumerates all projects in the local database scoped to `--data-root`. |
| `zine-layout workflow projects get --project-id` | `repos.Projects.Get` | `cmd/zine-layout/cmds/workflow/projects/get.go` | Prints metadata for a single project. |
| `zine-layout workflow projects create --name --description` | `repos.Projects.Create`, `projects.EnsureProjectDirs` | `cmd/zine-layout/cmds/workflow/projects/create.go` | `--skip-scaffold` skips creating on-disk folders under `projects/`. |
| `zine-layout workflow projects delete --project-id` | `repos.Projects.Delete` | `cmd/zine-layout/cmds/workflow/projects/delete.go` | Pass `--keep-artifacts` to preserve the `projects/<id>/` directory. |

### 5.3 Assets
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow assets list --project-id` | `repos.Assets.ListByProject` | `cmd/zine-layout/cmds/workflow/assets/list.go` | Emits filenames, rel paths, dimensions, and metadata JSON. |
| `zine-layout workflow assets create --project-id --file path[,path...]` | `projects.SavePNGImageFromPath`, `repos.Assets.Create` | `cmd/zine-layout/cmds/workflow/assets/create.go` | Copies PNGs into `projects/<id>/images/` and records asset rows. Multiple `--file` flags are supported. |

### 5.4 Image Layout Templates
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow image-layout-templates list [--project-id]` | `repos.ImageLayoutTemplates.ListGlobal/ListByProject` | `cmd/zine-layout/cmds/workflow/image_layout_templates/list.go` | With `--project-id`, merges project-scoped and global templates. |
| `zine-layout workflow image-layout-templates get --template-id` | `repos.ImageLayoutTemplates.Get` | `cmd/zine-layout/cmds/workflow/image_layout_templates/get.go` | Returns normalized settings JSON and metadata. |
| `zine-layout workflow image-layout-templates create [--project-id] --settings-json|--file` | `repos.ImageLayoutTemplates.Create` | `cmd/zine-layout/cmds/workflow/image_layout_templates/create.go` | Defaults to global scope; accepts inline or file-based JSON. |
| `zine-layout workflow image-layout-templates update --template-id [...]` | `repos.ImageLayoutTemplates.Update` | `cmd/zine-layout/cmds/workflow/image_layout_templates/update.go` | Supports renaming, re-scoping (`--project-id` or `--clear-project`), and new settings payloads. |
| `zine-layout workflow image-layout-templates delete --template-id` | `repos.ImageLayoutTemplates.Delete` | `cmd/zine-layout/cmds/workflow/image_layout_templates/delete.go` | Removes the template; ensure dependent laid-out images are updated. |

### 5.5 Laid-Out Images
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow laid-out-images list --project-id` | `repos.LaidOutImages.ListByProject` | `cmd/zine-layout/cmds/workflow/laid_out_images/list.go` | Lists placement metadata for every laid-out image in a project. |
| `zine-layout workflow laid-out-images get --id` | `repos.LaidOutImages.Get` | `cmd/zine-layout/cmds/workflow/laid_out_images/get.go` | Displays overrides and the persisted layout computation JSON. |
| `zine-layout workflow laid-out-images create --project-id --asset-id --template-id [--overrides-json|--file]` | `services.LayoutService.CreateLaidOutImage` | `cmd/zine-layout/cmds/workflow/laid_out_images/create.go` | Runs imagelayout math directly; overrides are optional JSON. |
| `zine-layout workflow laid-out-images update --id [--template-id] [--overrides-json|--file|--clear-overrides]` | `services.LayoutService.RecomputeLaidOutImage` | `cmd/zine-layout/cmds/workflow/laid_out_images/update.go` | Swaps templates or overrides and recomputes placements. |
| `zine-layout workflow laid-out-images delete --id` | `repos.LaidOutImages.Delete` | `cmd/zine-layout/cmds/workflow/laid_out_images/delete.go` | Removes the record; downstream pages should be refreshed. |

### 5.6 Image Sequences
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow image-sequences list --project-id` | `repos.ImageSequences.ListByProject` | `cmd/zine-layout/cmds/workflow/image_sequences/list.go` | Lists sequences ordered by `updated_at`. |
| `zine-layout workflow image-sequences get --sequence-id` | `repos.ImageSequences.Get`, `repos.ImageSequences.ListItems` | `cmd/zine-layout/cmds/workflow/image_sequences/get.go` | Emits a header row plus one row per sequence item. |
| `zine-layout workflow image-sequences create --project-id` | `repos.ImageSequences.Create` | `cmd/zine-layout/cmds/workflow/image_sequences/create.go` | Optional `--name`, `--description`; defaults to “Untitled Sequence”. |
| `zine-layout workflow image-sequences update --sequence-id` | `repos.ImageSequences.Update` | `cmd/zine-layout/cmds/workflow/image_sequences/update.go` | Renames or updates description. |
| `zine-layout workflow image-sequences delete --sequence-id` | `repos.ImageSequences.Delete` | `cmd/zine-layout/cmds/workflow/image_sequences/delete.go` | Deletes the sequence and its items. |
| `zine-layout workflow image-sequences add-item --sequence-id [--asset-id|--gap]` | `repos.ImageSequences.AddItem` | `cmd/zine-layout/cmds/workflow/image_sequences/add_item.go` | Validates asset ownership before appending; `--gap` inserts blanks. |
| `zine-layout workflow image-sequences delete-item --sequence-id --position` | `repos.ImageSequences.DeleteItem` | `cmd/zine-layout/cmds/workflow/image_sequences/delete_item.go` | Removes one position and reindexes the remainder. |
| `zine-layout workflow image-sequences reorder --sequence-id --items` | `repos.ImageSequences.ReplaceItems` | `cmd/zine-layout/cmds/workflow/image_sequences/reorder.go` | Comma-separated tokens (`gap`/`_`/`-` insert blanks). Asset IDs are validated. |

### 5.7 Layout Sequences
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow layout-sequences list --project-id` | `repos.LayoutSequences.ListByProject` | `cmd/zine-layout/cmds/workflow/layout_sequences/list.go` | Lists layout sequences (ordered laid-out images). |
| `zine-layout workflow layout-sequences get --sequence-id` | `repos.LayoutSequences.Get`, `repos.LayoutSequences.ListItems` | `cmd/zine-layout/cmds/workflow/layout_sequences/get.go` | Returns sequence metadata plus item rows. |
| `zine-layout workflow layout-sequences create --project-id` | `repos.LayoutSequences.Create` | `cmd/zine-layout/cmds/workflow/layout_sequences/create.go` | Optional `--name`, `--description`. |
| `zine-layout workflow layout-sequences update --sequence-id` | `repos.LayoutSequences.Update` | `cmd/zine-layout/cmds/workflow/layout_sequences/update.go` | Renames or updates description. |
| `zine-layout workflow layout-sequences delete --sequence-id` | `repos.LayoutSequences.Delete` | `cmd/zine-layout/cmds/workflow/layout_sequences/delete.go` | Deletes sequence and items. |
| `zine-layout workflow layout-sequences add-item --sequence-id --laid-out-image-id` | `repos.LayoutSequences.AddItem` | `cmd/zine-layout/cmds/workflow/layout_sequences/add_item.go` | Appends a laid-out image after verifying project ownership. |
| `zine-layout workflow layout-sequences delete-item --sequence-id --position` | `repos.LayoutSequences.DeleteItem` | `cmd/zine-layout/cmds/workflow/layout_sequences/delete_item.go` | Removes a position (no gaps supported). |
| `zine-layout workflow layout-sequences reorder --sequence-id --order` | `repos.LayoutSequences.ReplaceItems` | `cmd/zine-layout/cmds/workflow/layout_sequences/reorder.go` | Provide comma-separated laid-out image IDs; each is validated. |

### 5.8 Page Templates
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow page-templates list --project-id` | `repos.PageTemplates.ListByProject` | `cmd/zine-layout/cmds/workflow/page_templates/list.go` | Requires `--project-id`; prints template IDs and timestamps. |
| `zine-layout workflow page-templates get --page-template-id` | `repos.PageTemplates.Get` | `cmd/zine-layout/cmds/workflow/page_templates/get.go` | Shows stored JSON plus metadata. |
| `zine-layout workflow page-templates create --project-id --template-json|--file` | `repos.PageTemplates.Create` | `cmd/zine-layout/cmds/workflow/page_templates/create.go` | Provide JSON inline or via file; defaults name to “Untitled Page Template”. |
| `zine-layout workflow page-templates delete --page-template-id` | `repos.PageTemplates.Delete` | `cmd/zine-layout/cmds/workflow/page_templates/delete.go` | Removes template; caller must clean dependent pages manually. |

### 5.9 Laid-Out Pages
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow laid-out-pages list --project-id` | `repos.LaidOutPages.ListByProject` | `cmd/zine-layout/cmds/workflow/laid_out_pages/list.go` | Lists renderable pages linked to a project. |
| `zine-layout workflow laid-out-pages get --page-id` | `repos.LaidOutPages.Get` | `cmd/zine-layout/cmds/workflow/laid_out_pages/get.go` | Returns template linkage plus render metadata. |
| `zine-layout workflow laid-out-pages create --project-id --template-id --laid-out-image-id` | `repos.LaidOutPages.Create` | `cmd/zine-layout/cmds/workflow/laid_out_pages/create.go` | Creates a page record; run `render` afterwards to produce PNGs. |
| `zine-layout workflow laid-out-pages update-image --page-id --laid-out-image-id` | `services.PagesService.UpdatePageImage` | `cmd/zine-layout/cmds/workflow/laid_out_pages/update_image.go` | Switches the laid-out image backing a page. |
| `zine-layout workflow laid-out-pages render --page-id` | `services.PagesService.RenderPage` | `cmd/zine-layout/cmds/workflow/laid_out_pages/render.go` | Generates `thumbnail.png`, `full.png`, `combined.png`, `left.png`, `right.png` under the project data root. |
| `zine-layout workflow laid-out-pages delete --page-id` | `repos.LaidOutPages.Delete` | `cmd/zine-layout/cmds/workflow/laid_out_pages/delete.go` | Deletes the record; remove files manually if needed. |

### 5.10 Zines
| CLI verb | Services / Repo | CLI entry | Notes |
|----------|-----------------|-----------|-------|
| `zine-layout workflow zines list --project-id` | `repos.Zines.ListByProject` | `cmd/zine-layout/cmds/workflow/zines/list.go` | Lists zines scoped to a project. |
| `zine-layout workflow zines get --zine-id` | `repos.Zines.Get`, `repos.ZinePages.ListByZine` | `cmd/zine-layout/cmds/workflow/zines/get.go` | Displays zine metadata and ordered page IDs. |
| `zine-layout workflow zines create --project-id --pages` | `repos.Zines.Create`, `repos.ZinePages.Replace` | `cmd/zine-layout/cmds/workflow/zines/create.go` | Accepts comma-separated page IDs; can start empty. |
| `zine-layout workflow zines set-pages --zine-id --pages` | `repos.ZinePages.Replace` | `cmd/zine-layout/cmds/workflow/zines/set_pages.go` | Replaces the page order in one shot. |
| `zine-layout workflow zines export --zine-id --preset --out --dpi` | `services.ImpositionService.ImposeZine`, `pkg/export.SheetsToPDF` | `cmd/zine-layout/cmds/workflow/zines/export.go` | Streams PDF to `--out` (or stdout) after composing sheets from rendered pages. |
| `zine-layout workflow zines delete --zine-id` | `repos.Zines.Delete` | `cmd/zine-layout/cmds/workflow/zines/delete.go` | Removes the zine and its page associations. |

### 5.11 Data locations
- Workflow commands write under `<data-root>/projects/<project-id>/`. Page renders land in `pages/<page-id>/` as PNG variants, and render metadata is persisted in `laid_out_pages.result_json`.
- Zine exports default to the working directory unless `--out` is specified; keep them outside the data root to avoid mixing artifacts.
- Presets are looked up under `<data-root>/presets/`; `serve` seeds defaults but workflow commands expect them to exist already.

## 6. CLI-Only Rendering Pipeline (current state)
1. **Run fully serverless setup** (no HTTP API required):
   - Create a project (`workflow projects create --name "CLI Demo"`).
   - Import PNGs (`workflow assets create --project-id $PROJECT --file img1.png --file img2.png`).
   - Define layout logic (`workflow image-layout-templates create --project-id $PROJECT --file imagelayout.json`).
   - Compute placements (`workflow laid-out-images create --project-id $PROJECT --asset-id $ASSET --template-id $TEMPLATE`).
   - Author page templates (`workflow page-templates create --project-id $PROJECT --file page_template.json`) and render pages (`workflow laid-out-pages create ...` + `workflow laid-out-pages render ...`).
   - Optionally assemble sequences (`workflow image-sequences ...`, `workflow layout-sequences ...`) before exporting a zine (`workflow zines export --out ./out.pdf`).
2. **Mix and match with API verbs** when you specifically need HTTP surface area (e.g., to share data with the running server).
3. **Reuse artifacts**: persist Glazed JSON output under `./tmp-e2e` as described in `ttmp/2025-10-11/21-end-to-end-cli-test-plan.md` so downstream tasks can rehydrate IDs without recomputation.
4. **Validate outputs**: inspect rendered PNGs inside the data root and confirm the PDF with `file` or `pdfinfo` before handing off to frontend work.

## 7. Serverless Coverage Backlog
- Optional niceties: add `workflow assets delete` and update/rename operations for projects if needed.
- Keep this inventory in sync with any new verbs; cross-link updates in `ttmp/2025-10-10/07-phase2-backend-and-ui-progress-changelog.md` and `ttmp/2025-10-10/09-system-specification-after-phase1-and-phase2.md`.

## 8. Handy References
- `zine-layout help` and `zine-layout api image-layout-templates --help` surface Glazed auto-generated usage for any verb.
- `cmd/zine-layout/cmds/api/README.md` contains additional examples (be aware it still mentions the deprecated `images-sync` command).
- Cross-check HTTP handler behaviour in `pkg/serve/*_routes.go` when debugging responses, and align service expectations with `pkg/services/pages.go`, `pkg/services/zines.go`, and `pkg/services/imposition.go`.
