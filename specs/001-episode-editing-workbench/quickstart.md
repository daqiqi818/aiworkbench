# Quickstart: Multi-Episode Editing Workbench

## Prerequisites

Use an isolated local backend data directory. Do not run against a production workspace. The
feature requires the backend to migrate its local schema and the web dependencies already present.

## Validation flow

### Verified Task Deletion

1. In the rebuilt local browser workbench, create two disposable completed tasks with distinct
   source videos and record their task IDs and source asset IDs.
2. Select both tasks, open the delete confirmation, and confirm that **同时删除剧集文件** is visible
   and unchecked. Confirm deletion; verify both task cards disappear while source videos, delivery
   videos and reports remain in the asset library.
3. Repeat with disposable completed tasks and select **同时删除剧集文件**. Verify only eligible
   source episode assets are removed; delivery/report assets remain. Keep or create a referenced
   source asset and verify it is retained with an actionable reason.
4. Record the actual browser URL, selected task count, request result, post-delete task list and
   source-asset outcome. A frontend build or direct endpoint result alone does not complete this
   validation.

### Active OpenMontage Task Deletion

1. Create an isolated two-source task and approve it until it is `processing` or waiting at an
   OpenMontage checkpoint; record its BeefTV task ID and native production ID.
2. Select that task in the list. The confirmation must say that active OpenMontage production will
   be stopped before deletion. Confirm without selecting source-file deletion.
3. Verify the request returns a deleted task result, the task disappears after the list refresh,
   and the Runner status for the recorded production is `cancelled`.
4. Repeat against a controlled unavailable Runner or stop failure. Verify the task remains visible
   with the returned stop reason while another independently deletable task in the same request is
   still removed.

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

- Passed: rebuilt and restarted the modified `web` service. In a real browser, opened completed
  task `3223` from the short-drama task list. The in-page portrait player rendered in the compact
  560px modal with its height constrained to the viewport. Closing it removed the dialog and
  video element from the accessibility tree; its unified close handler pauses, removes `src`, and
  reloads the media element before unmounting it.
- Passed: repeated the same real-browser sequence for task `3223`: open and play, close, then
  reopen the same delivery. The second open loaded and displayed a fresh video frame, proving the
  player source is reattached after cleanup rather than reusing the emptied media element.

- Passed: fixed and rebuilt the native compose path after `DE_000015` was cancelled. The Runner
  now writes the OpenMontage-schema-valid `render_report.version: "1.0"` and stores editorial
  details at `metadata.editorial_report`; a compose failure is durably returned as `blocked` with
  its error text rather than being overwritten as `processing` by BeefTV polling.
- Passed: `GOCACHE=/private/tmp/aiworkbench-gocache go test ./internal/dramaedit ./internal/localapp`,
  `PYTHONPATH=. /Users/dajiang/Documents/projects/OpenMontage/.venv/bin/python -m pytest tests -q`
  (`4 passed`), and `cd web && bun run build`.
- Passed: rebuilt and force-recreated `backend`, `web`, and `openmontage-runner` with
  `docker-compose.local.yml`; all three containers subsequently reported healthy.
- Passed: Playwright browser acceptance created `DE_000016` (`修复验收 20261030`) from real
  `67906322_1` and `67906322_2` BeefTV assets and approved
  `proposal -> script -> scene_plan -> assets -> publish`. The formerly failing assets approval
  took 15.52 seconds and correctly advanced to `publish` at 100%, rather than staying at 71%.
  Publish completed with BeefTV delivery resource `4aa59cb01b6ee0118b4fbe4a71f59799` and report
  resource `96c5c2229839eeb7c9c9005a5a13f5d8`; no `timeline_render` task was used. `ffprobe`
  confirmed H.264/AAC, 1080x1920, 15.000000 seconds, 5,664,774 bytes. The browser report modal
  displayed episode navigation, hook, story progression, suspense, and process sections from
  `metadata.editorial_report`.

- Passed: rebuilt `backend` and `web` in `docker-compose.local.yml`; backend container
  `sha256:c12202001e1c18aff16a4e8e8b560013a55e1bccacec40f9081005e8133620a5`
  reported healthy, and the browser workbench loaded at `http://localhost:3000/drama-editing`.
- Passed: real Playwright browser deletion with source deletion unchecked selected completed
  `DE_000010` (`Final Strict E2E 20261030`) and `DE_000008` (`UI E2E 20261030`). The confirmation
  showed **同时删除剧集文件** unchecked by default; `DELETE /api/drama-editing/tasks` returned 200,
  both task cards disappeared, and the shared `67906322` source assets remained candidates.
