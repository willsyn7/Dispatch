package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/ats/ashby"
	"github.com/example/job-application-assistant/apps/api/internal/ats/greenhouse"
	"github.com/example/job-application-assistant/apps/api/internal/ats/lever"
	"github.com/example/job-application-assistant/apps/api/internal/ats/workday"
	"github.com/example/job-application-assistant/apps/api/internal/config"
	"github.com/example/job-application-assistant/apps/api/internal/httpapi"
	"github.com/example/job-application-assistant/apps/api/internal/service"
)

func main() {
	cfg := config.FromEnvironment()
	logger := log.New(os.Stdout, "api ", log.LstdFlags|log.LUTC)

	dispatcher := ats.NewDispatcher(
		greenhouse.New(),
		lever.New(),
		workday.New(),
		ashby.New(),
	)
	jobs := service.NewJobService(dispatcher)
	handler := httpapi.NewHandler(logger, jobs, cfg.WebOrigin)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	logger.Printf("starting API address=%s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Printf("API stopped error=%q", err)
		os.Exit(1)
	}
}
