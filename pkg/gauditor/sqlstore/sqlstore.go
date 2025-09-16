package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	gauditor "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
)

// Store implements gauditor.Storage using database/sql.
// Compatible with Postgres and MySQL given the simple schema.
type Store struct {
	bb    *sql.DB
	table string
}

// New constructs a Store backed by the provided *sql.DB.
func New(db *sql.DB) *Store { return &Store{bb: db, table: "gauditor_events"} }

// Option configures the Store.
type Option func(*Store)

// WithTablePrefix sets a prefix for the table name. The base table is "gauditor_events".
func WithTablePrefix(prefix string) Option {
	return func(s *Store) { s.table = prefix + "gauditor_events" }
}

// WithTableName sets an explicit table name.
func WithTableName(name string) Option { return func(s *Store) { s.table = name } }

// ApplyOptions applies the provided options to the store and returns it.
func (s *Store) ApplyOptions(opts ...Option) *Store {
	for _, o := range opts {
		o(s)
	}
	return s
}

// EnsureSchema creates the minimal table if it does not exist.
// For Postgres/MySQL compatible types.
func (s *Store) EnsureSchema(ctx context.Context) error {
	stmt := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
  id        VARCHAR(64) PRIMARY KEY,
  ts        TIMESTAMP NOT NULL,
  tenant    VARCHAR(128) NOT NULL,
  actor_id  VARCHAR(128) NULL,
  action    VARCHAR(128) NOT NULL,
  target_id VARCHAR(128) NULL,
  actor_json   TEXT NULL,
  target_json  TEXT NULL,
  data_json    TEXT NULL
);
CREATE INDEX IF NOT EXISTS idx_%s_tenant_ts ON %s(tenant, ts);
`, s.table, s.table, s.table)
	_, err := s.bb.ExecContext(ctx, stmt)
	return err
}

// Save inserts the event row. JSON columns store full structs as JSON.
func (s *Store) Save(ctx context.Context, e gauditor.Event) (gauditor.Event, error) {
	actorJSON, targetJSON, dataJSON, err := marshalParts(e)
	if err != nil {
		return e, err
	}
	query := fmt.Sprintf("INSERT INTO %s (id, ts, tenant, actor_id, action, target_id, actor_json, target_json, data_json) VALUES (?,?,?,?,?,?,?,?,?)", s.table)
	_, err = s.bb.ExecContext(ctx,
		query,
		e.ID, e.Timestamp, e.Tenant, e.Actor.ID, e.Action, e.Target.ID, actorJSON, targetJSON, dataJSON,
	)
	return e, err
}

// Query selects rows with simple filters and maps them back to events.
func (s *Store) Query(ctx context.Context, q gauditor.Query) ([]gauditor.Event, error) {
	where := "WHERE 1=1"
	args := make([]any, 0, 6)
	if q.Tenant != "" {
		where += " AND tenant = ?"
		args = append(args, q.Tenant)
	}
	if q.ActorID != "" {
		where += " AND actor_id = ?"
		args = append(args, q.ActorID)
	}
	if q.Action != "" {
		where += " AND action = ?"
		args = append(args, q.Action)
	}
	if q.TargetID != "" {
		where += " AND target_id = ?"
		args = append(args, q.TargetID)
	}
	if q.Since != nil {
		where += " AND ts >= ?"
		args = append(args, q.Since)
	}
	if q.Until != nil {
		where += " AND ts <= ?"
		args = append(args, q.Until)
	}
	limit := ""
	if q.Limit > 0 {
		limit = " LIMIT ?"
		args = append(args, q.Limit)
	}
	qstr := fmt.Sprintf("SELECT id, ts, tenant, actor_id, action, target_id, actor_json, target_json, data_json FROM %s ", s.table) + where + " ORDER BY ts ASC" + limit
	rows, err := s.bb.QueryContext(ctx, qstr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]gauditor.Event, 0)
	for rows.Next() {
		var id, tenant, actorID, action, targetID string
		var ts time.Time
		var actorJSON, targetJSON, dataJSON sql.NullString
		if err := rows.Scan(&id, &ts, &tenant, &actorID, &action, &targetID, &actorJSON, &targetJSON, &dataJSON); err != nil {
			return nil, err
		}
		e := gauditor.Event{ID: id, Timestamp: ts, Tenant: tenant, Actor: gauditor.Actor{ID: actorID}, Action: action, Target: gauditor.Target{ID: targetID}}
		if actorJSON.Valid {
			_ = jsonUnmarshal([]byte(actorJSON.String), &e.Actor)
		}
		if targetJSON.Valid {
			_ = jsonUnmarshal([]byte(targetJSON.String), &e.Target)
		}
		if dataJSON.Valid {
			_ = jsonUnmarshal([]byte(dataJSON.String), &e.Data)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// helpers
func marshalParts(e gauditor.Event) (actor, target, data []byte, err error) {
	if actor, err = jsonMarshal(e.Actor); err != nil {
		return
	}
	if target, err = jsonMarshal(e.Target); err != nil {
		return
	}
	if data, err = jsonMarshal(e.Data); err != nil {
		return
	}
	return
}

var (
	jsonMarshal   = func(v any) ([]byte, error) { return json.Marshal(v) }
	jsonUnmarshal = func(b []byte, v any) error { return json.Unmarshal(b, v) }
)
