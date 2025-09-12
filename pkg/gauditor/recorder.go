package gauditor

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Recorder validates and persists events.
type Recorder struct {
	store Storage
	clock func() time.Time
}

// Option configures a Recorder.
type Option func(*Recorder)

// WithClock overrides time source.
func WithClock(clock func() time.Time) Option { return func(r *Recorder) { r.clock = clock } }

// NewRecorder constructs a Recorder with the provided Storage.
func NewRecorder(store Storage, opts ...Option) *Recorder {
	r := &Recorder{store: store, clock: time.Now}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Record assigns defaults, performs basic validation, and saves the event.
func (r *Recorder) Record(ctx context.Context, e Event) (Event, error) {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = r.clock().UTC()
	}
	// Minimal validation
	if e.Tenant == "" || e.Action == "" {
		return e, ErrInvalidEvent
	}
	return r.store.Save(ctx, e)
}

// Query delegates to the storage.
func (r *Recorder) Query(ctx context.Context, q Query) ([]Event, error) {
	return r.store.Query(ctx, q)
}
