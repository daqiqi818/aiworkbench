package dramaedit

import (
	"errors"
	"time"

	"infinite-canvas/backend/internal/model"
	"infinite-canvas/backend/internal/repository"

	"gorm.io/gorm"
)

type store struct{ repo *repository.Repository }

func newStore(repo *repository.Repository) *store { return &store{repo: repo} }

func (s *store) create(task *model.DramaEditTask, decision *model.DramaEditDecision) error {
	return s.repo.Transaction(func(tx *repository.Repository) error {
		if err := tx.Create(task); err != nil {
			return err
		}
		return tx.Create(decision)
	})
}

func (s *store) list(userID string, status string) ([]model.DramaEditTask, error) {
	var tasks []model.DramaEditTask
	query := s.repo.DB().Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	return tasks, query.Order("created_at DESC, id DESC").Find(&tasks).Error
}

func (s *store) delete(userID, id string) error {
	return s.repo.Transaction(func(tx *repository.Repository) error {
		result := tx.DB().Where("id = ? AND user_id = ?", id, userID).Delete(&model.DramaEditTask{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrNotFound
		}
		return tx.DB().Where("task_id = ? AND user_id = ?", id, userID).Delete(&model.DramaEditDecision{}).Error
	})
}

func (s *store) episodeReferencedByTask(userID, assetID string) (bool, error) {
	var count int64
	err := s.repo.DB().Model(&model.DramaEditTask{}).
		Where("user_id = ? AND episodes_json LIKE ?", userID, "%\"assetId\":\""+assetID+"\"%").
		Count(&count).Error
	return count > 0, err
}

func (s *store) get(userID, id string) (*model.DramaEditTask, error) {
	var task model.DramaEditTask
	err := s.repo.DB().Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *store) byOpenMontageProduction(userID, productionID string) (*model.DramaEditTask, error) {
	var task model.DramaEditTask
	err := s.repo.DB().Where("user_id = ? AND open_montage_production_id = ?", userID, productionID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &task, err
}

func (s *store) byOpenMontageProductionAnyOwner(productionID string) (*model.DramaEditTask, error) {
	var task model.DramaEditTask
	err := s.repo.DB().Where("open_montage_production_id = ?", productionID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &task, err
}

func (s *store) decisions(userID, taskID string) ([]model.DramaEditDecision, error) {
	var decisions []model.DramaEditDecision
	err := s.repo.DB().Where("user_id = ? AND task_id = ?", userID, taskID).Order("created_at ASC, id ASC").Find(&decisions).Error
	return decisions, err
}

func (s *store) episodeFolders(userID string) ([]EpisodeFolderCandidate, error) {
	rows, err := s.repo.EpisodeAssetFolders(userID)
	if err != nil {
		return nil, err
	}
	candidates := make([]EpisodeFolderCandidate, 0, len(rows))
	for _, row := range rows {
		videos := make([]EpisodeAssetCandidate, 0, len(row.Videos))
		for _, video := range row.Videos {
			videos = append(videos, EpisodeAssetCandidate{AssetID: video.AssetID, AssetTitle: video.AssetTitle})
		}
		candidates = append(candidates, EpisodeFolderCandidate{FolderID: row.FolderID, FolderName: row.FolderName, VideoCount: row.VideoCount, Videos: videos})
	}
	return candidates, nil
}

func (s *store) transition(task *model.DramaEditTask, decision *model.DramaEditDecision, from []model.DramaEditStatus) error {
	return s.repo.Transaction(func(tx *repository.Repository) error {
		updates := map[string]any{
			"status": task.Status, "stage": task.Stage, "next_action": task.NextAction,
			"proposal_json": task.ProposalJSON, "updated_at": task.UpdatedAt,
		}
		result := tx.DB().Model(&model.DramaEditTask{}).Where("id = ? AND user_id = ? AND status IN ?", task.ID, task.UserID, from).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrInvalidTransition
		}
		return tx.Create(decision)
	})
}

func (s *store) task(userID, id string) (*model.Task, error) {
	var task model.Task
	err := s.repo.DB().Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &task, err
}

func (s *store) episodeResources(userID string, assetIDs []string) (map[string]model.Resource, error) {
	if len(assetIDs) == 0 {
		return map[string]model.Resource{}, nil
	}
	type row struct {
		AssetID string
		model.Resource
	}
	var rows []row
	err := s.repo.DB().Table("assets").
		Select("assets.id AS asset_id, resources.*").
		Joins("JOIN asset_versions ON asset_versions.id = assets.primary_version_id").
		Joins("JOIN asset_representations ON asset_representations.asset_version_id = asset_versions.id AND asset_representations.role = ?", "output").
		Joins("JOIN resources ON resources.id = asset_representations.resource_id AND resources.user_id = assets.user_id").
		Where("assets.user_id = ? AND assets.id IN ? AND resources.status = ?", userID, assetIDs, model.ResourceStatusReady).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]model.Resource, len(rows))
	for _, row := range rows {
		if _, exists := result[row.AssetID]; !exists {
			result[row.AssetID] = row.Resource
		}
	}
	// Browser-local assets have no asset-version representation until the user starts
	// production. Their client-side handoff persists data.storageKey as resource:<id>.
	// Resolve that owned, ready source as a fallback without weakening the established
	// representation lookup above.
	missing := make([]string, 0, len(assetIDs))
	for _, assetID := range assetIDs {
		if _, exists := result[assetID]; !exists {
			missing = append(missing, assetID)
		}
	}
	if len(missing) == 0 {
		return result, nil
	}
	var payloadRows []row
	err = s.repo.DB().Table("assets").
		Select("assets.id AS asset_id, resources.*").
		Joins("JOIN resources ON resources.id = substr(json_extract(assets.payload_json, '$.data.storageKey'), ?) AND resources.user_id = assets.user_id", len("resource:")+1).
		Where("assets.user_id = ? AND assets.id IN ? AND assets.kind = ? AND assets.status <> ? AND resources.status = ? AND resources.kind = ? AND resources.mime_type LIKE ?", userID, missing, "video", model.AssetVersionStatusArchived, model.ResourceStatusReady, "video", "video/%").
		Scan(&payloadRows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range payloadRows {
		if _, exists := result[row.AssetID]; !exists {
			result[row.AssetID] = row.Resource
		}
	}
	return result, nil
}

func (s *store) updateProduction(task *model.DramaEditTask) error {
	return s.repo.DB().Model(&model.DramaEditTask{}).Where("id = ? AND user_id = ?", task.ID, task.UserID).Updates(map[string]any{
		"status": task.Status, "stage": task.Stage, "next_action": task.NextAction,
		"production_task_id": task.ProductionTaskID, "delivery_resource_id": task.DeliveryResourceID,
		"production_error": task.ProductionError, "updated_at": task.UpdatedAt,
		"open_montage_production_id":      task.OpenMontageProductionID,
		"open_montage_access_token":       task.OpenMontageAccessToken,
		"open_montage_stage":              task.OpenMontageStage,
		"open_montage_checkpoint_status":  task.OpenMontageCheckpointStatus,
		"open_montage_report_resource_id": task.OpenMontageReportResourceID,
	}).Error
}

func nextID(repo *repository.Repository, prefix string) (string, error) {
	return repo.NextPrefixedID(prefix)
}

func now() time.Time { return time.Now().UTC() }
