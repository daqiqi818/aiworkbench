package dramaedit

import "time"

type Source struct {
	Reference       string `json:"reference"`
	RightsConfirmed bool   `json:"rightsConfirmed"`
	Notes           string `json:"notes,omitempty"`
}

type Episode struct {
	EpisodeID       string   `json:"episodeId"`
	SequenceNumber  int      `json:"sequenceNumber"`
	Summary         string   `json:"summary"`
	Inclusion       string   `json:"inclusion"`
	ExclusionReason string   `json:"exclusionReason,omitempty"`
	Restrictions    []string `json:"restrictions,omitempty"`
	Sources         []Source `json:"sources"`
}

type DeliveryPreferences struct {
	AspectRatio          string `json:"aspectRatio"`
	TargetDurationSecond int    `json:"targetDurationSeconds"`
	Language             string `json:"language"`
	DestinationLabel     string `json:"destinationLabel"`
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
	CampaignObjective   string              `json:"campaignObjective"`
	TargetAudience      string              `json:"targetAudience"`
	DeliveryPreferences DeliveryPreferences `json:"deliveryPreferences"`
	SpoilerPolicy       SpoilerPolicy       `json:"spoilerPolicy"`
	Episodes            []Episode           `json:"episodes"`
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
	CampaignObjective   string              `json:"campaignObjective"`
	TargetAudience      string              `json:"targetAudience"`
	DeliveryPreferences DeliveryPreferences `json:"deliveryPreferences"`
	SpoilerPolicy       SpoilerPolicy       `json:"spoilerPolicy"`
	Episodes            []Episode           `json:"episodes"`
	Proposal            Proposal            `json:"proposal"`
	Status              string              `json:"status"`
	Stage               string              `json:"stage"`
	NextAction          string              `json:"nextAction"`
	Decisions           []Decision          `json:"decisions,omitempty"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
}
