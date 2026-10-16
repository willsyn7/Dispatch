package service

import (
	"context"
	"testing"

	"github.com/example/job-application-assistant/apps/api/internal/domain"
	"github.com/example/job-application-assistant/apps/api/internal/drafting"
	"github.com/example/job-application-assistant/apps/api/internal/repository/memory"
)

type fakeGenerator struct{ calls int }

func (g *fakeGenerator) Generate(context.Context, domain.JobPosting, domain.Profile) (drafting.Result, error) {
	g.calls++
	return drafting.Result{CoverLetter: "Letter", MatchSummary: "Match"}, nil
}

func TestGenerateDraftPersistsApplication(t *testing.T) {
	ctx := context.Background()
	repo := memory.New()
	generator := &fakeGenerator{}
	service := NewApplicationService("owner", repo, generator, nil)
	if _, err := service.SaveProfile(ctx, "Go developer", "Distributed systems"); err != nil {
		t.Fatal(err)
	}
	draft, err := service.GenerateDraft(ctx, domain.JobPosting{Source: "manual", Title: "Engineer", Company: "Acme", Description: "Build APIs"})
	if err != nil {
		t.Fatal(err)
	}
	if draft.ApplicationID == "" || draft.CoverLetter != "Letter" {
		t.Fatalf("unexpected draft: %#v", draft)
	}
	applications, err := service.ListApplications(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(applications) != 1 || applications[0].LatestDraft != draft.ID {
		t.Fatalf("unexpected applications: %#v", applications)
	}
	updated, err := service.UpdateStatus(ctx, draft.ApplicationID, domain.StatusApplied)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.StatusApplied {
		t.Fatalf("status was not updated: %#v", updated)
	}
	drafts, err := service.ListDrafts(ctx, draft.ApplicationID)
	if err != nil {
		t.Fatal(err)
	}
	if len(drafts) != 1 || drafts[0].ID != draft.ID {
		t.Fatalf("unexpected draft history: %#v", drafts)
	}
}
