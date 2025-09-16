package gauditorenv

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"

	gauditor "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
	"github.com/antoniomarcosferreira/gauditor/pkg/gauditor/redisstore"
	"github.com/antoniomarcosferreira/gauditor/pkg/gauditor/s3store"
	"github.com/antoniomarcosferreira/gauditor/pkg/gauditor/sqlstore"
)

// NewRecorderFromEnv creates a Recorder using storage configured via environment variables.
//
// Variables:
//
//	GAUDITOR_STORAGE: memory | redis | sql | s3 (default: memory)
//	Redis: REDIS_ADDR (default 127.0.0.1:6379), REDIS_KEY_PREFIX (default "gauditor:")
//	SQL:   SQL_DRIVER (postgres|mysql), SQL_DSN (driver-specific DSN)
//	       GAUDITOR_SQL_ENSURE_SCHEMA=1 (default) to auto-create table
//	S3:    S3_BUCKET (required), S3_PREFIX (default "gauditor") + standard AWS_* envs
func NewRecorderFromEnv(ctx context.Context, opts ...gauditor.Option) (*gauditor.Recorder, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv("GAUDITOR_STORAGE")))
	if backend == "" || backend == "memory" {
		return gauditor.NewRecorder(gauditor.NewMemoryStorage(), opts...), nil
	}

	switch backend {
	case "redis":
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = "127.0.0.1:6379"
		}
		prefix := os.Getenv("REDIS_KEY_PREFIX")
		if prefix == "" {
			prefix = "gauditor:"
		}
		rdb := redis.NewClient(&redis.Options{Addr: addr})
		store := redisstore.New(rdb, redisstore.WithKeyPrefix(prefix))
		return gauditor.NewRecorder(store, opts...), nil

	case "sql":
		driver := os.Getenv("SQL_DRIVER")
		dsn := os.Getenv("SQL_DSN")
		db, err := sql.Open(driver, dsn)
		if err != nil {
			return nil, err
		}
		store := sqlstore.New(db)
		if os.Getenv("GAUDITOR_SQL_ENSURE_SCHEMA") != "0" {
			if err := store.EnsureSchema(ctx); err != nil {
				return nil, err
			}
		}
		return gauditor.NewRecorder(store, opts...), nil

	case "s3":
		bucket := os.Getenv("S3_BUCKET")
		prefix := os.Getenv("S3_PREFIX")
		if prefix == "" {
			prefix = "gauditor"
		}
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
		cli := s3.NewFromConfig(cfg)
		store := s3store.New(cli, bucket, prefix)
		return gauditor.NewRecorder(store, opts...), nil
	}

	// Fallback to memory for unknown backend
	return gauditor.NewRecorder(gauditor.NewMemoryStorage(), opts...), nil
}
