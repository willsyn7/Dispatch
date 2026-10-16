package ashby

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
func (*Source) Name() string { return "ashby" }
func (*Source) Detect(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "ashbyhq.com" || strings.HasSuffix(host, ".ashbyhq.com")
}

func (s *Source) GetJobDescription(ctx context.Context, rawURL string) (domain.JobPosting, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return domain.JobPosting{}, atsutil.ErrInvalidURL
	}
	parts := splitPath(parsed.Path)
	if len(parts) < 2 {
		return domain.JobPosting{}, atsutil.ErrInvalidURL
	}
	company, jobKey := parts[0], parts[1]
	endpoint := fmt.Sprintf("https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true", url.PathEscape(company))
	var response struct {
		Jobs []struct {
			ID               string `json:"id"`
			Title            string `json:"title"`
			DescriptionPlain string `json:"descriptionPlain"`
			DescriptionHTML  string `json:"descriptionHtml"`
			JobURL           string `json:"jobUrl"`
		} `json:"jobs"`
	}
	if err := atsutil.GetJSON(ctx, s.client, endpoint, &response); err != nil {
		return domain.JobPosting{}, fmt.Errorf("ashby: %w", err)
	}
	for _, job := range response.Jobs {
		if job.ID != jobKey && !strings.HasSuffix(strings.TrimSuffix(job.JobURL, "/"), "/"+jobKey) {
			continue
		}
		description := job.DescriptionPlain
		if description == "" {
			description = atsutil.HTMLToText(job.DescriptionHTML)
		}
		return domain.JobPosting{Source: s.Name(), CompanySlug: company, JobID: job.ID, Title: job.Title, Company: atsutil.DisplayName(company), Description: description, URL: job.JobURL, FetchedAt: time.Now().UTC()}, nil
	}
	return domain.JobPosting{}, fmt.Errorf("ashby: job %q not found", jobKey)
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
