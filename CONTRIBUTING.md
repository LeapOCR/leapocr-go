# Contributing to OCR Go SDK

Thank you for your interest in contributing to the OCR Go SDK! This guide will help you get started.

## Development Setup

1. **Prerequisites**

   - Go 1.21 or higher
   - Make
   - Git

2. **Clone the repository**

   ```bash
   git clone https://github.com/leapOCR/go-sdk.git
   cd ocr-go-sdk
   ```

3. **Set up development environment**

   ```bash
   make dev-setup
   ```

4. **Run tests**
   ```bash
   make test
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**

   - Follow Go conventions and best practices
   - Add tests for new functionality
   - Update documentation as needed

3. **Run quality checks**

   ```bash
   make lint
   make test
   make format
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

### Code Generation

The SDK uses code generation from the OpenAPI specification:

```bash
# Regenerate client code when the API changes
make generate

# Validate the OpenAPI spec is accessible
make validate-spec
```

### Testing

We use multiple levels of testing:

1. **Unit tests** - Test individual components in isolation

   ```bash
   make test
   ```

2. **Integration tests** - Test against a running API instance

   ```bash
   # Set up environment variables
   export OCR_API_KEY=your_test_api_key
   export OCR_BASE_URL=http://localhost:8080

   # Run integration tests
   make test-integration
   ```

3. **Example tests** - Ensure examples compile and run
   ```bash
   make examples
   ```

### Running Examples

Examples require environment setup:

```bash
export OCR_API_KEY=your_api_key
cd examples/basic
go run main.go
```

## Code Standards

### Go Code Style

- Follow standard Go formatting (`gofmt`, `goimports`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Handle errors appropriately

### Commit Messages

We use conventional commits format:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Test additions or modifications
- `chore:` - Build process or auxiliary tool changes

Examples:

- `feat: add batch processing support`
- `fix: handle timeout errors gracefully`
- `docs: update authentication examples`

### API Design Principles

1. **Consistency** - Follow established patterns in the codebase
2. **Simplicity** - Make common tasks easy, complex tasks possible
3. **Type Safety** - Use Go's type system effectively
4. **Error Handling** - Provide clear, actionable error messages
5. **Context Support** - All network operations should accept `context.Context`

## Project Structure

```
ocr-go-sdk/
├── client/          # Core HTTP client and configuration
├── ocr/             # OCR-specific operations
├── analytics/       # Analytics service client
├── auth/           # Authentication service client
├── credits/        # Credits management client
├── examples/       # Usage examples
├── test/           # Test utilities and integration tests
├── types/          # Generated and custom types
└── internal/       # Internal utilities (not exported)
```

## Adding New Features

### Adding a New Service Client

1. Create a new directory under the root (e.g., `templates/`)
2. Implement the client with the standard interface:

   ```go
   type Client struct {
       client ClientInterface
   }

   type ClientInterface interface {
       DoRequestWithResponse(ctx context.Context, method, path string, body interface{}, response interface{}) error
   }
   ```

3. Add the client to the main client in `client/client.go`
4. Add tests in `service_name/client_test.go`
5. Add integration tests if needed

### Adding New Types

- Generated types go in `types/` directory
- Custom types can be added to service-specific directories
- Ensure all exported types have documentation

## Testing Guidelines

### Unit Tests

- Use `testify` for assertions
- Mock external dependencies
- Test both success and error cases
- Use table-driven tests when appropriate

```go
func TestClientFunction(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expectedErr bool
    }{
        {
            name:        "valid input",
            input:       "valid",
            expectedErr: false,
        },
        {
            name:        "invalid input",
            input:       "",
            expectedErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

- Use build tags: `//go:build integration`
- Require environment configuration
- Clean up resources after tests
- Skip gracefully when dependencies aren't available

## Documentation

### Code Documentation

- All exported functions, types, and constants must have comments
- Comments should start with the name being documented
- Use examples in documentation when helpful

```go
// ProcessFile processes a PDF file and returns a job ID.
// The file can be provided as either a file path or an io.Reader.
//
// Example:
//   job, err := client.ProcessFile(ctx, &ProcessFileRequest{
//       FilePath: "./document.pdf",
//       Mode:     ocr.ModeTextAndImage,
//   })
func (c *Client) ProcessFile(ctx context.Context, req *ProcessFileRequest) (*JobResponse, error) {
    // Implementation
}
```

### README Updates

When adding new features, update the relevant documentation:

- Main README.md
- Example code
- API reference (if applicable)

## Pull Request Process

1. **Create a pull request**

   - Use a descriptive title
   - Fill out the PR template
   - Link any related issues

2. **Address review feedback**

   - Respond to all comments
   - Make requested changes
   - Push updates to your branch

3. **Ensure CI passes**

   - All tests must pass
   - Code must pass linting
   - Examples must compile

4. **Merge requirements**
   - At least one approval from a maintainer
   - All CI checks passing
   - Up to date with main branch

## Release Process

Releases are automated when tags are pushed:

1. **Version bump**

   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

2. **Automatic processes**
   - CI/CD runs full test suite
   - GitHub release is created
   - Documentation is updated

## Getting Help

- **Issues** - Create a GitHub issue for bugs or feature requests
- **Discussions** - Use GitHub discussions for questions
- **Documentation** - Check existing documentation and examples

## Code of Conduct

This project follows the [Go Community Code of Conduct](https://golang.org/conduct). Please be respectful and inclusive in all interactions.

Thank you for contributing to the OCR Go SDK!
