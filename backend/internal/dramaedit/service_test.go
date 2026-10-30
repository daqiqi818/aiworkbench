package dramaedit

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"infinite-canvas/backend/internal/editing"
	"infinite-canvas/backend/internal/model"
	"infinite-canvas/backend/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.IDSequence{}, &model.Asset{}, &model.AssetFolder{}, &model.AssetVersion{}, &model.AssetRepresentation{}, &model.Resource{}, &model.Task{}, &model.DramaEditTask{}, &model.DramaEditDecision{}); err != nil {
		t.Fatal(err)
	}
	return New(repository.New(db)), db
}

func seedRenderableVideo(t *testing.T, db *gorm.DB, assetID string, durationMs int64) {
	t.Helper()
	now := time.Now().UTC()
	versionID, resourceID := assetID+"-v1", assetID+"-resource"
	if err := db.Model(&model.Asset{}).Where("id = ?", assetID).Update("primary_version_id", versionID).Error; err != nil {
		t.Fatal(err)
	}
	for _, record := range []any{
		&model.AssetVersion{ID: versionID, AssetID: assetID, Version: 1, Status: model.AssetVersionStatusConfirmed, CreatedAt: now, UpdatedAt: now},
		&model.Resource{ID: resourceID, UserID: "owner", Kind: "video", Status: model.ResourceStatusReady, MimeType: "video/mp4", DurationMs: durationMs, CreatedAt: now, UpdatedAt: now},
		&model.AssetRepresentation{ID: assetID + "-repr", TaskID: assetID + "-task", AssetVersionID: versionID, ResourceID: resourceID, MediaType: "video", Role: "output", CreatedAt: now},
	} {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func newProductionTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	service, db := newTestService(t)
	service.production = ProductionStarterFunc(func(userID string, timeline editing.Project, options editing.Options, operationID string) (*model.Task, error) {
		work := &model.Task{ID: "render-" + operationID, UserID: userID, Type: model.TaskTypeTimelineRender, Status: model.TaskStatusQueued, Stage: "等待队列调度", Progress: 5, InputJSON: "{}", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := db.Create(work).Error; err != nil {
			return nil, err
		}
		return work, nil
	})
	return service, db
}

func seedEpisodeFolder(t *testing.T, db *gorm.DB, userID, folderID, name string, position int, videoIDs ...string) {
	t.Helper()
	now := time.Now().UTC()
	if err := db.Create(&model.AssetFolder{ID: folderID, UserID: userID, Name: name, NameKey: name, Position: position, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if len(videoIDs) == 0 {
		videoIDs = []string{"asset-" + folderID}
	}
	for _, videoID := range videoIDs {
		if err := db.Create(&model.Asset{ID: videoID, UserID: userID, FolderID: folderID, Kind: "video", Category: model.AssetCategoryOther, Status: model.AssetVersionStatusConfirmed, Title: name + " " + videoID, PayloadJSON: "{}", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func validRequest() CreateRequest {
	return CreateRequest{DramaTitle: "信号之外", DeliveryPreferences: DeliveryPreferences{AspectRatio: "9:16", MinDurationSeconds: 15, MaxDurationSeconds: 60, Language: "zh", Platform: "douyin"}, SpoilerPolicy: SpoilerPolicy{Policy: "no-major-spoilers"}, EpisodeAssetIDs: []string{"asset-folder-1", "asset-folder-2"}}
}

func TestCreateUsesSelectedEpisodeVideos(t *testing.T) {
	service, db := newTestService(t)
	seedEpisodeFolder(t, db, "owner", "folder-1", "第 1 集", 0)
	seedEpisodeFolder(t, db, "owner", "folder-2", "第 2 集", 1)
	task, err := service.Create("owner", validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != string(model.DramaEditStatusAwaitingApproval) || len(task.Episodes) != 2 || task.Episodes[0].FolderName != "第 1 集" || task.Episodes[0].AssetID != "asset-folder-1" {
		t.Fatalf("task = %#v", task)
	}
	if _, err := service.Get("other", task.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign get err = %v", err)
	}
}

func TestCreateUsesMultipleVideosFromOneAssetFolder(t *testing.T) {
	service, db := newTestService(t)
	seedEpisodeFolder(t, db, "owner", "67906322", "67906322", 0, "video-1", "video-2", "video-3")
	request := validRequest()
	request.EpisodeAssetIDs = []string{"video-1", "video-2", "video-3"}
	task, err := service.Create("owner", request)
	if err != nil {
		t.Fatal(err)
	}
	if len(task.Episodes) != 3 || task.Episodes[0].FolderID != "67906322" || task.Episodes[2].AssetID != "video-3" {
		t.Fatalf("task episodes = %#v", task.Episodes)
	}
}

func TestCreateRejectsInvalidDeliveryAndEpisodeFolders(t *testing.T) {
	service, db := newTestService(t)
	seedEpisodeFolder(t, db, "owner", "folder-1", "第 1 集", 0)
	seedEpisodeFolder(t, db, "owner", "folder-2", "第 2 集", 1)
	for name, mutate := range map[string]func(*CreateRequest){
		"one video":        func(request *CreateRequest) { request.EpisodeAssetIDs = request.EpisodeAssetIDs[:1] },
		"duplicate video":  func(request *CreateRequest) { request.EpisodeAssetIDs[1] = "asset-folder-1" },
		"foreign video":    func(request *CreateRequest) { request.EpisodeAssetIDs[1] = "foreign" },
		"too short":        func(request *CreateRequest) { request.DeliveryPreferences.MinDurationSeconds = 14 },
		"inverted range":   func(request *CreateRequest) { request.DeliveryPreferences.MinDurationSeconds = 61 },
		"invalid language": func(request *CreateRequest) { request.DeliveryPreferences.Language = "ja" },
		"invalid platform": func(request *CreateRequest) { request.DeliveryPreferences.Platform = "youtube" },
	} {
		t.Run(name, func(t *testing.T) {
			request := validRequest()
			mutate(&request)
			if _, err := service.Create("owner", request); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestEpisodeFoldersExcludeArchivedAndForeignAssets(t *testing.T) {
	service, db := newTestService(t)
	seedEpisodeFolder(t, db, "owner", "folder-1", "第 1 集", 0)
	seedEpisodeFolder(t, db, "other", "foreign", "外部剧集", 0)
	now := time.Now().UTC()
	if err := db.Create(&model.AssetFolder{ID: "archived", UserID: "owner", Name: "已归档", NameKey: "已归档", Position: 2, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Asset{ID: "archived-video", UserID: "owner", FolderID: "archived", Kind: "video", Category: model.AssetCategoryOther, Status: model.AssetVersionStatusArchived, Title: "归档视频", PayloadJSON: "{}", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	folders, err := service.EpisodeFolders("owner")
	if err != nil || len(folders) != 1 || folders[0].FolderID != "folder-1" || folders[0].VideoCount != 1 || len(folders[0].Videos) != 1 || folders[0].Videos[0].AssetID != "asset-folder-1" {
		t.Fatalf("folders=%#v err=%v", folders, err)
	}
}

func TestDecisionsFollowLifecycle(t *testing.T) {
	service, db := newTestService(t)
	seedEpisodeFolder(t, db, "owner", "folder-1", "第 1 集", 0)
	seedEpisodeFolder(t, db, "owner", "folder-2", "第 2 集", 1)
	task, err := service.Create("owner", validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.RequestRevision("owner", task.ID, "减少第二集剧透"); err != nil {
		t.Fatal(err)
	}
	approved, err := service.Approve("owner", task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if approved.Status != string(model.DramaEditStatusApproved) || len(approved.Decisions) != 3 {
		t.Fatalf("approved = %#v", approved)
	}
	if _, err := service.RequestRevision("owner", task.ID, "再次修改"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("revision after approve err = %v", err)
	}
}

func TestApprovalStartsProductionAndProjectsCompletion(t *testing.T) {
	service, db := newProductionTestService(t)
	seedEpisodeFolder(t, db, "owner", "folder-1", "第 1 集", 0)
	seedEpisodeFolder(t, db, "owner", "folder-2", "第 2 集", 1)
	seedRenderableVideo(t, db, "asset-folder-1", 9000)
	seedRenderableVideo(t, db, "asset-folder-2", 9000)
	created, err := service.Create("owner", validRequest())
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.Approve("owner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != string(model.DramaEditStatusProcessing) || started.ProductionTaskID == "" || started.ProductionProgress != 5 {
		t.Fatalf("started = %#v", started)
	}
	if err := db.Model(&model.Task{}).Where("id = ?", started.ProductionTaskID).Updates(map[string]any{"status": model.TaskStatusSucceeded, "stage": "渲染完成", "progress": 100, "result_json": `{"resourceId":"rendered-output"}`}).Error; err != nil {
		t.Fatal(err)
	}
	completed, err := service.Get("owner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != string(model.DramaEditStatusCompleted) || completed.DeliveryResourceID != "rendered-output" || completed.DeliveryURL != "/api/resources/rendered-output/file" {
		t.Fatalf("completed = %#v", completed)
	}
}

func TestApprovalBlocksWhenProductionSourcesAreUnavailable(t *testing.T) {
	service, db := newProductionTestService(t)
	seedEpisodeFolder(t, db, "owner", "folder-1", "第 1 集", 0)
	seedEpisodeFolder(t, db, "owner", "folder-2", "第 2 集", 1)
	created, err := service.Create("owner", validRequest())
	if err != nil {
		t.Fatal(err)
	}
	blocked, err := service.Approve("owner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if blocked.Status != string(model.DramaEditStatusBlocked) || blocked.ProductionError == "" {
		t.Fatalf("blocked = %#v", blocked)
	}
	if retried, err := service.StartProduction("owner", created.ID); err != nil || retried.Status != string(model.DramaEditStatusBlocked) {
		t.Fatalf("retry = %#v err=%v", retried, err)
	}
}

func TestRetryCreatesNewWorkAfterTerminalRenderFailure(t *testing.T) {
	service, db := newProductionTestService(t)
	seedEpisodeFolder(t, db, "owner", "folder-1", "第 1 集", 0)
	seedEpisodeFolder(t, db, "owner", "folder-2", "第 2 集", 1)
	seedRenderableVideo(t, db, "asset-folder-1", 9000)
	seedRenderableVideo(t, db, "asset-folder-2", 9000)
	created, err := service.Create("owner", validRequest())
	if err != nil {
		t.Fatal(err)
	}
	first, err := service.Approve("owner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.Task{}).Where("id = ?", first.ProductionTaskID).Updates(map[string]any{"status": model.TaskStatusFailed, "error": "ffmpeg unavailable"}).Error; err != nil {
		t.Fatal(err)
	}
	retried, err := service.StartProduction("owner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if retried.Status != string(model.DramaEditStatusProcessing) || retried.ProductionTaskID == "" || retried.ProductionTaskID == first.ProductionTaskID {
		t.Fatalf("retry = %#v; first = %#v", retried, first)
	}
}

func TestApprovalUsesReadyResourceStorageKeyWithoutRepresentation(t *testing.T) {
	service, db := newProductionTestService(t)
	seedEpisodeFolder(t, db, "owner", "folder-1", "第 1 集", 0)
	seedEpisodeFolder(t, db, "owner", "folder-2", "第 2 集", 1)
	now := time.Now().UTC()
	for _, assetID := range []string{"asset-folder-1", "asset-folder-2"} {
		resourceID := assetID + "-handoff"
		if err := db.Create(&model.Resource{ID: resourceID, UserID: "owner", Kind: "video", Status: model.ResourceStatusReady, MimeType: "video/mp4", DurationMs: 9000, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Model(&model.Asset{}).Where("id = ?", assetID).Update("payload_json", fmt.Sprintf(`{"data":{"storageKey":"resource:%s"}}`, resourceID)).Error; err != nil {
			t.Fatal(err)
		}
	}
	created, err := service.Create("owner", validRequest())
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.Approve("owner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != string(model.DramaEditStatusProcessing) || started.ProductionTaskID == "" {
		t.Fatalf("started = %#v", started)
	}
}
