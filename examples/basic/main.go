package main

import (
	"context"
	"fmt"
	"time"

	"github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
)

func main() {
	rec := gauditor.NewRecorder(gauditor.NewMemoryStorage())
	ez := gauditor.NewEasyRecorder(rec)

	e, err := ez.Record(context.Background(), "acme", map[string]any{"name": "Marcos"}, "update", map[string]any{"model": "users", "data": map[string]any{"dsds": 22}})
	if err != nil {
		panic(err)
	}

	until := time.Now().Add(1 * time.Hour)
	res, err := rec.Query(context.Background(), gauditor.Query{Tenant: "acme", Until: &until})
	if err != nil {
		panic(err)
	}

	fmt.Printf("stored event id=%s action=%s\n", e.ID, e.Action)
	fmt.Printf("queried %d events\n", len(res))
}
