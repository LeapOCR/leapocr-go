# LeapOCR Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/leapocr/leapocr-go.svg)](https://pkg.go.dev/github.com/leapocr/leapocr-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/leapocr/leapocr-go)](https://goreportcard.com/report/github.com/leapocr/leapocr-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official Go SDK for [LeapOCR](https://www.leapocr.com/) - Transform documents into structured data using AI-powered OCR.

## Overview

LeapOCR provides enterprise-grade document processing with AI-powered data extraction. This SDK offers a Go-native interface for seamless integration into your applications.

## Installation

```bash
go get github.com/leapocr/leapocr-go
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- LeapOCR API key ([sign up here](https://www.leapocr.com/signup))

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/leapocr/leapocr-go"
)

func main() {
    // Initialize the SDK with your API key
    client, err := ocr.New(os.Getenv("LEAPOCR_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Submit a document for processing
    job, err := client.ProcessURL(ctx,
        "https://example.com/document.pdf",
        ocr.WithFormat(ocr.FormatStructured),
        ocr.WithModel(ocr.ModelStandardV1),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Wait for processing to complete
    result, err := client.WaitUntilDone(ctx, job.ID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Extracted data: %+v\n", result.Data)

    // Optional: Delete the job to remove sensitive data
    if err := client.DeleteJob(ctx, job.ID); err != nil {
        log.Printf("Failed to delete job: %v", err)
    }
}
```

## Key Features

- **Idiomatic Go API** - Clean, type-safe interface following Go best practices
- **Multiple Processing Formats** - Structured data extraction, markdown output, or per-page processing
- **Flexible Model Selection** - Choose from standard, pro, or custom AI models
- **Custom Schema Support** - Define extraction schemas for your specific use case
- **Built-in Retry Logic** - Automatic handling of transient failures
- **Context Support** - Full context.Context integration for timeouts and cancellation
- **Direct File Upload** - Efficient multipart uploads for local files

## Processing Models

| Model               | Use Case                           | Credits/Page | Priority |
| ------------------- | ---------------------------------- | ------------ | -------- |
| `ModelStandardV1`   | General purpose (default)          | 1            | 1        |
| `ModelEnglishProV1` | English documents, premium quality | 2            | 4        |
| `ModelProV1`        | Highest quality, all languages     | 5            | 5        |

Use `WithModel()` to specify a model, or `WithModelString()` for custom models. Defaults to `ModelStandardV1`.

## Usage Examples

### Processing from URL

```go
ctx := context.Background()

job, err := client.ProcessURL(ctx,
    "https://example.com/invoice.pdf",
    ocr.WithFormat(ocr.FormatStructured),
    ocr.WithModel(ocr.ModelStandardV1),
    ocr.WithInstructions("Extract invoice number, date, and total amount"),
)
if err != nil {
    log.Fatal(err)
}

result, err := client.WaitUntilDone(ctx, job.ID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Processing completed in %v\n", result.Duration)
fmt.Printf("Credits used: %d\n", result.Credits)
fmt.Printf("Data: %+v\n", result.Data)
```

### Processing Local Files

```go
file, err := os.Open("invoice.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

job, err := client.ProcessFile(ctx, file, "invoice.pdf",
    ocr.WithFormat(ocr.FormatStructured),
    ocr.WithModel(ocr.ModelProV1),
    ocr.WithSchema(map[string]interface{}{
        "invoice_number": "string",
        "total_amount":   "number",
        "invoice_date":   "string",
        "vendor_name":    "string",
    }),
)
```

### Using Templates

Use pre-configured templates for common document types:

```go
// Use an existing template by slug
job, err := client.ProcessFile(ctx, file, "invoice.pdf",
    ocr.WithFormat(ocr.FormatStructured),
    ocr.WithTemplateSlug("invoice-template"),
)
```

### Custom Schema Extraction

Define custom extraction schemas for specific use cases:

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "patient_name": map[string]string{"type": "string"},
        "date_of_birth": map[string]string{"type": "string"},
        "medications": map[string]interface{}{
            "type": "array",
            "items": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "name": map[string]string{"type": "string"},
                    "dosage": map[string]string{"type": "string"},
                },
            },
        },
    },
}

job, err := client.ProcessFile(ctx, file, "medical-record.pdf",
    ocr.WithFormat(ocr.FormatStructured),
    ocr.WithSchema(schema),
)
```

### Output Formats

| Format                    | Description        | Use Case                                       |
| ------------------------- | ------------------ | ---------------------------------------------- |
| `FormatStructured`        | Single JSON object | Extract specific fields across entire document |
| `FormatMarkdown`          | Text per page      | Convert document to readable text              |
| `FormatPerPageStructured` | JSON per page      | Extract fields from multi-section documents    |

### Monitoring Job Progress

```go
// Poll for status updates
ticker := time.NewTicker(2 * time.Second)
defer ticker.Stop()

