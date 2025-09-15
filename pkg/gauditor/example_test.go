package gauditor_test

import (
	"context"
	"fmt"
	"time"

	g "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
)

// ExampleRecorder demonstrates recording and querying events.
func ExampleRecorder() {
	rec := g.NewRecorder(
		g.NewMemoryStorage(),
		g.WithClock(func() time.Time { return time.Unix(0, 0).UTC() }),
		g.WithIDGenerator(func() string { return "id-1" }),
	)

	// Record a login event
	_, err := rec.Record(context.Background(), g.Event{
		Tenant: "acme",
		Actor:  g.Actor{ID: "u1"},
		Action: "login",
	})
	if err != nil {
		panic(err)
	}

	// Query back
	list, err := rec.Query(context.Background(), g.Query{Tenant: "acme", Limit: 10})
	if err != nil {
		panic(err)
	}

	fmt.Printf("stored first id=%s action=%s count=%d\n", list[0].ID, list[0].Action, len(list))
	// Output:
	// stored first id=id-1 action=login count=1
}

// ExampleEasyRecorder shows the convenience facade.
func ExampleEasyRecorder() {
	rec := g.NewRecorder(g.NewMemoryStorage(), g.WithIDGenerator(func() string { return "id-42" }))
	ez := g.NewEasyRecorder(rec)
	e, err := ez.Record(context.Background(), "acme", map[string]any{"name": "Alice"}, "update", map[string]any{"model": "users"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("action=%s tenant=%s id=%s\n", e.Action, e.Tenant, e.ID)
	// Output:
	// action=update tenant=acme id=id-42
}