- Passed: real Playwright browser opted-in deletion used isolated folder
  `DELETE-E2E-20261030` with disposable assets `06b5cb650c584511831ee07a6b79d3fd` and
  `bbb1c31a1068424bbd0f7bf4f9328f6a`. A UI-created/cancelled task `DE_000013` was deleted after
  checking **同时删除剧集文件**; the browser request returned 200, the task disappeared, and both
  asset reads plus resources `168beb750402d28b7edcb0bf12cf7274` and
  `4d6f77908ac670692a84856b799beecb` returned HTTP 404.
- Passed: rebuilt and restarted the affected `backend` and `web` services. In a real Playwright
  browser at `http://127.0.0.1:3000/drama-editing`, created and approved isolated task
  `DE_000014` (`Active Delete E2E 20261030`) until it reached the native `proposal` stage with
  `processing` status. Its delete dialog stated that one active task would stop OpenMontage before
  deletion. Confirming the default task-only deletion removed the card from the refreshed list.
  The persistent Runner record for `om-DE_000014` then reported `status: cancelled` with
  `cancelledAt: 2026-10-30T14:37:00.304033+00:00`.
- Passed: `cd web && bun run build` after adding short-drama terminology, folder-first selection,
  preference restoration, task actions, in-page delivery playback, report modal and bulk deletion.
- Passed: `cd backend && GOCACHE=/private/tmp/aiworkbench-gocache go test ./internal/dramaedit`.
- The combined handler package test is blocked in this environment by existing tests that attempt
  to bind `127.0.0.1:0` / `::1:0`, which sandbox policy rejects; it is not a short-drama compile
  or route failure.
- Runner validation passed three non-network tests. Its real-video test is blocked at creation of
  a local source broker on `127.0.0.1:0` by the same sandbox listener policy; syntax compilation
  passed and this environment cannot provide the required loopback test server.

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
- Passed: vendored clean OpenMontage source snapshot at `9590f5639a0fe8b32d54dd74e196b8ea02191453`; exclusion checks confirmed no `.env`, `.git`, virtual environment, Node dependencies, or project workspace entered `third_party/openmontage`.
- Passed: `PYTHONPATH=openmontage-runner /Users/dajiang/Documents/projects/OpenMontage/.venv/bin/python -m pytest openmontage-runner/tests -q` (3 passed) and Python compile validation. The tests confirm the native `cinematic` stage order, authenticated internal-only access, owner-scoped workspace/checkpoint state, and scoped cancellation.
- Passed: `POSTGRES_PASSWORD=validation-only DATABASE_URL=postgres://validation:validation@postgres:5432/validation OPENMONTAGE_RUNNER_TOKEN=validation-only docker compose -f docker-compose.deploy.yml config` and `docker compose -f docker-compose.local.yml config`.
- Passed: `GOCACHE=/private/tmp/aiworkbench-gocache go test ./internal/dramaedit -run 'Test(Approval|Create|Episode|Retry)'` and `cd web && bun run build` after adding the OpenMontage runtime foundation.
- Not completed in this environment: Docker image build for `openmontage-runner`. Docker began installing the Runner's Debian media dependencies but this execution environment terminated the build before it could leave an image. The Dockerfile and Compose syntax were validated; rebuild the runner image in a normal Docker build window before publishing.
- Passed: rebuilt and started the local `openmontage-runner` image. `PYTHONPATH=openmontage-runner /Users/dajiang/Documents/projects/OpenMontage/.venv/bin/python -m pytest openmontage-runner/tests -q` (4 passed) covers owner-scoped BeefTV source references, production-token enforcement, native approval gates, FFmpeg composition, and cleanup of temporary source files.
- Passed: real BeefTV task `DE_000005` (`67906322 E2E`) selected existing `67906322_1` and `67906322_2` resources, then completed `proposal -> script -> scene_plan -> assets -> publish` approvals through the workbench endpoint. It reached `completed` without creating a `timeline_render` task. BeefTV imported final resource `1c0711d8e49572f5b30f3917f0f8d7c3` and render-report resource `c14c8947a39dc3f9c499ef963dbd2a15`; the final video was registered as asset `bf3e03cd13f05cff29b656ada93258af` in folder `67906322`. FFprobe confirmed a 15.021333-second 1080x1920 H.264/AAC MP4 (5,751,656 bytes). The Runner workspace retains checkpoint/report/output files only; source-media copies were absent after rendering.
- Passed: rebuilt the Runner and Web after removing browser-side source handoff and exempting native production controls from the four-second JSON timeout. A Playwright-created task `DE_000008` advanced through every native gate and completed with a final video and report returned to BeefTV.
- Passed: final real-source acceptance task `DE_000010` (`Final Strict E2E 20261030`) used `67906322_1` and `67906322_2`, completed `proposal -> script -> scene_plan -> assets -> edit -> compose -> publish`, and returned final resource `382cd0dc70fdaa00bc7419cec6cf7c7d` plus report `2c112b91c785bdbdf2bcad5321317aaa`. FFprobe confirmed H.264/AAC, 1080x1920, 5,664,774 bytes, and exactly `15.000000` seconds. Asset `9403a35b554cfda797f0071f4a9a49e4` was confirmed in the original `67906322` folder. The rebuilt Runner test suite passed `4 passed`, including the maximum-duration assertion.

