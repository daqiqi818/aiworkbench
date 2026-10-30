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
