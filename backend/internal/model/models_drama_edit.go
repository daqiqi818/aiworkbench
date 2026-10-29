package model

import "time"

type DramaEditStatus string

const (
	DramaEditStatusAwaitingApproval  DramaEditStatus = "awaiting_approval"
	DramaEditStatusRevisionRequested DramaEditStatus = "revision_requested"
	DramaEditStatusApproved          DramaEditStatus = "approved"
	DramaEditStatusBlocked           DramaEditStatus = "blocked"
	DramaEditStatusCancelled         DramaEditStatus = "cancelled"
)

type DramaEditTask struct {
	ID                string          `json:"id" gorm:"primaryKey;size:36"`
	UserID            string          `json:"userId" gorm:"index;size:36;index:idx_drama_edit_tasks_user_updated,priority:1"`
	DramaTitle        string          `json:"dramaTitle" gorm:"size:200"`
	CampaignObjective string          `json:"campaignObjective" gorm:"type:text"`
	TargetAudience    string          `json:"targetAudience" gorm:"type:text"`
	DeliveryJSON      string          `json:"delivery" gorm:"type:text"`
	SpoilerJSON       string          `json:"spoilerPolicy" gorm:"type:text"`
	EpisodesJSON      string          `json:"episodes" gorm:"type:text"`
	ProposalJSON      string          `json:"proposal" gorm:"type:text"`
	BlockerJSON       string          `json:"blocker,omitempty" gorm:"type:text"`
	Status            DramaEditStatus `json:"status" gorm:"index;size:32;index:idx_drama_edit_tasks_user_updated,priority:2"`
	Stage             string          `json:"stage" gorm:"size:120"`
	NextAction        string          `json:"nextAction" gorm:"size:500"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt" gorm:"index:idx_drama_edit_tasks_user_updated,priority:3"`
}

type DramaEditDecision struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	TaskID    string    `json:"taskId" gorm:"index;size:36;index:idx_drama_edit_decisions_task_created,priority:1"`
	UserID    string    `json:"userId" gorm:"index;size:36"`
	Action    string    `json:"action" gorm:"size:48"`
	Note      string    `json:"note" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"index:idx_drama_edit_decisions_task_created,priority:2"`
}
