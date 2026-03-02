APP_NAME := ExchangeService

ifeq ($(POSTGRES_SETUP),)
	POSTGRES_SETUP := user=exchange-user password=exchange-password dbname=exchange-db host=localhost port=5432 sslmode=disable
endif

ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=exchange-user password=exchange-password dbname=exchange-db-test host=localhost port=5433 sslmode=disable
endif

INTERNAL_PATH=$(CURDIR)/internal/storage
MIGRATION_FOLDER=$(INTERNAL_PATH)/migrations

.PHONY: compose-up
compose-up:
	docker-compose up -d

.PHONY: migration-create
migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "$(name)" sql

.PHONY: migration-up
migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP)" up

.PHONY: migration-down
migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP)" down

.PHONY: test-migration-up
test-migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up

.PHONY: test-migration-down
test-migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" down

build:
	@echo Building application...
	@go build -o $(APP_NAME) ./cmd

deps:
	@echo Installing dependencies...
	@go mod tidy

run: build
	./$(APP_NAME)