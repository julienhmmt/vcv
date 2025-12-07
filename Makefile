# Makefile VCV

.PHONY: help build

# Couleurs pour l'affichage
BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

help:
	@echo "$(BLUE)VCV - Commandes disponibles:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'
	@echo ""

dev: ## Construit le binaire Go et le container docker
	@echo "Building binary" && cd /Users/jh/git/vcv && go clean -cache && go build -C app -o ../vcv ./cmd/server
	@echo "Binary built successfully"
	@echo ""
	@echo "Remove old app logs and create new one"
	@rm -f vcv.log
	@touch vcv.log
	@echo "Successfully cleaned app logs"
	@echo "Building and running docker container"
	docker compose -f docker-compose.dev.yml down
	docker buildx build --platform linux/arm64 --load \
		--build-arg VERSION=dev \
		-t jhmmt/vcv:dev ./app
	docker compose -f docker-compose.dev.yml up -d
	@echo ""

docker-build: ## Construit les images docker (arm64 et amd64) et push sur Docker Hub
	@echo "Building Docker images for multiple architectures..."
	@echo "Usage: make docker-build VCV_TAG=your-tag"
	@echo "Default tag: latest"
	@docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(VCV_TAG) -t jhmmt/vcv:$(or $(VCV_TAG),latest)

test-offline: ## Run unit tests offline (no Vault) with coverage
	cd app && go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic

test-dev: ## Run tests against dev stack (docker-compose)
	cd app && VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=root go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic
