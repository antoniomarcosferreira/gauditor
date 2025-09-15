package gauditor

import (
	"context"
	"sort"
	"sync"
)

// Storage defines persistence for events.
type Storage interface {
	Save(ctx context.Context, event Event) (Event, error)
	Query(ctx context.Context, query Query) ([]Event, error)
}

// MemoryStorage is an in-memory implementation of Storage for development/testing.
type MemoryStorage struct {
	mu     sync.RWMutex
	events []Event
}

// NewMemoryStorage constructs a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{events: make([]Event, 0, 1024)}
}

// Save appends an event.
func (m *MemoryStorage) Save(ctx context.Context, event Event) (Event, error) {
	select {
	case <-ctx.Done():
		return event, ctx.Err()
	default:
	}
	m.mu.Lock()
	m.events = append(m.events, event)
	m.mu.Unlock()
	return event, nil
}

// Query returns events matching the filter. Results are sorted by timestamp ascending.
func (m *MemoryStorage) Query(ctx context.Context, q Query) ([]Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]Event, 0, len(m.events))
	for _, e := range m.events {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if q.Tenant != "" && e.Tenant != q.Tenant {
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
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Timestamp.Before(results[j].Timestamp) })

	if q.Limit > 0 && len(results) > q.Limit {
		results = results[:q.Limit]
	}
	return results, nil
}
