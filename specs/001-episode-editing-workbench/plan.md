# Implementation Plan: Multi-Episode Editing Workbench

**Branch**: `001-episode-editing-workbench` | **Date**: 2026-10-29 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-episode-editing-workbench/spec.md`

## Summary

Add a local, independently navigable multi-episode editing workbench. A creator can create a
short-drama editing task from 2-12 ordered episodes, review a deterministic proposal before
approval, record revision or cancellation decisions, and inspect its current status. The backend
owns validation, authorization, lifecycle transitions, and SQLite persistence; the frontend owns
the workbench view and invokes the typed API. The first increment records an approved task as ready
for a future rendering executor rather than claiming that an external OpenMontage process rendered
a video.

## Technical Context

**Language/Version**: Go 1.25 backend; TypeScript with React 19 frontend

**Primary Dependencies**: Gin, GORM, SQLite; Vite, React Router, Ant Design, TanStack Query,
Zustand, Lucide

**Storage**: Local SQLite through the managed local schema; no browser-only source of truth

**Testing**: Go package tests and `go test ./...`; frontend build with Bun

**Target Platform**: Local BeefTV desktop workspace and local backend

**Project Type**: Local web application with a Go HTTP API

**Performance Goals**: List and open a task locally within 1 second for 100 task records; submit a
valid task without a blocking external request

**Constraints**: User-scoped data; no upstream credentials or media bytes in task payloads; 2-12
episodes per task; no rendering claim without an attached executor result

**Scale/Scope**: One local workspace user; one short drama and one intended deliverable per task

## Constitution Check

*GATE: Pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle | Plan response | Status |
| --- | --- | --- |
| Evidence and focused change | Uses existing Gin envelope, local schema, route registration, sidebar, and `http` patterns. | PASS |
| Explicit boundaries and contracts | Adds a domain port, handler adapter, typed API module, and isolated page. | PASS |
| Secure, owned, truthful data | Service validates owner IDs and transitions; only references and permission declarations persist. | PASS |
| Proportionate verification | Domain and handler tests plus targeted frontend build are planned. | PASS |
| Upstream-preserving evolution | Adds an isolated menu and local domain without importing the OpenMontage runtime. | PASS |

Post-design review: PASS. The contracts and data model keep rendering as an explicit future
integration, so no unsupported production status or deliverable is fabricated.

## Project Structure

### Documentation (this feature)

```text
specs/001-episode-editing-workbench/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── internal/model/models_drama_edit.go
├── internal/dramaedit/                 # User-owned task lifecycle and validation
├── internal/localapp/                  # Narrow DramaEditingPort
├── internal/handler/drama_edit.go      # HTTP binding and envelopes
└── internal/database/local_schema.go    # v14 schema migration

web/src/
├── pages/drama-editing/                 # Page and page-private stylesheet/components
├── services/api/drama-editing.ts         # Typed, unwrapped business API
├── components/layout/workspace-sidebar-nav.tsx
├── lib/workspace-route-modules.ts
└── router.tsx

docs/content/docs/backend/backend-database.mdx
```

**Structure Decision**: Keep the feature self-contained under `internal/dramaedit` and
`pages/drama-editing`. The shared local application port and runtime dependency adapter preserve
the established handler-to-domain dependency direction. The source reference is a string, not an
OpenMontage checkout path or copied Python runtime.

## Complexity Tracking

No constitution violations or additional project roots are required.

## Amendment: Asset-Folder Episode Workflow

### Summary

Replace manual episode/source entry with user-owned asset folders that contain active video assets.
Each selected folder is an episode. The creation form accepts a duration range, enum-restricted
language and platform values, and selected folder IDs. It no longer accepts campaign objective or
target audience.

### Technical Context Changes

**Persistence**: Preserve the existing task and decision tables. A v15 additive schema migration
removes no historical columns and uses the existing JSON columns for the new delivery and episode
snapshots. Existing records remain readable; new records use the amended request shape.

**Validation**: Service-level validation reloads requested folders and active video asset counts
from the repository under the current user scope. The request must select 2-12 unique folder IDs;
the submitted order is authoritative. Minimum duration is 15-300 seconds, maximum duration is
15-300 seconds, and minimum cannot exceed maximum.

**Presentation**: The drama-editing page reads folder candidates via a typed API endpoint, selects
all initial candidates by default, and displays a compact checkbox list with per-folder video
counts. The assets page renders an inline new-folder input with explicit confirm/cancel controls.

### Constitution Check Amendment

| Principle | Plan response | Status |
| --- | --- | --- |
| Explicit boundaries and contracts | Folder lookup remains in the domain service and repository; handler only binds the request. | PASS |
| Secure, owned, truthful data | Candidate listing and task creation filter folder and assets by current owner; snapshots avoid media bytes and credentials. | PASS |
| Proportionate verification | Add focused folder-candidate, validation, and creation-path tests before backend implementation; run the frontend build. | PASS |
| Upstream-preserving evolution | Extends the isolated workbench and existing asset API without replacing upstream systems. | PASS |

### Project Structure Changes

```text
backend/
├── internal/repository/asset_library.go       # owner-scoped video-folder candidate query
├── internal/dramaedit/                        # amended request, snapshots, validation, tests
├── internal/handler/drama_edit.go             # folder candidate route
└── internal/database/local_schema.go           # v15 additive migration

