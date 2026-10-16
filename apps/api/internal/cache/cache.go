package cache

import (
	"context"
	"encoding/json"
	"time"
)

type Store interface {
	Get(ctx context.Context, key string, destination any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

type Noop struct{}

func (Noop) Get(context.Context, string, any) (bool, error)        { return false, nil }
func (Noop) Set(context.Context, string, any, time.Duration) error { return nil }

func Marshal(value any) ([]byte, error)             { return json.Marshal(value) }
func Unmarshal(value []byte, destination any) error { return json.Unmarshal(value, destination) }
