# CLI Testing Playbook for `zine-layout`

This guide captures the repeatable workflow for exercising the Go-based CLI after changes. Follow it whenever you touch CLI verbs, backend routes, or data flows so the next developer can quickly validate behaviour.

---

## 1. Environment Checklist

- **Go toolchain**: `go version` ≥ 1.21.
- **Node/Vite setup** (only if you need the dev server running while testing API verbs): see `README.md`.
- **Working directory**: repository root (`/home/manuel/workspaces/2025-09-23/book-spread-generator/zine-layout`).
- **Data root**: prefer an isolated folder (e.g. `tmp-workflow-test/`). The CLI will create SQLite databases inside `<data-root>/zine-layout.db`.
- **Binary entry point**: `cmd/zine-layout`. Commands should be invoked via `go run ./cmd/zine-layout …` during development.

Whenever tests rely on long-running services, use a separate terminal to keep processes alive.

---

## 2. Discovering the Command Tree

1. `go build ./cmd/zine-layout` — ensures the tree compiles.
2. `go run ./cmd/zine-layout --help` — confirms verb groups are registered once.
3. Optional: `go run ./cmd/zine-layout help workflow` for deep Glazed help (shows layers, flags, examples).

> **Code reference**: verb namespaces live in `cmd/zine-layout/cmds/<group>/`. Each group exposes `NewCommand()` returning a Cobra command wired with Glazed.

---

## 3. Imagelayout Engine Smoke Tests

These exercise `cmd/zine-layout/cmds/imagelayout/compute.go`.

1. **Basic fit**  
   ```bash
   go run ./cmd/zine-layout imagelayout compute \
     --source-width 2400 --source-height 1600 \
     --mode fit --fit-width 1200 --fit-height 800
   ```
   Expect JSON output with `mode: contain` in the trace.

2. **Spec-driven run**  
   ```bash
   cat > /tmp/layout-spec.yaml <<'YAML'
   settings:
     mode: crop
     crop_to_fill: true
     anchor_preset: top-left
     margin_top_in: 0.1
   image:
     width: 4032
     height: 3024
   YAML

   go run ./cmd/zine-layout imagelayout compute \
     --spec /tmp/layout-spec.yaml --fit-width 1500
   ```
   Ensure CLI honours YAML overrides and mixes flag overrides (fit width in this example).

3. **Focus point override**  
   ```bash
   go run ./cmd/zine-layout imagelayout compute \
     --source-width 3000 --source-height 2000 \
     --mode crop --crop-to-fill \
     --focus-source-x 1200 --focus-source-y 800 \
     --focus-target-x 0.3 --focus-target-y 0.6
   ```
   Confirm `trace.steps` contains `focus_applied: true`.

---

## 4. Render Command Regression

Reference: `cmd/zine-layout/cmds/render/command.go`.

1. Generate fixture input:
   ```bash
   mkdir -p /tmp/zine-render
   go run ./cmd/zine-layout render --test \
     --spec examples/specs/sample-layout.yaml \
     --output-dir /tmp/zine-render --verbose
   ```
   Validate the CLI prints saved file paths and the directory contains the generated PNGs.

2. Exercise Glazed output layers:
   - `--output json` (via Glazed) — inspect `--output` rendering from `--help`.
   - Ensure errors are surfaced with non-zero exit codes (e.g. missing spec).

---

## 5. Serve Command Smoke Test

Reference: `cmd/zine-layout/cmds/serve/command.go`.

1. Start server:
   ```bash
   go run ./cmd/zine-layout serve \
     --root web/dist \
     --data-root tmp-playbook-data \
     --addr :8090
   ```
   Observe log line `listening on :8090`.

2. In another terminal, hit health endpoints or CLI API verbs (next section) to confirm the server responds.

3. Stop with `Ctrl+C` and watch for graceful shutdown.

---

## 6. API Verb Group (HTTP Client) Tests

Reference: `cmd/zine-layout/cmds/api`.

### 6.1 Projects

```bash
go run ./cmd/zine-layout api projects-create \
  --server http://localhost:8090 \
  --name "Playbook Project"

go run ./cmd/zine-layout api projects-list \
  --server http://localhost:8090 --output json

go run ./cmd/zine-layout api projects-get \
  --server http://localhost:8090 \
  --id <project-id>
```

Expect tabular output by default, JSON when requested. Capture IDs for downstream steps.

### 6.2 Image Upload and Listing

```bash
go run ./cmd/zine-layout api images-upload \
  --server http://localhost:8090 \
  --project-id <project-id> \
  --files web/public/demo/001.png,web/public/demo/002.png

go run ./cmd/zine-layout api images-list \
  --server http://localhost:8090 \
  --project-id <project-id> --fields name,width,height
```

