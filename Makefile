.PHONY: build run run-worker migrate migrate-down migrate-status migrate-plan \
        health config test lint fmt module \
        docker-build docker-up docker-down docker-dev docker-dev-down \
        docker-migrate docker-logs docker-ps

APP_NAME  := app
BUILD_DIR := bin

# ── Local ────────────────────────────────────────────────
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/app

run:
	go run ./cmd/app serve

run-worker:
	go run ./cmd/app queue:work

migrate:
	go run ./cmd/app migrate:up

migrate-down:
	go run ./cmd/app migrate:down

migrate-status:
	go run ./cmd/app migrate:status

migrate-plan:
	go run ./cmd/app migrate:plan

health:
	go run ./cmd/app health

config:
	go run ./cmd/app config:dump

test:
	go test -race -count=1 -timeout=30s ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	goimports -local github.com/shuldan/skeleton -w .

# ── Scaffolding ──────────────────────────────────────────
module:
	@echo "Enter module name:" && \
	read -r MODULE && \
	if [ -z "$$MODULE" ]; then \
		echo "Module name cannot be empty"; \
		exit 1; \
	fi && \
	MODULE_DIR="internal/module/$$MODULE" && \
	if [ -d "$$MODULE_DIR" ]; then \
		echo "Module '$$MODULE' already exists at $$MODULE_DIR"; \
		echo "   Aborting to avoid data loss."; \
		exit 1; \
	fi && \
	echo "Creating module structure for: $$MODULE" && \
	mkdir -p "$$MODULE_DIR"/{domain/{model,persistence,business/{emitter,operation}},application/{interactor,business/{emitter,operation},port},infrastructure/{persistence,migration,adapter},presentation/{api,job,listener}} && \
	printf '%s\n' "package $${MODULE}" > "$$MODULE_DIR/module.go" && \
	echo "✅ Module '$$MODULE' created at $$MODULE_DIR"

# ── Docker (production) ──────────────────────────────────
docker-build:
	docker build \
		-f deployments/Dockerfile \
		--build-arg GO_VERSION=$(or $(GO_VERSION),1.25-alpine) \
		-t $(APP_NAME) .

docker-up:
	docker compose \
		--env-file deployments/.env \
		-f deployments/docker-compose.yml \
		up -d

docker-down:
	docker compose \
		--env-file deployments/.env \
		-f deployments/docker-compose.yml \
		down

# ── Docker (development) ─────────────────────────────────
docker-dev:
	docker compose \
		--env-file deployments/.env \
		-f deployments/docker-compose.yml \
		-f deployments/docker-compose.dev.yml \
		up -d

docker-dev-down:
	docker compose \
		--env-file deployments/.env \
		-f deployments/docker-compose.yml \
		-f deployments/docker-compose.dev.yml \
		down

# ── Docker utilities ─────────────────────────────────────
docker-migrate:
	docker compose \
		--env-file deployments/.env \
		-f deployments/docker-compose.yml \
		run --rm migrate

docker-logs:
	docker compose \
		--env-file deployments/.env \
		-f deployments/docker-compose.yml \
		logs -f $(SVC)

docker-ps:
	docker compose \
		--env-file deployments/.env \
		-f deployments/docker-compose.yml \
		ps
