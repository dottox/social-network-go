include .env


# Command to create new migration files
# Filter-out is used to extract the migration name from the make command
.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

# Command to run all the migrations
.PHONY: migrate-up
migrate-up:
	@migrate --path=$(MIGRATIONS_PATH) --database="$(DB_ADDR)" up

# Command to rollback the last n migrations
.PHONY: migrate-down
migrate-down:
	@migrate --path=$(MIGRATIONS_PATH) --database="$(DB_ADDR)" down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-version
migrate-version:
	@migrate --path=$(MIGRATIONS_PATH) --database="$(DB_ADDR)" version

.PHONY: migrate-drop
migrate-drop:
	@migrate --path=$(MIGRATIONS_PATH) --database="$(DB_ADDR)" drop $(filter-out $@,$(MAKECMDGOALS))

.PHONY: force-version
force-version:
	@migrate --path=$(MIGRATIONS_PATH) --database="$(DB_ADDR)" force $(filter-out $@,$(MAKECMDGOALS))

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./main.go -d cmd,internal/api,internal/db,internal/model,internal/store,internal/env && swag fmt