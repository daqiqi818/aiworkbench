# Research: Multi-Episode Editing Workbench

## Domain ownership

**Decision**: Add a small `internal/dramaedit` domain backed by local SQLite and expose it through
a `localapp` port.

**Rationale**: Existing generation tasks represent provider execution and must not be overloaded
with product approval state. A distinct aggregate lets the service enforce user ownership and
transitions while keeping its JSON payload free of upstream credentials.

**Alternatives considered**:

- Store drafts only in browser storage: rejected because review and status must survive refresh and
  remain distinct from a failed server save.
- Add fields to generic `Task`: rejected because generic task workers would treat the new type as a
  provider job and conflate task execution with creative approval.

## Proposal and execution boundary

**Decision**: Generate a reviewable proposal from submitted campaign data, then leave approved
tasks in `approved` until an explicitly designed rendering executor is attached.

**Rationale**: The OpenMontage adapter validates and projects a workflow but does not provide an
embeddable Go rendering engine. Reporting completion without an executor or output violates the
truthful-data principle.

**Alternatives considered**:

- Run OpenMontage Python from the backend: rejected because it imports a separate runtime,
  dependencies, credentials, and data layout into the product without an approved deployment plan.
- Mark approval as completed: rejected because approval is not a delivered edit.

## State model

**Decision**: Persist `awaiting_approval`, `revision_requested`, `approved`, `blocked`, and
`cancelled` states, with action-specific transition checks.

**Rationale**: These are directly observable states required by the user flow. A future executor
can add `in_progress` and `completed` only when it writes an actual output reference.

**Alternatives considered**:

- Free-form status strings set by the browser: rejected because they cannot preserve approval
  gates or explain allowed next actions.

## Navigation and presentation

**Decision**: Add a lazy `/drama-editing` route and one sidebar item, using page-private UI.

**Rationale**: The feature is a workbench, not a variant of the generic task center. This keeps it
discoverable without reopening the disabled `/tasks` route.

**Alternatives considered**:

- Add a project-detail tab: rejected because a task may begin from episode references and the
  requested entry point is an independent menu.

## Asset folders as episodes

**Decision**: Treat each current user's flat asset folder containing at least one active video as
one selectable episode. Persist a snapshot of its ID, name, selected order, and video count in the
task.

**Rationale**: The current product has only flat user asset folders. This makes the creator's
existing episode organization directly usable, preserves ownership checks, and avoids inventing a
nested-folder model that no other asset UI supports.

**Alternatives considered**:

- Interpret child folders as episodes: rejected because asset folders have no parent relationship.
- Trust folder details sent by the browser: rejected because a folder may be removed, reassigned,
  or emptied between form load and task creation.

## Restricted delivery controls

**Decision**: Store `minDurationSeconds`, `maxDurationSeconds`, language enum `zh|en`, and platform
enum `tiktok|facebook|wechat_channels|douyin` in delivery preferences.

**Rationale**: Explicit enums make UI dropdown choices and server validation agree. A 15-300 second
inclusive range accommodates the requested limits and is clear in proposals.

**Alternatives considered**:

- Free-form language/platform text: rejected because it admits unsupported destinations.
- A single target duration: rejected because it cannot describe the requested delivery range.

## Asset-folder creation interaction

**Decision**: Render a focused inline editor and explicit confirm/cancel controls whenever the
asset page enters the new-folder state.

**Rationale**: Existing state and save logic are present, but the `new` state has no rendered
input or submit control, so both new-folder entry points appear inert.

**Alternatives considered**:

- Replace with a global modal: rejected because the existing sidebar uses inline rename and a
  local inline editor is faster for recurring episode-folder creation.

## Browser-local asset metadata bridge

**Decision**: Keep browser-local files in scoped IndexedDB and mirror only folder and asset
metadata through the existing local workspace API before reading episode candidates.

**Rationale**: The candidate endpoint already validates ownership and active-video counts. Giving
it metadata visibility preserves that authority without duplicating a video or adding a
browser-only task path.

**Alternatives considered**:

- Read IndexedDB directly from the workbench: rejected because it bypasses backend ownership
  validation and the persisted task contract.
- Re-upload each local video: rejected because it duplicates storage and violates the current
  browser-local media policy.
