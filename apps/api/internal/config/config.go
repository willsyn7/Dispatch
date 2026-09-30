package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr          string
	WebOrigin         string
	ReadHeaderTimeout time.Duration
}

func FromEnvironment() Config {
	return Config{
		HTTPAddr:          envOrDefault("HTTP_ADDR", ":8080"),
		WebOrigin:         envOrDefault("WEB_ORIGIN", "http://localhost:4200"),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
