package httpapi

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/service"
)

func TestManualJobResolution(t *testing.T) {
	dispatcher := ats.NewDispatcher()
	jobs := service.NewJobService(dispatcher)
	handler := NewHandler(log.New(io.Discard, "", 0), jobs, "http://localhost:4200")
	request := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/resolve", strings.NewReader(`{
		"title":"Engineer","company":"Acme","description":"Build useful things."
	}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"source":"manual"`) {
		t.Fatalf("expected manual posting, got %s", response.Body.String())
	}
}

func TestUnsupportedURL(t *testing.T) {
	dispatcher := ats.NewDispatcher()
	jobs := service.NewJobService(dispatcher)
	handler := NewHandler(log.New(io.Discard, "", 0), jobs, "http://localhost:4200")
	request := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/resolve", strings.NewReader(`{"url":"https://example.com/jobs/1"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request.WithContext(context.Background()))

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.Code)
	}
}
