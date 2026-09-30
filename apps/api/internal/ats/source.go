package ats

import (
	"context"
	"errors"

	"github.com/example/job-application-assistant/apps/api/internal/domain"
)

var ErrUnsupportedURL = errors.New("no ATS adapter supports this URL")

type JobSource interface {
	Name() string
	Detect(rawURL string) bool
	GetJobDescription(ctx context.Context, rawURL string) (domain.JobPosting, error)
}

type Dispatcher struct {
	sources []JobSource
}

func NewDispatcher(sources ...JobSource) *Dispatcher {
	return &Dispatcher{sources: sources}
}

func (d *Dispatcher) Resolve(ctx context.Context, rawURL string) (domain.JobPosting, error) {
	for _, source := range d.sources {
		if source.Detect(rawURL) {
			return source.GetJobDescription(ctx, rawURL)
		}
	}
	return domain.JobPosting{}, ErrUnsupportedURL
}