web/src/
├── pages/assets/index.tsx                      # usable new-folder inline editor
├── pages/drama-editing/                        # compact folder-driven creation form
└── services/api/drama-editing.ts               # folder candidate and revised task types
```

## Amendment: Browser-Local Asset Metadata Bridge

### Summary

Keep browser-local media blobs in their current scoped IndexedDB storage while mirroring only
folder and asset metadata to the local backend. This gives the existing owner-scoped episode-folder
query visibility into folders such as `67906322` without treating a browser-local blob as a new
remote upload.

### Technical Context Changes

**Synchronization**: Before the episode-folder query, migrate any browser-local folder records by
normalized name to the local backend, rebind local asset `folderId` values to the resulting backend
IDs, then upsert the metadata for non-entity local assets. New browser-local folders use the same
backend folder API immediately; uploads and folder moves flush their metadata after local storage
has succeeded.

**Data integrity**: The backend remains the authority for folder ownership and candidate validation.
The bridge writes no media bytes and preserves browser-local storage keys, so an unavailable blob
does not cause metadata deletion. A failed metadata bridge propagates to the task-creation view
instead of appearing as an empty candidate list.

### Constitution Check Amendment

| Principle | Plan response | Status |
| --- | --- | --- |
| Explicit boundaries and contracts | The bridge is isolated in existing workspace asset services and uses typed `http` APIs; the workbench remains a consumer of the domain endpoint. | PASS |
| Secure, owned, truthful data | Backend folder ownership remains enforced; the bridge sends no media bytes, credentials, or fabricated resource references. | PASS |
| Proportionate verification | Add focused browser-local migration tests and run the web build plus targeted backend tests. | PASS |

### Project Structure Changes

```text
web/src/
├── services/workspace-asset-folders.ts       # browser-local folder migration and canonical IDs
├── services/workspace-asset-repository.ts    # metadata-only local asset bridge
└── pages/drama-editing/index.tsx             # synchronize before loading episode candidates
```

## Amendment: Task Submission Feedback

### Summary

Treat client-side form validation and candidate-query state as explicit submission outcomes. The
creation action will report missing fields, current episode count, loading, or sync errors before
making a request; only a valid, settled form issues the existing creation request.

### Technical Context Changes

**Presentation**: Keep values in the Ant Design form on failure, show field-level validation plus
one concise action message, and lock the confirmation action while submission is in progress.
Successful responses continue through the existing close, selected-task, and list-refresh path.

### Constitution Check Amendment

| Principle | Plan response | Status |
| --- | --- | --- |
| Proportionate verification and user-facing quality | Exercise invalid, loading, insufficient-episode, and valid submission paths; preserve exactly-once UI submission behavior. | PASS |

## Amendment: Approved Production Execution

### Technical Context Changes

**Execution**: Approval creates an idempotent local timeline-render work item from the ordered,
owned source videos. The timeline consumes sources in selection order until the configured maximum
duration, and admission rejects missing resources or a total duration shorter than the configured
minimum.

**Durability**: The editing aggregate persists the production work ID, output resource ID, and
last production error. On every list/detail projection, the aggregate maps the durable worker row
to `processing`, `completed`, or `blocked`; it never trusts a browser-held progress flag.

**Historical recovery**: An approved task without a production work, such as `1222`, has an
explicit idempotent start endpoint. A failed work can be started again only after its old work has
reached a terminal state.

### Constitution Check Amendment

| Principle | Plan response | Status |
| --- | --- | --- |
| Explicit boundaries and contracts | The editing domain accepts a narrow production-starter port; task admission and FFmpeg remain owned by the task domain. | PASS |
| Secure, owned, truthful data | Source assets resolve through owned representations and ready resources; no media path or bytes are stored in the editing task. | PASS |
| Proportionate verification | Domain tests cover automatic admission, completed projection, and blocking; the web build verifies the status surface. | PASS |

## Amendment: Production Source Handoff

### Technical Context Changes

**Handoff**: In browser-local mode, the workbench resolves each selected asset from the scoped
asset store, reads its IndexedDB Blob, and uploads only sources that do not already have a
`resource:` storage key. The existing resumable resource uploader supplies a stable key based on
the immutable asset ID and original local storage key. After a successful upload, the same asset
metadata is persisted with the resource reference before production is requested.

**Admission**: The drama-edit source resolver first retains the existing asset-representation
lookup, then resolves remaining selected assets through an owned ready video resource named by the
asset payload's storage key. This supports browser-originated assets without adding a parallel
asset model or storing media paths in the drama-edit record.

### Constitution Check Amendment

| Principle | Plan response | Status |
| --- | --- | --- |
| Explicit boundaries and contracts | The page orchestrates the existing resource and asset APIs; the domain continues to own source validation and timeline construction. | PASS |
| Secure, owned, truthful data | Every upload and payload update is user-scoped; the domain only accepts ready owned video resources. | PASS |
| Proportionate verification | Add domain coverage for payload resource resolution and run the source-handoff UI flow against a real browser profile. | PASS |

## Amendment: Local Rendering Runtime

### Technical Context Changes

The final local backend image installs Alpine's `ffmpeg` package, which provides both `ffmpeg`
and `ffprobe` on the container `PATH`. No host volume or undocumented `CANVAS_FFMPEG_PATH` is
required. Deployment verification checks those binaries in the running container, then retries an
already source-ready task and polls its durable editing-task state to a terminal result.
