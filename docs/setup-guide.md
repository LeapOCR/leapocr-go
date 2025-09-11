# OCR Go SDK Setup Guide

This guide walks you through setting up and using the OCR Go SDK with your OCR API.

## Prerequisites

- Go 1.21 or higher
- OCR API server running (default: http://localhost:8080)
- Valid API key

## Installation

```bash
go get github.com/your-org/ocr-go-sdk
```

## Quick Start

### 1. Get Your API Key

Obtain your API key from the OCR dashboard or create one programmatically:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/your-org/ocr-go-sdk"
)

func main() {
    // Initialize with existing API key
    client := ocr.New("pk_live_your_api_key_here")

    // Or create one programmatically (requires authentication)
    // apiKey, err := client.Auth.CreateAPIKey(ctx, &CreateAPIKeyRequest{
    //     Name: "My SDK Key",
    //     Permissions: []string{"read", "write"},
    // })
}
```

### 2. Configure the Client

```go
// Basic configuration
client := ocr.New("your-api-key")

// Advanced configuration
config := ocr.NewConfig("your-api-key")
config.SetBaseURL("https://api.example.com")
config.WithTimeout(60 * time.Second)
config.WithRetries(5, time.Second, 2*time.Minute)
config.WithUserAgent("my-app/1.0.0")

client := ocr.NewWithConfig(config)
```

### 3. Process Your First Document

#### From File Path

```go
ctx := context.Background()

req := &ocr.ProcessFileRequest{
    FilePath: "./invoice.pdf",
    Mode:     ocr.ModeTextAndImage,
    Schema:   "invoice", // optional predefined schema
    CustomInstructions: "Extract invoice number, date, and total amount",
    Priority: ocr.PriorityNormal,
}

