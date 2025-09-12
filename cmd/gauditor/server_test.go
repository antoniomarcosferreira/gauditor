package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	g "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
)

func TestHTTP_IngestAndQuery(t *testing.T) {
	rec := g.NewRecorder(g.NewMemoryStorage())
	srv := httptest.NewServer(newServer(rec))
	t.Cleanup(srv.Close)

	// POST an event
	payload := map[string]any{
		"tenant": "acme",
		"actor": map[string]any{
			"id":         "u1",
			"attributes": map[string]any{"type": "user"},
		},
		"action": "login",
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(srv.URL+"/v1/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("want 201, got %d", resp.StatusCode)
	}

	// GET query
	r2, err := http.Get(srv.URL + "/v1/events?tenant=acme")
	if err != nil {
		t.Fatal(err)
	}
	defer r2.Body.Close()
	raw, _ := io.ReadAll(r2.Body)
	if !bytes.Contains(raw, []byte("\"action\":\"login\"")) {
		t.Fatalf("response missing event: %s", string(raw))
	}
}

func TestHTTP_InvalidJSON(t *testing.T) {
	rec := g.NewRecorder(g.NewMemoryStorage())
	srv := httptest.NewServer(newServer(rec))
	t.Cleanup(srv.Close)

	resp, err := http.Post(srv.URL+"/v1/events", "application/json", bytes.NewBufferString("{"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
}

func TestHTTP_MethodNotAllowed(t *testing.T) {
	rec := g.NewRecorder(g.NewMemoryStorage())
	srv := httptest.NewServer(newServer(rec))
	t.Cleanup(srv.Close)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/v1/events", nil)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", r.StatusCode)
	}
}

func TestHTTP_MissingRequiredFields(t *testing.T) {
	rec := g.NewRecorder(g.NewMemoryStorage())
	srv := httptest.NewServer(newServer(rec))
	t.Cleanup(srv.Close)

	// Missing tenant
	payload := map[string]any{
		"action": "login",
		"actor":  map[string]any{"id": "u1"},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(srv.URL+"/v1/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
}

func TestHTTP_QueryFilters(t *testing.T) {
	rec := g.NewRecorder(g.NewMemoryStorage())
	// seed directly
	_, _ = rec.Record(context.Background(), g.Event{Tenant: "t1", Action: "x", Actor: g.Actor{ID: "a"}})
	_, _ = rec.Record(context.Background(), g.Event{Tenant: "t1", Action: "y", Actor: g.Actor{ID: "b"}})
	_, _ = rec.Record(context.Background(), g.Event{Tenant: "t2", Action: "x", Actor: g.Actor{ID: "a"}})

	srv := httptest.NewServer(newServer(rec))
	t.Cleanup(srv.Close)

	r, err := http.Get(srv.URL + "/v1/events?tenant=t1&action=x")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	raw, _ := io.ReadAll(r.Body)
	if !bytes.Contains(raw, []byte("\"action\":\"x\"")) || bytes.Contains(raw, []byte("\"action\":\"y\"")) {
		t.Fatalf("unexpected filter results: %s", string(raw))
	}
}

func TestRun_InvalidAddr(t *testing.T) {
	rec := g.NewRecorder(g.NewMemoryStorage())
	h := newServer(rec)
	if err := run("bad-addr", h); err == nil {
		t.Fatalf("expected error for invalid addr")
	}
}

func TestRealMain_NoServe(t *testing.T) {
	t.Setenv("GAUDITOR_NO_SERVE", "1")
	// Call realMain to cover init path without serving
	if code := realMain(); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// errorStorage implements gauditor.Storage returning error on Query to hit 500 path.
type errorStorage struct{}

func (e errorStorage) Save(_ context.Context, ev g.Event) (g.Event, error) { return ev, nil }
func (e errorStorage) Query(_ context.Context, _ g.Query) ([]g.Event, error) {
	return nil, errors.New("boom")
}

func TestHTTP_InternalErrorOnQuery(t *testing.T) {
	rec := g.NewRecorder(errorStorage{})
	srv := httptest.NewServer(newServer(rec))
	t.Cleanup(srv.Close)
	r, err := http.Get(srv.URL + "/v1/events?tenant=any")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", r.StatusCode)
	}
}

func TestRealMain_NoServeEnv(t *testing.T) {
	t.Setenv("GAUDITOR_NO_SERVE", "1")
	// We cannot call main() because it calls os.Exit, but we can ensure that
	// initialization path is reachable by creating server and avoiding listen.
	rec := g.NewRecorder(g.NewMemoryStorage())
	if h := newServer(rec); h == nil {
		t.Fatal("nil handler")
	}
}
