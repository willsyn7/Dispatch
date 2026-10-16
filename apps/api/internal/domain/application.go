package domain

import "time"

type ApplicationStatus string

const (
	StatusDrafting     ApplicationStatus = "drafting"
	StatusApplied      ApplicationStatus = "applied"
	StatusInterviewing ApplicationStatus = "interviewing"
	StatusRejected     ApplicationStatus = "rejected"
	StatusOffer        ApplicationStatus = "offer"
)

type Application struct {
	ID          string            `json:"id"`
	UserID      string            `json:"userId"`
	Company     string            `json:"company"`
	Title       string            `json:"title"`
	Job         JobPosting        `json:"job"`
	Status      ApplicationStatus `json:"status"`
	LatestDraft string            `json:"latestDraftId,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type Draft struct {
	ID             string    `json:"id"`
	ApplicationID  string    `json:"applicationId"`
	UserID         string    `json:"userId"`
	Version        int64     `json:"version"`
	ProfileVersion string    `json:"profileVersion"`
	CoverLetter    string    `json:"coverLetter"`
	MatchSummary   string    `json:"matchSummary"`
	CreatedAt      time.Time `json:"createdAt"`
}
