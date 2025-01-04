include .envrc
MIGRATION_PATH=./cmd/migrate/migrations

.PHONY: migrate-create
migration: 
	@migrate create -seq -ext  sql -dir $(MIGRATION_PATH) $(filter-out $@, $(MAKECMDGOALS))

.PHONY: migrate-up
migration-up:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) up

.PHONY: migrate-down
migration-down:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) down $(filter-out $@, $(MAKECMDGOALS))

.PHONY: seed
seed:
	@go run ./cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt