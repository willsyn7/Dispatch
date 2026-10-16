package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/job-application-assistant/apps/api/internal/cache"
	"github.com/example/job-application-assistant/apps/api/internal/domain"
	"github.com/example/job-application-assistant/apps/api/internal/drafting"
	"github.com/example/job-application-assistant/apps/api/internal/repository"
)

var (
	ErrProfileRequired     = errors.New("save a resume profile before generating a draft")
	ErrResumeRequired      = errors.New("resume text is required")
	ErrPostingRequired     = errors.New("job title and description are required")
	ErrInvalidStatus       = errors.New("invalid application status")
	ErrApplicationNotFound = errors.New("application not found")
)

type ApplicationService struct {
	userID     string
	repository repository.Repository
	generator  drafting.Generator
	cache      cache.Store
}

func NewApplicationService(userID string, repo repository.Repository, generator drafting.Generator, store cache.Store) *ApplicationService {
	if store == nil {
		store = cache.Noop{}
	}
	return &ApplicationService{userID: userID, repository: repo, generator: generator, cache: store}
}

func (s *ApplicationService) GetProfile(ctx context.Context) (domain.Profile, error) {
	profile, err := s.repository.GetProfile(ctx, s.userID)
	if errors.Is(err, repository.ErrNotFound) {
		return domain.Profile{}, ErrProfileRequired
	}
	return profile, err
}

func (s *ApplicationService) SaveProfile(ctx context.Context, resumeText, background string) (domain.Profile, error) {
	resumeText = strings.TrimSpace(resumeText)
	background = strings.TrimSpace(background)
	if resumeText == "" {
		return domain.Profile{}, ErrResumeRequired
	}
	versionHash := sha256.Sum256([]byte(resumeText + "\x00" + background))
	profile := domain.Profile{ID: "profile", UserID: s.userID, Version: hex.EncodeToString(versionHash[:8]), ResumeText: resumeText, Background: background, UpdatedAt: time.Now().UTC()}
	if err := s.repository.UpsertProfile(ctx, profile); err != nil {
		return domain.Profile{}, err
	}
	return profile, nil
}

func (s *ApplicationService) GenerateDraft(ctx context.Context, posting domain.JobPosting) (domain.Draft, error) {
	if strings.TrimSpace(posting.Title) == "" || strings.TrimSpace(posting.Description) == "" {
		return domain.Draft{}, ErrPostingRequired
	}
	profile, err := s.GetProfile(ctx)
	if err != nil {
		return domain.Draft{}, err
	}
	jobFingerprint := fingerprint(posting.Source, posting.JobID, posting.URL, posting.Title, posting.Company, posting.Description)
	cacheKey := fmt.Sprintf("draft:%s:%s", jobFingerprint, profile.Version)
	var generated drafting.Result
	if found, _ := s.cache.Get(ctx, cacheKey, &generated); !found {
		if s.generator == nil {
			return domain.Draft{}, errors.New("drafting service is not configured")
		}
		generated, err = s.generator.Generate(ctx, posting, profile)
		if err != nil {
			return domain.Draft{}, err
		}
		_ = s.cache.Set(ctx, cacheKey, generated, 24*time.Hour)
	}

	now := time.Now().UTC()
	applicationID := "app_" + fingerprint(s.userID, jobFingerprint)[:20]
	draft := domain.Draft{ID: "draft_" + randomID(), ApplicationID: applicationID, UserID: s.userID, Version: now.UnixMilli(), ProfileVersion: profile.Version, CoverLetter: generated.CoverLetter, MatchSummary: generated.MatchSummary, CreatedAt: now}
	application := domain.Application{ID: applicationID, UserID: s.userID, Company: posting.Company, Title: posting.Title, Job: posting, Status: domain.StatusDrafting, LatestDraft: draft.ID, CreatedAt: now, UpdatedAt: now}
	if existing, getErr := s.repository.GetApplication(ctx, s.userID, applicationID); getErr == nil {
		application.CreatedAt = existing.CreatedAt
		application.Status = existing.Status
	} else if !errors.Is(getErr, repository.ErrNotFound) {
		return domain.Draft{}, getErr
	}
	if err := s.repository.SaveDraft(ctx, draft); err != nil {
		return domain.Draft{}, err
	}
	if err := s.repository.SaveApplication(ctx, application); err != nil {
		return domain.Draft{}, err
	}
	return draft, nil
}

func (s *ApplicationService) ListApplications(ctx context.Context) ([]domain.Application, error) {
	return s.repository.ListApplications(ctx, s.userID)
}

func (s *ApplicationService) UpdateStatus(ctx context.Context, applicationID string, status domain.ApplicationStatus) (domain.Application, error) {
	if !validStatus(status) {
		return domain.Application{}, ErrInvalidStatus
	}
	application, err := s.repository.GetApplication(ctx, s.userID, applicationID)
	if errors.Is(err, repository.ErrNotFound) {
		return domain.Application{}, ErrApplicationNotFound
	}
	if err != nil {
		return domain.Application{}, err
	}
	application.Status = status
	application.UpdatedAt = time.Now().UTC()
	if err := s.repository.SaveApplication(ctx, application); err != nil {
		return domain.Application{}, err
	}
	return application, nil
}

func (s *ApplicationService) ListDrafts(ctx context.Context, applicationID string) ([]domain.Draft, error) {
	if _, err := s.repository.GetApplication(ctx, s.userID, applicationID); errors.Is(err, repository.ErrNotFound) {
		return nil, ErrApplicationNotFound
	} else if err != nil {
		return nil, err
	}
	return s.repository.ListDrafts(ctx, applicationID)
}

func validStatus(status domain.ApplicationStatus) bool {
	switch status {
	case domain.StatusDrafting, domain.StatusApplied, domain.StatusInterviewing, domain.StatusRejected, domain.StatusOffer:
		return true
	default:
		return false
	}
}

func fingerprint(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte(part))
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
func randomID() string {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err != nil {
		return fmt.Sprint(time.Now().UnixNano())
	}
	return hex.EncodeToString(value)
}
