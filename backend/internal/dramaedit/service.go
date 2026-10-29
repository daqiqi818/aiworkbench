package dramaedit

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"infinite-canvas/backend/internal/model"
	"infinite-canvas/backend/internal/repository"
)

var (
	ErrNotFound          = errors.New("剪辑任务不存在")
	ErrInvalidTransition = errors.New("当前剪辑任务状态不支持此操作")
)

type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }

type Service struct {
	repo  *repository.Repository
	store *store
}

func New(repo *repository.Repository) *Service { return &Service{repo: repo, store: newStore(repo)} }

func (s *Service) Create(userID string, request CreateRequest) (Task, error) {
	if strings.TrimSpace(userID) == "" {
		return Task{}, &ValidationError{Message: "当前用户无效"}
	}
	if err := validateRequest(&request); err != nil {
		return Task{}, err
	}
	delivery, _ := json.Marshal(request.DeliveryPreferences)
	spoiler, _ := json.Marshal(request.SpoilerPolicy)
	episodes, _ := json.Marshal(request.Episodes)
	proposal := buildProposal(request)
	proposalJSON, _ := json.Marshal(proposal)
	id, err := nextID(s.repo, "DE")
	if err != nil {
		return Task{}, err
	}
	decisionID, err := nextID(s.repo, "DED")
	if err != nil {
		return Task{}, err
	}
	timestamp := now()
	stored := model.DramaEditTask{ID: id, UserID: userID, DramaTitle: request.DramaTitle, CampaignObjective: request.CampaignObjective, TargetAudience: request.TargetAudience, DeliveryJSON: string(delivery), SpoilerJSON: string(spoiler), EpisodesJSON: string(episodes), ProposalJSON: string(proposalJSON), Status: model.DramaEditStatusAwaitingApproval, Stage: "等待审核", NextAction: "审核剪辑方案后批准、请求修改或取消", CreatedAt: timestamp, UpdatedAt: timestamp}
	decision := model.DramaEditDecision{ID: decisionID, TaskID: id, UserID: userID, Action: "created", CreatedAt: timestamp}
	if err := s.store.create(&stored, &decision); err != nil {
		return Task{}, err
	}
	return decodeTask(stored, []model.DramaEditDecision{decision})
}

func (s *Service) List(userID, status string) ([]Task, error) {
	if status != "" && !validStatus(status) {
		return nil, &ValidationError{Message: "任务状态无效"}
	}
	items, err := s.store.list(userID, status)
	if err != nil {
		return nil, err
	}
	result := make([]Task, 0, len(items))
	for _, item := range items {
		task, decodeErr := decodeTask(item, nil)
		if decodeErr != nil {
			return nil, decodeErr
		}
		result = append(result, task)
	}
	return result, nil
}

func (s *Service) Get(userID, id string) (Task, error) {
	stored, err := s.store.get(userID, id)
	if err != nil {
		return Task{}, err
	}
	decisions, err := s.store.decisions(userID, id)
	if err != nil {
		return Task{}, err
	}
	return decodeTask(*stored, decisions)
}

func (s *Service) Approve(userID, id string) (Task, error) {
	return s.decide(userID, id, "approved", "", []model.DramaEditStatus{model.DramaEditStatusAwaitingApproval, model.DramaEditStatusRevisionRequested})
}
func (s *Service) Cancel(userID, id, note string) (Task, error) {
	return s.decide(userID, id, "cancelled", strings.TrimSpace(note), []model.DramaEditStatus{model.DramaEditStatusAwaitingApproval, model.DramaEditStatusRevisionRequested, model.DramaEditStatusApproved, model.DramaEditStatusBlocked})
}

func (s *Service) RequestRevision(userID, id, note string) (Task, error) {
	note = strings.TrimSpace(note)
	if note == "" || len([]rune(note)) > 1000 {
		return Task{}, &ValidationError{Message: "修改说明需为 1 到 1000 个字符"}
	}
	return s.decide(userID, id, "revision_requested", note, []model.DramaEditStatus{model.DramaEditStatusAwaitingApproval, model.DramaEditStatusRevisionRequested})
}

