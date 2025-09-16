// Package sqlstore implements gauditor.Storage via database/sql for Postgres/MySQL.
//
// The default table name is "gauditor_events". Use WithTablePrefix or WithTableName
// to change it when sharing the database with your application. EnsureSchema creates
// the table and index if missing. Save/Query provide a minimal, portable mapping.
package sqlstore
