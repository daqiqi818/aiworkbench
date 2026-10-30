---
description: "Task list for the multi-episode editing workbench"
---

# Tasks: Multi-Episode Editing Workbench

**Input**: Design documents from `/specs/001-episode-editing-workbench/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Go domain and handler tests are required because ownership and lifecycle transitions are
the feature's security and data-integrity boundary. Frontend validation uses the production build.

## Phase 1: Setup

- [X] T001 Create `backend/internal/dramaedit/` and `web/src/pages/drama-editing/` feature directories.
- [X] T002 [P] Review `specs/001-episode-editing-workbench/contracts/drama-editing-api.md` against the implemented public route surface.

## Phase 2: Foundational

- [X] T003 Add `DramaEditTask` and append-only `DramaEditDecision` models in `backend/internal/model/models_drama_edit.go`.
- [X] T004 Add local schema v14 migration, models, readiness checks, and a documented v14 entry in `backend/internal/database/local_schema.go` and `docs/content/docs/backend/backend-database.mdx`.
- [X] T005 Implement owned SQLite reads and transactional writes for task and decision records in `backend/internal/dramaedit/store.go`.
- [X] T006 Define the narrow `DramaEditingPort` and wire it through `backend/internal/localapp/ports.go`, `backend/internal/localapp/app.go`, `backend/internal/handler/runtime_dependencies.go`, and `backend/internal/bootstrap/runtime.go`.
- [X] T007 Register `RegisterDramaEditingRoutes` with the desktop API graph in `backend/internal/handler/api.go`.

**Checkpoint**: Foundation is ready for task lifecycle work.

## Phase 3: User Story 1 - New Multi-Episode Editing Task (Priority: P1) MVP

**Goal**: A creator can create and reopen a user-owned task containing 2-12 validated episodes.

**Independent Test**: A valid two-episode request persists and returns its proposal; one episode,
duplicate positions, missing source rights, or a foreign task read are rejected.

- [X] T008 [P] [US1] Write create, validation, ownership, and list tests in `backend/internal/dramaedit/service_test.go`.
- [X] T009 [P] [US1] Write create/list/get HTTP envelope tests in `backend/internal/handler/drama_edit_test.go`.
- [X] T010 [US1] Implement request validation, proposal construction, creation, list, and get behavior in `backend/internal/dramaedit/service.go` using the documented 2-12 episode, unique order, and confirmed-rights rules.
- [X] T011 [US1] Implement `GET` and `POST /drama-editing/tasks` plus `GET /drama-editing/tasks/:id` in `backend/internal/handler/drama_edit.go`.
- [X] T012 [P] [US1] Create typed request, proposal, and task API functions using `http` in `web/src/services/api/drama-editing.ts`.
- [X] T013 [US1] Implement the task list and new-task form with two editable episode rows in `web/src/pages/drama-editing/index.tsx` and `web/src/pages/drama-editing/drama-editing.css`.
- [X] T014 [US1] Add lazy route loading and the independent sidebar menu item in `web/src/router.tsx`, `web/src/lib/workspace-route-modules.ts`, and `web/src/components/layout/workspace-sidebar-nav.tsx`.

**Checkpoint**: A creator can submit and reopen a valid local multi-episode task from the menu.

## Phase 4: User Story 2 - Review and Approve a Proposal (Priority: P2)

**Goal**: A creator can review, revise, approve, or cancel a submitted task without starting a
fictional render.

**Independent Test**: An awaiting task transitions only through valid approve, revision, and cancel
actions and preserves an append-only decision history.

- [X] T015 [P] [US2] Add lifecycle transition and decision-history tests in `backend/internal/dramaedit/service_test.go`.
- [X] T016 [P] [US2] Add approve, revision, and cancel HTTP route tests in `backend/internal/handler/drama_edit_test.go`.
- [X] T017 [US2] Implement transition guards, decision recording, and next-action projection in `backend/internal/dramaedit/service.go`.
- [X] T018 [US2] Implement approve, revision, and cancel handlers in `backend/internal/handler/drama_edit.go`.
- [X] T019 [US2] Add approval actions and the proposal/decision detail panel in `web/src/pages/drama-editing/index.tsx`.

**Checkpoint**: Unapproved tasks cannot be represented as started; approved tasks clearly report
that a rendering executor is not yet connected.

## Phase 5: User Story 3 - Track Task Status (Priority: P3)

**Goal**: A creator can scan task state and reopen any task to understand its current stage and
next action.

**Independent Test**: Task list sorting and detail projection show consistent statuses and do not
expose another user's task.

- [X] T020 [P] [US3] Add summary projection and status-filter tests in `backend/internal/dramaedit/service_test.go`.
- [X] T021 [US3] Add status filtering to `GET /drama-editing/tasks` in `backend/internal/handler/drama_edit.go` and `backend/internal/dramaedit/service.go`.
- [X] T022 [US3] Add status filters, task summary metadata, empty states, and reload behavior in `web/src/pages/drama-editing/index.tsx`.

**Checkpoint**: The workbench presents an actionable, persistent task history.

## Phase 6: Polish and Validation

- [X] T023 Run `gofmt` on `backend/internal/model/models_drama_edit.go`, `backend/internal/dramaedit/`, and `backend/internal/handler/drama_edit*.go`.
- [X] T024 Run focused Go tests and `cd backend && go test ./...` (the complete suite reports existing unrelated channel-model failures).
- [X] T025 Run `cd web && bun run build` and resolve route/type failures.
- [X] T026 Reconcile `specs/001-episode-editing-workbench/quickstart.md` with executed validation and mark completed tasks in `specs/001-episode-editing-workbench/tasks.md`.

## Phase 7: Asset-Folder Episode Workflow Amendment

**Goal**: Creators can create episode folders and use their available video assets to create a
compact multi-episode task with constrained delivery choices.

**Independent Test**: A newly created asset folder is visible after save. A two-folder task with
valid delivery enums and a 15-300 second range persists folder snapshots; empty, foreign,
duplicate, out-of-range, inverted, or unsupported selections are rejected.

- [X] T027 [P] Add owner-scoped active-video folder candidate query coverage in `backend/internal/repository/asset_library_test.go`.
- [X] T028 [P] Add folder-backed creation, duration range, enum, order, and unavailable-folder regression tests in `backend/internal/dramaedit/service_test.go`.
- [X] T029 [P] Add episode-folder endpoint and amended create-envelope tests in `backend/internal/handler/drama_edit_test.go`.
- [X] T030 Add `folderId`, `folderName`, `sequenceNumber`, and positive `videoCount` episode snapshot rules; quote constraints: "2-12 unique folder IDs", "15-300 seconds inclusive", "minimum cannot exceed maximum", "language: zh|en", and "platform: tiktok|facebook|wechat_channels|douyin" in `backend/internal/dramaedit/types.go` and `backend/internal/dramaedit/service.go`.
- [X] T031 Implement current-user active-video folder candidate reads and creation-time ownership/availability revalidation in `backend/internal/repository/asset_library.go` and `backend/internal/dramaedit/store.go`.
- [X] T032 Add the `GET /drama-editing/episode-folders` route and update task binding in `backend/internal/handler/drama_edit.go` and `backend/internal/localapp/ports.go`.
- [X] T033 Add the v15 additive migration/readiness validation and document the task payload revision in `backend/internal/database/local_schema.go` and `docs/content/docs/backend/backend-database.mdx`.
- [X] T034 Repair the rendered new-folder state with a focused input, explicit save/cancel controls, scope-aware refresh, and error retention in `web/src/pages/assets/index.tsx`.
- [X] T035 Update typed drama-editing API contracts for episode folder candidates, duration range, language enum, platform enum, and folder snapshots in `web/src/services/api/drama-editing.ts`.
- [X] T036 Implement the compact folder-driven creation form, default-all selection, selected-count feedback, and revised task/detail summaries in `web/src/pages/drama-editing/index.tsx` and `web/src/pages/drama-editing/drama-editing.css`.
- [X] T037 Run `gofmt`, focused Go tests, `cd backend && go test ./...`, and `cd web && bun run build`; record only actual results in `specs/001-episode-editing-workbench/quickstart.md` and mark completed tasks here.

## Phase 8: Browser-Local Asset Folder Availability Amendment

**Goal**: Browser-local asset folders and videos are visible to the owner-scoped episode candidate
endpoint without uploading media bytes again.

**Independent Test**: A browser-local folder named `67906322` with a video is migrated by name,
its local asset reference is rebound to the canonical folder ID, and the new-task query receives it
as a positive-video candidate.

- [X] T038 [US4] Record the browser-local folder migration and metadata bridge regression flow in `specs/001-episode-editing-workbench/quickstart.md`.
- [X] T039 [US4] Implement name-deduplicated browser-local folder migration, local asset-folder rebinding, and canonical folder creation in `web/src/services/workspace-asset-folders.ts`.
- [X] T040 [US4] Implement metadata-only browser-local asset synchronization after local persistence in `web/src/services/workspace-asset-repository.ts`.
- [X] T041 [US4] Synchronize browser-local asset metadata before the episode-folder query and surface synchronization errors in `web/src/pages/drama-editing/index.tsx`.
- [X] T042 Run focused web tests and `cd web && bun run build`; record actual results in `specs/001-episode-editing-workbench/quickstart.md` and mark completed tasks here.

## Phase 9: Task Submission Feedback Amendment

**Goal**: Every creation confirmation produces an observable validation, loading, error, or
success result instead of leaving the dialog apparently inactive.

**Independent Test**: An empty form, one selected episode, loading candidates, and a valid
two-episode form each show their distinct expected outcome.

- [X] T043 [US5] Add explicit form-validation, episode-count, and candidate-query submission feedback in `web/src/pages/drama-editing/index.tsx`.
- [X] T044 [US5] Record invalid and valid creation-dialog regression scenarios in `specs/001-episode-editing-workbench/quickstart.md`.
- [X] T045 Run `cd web && bun run build` and mark the completed task submission feedback work in `specs/001-episode-editing-workbench/tasks.md`.
- [X] T046 [US5] Keep task-creation blocking feedback visible inside the dialog in `web/src/pages/drama-editing/index.tsx`.
- [X] T047 Verify the one-episode validation state in a browser and run `cd web && bun run build` in `specs/001-episode-editing-workbench/quickstart.md`.

## Phase 10: Folder Video Episode Selection Correction

- [X] T048 [US5] Return individual eligible videos under each asset-folder candidate in `backend/internal/repository/asset_library.go`, `backend/internal/dramaedit/store.go`, and `backend/internal/dramaedit/types.go`.
- [X] T049 [US5] Validate and persist ordered `episodeAssetIds`, allowing multiple selected videos from one folder, in `backend/internal/dramaedit/service.go`.
- [X] T050 [US5] Select folders then individual video episodes in `web/src/pages/drama-editing/index.tsx` and update `web/src/services/api/drama-editing.ts`.
- [X] T051 [US5] Cover one-folder multi-video creation in `backend/internal/dramaedit/service_test.go` and candidate expansion in `backend/internal/repository/asset_library_test.go`.

## Phase 11: Approved Production Execution

**Goal**: Approval starts one durable local production work and the editing task exposes its real
queue, result, and failure state.

**Independent Test**: A renderable two-video task transitions from approval to processing and then
completed when the local work returns an output resource; a missing source becomes blocked.

- [X] T052 [US6] Add v16 production task, delivery resource, and failure persistence fields plus additive migration in `backend/internal/model/models_drama_edit.go` and `backend/internal/database/local_schema.go`.
- [X] T053 [US6] Implement owned source-resource resolution, ordered duration-bounded timeline construction, production admission, and worker outcome projection in `backend/internal/dramaedit/store.go` and `backend/internal/dramaedit/service.go`.
- [X] T054 [US6] Wire the narrow local render starter and historical `POST /drama-editing/tasks/:id/start` endpoint through `backend/internal/app/service.go`, `backend/internal/localapp/ports.go`, `backend/internal/handler/drama_edit.go`, and `backend/internal/handler/runtime_dependencies.go`.
- [X] T055 [US6] Add automatic approval, completed output, and source-blocked regression coverage in `backend/internal/dramaedit/service_test.go`.
- [X] T056 [US6] Show processing progress, a blocked reason, final output link, and historical start command in `web/src/pages/drama-editing/index.tsx` and `web/src/services/api/drama-editing.ts`.
- [X] T057 Run targeted Go lifecycle/migration tests and `cd web && bun run build`; record results in `specs/001-episode-editing-workbench/quickstart.md`.

## Phase 12: Browser-Local Production Source Handoff

**Goal**: A task created from browser-local videos transfers its selected source bytes to the
local backend exactly once before durable FFmpeg production begins.

**Independent Test**: A ready owned `resource:` storage key resolves without an asset
representation; the UI prepares local source blobs before its approval/start request.

- [X] T058 [US8] Add a focused domain test proving a ready owned `resource:` storage key can supply an episode without an asset representation in `backend/internal/dramaedit/service_test.go`.
- [X] T059 [US8] Resolve remaining episode assets through owned ready video resource storage keys in `backend/internal/dramaedit/store.go`.
- [X] T060 [US8] Prepare selected browser-local source blobs with stable idempotent resource uploads and persist their resource-backed asset metadata in `web/src/pages/drama-editing/index.tsx`.
- [X] T061 [US8] Invoke source preparation before approve/start, preserve a truthful failure in the task UI, and avoid duplicate uploads in `web/src/pages/drama-editing/index.tsx`.
- [X] T062 [US8] Update the production source-handoff validation steps in `specs/001-episode-editing-workbench/quickstart.md`.
- [X] T063 Validate payload source resolution with `go test ./internal/dramaedit`, build `web/`, and exercise the start flow in a real browser profile.

## Phase 13: Local Rendering Runtime

- [X] T064 [US9] Install the FFmpeg/FFprobe runtime dependency in the final local backend image in `backend/Dockerfile`.
- [X] T065 [US9] Rebuild the local Docker services, verify both binaries inside `backend`, retry the source-ready drama-edit task, and record its terminal production result in `specs/001-episode-editing-workbench/quickstart.md`.
- [X] T066 [US9] Create a new idempotent operation ID for a terminal failed production retry, preserve active/succeeded reuse, and cover it in `backend/internal/dramaedit/service.go` and `backend/internal/dramaedit/service_test.go`.

## Dependencies & Execution Order

`T001-T007` block all user stories. US1 (`T008-T014`) creates the aggregate and menu. US2
(`T015-T019`) depends on US1's aggregate. US3 (`T020-T022`) depends on list/read behavior from
US1 and can be completed after US2 because its page changes share the same file. Polish follows all
stories.
The amendment tasks (`T027-T037`) depend on the existing task aggregate and can run after Phase 6;
T030-T033 provide the backend contract before T035-T036 consume it.
The browser-local availability tasks (`T038-T042`) depend on the amended folder workflow and are
completed after T027-T037.

## Parallel Opportunities

- T002 can proceed while directories are created.
- The two backend test files in each story can be prepared independently of the frontend API module.
- T012 can proceed after the API contract is stable while backend tests are being written.
- T027-T029 can be prepared in parallel before the corresponding implementation tasks.

## Implementation Strategy

Implement T001-T014 first to establish a usable MVP. Add lifecycle controls next, then status
filters and presentation. Do not attach an unreviewed external rendering runtime; the approved
state is an honest integration boundary. Implement T027-T034 before the amended web contract and
compact form, then validate the complete amended workflow.
