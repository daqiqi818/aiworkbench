package dramaedit

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"infinite-canvas/backend/internal/assets"
	"infinite-canvas/backend/internal/editing"
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
	repo         *repository.Repository
	store        *store
	production   ProductionStarter
	native       NativeRuntime
	importer     DeliveryImporter
	assetDeleter AssetDeleter
}

func NewNative(repo *repository.Repository, runtime NativeRuntime, importer DeliveryImporter, deleter ...AssetDeleter) *Service {
	service := &Service{repo: repo, store: newStore(repo), native: runtime, importer: importer}
	if len(deleter) > 0 {
		service.assetDeleter = deleter[0]
	}
	return service
}

func New(repo *repository.Repository, production ...ProductionStarter) *Service {
	service := &Service{repo: repo, store: newStore(repo)}
	if len(production) > 0 {
		service.production = production[0]
	}
	return service
}

func (s *Service) Create(userID string, request CreateRequest) (Task, error) {
	if strings.TrimSpace(userID) == "" {
		return Task{}, &ValidationError{Message: "当前用户无效"}
	}
	if err := validateRequest(&request); err != nil {
		return Task{}, err
	}
	episodes, err := s.selectedEpisodes(userID, request.EpisodeAssetIDs)
	if err != nil {
		return Task{}, err
	}
	delivery, _ := json.Marshal(request.DeliveryPreferences)
	spoiler, _ := json.Marshal(request.SpoilerPolicy)
	episodesJSON, _ := json.Marshal(episodes)
	proposal := buildProposal(request, episodes)
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
	stored := model.DramaEditTask{ID: id, UserID: userID, DramaTitle: request.DramaTitle, DeliveryJSON: string(delivery), SpoilerJSON: string(spoiler), EpisodesJSON: string(episodesJSON), ProposalJSON: string(proposalJSON), Status: model.DramaEditStatusAwaitingApproval, Stage: "等待审核", NextAction: "审核剪辑方案后批准、请求修改或取消", CreatedAt: timestamp, UpdatedAt: timestamp}
	decision := model.DramaEditDecision{ID: decisionID, TaskID: id, UserID: userID, Action: "created", CreatedAt: timestamp}
	if err := s.store.create(&stored, &decision); err != nil {
		return Task{}, err
	}
	return decodeTask(stored, []model.DramaEditDecision{decision})
}

func (s *Service) EpisodeFolders(userID string) ([]EpisodeFolderCandidate, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, &ValidationError{Message: "当前用户无效"}
	}
	return s.store.episodeFolders(userID)
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
		task, decodeErr := s.projectTask(item)
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
	result, err := s.projectTask(*stored)
	if err != nil {
		return Task{}, err
	}
	result.Decisions = nil
	for _, item := range decisions {
		result.Decisions = append(result.Decisions, Decision{ID: item.ID, Action: item.Action, Note: item.Note, CreatedAt: item.CreatedAt})
	}
	return result, nil
}

func (s *Service) Approve(userID, id string) (Task, error) {
	task, err := s.decide(userID, id, "approved", "", []model.DramaEditStatus{model.DramaEditStatusAwaitingApproval, model.DramaEditStatusRevisionRequested})
	if err != nil {
		return Task{}, err
	}
	return s.startProduction(userID, task.ID, false)
}