// Start processing
job, err := client.OCR.ProcessFile(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Job started: %s\n", job.ID)

// Wait for completion
result, err := client.OCR.WaitForCompletion(ctx, job.ID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Credits used: %d\n", result.CreditsCost)
fmt.Printf("Extracted data: %+v\n", result.Data)
```

#### From URL

```go
job, err := client.OCR.ProcessURL(ctx, "https://example.com/document.pdf", &ocr.ProcessFileRequest{
    Mode: ocr.ModeTextOnly,
    Priority: ocr.PriorityHigh,
})
```

#### From io.Reader

```go
file, err := os.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

req := &ocr.ProcessFileRequest{
    FileReader: file,
    FileName:   "document.pdf",
    Mode:       ocr.ModeTextAndImage,
}

job, err := client.OCR.ProcessFile(ctx, req)
```

### 4. Monitor Progress

```go
// Poll for status updates
ticker := time.NewTicker(2 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-ticker.C:
        status, err := client.OCR.GetStatus(ctx, job.ID)
        if err != nil {
            return err
        }

        fmt.Printf("Status: %s", status.Status)
        if status.Progress != nil {
            fmt.Printf(" (%.1f%% complete)", status.Progress.Percentage)
        }
        fmt.Println()

        if status.Status == ocr.StatusCompleted {
            return nil
        } else if status.Status == ocr.StatusFailed {
            return fmt.Errorf("job failed: %s", *status.Error)
        }
    }
}
```

## Advanced Features

### Batch Processing

```go
files := []string{"doc1.pdf", "doc2.pdf", "doc3.pdf"}
jobIDs := make([]string, 0, len(files))

// Start all jobs
for _, file := range files {
    req := &ocr.ProcessFileRequest{
        FilePath: file,
        Mode:     ocr.ModeTextAndImage,
        Priority: ocr.PriorityNormal,
    }

    job, err := client.OCR.ProcessFile(ctx, req)
    if err != nil {
        log.Printf("Failed to start %s: %v", file, err)
        continue
    }

    jobIDs = append(jobIDs, job.ID)
}

// Wait for all to complete
for _, jobID := range jobIDs {
    result, err := client.OCR.WaitForCompletion(ctx, jobID)
    if err != nil {
        log.Printf("Job %s failed: %v", jobID, err)
        continue
    }

    fmt.Printf("Job %s completed - Credits: %d\n", jobID, result.CreditsCost)
}
```

### Analytics

```go
// Get credits analytics
creditsData, err := client.Analytics.GetCreditsAnalytics(ctx)
if err != nil {
    log.Fatal(err)
}

// Get jobs analytics
jobsData, err := client.Analytics.GetJobsAnalytics(ctx)
if err != nil {
    log.Fatal(err)
}

// List recent jobs
jobs, err := client.Jobs.List(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Error Handling

```go
result, err := client.OCR.GetResult(ctx, jobID)
if err != nil {
    // Check if it's an API error
    if client.IsAPIError(err, 404) {
        fmt.Println("Job not found")
    } else if client.IsAPIError(err, 429) {
        fmt.Println("Rate limit exceeded")
    } else {
        fmt.Printf("Unexpected error: %v", err)
    }
    return
}
```

## Environment Variables

Set these environment variables for easier configuration:

```bash
export OCR_API_KEY=pk_live_your_api_key_here
export OCR_BASE_URL=https://api.example.com  # optional
```

Then use in code:

```go
apiKey := os.Getenv("OCR_API_KEY")
if apiKey == "" {
    log.Fatal("OCR_API_KEY environment variable is required")
}

client := ocr.New(apiKey)

// Optionally override base URL
if baseURL := os.Getenv("OCR_BASE_URL"); baseURL != "" {
    config := ocr.NewConfig(apiKey)
    config.SetBaseURL(baseURL)
    client = ocr.NewWithConfig(config)
}
```

## Processing Modes

- `ocr.ModeTextOnly` - Extract text only (fastest, lowest cost)
- `ocr.ModeImageOnly` - Extract images only
- `ocr.ModeTextAndImage` - Extract both text and images (default)
- `ocr.ModeAutoDetect` - Automatically determine best mode

## Job Priorities

- `ocr.PriorityLow` - Process when resources available
- `ocr.PriorityNormal` - Standard processing (default)
- `ocr.PriorityHigh` - Expedited processing (higher cost)

## Best Practices

1. **Always use context with timeouts**

   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
   defer cancel()
   ```

2. **Handle errors gracefully**

   ```go
   if err != nil {
       log.Printf("Operation failed: %v", err)
       // Don't panic in production code
   }
   ```

3. **Reuse clients**

   ```go
   // Good: Create once, reuse
   client := ocr.New(apiKey)

   // Bad: Create new client for each request
   ```

4. **Monitor credit usage**

   ```go
   result, err := client.OCR.WaitForCompletion(ctx, jobID)
   if err == nil {
       fmt.Printf("Credits used: %d\n", result.CreditsCost)
   }
   ```

5. **Use appropriate processing modes**

   ```go
   // For text extraction only
   req.Mode = ocr.ModeTextOnly

   // For documents with important images
   req.Mode = ocr.ModeTextAndImage
   ```

## Troubleshooting

### Common Issues

1. **Authentication errors**

   - Verify API key is correct
   - Check key permissions
   - Ensure key hasn't expired

2. **Connection timeouts**

   - Increase timeout duration
   - Check network connectivity
   - Verify base URL is correct

3. **Processing failures**

   - Check file format is supported
   - Verify file isn't corrupted
   - Check file size limits

4. **Rate limiting**
   - Implement exponential backoff
   - Consider reducing concurrency
   - Contact support for rate limit increases

### Debug Mode

Enable debug logging by setting log level:

```go
import "log"

// Enable detailed logging
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

## Support

- **Documentation**: [API Reference](../README.md)
- **Examples**: [Examples Directory](../examples/)
- **Issues**: [GitHub Issues](https://github.com/leapOCR/go-sdk/issues)
- **Discussions**: [GitHub Discussions](https://github.com/leapOCR/go-sdk/discussions)
