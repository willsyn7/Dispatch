package ats_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/ats/ashby"
	"github.com/example/job-application-assistant/apps/api/internal/ats/greenhouse"
	"github.com/example/job-application-assistant/apps/api/internal/ats/lever"
	"github.com/example/job-application-assistant/apps/api/internal/ats/workday"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestAdaptersResolvePublicJobURLs(t *testing.T) {
	tests := []struct {
		name        string
		source      func(*http.Client) ats.JobSource
		jobURL      string
		response    string
		requestPath string
		wantSource  string
		wantID      string
		wantTitle   string
	}{
		{name: "greenhouse", source: func(c *http.Client) ats.JobSource { return greenhouse.New(c) }, jobURL: "https://job-boards.greenhouse.io/acme/jobs/123", requestPath: "/v1/boards/acme/jobs/123", response: `{"id":123,"title":"Platform Engineer","content":"<p>Build systems.</p>","absolute_url":"https://job-boards.greenhouse.io/acme/jobs/123","company_name":"Acme"}`, wantSource: "greenhouse", wantID: "123", wantTitle: "Platform Engineer"},
		{name: "lever", source: func(c *http.Client) ats.JobSource { return lever.New(c) }, jobURL: "https://jobs.lever.co/acme/abc-123", requestPath: "/v0/postings/acme/abc-123", response: `{"id":"abc-123","text":"Backend Engineer","descriptionPlain":"Build APIs.","hostedUrl":"https://jobs.lever.co/acme/abc-123","lists":[{"text":"What you'll do","content":"<li>Ship code</li>"}]}`, wantSource: "lever", wantID: "abc-123", wantTitle: "Backend Engineer"},
		{name: "workday", source: func(c *http.Client) ats.JobSource { return workday.New(c) }, jobURL: "https://acme.wd5.myworkdayjobs.com/en-US/careers/job/Seattle/Engineer_R123", requestPath: "/wday/cxs/acme/careers/job/Seattle/Engineer_R123", response: `{"jobPostingInfo":{"jobReqId":"R123","title":"Systems Engineer","jobDescription":"<p>Build platforms.</p>","externalUrl":"https://acme.wd5.myworkdayjobs.com/en-US/careers/job/Seattle/Engineer_R123"}}`, wantSource: "workday", wantID: "R123", wantTitle: "Systems Engineer"},
		{name: "ashby", source: func(c *http.Client) ats.JobSource { return ashby.New(c) }, jobURL: "https://jobs.ashbyhq.com/acme/job-123", requestPath: "/posting-api/job-board/acme", response: `{"jobs":[{"id":"job-123","title":"Product Engineer","descriptionHtml":"<p>Build products.</p>","jobUrl":"https://jobs.ashbyhq.com/acme/job-123"}]}`, wantSource: "ashby", wantID: "job-123", wantTitle: "Product Engineer"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.URL.Path != test.requestPath {
					t.Fatalf("unexpected request path %q", request.URL.Path)
				}
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(test.response))}, nil
			})}
			posting, err := test.source(client).GetJobDescription(context.Background(), test.jobURL)
			if err != nil {
				t.Fatalf("resolve posting: %v", err)
			}
			if posting.Source != test.wantSource || posting.JobID != test.wantID || posting.Title != test.wantTitle {
				t.Fatalf("unexpected posting: %#v", posting)
			}
			if strings.Contains(posting.Description, "<") {
				t.Fatalf("description was not normalized: %q", posting.Description)
			}
		})
	}
}
