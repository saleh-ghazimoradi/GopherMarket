.PHONY: help docker-up docker-down fmt vet run build lint migrate-up migrate-down

help:
	@echo "Available commands:"
	@echo "  make build      - Build the GopherMarket Application"
	@echo "  make run      	 - Run the GopherMarket Application"
	@echo "  make lint       - Run the linter on the codebase"
	@echo "  make docker-up  - Run all the docker containers"
	@echo "  make docker-down- Stop all the docker containers"
	@echo "  make fmt      	 - Format all the code"
	@echo "  make vet      	 - Run vet on all the codebase"
	@echo "  make swagger    - Generate Swagger Docs"

docker-up:
	docker compose up -d

docker-down:
	docker compose down

fmt:
	go fmt ./...

vet:
	go vet ./...

run: fmt vet
	go run . run

build:
	mkdir -p bin
	go build -o bin/gopherMarket

lint:
	golangci-lint run ./...

migrate-up:
	go run . migrateUp

migrate-down:
	go run . migrateDown

swagger:
	mkdir -p docs
	swag init -g main.go -o docs --parseDependency --parseInternal --exclude .git,docker-compose.yml,infra



