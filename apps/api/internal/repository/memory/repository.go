package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/example/job-application-assistant/apps/api/internal/domain"
	"github.com/example/job-application-assistant/apps/api/internal/repository"
)

type Repository struct {
	mu           sync.RWMutex
	profiles     map[string]domain.Profile
	applications map[string]domain.Application
	drafts       map[string]domain.Draft
}

func New() *Repository {
	return &Repository{profiles: map[string]domain.Profile{}, applications: map[string]domain.Application{}, drafts: map[string]domain.Draft{}}
}

func (r *Repository) GetProfile(_ context.Context, userID string) (domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.profiles[userID]
	if !ok {
		return domain.Profile{}, repository.ErrNotFound
	}
	return profile, nil
}
func (r *Repository) UpsertProfile(_ context.Context, profile domain.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[profile.UserID] = profile
	return nil
}
func (r *Repository) SaveApplication(_ context.Context, application domain.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.applications[application.ID] = application
	return nil
}
func (r *Repository) GetApplication(_ context.Context, userID, applicationID string) (domain.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	application, ok := r.applications[applicationID]
	if !ok || application.UserID != userID {
		return domain.Application{}, repository.ErrNotFound
	}
	return application, nil
}
func (r *Repository) SaveDraft(_ context.Context, draft domain.Draft) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drafts[draft.ID] = draft
	return nil
}
func (r *Repository) ListApplications(_ context.Context, userID string) ([]domain.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Application, 0)
	for _, application := range r.applications {
		if application.UserID == userID {
			result = append(result, application)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}
func (r *Repository) ListDrafts(_ context.Context, applicationID string) ([]domain.Draft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Draft, 0)
	for _, draft := range r.drafts {
		if draft.ApplicationID == applicationID {
			result = append(result, draft)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}
