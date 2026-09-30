package ats

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/example/job-application-assistant/apps/api/internal/domain"
)

type StubSource struct {
	SourceName string
	Hosts      []string
}

func (s StubSource) Name() string { return s.SourceName }

func (s StubSource) Detect(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	for _, suffix := range s.Hosts {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}

func (s StubSource) GetJobDescription(context.Context, string) (domain.JobPosting, error) {
	return domain.JobPosting{}, fmt.Errorf("%s adapter: %w", s.SourceName, ErrNotImplemented)
}

var ErrNotImplemented = errors.New("adapter implementation is not part of the initial scaffold")
