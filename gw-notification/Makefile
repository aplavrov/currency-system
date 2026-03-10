APP_NAME := NotificationService

.PHONY: compose-up
compose-up:
	docker-compose up -d

build:
	@echo Building application...
	@go build -o $(APP_NAME) ./cmd

deps:
	@echo Installing dependencies...
	@go mod tidy

run: build
	./$(APP_NAME)