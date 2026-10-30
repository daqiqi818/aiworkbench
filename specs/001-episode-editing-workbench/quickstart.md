# Quickstart: Multi-Episode Editing Workbench

## Prerequisites

Use an isolated local backend data directory. Do not run against a production workspace. The
feature requires the backend to migrate its local schema and the web dependencies already present.

## Validation flow

1. Run the focused Go domain and handler tests listed in the implementation tasks.
2. Run `cd backend && go test ./...` after the feature tests pass.
3. Run `cd web && bun run build` to type-check and bundle the lazy route.
4. Start the local backend only when browser validation is explicitly required, using
   `CANVAS_BACKEND_DATA_DIR=../.local/project-workbench-debug go run ./cmd/server` from `backend/`.
5. In Assets, create an asset folder and place at least two episode videos in it. Confirm the
   newly created folder remains in the sidebar after refresh.
6. Open `/drama-editing`, select the asset folder, confirm its videos are selected by default, submit a task with a
   15-300 second valid duration range and supported language/platform, then verify its proposal
   and `awaiting_approval` status.
7. Verify that a range below 15 seconds, above 300 seconds, inverted range, unsupported delivery
   value, one selected video, or an emptied folder is rejected with a clear message.
8. Request a revision, approve the updated proposal, and cancel a separate task. Verify approval
   creates one local production work, then verify its queued/running/completed or blocked state is
   reflected in the editing task without manual status editing.

See [data-model.md](./data-model.md) for validation rules and
[contracts/drama-editing-api.md](./contracts/drama-editing-api.md) for the request surface.

## Executed Validation

- Passed: `GOCACHE=/tmp/aiworkbench-go-cache go test ./internal/dramaedit ./internal/repository ./internal/database ./internal/localapp`.
- Passed: `GOCACHE=/tmp/aiworkbench-go-cache go test ./internal/handler -run 'Test(FailDramaEditingProjectsPublicErrors|RegisterDramaEditingRoutesIncludesEpisodeFolders)$'`.
- Passed: `cd web && bun run build`.
- `cd backend && GOCACHE=/tmp/aiworkbench-go-cache go test ./...` still reports existing environment and dependency failures: loopback listener permissions, missing packaged plugins, and an FFmpeg installation without the `subtitles` filter. The run no longer reports the drama-editing port compile failure found during this change.
- Passed: `cd web && bun run build` after the browser-local folder metadata bridge.
- Passed: `cd web && bun run build` after task-submission feedback handling.
- Passed: `GOCACHE=/private/tmp/aiworkbench-gocache go test ./internal/dramaedit ./internal/database -run 'Test(Approval|Decisions|Create|Episode|AssistantTurnsMigrationFromPreviewV9)'`.
- Passed: `GOCACHE=/private/tmp/aiworkbench-gocache go test ./internal/handler -run 'Test(FailDramaEditingProjectsPublicErrors|RegisterDramaEditingRoutesIncludesEpisodeFolders)$'` and `go test ./internal/localapp`.
- Passed: `cd web && bun run build` after adding production status, output, and recovery controls.
- Passed: `GOCACHE=/private/tmp/aiworkbench-gocache go test ./internal/dramaedit -run 'Test(Approval|Create|Episode)'`, including resource-backed browser-local handoff admission.
- Passed: `cd web && bun run build` after adding browser-local production source handoff.
- Passed: deployed `http://127.0.0.1:3000/drama-editing` Playwright inspection of task `1222`; its 7 source episodes are present and an isolated browser with no IndexedDB source Blob receives the in-drawer `素材交接失败` error without creating another production work.
- Passed: rebuilt `docker-compose.local.yml`; the final backend container reports `/usr/bin/ffmpeg`, `/usr/bin/ffprobe`, and FFmpeg `6.1.2`.
- Passed: source-ready task `2333` (`DE_000002`) completed after retry. The linked production work `47de257606182ec08bc30e56ad350c24` reached `succeeded` at 100% and produced video resource `ab6a13f9ccce29ad419fa38c7490079f`.
- Passed: downloaded the resulting delivery resource through `/api/resources/ab6a13f9ccce29ad419fa38c7490079f/file`: HTTP 200, `video/mp4`, 24,561,842 bytes. `ffprobe` confirmed a 60-second 1080x1920 H.264 video stream with AAC audio.

## Production Execution Verification

1. Create a task from at least two source videos with ready local resource representations whose
   total duration is at least the requested minimum.
2. Approve the proposal. Confirm one `timeline_render` work is linked and the task enters
   `processing`; refreshing must retain the work ID and progress.
3. On successful rendering, reload the editing task and confirm it is `completed` with an output
   resource URL. On missing resources, insufficient duration, or FFmpeg failure, confirm it is
   `blocked` with the actual reason and a start/retry action.
4. For an older approved task without a work ID, use **开始制作** once. Repeating it while the work
   is active must return the same linked work rather than create a duplicate.

## Browser-Local Production Source Handoff Verification

1. In browser-local mode, use a task whose selected video assets still have local (non-`resource:`)
   storage keys and whose IndexedDB Blobs are available.
2. Approve the task or choose **开始制作**. Confirm each selected source is uploaded once through
   the existing local resource upload path and its asset record changes to a `resource:` storage
   key before the drama-editing start request is issued.
3. Confirm the resulting production task enters `processing` rather than reporting that the first
   selected episode has no readable video resource. Retry before completion and verify no second
   source resource is created.
4. Delete or make one local Blob unreadable and confirm no production request is created; the
   drawer must name the affected episode and direct the creator to re-import it.

## Browser-Local Folder Availability Verification

1. In browser-local mode, create an asset folder named `67906322` and upload a video into it.
2. Reload the assets page. The folder remains listed; no second media upload is requested.
3. Open **多剧集剪辑** and select **新建剪辑任务**. The folder appears with a positive video
   count and is included in the default selection.
4. Confirm a metadata synchronization failure is shown as an actionable error rather than the
   no-folder empty state.

## Task Submission Feedback Verification

1. Open the creation dialog with the short-drama name empty and confirm. The dialog stays open,
   the field is marked, and an error explains that required settings need completion.
2. Select only one episode video and confirm. The dialog reports the current selected count and the
   two-episode minimum without issuing a create request.
3. With two candidate videos from one selected folder, a short-drama name, and valid delivery settings, confirm once.
   The dialog closes, the new task is selected, and the list refreshes.
