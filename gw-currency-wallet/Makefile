APP_NAME := WalletService

ifeq ($(POSTGRES_SETUP),)
	POSTGRES_SETUP := user=wallet-user password=wallet-password dbname=wallet-db host=localhost port=5432 sslmode=disable
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

build:
	@echo Building application...
	@go build -o $(APP_NAME) ./cmd

deps:
	@echo Installing dependencies...
	@go mod tidy

run: build
	./$(APP_NAME)