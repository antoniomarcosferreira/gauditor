package redisstore

import (
	"context"
	"encoding/json"

	gauditor "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
	"github.com/redis/go-redis/v9"
)

// Store implements gauditor.Storage backed by Redis.
//
// This implementation is intentionally simple and best for demos/development.
// Events are appended to a Redis list per-tenant and filtered in-application.
type Store struct {
	rdb       *redis.Client
	keyPrefix string
}

// Option configures the Store.
type Option func(*Store)

// WithKeyPrefix sets a prefix for Redis keys. Default: "gauditor:".
func WithKeyPrefix(prefix string) Option { return func(s *Store) { s.keyPrefix = prefix } }

// New constructs a Redis-backed Store.
func New(rdb *redis.Client, opts ...Option) *Store {
	s := &Store{rdb: rdb, keyPrefix: "gauditor:"}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *Store) keyForTenant(tenant string) string { return s.keyPrefix + tenant + ":events" }

// Save pushes the event to the tenant list (newest-first).
func (s *Store) Save(ctx context.Context, e gauditor.Event) (gauditor.Event, error) {
	raw, err := json.Marshal(e)
	if err != nil {
		return e, err
	}
	if err := s.rdb.LPush(ctx, s.keyForTenant(e.Tenant), raw).Err(); err != nil {
		return e, err
	}
	return e, nil
}

// Query scans the tenant list newest-to-oldest, returning ascending by timestamp.
func (s *Store) Query(ctx context.Context, q gauditor.Query) ([]gauditor.Event, error) {
	if q.Tenant == "" {
		return nil, nil
	}
	vals, err := s.rdb.LRange(ctx, s.keyForTenant(q.Tenant), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	results := make([]gauditor.Event, 0, len(vals))
	for i := len(vals) - 1; i >= 0; i-- {
		var e gauditor.Event
		if err := json.Unmarshal([]byte(vals[i]), &e); err != nil {
			continue
		}
		if q.ActorID != "" && e.Actor.ID != q.ActorID {
			continue
		}
		if q.Action != "" && e.Action != q.Action {
			continue
		}
		if q.TargetID != "" && e.Target.ID != q.TargetID {
			continue
		}
		if q.Since != nil && e.Timestamp.Before(*q.Since) {
			continue
		}
		if q.Until != nil && e.Timestamp.After(*q.Until) {
			continue
		}
		results = append(results, e)
		if q.Limit > 0 && len(results) >= q.Limit {
			break
		}
	}
	return results, nil
}
