package ocr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/leapocr/leapocr-go/internal/generated"
)

// uploadFileParts uploads file parts to presigned URLs and returns completed parts with ETags
func (s *SDK) uploadFileParts(ctx context.Context, resp *generated.UploadDirectUploadResponse, file io.Reader) ([]generated.UploadCompletedPart, error) {
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

	completedParts := make([]generated.UploadCompletedPart, 0, len(resp.Parts))

	// Upload each part
	for _, part := range resp.Parts {
		completedPart, err := s.uploadSinglePart(ctx, client, part, fileContent)
		if err != nil {
			return nil, err
		}
		completedParts = append(completedParts, completedPart)
	}

	return completedParts, nil
}

// uploadSinglePart uploads a single file part to a presigned URL
func (s *SDK) uploadSinglePart(ctx context.Context, client *http.Client, part generated.UploadMultipartPart, fileContent []byte) (generated.UploadCompletedPart, error) {
	if part.UploadUrl == nil || part.StartByte == nil || part.EndByte == nil || part.PartNumber == nil {
		return generated.UploadCompletedPart{}, NewSDKError(ErrorTypeUploadError, "invalid part configuration", nil)
	}

	startByte := int(*part.StartByte)
	endByte := int(*part.EndByte)

	// Ensure we don't exceed file size
	if startByte >= len(fileContent) {
		return generated.UploadCompletedPart{}, NewSDKError(ErrorTypeUploadError,
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
		return generated.UploadCompletedPart{}, NewSDKError(ErrorTypeUploadError, "failed to create upload request", err)
	}

	// Upload the chunk
	uploadResp, err := client.Do(req)
	if err != nil {
		return generated.UploadCompletedPart{}, NewSDKError(ErrorTypeUploadError, "failed to upload chunk", err)
	}
	defer func() { _ = uploadResp.Body.Close() }() //nolint:errcheck

	// Check response status
	if uploadResp.StatusCode < 200 || uploadResp.StatusCode >= 300 {
		return generated.UploadCompletedPart{}, NewSDKError(ErrorTypeUploadError,
			fmt.Sprintf("upload failed with status %d", uploadResp.StatusCode), nil)
	}

	// Extract ETag from response header
	etag := uploadResp.Header.Get("ETag")
	if etag == "" {
		// Try lowercase header name as fallback
		etag = uploadResp.Header.Get("etag")
	}
	// Remove quotes if present (S3-compatible services return quoted ETags)
	etag = strings.Trim(etag, `"`)

	// Create completed part with ETag
	var partNumber int32
	if part.PartNumber != nil {
		partNumber = *part.PartNumber
	}
	completedPart := generated.UploadCompletedPart{
		PartNumber: partNumber,
		Etag:       etag,
	}

	return completedPart, nil
}

// completeDirectUpload completes the multipart upload by sending ETags
func (s *SDK) completeDirectUpload(ctx context.Context, jobID string, completedParts []generated.UploadCompletedPart) error {
	if len(completedParts) == 0 {
		return NewSDKError(ErrorTypeUploadError, "no upload parts to complete", nil)
	}

	// Create completion request
	completePayload := generated.UploadDirectUploadCompleteRequest{
		Parts: completedParts,
	}

	// Make the API call to complete the upload
	apiRequest := s.client.SDKAPI.CompleteDirectUpload(ctx, jobID)
	apiRequest = apiRequest.CompleteDirectUploadRequest(
		generated.UploadDirectUploadCompleteRequestAsCompleteDirectUploadRequest(&completePayload),
	)

	_, httpResp, err := apiRequest.Execute()
	if err != nil {
		return s.handleAPIError(err, httpResp, "failed to complete direct upload")
	}

	return nil
}
