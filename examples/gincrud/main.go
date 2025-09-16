package main

import (
	"net/http"
	"sync"

	g "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
	"github.com/gin-gonic/gin"
)

// User is a simple demo model.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// store is an in-memory CRUD storage for users.
type store struct {
	mu    sync.RWMutex
	items map[string]User
}

func newStore() *store { return &store{items: make(map[string]User)} }

func (s *store) create(u User) User {
	s.mu.Lock()
	s.items[u.ID] = u
	s.mu.Unlock()
	return u
}

func (s *store) get(id string) (User, bool) {
	s.mu.RLock()
	u, ok := s.items[id]
	s.mu.RUnlock()
	return u, ok
}

func (s *store) list() []User {
	s.mu.RLock()
	out := make([]User, 0, len(s.items))
	for _, u := range s.items {
		out = append(out, u)
	}
	s.mu.RUnlock()
	return out
}

func (s *store) update(id string, u User) (User, bool) {
	s.mu.Lock()
	if _, ok := s.items[id]; !ok {
		s.mu.Unlock()
		return User{}, false
	}
	u.ID = id
	s.items[id] = u
	s.mu.Unlock()
	return u, true
}

func (s *store) delete(id string) bool {
	s.mu.Lock()
	if _, ok := s.items[id]; !ok {
		s.mu.Unlock()
		return false
	}
	delete(s.items, id)
	s.mu.Unlock()
	return true
}

func main() {
	// Initialize gauditor recorder and gin engine
	recorder := g.NewRecorder(g.NewMemoryStorage())
	r := gin.Default()

	// Simple header-based actor for demo (X-User-ID)
	actorFrom := func(c *gin.Context) g.Actor {
		id := c.GetHeader("X-User-ID")
		if id == "" {
			id = "anonymous"
		}
		return g.Actor{ID: id}
	}

	// auditMiddleware auto-records successful requests as audit events.
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

	// Install middleware to capture audit events automatically
	r.Use(auditMiddleware(recorder, actorFrom))

	// Create
	r.POST("/users", func(c *gin.Context) {
		var in User
		if err := c.ShouldBindJSON(&in); err != nil || in.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body (id,name,email required)"})
			return
		}
		out := st.create(in)
		c.JSON(http.StatusCreated, out)
	})

	// Get by ID
	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if u, ok := st.get(id); ok {
			c.JSON(http.StatusOK, u)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	// List
	r.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, st.list())
	})

	// Update
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

	// Delete
	r.DELETE("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if st.delete(id) {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	// Run sample server (different port to avoid clashing with main server)
	_ = r.Run(":8092")
}
