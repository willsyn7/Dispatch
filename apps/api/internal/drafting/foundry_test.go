package drafting

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/example/job-application-assistant/apps/api/internal/domain"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestFoundryGenerate(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != "https://example.openai.azure.com/openai/v1/chat/completions" {
			t.Fatalf("unexpected URL: %s", request.URL)
		}
		if request.Header.Get("api-key") != "secret" {
			t.Fatal("missing API key")
		}
		body := `{"choices":[{"message":{"content":"{\"coverLetter\":\"Dear team,\",\"matchSummary\":\"Strong Go experience.\"}"}}]}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	foundry, err := NewFoundry("https://example.openai.azure.com", "model-deployment", "secret", nil, client)
	if err != nil {
		t.Fatal(err)
	}
	result, err := foundry.Generate(context.Background(), domain.JobPosting{Title: "Engineer", Company: "Acme", Description: "Build APIs"}, domain.Profile{ResumeText: "Go developer"})
	if err != nil {
		t.Fatal(err)
	}
	if result.CoverLetter != "Dear team," || result.MatchSummary == "" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
