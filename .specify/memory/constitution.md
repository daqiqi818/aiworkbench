<!--
Sync Impact Report
- Version change: unversioned template -> 1.0.0
- Modified principles: none; initial constitution adopted from the resolved scaffold
- Added sections: Core Principles, Technical and Data Constraints, Development Workflow
- Removed sections: none
- Follow-up TODOs: TODO(RATIFICATION_DATE): original adoption date is not recorded in the repository.
-->
# BeefTV Constitution

## Core Principles

### I. Evidence and Focused Change
Every conclusion and change MUST be traceable to current code, configuration, tests, logs, or
documentation. Contributors MUST inspect affected entry points, callers, contracts, and adjacent
tests before editing. Changes MUST remain within the requested ownership boundary; unrelated
refactors, generated artifacts, lockfile churn, and another contributor's changes are prohibited.
This keeps a fast-moving workspace reviewable and prevents historical assumptions from becoming
implementation decisions.

### II. Explicit Boundaries and Contracts
Frontend modules MUST follow the established page, component, service, store, and library
boundaries. Backend requests MUST flow through handler, local application or domain port, domain
implementation, and repository or model layers; handlers MUST NOT contain database queries.
Business API calls MUST use the shared `http` client and its envelope/error contract. New helpers
MUST isolate a real protocol or remove meaningful duplication. This makes ownership, error flow,
and future replacement of local extensions explicit.

### III. Secure, Owned, and Truthful Data
Every object read, update, upload, generation, activation, approval, permission, deletion, quota,
billing, or key operation MUST validate the current user and resource ownership in the service
layer. Sensitive URLs, cookies, and API keys MUST NOT enter logs, URLs, local storage, or durable
task payloads. Private upstreams are denied unless precisely allowlisted. Local cache success MUST
be distinguished from server persistence, and an asset record MUST remain when its physical object
cannot be removed. These rules protect user data and preserve reliable recovery semantics.

### IV. Proportionate Verification and User-Facing Quality
Each change MUST receive the smallest sufficient verification for its risk and blast radius.
Contract, persistence, permissions, task, streaming, and resource changes MUST include focused
tests or an equivalent executable check. UI changes that can be run in a browser MUST inspect the
relevant route, themes, scrolling, overlays, empty states, and primary interaction. Results and
unrun checks MUST be reported accurately; static review is not runtime verification. This prevents
unverified behavior from being represented as working product behavior.

### V. Upstream-Preserving Product Evolution
Secondary development MUST retain a practical path to synchronize BeefTV upstream changes.
The multi-episode short-drama editing workbench MUST be developed as an additive main-menu
capability, using the existing stack and local extension mechanisms rather than replacing product
subsystems without a documented migration. New product work MUST begin with Spec Kit in the order
`spec -> plan -> task -> implement`. This maintains both product direction and sustainable
integration with the upstream project.

## Technical and Data Constraints

The web application MUST use Vite, React, TypeScript, React Router, Ant Design, Tailwind, Zustand,
and TanStack Query according to existing local patterns. The backend MUST use Go, Gin, GORM, and
SQLite for the local workspace unless an approved architecture change states otherwise. Canvas
state, algorithms, and components MUST remain in their designated canvas directories; large data
MUST use scoped localforage rather than `localStorage`.

API responses MUST retain the `{ code, data, msg, reason }` envelope. Binary or non-envelope
responses MUST use `http.raw`; write failures MUST propagate rather than silently returning a
default. Database schema changes MUST update the backend database documentation. Local development
MUST use the ignored `.local/project-workbench-debug` data directory through
`CANVAS_BACKEND_DATA_DIR`; committed databases, uploads, environment files, real keys, and build
artifacts are prohibited.

## Development Workflow

Before implementation, contributors MUST inspect the relevant files and check the working tree
without reverting or cleaning unrelated modifications. Feature work MUST complete the Spec Kit
sequence `speckit-specify`, `speckit-plan`, `speckit-tasks`, and `speckit-implement` before code is
changed. Documentation belongs in the relevant topical documentation when APIs, schemas, SSE,
storage, deployment, or security boundaries change.

Verification MUST use the narrowest applicable project command: frontend builds or focused Bun
tests for web work, `go test ./...` for backend work, and documentation type checks or builds for
documentation work. Development servers MUST remain stopped unless the user explicitly requests a
browser preview or integration check. Commits MUST use
`<type>(<scope>): <business module> - <change summary>` with an approved type.

## Governance

This constitution supersedes informal development preferences and governs repository work together
with explicit user requests and applicable security requirements. A proposed amendment MUST explain
the affected principles, its compatibility impact, the semantic version change, and any required
migration or documentation updates. Maintainers MUST review changed work for boundary, ownership,
security, documentation, and verification compliance before approval.

Constitution versions use semantic versioning: MAJOR for incompatible removal or redefinition of a
governance rule, MINOR for a new rule or material expansion, and PATCH for clarification or
non-semantic editorial refinement. Every amendment MUST update the Sync Impact Report for human
review before commit and the `Last Amended` date. This document is the governing source; runtime
development details remain in `AGENTS.md` and the relevant repository documentation.

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): original adoption date is not recorded | **Last Amended**: 2026-10-29
