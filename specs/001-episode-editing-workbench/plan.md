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
