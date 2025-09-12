SHELL := /bin/bash

.PHONY: tidy build test lint format clean coverage coverage-watch

tidy:
	go mod tidy

build:
	go build ./...

test:
	go test ./...

coverage:
	bash scripts/update_coverage.sh

coverage-watch:
	bash scripts/watch_coverage.sh

lint:
	golangci-lint run || true

format:
	gofmt -s -w .

clean:
	rm -rf bin dist build

