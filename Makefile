COMPOSE := docker compose -f infra/docker/compose.yml

.PHONY: install infra-up infra-down infra-logs db-up db-users \
	dev-api dev-web dev-corporate dev-market dev-wholesale dev-seller dev-staff \
	dev-customer dev-logistics tokens api-types test-api-client \
	migrate-up migrate-down migrate-status migrate-create sqlc check

# ---------- Setup ----------

install:
	cd frontend && npm install
	cd mobile && npm install
	cd backend && go mod download

# ---------- Local infrastructure (Postgres + nginx) ----------

infra-up:
	$(COMPOSE) up -d

infra-down:
	$(COMPOSE) down

infra-logs:
	$(COMPOSE) logs -f

db-up: ## Postgres only
	$(COMPOSE) up -d postgres

db-users: ## Create the API's least-privilege login (after migrations): make db-users APP_DB_PASSWORD=…
	@test $${#APP_DB_PASSWORD} -ge 16 || (echo "set APP_DB_PASSWORD (16+ random characters)" && exit 1)
	$(COMPOSE) exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -v app_password="$$1" -f -' sh "$$APP_DB_PASSWORD" < infra/docker/postgres/app-users.sql

# ---------- Run ----------

dev-api:
	$(MAKE) -C backend run

dev-web: ## All five web apps
	cd frontend && npm run dev

dev-corporate dev-market dev-wholesale dev-seller dev-staff:
	cd frontend && npm run $(subst dev-,dev:,$@)

dev-customer:
	cd mobile && npm run customer

dev-logistics:
	cd mobile && npm run logistics

# ---------- Design tokens ----------

tokens: ## Regenerate web CSS from shared/design-tokens
	cd shared/design-tokens && node scripts/build-css.ts

# ---------- API contract ----------

api-types: ## Regenerate shared/api-client types from backend/api/openapi.yaml
	cd frontend && npm run api:types

test-api-client: ## Unit tests; contract tests need a seeded local API: API_URL=… SEED_FILE=… npm run test:contract
	cd shared/api-client && ../../frontend/node_modules/.bin/tsc -p . && npm test

# ---------- Database ----------

migrate-up migrate-down migrate-status sqlc:
	$(MAKE) -C backend $@

migrate-create: ## Usage: make migrate-create name=create_users
	$(MAKE) -C backend migrate-create name=$(name)

# ---------- Everything CI runs ----------

check:
	cd shared/design-tokens && node scripts/build-css.ts --check
	cd frontend && npm run api:types && git diff --exit-code -- ../shared/api-client/src/generated
	$(MAKE) test-api-client
	cd backend && test -z "$$(gofmt -l .)" && go vet ./... && go test ./... && go build -o /dev/null ./cmd/api
	cd frontend && npm run format:check && npm run lint && npm run typecheck && npm run build
	cd mobile && CI=1 npm run lint && npm run typecheck && npm test
