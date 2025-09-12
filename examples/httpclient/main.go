package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// Ingest an event via HTTP
	payload := map[string]any{
		"tenant": "acme",
		"actor": map[string]any{
			"id": "uuid",
			"attributes": map[string]any{
				"type":  "user",
				"email": "alice@example.com",
				"name":  "Alice",
			},
		},
		"action": "login",
		"data":   map[string]any{"method": "password"},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:8080/v1/events", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	created, _ := io.ReadAll(resp.Body)
	fmt.Printf("created: %s\n", string(created))

	// Query events via HTTP
	res, err := http.Get("http://localhost:8080/v1/events?tenant=acme")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	list, _ := io.ReadAll(res.Body)
	fmt.Printf("query: %s\n", string(list))
}