Include a failure test (non-existent project) to ensure HTTP errors propagate.

### 6.3 Image Sequences & Layout Sequences

Each verb group exposes `list`, `get`, `create`, `update`, `delete`, `add-item`, `delete-item`, and `reorder`. Run canonical workflow:

1. Create sequence with ordered images.
2. Verify `get` returns the item list.
3. Call `reorder` to swap positions; rerun `get` to confirm.
4. Remove an item and delete the sequence.

Use `--glazed-limit` and `--output json` on `list` to confirm Glazed features remain intact.

### 6.4 Image Layout Templates & Laid-Out Images

Similar routine:

1. `image-layout-templates create/list/get/update/delete`.
2. `laid-out-images create list get preview update delete`.
3. The preview verb may return binary data; test with `--output json` to inspect metadata or `curl` if easier.

---

## 7. Workflow Verb Group (Direct Repository Access)

Reference directories mirror the verb groups:

- `cmd/zine-layout/cmds/workflow/page_templates`
- `cmd/zine-layout/cmds/workflow/laid_out_pages`
- `cmd/zine-layout/cmds/workflow/zines`

### 7.1 Page Templates

```bash
DATA_ROOT=tmp-workflow-test

go run ./cmd/zine-layout workflow page-templates create \
  --data-root $DATA_ROOT \
  --name "CLI Template" \
  --template-json '{"settings":{"paper_width_in":8.5}}'

go run ./cmd/zine-layout workflow page-templates list \
  --data-root $DATA_ROOT --output json
```

Run `get` with returned ID, then `delete` and ensure it disappears from `list`.

### 7.2 Laid-Out Pages

Needs existing project + page template + laid-out images:

1. Seed via API commands (or reuse `tmp-workflow-test` seeded by previous workflows).
2. `create`: supply `--project-id`, `--template-id`, and `--inputs`.
3. `get`: validate inputs array.
4. `set-inputs`: reorder `inputs`, confirm via `get`.
5. `delete`: ensure cleanup.

### 7.3 Zines

1. `create`: requires project ID and laid-out pages.
2. `list`, `get`: verify page ordering.
3. `set-pages`: reorder or replace pages.
4. `delete`: confirm removal.

For each mutation command, verify row output includes status (e.g. `inputs-updated`, `deleted`).

---

## 8. Render & Workflow Integration Loop

After mutating workflow entities, run integration sweeps:

1. Serve the API (`go run ./cmd/zine-layout serve …`).
2. Use API verbs to fetch the newly created resources; ensure they align with direct repo changes.
3. Optionally call `imagelayout compute` on the settings generated from layout templates to verify parity with front-end expectations.

---

## 9. Output Format Matrix

For every `list` command (API or workflow):

- Table (default).
- `--output json`.
- `--output csv`.
- `--output markdown`.
- `--fields id,name`.
- `--filter` to exclude columns.
- `--glazed-limit` + `--glazed-skip` to test pagination helpers.

This matrix confirms the Glazed middleware is still wired correctly.

---

## 10. Error Handling Checks

1. Missing required flag: e.g. omit `--project-id` on `workflow laid-out-pages list`; expect a descriptive error.
2. Server offline: run an API command without the server; should produce `connect: connection refused`.
3. Invalid JSON/Spec: supply malformed `--template-json` or spec file to confirm parsing errors bubble up.
4. Permission issues: point `--data-root` at read-only directory; CLI should fail gracefully.

Document any deviations in `ttmp/2025-10-10/07-phase2-backend-and-ui-progress-changelog.md`.

---

## 11. Cleanup

- Remove temporary data roots (`rm -rf tmp-workflow-test tmp-playbook-data`).
- Delete spec files under `/tmp`.
- Stop background servers.
- Reset Glazed output preferences if you toggled environment variables.

---

## 12. Quick Reference Command Set

```
go build ./cmd/zine-layout
go run ./cmd/zine-layout --help
go run ./cmd/zine-layout imagelayout compute --source-width 2400 --source-height 1600 --mode fit --fit-width 1200 --fit-height 800
go run ./cmd/zine-layout render --test --spec examples/specs/sample-layout.yaml --output-dir /tmp/zine-render
go run ./cmd/zine-layout serve --root web/dist --data-root tmp-playbook-data --addr :8090
go run ./cmd/zine-layout api projects-list --server http://localhost:8090 --output json
go run ./cmd/zine-layout workflow page-templates list --data-root tmp-workflow-test
```

Keep this cheat sheet updated if you add new verb groups or options. Mention any updates in `05-expansion-plan-for-zine-layout-platform.md` (documentation cadence) and the relevant changelog entries.
