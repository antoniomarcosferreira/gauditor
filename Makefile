SHELL := /bin/bash

.PHONY: tidy build test lint format clean coverage coverage-watch help

# Usage:
#   make test        # run unit tests
#   make coverage    # run tests with coverage and generate docs/coverage.html
#   make coverage-watch  # watch files (requires fswatch) and regenerate coverage
# After running `make coverage`, open the dashboard at docs/coverage.html

help: ## Show common commands
	@echo "make test              # run unit tests"
	@echo "make coverage          # generate HTML coverage report at docs/coverage.html"
	@echo "make coverage-watch    # watch and regenerate coverage (requires fswatch)"

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

