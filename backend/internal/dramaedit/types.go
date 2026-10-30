package dramaedit

import (
	"time"

	"infinite-canvas/backend/internal/editing"
	"infinite-canvas/backend/internal/model"
)

type Episode struct {
	FolderID       string `json:"folderId"`
	FolderName     string `json:"folderName"`
	AssetID        string `json:"assetId"`
	AssetTitle     string `json:"assetTitle"`
	SequenceNumber int    `json:"sequenceNumber"`
}

type DeliveryPreferences struct {
	AspectRatio        string `json:"aspectRatio"`
	MinDurationSeconds int    `json:"minDurationSeconds"`
	MaxDurationSeconds int    `json:"maxDurationSeconds"`
	Language           string `json:"language"`
	Platform           string `json:"platform"`
}

type SpoilerPolicy struct {
	Policy          string   `json:"policy"`
	RestrictedItems []string `json:"restrictedItems"`
}

type Proposal struct {
	Hook               string   `json:"hook"`
	EmotionalArc       string   `json:"emotionalArc"`
	EpisodeCoverage    []string `json:"episodeCoverage"`
	ExcludedEpisodes   []string `json:"excludedEpisodes"`
	DeliveryDefinition string   `json:"deliveryDefinition"`
	CostEstimate       string   `json:"costEstimate"`
	SourceLimitations  []string `json:"sourceLimitations"`
}

type CreateRequest struct {
	DramaTitle          string              `json:"dramaTitle"`
	DeliveryPreferences DeliveryPreferences `json:"deliveryPreferences"`
	SpoilerPolicy       SpoilerPolicy       `json:"spoilerPolicy"`
	EpisodeAssetIDs     []string            `json:"episodeAssetIds"`
}

type Decision struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Task struct {
	ID                  string              `json:"id"`
	DramaTitle          string              `json:"dramaTitle"`
	DeliveryPreferences DeliveryPreferences `json:"deliveryPreferences"`
	SpoilerPolicy       SpoilerPolicy       `json:"spoilerPolicy"`
	Episodes            []Episode           `json:"episodes"`
	Proposal            Proposal            `json:"proposal"`
	Status              string              `json:"status"`
	Stage               string              `json:"stage"`
	NextAction          string              `json:"nextAction"`
	ProductionTaskID    string              `json:"productionTaskId,omitempty"`
	ProductionStatus    string              `json:"productionStatus,omitempty"`
	ProductionProgress  int                 `json:"productionProgress,omitempty"`
	DeliveryResourceID  string              `json:"deliveryResourceId,omitempty"`
	DeliveryURL         string              `json:"deliveryUrl,omitempty"`
	ProductionError     string              `json:"productionError,omitempty"`
	Decisions           []Decision          `json:"decisions,omitempty"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
}

// ProductionStarter keeps the editing domain independent from the task runtime.
type ProductionStarter interface {
	StartProduction(string, editing.Project, editing.Options, string) (*model.Task, error)
}

type ProductionStarterFunc func(string, editing.Project, editing.Options, string) (*model.Task, error)

func (f ProductionStarterFunc) StartProduction(userID string, timeline editing.Project, options editing.Options, operationID string) (*model.Task, error) {
	return f(userID, timeline, options, operationID)
}

type EpisodeFolderCandidate struct {
	FolderID   string                  `json:"folderId"`
	FolderName string                  `json:"folderName"`
	VideoCount int                     `json:"videoCount"`
	Videos     []EpisodeAssetCandidate `json:"videos"`
}

type EpisodeAssetCandidate struct {
	AssetID    string `json:"assetId"`
	AssetTitle string `json:"assetTitle"`
}
