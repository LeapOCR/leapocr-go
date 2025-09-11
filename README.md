# OCR Go SDK

Official Go SDK for the OCR API - Process PDFs and extract structured data using AI.

## Installation

```bash
go get github.com/leapocr/go-sdk
```

## Quick Start

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
    sdk, err := ocr.New(os.Getenv("OCR_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Process a file from URL
    job, err := sdk.ProcessURL(context.Background(), "https://example.com/document.pdf",
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
    
    fmt.Printf("Extracted data: %+v\n", result.Data)
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

### Process Local File

```go
file, err := os.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

job, err := sdk.ProcessFile(ctx, file, "document.pdf",
    ocr.WithFormat(ocr.FormatStructured),
    ocr.WithInstructions("Extract invoice data"))
if err != nil {
    log.Fatal(err)
}

result, err := sdk.WaitUntilDone(ctx, job.ID)
// Handle result...
```

### Custom Configuration

```go
config := ocr.DefaultConfig("your-api-key")
config.BaseURL = "https://api-staging.example.com"
config.Timeout = 60 * time.Second

sdk, err := ocr.NewSDK(config)
```

### Concurrent Processing

```go
var wg sync.WaitGroup
results := make(chan *ocr.OCRResult, len(urls))

for _, url := range urls {
    wg.Add(1)
    go func(fileURL string) {
        defer wg.Done()
        
        job, err := sdk.ProcessURL(ctx, fileURL, 
            ocr.WithFormat(ocr.FormatMarkdown))
        if err != nil {
            log.Printf("Failed: %v", err)
            return
        }
        
        result, err := sdk.WaitUntilDone(ctx, job.ID)
        if err != nil {
            log.Printf("Failed: %v", err)
            return
        }
        
        results <- result
    }(url)
}

go func() {
    wg.Wait()
    close(results)
}()

for result := range results {
    fmt.Printf("Completed: %+v\n", result.Data)
}
```

### Manual Status Polling

```go
job, err := sdk.ProcessURL(ctx, url, ocr.WithFormat(ocr.FormatStructured))
if err != nil {
    log.Fatal(err)
}

// Poll manually instead of using WaitUntilDone
ticker := time.NewTicker(2 * time.Second)
defer ticker.Stop()

for {
    status, err := sdk.GetJobStatus(ctx, job.ID)
    if err != nil {
        log.Printf("Failed to get status: %v", err)
        continue
    }
    
    fmt.Printf("Status: %s (%.1f%% complete)\n", status.Status, status.Progress)
    
    if status.Status == "completed" {
        result, err := sdk.GetJobResult(ctx, job.ID)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Result: %+v\n", result.Data)
        break
    }
    
    <-ticker.C
}
```

### Schema-Based Extraction

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "invoice_number": map[string]interface{}{"type": "string"},
        "total_amount":   map[string]interface{}{"type": "number"},
        "vendor_name":    map[string]interface{}{"type": "string"},
    },
}

job, err := sdk.ProcessURL(ctx, invoiceURL,
    ocr.WithFormat(ocr.FormatStructured),
    ocr.WithTier(ocr.TierIntelli),
    ocr.WithSchema(schema),
    ocr.WithInstructions("Extract invoice data according to schema"))
```

## API Reference

### SDK Methods

- `New(apiKey string) (*SDK, error)` - Create SDK with default config
- `NewSDK(config *Config) (*SDK, error)` - Create SDK with custom config
- `ProcessURL(ctx, url, ...options) (*Job, error)` - Process file from URL
- `ProcessFile(ctx, file, filename, ...options) (*Job, error)` - Process uploaded file
- `WaitUntilDone(ctx, jobID) (*OCRResult, error)` - Wait for job completion
- `WaitUntilDoneWithOptions(ctx, jobID, options) (*OCRResult, error)` - Wait with custom options
- `GetJobStatus(ctx, jobID) (*JobStatusInfo, error)` - Get current job status without waiting
- `GetJobResult(ctx, jobID) (*OCRResult, error)` - Get final job result (for completed jobs)

### Processing Options

- `WithFormat(format Format)` - Set output format (Markdown, Structured, PerPageStructured)
- `WithTier(tier Tier)` - Set processing tier (Swift, Core, Intelli)
- `WithSchema(schema map[string]interface{})` - Custom extraction schema
- `WithInstructions(instructions string)` - Custom processing instructions
- `WithCategoryID(categoryID string)` - Document category ID

### Types

- `Format`: `FormatMarkdown`, `FormatStructured`, `FormatPerPageStructured`
- `Tier`: `TierSwift`, `TierCore`, `TierIntelli`
- `OCRResult`: Final processing result with text, data, pages, credits
- `JobStatusInfo`: Job status information with progress
- `SDKError`: Custom error type with retry logic

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
make test-integration        # Integration tests (requires OCR_API_KEY)
make test-coverage          # Coverage report
```

## Examples

See the `examples/` directory for complete working examples:

- **Basic**: Simple file processing from URL and local file
- **Advanced**: Concurrent processing, custom configuration, schema extraction  
- **Validation**: Error handling and timeout scenarios

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please refer to the [API documentation](https://docs.example.com) or open an issue.