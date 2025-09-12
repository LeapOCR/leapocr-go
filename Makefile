.PHONY: help generate-sdk-endpoints generate-full build test lint clean install format coverage examples validate-spec dev-setup ci-test release-check

# Variables
OPENAPI_URL := http://localhost:8080/api/v1/swagger.json
GENERATOR_VERSION := 7.9.0
PACKAGE_NAME := github.com/leapocr/leapocr-go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-25s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	go mod download
	go mod tidy

# SDK Generation Targets
generate-sdk-endpoints: ## Generate Go SDK only for SDK-tagged endpoints (recommended)
	@echo "🎯 Generating SDK for SDK-tagged endpoints only..."
	@echo "📋 Step 1: Downloading OpenAPI spec"
	curl -s $(OPENAPI_URL) > openapi-full.json
	@echo "📋 Step 2: Filtering SDK endpoints"
	@./scripts/filter-sdk-endpoints.sh openapi-full.json openapi-sdk.json
	@echo "📋 Step 3: Generating Go client from filtered spec"
	openapi-generator-cli generate \
		-i openapi-sdk.json \
		-g go \
		-o ./generated-sdk \
		--skip-validate-spec \
		--additional-properties=packageName=gen,generateInterfaces=true,structPrefix=false,enumClassPrefix=false \
		--global-property=models,apis,supportingFiles
	@echo "📋 Step 4: Organizing generated code"
	@mkdir -p gen/
	@# Clean up old generated files first
	@rm -f gen/model_*.go gen/api_*.go gen/client.go gen/configuration.go gen/response.go gen/utils.go 2>/dev/null || true
	@# Copy models and core files
	@cp generated-sdk/model_*.go gen/ 2>/dev/null || echo "No model files found"
	@cp generated-sdk/client.go gen/ 2>/dev/null || echo "No client file found"
	@cp generated-sdk/configuration.go gen/ 2>/dev/null || echo "No configuration file found"
	@cp generated-sdk/response.go gen/ 2>/dev/null || echo "No response file found"
	@cp generated-sdk/utils.go gen/ 2>/dev/null || echo "No utils file found"
	@# Copy only one API file to avoid conflicts (prefer SDK API)
	@cp generated-sdk/api_sdk.go gen/ 2>/dev/null || cp generated-sdk/api_ocr.go gen/ 2>/dev/null || echo "No API files found"
	@echo "📋 Step 5: Cleaning up temporary files"
	@rm -rf generated-sdk openapi-full.json openapi-sdk.json
	@echo "📋 Step 6: Fixing generated client"
	@./scripts/fix-generated-client.sh
	@echo "📋 Step 7: Formatting generated code"
	@go fmt ./gen/... 2>/dev/null || echo "No generated package to format"
	@gofumpt -w .
	@go mod tidy
	@echo "✅ SDK generation complete!"

generate-full: ## Generate Go SDK for ALL endpoints (not recommended)
	@echo "⚠️  WARNING: Generating SDK for ALL endpoints (not just SDK-tagged ones)"
	@echo "📋 Downloading OpenAPI spec..."
	curl -s $(OPENAPI_URL) > openapi.json
	@echo "📋 Generating Go client..."
	openapi-generator-cli generate \
		-i openapi.json \
		-g go \
		-o ./generated \
		--skip-validate-spec \
		--additional-properties=packageName=generated,generateInterfaces=true,structPrefix=true,enumClassPrefix=true
	@echo "📋 Copying generated files..."
	@mkdir -p types/
	@cp -r generated/*.go types/ 2>/dev/null || true
	@cp -r generated/model_*.go types/ 2>/dev/null || true
	@echo "📋 Cleaning up..."
	@rm -rf generated openapi.json
	@echo "📋 Formatting generated code..."
	@go fmt ./types/...
	@echo "✅ Full generation complete!"

# Default generation (SDK endpoints only)
generate: generate-sdk-endpoints ## Generate SDK (defaults to SDK endpoints only)

# Analysis and Validation
list-sdk-endpoints: ## List all endpoints tagged with 'SDK'
	@echo "📋 SDK-tagged endpoints:"
	@curl -s $(OPENAPI_URL) | jq -r '.paths | to_entries[] | select(.value | to_entries[] | .value.tags[]? == "SDK") | "  \(.key) (\(.value | keys | join(", ")))"' | sort

list-all-endpoints: ## List all API endpoints
	@echo "📋 All API endpoints:"
	@curl -s $(OPENAPI_URL) | jq -r '.paths | keys[]' | sort

validate-spec: ## Validate OpenAPI spec is accessible
	@echo "🔍 Checking OpenAPI spec accessibility..."
	@curl -f -s $(OPENAPI_URL) > /dev/null && echo "✅ OpenAPI spec is accessible" || echo "❌ Cannot access OpenAPI spec"

analyze-spec: ## Analyze OpenAPI spec for SDK-tagged endpoints
	@echo "📊 Analyzing OpenAPI specification..."
	@echo ""
	@echo "📋 SDK-tagged endpoints:"
	@curl -s $(OPENAPI_URL) | jq -r '.paths | to_entries[] | select(.value | to_entries[] | .value.tags[]? == "SDK") | "  \(.key) (\(.value | keys | join(", ")))"' | sort
	@echo ""
	@echo "📋 Available tags:"
	@curl -s $(OPENAPI_URL) | jq -r '.paths | to_entries[] | .value | to_entries[] | .value.tags[]?' | sort -u | sed 's/^/  /'
	@echo ""
	@echo "📋 Total endpoints: $$(curl -s $(OPENAPI_URL) | jq '.paths | length')"
	@echo "📋 SDK endpoints: $$(curl -s $(OPENAPI_URL) | jq '.paths | to_entries[] | select(.value | to_entries[] | .value.tags[]? == "SDK") | .key' | wc -l)"

# Build and Test
build: ## Build the SDK
	@echo "🔨 Building SDK..."
	go build ./...

test: ## Run unit tests
	@echo "🧪 Running tests..."
	go test -race -v ./...

test-coverage: ## Run tests with coverage report
	@echo "🧪 Running tests with coverage..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

test-integration: ## Run integration tests (requires LEAPOCR_API_KEY)
	@echo "🧪 Running integration tests..."
	@if [ -z "$$LEAPOCR_API_KEY" ]; then \
		echo "❌ LEAPOCR_API_KEY environment variable is required"; \
		exit 1; \
	fi
	go test -race -v -tags=integration ./test/integration/...

# Code Quality
lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run ./...

format: ## Format code
	@echo "💅 Formatting code..."
	go fmt ./...
	goimports -w . 2>/dev/null || echo "goimports not available, using go fmt only"
	gofumpt -w .

# Examples
examples: ## Build all examples
	@echo "📚 Building examples..."
	@for example in examples/*/; do \
		if [ -f "$$example/main.go" ]; then \
			echo "Building example: $$example"; \
			cd "$$example" && go build . && cd ../..; \
		fi; \
	done

