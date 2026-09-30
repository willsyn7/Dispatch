package domain

import "time"

type JobPosting struct {
	Source      string    `json:"source"`
	CompanySlug string    `json:"companySlug,omitempty"`
	JobID       string    `json:"jobId,omitempty"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Description string    `json:"description"`
	URL         string    `json:"url,omitempty"`
	FetchedAt   time.Time `json:"fetchedAt"`
}
