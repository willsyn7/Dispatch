package greenhouse

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/ats/atsutil"
	"github.com/example/job-application-assistant/apps/api/internal/domain"
)

type Source struct{ client *http.Client }

func New(client ...*http.Client) ats.JobSource {
	var selected *http.Client
	if len(client) > 0 {
		selected = client[0]
	}
	return &Source{client: atsutil.ClientOrDefault(selected)}
}

func (*Source) Name() string { return "greenhouse" }

func (*Source) Detect(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "greenhouse.io" || strings.HasSuffix(host, ".greenhouse.io") || host == "greenhouse.com" || strings.HasSuffix(host, ".greenhouse.com")
}

func (s *Source) GetJobDescription(ctx context.Context, rawURL string) (domain.JobPosting, error) {
	company, jobID, err := identifiers(rawURL)
	if err != nil {
		return domain.JobPosting{}, err
	}
	endpoint := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs/%s", url.PathEscape(company), url.PathEscape(jobID))
	var response struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		AbsoluteURL string `json:"absolute_url"`
		CompanyName string `json:"company_name"`
	}
	if err := atsutil.GetJSON(ctx, s.client, endpoint, &response); err != nil {
		return domain.JobPosting{}, fmt.Errorf("greenhouse: %w", err)
	}
	if response.CompanyName == "" {
		response.CompanyName = atsutil.DisplayName(company)
	}
	return domain.JobPosting{Source: s.Name(), CompanySlug: company, JobID: fmt.Sprint(response.ID), Title: response.Title, Company: response.CompanyName, Description: atsutil.HTMLToText(response.Content), URL: response.AbsoluteURL, FetchedAt: time.Now().UTC()}, nil
}

func identifiers(rawURL string) (string, string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", atsutil.ErrInvalidURL
	}
	parts := splitPath(parsed.Path)
	for index, part := range parts {
		if part == "jobs" && index > 0 && index+1 < len(parts) {
			return parts[index-1], parts[index+1], nil
		}
	}
	return "", "", atsutil.ErrInvalidURL
}

func splitPath(path string) []string {
	var result []string
	for _, part := range strings.Split(strings.Trim(path, "/"), "/") {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
