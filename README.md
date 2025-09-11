# Go SDK for LeapOCR

[![Go Reference](https://pkg.go.dev/badge/github.com/leapocr/go-sdk.svg)](https://pkg.go.dev/github.com/leapocr/go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/leapocr/go-sdk)](https://goreportcard.com/report/github.com/leapocr/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go SDK for the [LeapOCR API](https://www.leapocr.com/) - Process PDFs and extract structured data using AI.

## Project Status

**Version**: `v0.0.0`

This SDK is currently in **beta** and is subject to change.

## Installation

```bash
go get github.com/leapocr/go-sdk
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

    "github.com/leapocr/go-sdk"
)

func main() {
    // Initialize SDK with API key
    sdk, err := ocr.New(os.Getenv("LEAPLEAPOCR_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }

    // Process a file from URL
    job, err := sdk.ProcessURL(context.Background(), "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf",
        ocr.WithFormat(ocr.FormatStructured),
        ocr.WithTier(ocr.TierCore))
    if err != nil {
        log.Fatal(err)
    }

    // Wait for completion
    result, err := sdk.WaitUntilDone(context.Background(), job.ID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Extracted data: %+v
", result.Data)
}
```

## Features

- **Go-Native Interface**: Clean, idiomatic Go API with functional options
- **Type-Safe**: Strong typing with compile-time error checking
- **Concurrent-Friendly**: Works seamlessly with goroutines and channels
- **Flexible Configuration**: Support for custom schemas, instructions, and processing tiers
- **Robust Error Handling**: Comprehensive error types with retry logic
- **File Upload Support**: Direct file upload with presigned URLs

## Usage Examples

See the `examples/` directory for more detailed examples.

## API Reference

For a complete API reference, see the [Go documentation](https://pkg.go.dev/github.com/leapocr/go-sdk).

## Error Handling

The SDK provides comprehensive error handling:

```go
result, err := sdk.WaitUntilDone(ctx, job.ID)
if err != nil {
    if sdkErr, ok := err.(*ocr.SDKError); ok {
        fmt.Printf("SDK Error: %s
", sdkErr.Type)
        fmt.Printf("Message: %s
", sdkErr.Message)

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
