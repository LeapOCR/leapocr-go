# Go SDK for LeapOCR

[![Go Reference](https://pkg.go.dev/badge/github.com/leapocr/leapocr-go.svg)](https://pkg.go.dev/github.com/leapocr/leapocr-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/leapocr/leapocr-go)](https://goreportcard.com/report/github.com/leapocr/leapocr-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go SDK for the [LeapOCR API](https://www.leapocr.com/) - Process PDFs and extract structured data using AI.

## Project Status

**Version**: `v0.0.1`

This SDK is currently in **beta** and is subject to change.

## Installation

```bash
go get github.com/leapocr/leapocr-go
```

## Getting Started

### 1. Get Your API Key

To use the LeapOCR API, you'll need an API key. You can get one by signing up on the [LeapOCR website](https://www.leapocr.com/signup).

### 2. Quick Start

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
    // Initialize SDK with API key
    sdk, err := ocr.New(os.Getenv("LEAPOCR_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }

    // Process a file from URL
    job, err := sdk.ProcessURL(context.Background(), "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf",
        ocr.WithFormat(ocr.FormatStructured),
        ocr.WithModel(ocr.ModelStandardV1))
    if err != nil {
        log.Fatal(err)
    }

    // Wait for completion
    result, err := sdk.WaitUntilDone(context.Background(), job.ID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Extracted data: %+v\n", result.Data)
}
```

## Features

- **Go-Native Interface**: Clean, idiomatic Go API with functional options
- **Type-Safe**: Strong typing with compile-time error checking
- **Concurrent-Friendly**: Works seamlessly with goroutines and channels
- **Flexible Configuration**: Support for custom schemas, instructions, and OCR models
- **Robust Error Handling**: Comprehensive error types with retry logic
- **File Upload Support**: Multipart direct uploads with automatic chunking

## Available Models

The SDK supports multiple OCR models with different quality and cost characteristics:

- **`ModelStandardV1`**: Baseline model that handles all cases (1 credit/page, Priority: 1) - **Default**
- **`ModelEnglishProV1`**: Premium quality for English documents only (2 credits/page, Priority: 4)
- **`ModelProV1`**: Highest quality model that handles all cases (5 credits/page, Priority: 5)

You can also use custom model names with `WithModelString()`. If no model is specified, `ModelStandardV1` is used by default.

## Usage Examples

### Process File from URL

```go
job, err := sdk.ProcessURL(ctx, "https://example.com/document.pdf",
    ocr.WithFormat(ocr.FormatStructured),
    ocr.WithModel(ocr.ModelStandardV1),
    ocr.WithInstructions("Extract invoice details"),
)
```

### Process Local File

```go
file, err := os.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

job, err := sdk.ProcessFile(ctx, file, "document.pdf",
    ocr.WithFormat(ocr.FormatStructured),
    ocr.WithModel(ocr.ModelProV1),
    ocr.WithSchema(map[string]interface{}{
        "amount": "number",
        "date":   "string",
    }),
)
```

### Available Formats

- **`FormatStructured`**: Structured data extraction (JSON)
- **`FormatMarkdown`**: Page-by-page OCR text output
- **`FormatPerPageStructured`**: Per-page structured extraction

See the `examples/` directory for more detailed examples.

## API Reference

For a complete API reference, see the [Go documentation](https://pkg.go.dev/github.com/leapocr/leapocr-go).

## Error Handling

The SDK provides comprehensive error handling:

```go
result, err := sdk.WaitUntilDone(ctx, job.ID)
if err != nil {
    if sdkErr, ok := err.(*ocr.SDKError); ok {
        fmt.Printf("SDK Error: %s\n", sdkErr.Type)
        fmt.Printf("Message: %s\n", sdkErr.Message)

        if sdkErr.IsRetryable() {
            // Implement retry logic
        }
    }
}
```

## Development

### Building

```bash
make build        # Build SDK
make test         # Run tests
make examples     # Build examples
```

### Code Generation

```bash
make generate     # Generate from OpenAPI spec
make clean        # Clean generated files
```

### Testing

```bash
make test                    # Unit tests
make test-coverage          # Coverage report
```

## Contributing

Contributions are welcome! Please see our [Contributing Guidelines](CONTRIBUTING.md) for more details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions, please refer to the [API documentation](https://docs.leapocr.com) or open an issue.
