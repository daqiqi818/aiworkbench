# Quickstart: Multi-Episode Editing Workbench

## Prerequisites

Use an isolated local backend data directory. Do not run against a production workspace. The
feature requires the backend to migrate its local schema and the web dependencies already present.

## Validation flow

1. Run the focused Go domain and handler tests listed in the implementation tasks.
2. Run `cd backend && go test ./...` after the feature tests pass.
3. Run `cd web && bun run build` to type-check and bundle the lazy route.
4. Start the local backend only when browser validation is explicitly required, using
   `CANVAS_BACKEND_DATA_DIR=../.local/project-workbench-debug go run ./cmd/server` from `backend/`.
5. Open `/drama-editing`, create a task containing two included episodes with confirmed source
   rights, and verify its proposal and `awaiting_approval` status.
6. Request a revision, approve the updated proposal, and cancel a separate task. Verify each
   decision survives page reload and that an approved task states execution is pending rather than
   presenting a fictional output.

See [data-model.md](./data-model.md) for validation rules and
[contracts/drama-editing-api.md](./contracts/drama-editing-api.md) for the request surface.
