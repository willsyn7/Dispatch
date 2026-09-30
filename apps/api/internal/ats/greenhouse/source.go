package greenhouse

import "github.com/example/job-application-assistant/apps/api/internal/ats"

func New() ats.JobSource {
	return ats.StubSource{SourceName: "greenhouse", Hosts: []string{"greenhouse.io", "greenhouse.com"}}
}
