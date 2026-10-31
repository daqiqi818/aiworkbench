# Data Model: Multi-Episode Editing Workbench

## Short-Drama Operations Amendment

- **Task deletion result**: `id`, `deleted`, and optional `reason`. A task in `approved` or
  `processing` is retained until its production is cancelled. Deleting a task does not delete its
  BeefTV assets or delivery resources.
- **Creator report**: an owned delivery-report projection with per-source episode subtitle status,
  hook, ordered story clips, suspense status, and native workflow entries. Every source/final time
  range and rationale originates from the run; unavailable source subtitles or editorial markers
  remain explicitly unavailable.
- **Creation preference**: device-local, creator-scoped form state containing folder, selected
  videos, delivery settings and spoiler policy. It excludes `dramaTitle` and is validated against
  current folder candidates before use.

## Verified Deletion Amendment

- **Delete command**: `ids` is a non-empty set of at most 100 owned task IDs; `deleteEpisodeFiles`
  is false unless the creator explicitly selects it.
- **Task result**: retains `id`, `deleted`, and optional `reason`; adds individual source-file
  outcomes when source deletion was requested. A source asset is eligible only after all selected
  task records are removed and no remaining short-drama task references it.
- **Source file result**: identifies the episode asset and whether it was deleted or retained with
  a reference-safety reason. Delivery video and report assets are never candidates.

## Active Production Deletion Amendment

- **Active deletion precondition**: an `approved` or `processing` task must have an owned
  `openMontageProductionId` and an available native stop capability. The task record remains until
  that production returns terminal `cancelled`.
- **Active deletion result**: the existing task result returns `deleted: true` only after native
  cancellation is acknowledged. A failed or unavailable stop returns `deleted: false` with a reason;
  it does not mutate the task lifecycle or delete associated BeefTV assets.

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
## Amendment: OpenMontage Published Production

### `OpenMontageRuntimeVersion`

| Field | Meaning | Rules |
| --- | --- | --- |
| `id` | Immutable runtime version identity | Unique; never rewritten after publication. |
| `upstreamCommit` | Vendored OpenMontage source commit | Required, immutable, auditable. |
| `pipelineName` | Native supported pipeline | Initial value is `cinematic`. |
| `capabilityManifest` | Declared tools, resource classes and availability | Must be available before a run is admitted. |
| `licenseNoticeVersion` | Published AGPL/source-offer notice version | Required for public runtime admission. |

### `OpenMontageProductionRun`

| Field | Meaning | Rules |
| --- | --- | --- |
| `id` | Durable production identity | Globally unique. |
| `userId` | Owning user | Required for all reads and mutations. |
| `dramaEditTaskId` | Owning editing task | One active production per task and versioned retry history. |
| `runtimeVersionId` | Fixed runtime source version | Required and immutable after creation. |
| `workspaceKey` | Opaque scoped native workspace location | Never expose filesystem paths to users. |
| `pipelineName` | Native pipeline | Initially `cinematic`; immutable after creation. |
| `status` | Projection of native checkpoint state | `queued`, `in_progress`, `awaiting_approval`, `blocked`, `completed`, `cancelled`. |
| `currentStage` | Native cinematic stage | Must be declared by the fixed manifest. |
| `inputSnapshot` | Immutable source/brief/budget/capability references | Contains resource IDs only, never credentials or raw local paths. |
| `lastCheckpointAt` | Latest accepted checkpoint timestamp | Monotonic per production. |
| `deliveryResourceId` | Imported final video resource | Required only for `completed`. |
| `renderReportResourceId` | Imported native render report resource | Required only for `completed`. |

### `OpenMontageStageApproval`

One append-only decision for a production stage and artifact revision. It records the owning user,
stage, decision (`approved`, `revision_requested`, `rejected`, `cancelled`), optional note, and
timestamp. A required stage cannot advance without the corresponding approved decision.

### Relationships and Invariants

- A `DramaEditTask` may have many historical production runs but at most one non-terminal run.
- A production run belongs to exactly one user, one drama-editing task, and one runtime version.
- Native OpenMontage checkpoint and artifact files are authoritative; BeefTV projections are
  cacheable views that must be refreshed from the scoped workspace.
- `completed` requires both a final video resource and a render-report resource. Any missing
  native workspace, checkpoint, output, or ownership relationship projects as `blocked`.

## Amendment: Eligible Episode Asset

`EpisodeFolderCandidate.videos` is a transient owner-scoped source selection list. An eligible video is owned by
the current user, assigned to the folder, non-archived, and does not contain `data.source: "OpenMontage"` in its
persisted asset payload. `videoCount` is the length of this same set. OpenMontage delivery assets remain ordinary
library assets and retain their folder relationship; they are simply not selectable as future episode input.

## Amendment: Task Card Presentation

`DramaEditTask.openMontageStage` remains native operational metadata and is not changed or persisted differently.
The task-card projection derives an ephemeral creator-facing label from it; no progress value is part of the
task-card presentation contract.

## Amendment: Compact Task Status Presentation

`DramaEditTask.productionProgress` remains the authoritative numeric value for an in-progress task. A task card presents it only within the right-side processing status label; stage-description text is not part of the card projection.

## Amendment: Creation Defaults and Episode Label

`CreateDramaEditTask.spoilerPolicy` remains required by the existing creation contract and is fixed to
`{ policy: "no-major-spoilers", restrictedItems: [] }` by the creation form. Episode asset identity and selection
remain unchanged; `assetTitle` has a compact visual projection with full text available through its hover label.
