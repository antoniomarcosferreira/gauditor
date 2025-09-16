// Package redisstore provides a simple Redis-backed gauditor.Storage implementation.
//
// It stores events in a Redis list per tenant (newest-first) and filters
// in application code on query. This is ideal for development and demos.
// Production users may want to implement pagination and sharding.
package redisstore