## Production Execution Verification

1. Create a task from at least two source videos with ready local resource representations whose
   total duration is at least the requested minimum.
2. Approve the proposal. Confirm one OpenMontage production is linked, the task enters
   `processing`, and the first native approval gate is visible; no `timeline_render` task is created.
3. Approve each required native stage. On successful rendering, reload the editing task and confirm
   it is `completed` with final-video and render-report resource URLs. On source, Runner, import,
   or FFmpeg failure, confirm it is `blocked` with the actual reason.
4. Inspect the Runner workspace and confirm no selected source video persists there; only native
   checkpoints, reports, derived working files, and final output may remain.

## OpenMontage Published Runtime Verification

1. Verify the vendored source manifest identifies the selected upstream commit, AGPLv3 license,
   exclusions, and supported native `cinematic` pipeline.
2. Build the internal runner image. Confirm it reports the native cinematic stage order and has no
   provider credentials in its capability response or logs.
3. Create two runner productions with different user and production identities. Confirm their
   workspaces, checkpoint results, cancellation requests and artifact listings are isolated.
4. Use fixture-only source resources to create a production. Confirm the canonical project marker
   and `research` checkpoint are persisted on durable storage, then restart the runner and reload
   the same projected state.
5. Attempt to advance a required-approval stage without approval. Confirm the runner rejects it;
   record approval, then verify the next manifest-valid action is exposed.
6. Before enabling paid capabilities, run the cinematic fixture through every stage with a
   non-billed tool profile. Verify the final video and `render_report` are both imported before
   the workbench shows `completed`.

## BeefTV-Owned Source Verification

1. Create a task only from episode candidates returned by BeefTV's owner-scoped asset endpoint.
2. Approve the task. Confirm the browser makes no source upload or local Blob read; the backend
   resolves the task's owned `resource:` references and creates the OpenMontage production.
3. Confirm a missing or unreadable source is reported by the backend as the task's actual blocker.
4. Confirm Runner workspaces retain no selected source-media copies after a completed render.

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

## Source Candidate and Asset Workspace Verification

1. Place seven source videos and one or more completed OpenMontage delivery videos in the same `67906322` folder.
   Open **短剧剪辑** and confirm the creation action reads **新建任务**, the folder label reads `67906322（7集）`,
   and exactly seven source-video checkboxes are shown.
2. Open **资产** directly. Confirm the classification sidebar is visible, includes **我的短剧**, and has no
   **快捷筛选** section or right-side **生成历史 / 个人资产库** rail.
3. Confirm the header offers distinct **新建文件夹** and **上传资产** controls. Verify the first opens the existing
   folder editor and the second opens the native file chooser without creating an asset before a file is chosen.

### 2026-10-31 Validation Evidence

- `GOCACHE=/private/tmp/aiworkbench-gocache go test ./internal/repository ./internal/dramaedit` passed.
- `cd web && bun run build` passed after the asset-workspace update.
- Rebuilt and restarted `backend` and `web` using `docker-compose.local.yml`; backend reported healthy and web
  was available at `http://127.0.0.1:3000`. The unchanged OpenMontage Runner remained running.
- Browser validation against the rebuilt stack confirmed the assets sidebar is initially expanded, uses **我的短剧**,
  has no quick-filter section or source rail, and presents separate **新建文件夹** and **上传资产** buttons. The folder
  button opens only the inline editor after removal of the duplicate modal.
- The live owner-scoped episode endpoint returned `67906322` with `videoCount: 7` and source files `_1` through `_7`.
  In the browser, the selector displayed `67906322（7集）`; selecting it displayed `7/7 集已选` and exactly seven checked
  source-video boxes. OpenMontage delivery artifacts were absent from the candidate list.

### Backend-Readable Source Validation