func (s *Service) decide(userID, id, action, note string, allowed []model.DramaEditStatus) (Task, error) {
	task, err := s.store.get(userID, id)
	if err != nil {
		return Task{}, err
	}
	if !containsStatus(allowed, task.Status) {
		return Task{}, ErrInvalidTransition
	}
	timestamp := now()
	switch action {
	case "approved":
		task.Status, task.Stage, task.NextAction = model.DramaEditStatusApproved, "已批准", "剪辑执行器尚未接入；当前方案已锁定，等待执行"
	case "revision_requested":
		task.Status, task.Stage, task.NextAction = model.DramaEditStatusRevisionRequested, "等待方案修改", "查看修改说明并重新审核方案"
	case "cancelled":
		task.Status, task.Stage, task.NextAction = model.DramaEditStatusCancelled, "已取消", "任务已停止，不会进入剪辑执行"
	default:
		return Task{}, fmt.Errorf("unknown drama edit action %q", action)
	}
	task.UpdatedAt = timestamp
	decisionID, err := nextID(s.repo, "DED")
	if err != nil {
		return Task{}, err
	}
	decision := model.DramaEditDecision{ID: decisionID, TaskID: id, UserID: userID, Action: action, Note: note, CreatedAt: timestamp}
	if err := s.store.transition(task, &decision, allowed); err != nil {
		return Task{}, err
	}
	return s.Get(userID, id)
}

func validateRequest(request *CreateRequest) error {
	request.DramaTitle, request.CampaignObjective, request.TargetAudience = strings.TrimSpace(request.DramaTitle), strings.TrimSpace(request.CampaignObjective), strings.TrimSpace(request.TargetAudience)
	if !between(request.DramaTitle, 1, 200) || !between(request.CampaignObjective, 1, 1000) || !between(request.TargetAudience, 1, 1000) {
		return &ValidationError{Message: "短剧名称、剪辑目标和目标受众不能为空且长度必须有效"}
	}
	if request.DeliveryPreferences.AspectRatio != "9:16" && request.DeliveryPreferences.AspectRatio != "16:9" && request.DeliveryPreferences.AspectRatio != "1:1" {
		return &ValidationError{Message: "画幅比例无效"}
	}
	if request.DeliveryPreferences.TargetDurationSecond < 10 || request.DeliveryPreferences.TargetDurationSecond > 180 || strings.TrimSpace(request.DeliveryPreferences.Language) == "" || strings.TrimSpace(request.DeliveryPreferences.DestinationLabel) == "" {
		return &ValidationError{Message: "交付偏好无效"}
	}
	if request.SpoilerPolicy.Policy != "no-major-spoilers" && request.SpoilerPolicy.Policy != "limited-spoilers" && request.SpoilerPolicy.Policy != "approved-spoilers" {
		return &ValidationError{Message: "剧透策略无效"}
	}
	if len(request.Episodes) < 2 || len(request.Episodes) > 12 {
		return &ValidationError{Message: "每项剪辑任务必须包含 2 到 12 集"}
	}
	ids, positions, included := map[string]struct{}{}, map[int]struct{}{}, 0
	for index := range request.Episodes {
		episode := &request.Episodes[index]
		episode.EpisodeID, episode.Summary = strings.TrimSpace(episode.EpisodeID), strings.TrimSpace(episode.Summary)
		if episode.EpisodeID == "" || episode.Summary == "" || episode.SequenceNumber < 1 {
			return &ValidationError{Message: "剧集编号、顺序和摘要不能为空"}
		}
		if _, exists := ids[episode.EpisodeID]; exists {
			return &ValidationError{Message: "剧集编号不能重复"}
		}
		ids[episode.EpisodeID] = struct{}{}
		if _, exists := positions[episode.SequenceNumber]; exists {
			return &ValidationError{Message: "剧集顺序不能重复"}
		}
		positions[episode.SequenceNumber] = struct{}{}
		if episode.Inclusion != "included" && episode.Inclusion != "excluded" {
			return &ValidationError{Message: "剧集纳入状态无效"}
		}
		if episode.Inclusion == "included" {
			included++
		} else if strings.TrimSpace(episode.ExclusionReason) == "" {
			return &ValidationError{Message: "排除剧集必须说明原因"}
		}
		if len(episode.Sources) == 0 {
			return &ValidationError{Message: "每集至少需要一个素材引用"}
		}
		for sourceIndex := range episode.Sources {
			source := &episode.Sources[sourceIndex]
			source.Reference = strings.TrimSpace(source.Reference)
			if source.Reference == "" || !source.RightsConfirmed {
				return &ValidationError{Message: "所有素材必须有引用且已确认使用权限"}
			}
		}
	}
	if included < 2 {
		return &ValidationError{Message: "至少需要纳入两集素材"}
	}
	return nil
}