for {
    status, err := client.GetJobStatus(ctx, job.ID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Status: %s (%.1f%% complete)\n", status.Status, status.Progress)

    if status.Status == "completed" {
        result, _ := client.GetJobResult(ctx, job.ID)
        fmt.Println("Processing complete!")
        break
    }

    <-ticker.C
}
```

### Deleting Jobs

Delete jobs to remove sensitive data and free up storage. Jobs and their associated files are automatically deleted after 7 days, but you can delete them immediately after processing:

```go
// Process and delete immediately after retrieving results
result, err := client.WaitUntilDone(ctx, job.ID)
if err != nil {
    log.Fatal(err)
}

// Use the result
fmt.Printf("Extracted data: %+v\n", result.Data)

// Delete the job (redacts content and marks as deleted)
err = client.DeleteJob(ctx, job.ID)
if err != nil {
    log.Printf("Failed to delete job: %v", err)
}
```

For more examples, see the [`examples/`](./examples) directory.

## Configuration

### Custom Configuration

```go
config := &ocr.Config{
    APIKey:     "your-api-key",
    BaseURL:    "https://api.leapocr.com",
    HTTPClient: &http.Client{Timeout: 60 * time.Second},
    UserAgent:  "my-app/1.0",
    Timeout:    30 * time.Second,
}

client, err := ocr.NewSDK(config)
```

### Environment Variables

```bash
export LEAPOCR_API_KEY="your-api-key"
export OCR_BASE_URL="https://api.leapocr.com"  # optional
```

## Error Handling

The SDK provides typed errors for robust error handling:

```go
result, err := client.WaitUntilDone(ctx, job.ID)
if err != nil {
    if sdkErr, ok := err.(*ocr.SDKError); ok {
        switch sdkErr.Type {
        case ocr.ErrorTypeAuth:
            log.Fatal("Authentication failed - check your API key")
        case ocr.ErrorTypeValidation:
            log.Printf("Validation error: %s", sdkErr.Message)
        case ocr.ErrorTypeNetwork:
            if sdkErr.IsRetryable() {
                // Retry the operation
            }
        case ocr.ErrorTypeProcessing:
            log.Printf("Processing failed: %s", sdkErr.Message)
        }
    }
}
```

### Error Types

- `ErrorTypeInvalidConfig` - Configuration errors
- `ErrorTypeAuth` - Authentication failures
- `ErrorTypeValidation` - Input validation errors
- `ErrorTypeNetwork` - Network/connectivity issues (retryable)
- `ErrorTypeProcessing` - Document processing errors
- `ErrorTypeTimeout` - Operation timeouts

## API Reference

Full API documentation is available at [pkg.go.dev/github.com/leapocr/leapocr-go](https://pkg.go.dev/github.com/leapocr/leapocr-go).

### Core Methods

```go
// Initialize SDK
New(apiKey string) (*SDK, error)
NewSDK(config *Config) (*SDK, error)

// Process documents
ProcessURL(ctx context.Context, url string, opts ...ProcessingOption) (*Job, error)
ProcessFile(ctx context.Context, file io.Reader, filename string, opts ...ProcessingOption) (*Job, error)

// Job management
GetJobStatus(ctx context.Context, jobID string) (*JobStatus, error)
GetJobResult(ctx context.Context, jobID string) (*OCRResult, error)
WaitUntilDone(ctx context.Context, jobID string) (*OCRResult, error)
DeleteJob(ctx context.Context, jobID string) error
```

### Processing Options

```go
WithFormat(format Format)                 // Set output format
WithModel(model Model)                    // Set OCR model
WithModelString(model string)             // Set custom model
WithSchema(schema map[string]interface{}) // Define extraction schema
WithInstructions(instructions string)     // Add processing instructions
WithTemplateSlug(templateSlug string)     // Use existing template
```

## Development

### Prerequisites

- Go 1.21+
- [golangci-lint](https://golangci-lint.run/) (for linting)
- [OpenAPI Generator](https://openapi-generator.tech/) (for code generation)

### Setup

```bash
# Clone the repository
git clone https://github.com/leapocr/leapocr-go.git
cd leapocr-go

# Install dependencies
make install

# Run tests
make test
```

### Common Tasks

```bash
make build              # Build the SDK
make test               # Run unit tests
make test-coverage      # Generate coverage report
make test-integration   # Run integration tests (requires API key)
make lint               # Run linters
make format             # Format code
make examples           # Build examples
```

### Code Generation

The SDK is partially generated from the OpenAPI specification:

```bash
make generate           # Regenerate client from OpenAPI spec
make clean              # Remove generated files
```

## Contributing

We welcome contributions! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support & Resources

- **Documentation**: [docs.leapocr.com](https://docs.leapocr.com)
- **API Reference**: [pkg.go.dev/github.com/leapocr/leapocr-go](https://pkg.go.dev/github.com/leapocr/leapocr-go)
- **Issues**: [GitHub Issues](https://github.com/leapocr/leapocr-go/issues)
- **Website**: [leapocr.com](https://www.leapocr.com)

---

**Version**: 0.0.3
