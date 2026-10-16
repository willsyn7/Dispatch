package domain

import "time"

type Profile struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Version    string    `json:"version"`
	ResumeText string    `json:"resumeText"`
	Background string    `json:"background,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
