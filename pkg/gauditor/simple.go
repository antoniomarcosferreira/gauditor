package gauditor

import "context"

// SimpleRecorder defines a minimal API for recording events without constructing
// a full Event. It is useful for call-sites that only need to provide a tenant,
// actor attributes, an action, and optional data payload.
//
// Example:
//
//	ez.Record(ctx, "123434", map[string]any{"name": "Marcos"}, "update", map[string]any{"model": "users"})
type SimpleRecorder interface {
	Record(ctx context.Context, tenant string, actorAttributes map[string]any, action string, data map[string]any) (Event, error)
}

// EasyRecorder is a small facade that implements SimpleRecorder on top of Recorder.
type EasyRecorder struct {
	recorder *Recorder
}

// NewEasyRecorder constructs an EasyRecorder using an existing Recorder.
func NewEasyRecorder(recorder *Recorder) *EasyRecorder {
	return &EasyRecorder{recorder: recorder}
}

// Record builds an Event from simple parameters and delegates to Recorder.Record.
// It sets the Actor.Attributes map as provided.
func (e *EasyRecorder) Record(ctx context.Context, tenant string, actorAttributes map[string]any, action string, data map[string]any) (Event, error) {
	return e.recorder.Record(ctx, Event{
		Tenant: tenant,
		Actor:  Actor{Attributes: actorAttributes},
		Action: action,
		Data:   data,
	})
}
