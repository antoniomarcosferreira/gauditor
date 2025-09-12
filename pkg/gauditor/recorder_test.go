package gauditor

import (
	"context"
	"testing"
	"time"
)

func TestRecorder_RecordAssignsDefaultsAndValidates(t *testing.T) {
	rec := NewRecorder(NewMemoryStorage(), WithClock(func() time.Time { return time.Unix(1000, 0).UTC() }))
	ev, err := rec.Record(context.Background(), Event{
		Tenant: "acme",
		Action: "login",
		Actor:  Actor{ID: "u1", Attributes: map[string]any{"email": "a@example.com"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.ID == "" {
		t.Fatalf("expected id to be set")
	}
	if ev.Timestamp.IsZero() {
		t.Fatalf("expected timestamp to be set")
	}
	if ev.Timestamp != time.Unix(1000, 0).UTC() {
		t.Fatalf("timestamp not from clock")
	}
}

func TestRecorder_RecordInvalid(t *testing.T) {
	rec := NewRecorder(NewMemoryStorage())
	_, err := rec.Record(context.Background(), Event{Action: "x"})
	if err == nil {
		t.Fatalf("expected error for missing tenant")
	}
	_, err = rec.Record(context.Background(), Event{Tenant: "t"})
	if err == nil {
		t.Fatalf("expected error for missing action")
	}
}

func TestMemoryStorage_QueryFilters(t *testing.T) {
	store := NewMemoryStorage()
	rec := NewRecorder(store, WithClock(func() time.Time { return time.Unix(1000, 0).UTC() }))
	// seed
	_, _ = rec.Record(context.Background(), Event{Tenant: "a", Action: "x", Actor: Actor{ID: "u1"}})
	_, _ = rec.Record(context.Background(), Event{Tenant: "a", Action: "y", Actor: Actor{ID: "u2"}})
	_, _ = rec.Record(context.Background(), Event{Tenant: "b", Action: "x", Actor: Actor{ID: "u1"}})

	res, err := rec.Query(context.Background(), Query{Tenant: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("want 2, got %d", len(res))
	}

	res, _ = rec.Query(context.Background(), Query{Tenant: "a", Action: "x"})
	if len(res) != 1 {
		t.Fatalf("want 1, got %d", len(res))
	}

	res, _ = rec.Query(context.Background(), Query{ActorID: "u1"})
	if len(res) != 2 {
		t.Fatalf("want 2, got %d", len(res))
	}
}
