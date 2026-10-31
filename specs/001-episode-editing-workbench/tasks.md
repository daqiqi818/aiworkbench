---
description: "Task list for the multi-episode editing workbench"
---

# Tasks: Multi-Episode Editing Workbench

**Input**: Design documents from `/specs/001-episode-editing-workbench/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Go domain and handler tests are required because ownership and lifecycle transitions are
the feature's security and data-integrity boundary. Frontend validation uses the production build.

## Phase 15: Short-Drama Task Operations and Creator Report Amendment

- [X] T085 Add owner-scoped batch task deletion results and creation-time sorting in `backend/internal/dramaedit/store.go` and `backend/internal/dramaedit/service.go`.
- [X] T086 Add the delete task route and narrow domain-port wiring in `backend/internal/handler/drama_edit.go`, `backend/internal/handler/runtime_dependencies.go`, and `backend/internal/localapp/ports.go`.
- [X] T087 Add typed task deletion API support in `web/src/services/api/drama-editing.ts`.
- [X] T088 Replace creator-facing multi-episode terminology, add folder-first video selection, persisted non-name preferences, progress projection, bulk selection, delivery playback/download and report modal in `web/src/pages/drama-editing/index.tsx` and `web/src/pages/drama-editing/drama-editing.css`.
- [X] T089 Generate an owner-scoped editorial report from native source cuts and checkpoints in `openmontage-runner/app/native_workspace.py`.
- [X] T090 Update feature design artifacts and run focused backend/frontend validation in `specs/001-episode-editing-workbench/plan.md`, `data-model.md`, `contracts/drama-editing-api.md`, and `quickstart.md`.

## Phase 16: Verified Task and Source Deletion Amendment

**Goal**: A creator can reliably remove selected completed tasks and can explicitly opt into
safe source-video deletion without affecting delivery assets.

**Independent Test**: In a real browser, select two completed tasks, delete with source deletion
unchecked, and verify both task cards disappear while their assets remain; repeat with source
deletion checked and verify only eligible source assets are removed.

- [X] T091 [P] [US17] Add focused deletion tests for two completed tasks, duplicate/foreign IDs, active-task rejection, shared source retention, and opted-in source deletion in `backend/internal/dramaedit/service_test.go`.
- [X] T092 [P] [US17] Add JSON deletion request/response route tests in `backend/internal/handler/drama_edit_test.go`.
- [X] T093 [US17] Extend the deletion command/result, delegate owner-scoped asset deletion, and recheck remaining drama-task source references in `backend/internal/dramaedit/types.go`, `backend/internal/dramaedit/store.go`, and `backend/internal/dramaedit/service.go`.
- [X] T094 [US17] Bind the JSON deletion command and wire the existing asset deletion authority into the short-drama service in `backend/internal/handler/drama_edit.go` and `backend/internal/app/service.go`.
- [X] T095 [US17] Send deletion IDs and `deleteEpisodeFiles` as a JSON command and expose typed source-file results in `web/src/services/api/request.ts` and `web/src/services/api/drama-editing.ts`.
- [X] T096 [US17] Replace the static confirmation text with a default-unchecked “同时删除剧集文件” checkbox and present task/source deletion outcomes in `web/src/pages/drama-editing/index.tsx`.
- [X] T097 [US17] Rebuild the local stack and execute the two-completed-task browser deletion workflow in `specs/001-episode-editing-workbench/quickstart.md`; record exact evidence and mark this task only after both unchecked and checked paths are verified.

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

## Phase 14: OpenMontage Published Runtime Foundation

**Goal**: The repository contains a reproducible, license-compliant OpenMontage source snapshot
and an internal runner that creates and reports isolated native cinematic workspaces without
pretending to run unavailable paid production stages.

**Independent Test**: Two scoped fixture productions return the native `cinematic` stage order,
persist distinct project markers and research checkpoints, deny cross-owner reads/cancellation,
and survive a runner restart from the same workspace volume.

- [X] T067 [US10] Vendor the clean tracked OpenMontage source at a recorded commit, exclude `.env`, `.git`, `node_modules`, `.venv`, `projects`, caches and generated artifacts, and add `third_party/openmontage/UPSTREAM.md`.
- [X] T068 [P] [US10] Add native cinematic manifest/checkpoint and owner-isolation tests in `openmontage-runner/tests/test_runner.py`.
- [X] T069 [US10] Implement the authenticated runner capability, production creation, checkpoint projection, approval and cancellation contract in `openmontage-runner/app/main.py` and `openmontage-runner/app/native_workspace.py`.
- [X] T070 [US10] Build the internal CPU runtime with the vendored source, Python, Node/Remotion and FFmpeg in `openmontage-runner/Dockerfile`, `openmontage-runner/requirements.txt`, and `openmontage-runner/entrypoint.sh`.
- [X] T071 [US11] Enforce opaque owner-and-production workspace resolution, path-component validation and token authentication in `openmontage-runner/app/main.py` and cover cross-owner rejection in `openmontage-runner/tests/test_runner.py`.
- [X] T072 [US12] Add the runner service and durable workspace volume to `docker-compose.local.yml`, `docker-compose.deploy.yml`, and `docker-compose.build.yml` without publishing its port.
- [X] T073 [US12] Replace root MIT release statements with the selected AGPLv3 notices, preserve BeefTV attribution, add the OpenMontage source offer and third-party record in `LICENSE`, `NOTICE`, `README.md`, and `THIRD_PARTY_NOTICES.md`.
- [X] T074 Run runner fixture tests, native manifest validation, relevant Go tests, web build, and Docker Compose configuration validation; record actual results in `specs/001-episode-editing-workbench/quickstart.md`.

## Phase 15: Native Cinematic Production Execution

**Goal**: An approved workbench task runs the vendored OpenMontage `cinematic` pipeline against
the owner's real source videos. It creates schema-valid native artifacts, stops at every manifest
approval gate, renders a real final video, validates it with FFprobe, and returns the final video
and render report to BeefTV as owned resources.

**Independent Test**: A production seeded with two real local video files advances through
`research -> proposal -> script -> scene_plan -> assets -> edit -> compose -> publish`, requires
approval at every declared native gate, and produces a distinct H.264/AAC MP4 plus a valid native
`render_report`.

- [X] T075 [US13] Replace the internal Runner source-upload contract with owner-scoped BeefTV resource references and production capability checks in `openmontage-runner/app/main.py`, `openmontage-runner/app/native_workspace.py`, and `openmontage-runner/tests/test_runner.py`.
- [X] T076 [US13] Implement the native source-led cinematic executor in `openmontage-runner/app/native_workspace.py`: source inspection, schema-valid artifacts, manifest gate enforcement, FFmpeg composition, FFprobe verification, and final output/report projection.
- [X] T077 [US13] Add a typed backend OpenMontage runner client and source-resource handoff that never exposes internal runner credentials to the browser.
- [X] T078 [US13] Persist and project the OpenMontage production identity, current checkpoint, gate/action, blockers, final resource and render-report resource on the owner-scoped drama-edit task.
- [X] T079 [US13] Route approval/start/status/cancel and native checkpoint approval through the narrow drama-editing port and public handlers; preserve `timeline_render` only as an explicitly labeled legacy fast path.
- [X] T080 [US13] Update the workbench typed API and UI to show native pipeline state, stage approvals, terminal render links, and truthful unavailable/blocking states.
- [X] T081 [US13] Build and run the Runner together with local BeefTV, then execute the complete real-source `67906322` production through every required approval; retain FFprobe and native checkpoint/report evidence in `quickstart.md`.

**Non-negotiable execution rule**: T077-T081 replace the automatic `timeline_render` path for
new drama-edit approvals. A Runner-only result is not completion. Mark T081 complete only after
the final video and render report are imported into BeefTV as owner-scoped resources and are
visible from the original workbench task.

**Asset ownership rule**: T075-T081 must be revised during implementation so the Runner receives
only authenticated BeefTV resource references, never copied/uploaded source video bytes. BeefTV
remains the only asset store; the generated final video is registered under the originating drama
folder before task completion.

## Phase 16: Backend-Owned Source Approval Correction

**Goal**: Approving a task whose episode sources are already owned by BeefTV must not depend on a
browser-local asset lookup or its request timeout.

- [X] T082 [US13] Remove the browser source-handoff path from approval and start; the owner-scoped
  backend API alone resolves canonical BeefTV `resource:` references in
  `web/src/pages/drama-editing/index.tsx`.

## Phase 17: Native Production Control Timeout Correction

- [X] T083 [US13] Exempt only approval/start/native-stage production controls from the shared
  four-second browser JSON timeout in `web/src/services/api/drama-editing.ts`, because compose
  and publish synchronously execute real OpenMontage/FFmpeg work before returning their state.

## Phase 18: Maximum Duration Boundary Correction

- [X] T084 [US13] Reserve one output frame during native FFmpeg concatenation and reject any
  verified output above its approved target duration; cover the boundary with the Runner's
  real-video integration test in `openmontage-runner/app/native_workspace.py` and
  `openmontage-runner/tests/test_runner.py`.

## Phase 19: Active Native Task Deletion

**Goal**: A creator can delete an approved or processing task without leaving an OpenMontage
production running after its BeefTV task is removed.

**Independent Test**: A processing task calls the owner-scoped native cancellation port before
deletion; a failed stop retains that task while other batch entries proceed.

- [X] T098 [P] [US18] Add domain coverage for acknowledged and failed OpenMontage cancellation before active task deletion in `backend/internal/dramaedit/service_test.go`.
- [X] T099 [US18] Stop an approved or processing owner-scoped OpenMontage production and delete only after it reports `cancelled` in `backend/internal/dramaedit/service.go`.
- [X] T100 [US18] Keep native production-control timeout handling for active deletion and tell creators when their selected tasks will be stopped first in `web/src/services/api/drama-editing.ts` and `web/src/pages/drama-editing/index.tsx`.
- [X] T101 [US18] Rebuild and restart affected `backend` and `web` services; validate a real active-task deletion through the browser and record results in `specs/001-episode-editing-workbench/quickstart.md`.

## Phase 20: Native Compose Failure Integrity

- [X] T102 [US13] Keep the native `render_report` schema-valid while preserving editorial details under `metadata.editorial_report`, persist compose failures in the Runner, and project the resulting blocked reason through BeefTV polling; cover both paths with Runner and Go tests.
- [X] T103 [US13] Rebuild and restart `openmontage-runner`, `backend`, and `web`, then complete one real-source browser production through publish and BeefTV delivery import; record the checkpoint, output probe, and report-modal evidence in `quickstart.md`.

## Phase 21: In-Page Delivery Playback

- [X] T104 [US14] Constrain the in-page delivery player to a compact responsive modal and make every close path pause, clear, and unload the HTML video element in `web/src/pages/drama-editing/index.tsx` and `web/src/pages/drama-editing/drama-editing.css`; verify it in a real browser after rebuilding `web`.

## Phase 22: Source Candidate and Asset Workspace Clarity

**Goal**: Creators can select only original episode sources, while the asset workspace exposes a concise default-visible folder view.

**Independent Test**: A folder with seven source videos plus OpenMontage outputs reads as seven episodes; the asset page exposes the requested sidebar and direct actions without quick filters or a source rail.

- [X] T105 [P] [US19] Add the seven-source-plus-OpenMontage-delivery candidate regression case in `backend/internal/repository/asset_library_test.go`.
- [X] T106 [US19] Apply the shared persisted OpenMontage-delivery provenance predicate to episode-folder counts and returned video rows in `backend/internal/repository/asset_library.go`.
- [X] T107 [P] [US19] Rename visible short-drama creation controls and modal title to “新建任务” in `web/src/pages/drama-editing/index.tsx`.
- [X] T108 [US20] Default the asset sidebar open; remove its quick-filter block and source-navigation rail; rename the folder group; and expose visible folder/upload actions in `web/src/pages/assets/index.tsx`.
- [X] T109 [US20] Remove obsolete source-rail/new-menu style behavior in `web/src/styles/workspace-product.css` if it is no longer referenced.
- [X] T110 Run focused repository tests, `cd web && bun run build`, rebuild and restart `backend` and `web`, then complete browser validation and record evidence in `specs/001-episode-editing-workbench/quickstart.md`.

## Phase 23: Backend-Readable Episode Sources

**Goal**: A video becomes a short-drama source only after BeefTV owns readable bytes; neither a
browser Blob nor metadata-only mirror can enter a task or reach OpenMontage.

**Independent Test**: Upload two real videos to a folder in browser-local mode, verify their asset
payloads use `resource:` keys, create and approve a task, and verify the Runner receives only those
BeefTV resource IDs. A legacy Blob-only video is excluded when its bytes cannot be recovered.

- [X] T111 [US21] Add historical-resource-backfill coverage through the browser-local asset persistence path in `web/src/services/workspace-asset-repository.ts`.
- [X] T112 [US21] Add owned-ready-resource candidate count/list and stale-payload exclusion regression coverage in `backend/internal/repository/asset_library_test.go`.
- [X] T113 [US21] Make browser-local video uploads use the owner-scoped Resource API, preserve local recovery only on failure, and require `resource:` persistence for video assets in `web/src/services/file-storage.ts` and `web/src/services/workspace-resource-storage.ts`.
- [X] T114 [US21] Backfill readable historical local video bytes before episode-folder lookup and rewrite successful assets with their owned resource references in `web/src/services/workspace-asset-repository.ts`.
- [X] T115 [US21] Apply the same owned-ready-video predicate to episode folder counts and candidate rows in `backend/internal/repository/asset_library.go`.
- [X] T116 [US21] Run focused tests, Go repository/domain tests, `cd web && bun run build`; rebuild/restart changed services and complete real two-video upload-to-Runner admission verification in `specs/001-episode-editing-workbench/quickstart.md`.

## Phase 24: Creator-Facing Task Labels

**Goal**: The navigation and task list use short-drama terminology without exposing native pipeline vocabulary or an imprecise progress bar.

**Independent Test**: A task at native `publish` is rendered with a Chinese creator-facing description, no `publish` text, and no task-card progress bar.

- [X] T117 [US22] Rename the drama-editing workspace navigation entry to “短剧剪辑” in `web/src/components/layout/workspace-sidebar-nav.tsx`.
- [X] T118 [US22] Remove the task-card progress control and translate native task stages to concise creator-facing Chinese descriptions in `web/src/pages/drama-editing/index.tsx`.
- [X] T119 [US22] Build the web bundle, rebuild/restart only the changed `web` Compose service, and verify the sidebar and task card in a real browser; record evidence in `specs/001-episode-editing-workbench/quickstart.md`.

## Phase 25: Compact Task Status

**Goal**: Task cards show production progress once in the right-side status tag and keep their left summary limited to task metadata.

**Independent Test**: A processing task has no stage text on the left and displays `制作中 N%` on the right using its current progress.

- [X] T120 [US23] Remove task-card stage copy and append the existing processing progress to the right-side status tag in `web/src/pages/drama-editing/index.tsx`.
- [X] T121 [US23] Build the web bundle, rebuild/restart only the changed `web` Compose service, and browser-verify a processing card's compact status; record evidence in `specs/001-episode-editing-workbench/quickstart.md`.

## Phase 26: Compact Task Creation

**Goal**: The creation dialog hides fixed spoiler policy and presents episode sources in a compact, discoverable five-column grid.

**Independent Test**: A seven-episode folder has five first-row entries; a long title is ellipsized with full hover text, and no spoiler control appears.

- [X] T122 [US24] Fix the creation form's hidden spoiler policy to `no-major-spoilers`, remove its visible control, and add full-title hover text to episode labels in `web/src/pages/drama-editing/index.tsx`.
- [X] T123 [US24] Render the episode chooser in five stable desktop columns with truncating labels and responsive narrow-screen tracks in `web/src/pages/drama-editing/drama-editing.css`.
- [X] T125 [US24] Reflow short-drama title, aspect ratio, language, platform, and duration controls into the compact five-track creation form in `web/src/pages/drama-editing/index.tsx` and `web/src/pages/drama-editing/drama-editing.css`.
- [X] T124 [US24] Build the web bundle, rebuild/restart only the changed `web` Compose service, and browser-verify the compact creation form; record evidence in `specs/001-episode-editing-workbench/quickstart.md`.
