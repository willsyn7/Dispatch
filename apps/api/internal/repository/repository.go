package repository

import (
	"context"
	"errors"

	"github.com/example/job-application-assistant/apps/api/internal/domain"
)

var ErrNotFound = errors.New("record not found")

type Repository interface {
	GetProfile(ctx context.Context, userID string) (domain.Profile, error)
	UpsertProfile(ctx context.Context, profile domain.Profile) error
	SaveApplication(ctx context.Context, application domain.Application) error
	GetApplication(ctx context.Context, userID, applicationID string) (domain.Application, error)
	SaveDraft(ctx context.Context, draft domain.Draft) error
	ListApplications(ctx context.Context, userID string) ([]domain.Application, error)
	ListDrafts(ctx context.Context, applicationID string) ([]domain.Draft, error)
}
