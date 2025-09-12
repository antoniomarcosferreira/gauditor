package gauditor

import (
	"context"
	"testing"
)

func TestEasyRecorder_Record(t *testing.T) {
	rec := NewRecorder(NewMemoryStorage())
	ez := NewEasyRecorder(rec)
	ev, err := ez.Record(context.Background(), "tenant-x", map[string]any{"name": "Marcos"}, "update", map[string]any{"model": "users"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Tenant != "tenant-x" {
		t.Fatalf("wrong tenant: %s", ev.Tenant)
	}
	if ev.Action != "update" {
		t.Fatalf("wrong action: %s", ev.Action)
	}
	if ev.Actor.Attributes["name"] != "Marcos" {
		t.Fatalf("missing actor attribute")
	}
}
