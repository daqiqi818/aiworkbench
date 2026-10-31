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

## Published OpenMontage integration

**Decision**: Vendor the tracked OpenMontage repository unchanged at a pinned commit and run it
inside an internal `openmontage-runner` service. Do not copy its files into Go or web modules.

**Rationale**: Its cinematic manifest, skills, schemas, tools, project layout, and checkpoint
format form one production contract. Keeping them together makes upstream refreshes auditable and
lets production state remain truthful.

**Alternatives considered**:

- Reimplement cinematic stages in BeefTV: rejected because it would immediately diverge from the
  declared OpenMontage workflow.
- Start an ad-hoc local Python script from the backend: rejected for published multi-user use;
  it cannot isolate workspaces, scale rendering, or survive instance migration.

## Production control plane

**Decision**: Use BeefTV's current `agent-host` official pi SDK integration as the only production
control plane and let the runner expose canonical project state and approved tool capability.

**Rationale**: OpenMontage explicitly defines the coding agent as its orchestrator. Introducing a
second custom conversation or model loop would violate the current product runtime boundary and
create contradictory approval semantics.

## License and source boundaries

**Decision**: Publish the integrated product under AGPLv3 obligations and vendor only files tracked
at the selected upstream commit. Exclude `.env`, `.git`, `node_modules`, virtual environments,
project workspaces, cache, bytecode, and build output.

**Rationale**: The source repository is AGPLv3 while BeefTV is currently MIT; the publisher chose
AGPL-compliant distribution. A clean tracked snapshot is reproducible and avoids copying local
credentials or unrelated work-in-progress.

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
# Active Native Deletion Decision

- **Decision**: Reuse the existing owner-scoped OpenMontage cancellation capability from the
  drama-edit service before deletion of an `approved` or `processing` task.
- **Rationale**: The native Runner owns canonical production state and can preserve its checkpoint
  cancellation history. Calling it from the service retains credential and ownership boundaries.
- **Alternatives considered**: Browser-to-Runner cancellation was rejected because it exposes an
  internal capability; deleting first was rejected because it can orphan a continuing production;
  requiring a separate cancel action was rejected because the delete command can safely compose the
  two existing operations.

## Source Candidate and Asset Workspace Clarity

**Decision**: Exclude only assets whose persisted delivery provenance is `data.source = "OpenMontage"`
from the episode-folder query, and use this identical rule in both aggregate and detail reads.

**Rationale**: Imported final videos intentionally live beside their source episodes. Provenance remains correct
after a creator renames either item; file-name matching and general generated-asset flags do not.

**Alternatives considered**:

- Exclude videos by a title suffix such as `成片`: rejected because titles are user-editable.
- Exclude all generated assets: rejected because a creator may deliberately use a generated source video.

## Asset Workspace Controls

**Decision**: Keep current folder creation and upload callbacks, expose them as two visible buttons, and remove
only the redundant quick-filter and source-rail presentation blocks.

**Rationale**: This meets the requested information hierarchy without changing upload persistence, folder ownership,
or any direct legacy history URL.

## Creator-Facing Task Labels

**Decision**: Keep native stage identifiers in the task data for approval and polling, but map them to concise Chinese task-card descriptions at the presentation boundary. Remove the task-card progress control.

**Rationale**: The native stage is operational metadata, while creators need a readable explanation. Removing it from the API would break stage approval; showing it directly exposes implementation vocabulary such as `publish`.

**Alternatives considered**:

- Remove all task-card stage text: rejected because a concise current-action explanation remains useful while a task is processing.
- Rename native Runner stages: rejected because it changes the production protocol for a presentation-only requirement.

## Compact Task Status

**Decision**: Remove the card's secondary stage line entirely. Reuse the existing task progress value only in the right-side `processing` status tag, with the same fallback already used by the detail progress display when a running task has not yet reported a value.

**Rationale**: This keeps the card scan-friendly and presents progress once, at the established state indicator, without changing production data or poll behavior.

**Alternatives considered**:

- Add another visual progress component to the status area: rejected because the requested compact presentation calls for a text percentage rather than a bar.
- Remove progress from processing cards altogether: rejected because the requested state label explicitly needs the current percentage.

## Compact Task Creation

**Decision**: Keep the existing default spoiler policy as a hidden fixed creation value and overwrite restored preferences at both form initialization and request submission. Use a five-column episode checkbox grid at the existing desktop dialog width, with text overflow ellipsis and a native hover title on each episode name.

**Rationale**: Hidden defaults remove an unnecessary decision without loosening the server contract. Stable grid tracks preserve a compact layout, while the browser-native title makes truncation reversible without adding another control.

**Alternatives considered**:

- Retain the selector but hide only its label: rejected because it remains a visible, unsupported choice.
- Allow card widths to expand for long names: rejected because one long filename would disrupt the requested five-item row.

## Compact Creation Controls

**Decision**: Use the five desktop grid tracks already required for source selection in the settings row: the short-drama name spans two tracks and aspect ratio, language, and platform use one track each. Duration controls move to the following compact row.

**Rationale**: This gives the title enough room while keeping the small enum controls close together and reduces dialog height without reducing readable control widths.

**Alternatives considered**:

- Use a single horizontal row for every field: rejected because duration inputs would make controls too narrow at the existing dialog width.
- Increase the dialog width: rejected because compactness is better achieved through layout, not a wider overlay.
