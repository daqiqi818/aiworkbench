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

## Amendment: Asset-Folder Episode Workflow

### Revised Drama Edit Task Fields

| Field | Rules |
| --- | --- |
| `dramaTitle` | Required, trimmed, 1-200 characters. |
| `campaignObjective` / `targetAudience` | No longer accepted for new tasks or displayed in the new-task workflow. Historical persisted values may remain unread by new UI. |
| `delivery.minDurationSeconds` | Required integer, 15-300 inclusive. |
| `delivery.maxDurationSeconds` | Required integer, 15-300 inclusive and greater than or equal to `minDurationSeconds`. |
| `delivery.language` | Required enum: `zh` or `en`. |
| `delivery.platform` | Required enum: `tiktok`, `facebook`, `wechat_channels`, or `douyin`. |
| `episodes` | Ordered JSON set of 2-12 unique selected video snapshots. Each snapshot has `folderId`, `folderName`, `assetId`, `assetTitle`, and `sequenceNumber`. Videos from the same folder remain distinct episodes. |

### Episode Folder Candidate

This is a transient owner-scoped read model, never a browser-supplied authority.

| Field | Rules |
| --- | --- |
| `folderId` | Current user's asset-folder ID. |
| `folderName` | Current folder name, sorted by current folder position. |
| `videoCount` | Count of active video assets assigned to the folder; must be greater than zero. |
| `videos` | Active non-archived video candidates, each with `assetId` and `assetTitle`. |

At creation time the service re-reads selected video assets, rejects missing, foreign, archived,
duplicate, empty, or more-than-12 choices, and writes snapshots in submitted order.

## Amendment: Browser-Local Metadata Mirror

Browser-local media remains a scoped local storage object. Its corresponding asset record and
folder relationship are mirrored to the local backend before an episode candidate read. A legacy
browser-only folder is matched to an owned backend folder by normalized name; all affected local
asset `folderId` values are rebound to that canonical ID before metadata is mirrored. The mirror
does not add a media-byte field or a duplicate media object.

## Amendment: Submission Read Model

The create dialog has transient submission state: `idle`, `validating`, `submitting`, and candidate
read state. It does not alter persisted task data. A request is permitted only when the read state
is ready, all required delivery fields validate, and the ordered episode selection contains at
least two videos.

## Amendment: Production Execution

`DramaEditTask` additionally records a nullable `productionTaskId`, `deliveryResourceId`, and
`productionError`. The public projection includes the linked work status and progress. The source
episode snapshot keeps the asset identity; source-resource ownership and readability are rechecked
when production starts. Lifecycle projection is: approved with no work -> approved; queued/running
work -> processing; succeeded work with delivery resource -> completed; failed work or failed
admission -> blocked.

## Amendment: Production Source Handoff

A browser-local video asset may initially contain a scoped local `data.storageKey`. Before a
production request, the client may replace it with `resource:<resourceId>` after the source Blob
has been uploaded. The `resourceId` is not copied into `DramaEditTask`; the task retains its
asset ID and production admission resolves the current owned ready video resource through either
the asset's output representation or this resource storage key.
