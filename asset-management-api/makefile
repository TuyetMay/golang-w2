
# Makefile
.PHONY: build run test clean docker-build docker-run setup

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=asset-management-api
BINARY_PATH=./cmd/server

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(BINARY_PATH)

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v $(BINARY_PATH) && ./$(BINARY_NAME)

# Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Test the application
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Download dependencies
deps:
	$(GOMOD) download

# Docker build
docker-build:
	docker build -t $(BINARY_NAME) .

# Docker run
docker-run:
	docker run -p 8000:8000 --env-file .env $(BINARY_NAME)

# Setup development environment
setup:
	$(GOMOD) download
	docker-compose up -d
	sleep 5
	PGPASSWORD=password123 psql -h localhost -U postgres -d asset_db -f migrations/001_create_tables.sql

# Stop development environment
stop:
	docker-compose down

# Database migration
migrate:
	PGPASSWORD=password123 psql -h localhost -U postgres -d asset_db -f migrations/001_create_tables.sql

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Security scan (requires gosec)
security:
	gosec ./...