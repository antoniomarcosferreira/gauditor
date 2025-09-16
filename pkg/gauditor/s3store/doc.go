// Package s3store provides an S3-backed gauditor.Storage implementation.
//
// Each event is stored as a JSON object under a configurable prefix and tenant.
// Queries list and filter client-side and are suitable for append-only/archive use.
package s3store
