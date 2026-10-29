package dramaedit

import (
	"errors"
	"testing"

	"infinite-canvas/backend/internal/model"
	"infinite-canvas/backend/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.IDSequence{}, &model.DramaEditTask{}, &model.DramaEditDecision{}); err != nil {
		t.Fatal(err)
	}
	return New(repository.New(db))
}

func validRequest() CreateRequest {
	return CreateRequest{DramaTitle: "信号之外", CampaignObjective: "突出主角的谜团", TargetAudience: "悬疑短剧观众", DeliveryPreferences: DeliveryPreferences{AspectRatio: "9:16", TargetDurationSecond: 60, Language: "中文", DestinationLabel: "短视频平台"}, SpoilerPolicy: SpoilerPolicy{Policy: "no-major-spoilers"}, Episodes: []Episode{{EpisodeID: "E01", SequenceNumber: 1, Summary: "收到神秘信件", Inclusion: "included", Sources: []Source{{Reference: "asset:E01", RightsConfirmed: true}}}, {EpisodeID: "E02", SequenceNumber: 2, Summary: "揭开线索", Inclusion: "included", Sources: []Source{{Reference: "asset:E02", RightsConfirmed: true}}}}}
}

func TestCreateValidatesOwnershipAndEpisodeRules(t *testing.T) {
	service := newTestService(t)
	task, err := service.Create("owner", validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != string(model.DramaEditStatusAwaitingApproval) || len(task.Proposal.EpisodeCoverage) != 2 {
		t.Fatalf("task = %#v", task)
	}
	if _, err := service.Get("other", task.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign get err = %v", err)
	}
	invalid := validRequest()
	invalid.Episodes = invalid.Episodes[:1]
	if _, err := service.Create("owner", invalid); err == nil {
		t.Fatal("expected episode validation error")
	}
	invalid = validRequest()
	invalid.Episodes[1].SequenceNumber = 1
	if _, err := service.Create("owner", invalid); err == nil {
		t.Fatal("expected duplicate position error")
	}
	invalid = validRequest()
	invalid.Episodes[0].Sources[0].RightsConfirmed = false
	if _, err := service.Create("owner", invalid); err == nil {
		t.Fatal("expected rights error")
	}
}

func TestDecisionsFollowLifecycle(t *testing.T) {
	service := newTestService(t)
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
	cancelled, err := service.Cancel("owner", task.ID, "暂停活动")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != string(model.DramaEditStatusCancelled) {
		t.Fatalf("cancelled = %#v", cancelled)
	}
}
