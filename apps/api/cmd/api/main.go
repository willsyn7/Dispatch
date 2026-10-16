package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/example/job-application-assistant/apps/api/internal/ats"
	"github.com/example/job-application-assistant/apps/api/internal/ats/ashby"
	"github.com/example/job-application-assistant/apps/api/internal/ats/greenhouse"
	"github.com/example/job-application-assistant/apps/api/internal/ats/lever"
	"github.com/example/job-application-assistant/apps/api/internal/ats/workday"
	"github.com/example/job-application-assistant/apps/api/internal/auth"
	"github.com/example/job-application-assistant/apps/api/internal/cache"
	"github.com/example/job-application-assistant/apps/api/internal/config"
	"github.com/example/job-application-assistant/apps/api/internal/drafting"
	"github.com/example/job-application-assistant/apps/api/internal/httpapi"
	"github.com/example/job-application-assistant/apps/api/internal/repository"
	cosmosrepo "github.com/example/job-application-assistant/apps/api/internal/repository/cosmos"
	memoryrepo "github.com/example/job-application-assistant/apps/api/internal/repository/memory"
	"github.com/example/job-application-assistant/apps/api/internal/service"
)

func main() {
	cfg := config.FromEnvironment()
	logger := log.New(os.Stdout, "api ", log.LstdFlags|log.LUTC)
	httpClient := &http.Client{Timeout: 20 * time.Second}

	cacheStore, closeCache := configureCache(cfg, logger)
	defer closeCache()
	credential := configureCredential(cfg, logger)
	repo := configureRepository(cfg, credential, logger)
	generator := configureGenerator(cfg, credential, httpClient, logger)

	dispatcher := ats.NewDispatcher(
		greenhouse.New(httpClient),
		lever.New(httpClient),
		workday.New(httpClient),
		ashby.New(httpClient),
	)
	jobs := service.NewJobService(dispatcher, cacheStore)
	applications := service.NewApplicationService(cfg.UserID, repo, generator, cacheStore)
	handler := httpapi.NewHandler(logger, jobs, cfg.WebOrigin, applications)
	handler = auth.New(cfg.RequireAuth, cfg.AllowedPrincipal).Wrap(handler)

	server := &http.Server{Addr: cfg.HTTPAddr, Handler: handler, ReadHeaderTimeout: cfg.ReadHeaderTimeout, ReadTimeout: 30 * time.Second, WriteTimeout: 45 * time.Second, IdleTimeout: 60 * time.Second}
	logger.Printf("starting API address=%s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Printf("API stopped error=%q", err)
		os.Exit(1)
	}
}

func configureCache(cfg config.Config, logger *log.Logger) (cache.Store, func()) {
	if cfg.RedisURL == "" {
		logger.Print("Redis is not configured; using no-op cache")
		return cache.Noop{}, func() {}
	}
	store, err := cache.NewRedis(cfg.RedisURL)
	if err != nil {
		logger.Fatalf("configure Redis: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		_ = store.Close()
		logger.Printf("Redis is unavailable; continuing with no-op cache error=%q", err)
		return cache.Noop{}, func() {}
	}
	return store, func() { _ = store.Close() }
}

func configureCredential(cfg config.Config, logger *log.Logger) azcore.TokenCredential {
	needsCredential := (cfg.CosmosEndpoint != "" && cfg.CosmosConnection == "") || (cfg.AIEndpoint != "" && cfg.AIAPIKey == "")
	if !needsCredential {
		return nil
	}
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		logger.Fatalf("configure managed identity: %v", err)
	}
	return credential
}

func configureRepository(cfg config.Config, credential azcore.TokenCredential, logger *log.Logger) repository.Repository {
	if cfg.CosmosEndpoint == "" && cfg.CosmosConnection == "" {
		logger.Print("Cosmos DB is not configured; using in-memory persistence")
		return memoryrepo.New()
	}
	var client *azcosmos.Client
	var err error
	if cfg.CosmosConnection != "" {
		client, err = azcosmos.NewClientFromConnectionString(cfg.CosmosConnection, nil)
	} else {
		client, err = azcosmos.NewClient(cfg.CosmosEndpoint, credential, nil)
	}
	if err != nil {
		logger.Fatalf("configure Cosmos DB: %v", err)
	}
	repo, err := cosmosrepo.New(client, cfg.CosmosDatabase)
	if err != nil {
		logger.Fatalf("configure Cosmos containers: %v", err)
	}
	return repo
}

func configureGenerator(cfg config.Config, credential azcore.TokenCredential, client *http.Client, logger *log.Logger) drafting.Generator {
	if cfg.AIEndpoint == "" || cfg.AIDeployment == "" {
		logger.Print("Azure AI Foundry is not configured; draft generation is disabled")
		return nil
	}
	var tokenProvider drafting.TokenProvider
	if cfg.AIAPIKey == "" {
		tokenProvider = func(ctx context.Context) (string, error) {
			token, err := credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"https://cognitiveservices.azure.com/.default"}})
			return token.Token, err
		}
	}
	generator, err := drafting.NewFoundry(cfg.AIEndpoint, cfg.AIDeployment, cfg.AIAPIKey, tokenProvider, client)
	if err != nil {
		logger.Fatalf("configure Azure AI Foundry: %v", err)
	}
	return generator
}
