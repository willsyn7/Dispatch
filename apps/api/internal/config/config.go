package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr          string
	WebOrigin         string
	ReadHeaderTimeout time.Duration
	UserID            string
	RequireAuth       bool
	AllowedPrincipal  string
	RedisURL          string
	CosmosEndpoint    string
	CosmosConnection  string
	CosmosDatabase    string
	AIEndpoint        string
	AIDeployment      string
	AIAPIKey          string
}

func FromEnvironment() Config {
	return Config{
		HTTPAddr:          envOrDefault("HTTP_ADDR", ":8080"),
		WebOrigin:         envOrDefault("WEB_ORIGIN", "http://localhost:4200"),
		ReadHeaderTimeout: 5 * time.Second,
		UserID:            envOrDefault("APP_USER_ID", "owner"),
		RequireAuth:       envBool("REQUIRE_ENTRA_AUTH", false),
		AllowedPrincipal:  os.Getenv("ALLOWED_ENTRA_PRINCIPAL_ID"),
		RedisURL:          os.Getenv("REDIS_URL"),
		CosmosEndpoint:    os.Getenv("AZURE_COSMOS_ENDPOINT"),
		CosmosConnection:  os.Getenv("AZURE_COSMOS_CONNECTION_STRING"),
		CosmosDatabase:    envOrDefault("COSMOS_DATABASE", "job-application-assistant"),
		AIEndpoint:        os.Getenv("AZURE_AI_ENDPOINT"),
		AIDeployment:      os.Getenv("AZURE_AI_DEPLOYMENT"),
		AIAPIKey:          os.Getenv("AZURE_AI_API_KEY"),
	}
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
