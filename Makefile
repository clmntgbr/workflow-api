COMPOSE_DEV := docker compose -f compose.dev.yaml

# ============================================
# Development (compose.dev.yaml)
# ============================================

dev:
	$(COMPOSE_DEV) up -d --build

dev-down:
	$(COMPOSE_DEV) down

dev-logs:
	$(COMPOSE_DEV) logs -f api worker executor scheduler

api-logs:
	$(COMPOSE_DEV) logs -f api

worker-logs:
	$(COMPOSE_DEV) logs -f worker

executor-logs:
	$(COMPOSE_DEV) logs -f executor

scheduler-logs:
	$(COMPOSE_DEV) logs -f scheduler

dev-restart:
	$(COMPOSE_DEV) restart api worker executor scheduler

lint:
	$(COMPOSE_DEV) exec api golangci-lint run --fix

# ============================================
# CLI Commands (via Docker)
# ============================================

cli-build:
	@$(COMPOSE_DEV) exec api go build -o bin/cli ./cmd/cli

migrate: cli-build
	@echo "Applying SQL migrations..."
	@$(COMPOSE_DEV) exec api ./bin/cli migrate up
	@echo "Checking model/DB schema..."
	@$(COMPOSE_DEV) exec api ./bin/cli migrate check

migrate-status: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli migrate status

migrate-down: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli migrate down

migrate-check: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli migrate check

shell:
	$(COMPOSE_DEV) exec api sh

workflow-run: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli workflow run 3aa5ce64-6859-4d17-bf67-607b6fff4144