examples-run: ## Run all examples (requires LEAPOCR_API_KEY)
	@echo "🚀 Running examples..."
	@if [ -z "$$LEAPOCR_API_KEY" ]; then \
		echo "❌ LEAPOCR_API_KEY environment variable is required to run examples"; \
		exit 1; \
	fi
	@for example in examples/*/; do \
		if [ -f "$$example/main.go" ]; then \
			echo "Running example: $$example"; \
			cd "$$example" && timeout 30s go run . || echo "Example completed or timed out" && cd ../..; \
		fi; \
	done

# Cleanup
clean: ## Clean build artifacts and generated files
	@echo "🧹 Cleaning up..."
	go clean ./...
	rm -f coverage.out coverage.html
	rm -rf generated/ generated-sdk/ types/ gen/
	rm -f openapi.json openapi-full.json openapi-sdk.json

clean-types: ## Clean only generated types (keep other artifacts)
	@echo "🧹 Cleaning generated types..."
	rm -rf types/ generated/ generated-sdk/ gen/

# Development Workflow
dev-setup: install validate-spec generate build ## Complete development setup
	@echo "🎉 Development environment ready!"
	@echo ""
	@echo "📋 Next steps:"
	@echo "  1. Set your API key: export LEAPOCR_API_KEY=pk_live_your_key_here"
	@echo "  2. Run tests: make test"
	@echo "  3. Try examples: make examples-run"

dev-reset: clean dev-setup ## Reset development environment completely

# CI/CD targets
ci-test: lint test ## Run all CI tests
	@echo "✅ All CI tests passed!"

ci-test-full: lint test test-integration examples ## Run full CI test suite
	@echo "✅ Full CI test suite passed!"

# Release Process
release-check: ci-test examples ## Pre-release validation
	@echo "🚀 Release checks passed!"

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@mkdir -p docs/api
	godoc -html . > docs/api/index.html 2>/dev/null || echo "godoc not available"

# Advanced SDK Generation with Custom Filtering
generate-custom: ## Generate SDK with custom endpoint filtering
	@echo "🎯 Custom SDK generation..."
	@read -p "Enter tag to filter by (default: SDK): " TAG; \
	TAG=$${TAG:-SDK}; \
	echo "Filtering endpoints with tag: $$TAG"; \
	curl -s $(OPENAPI_URL) > openapi-full.json; \
	./scripts/filter-endpoints-by-tag.sh openapi-full.json openapi-custom.json "$$TAG"; \
	openapi-generator-cli generate \
		-i openapi-custom.json \
		-g go \
		-o ./generated-custom \
		--skip-validate-spec \
		--additional-properties=packageName=types; \
	mkdir -p types/; \
	cp generated-custom/model_*.go types/ 2>/dev/null || true; \
	rm -rf generated-custom openapi-full.json openapi-custom.json; \
	go fmt ./types/...; \
	echo "✅ Custom generation complete!"

# Show current SDK status
status: ## Show current SDK status
	@echo "📊 OCR Go SDK Status"
	@echo "===================="
	@echo ""
	@echo "📁 Project Structure:"
	@find . -name "*.go" -not -path "./examples/*" -not -path "./gen/*" | head -20
	@echo ""
	@echo "🏷️  Available endpoint tags:"
	@curl -s $(OPENAPI_URL) 2>/dev/null | jq -r '.paths | to_entries[] | .value | to_entries[] | .value.tags[]?' | sort -u | sed 's/^/  /' || echo "  API not accessible"
	@echo ""
	@echo "🎯 SDK endpoints (current focus):"
	@curl -s $(OPENAPI_URL) 2>/dev/null | jq -r '.paths | to_entries[] | select(.value | to_entries[] | .value.tags[]? == "SDK") | "  \(.key)"' | sort || echo "  API not accessible"
	@echo ""
	@echo "📦 Go modules:"
	@go list ./... | sed 's/^/  /'