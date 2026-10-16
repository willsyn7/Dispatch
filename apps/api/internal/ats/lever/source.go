package lever

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
func (*Source) Name() string { return "lever" }
func (*Source) Detect(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "lever.co" || strings.HasSuffix(host, ".lever.co")
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
	company, jobID := parts[0], parts[1]
	endpoint := fmt.Sprintf("https://api.lever.co/v0/postings/%s/%s?mode=json", url.PathEscape(company), url.PathEscape(jobID))
	var response struct {
		ID               string `json:"id"`
		Text             string `json:"text"`
		DescriptionPlain string `json:"descriptionPlain"`
		HostedURL        string `json:"hostedUrl"`
		Categories       struct {
			Team string `json:"team"`
		} `json:"categories"`
		Lists []struct {
			Text    string `json:"text"`
			Content string `json:"content"`
		} `json:"lists"`
	}
	if err := atsutil.GetJSON(ctx, s.client, endpoint, &response); err != nil {
		return domain.JobPosting{}, fmt.Errorf("lever: %w", err)
	}
	description := response.DescriptionPlain
	for _, section := range response.Lists {
		description += "\n\n" + section.Text + "\n" + atsutil.HTMLToText(section.Content)
	}
	return domain.JobPosting{Source: s.Name(), CompanySlug: company, JobID: response.ID, Title: response.Text, Company: atsutil.DisplayName(company), Description: strings.TrimSpace(description), URL: response.HostedURL, FetchedAt: time.Now().UTC()}, nil
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