func buildProposal(request CreateRequest) Proposal {
	episodes := append([]Episode(nil), request.Episodes...)
	sort.Slice(episodes, func(i, j int) bool { return episodes[i].SequenceNumber < episodes[j].SequenceNumber })
	coverage, excluded, limitations := []string{}, []string{}, []string{}
	for _, episode := range episodes {
		if episode.Inclusion == "included" {
			coverage = append(coverage, episode.EpisodeID)
		} else {
			excluded = append(excluded, episode.EpisodeID)
		}
		if len(episode.Restrictions) > 0 {
			limitations = append(limitations, fmt.Sprintf("%s: %s", episode.EpisodeID, strings.Join(episode.Restrictions, "、")))
		}
	}
	if len(limitations) == 0 {
		limitations = []string{"使用已确认授权的素材，并遵守剧透限制"}
	}
	return Proposal{Hook: request.CampaignObjective, EmotionalArc: "从跨剧集冲突建立悬念，并在交付时保留故事张力", EpisodeCoverage: coverage, ExcludedEpisodes: excluded, DeliveryDefinition: fmt.Sprintf("%s，%d 秒，%s，面向 %s", request.DeliveryPreferences.AspectRatio, request.DeliveryPreferences.TargetDurationSecond, request.DeliveryPreferences.Language, request.DeliveryPreferences.DestinationLabel), CostEstimate: "待接入剪辑执行器后估算", SourceLimitations: limitations}
}

func decodeTask(stored model.DramaEditTask, decisions []model.DramaEditDecision) (Task, error) {
	result := Task{ID: stored.ID, DramaTitle: stored.DramaTitle, CampaignObjective: stored.CampaignObjective, TargetAudience: stored.TargetAudience, Status: string(stored.Status), Stage: stored.Stage, NextAction: stored.NextAction, CreatedAt: stored.CreatedAt, UpdatedAt: stored.UpdatedAt}
	if err := json.Unmarshal([]byte(stored.DeliveryJSON), &result.DeliveryPreferences); err != nil {
		return Task{}, err
	}
	if err := json.Unmarshal([]byte(stored.SpoilerJSON), &result.SpoilerPolicy); err != nil {
		return Task{}, err
	}
	if err := json.Unmarshal([]byte(stored.EpisodesJSON), &result.Episodes); err != nil {
		return Task{}, err
	}
	if err := json.Unmarshal([]byte(stored.ProposalJSON), &result.Proposal); err != nil {
		return Task{}, err
	}
	for _, item := range decisions {
		result.Decisions = append(result.Decisions, Decision{ID: item.ID, Action: item.Action, Note: item.Note, CreatedAt: item.CreatedAt})
	}
	return result, nil
}

func between(value string, min, max int) bool {
	count := len([]rune(value))
	return count >= min && count <= max
}
func validStatus(status string) bool {
	return containsStatus([]model.DramaEditStatus{model.DramaEditStatusAwaitingApproval, model.DramaEditStatusRevisionRequested, model.DramaEditStatusApproved, model.DramaEditStatusBlocked, model.DramaEditStatusCancelled}, model.DramaEditStatus(status))
}
func containsStatus(values []model.DramaEditStatus, target model.DramaEditStatus) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
