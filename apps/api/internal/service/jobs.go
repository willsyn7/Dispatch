package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/cache"
	"github.com/example/job-application-assistant/apps/api/internal/domain"
)

var ErrMissingJobInput = errors.New("provide either url or description")

type JobResolver interface {
	Resolve(ctx context.Context, rawURL string) (domain.JobPosting, error)
}

type JobService struct {
	resolver JobResolver
	cache    cache.Store
}

func NewJobService(resolver JobResolver, stores ...cache.Store) *JobService {
	var store cache.Store = cache.Noop{}
	if len(stores) > 0 && stores[0] != nil {
		store = stores[0]
	}
	return &JobService{resolver: resolver, cache: store}
}

func (s *JobService) Resolve(ctx context.Context, rawURL, title, company, description string) (domain.JobPosting, error) {
	if strings.TrimSpace(rawURL) != "" {
		key := fmt.Sprintf("jobposting:url:%x", sha256.Sum256([]byte(strings.TrimSpace(rawURL))))
		var posting domain.JobPosting
		if found, _ := s.cache.Get(ctx, key, &posting); found {
			return posting, nil
		}
		posting, err := s.resolver.Resolve(ctx, rawURL)
		if err != nil {
			return domain.JobPosting{}, err
		}
		_ = s.cache.Set(ctx, key, posting, 12*time.Hour)
		return posting, nil
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