func (s *Service) StartProduction(userID, id string) (Task, error) {
	return s.startProduction(userID, id, true)
}
func (s *Service) Delete(userID string, request DeleteRequest) ([]DeleteResult, error) {
	ids := request.IDs
	if len(ids) == 0 || len(ids) > 100 {
		return nil, &ValidationError{Message: "请选择 1 到 100 个任务"}
	}
	if request.DeleteEpisodeFiles && s.assetDeleter == nil {
		return nil, &ValidationError{Message: "剧集文件删除服务不可用"}
	}
	results := make([]DeleteResult, 0, len(ids))
	filesByTask := make(map[string][]string)
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		stored, err := s.store.get(userID, id)
		if err != nil {
			results = append(results, DeleteResult{ID: id, Reason: "任务不存在或无权访问"})
			continue
		}
		if stored.Status == model.DramaEditStatusProcessing || stored.Status == model.DramaEditStatusApproved {
			if stored.OpenMontageProductionID == "" || s.native == nil {
				results = append(results, DeleteResult{ID: id, Reason: "进行中的任务缺少可用的 OpenMontage 停止通道，未删除"})
				continue
			}
			state, cancelErr := s.native.Cancel(userID, stored.OpenMontageProductionID)
			if cancelErr != nil {
				results = append(results, DeleteResult{ID: id, Reason: "无法停止 OpenMontage 制作，未删除：" + cancelErr.Error()})
				continue
			}
			if state.Status != "cancelled" {
				results = append(results, DeleteResult{ID: id, Reason: "OpenMontage 未确认停止，未删除"})
				continue
			}
		}
		if err := s.store.delete(userID, id); err != nil {
			return nil, err
		}
		if request.DeleteEpisodeFiles {
			var episodes []Episode
			if err := json.Unmarshal([]byte(stored.EpisodesJSON), &episodes); err != nil {
				return nil, err
			}
			for _, episode := range episodes {
				filesByTask[id] = append(filesByTask[id], episode.AssetID)
			}
		}
		results = append(results, DeleteResult{ID: id, Deleted: true})
	}
	if !request.DeleteEpisodeFiles {
		return results, nil
	}
	fileResults := make(map[string]EpisodeFileDeleteResult)
	for _, assetID := range uniqueAssetIDs(filesByTask) {
		referenced, err := s.store.episodeReferencedByTask(userID, assetID)
		if err != nil {
			return nil, err
		}
		if referenced {
			fileResults[assetID] = EpisodeFileDeleteResult{AssetID: assetID, Reason: "仍被其他短剧剪辑任务引用"}
			continue
		}
		if err := s.assetDeleter.DeleteUserAsset(userID, assetID); err != nil {
			fileResults[assetID] = EpisodeFileDeleteResult{AssetID: assetID, Reason: err.Error()}
			continue
		}
		fileResults[assetID] = EpisodeFileDeleteResult{AssetID: assetID, Deleted: true}
	}
	for index := range results {
		for _, assetID := range filesByTask[results[index].ID] {
			results[index].EpisodeFiles = append(results[index].EpisodeFiles, fileResults[assetID])
		}
	}
	return results, nil
}

