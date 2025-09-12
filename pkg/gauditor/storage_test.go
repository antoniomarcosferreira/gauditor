package gauditor

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStorage_SinceUntilLimitAndSort(t *testing.T) {
	store := NewMemoryStorage()
	rec := NewRecorder(store)

	// Create three events with controlled timestamps
	t0 := time.Unix(1_000, 0).UTC()
	t1 := time.Unix(2_000, 0).UTC()
	t2 := time.Unix(3_000, 0).UTC()

	_, _ = rec.Record(context.Background(), Event{Tenant: "t", Action: "x", Actor: Actor{ID: "a"}, Timestamp: t0})
	_, _ = rec.Record(context.Background(), Event{Tenant: "t", Action: "y", Actor: Actor{ID: "b"}, Timestamp: t1})
	_, _ = rec.Record(context.Background(), Event{Tenant: "t", Action: "z", Actor: Actor{ID: "c"}, Timestamp: t2})

	// Since filters out earlier ones
	since := t1
	res, err := store.Query(context.Background(), Query{Tenant: "t", Since: &since})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("want 2, got %d", len(res))
	}
	if !res[0].Timestamp.Equal(t1) || !res[1].Timestamp.Equal(t2) {
		t.Fatalf("results not sorted asc by time")
	}

	// Until filters out later ones
	until := t1
	res, err = store.Query(context.Background(), Query{Tenant: "t", Until: &until})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("want 2, got %d", len(res))
	}
	if !res[0].Timestamp.Equal(t0) || !res[1].Timestamp.Equal(t1) {
		t.Fatalf("unexpected until results")
	}

	// Limit returns only first N after sort
	res, err = store.Query(context.Background(), Query{Tenant: "t", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || !res[0].Timestamp.Equal(t0) {
		t.Fatalf("limit did not apply")
	}
}

func TestMemoryStorage_TargetFilter(t *testing.T) {
	store := NewMemoryStorage()
	rec := NewRecorder(store)
	_, _ = rec.Record(context.Background(), Event{Tenant: "t", Action: "a", Target: Target{ID: "X"}})
	_, _ = rec.Record(context.Background(), Event{Tenant: "t", Action: "b", Target: Target{ID: "Y"}})
	res, err := store.Query(context.Background(), Query{Tenant: "t", TargetID: "Y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].Action != "b" {
		t.Fatalf("target filter failed")
	}
}
