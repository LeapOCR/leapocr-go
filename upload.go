package ocr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// uploadFile uploads a file to the given presigned URL
func (s *SDK) uploadFile(ctx context.Context, presignedURL string, file io.Reader, filename string) error {
	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return NewSDKError(ErrorTypeUploadError, "failed to read file content", err)
	}

	// Validate file size
	if len(fileContent) > MaxFileSizeBytes {
		return NewSDKError(ErrorTypeValidationError,
			"file too large",
			NewValidationError("file_size",
				fmt.Sprintf("file size (%d bytes) exceeds maximum allowed size (%d bytes)",
					len(fileContent), MaxFileSizeBytes)))
	}

	if len(fileContent) == 0 {
		return NewSDKError(ErrorTypeValidationError,
			"empty file",
			NewValidationError("file_size", "file cannot be empty"))
	}

	// Create multipart form
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Add file field
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return NewSDKError(ErrorTypeUploadError, "failed to create form file", err)
	}

	if _, err := part.Write(fileContent); err != nil {
		return NewSDKError(ErrorTypeUploadError, "failed to write file content", err)
	}

	if err := writer.Close(); err != nil {
		return NewSDKError(ErrorTypeUploadError, "failed to close multipart writer", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", presignedURL, &buffer)
	if err != nil {
		return NewSDKError(ErrorTypeUploadError, "failed to create upload request", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make the upload request
	client := s.config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return NewSDKError(ErrorTypeUploadError, "failed to upload file", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return WrapHTTPError(resp, err)
	}

	return nil
}
