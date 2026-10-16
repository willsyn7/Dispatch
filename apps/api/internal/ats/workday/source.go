package workday

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
func (*Source) Name() string { return "workday" }
func (*Source) Detect(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "myworkdayjobs.com" || strings.HasSuffix(host, ".myworkdayjobs.com")
}

func (s *Source) GetJobDescription(ctx context.Context, rawURL string) (domain.JobPosting, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return domain.JobPosting{}, atsutil.ErrInvalidURL
	}
	parts := splitPath(parsed.Path)
	jobIndex := -1
	for index, part := range parts {
		if part == "job" {
			jobIndex = index
			break
		}
	}
	if jobIndex < 1 || jobIndex+1 >= len(parts) {
		return domain.JobPosting{}, atsutil.ErrInvalidURL
	}
	tenant := strings.Split(parsed.Hostname(), ".")[0]
	site := parts[jobIndex-1]
	jobParts := parts[jobIndex+1:]
	for index, part := range jobParts {
		jobParts[index] = url.PathEscape(part)
	}
	jobPath := strings.Join(jobParts, "/")
	endpoint := fmt.Sprintf("%s://%s/wday/cxs/%s/%s/job/%s", parsed.Scheme, parsed.Host, url.PathEscape(tenant), url.PathEscape(site), jobPath)
	var response struct {
		JobPostingInfo struct {
			ID             string `json:"id"`
			JobReqID       string `json:"jobReqId"`
			Title          string `json:"title"`
			JobDescription string `json:"jobDescription"`
			ExternalURL    string `json:"externalUrl"`
		} `json:"jobPostingInfo"`
	}
	if err := atsutil.GetJSON(ctx, s.client, endpoint, &response); err != nil {
		return domain.JobPosting{}, fmt.Errorf("workday: %w", err)
	}
	info := response.JobPostingInfo
	jobID := info.JobReqID
	if jobID == "" {
		jobID = info.ID
	}
	return domain.JobPosting{Source: s.Name(), CompanySlug: tenant, JobID: jobID, Title: info.Title, Company: atsutil.DisplayName(tenant), Description: atsutil.HTMLToText(info.JobDescription), URL: info.ExternalURL, FetchedAt: time.Now().UTC()}, nil
}

func splitPath(path string) []string {
	var result []string
	for _, part := range strings.Split(strings.Trim(path, "/"), "/") {
		if part != "" {
			decoded, _ := url.PathUnescape(part)
			result = append(result, decoded)
		}
	}
	return result
}
