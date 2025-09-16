package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	g "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
	"github.com/gin-gonic/gin"
)

// buildTestServer replicates the example setup and returns the recorder and router.
func buildTestServer() (*g.Recorder, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	recorder := g.NewRecorder(g.NewMemoryStorage())
	r := gin.New()
	r.Use(gin.Recovery())

	actorFrom := func(c *gin.Context) g.Actor {
		id := c.GetHeader("X-User-ID")
		if id == "" {
			id = "anonymous"
		}
		return g.Actor{ID: id}
	}

	auditMiddleware := func(rec *g.Recorder, actor func(*gin.Context) g.Actor) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Next()
			status := c.Writer.Status()
			if status >= 200 && status < 400 {
				action := c.Request.Method + " " + c.FullPath()
				target := g.Target{ID: c.Param("id")}
				_, _ = rec.Record(c, g.Event{Tenant: "acme", Actor: actor(c), Action: action, Target: target})
			}
		}
	}

	st := newStore()
	r.Use(auditMiddleware(recorder, actorFrom))

	r.POST("/users", func(c *gin.Context) {
		var in User
		if err := c.ShouldBindJSON(&in); err != nil || in.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body (id,name,email required)"})
			return
		}
		out := st.create(in)
		c.JSON(http.StatusCreated, out)
	})

	r.PUT("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		var in User
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if out, ok := st.update(id, in); ok {
			c.JSON(http.StatusOK, out)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	r.DELETE("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if st.delete(id) {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	return recorder, r
}

func TestGinCRUD_AuditEventsRecorded(t *testing.T) {
	rec, router := buildTestServer()
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	// Create
	body, _ := json.Marshal(map[string]any{"id": "1", "name": "Alice", "email": "a@ex.com"})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "u1")
	r1, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	r1.Body.Close()
	if r1.StatusCode != http.StatusCreated {
		t.Fatalf("create got %d", r1.StatusCode)
	}

	// Update
	body2, _ := json.Marshal(map[string]any{"name": "Alice Doe", "email": "b@ex.com"})
	req2, _ := http.NewRequest(http.MethodPut, srv.URL+"/users/1", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-User-ID", "u1")
	r2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	r2.Body.Close()
	if r2.StatusCode != http.StatusOK {
		t.Fatalf("update got %d", r2.StatusCode)
	}

	// Delete
	req3, _ := http.NewRequest(http.MethodDelete, srv.URL+"/users/1", nil)
	req3.Header.Set("X-User-ID", "u1")
	r3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatal(err)
	}
	r3.Body.Close()
	if r3.StatusCode != http.StatusNoContent {
		t.Fatalf("delete got %d", r3.StatusCode)
	}

	// Assert audit events
	events, err := rec.Query(context.Background(), g.Query{Tenant: "acme"})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	wantActions := map[string]bool{
		"POST /users":       false,
		"PUT /users/:id":    false,
		"DELETE /users/:id": false,
	}
	for _, e := range events {
		if _, ok := wantActions[e.Action]; ok {
			wantActions[e.Action] = true
		}
	}
	for a, seen := range wantActions {
		if !seen {
			t.Fatalf("missing action %s in audit events", a)
		}
	}
}
