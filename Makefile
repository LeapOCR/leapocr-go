.PHONY: help generate generate-full build test lint clean install format test-coverage test-integration examples examples-run dev-setup dev-reset ci-test ci-test-full release-check tidy

OPENAPI_URL := http://localhost:8080/api/v1/docs/openapi.json

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-25s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	go mod download
	go mod tidy

tidy: ## Tidy and format code
	go mod tidy
	gofumpt -w .
	go fmt ./...

generate: ## Generate Go SDK for ALL endpoints 
	@. ./scripts/setup-java-path.sh || exit 1
	curl -s $(OPENAPI_URL) > openapi.json
	@. ./scripts/setup-java-path.sh && npx @openapitools/openapi-generator-cli generate \
		-i openapi.json \
		-g go \
		-o ./internal/generated \
		--skip-validate-spec \
		--additional-properties=packageName=generated,generateInterfaces=true,structPrefix=true,enumClassPrefix=true,generateTests=false \
		--global-property=apiDocs=false,modelDocs=false,withGoMod=false,generateInterfaces=true,apiTests=false,modelTests=false
	@rm -f ./internal/generated/go.mod ./internal/generated/go.sum
	@make tidy

build: ## Build the SDK
	go build ./...

test: ## Run unit tests
	go test -race -v ./...

test-coverage: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests (requires LEAPOCR_API_KEY)
	@if [ -z "$$LEAPOCR_API_KEY" ]; then \
		echo "LEAPOCR_API_KEY environment variable is required"; \
		exit 1; \
	fi
	go test -race -v -tags=integration ./test/integration/...

lint: ## Run linter
	golangci-lint run ./...
	go vet $(shell go list ./... | grep -v '/internal/generated')
	gosec -exclude-generated ./...

format: ## Format code
	go fmt ./...
	goimports -w . 2>/dev/null || echo "goimports not available, using go fmt only"
	gofumpt -w .

examples: ## Build all examples
	@for example in examples/*/; do \
		if [ -f "$$example/main.go" ]; then \
			cd "$$example" && go build . && cd ../..; \
		fi; \
	done

examples-run: ## Run all examples (requires LEAPOCR_API_KEY)
	@if [ -z "$$LEAPOCR_API_KEY" ]; then \
		echo "LEAPOCR_API_KEY environment variable is required to run examples"; \
		exit 1; \
	fi
	@for example in examples/*/; do \
		if [ -f "$$example/main.go" ]; then \
			cd "$$example" && timeout 30s go run . || echo "Example completed or timed out" && cd ../..; \
		fi; \
	done

clean: ## Clean build artifacts and generated files
	go clean ./...
	rm -f coverage.out coverage.html
	rm -rf internal/generated/
	rm -f openapi.json openapi-full.json openapi-sdk.json

dev-setup: install generate build ## Complete development setup

dev-reset: clean dev-setup ## Reset development environment completely

ci-test: lint test ## Run all CI tests

ci-test-full: lint test test-integration examples ## Run full CI test suite

release-check: ci-test examples ## Pre-release validation