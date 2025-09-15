package gauditor

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Recorder validates and persists events using a pluggable Storage backend.
//
// Recorder is safe for concurrent use by multiple goroutines.
// The zero value is not usable; use NewRecorder to construct one.
type Recorder struct {
	store Storage
	clock func() time.Time
	idgen func() string
}

// Option configures a Recorder instance created via NewRecorder.
// Options are applied in the order provided.
type Option func(*Recorder)

// WithClock overrides the time source used to populate Event.Timestamp.
// The provided function should return a UTC or local time; Recorder will call UTC().
func WithClock(clock func() time.Time) Option { return func(r *Recorder) { r.clock = clock } }

// WithIDGenerator overrides the ID generator used for new events.
// Defaults to github.com/google/uuid.NewString.
func WithIDGenerator(gen func() string) Option { return func(r *Recorder) { r.idgen = gen } }

// NewRecorder constructs a Recorder with the provided Storage and Optional settings.
// By default, it uses time.Now for the clock and uuid.NewString for ID generation.
func NewRecorder(store Storage, opts ...Option) *Recorder {
	r := &Recorder{store: store, clock: time.Now, idgen: uuid.NewString}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Record assigns defaults (ID, Timestamp), validates required fields, and persists the event.
// It returns the stored Event, which may include defaults applied by the storage backend.
func (r *Recorder) Record(ctx context.Context, e Event) (Event, error) {
	if e.ID == "" {
		e.ID = r.idgen()
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

// Query retrieves events from the underlying Storage that match the provided filter.
// The ordering and pagination are storage-defined; MemoryStorage returns ascending by timestamp.
func (r *Recorder) Query(ctx context.Context, q Query) ([]Event, error) {
	return r.store.Query(ctx, q)
}
