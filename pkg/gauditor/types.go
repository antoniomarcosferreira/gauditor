package gauditor

import "time"

// Actor describes who performed the action.
// ID should be a stable identifier for the actor (for example, a user ID).
// Optional contextual attributes can be attached via Attributes.
type Actor struct {
	ID         string         `json:"id,omitempty"`
	IP         string         `json:"ip,omitempty"`
	UserAgent  string         `json:"userAgent,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Target describes what the action was performed on.
type Target struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

// Event is the core audit record.
// ID and Timestamp are populated by the Recorder if unset.
type Event struct {
	ID        string         `json:"id"`
	Timestamp time.Time      `json:"timestamp"`
	Tenant    string         `json:"tenant"`
	Actor     Actor          `json:"actor"`
	Action    string         `json:"action"`
	Target    Target         `json:"target,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

// Query defines filters for retrieving events.
// Limit applies after filtering; storage may cap the maximum.
type Query struct {
	Tenant   string     `json:"tenant,omitempty"`
	ActorID  string     `json:"actorId,omitempty"`
	Action   string     `json:"action,omitempty"`
	TargetID string     `json:"targetId,omitempty"`
	Since    *time.Time `json:"since,omitempty"`
	Until    *time.Time `json:"until,omitempty"`
	Limit    int        `json:"limit,omitempty"`
}