- Passed: Browser-uploaded `qa-grok-imagine-720square.mp4` into `资源验证剧` created asset
  `urtulegGYBwaugaz_MqMR` with `storageKey: resource:7f09d7675a06c95a3a66978df42d068c`. Its owner-scoped
  resource record is `ready`, `video/mp4`, and has the expected dimensions, duration and byte count.
- Passed: the episode candidate endpoint reported only `67906322（7集）` plus the one newly uploaded source;
  `第二剧`'s ten historical `blob:` / `video:` assets were excluded rather than presented as renderable sources.
- Passed: browser-created task `DE_000019` (`资源链路验收`) selected two existing BeefTV resource-backed sources,
  progressed through `proposal -> script -> scene_plan -> assets -> edit -> compose -> publish`, and completed.
  BeefTV imported final video resource `54d96423c3e464103d32d4efca3ae058` and render report
  `965aab78acd4494a3ad506fe8e99587c`; the workbench showed the completed task's in-page playback, download and
  report actions. A temporary download followed by FFprobe confirmed a 60.000-second 1080x1920 H.264/AAC MP4
  (22,711,116 bytes). The task did not encounter the prior unreadable-video blocker.
- Passed: `GOCACHE=/private/tmp/aiworkbench-gocache go test ./internal/repository ./internal/dramaedit` and
  `cd web && bun run build`. Rebuilt and restarted `backend` and `web`; all Compose services are healthy.

## Creator-Facing Task Label Verification

1. Open `/drama-editing` and confirm the workspace sidebar label is **短剧剪辑**.
2. Inspect a task whose native stage is `publish` (or another active native stage). Confirm its list card has no
   progress bar and no native code such as `publish`; its short description is Chinese and creator-facing.
3. Confirm the task detail can still use the existing native approval flow, demonstrating that presentation did not
   alter task metadata or production controls.

### 2026-10-31 Validation Evidence

- Passed: `cd web && bun run build` completed with TypeScript checking and production bundle generation.
- Passed: rebuilt and restarted the changed `web` service with `docker compose -f docker-compose.local.yml up -d --build web`; the service reported healthy at `http://127.0.0.1:3000`.
- Passed: real-browser validation at `/drama-editing` found the sidebar link **短剧剪辑**. The two rendered task cards had `0` `.ant-progress` elements and neither card text contained `publish`; completed tasks displayed **成片已回传** instead.

## Compact Task Status Verification

1. Open `/drama-editing` with a processing task returned by the task endpoint.
2. Confirm the task's left summary contains only its title and episode/duration summary, with no stage wording such as “剪辑制作中”.
3. Confirm the right status tag reads `制作中 N%`, where `N` is the task's reported current progress; confirm non-processing tasks retain their ordinary status labels.

## Compact Task Creation Verification

1. Open **新建任务** and confirm no剧透设置控件 is visible.
2. Select `67906322（7集）` and confirm the episode chooser's first row contains five selectable entries.
3. Verify a long episode name is ellipsized within its entry, then hover it and confirm the full name is available without changing the grid layout.
4. Submit a valid task and inspect its saved policy through the existing task detail or request: it must be **避免重大剧透**.
5. Confirm the short-drama name, aspect ratio, language, and platform appear together in the compact first row; confirm duration controls begin the next row without clipped text.

### 2026-10-31 Validation Evidence

- Passed: `cd web && bun run build` completed with TypeScript checking and production bundle generation.
- Passed: rebuilt and restarted only the changed `web` service using `docker compose -f docker-compose.local.yml up -d --build web`; it reported healthy at `http://127.0.0.1:3000`.
- Passed: real-browser verification selected `67906322（7集）`. The creation dialog contained no **剧透限制** control, selected all seven episodes by default, and rendered the first five episode options in one row.
- Passed: a browser-only long-title check produced `clientWidth: 42`, `scrollWidth: 423`, and `textOverflow: ellipsis`; the element retained the full title for hover. The compact first-row controls measured 248px for **短剧名称** and 120px each for **画幅**、**语言**、**投放平台**; the two duration controls began the next row. No task or asset was created or changed during this validation.

### 2026-10-31 Validation Evidence

- Passed: `cd web && bun run build` completed with TypeScript checking and production bundle generation.
- Passed: rebuilt and restarted only the changed `web` service using `docker compose -f docker-compose.local.yml up -d --build web`; it reported healthy at `http://127.0.0.1:3000`.
- Passed: the live task list showed a completed task whose left summary contains only title and episode/duration metadata. An isolated browser-only `processing` task response with `productionProgress: 50` rendered left text **制作状态验证 2 集 · 15-60 秒**, right tag **制作中 50%**, and `0` task-card progress controls. The isolated response did not write or alter any BeefTV task.
