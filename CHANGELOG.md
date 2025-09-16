# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project adheres to Semantic Versioning.

## [Unreleased]

- Placeholder for upcoming changes

## [v0.0.1] - 2025-09-15

Initial alpha release.

Highlights:
- Core `Recorder` with options (`WithIDGenerator`, `WithClock`)
- `Storage` interface and implementations: `MemoryStorage`, `redisstore`, `sqlstore`, `s3store`
- Env-based bootstrap: `gauditorenv.NewRecorderFromEnv`
- HTTP server (`cmd/gauditor`) and examples (`basic`, `httpclient`, `gincrud`, `redis`)
- Documentation for packages and storage backends; CI with tests, coverage, and `govulncheck`

[Unreleased]: https://github.com/antoniomarcosferreira/gauditor/compare/v0.0.1...HEAD
[v0.0.1]: https://github.com/antoniomarcosferreira/gauditor/releases/tag/v0.0.1
