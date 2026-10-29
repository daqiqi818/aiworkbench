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
