// Package gauditorenv provides a helper to construct a gauditor.Recorder from
// environment variables, enabling a Rails-like developer experience.
//
// Supported backends via GAUDITOR_STORAGE: memory (default), redis, sql, s3.
// See docs/Storage.md and README for the list of environment variables.
package gauditorenv
