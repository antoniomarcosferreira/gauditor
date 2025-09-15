package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
)

var exitFunc = os.Exit

// newServer returns an http.Handler with routes configured for the recorder.
//
// Routes:
//
//	POST /v1/events  - ingest an event (JSON body of gauditor.Event)
//	GET  /v1/events  - query events with optional filters tenant, actorId, action, targetId, limit
func newServer(recorder *gauditor.Recorder) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Limit request body and decode strictly
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
			dec := json.NewDecoder(r.Body)
			dec.DisallowUnknownFields()
			var e gauditor.Event
			if err := dec.Decode(&e); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("invalid JSON"))
				return
			}
			// Ensure no trailing tokens
			if err := dec.Decode(new(struct{})); err != io.EOF {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("invalid JSON: trailing content"))
				return
			}
			out, err := recorder.Record(r.Context(), e)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(out)
		case http.MethodGet:
			q := gauditor.Query{
				Tenant:   r.URL.Query().Get("tenant"),
				ActorID:  r.URL.Query().Get("actorId"),
				Action:   r.URL.Query().Get("action"),
				TargetID: r.URL.Query().Get("targetId"),
				Limit:    100,
			}
			if s := r.URL.Query().Get("limit"); s != "" {
				if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 1000 {
					q.Limit = n
				}
			}
			res, err := recorder.Query(r.Context(), q)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(res)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func run(addr string, handler http.Handler) error {
	log.Printf("gauditor listening on %s", addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	if os.Getenv("GAUDITOR_TEST_EXIT") == "1" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return err
		}
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = srv.Shutdown(context.Background())
		}()
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}
	if os.Getenv("GAUDITOR_TEST_NORMAL") == "1" {
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = srv.Shutdown(context.Background())
		}()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}
	// Default branch. For tests, allow graceful shutdown trigger without special listener.
	if os.Getenv("GAUDITOR_TEST_DEFAULT") == "1" {
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = srv.Shutdown(context.Background())
		}()
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func realMain() int {
	fs := flag.NewFlagSet("gauditor", flag.ContinueOnError)
	addr := fs.String("addr", ":8091", "HTTP listen address")
	_ = fs.Parse(os.Args[1:])

	if env := os.Getenv("GAUDITOR_ADDR"); env != "" {
		*addr = env
	}

	recorder := gauditor.NewRecorder(gauditor.NewMemoryStorage())
	handler := newServer(recorder)

	if os.Getenv("GAUDITOR_NO_SERVE") == "1" {
		// Allows tests to execute initialization paths without binding ports
		return 0
	}
	if err := run(*addr, handler); err != nil {
		log.Println("server error:", err)
		return 1
	}
	return 0
}

func main() {
	code := realMain()
	_ = context.Background()
	exitFunc(code)
}
