<!--
Sync Impact Report
- Version change: unversioned scaffold -> 1.0.0
- Modified principles: none; replaced the unfilled template scaffold with five project principles.
- Added sections: Production Integrity; Development Workflow.
- Removed sections: none.
- Follow-up TODOs: RATIFICATION_DATE is unknown; confirm the original adoption date before commit.
-->
# OpenMontage Constitution

## Core Principles

### I. Instruction-Driven Orchestration
Agents MUST orchestrate production by reading the selected pipeline manifest and its stage director
skills. Python code MUST remain limited to tools, validation, and persistence; creative decisions,
workflow transitions, and review policy MUST remain declarative in instructions. This preserves
auditable, adaptable workflows instead of embedding agent behavior in opaque application code.

### II. Pipeline and Capability Integrity
Every video-production request MUST select a pipeline, pass registry-based preflight, and execute
one stage at a time with its required quality gates and approvals. Providers, models, render
runtimes, and fallbacks MUST be disclosed before consequential generation; an approved choice MUST
NOT be silently substituted. This makes deliverables reproducible and protects the approved creative
direction and budget.

### III. Canonical, Validated Artifacts
Each production MUST write its canonical artifacts, assets, checkpoints, and renders beneath
`projects/<project-id>/` using the project workspace convention. Artifacts MUST satisfy their
declared schemas before a stage advances, and output paths MUST be explicit. This keeps the Backlot
board, recovery workflows, and reviews aligned with the actual run state.

### IV. Explicit, Auditable Decisions
Material creative, provider, budget, and composition decisions MUST be recorded before execution.
When a recorded decision changes, a new append-only decision-log entry MUST reuse the same category
and subject and identify the superseded choice. This retains an intelligible history and prevents
the displayed production state from diverging from the delivered result.

### V. Verification Proportionate to Risk
Changes to tools, schemas, manifests, and runtime contracts MUST include focused automated tests or
validation that cover the changed behavior. Production outputs MUST undergo the applicable review
and delivery checks before release; higher-risk, cross-pipeline changes require broader integration
coverage. This keeps a multi-provider production system reliable without imposing unnecessary work
on isolated documentation changes.

## Production Integrity

Human approval gates defined by a pipeline are binding. Motion-dependent briefs MUST retain a
motion-capable treatment; static or lower-fidelity substitutes require explicit user approval.
When Remotion and HyperFrames are both available, the production proposal MUST present both
runtimes, their brief-specific tradeoffs, and a recommendation before locking a runtime. Cost
estimates, reservations, and reconciliations MUST use the project cost-tracking workflow.

## Development Workflow

Contributors MUST inspect existing conventions and the relevant contracts before modifying code.
New tools MUST inherit the project base tool contract, declare their capabilities for registry
discovery, return structured results, and avoid ad hoc orchestration. New pipelines MUST provide a
validated manifest, stage director skills, compatible style playbooks, and contract tests. Reviews
MUST verify constitution compliance alongside correctness, security, and regression risk.

## Governance

This constitution supersedes conflicting local development practices. Amendments MUST document the
proposed governance change, its rationale, its compatibility impact, and the required semantic
version bump. A MAJOR version removes or redefines an existing principle incompatibly; a MINOR
version adds a principle or materially expands governance; a PATCH version clarifies wording without
changing meaning. Every amendment MUST update the Sync Impact Report and receive compliance review
before commit. Pipeline manifests, skills, schemas, and reviews MUST be assessed against this
constitution whenever they are created or materially changed.

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): original adoption date is not recorded |
**Last Amended**: 2026-10-28
