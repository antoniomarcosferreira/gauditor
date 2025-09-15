// Package gauditor provides a small, idiomatic toolkit for recording and querying
// audit events in Go services. It focuses on clarity, testability, and minimal
// dependencies. The core concepts are:
//
//   - Event: a normalized audit record containing actor, action, target, and data
//   - Recorder: validates and persists events to a Storage implementation
//   - Storage: a pluggable interface for persistence (in-memory provided)
//   - EasyRecorder: a convenience facade for quick usage
//
// Typical usage:
//
//	rec := gauditor.NewRecorder(gauditor.NewMemoryStorage())
//	event, err := rec.Record(ctx, gauditor.Event{
//		Tenant: "acme",
//		Actor:  gauditor.Actor{ID: "u1"},
//		Action: "login",
//	})
//	if err != nil {
//		// handle validation/persistence error
//	}
//
// Storage is intentionally minimal; you can implement persistent backends
// (for example, PostgreSQL or cloud logging) by satisfying the Storage interface.
//
// All operations are context-aware to support cancellation and timeouts in callers.
package gauditor
