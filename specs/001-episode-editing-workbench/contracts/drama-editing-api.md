# Drama Editing API Contract

All endpoints are local business JSON endpoints under `/api/drama-editing`. Successful responses
use `{ "code": 0, "data": ..., "msg": "ok" }`; failures use the established `code`, `msg`, and
machine-readable `reason` envelope. Every operation uses the current workspace owner.

## `GET /drama-editing/tasks`

Returns task summaries ordered by most recently updated. Optional `status` filters one lifecycle
state.

## `POST /drama-editing/tasks`

Creates one task. Required fields are `dramaTitle`, `campaignObjective`, `targetAudience`,
`deliveryPreferences`, `spoilerPolicy`, and `episodes`. `episodes` contains 2-12 ordered entries;
each has `episodeId`, `sequenceNumber`, `summary`, `inclusion`, and one or more sources with
`reference` and `rightsConfirmed: true`. Excluded entries also require `exclusionReason`.

Returns the created task, including its initial proposal and `awaiting_approval` status.

## `GET /drama-editing/tasks/:id`

Returns the owned task, current proposal, and decision history. A task belonging to another user is
not disclosed.

## `POST /drama-editing/tasks/:id/approve`

Records explicit approval. Valid only while awaiting approval or after a revision request. Returns
the updated task with status `approved` and next action explaining that execution integration is
pending.

## `POST /drama-editing/tasks/:id/revision`

Requires a non-empty `note` up to 1,000 characters. Records the request and returns a refreshed
proposal with status `revision_requested`.

## `POST /drama-editing/tasks/:id/cancel`

Optionally accepts a note. Records cancellation and returns the task with status `cancelled`.
