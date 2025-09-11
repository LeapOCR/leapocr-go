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
    
    ocrsdk "github.com/leapocr/go-sdk"
)

func main() {
    // Initialize client with API key
    client := ocrsdk.New(os.Getenv("OCR_API_KEY"))
    
    // Process a local file
    result, err := client.OCR.ProcessFileFromPath(context.Background(), "./document.pdf",
        ocrsdk.WithFormat(ocrsdk.FormatStructured),
        ocrsdk.WithInstructions("Extract invoice details"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Wait for completion
    jobResult, err := client.OCR.WaitForCompletion(context.Background(), result.JobID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Extracted data: %+v\n", jobResult.Data)
    fmt.Printf("Credits used: %d\n", jobResult.CreditsUsed)
}
```

## Features

- ✅ Type-safe API client generated from OpenAPI 3.1 spec
- ✅ Automatic authentication with API keys
- ✅ Built-in retry logic and error handling
- ✅ Context support for all operations
- ✅ File and URL processing support
- ✅ Custom processing options and schemas
- ✅ Real-time job status monitoring
- ✅ Comprehensive test coverage
- ✅ Examples and documentation

## Client Configuration

### Basic Configuration

```go
// Simple client with API key
client := ocrsdk.New("your-api-key")

// Custom configuration
config := ocrsdk.NewConfig("your-api-key")
config.SetBaseURL("https://api.leapocr.com")
config.WithTimeout(60 * time.Second)
config.WithUserAgent("my-app/1.0.0")
config.WithRetries(5, time.Second, 2*time.Minute)

client := ocrsdk.NewWithConfig(config)
```

### Environment Variables

```bash
export OCR_API_KEY="your-api-key-here"
export OCR_BASE_URL="https://api.leapocr.com"  # optional
```

## Usage Examples

### Processing Local Files

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    ocrsdk "github.com/leapocr/go-sdk"
)

func main() {
    client := ocrsdk.New(os.Getenv("OCR_API_KEY"))
    
    // Process with options
    result, err := client.OCR.ProcessFileFromPath(context.Background(), "./invoice.pdf",
        ocrsdk.WithFormat(ocrsdk.FormatStructured),
        ocrsdk.WithSchema(map[string]interface{}{
            "invoice_number": "string",
            "total_amount":   "number",
            "date":          "date",
        }),
        ocrsdk.WithInstructions("Extract key invoice information"),
        ocrsdk.WithTier(ocrsdk.TierCore),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Wait for completion
    jobResult, err := client.OCR.WaitForCompletion(context.Background(), result.JobID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Status: %s\n", jobResult.Status)
    fmt.Printf("Data: %+v\n", jobResult.Data)
    fmt.Printf("Pages: %d\n", len(jobResult.Pages))
    fmt.Printf("Credits: %d\n", jobResult.CreditsUsed)
}
```

### Processing Files from URLs

```go
result, err := client.OCR.ProcessFileFromURL(context.Background(), 
    "https://example.com/document.pdf",
    ocrsdk.WithFormat(ocrsdk.FormatMarkdown),
    ocrsdk.WithTier(ocrsdk.TierSwift),
)
if err != nil {
    log.Fatal(err)
}

// Monitor progress
ticker := time.NewTicker(2 * time.Second)
defer ticker.Stop()

for {
    status, err := client.OCR.GetJobStatus(context.Background(), result.JobID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Progress: %.1f%%\n", status.Progress)
    
    if status.Status == "completed" {
        jobResult, _ := client.OCR.GetJobResult(context.Background(), result.JobID)
        fmt.Printf("Completed: %+v\n", jobResult.Data)
        break
    } else if status.Status == "failed" {
        log.Fatalf("Job failed: %s", status.Error)
    }
    
    <-ticker.C
}
```

### Batch Processing

```go
files := []string{"doc1.pdf", "doc2.pdf", "doc3.pdf"}
jobIDs := make([]string, 0, len(files))

// Start all jobs
for _, file := range files {
    result, err := client.OCR.ProcessFileFromPath(context.Background(), file)
    if err != nil {
        log.Printf("Failed to start %s: %v", file, err)
        continue
    }
    jobIDs = append(jobIDs, result.JobID)
}

// Wait for all completions
for _, jobID := range jobIDs {
    result, err := client.OCR.WaitForCompletion(context.Background(), jobID)
    if err != nil {
        log.Printf("Job %s failed: %v", jobID, err)
        continue
    }
    fmt.Printf("Job %s completed with %d credits\n", jobID, result.CreditsUsed)
}
```

## Processing Options

### Output Formats

```go
// Available formats
ocrsdk.FormatMarkdown            // Plain markdown text
ocrsdk.FormatStructured          // JSON structured data
ocrsdk.FormatPerPageStructured   // JSON per page
```

### Processing Tiers

```go
// Available tiers (speed vs accuracy tradeoff)
ocrsdk.TierSwift    // Fastest processing
ocrsdk.TierCore     // Balanced speed and accuracy
ocrsdk.TierIntelli  // Highest accuracy
```

### Custom Schemas

```go
// Define expected data structure
schema := map[string]interface{}{
    "invoice_number": "string",
    "vendor_name":    "string",
    "total_amount":   "number",
    "line_items": []map[string]interface{}{
        {
            "description": "string",
            "quantity":    "number",
            "unit_price":  "number",
        },
    },
}

result, err := client.OCR.ProcessFileFromPath(ctx, "invoice.pdf",
    ocrsdk.WithSchema(schema),
    ocrsdk.WithInstructions("Extract all invoice details including line items"),
)
```

## Error Handling

```go
result, err := client.OCR.ProcessFileFromPath(ctx, "document.pdf")
if err != nil {
    // Handle different error types
    switch {
    case strings.Contains(err.Error(), "authentication"):
        log.Fatal("Invalid API key")
    case strings.Contains(err.Error(), "file not found"):
        log.Fatal("Document file not found")
    case strings.Contains(err.Error(), "quota exceeded"):
        log.Fatal("Credit quota exceeded")
    default:
        log.Fatalf("Processing failed: %v", err)
    }
}

// Check job result for processing errors
jobResult, err := client.OCR.WaitForCompletion(ctx, result.JobID)
if err != nil {
    log.Fatal(err)
}

if jobResult.Status == "failed" {
    log.Fatalf("OCR processing failed: %s", jobResult.Error)
}
```

## Advanced Usage

### Custom HTTP Client

```go
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: false,
        },
    },
}

config := ocrsdk.NewConfig("your-api-key")
config.HTTPClient = httpClient
client := ocrsdk.NewWithConfig(config)
```

### Retry Configuration

```go
config := ocrsdk.NewConfig("your-api-key")
config.WithRetries(
    5,                    // max retries
    time.Second,         // initial delay
    2*time.Minute,       // max delay
)
client := ocrsdk.NewWithConfig(config)
```

## API Reference

### Core Methods

- `client.OCR.ProcessFileFromPath(ctx, path, ...options) (*ProcessResult, error)` - Process local file
- `client.OCR.ProcessFileFromURL(ctx, url, ...options) (*ProcessResult, error)` - Process file from URL
- `client.OCR.GetJobStatus(ctx, jobID) (*JobStatus, error)` - Get job status
- `client.OCR.GetJobResult(ctx, jobID) (*JobResult, error)` - Get job result
- `client.OCR.WaitForCompletion(ctx, jobID) (*JobResult, error)` - Wait for job completion

### Response Types

```go
type ProcessResult struct {
    JobID     string
    UploadURL string
    Status    string
}

type JobStatus struct {
    JobID         string
    Status        string
    Progress      float64
    EstimatedTime int
    Error         string
}

type JobResult struct {
    JobID          string
    Status         string
    Data           map[string]interface{}
    Pages          []PageResult
    ProcessingTime int
    CreditsUsed    int
    Error          string
}

type PageResult struct {
    PageNumber int                    `json:"page_number"`
    Text       string                 `json:"text"`
    Data       map[string]interface{} `json:"data"`
    Confidence float64                `json:"confidence"`
}
```

## Examples

See the [`examples/`](./examples/) directory for complete working examples:

- [`examples/basic/`](./examples/basic/) - Basic usage patterns
- [`examples/advanced/`](./examples/advanced/) - Advanced configuration and batch processing

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests (requires API key)
OCR_API_KEY=your-key go test ./test/integration/...
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](./LICENSE) for details.