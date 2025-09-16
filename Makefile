SHELL := /bin/bash

.PHONY: tidy build test lint format clean coverage coverage-watch help release-tag release-print build-versioned

# Usage:
#   make test        # run unit tests
#   make coverage    # run tests with coverage and generate docs/coverage.html
#   make coverage-watch  # watch files (requires fswatch) and regenerate coverage
# After running `make coverage`, open the dashboard at docs/coverage.html

help: ## Show common commands
	@echo "make test              # run unit tests"
	@echo "make coverage          # generate HTML coverage report at docs/coverage.html"
	@echo "make coverage-watch    # watch and regenerate coverage (requires fswatch)"
	@echo "make release-tag       # tag repo with VERSION and push tags"
	@echo "make release-print     # print current VERSION"

tidy:
	go mod tidy

build:
	go build ./...

build-versioned:
	go build -ldflags "-X main.version=$(shell cat VERSION)" ./...

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

release-print:
	@cat VERSION

release-tag:
	git tag $(shell cat VERSION)
	git push origin $(shell cat VERSION)