func uniqueAssetIDs(filesByTask map[string][]string) []string {
	seen := map[string]struct{}{}
	ids := []string{}
	for _, values := range filesByTask {
		for _, id := range values {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}
	return ids
}
func (s *Service) Cancel(userID, id, note string) (Task, error) {
	if stored, err := s.store.get(userID, id); err == nil && stored.OpenMontageProductionID != "" && s.native != nil {
		if _, cancelErr := s.native.Cancel(userID, stored.OpenMontageProductionID); cancelErr != nil {
			return Task{}, &ValidationError{Message: "无法取消 OpenMontage 制作：" + cancelErr.Error()}
		}
	}
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
		task.Status, task.Stage, task.NextAction = model.DramaEditStatusApproved, "已批准", "正在创建本地制作工作"
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
	request.DramaTitle = strings.TrimSpace(request.DramaTitle)
	if !between(request.DramaTitle, 1, 200) {
		return &ValidationError{Message: "短剧名称不能为空且长度必须有效"}
	}
	if request.DeliveryPreferences.AspectRatio != "9:16" && request.DeliveryPreferences.AspectRatio != "16:9" && request.DeliveryPreferences.AspectRatio != "1:1" {
		return &ValidationError{Message: "画幅比例无效"}
	}
	if request.DeliveryPreferences.MinDurationSeconds < 15 || request.DeliveryPreferences.MinDurationSeconds > 300 || request.DeliveryPreferences.MaxDurationSeconds < 15 || request.DeliveryPreferences.MaxDurationSeconds > 300 || request.DeliveryPreferences.MinDurationSeconds > request.DeliveryPreferences.MaxDurationSeconds {
		return &ValidationError{Message: "目标时长范围必须在 15 到 300 秒之间，且最小值不能大于最大值"}
	}
	if request.DeliveryPreferences.Language != "zh" && request.DeliveryPreferences.Language != "en" {
		return &ValidationError{Message: "仅支持中文或英文"}
	}
	if request.DeliveryPreferences.Platform != "tiktok" && request.DeliveryPreferences.Platform != "facebook" && request.DeliveryPreferences.Platform != "wechat_channels" && request.DeliveryPreferences.Platform != "douyin" {
		return &ValidationError{Message: "投放平台无效"}
	}
	if request.SpoilerPolicy.Policy != "no-major-spoilers" && request.SpoilerPolicy.Policy != "limited-spoilers" && request.SpoilerPolicy.Policy != "approved-spoilers" {
		return &ValidationError{Message: "剧透策略无效"}
	}
	if len(request.EpisodeAssetIDs) < 2 || len(request.EpisodeAssetIDs) > 12 {
		return &ValidationError{Message: "每项剪辑任务必须包含 2 到 12 集"}
	}
	seen := make(map[string]struct{}, len(request.EpisodeAssetIDs))
	for index, id := range request.EpisodeAssetIDs {
		id = strings.TrimSpace(id)
		if id == "" || len(id) > 80 {
			return &ValidationError{Message: "剧集视频无效"}
		}
		if _, exists := seen[id]; exists {
			return &ValidationError{Message: "剧集视频不能重复"}
		}
		seen[id] = struct{}{}
		request.EpisodeAssetIDs[index] = id
	}
	return nil
}

func (s *Service) selectedEpisodes(userID string, ids []string) ([]Episode, error) {
	candidates, err := s.store.episodeFolders(userID)
	if err != nil {
		return nil, err
	}
	type source struct {
		folder EpisodeFolderCandidate
		video  EpisodeAssetCandidate
	}
	byID := make(map[string]source)
	for _, candidate := range candidates {
		for _, video := range candidate.Videos {
			byID[video.AssetID] = source{folder: candidate, video: video}
		}
	}
	episodes := make([]Episode, 0, len(ids))
	for index, id := range ids {
		source, ok := byID[id]
		if !ok {
			return nil, &ValidationError{Message: "所选剧集视频不存在、无权访问或已不可用"}
		}
		episodes = append(episodes, Episode{FolderID: source.folder.FolderID, FolderName: source.folder.FolderName, AssetID: source.video.AssetID, AssetTitle: source.video.AssetTitle, SequenceNumber: index + 1})
	}
	return episodes, nil
}

func buildProposal(request CreateRequest, selected []Episode) Proposal {
	episodes := append([]Episode(nil), selected...)
	sort.Slice(episodes, func(i, j int) bool { return episodes[i].SequenceNumber < episodes[j].SequenceNumber })
	coverage, limitations := []string{}, []string{}
	for _, episode := range episodes {
		coverage = append(coverage, episode.AssetTitle)
		limitations = append(limitations, fmt.Sprintf("%s / %s：已确认可用视频素材", episode.FolderName, episode.AssetTitle))
	}
	if len(limitations) == 0 {
		limitations = []string{"使用已确认授权的素材，并遵守剧透限制"}
	}
	return Proposal{Hook: "从已选剧集素材中建立叙事钩子", EmotionalArc: "从跨剧集冲突建立悬念，并在交付时保留故事张力", EpisodeCoverage: coverage, DeliveryDefinition: fmt.Sprintf("%s，%d-%d 秒，%s，投放 %s", request.DeliveryPreferences.AspectRatio, request.DeliveryPreferences.MinDurationSeconds, request.DeliveryPreferences.MaxDurationSeconds, languageLabel(request.DeliveryPreferences.Language), platformLabel(request.DeliveryPreferences.Platform)), CostEstimate: "待接入剪辑执行器后估算", SourceLimitations: limitations}
}

func decodeTask(stored model.DramaEditTask, decisions []model.DramaEditDecision) (Task, error) {
	result := Task{ID: stored.ID, DramaTitle: stored.DramaTitle, Status: string(stored.Status), Stage: stored.Stage, NextAction: stored.NextAction, ProductionTaskID: stored.ProductionTaskID, DeliveryResourceID: stored.DeliveryResourceID, ProductionError: stored.ProductionError, OpenMontageProductionID: stored.OpenMontageProductionID, OpenMontageStage: stored.OpenMontageStage, OpenMontageCheckpointStatus: stored.OpenMontageCheckpointStatus, RenderReportResourceID: stored.OpenMontageReportResourceID, CreatedAt: stored.CreatedAt, UpdatedAt: stored.UpdatedAt}
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

func (s *Service) startProduction(userID, id string, historical bool) (Task, error) {
	stored, err := s.store.get(userID, id)
	if err != nil {
		return Task{}, err
	}
	if s.native != nil {
		return s.startNativeProduction(userID, stored)
	}
	retry := false
	if stored.ProductionTaskID != "" {
		work, lookupErr := s.store.task(userID, stored.ProductionTaskID)
		if lookupErr != nil {
			return Task{}, lookupErr
		}
		if work != nil && (work.Status == model.TaskStatusQueued || work.Status == model.TaskStatusRunning || work.Status == model.TaskStatusSucceeded) {
			return s.Get(userID, id)
		}
		stored.ProductionTaskID, stored.DeliveryResourceID, stored.ProductionError = "", "", ""
		stored.Status = model.DramaEditStatusApproved
		retry = true
	}
	if stored.Status != model.DramaEditStatusApproved && stored.Status != model.DramaEditStatusBlocked {
		return Task{}, ErrInvalidTransition
	}
	if s.production == nil {
		if historical {
			return Task{}, &ValidationError{Message: "本地制作服务不可用，请启动工作台后重试"}
		}
		return s.Get(userID, id)
	}
	var delivery DeliveryPreferences
	var episodes []Episode
	if err := json.Unmarshal([]byte(stored.DeliveryJSON), &delivery); err != nil {
		return Task{}, err
	}
	if err := json.Unmarshal([]byte(stored.EpisodesJSON), &episodes); err != nil {
		return Task{}, err
	}
	assetIDs := make([]string, 0, len(episodes))
	for _, episode := range episodes {
		assetIDs = append(assetIDs, episode.AssetID)
	}
	resources, err := s.store.episodeResources(userID, assetIDs)
	if err != nil {
		return Task{}, err
	}
	timeline, options, err := productionTimeline(episodes, resources, delivery)
	if err != nil {
		return s.blockProduction(stored, err.Error())
	}
	operationID := "drama-edit:" + stored.ID
	if retry {
		// A terminal work must not reuse its idempotency key: the task runtime
		// correctly returns the previous failed work for that key.
		operationID = fmt.Sprintf("%s:retry:%d", operationID, now().UnixNano())
	}
	work, err := s.production.StartProduction(userID, timeline, options, operationID)
	if err != nil {
		return s.blockProduction(stored, "无法创建本地制作工作："+err.Error())
	}
	stored.ProductionTaskID = work.ID
	stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusProcessing, "制作排队中", "本地制作工作已入队，等待开始", ""
	stored.UpdatedAt = now()
	if err := s.store.updateProduction(stored); err != nil {
		return Task{}, err
	}
	return s.Get(userID, id)
}

func (s *Service) startNativeProduction(userID string, stored *model.DramaEditTask) (Task, error) {
	if stored.Status != model.DramaEditStatusApproved && stored.Status != model.DramaEditStatusBlocked {
		return Task{}, ErrInvalidTransition
	}
	if stored.OpenMontageProductionID != "" {
		return s.Get(userID, stored.ID)
	}
	var delivery DeliveryPreferences
	var episodes []Episode
	if err := json.Unmarshal([]byte(stored.DeliveryJSON), &delivery); err != nil {
		return Task{}, err
	}
	if err := json.Unmarshal([]byte(stored.EpisodesJSON), &episodes); err != nil {
		return Task{}, err
	}
	assetIDs := make([]string, 0, len(episodes))
	for _, episode := range episodes {
		assetIDs = append(assetIDs, episode.AssetID)
	}
	resources, err := s.store.episodeResources(userID, assetIDs)
	if err != nil {
		return Task{}, err
	}
	resourceIDs := make([]string, 0, len(episodes))
	for _, episode := range episodes {
		resource, exists := resources[episode.AssetID]
		if !exists {
			return s.blockProduction(stored, "所选剧集没有可读取的视频资源")
		}
		resourceIDs = append(resourceIDs, resource.ID)
	}
	productionID := "om-" + stored.ID
	accessToken, err := newProductionToken()
	if err != nil {
		return Task{}, err
	}
	state, err := s.native.Create(userID, productionID, stored.DramaTitle, delivery, resourceIDs, accessToken)
	if err != nil {
		return s.blockProduction(stored, "无法创建 OpenMontage 制作："+err.Error())
	}
	// Persist the scoped broker capability before the Runner begins reading any
	// source. The broker resolves the production back to this task on each read.
	stored.OpenMontageProductionID, stored.OpenMontageAccessToken = productionID, accessToken
	applyNativeProjection(stored, state)
	stored.UpdatedAt = now()
	if err := s.store.updateProduction(stored); err != nil {
		return Task{}, err
	}
	state, err = s.native.Run(userID, productionID)
	if err != nil {
		return s.blockProduction(stored, "无法启动 OpenMontage 制作："+err.Error())
	}
	applyNativeProjection(stored, state)
	stored.UpdatedAt = now()
	if err := s.store.updateProduction(stored); err != nil {
		return Task{}, err
	}
	return s.Get(userID, stored.ID)
}

func newProductionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (s *Service) ApproveNativeStage(userID, id, stage, note string) (Task, error) {
	stored, err := s.store.get(userID, id)
	if err != nil {
		return Task{}, err
	}
	if s.native == nil || stored.OpenMontageProductionID == "" {
		return Task{}, ErrInvalidTransition
	}
	state, err := s.native.ApproveStage(userID, stored.OpenMontageProductionID, strings.TrimSpace(stage), strings.TrimSpace(note))
	if err != nil {
		return s.blockProduction(stored, "OpenMontage 阶段审批失败："+err.Error())
	}
	applyNativeProjection(stored, state)
	stored.UpdatedAt = now()
	if err := s.store.updateProduction(stored); err != nil {
		return Task{}, err
	}
	return s.Get(userID, id)
}

func (s *Service) ValidateOpenMontageSource(userID, productionID, resourceID, accessToken string) error {
	stored, err := s.store.byOpenMontageProduction(userID, productionID)
	if err != nil {
		return err
	}
	if stored.OpenMontageAccessToken == "" || accessToken != stored.OpenMontageAccessToken {
		return &ValidationError{Message: "OpenMontage 源资源访问未授权"}
	}
	var episodes []Episode
	if err := json.Unmarshal([]byte(stored.EpisodesJSON), &episodes); err != nil {
		return err
	}
	resources, err := s.store.episodeResources(userID, episodeAssetIDs(episodes))
	if err != nil {
		return err
	}
	for _, resource := range resources {
		if resource.ID == resourceID {
			return nil
		}
	}
	return &ValidationError{Message: "资源不属于此 OpenMontage 制作"}
}

func (s *Service) ValidateOpenMontageSourceByProduction(productionID, resourceID, accessToken string) (string, error) {
	stored, err := s.store.byOpenMontageProductionAnyOwner(productionID)
	if err != nil {
		return "", err
	}
	if err := s.ValidateOpenMontageSource(stored.UserID, productionID, resourceID, accessToken); err != nil {
		return "", err
	}
	return stored.UserID, nil
}

func episodeAssetIDs(episodes []Episode) []string {
	ids := make([]string, 0, len(episodes))
	for _, episode := range episodes {
		ids = append(ids, episode.AssetID)
	}
	return ids
}

func (s *Service) blockProduction(stored *model.DramaEditTask, reason string) (Task, error) {
	stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusBlocked, "制作受阻", "检查素材或本地制作环境后重新开始制作", reason
	stored.UpdatedAt = now()
	if err := s.store.updateProduction(stored); err != nil {
		return Task{}, err
	}
	return s.Get(stored.UserID, stored.ID)
}

func (s *Service) projectTask(stored model.DramaEditTask) (Task, error) {
	if stored.OpenMontageProductionID != "" && s.native != nil {
		state, err := s.native.Status(stored.UserID, stored.OpenMontageProductionID)
		if err != nil {
			return s.blockProduction(&stored, "无法读取 OpenMontage 制作状态："+err.Error())
		}
		applyNativeProjection(&stored, state)
		if state.Status == "completed" && (stored.DeliveryResourceID == "" || stored.OpenMontageReportResourceID == "") {
			if s.importer == nil {
				return s.blockProduction(&stored, "OpenMontage 交付导入服务不可用")
			}
			var episodes []Episode
			if err := json.Unmarshal([]byte(stored.EpisodesJSON), &episodes); err != nil || len(episodes) == 0 {
				return s.blockProduction(&stored, "无法确定成片归属文件夹")
			}
			video, report, importErr := s.importer.ImportOpenMontageDelivery(stored.UserID, episodes[0].FolderID, stored.DramaTitle, stored.OpenMontageProductionID, s.native)
			if importErr != nil {
				return s.blockProduction(&stored, "无法将 OpenMontage 成片导入 BeefTV 资产："+importErr.Error())
			}
			stored.DeliveryResourceID, stored.OpenMontageReportResourceID = video, report
			stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusCompleted, "成片已回传资产", "查看或下载成片和渲染报告", ""
		}
		if state.Status == "completed" && stored.DeliveryResourceID != "" && stored.OpenMontageReportResourceID != "" {
			stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusCompleted, "成片已回传资产", "查看或下载成片和渲染报告", ""
		}
		stored.UpdatedAt = now()
		if err := s.store.updateProduction(&stored); err != nil {
			return Task{}, err
		}
		result, err := decodeTask(stored, nil)
		if err != nil {
			return Task{}, err
		}
		result.OpenMontageRequiresApproval = state.RequiresApproval
		result.ProductionProgress = nativeProgress(state)
		if stored.DeliveryResourceID != "" {
			result.DeliveryURL = assets.FileURL(stored.DeliveryResourceID)
		}
		if stored.OpenMontageReportResourceID != "" {
			result.RenderReportURL = assets.FileURL(stored.OpenMontageReportResourceID)
		}
		return result, nil
	}
	if stored.ProductionTaskID != "" {
		work, err := s.store.task(stored.UserID, stored.ProductionTaskID)
		if err != nil {
			return Task{}, err
		}
		if work == nil {
			stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusBlocked, "制作受阻", "本地制作工作不存在，请重新开始制作", "关联的本地制作工作已丢失"
		} else {
			applyProductionProjection(&stored, work)
		}
		if err := s.store.updateProduction(&stored); err != nil {
			return Task{}, err
		}
	}
	result, err := decodeTask(stored, nil)
	if err != nil {
		return Task{}, err
	}
	if stored.DeliveryResourceID != "" {
		result.DeliveryURL = assets.FileURL(stored.DeliveryResourceID)
	}
	if stored.ProductionTaskID != "" {
		if work, err := s.store.task(stored.UserID, stored.ProductionTaskID); err == nil && work != nil {
			result.ProductionStatus, result.ProductionProgress = string(work.Status), work.Progress
		}
	}
	return result, nil
}

func nativeProgress(state NativeState) int {
	if state.Status == "completed" {
		return 100
	}
	stages := []string{"research", "proposal", "script", "scene_plan", "assets", "edit", "compose", "publish"}
	for index, stage := range stages {
		if state.CurrentStage == stage {
			return (index * 100) / (len(stages) - 1)
		}
	}
	return 0
}

func applyNativeProjection(stored *model.DramaEditTask, state NativeState) {
	stored.OpenMontageStage, stored.OpenMontageCheckpointStatus = state.CurrentStage, state.CheckpointStatus
	switch state.Status {
	case "awaiting_approval":
		stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusProcessing, "OpenMontage："+state.CurrentStage, "审核并批准当前 OpenMontage 阶段", ""
	case "ready_to_run", "running":
		stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusProcessing, "OpenMontage："+state.CurrentStage, "OpenMontage 制作进行中", ""
	case "completed":
		stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusProcessing, "OpenMontage 成片导入中", "正在导入成片和渲染报告到 BeefTV 资产", ""
	case "cancelled":
		stored.Status, stored.Stage, stored.NextAction = model.DramaEditStatusCancelled, "已取消", "任务已停止"
	case "blocked":
		stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusBlocked, "OpenMontage 制作受阻", "检查 OpenMontage 制作状态后重新开始", state.Error
		if stored.ProductionError == "" {
			stored.ProductionError = "OpenMontage 制作未完成"
		}
	default:
		stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusBlocked, "OpenMontage 制作受阻", "检查 OpenMontage 制作状态后重新开始", "OpenMontage 返回未知状态："+state.Status
	}
}

func applyProductionProjection(stored *model.DramaEditTask, work *model.Task) {
	switch work.Status {
	case model.TaskStatusQueued, model.TaskStatusRunning:
		stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusProcessing, work.Stage, "本地制作进行中", ""
	case model.TaskStatusSucceeded:
		var output struct {
			ResourceID string `json:"resourceId"`
		}
		if err := json.Unmarshal([]byte(work.ResultJSON), &output); err != nil || strings.TrimSpace(output.ResourceID) == "" {
			stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusBlocked, "制作受阻", "检查制作结果后重新开始制作", "本地制作未返回有效成片"
			return
		}
		stored.Status, stored.Stage, stored.NextAction, stored.DeliveryResourceID, stored.ProductionError = model.DramaEditStatusCompleted, "成片已生成", "查看或下载成片", output.ResourceID, ""
	case model.TaskStatusFailed, model.TaskStatusCancelled:
		stored.Status, stored.Stage, stored.NextAction, stored.ProductionError = model.DramaEditStatusBlocked, "制作受阻", "检查失败原因后重新开始制作", strings.TrimSpace(work.Error)
		if stored.ProductionError == "" {
			stored.ProductionError = "本地制作未完成"
		}
	}
}

func productionTimeline(episodes []Episode, resources map[string]model.Resource, delivery DeliveryPreferences) (editing.Project, editing.Options, error) {
	options := editing.DefaultOptions()
	switch delivery.AspectRatio {
	case "9:16":
		options.Width, options.Height = 1080, 1920
	case "1:1":
		options.Width, options.Height = 1080, 1080
	}
	remaining := int64(delivery.MaxDurationSeconds) * 1000
	minimum := int64(delivery.MinDurationSeconds) * 1000
	timeline := editing.Project{Version: editing.ProjectVersion, Tracks: []editing.Track{{ID: "episodes", Kind: editing.KindVideo}}}
	for index, episode := range episodes {
		resource, ok := resources[episode.AssetID]
		if !ok || !strings.HasPrefix(resource.MimeType, "video/") || resource.DurationMs <= 0 {
			return editing.Project{}, editing.Options{}, fmt.Errorf("%s 没有可读取的视频资源", episode.AssetTitle)
		}
		if remaining <= 0 {
			break
		}
		duration := min(resource.DurationMs, remaining)
		timeline.Clips = append(timeline.Clips, editing.Clip{ID: fmt.Sprintf("episode-%d", index+1), Kind: editing.KindVideo, NodeID: episode.AssetID, TrackID: "episodes", StartMs: timeline.DurationMs, DurationMs: duration, Volume: 1, DirectMedia: &editing.DirectMedia{ID: resource.ID, Kind: editing.KindVideo, StorageKey: "resource:" + resource.ID}})
		timeline.DurationMs += duration
		remaining -= duration
	}
	if timeline.DurationMs < minimum {
		return editing.Project{}, editing.Options{}, fmt.Errorf("可用素材总时长不足 %d 秒", delivery.MinDurationSeconds)
	}
	return timeline, options, nil
}

func languageLabel(value string) string {
	if value == "en" {
		return "英文"
	}
	return "中文"
}

func platformLabel(value string) string {
	switch value {
	case "tiktok":
		return "TikTok"
	case "facebook":
		return "Facebook"
	case "wechat_channels":
		return "视频号"
	default:
		return "抖音"
	}
}

func between(value string, min, max int) bool {
	count := len([]rune(value))
	return count >= min && count <= max
}
func validStatus(status string) bool {
	return containsStatus([]model.DramaEditStatus{model.DramaEditStatusAwaitingApproval, model.DramaEditStatusRevisionRequested, model.DramaEditStatusApproved, model.DramaEditStatusProcessing, model.DramaEditStatusCompleted, model.DramaEditStatusBlocked, model.DramaEditStatusCancelled}, model.DramaEditStatus(status))
}
func containsStatus(values []model.DramaEditStatus, target model.DramaEditStatus) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
