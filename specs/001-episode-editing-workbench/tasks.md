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

## Dependencies & Execution Order

`T001-T007` block all user stories. US1 (`T008-T014`) creates the aggregate and menu. US2
(`T015-T019`) depends on US1's aggregate. US3 (`T020-T022`) depends on list/read behavior from
US1 and can be completed after US2 because its page changes share the same file. Polish follows all
stories.

## Parallel Opportunities

- T002 can proceed while directories are created.
- The two backend test files in each story can be prepared independently of the frontend API module.
- T012 can proceed after the API contract is stable while backend tests are being written.

## Implementation Strategy

Implement T001-T014 first to establish a usable MVP. Add lifecycle controls next, then status
filters and presentation. Do not attach an unreviewed external rendering runtime; the approved
state is an honest integration boundary.
