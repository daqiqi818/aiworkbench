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

### Amended creation request

The following request shape replaces the manual episode/source and campaign fields for newly
created tasks:

```json
{
  "dramaTitle": "短剧名称",
  "deliveryPreferences": {
    "aspectRatio": "9:16",
    "minDurationSeconds": 15,
    "maxDurationSeconds": 300,
    "language": "zh",
    "platform": "douyin"
  },
  "spoilerPolicy": {
    "policy": "no-major-spoilers",
    "restrictedItems": []
  },
  "episodeAssetIds": ["video-episode-01", "video-episode-02"]
}
```

`episodeAssetIds` has 2-12 unique video asset IDs in the creator's desired episode order. Server
validation requires each ID to be owned by the current user and not archived. Multiple selected
episode videos may belong to one folder.
Language is one of `zh`, `en`; platform is one of `tiktok`, `facebook`, `wechat_channels`,
`douyin`; each duration is 15-300 seconds inclusive and minimum is not greater than maximum.

## `GET /drama-editing/episode-folders`

Returns `{ "folders": [...] }` for the current user. Each folder has `folderId`, `folderName`,
`videoCount`, and `videos`; each `videos` entry has `assetId` and `assetTitle`. Only folders with
one or more active non-archived video assets are returned, ordered by asset-folder ordering. This
endpoint supports selection display only; `POST /tasks` repeats ownership and availability checks.

### Browser-local asset preparation

Before this read in browser-local mode, the client mirrors its asset-folder and asset metadata to
the local workspace API. This uses the existing folder and asset contracts; it does not upload
media bytes or alter the response shape. A preparation failure is surfaced to the creator and is
not represented as an empty folder response.

### Creation submission behavior

The client sends `POST /drama-editing/tasks` only after its required fields are valid, episode
candidate preparation has succeeded, and at least two video episodes are selected. Client-side validation
or preparation failures keep the request unsent and are shown as actionable dialog feedback.

## `GET /drama-editing/tasks/:id`

Returns the owned task, current proposal, and decision history. A task belonging to another user is
not disclosed.

## `POST /drama-editing/tasks/:id/approve`

Records explicit approval. Valid only while awaiting approval or after a revision request. Returns
the updated task with its linked local production work. Approval automatically starts production
when all selected source videos can be read.

### Browser-local production preparation

Before `approve` or `start`, a browser-local client transfers each selected source that lacks a
`resource:` storage key through the existing owner-scoped `POST /resources` upload contract and
persists the corresponding `PUT /assets/:id` metadata update. The drama-editing endpoints remain
unchanged and are called only after this preparation succeeds. A local source read or upload
failure is returned by the client as an actionable preparation error and no production work is
created.

## `POST /drama-editing/tasks/:id/start`

Starts a previously approved task that has no linked production work. This is intended for a
historical approved task created before automatic production existed. It is owner-scoped and
idempotent: an existing linked work is returned instead of creating another one.

## `POST /drama-editing/tasks/:id/revision`

Requires a non-empty `note` up to 1,000 characters. Records the request and returns a refreshed
proposal with status `revision_requested`.

## `POST /drama-editing/tasks/:id/cancel`

Optionally accepts a note. Records cancellation and returns the task with status `cancelled`.
## OpenMontage Production Contract

## Short-Drama Operations Amendment

| Operation | Intent | Required outcome |
| --- | --- | --- |
| `DELETE /drama-editing/tasks` | Delete one or more owner tasks | Receives `{ "ids": ["DE_..."], "deleteEpisodeFiles": false }`; before deleting an active native production, stops it through the internal Runner and proceeds only after `cancelled` is acknowledged. |
| `GET /drama-editing/tasks` | Read task summaries | Returns tasks ordered by creation time descending, with projected status, progress and current/last native stage. |
| Delivery report resource | Read creator report | The owned report contains `editorial_report`; absent subtitles or editorial data are represented as unavailable rather than generated. |

`deleteEpisodeFiles` defaults to `false`. When true, only the selected tasks' source episode
assets are candidates; delivery video, report and folders are excluded. The service rechecks
remaining short-drama-task references, then delegates asset deletion to the existing owner-scoped
asset deletion policy. A protected source asset is returned as retained with its reason without
rolling back independent task deletion.

The published workbench adds production endpoints beneath the existing drama-editing task owner.
All responses retain the normal business envelope and all identifiers are owner-scoped.

| Operation | Intent | Required outcome |
| --- | --- | --- |
| `POST /drama-editing/tasks/:id/approve` | Create and start the active native `cinematic` production | Returns the task projection with runtime stage, status and opaque production ID; refuses unavailable runtime or invalid source ownership. |
| `GET /drama-editing/tasks/:id` | Read projected native production state | Returns stage, checkpoint status, required approval, blockers and delivery links. |
| `POST /drama-editing/tasks/:id/openmontage/stages/:stage/approve` | Record a user approval at a required native gate | Appends an approval and permits only the next manifest-valid stage. |
| `POST /drama-editing/tasks/:id/cancel` | Stop the owner’s active production | Cancels only that scoped run and preserves native checkpoint history. |
| `GET /openmontage/capabilities` | Read the current published capability envelope | Returns supported pipeline versions and unavailable reasons without credentials. |

The internal runner contract is not public. It receives an authenticated immutable production
identity, owner identity, runtime version, opaque workspace key, input resource references, budget
and allowed capability set. It returns only checkpoint-derived state, safe artifact metadata,
stage-gate requirements, and opaque delivery artifact references.

`GET /internal/openmontage/productions/:productionID/sources/:resourceID` is Runner-only. It
requires the persisted per-production `X-BeefTV-Production-Token`, verifies that the requested
resource belongs to the task owner and selected episodes, and streams it without registering a
copy in the Runner.

## Episode Candidate Provenance Amendment

`GET /drama-editing/episode-folders` excludes a video only when its persisted asset payload identifies it as an
OpenMontage delivery (`data.source = "OpenMontage"`). Both `videoCount` and `videos` apply this exact condition.
The endpoint continues to return owned, non-archived source videos and task creation rechecks ownership and
availability for all submitted IDs.

## Task Card Presentation Amendment

No HTTP contract changes. `stage` and `openMontageStage` remain available for task detail, polling, and stage
approval; list cards must translate native stage codes to creator-facing wording and must not render a progress bar.

## Compact Task Status Amendment

No HTTP contract changes. A processing task card renders the existing `productionProgress` in its right-side status label and does not render `stage` or `openMontageStage` in its left-side summary.

## Compact Task Creation Amendment

No HTTP contract changes. `POST /drama-editing/tasks` continues receiving `spoilerPolicy`; the creation form always
sends the existing `no-major-spoilers` value. Episode display truncation affects neither the submitted asset IDs nor
their order.
