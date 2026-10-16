package cosmos

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/example/job-application-assistant/apps/api/internal/domain"
	"github.com/example/job-application-assistant/apps/api/internal/repository"
)

type Repository struct {
	profiles     *azcosmos.ContainerClient
	applications *azcosmos.ContainerClient
	drafts       *azcosmos.ContainerClient
}

func New(client *azcosmos.Client, databaseName string) (*Repository, error) {
	database, err := client.NewDatabase(databaseName)
	if err != nil {
		return nil, err
	}
	profiles, err := database.NewContainer("profile")
	if err != nil {
		return nil, err
	}
	applications, err := database.NewContainer("applications")
	if err != nil {
		return nil, err
	}
	drafts, err := database.NewContainer("drafts")
	if err != nil {
		return nil, err
	}
	return &Repository{profiles: profiles, applications: applications, drafts: drafts}, nil
}

func (r *Repository) GetProfile(ctx context.Context, userID string) (domain.Profile, error) {
	response, err := r.profiles.ReadItem(ctx, azcosmos.NewPartitionKeyString(userID), "profile", nil)
	if isNotFound(err) {
		return domain.Profile{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.Profile{}, err
	}
	var profile domain.Profile
	if err := json.Unmarshal(response.Value, &profile); err != nil {
		return domain.Profile{}, err
	}
	return profile, nil
}

func (r *Repository) UpsertProfile(ctx context.Context, profile domain.Profile) error {
	return upsert(ctx, r.profiles, profile.UserID, profile)
}
func (r *Repository) SaveApplication(ctx context.Context, application domain.Application) error {
	return upsert(ctx, r.applications, application.UserID, application)
}
func (r *Repository) GetApplication(ctx context.Context, userID, applicationID string) (domain.Application, error) {
	response, err := r.applications.ReadItem(ctx, azcosmos.NewPartitionKeyString(userID), applicationID, nil)
	if isNotFound(err) {
		return domain.Application{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.Application{}, err
	}
	var application domain.Application
	if err := json.Unmarshal(response.Value, &application); err != nil {
		return domain.Application{}, err
	}
	return application, nil
}
func (r *Repository) SaveDraft(ctx context.Context, draft domain.Draft) error {
	return upsert(ctx, r.drafts, draft.ApplicationID, draft)
}

func (r *Repository) ListApplications(ctx context.Context, userID string) ([]domain.Application, error) {
	query := "SELECT * FROM c WHERE c.userId = @userId"
	options := &azcosmos.QueryOptions{QueryParameters: []azcosmos.QueryParameter{{Name: "@userId", Value: userID}}}
	pager := r.applications.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(userID), options)
	var result []domain.Application
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range page.Items {
			var application domain.Application
			if err := json.Unmarshal(item, &application); err != nil {
				return nil, err
			}
			result = append(result, application)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) ListDrafts(ctx context.Context, applicationID string) ([]domain.Draft, error) {
	query := "SELECT * FROM c WHERE c.applicationId = @applicationId"
	options := &azcosmos.QueryOptions{QueryParameters: []azcosmos.QueryParameter{{Name: "@applicationId", Value: applicationID}}}
	pager := r.drafts.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(applicationID), options)
	var result []domain.Draft
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range page.Items {
			var draft domain.Draft
			if err := json.Unmarshal(item, &draft); err != nil {
				return nil, err
			}
			result = append(result, draft)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func upsert(ctx context.Context, container *azcosmos.ContainerClient, partition string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = container.UpsertItem(ctx, azcosmos.NewPartitionKeyString(partition), encoded, nil)
	return err
}
func isNotFound(err error) bool {
	var responseError *azcore.ResponseError
	return errors.As(err, &responseError) && responseError.StatusCode == 404
}
