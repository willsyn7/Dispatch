package ashby

import "github.com/example/job-application-assistant/apps/api/internal/ats"

func New() ats.JobSource {
	return ats.StubSource{SourceName: "ashby", Hosts: []string{"ashbyhq.com"}}
}
