# Data Model: Multi-Episode Editing Workbench

## Drama Edit Task

| Field | Rules |
| --- | --- |
| `id` | Server-generated stable local ID. |
| `userId` | Required owner; every read and mutation filters by it. |
| `dramaTitle` | Required, trimmed, 1-200 characters. |
| `campaignObjective` / `targetAudience` | Required, trimmed, each 1-1,000 characters. |
| `delivery` | Required aspect ratio (`9:16`, `16:9`, `1:1`), 10-180 second target, language, and destination label. |
| `spoilerPolicy` | Required policy and unique restricted items. |
| `episodes` | Ordered JSON set of 2-12 unique episode IDs; at least two are included; excluded entries require a reason; every source has a non-empty reference and confirmed rights. |
| `status` | `awaiting_approval`, `revision_requested`, `approved`, `blocked`, or `cancelled`. |
| `proposal` | Generated review snapshot with hook, emotional arc, coverage, delivery definition, cost estimate, and source limitations. |
| `revisionNote` | Required when requesting revision; retained in the decision log. |
| `blocker` | Optional structured attempted action, failure, impact, options, and recommendation. |
| `createdAt` / `updatedAt` | Server timestamps. |

## Drama Edit Decision

Append-only record associated with one task.

| Field | Rules |
| --- | --- |
| `id` / `taskId` / `userId` | Server-generated ID and owned task relationship. |
| `action` | `created`, `revision_requested`, `approved`, or `cancelled`. |
| `note` | Required for revision request; optional for cancellation. |
| `createdAt` | Server timestamp. |

## Relationships and transitions

One task owns one ordered episode set, one current proposal, and many decisions. Creation generates
a proposal and enters `awaiting_approval`. A revision request moves to `revision_requested` and
regenerates the proposal after retaining the note. Approval moves only `awaiting_approval` or
`revision_requested` to `approved`. Cancellation is allowed only before a terminal state. A future
executor alone may add production and completion states after storing a verified output reference.
