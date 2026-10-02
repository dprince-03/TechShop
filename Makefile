COMPOSE := docker compose -f infra/docker/compose.yml

.PHONY: install infra-up infra-down infra-logs db-up \
	dev-api dev-web dev-corporate dev-market dev-wholesale dev-seller dev-staff \
	dev-customer dev-logistics tokens \
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

# ---------- Database ----------

migrate-up migrate-down migrate-status sqlc:
	$(MAKE) -C backend $@

migrate-create: ## Usage: make migrate-create name=create_users
	$(MAKE) -C backend migrate-create name=$(name)

# ---------- Everything CI runs ----------

check:
	cd shared/design-tokens && node scripts/build-css.ts --check
	cd backend && test -z "$$(gofmt -l .)" && go vet ./... && go test ./... && go build -o /dev/null ./cmd/api
	cd frontend && npm run format:check && npm run lint && npm run typecheck && npm run build
	cd mobile && CI=1 npm run lint && npm run typecheck
