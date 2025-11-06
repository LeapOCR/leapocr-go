# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the official Go SDK for the OCR API, providing a type-safe client for processing PDFs and extracting structured data using AI. The SDK is automatically generated from an OpenAPI 3.1 specification and includes a clean wrapper layer for developer convenience.

## Key Architecture

### OpenAPI-Generated Code

- The SDK uses `openapi-generator-cli` to generate Go client code from a live OpenAPI spec
- Generated code is placed in `ocr/` directory and includes models, API clients, and configuration
- The build system filters endpoints tagged with "SDK" to focus on public API operations

### Wrapper Architecture

- `client/` provides a clean, developer-friendly wrapper around generated code
- `ocr.go` re-exports key types and functions for simplified imports
- Custom types and options are defined in `client/client.go`

### Project Structure

```
leapocr-go/
├── ocr/                 # Generated OpenAPI client code
├── client/             # Wrapper client with simplified API
├── examples/           # Usage examples (basic, advanced)
├── test/               # Tests (integration, fixtures)
├── scripts/            # Build and generation scripts
└── ocr.go              # Main package entry point
```

## Development Commands

### SDK Generation

```bash
# Generate SDK for SDK-tagged endpoints only (recommended)
make generate

# Generate SDK for all endpoints (not recommended)
make generate-full

# Analyze available endpoints
make list-sdk-endpoints
make list-all-endpoints
make analyze-spec

# Validate OpenAPI spec accessibility
make validate-spec
```

### Build and Test

```bash
# Install dependencies
make install

# Build SDK
make build

# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests (requires LEAPOCR_API_KEY)
make test-integration

# Run code quality checks
make lint
make format

# Build and run examples
make examples
make examples-run  # requires LEAPOCR_API_KEY
```

### Development Setup

```bash
# Complete development setup
make dev-setup

# Reset development environment
make dev-reset

# Clean generated files
make clean
```

### CI/CD

```bash
# Run all CI tests
make ci-test

# Run full CI test suite
make ci-test-full

# Pre-release validation
make release-check
```

## Code Generation Workflow

The SDK follows a sophisticated generation process:

1. **Fetch OpenAPI Spec**: Downloads from `http://localhost:8080/api/v1/docs/openapi.json`
2. **Filter Endpoints**: Uses `scripts/filter-sdk-endpoints.sh` to keep only SDK-tagged endpoints
3. **Generate Code**: Uses `openapi-generator-cli` to create Go client
4. **Create Wrapper**: Runs `scripts/create-sdk-wrapper.sh` to add convenience layer
5. **Organize Files**: Moves generated files to appropriate directories

### Filtering Logic

The filtering script (`filter-sdk-endpoints.sh`) uses jq to:

- Extract endpoints with "SDK" tag
- Collect all referenced component schemas
- Create a minimal OpenAPI spec with only needed dependencies

## Client Architecture

### Main Components

- **Client**: Main client that wraps generated API client
- **OCRService**: Handles OCR-specific operations with simplified interface
- **Config**: Manages API configuration, authentication, and retry logic

### Key Types

```go
// Main client
client := ocrsdk.New("api-key")

// Processing options
result, err := client.OCR.ProcessFileFromPath(ctx, "file.pdf",
    ocrsdk.WithFormat(ocrsdk.FormatStructured),
    ocrsdk.WithTier(ocrsdk.TierCore),
)

// Response types
ProcessResult    // Initial job submission result
JobStatus       // Job status information
JobResult       // Final processing result
PageResult      // Individual page results
```

### Configuration

The client supports extensive configuration:

- Base URL customization
- HTTP client configuration
- Retry logic with exponential backoff
- Custom user agent
- Timeout settings

## Testing Strategy

### Test Structure

- **Unit Tests**: Client functionality and wrapper behavior
- **Integration Tests**: Real API calls (requires `LEAPOCR_API_KEY`)
- **Example Tests**: Ensure examples compile and run

### Running Tests

```bash
# Unit tests only
make test

# Integration tests (requires API key)
LEAPOCR_API_KEY=your-key make test-integration

# Coverage report
make test-coverage
```

## Environment Variables

### Required for Integration

```bash
LEAPOCR_API_KEY=your_api_key_here          # API authentication
OCR_BASE_URL=http://localhost:8080      # API base URL (optional)
```

### Build Dependencies

```bash
# For code generation
OPENAPI_URL=http://localhost:8080/api/v1/docs/openapi.json
GENERATOR_VERSION=7.9.0
```

## Code Standards

### Generated Code

- Generated code in `ocr/` should not be manually edited
- Linter excludes `types/` directory from most rules
- Use `make generate` to regenerate after API changes

### Wrapper Code

- Follow Go standard formatting and naming conventions
- Use functional options pattern for configuration
- Provide clear error messages with context
- All network operations accept `context.Context`

### Commit Format

Use conventional commits:

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Test additions/changes

## Important Development Notes

### API Changes

When the upstream API changes:

1. Update OpenAPI spec URL if needed
2. Run `make generate` to regenerate client code
3. Update wrapper layer if needed
4. Run tests to ensure compatibility

### Version Compatibility

- The SDK targets Go 1.21+
- Generated code compatibility depends on OpenAPI generator version
- Test with multiple Go versions if releasing publicly

### Error Handling

- Wrapper provides simplified error messages
- Generated code may return detailed API errors
- Retry logic handles transient failures automatically

## Common Development Patterns

### Adding New Features

1. Check if feature exists in OpenAPI spec
2. Regenerate SDK if needed
3. Add wrapper methods in `client/client.go`
4. Export new types in `ocr.go`
5. Add tests and examples

### Debugging Generation Issues

```bash
# Check spec accessibility
make validate-spec

# List available endpoints
make list-sdk-endpoints

# Analyze spec structure
make analyze-spec
```

### Testing New Functionality

1. Add unit tests for wrapper logic
2. Add integration tests for API calls
3. Create example in `examples/`
4. Verify with `make examples`
