# Makefile
.PHONY: build run test clean docker-build docker-run setup redis-cli

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
	$(GOGET) github.com/redis/go-redis/v9

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
	sleep 10
	PGPASSWORD=iloveyou044 psql -h localhost -p 5433 -U postgres -d asset_db -f migrations/001_create_tables.sql

# Stop development environment
stop:
	docker-compose down

# Database migration
migrate:
	PGPASSWORD=iloveyou044 psql -h localhost -p 5433 -U postgres -d asset_db -f migrations/001_create_tables.sql

# NEW: Redis operations
redis-cli:
	docker exec -it redis redis-cli

# NEW: Check Redis cache contents
redis-check:
	@echo "=== Redis Cache Status ==="
	@docker exec -it redis redis-cli info memory | grep used_memory_human
	@echo "=== Cache Keys Count ==="
	@docker exec -it redis redis-cli eval "return #redis.call('keys', '*')" 0
	@echo "=== Team Cache Keys ==="
	@docker exec -it redis redis-cli keys "team:*"
	@echo "=== Asset Cache Keys ==="
	@docker exec -it redis redis-cli keys "folder:*" 
	@docker exec -it redis redis-cli keys "note:*"
	@echo "=== ACL Cache Keys ==="
	@docker exec -it redis redis-cli keys "asset:*:acl"

# NEW: Clear Redis cache
redis-clear:
	docker exec -it redis redis-cli flushdb

# NEW: Monitor Redis operations
redis-monitor:
	docker exec -it redis redis-cli monitor

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Security scan (requires gosec)
security:
	gosec ./...

# NEW: Test Kafka and Redis integration
test-integration:
	@echo "Testing Kafka connectivity..."
	@curl -s http://localhost:8000/health | jq '.data.cache'
	@echo "\nTesting Redis cache..."
	@docker exec -it redis redis-cli ping
	@echo "Integration test completed!"

# NEW: Load test endpoints
load-test:
	@echo "Running basic load test..."
	@curl -X POST http://localhost:8000/test/login -s | jq -r '.data.token' > /tmp/token
	@echo "Token generated, testing folder operations..."
	@for i in {1..10}; do \
		curl -s -X POST http://localhost:8000/api/v1/folders \
		-H "Authorization: Bearer $$(cat /tmp/token)" \
		-H "Content-Type: application/json" \
		-d '{"name":"Test Folder '$$i'","description":"Load test folder"}' > /dev/null; \
	done
	@echo "Load test completed! Check Redis cache:"
	@make redis-check