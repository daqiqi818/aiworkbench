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
	return tasks, query.Order("updated_at DESC, id DESC").Find(&tasks).Error
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

func (s *store) decisions(userID, taskID string) ([]model.DramaEditDecision, error) {
	var decisions []model.DramaEditDecision
	err := s.repo.DB().Where("user_id = ? AND task_id = ?", userID, taskID).Order("created_at ASC, id ASC").Find(&decisions).Error
	return decisions, err
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

func nextID(repo *repository.Repository, prefix string) (string, error) {
	return repo.NextPrefixedID(prefix)
}

func now() time.Time { return time.Now().UTC() }
