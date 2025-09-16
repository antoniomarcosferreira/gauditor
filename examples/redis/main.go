package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	g "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
	"github.com/redis/go-redis/v9"
)

// redisStorage is a simple Redis-based Storage implementation using a list per tenant.
// This is a demo example: not optimized for production and no pagination.
type redisStorage struct {
	rdb *redis.Client
}

func newRedisStorage(addr string) *redisStorage {
	return &redisStorage{rdb: redis.NewClient(&redis.Options{Addr: addr})}
}

func (s *redisStorage) keyForTenant(tenant string) string { return "gauditor:" + tenant + ":events" }

func (s *redisStorage) Save(ctx context.Context, e g.Event) (g.Event, error) {
	raw, err := json.Marshal(e)
	if err != nil {
		return e, err
	}
	if err := s.rdb.LPush(ctx, s.keyForTenant(e.Tenant), raw).Err(); err != nil {
		return e, err
	}
	return e, nil
}

func (s *redisStorage) Query(ctx context.Context, q g.Query) ([]g.Event, error) {
	if q.Tenant == "" {
		return nil, nil
	}
	vals, err := s.rdb.LRange(ctx, s.keyForTenant(q.Tenant), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	out := make([]g.Event, 0, len(vals))
	// Stored newest-first, so reverse to ascending by timestamp
	for i := len(vals) - 1; i >= 0; i-- {
		var e g.Event
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
		out = append(out, e)
		if q.Limit > 0 && len(out) >= q.Limit {
			break
		}
	}
	return out, nil
}

func main() {
	ctx := context.Background()
	store := newRedisStorage("127.0.0.1:6379")
	if err := store.rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis not available: %v", err)
	}
	rec := g.NewRecorder(store)

	// Write an event
	e, err := rec.Record(ctx, g.Event{Tenant: "acme", Actor: g.Actor{ID: "u1"}, Action: "login"})
	if err != nil {
		log.Fatal(err)
	}

	// Query events
	until := time.Now().Add(1 * time.Hour)
	list, err := rec.Query(ctx, g.Query{Tenant: "acme", Until: &until, Limit: 10})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("stored id=%s action=%s count=%d\n", e.ID, e.Action, len(list))
}
