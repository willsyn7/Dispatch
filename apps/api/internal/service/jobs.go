package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/domain"
)

var ErrMissingJobInput = errors.New("provide either url or description")

type JobResolver interface {
	Resolve(ctx context.Context, rawURL string) (domain.JobPosting, error)
}

type JobService struct {
	resolver JobResolver
}

func NewJobService(resolver JobResolver) *JobService {
	return &JobService{resolver: resolver}
}

func (s *JobService) Resolve(ctx context.Context, rawURL, title, company, description string) (domain.JobPosting, error) {
	if strings.TrimSpace(rawURL) != "" {
		return s.resolver.Resolve(ctx, rawURL)
	}
	if strings.TrimSpace(description) == "" {
		return domain.JobPosting{}, ErrMissingJobInput
	}
	return domain.JobPosting{
		Source:      "manual",
		Title:       strings.TrimSpace(title),
		Company:     strings.TrimSpace(company),
		Description: strings.TrimSpace(description),
		FetchedAt:   time.Now().UTC(),
	}, nil
}

var _ JobResolver = (*ats.Dispatcher)(nil)
