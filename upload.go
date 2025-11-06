package ocr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/leapocr/leapocr-go/gen"
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

// uploadFileParts uploads file parts to presigned URLs and returns completed parts with ETags
func (s *SDK) uploadFileParts(ctx context.Context, resp *gen.UploadDirectUploadResponse, file io.Reader) ([]gen.UploadCompletedPart, error) {
	if len(resp.Parts) == 0 {
		return nil, NewSDKError(ErrorTypeUploadError, "no upload parts provided", nil)
	}

	// Read entire file into memory for chunking
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, NewSDKError(ErrorTypeUploadError, "failed to read file content", err)
	}

	client := s.config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	completedParts := make([]gen.UploadCompletedPart, 0, len(resp.Parts))

	// Upload each part
	for _, part := range resp.Parts {
		if part.UploadUrl == nil || part.StartByte == nil || part.EndByte == nil || part.PartNumber == nil {
			continue
		}

		startByte := int(*part.StartByte)
		endByte := int(*part.EndByte)
		
		// Ensure we don't exceed file size
		if startByte >= len(fileContent) {
			return nil, NewSDKError(ErrorTypeUploadError, 
				fmt.Sprintf("start byte %d exceeds file size %d", startByte, len(fileContent)), nil)
		}
		
		if endByte >= len(fileContent) {
			endByte = len(fileContent) - 1
		}

		// Extract chunk data
		chunk := fileContent[startByte : endByte+1]

		// Create PUT request to upload the chunk
		req, err := http.NewRequestWithContext(ctx, "PUT", *part.UploadUrl, bytes.NewReader(chunk))
		if err != nil {
			return nil, NewSDKError(ErrorTypeUploadError, "failed to create upload request", err)
		}

		// Upload the chunk
		uploadResp, err := client.Do(req)
		if err != nil {
			return nil, NewSDKError(ErrorTypeUploadError, "failed to upload chunk", err)
		}
		defer func() { _ = uploadResp.Body.Close() }() //nolint:errcheck

		// Check response status
		if uploadResp.StatusCode < 200 || uploadResp.StatusCode >= 300 {
			return nil, NewSDKError(ErrorTypeUploadError, 
				fmt.Sprintf("upload failed with status %d", uploadResp.StatusCode), nil)
		}

		// Extract ETag from response header
		etag := uploadResp.Header.Get("ETag")
		if etag == "" {
			// Some S3-compatible services return ETag in ETag header without quotes
			// Try to get it from the response body or use a default
			etag = uploadResp.Header.Get("etag")
		}

		// Create completed part with ETag
		completedPart := gen.UploadCompletedPart{
			PartNumber: part.PartNumber,
		}
		if etag != "" {
			completedPart.Etag = &etag
		}

		completedParts = append(completedParts, completedPart)
	}

	return completedParts, nil
}

// completeDirectUpload completes the multipart upload by sending ETags
func (s *SDK) completeDirectUpload(ctx context.Context, jobID string, completedParts []gen.UploadCompletedPart) error {
	if len(completedParts) == 0 {
		return NewSDKError(ErrorTypeUploadError, "no upload parts to complete", nil)
	}

	// Create completion request
	completeRequest := gen.UploadDirectUploadCompleteRequest{
		Parts: completedParts,
	}

	// Make the API call to complete the upload
	apiRequest := s.client.SDKAPI.CompleteDirectUpload(ctx, jobID)
	apiRequest = apiRequest.UploadDirectUploadCompleteRequest(completeRequest)

	_, httpResp, err := apiRequest.Execute()
	if err != nil {
		return s.handleAPIError(err, httpResp, "failed to complete direct upload")
	}

	return nil
}
