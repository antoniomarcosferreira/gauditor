package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	recorder := gauditor.NewRecorder(gauditor.NewMemoryStorage())

	http.HandleFunc("/v1/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var e gauditor.Event
			if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("invalid JSON"))
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

	log.Printf("gauditor listening on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Println("server error:", err)
		os.Exit(1)
	}

	_ = context.Background()
}
