.DEFAULT_GOAL := help

VCV_VERSION ?= dev-$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
VCV_TAG ?= latest

.PHONY: help web-install web-dev web-build web-check web-test web-test-coverage dev docker-build test-offline test-dev go-update go-lint go-lint-full go-coverage

help:
	@printf '%s\n' \
		'Available targets:' \
		'  web-install        Install frontend dependencies (pnpm)' \
		'  web-dev            Run Vite dev server (proxies /api + /i18n to Go on :52000)' \
		'  web-build          Build Svelte frontend to app/web/dist' \
		'  web-check          Run svelte-check and tsc on the frontend' \
		'  web-test           Run frontend unit tests (Vitest)' \
		'  web-test-coverage  Run frontend unit tests with coverage' \
		'  dev                Build the Go binary and start the development Docker stack' \
		'  docker-build       Build and push multi-architecture Docker images' \
		'  test-offline       Run Go unit tests offline with coverage' \
		'  test-dev           Run Go tests against the development stack with coverage' \
		'  go-update          Update Go dependencies' \
		'  go-lint            Run go fmt and go vet' \
		'  go-lint-full       Run golangci-lint (same checks and version as CI)' \
		'  go-coverage        Run Go unit tests with coverage' \
		'' \
		'Variables:' \
		'  VCV_VERSION        Development binary and Docker Compose version (default: dev-<git SHA>)' \
		'  VCV_TAG            Docker image tag (default: latest)'

web-install:
	cd app/web/frontend && pnpm install

web-dev:
	cd app/web/frontend && pnpm dev

web-build:
	cd app/web/frontend && pnpm build
	touch app/web/dist/.gitkeep

web-check:
	cd app/web/frontend && pnpm check

web-test:
	cd app/web/frontend && pnpm test

web-test-coverage:
	cd app/web/frontend && pnpm test:coverage

dev: web-build
	go clean -cache && go build -C app -ldflags="-X vcv/internal/version.Version=$(VCV_VERSION)" -o ../vcv ./cmd/server
	rm -f vcv.log
	touch vcv.log
	docker compose -f docker-compose.dev.yml down
	VCV_VERSION=$(VCV_VERSION) docker compose -f docker-compose.dev.yml build
	VCV_VERSION=$(VCV_VERSION) docker compose -f docker-compose.dev.yml up -d

docker-build:
	docker buildx use vcv-builder
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(VCV_TAG) -t jhmmt/vcv:$(VCV_TAG) -t jhmmt/vcv:latest -t ghcr.io/julienhmmt/vcv:$(VCV_TAG) -t ghcr.io/julienhmmt/vcv:latest --push ./app

test-offline:
	cd app && go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic 2>&1 && go tool cover -func=coverage.out

test-dev:
	cd app && VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=root go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic 2>&1 && go tool cover -func=coverage.out

go-update:
	cd app && go get -u all && go mod tidy -v && go clean -cache -v

go-lint:
	cd app && go fmt ./... && go vet ./...

# Pinned to the same version as .github/workflows/go.yml; bump both together.
go-lint-full:
	cd app && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run --timeout=3m

go-coverage:
	cd app && go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic 2>&1 && go tool cover -func=coverage.out

# --- SonarQube (local container, Go-only) -------------------------------------
COMPOSE_SONAR := docker compose -f docker-compose.sonarqube.yml

.PHONY: sonar-up sonar-down sonar-logs sonar-status sonar-bootstrap sonar-coverage sonar-scan sonar-scan-server sonar-query sonar-clean

sonar-up:
	$(COMPOSE_SONAR) up -d sonarqube

sonar-down:
	$(COMPOSE_SONAR) down

sonar-logs:
	$(COMPOSE_SONAR) logs -f sonarqube

sonar-status:
	@docker ps --filter "name=vcv-sonarqube" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

sonar-bootstrap:
	@chmod +x tools/sonar-bootstrap.sh && tools/sonar-bootstrap.sh && echo "SonarQube token ready in .sonar/token"

sonar-coverage:
	@chmod +x tools/sonar-coverage.sh && GO_TEST_FLAGS="$(GO_TEST_FLAGS)" SERVER_DIR="$(SERVER_DIR)" tools/sonar-coverage.sh && echo "Coverage reports ready in .sonar/"

sonar-scan: sonar-coverage
	@chmod +x tools/sonar-scan.sh && tools/sonar-scan.sh

sonar-scan-server:
	@chmod +x tools/sonar-scan.sh && tools/sonar-scan.sh server

sonar-query:
	python3 tools/sonar-query.py $(CMD)

sonar-clean: sonar-down
	rm -rf .sonar
