package ocr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DeleteJob soft deletes an OCR job by redacting all page content to [REDACTED],
// deleting associated files from storage, and marking the job as deleted.
// The job will no longer be accessible via normal fetch endpoints but will
// appear in job listings with a deleted flag.
func (s *SDK) DeleteJob(ctx context.Context, jobID string) error {
	if jobID == "" {
		return NewSDKError(ErrorTypeValidationError, "job ID is required", nil)
	}

	baseURL := strings.TrimRight(s.config.BaseURL, "/")
	url := fmt.Sprintf("%s/ocr/delete/%s", baseURL, jobID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return NewSDKError(ErrorTypeAPIError, "failed to build delete request", err)
	}

	req.Header.Set("X-API-KEY", s.config.APIKey)
	if s.config.UserAgent != "" {
		req.Header.Set("User-Agent", s.config.UserAgent)
	}

	httpClient := s.client.GetConfig().HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return NewSDKError(ErrorTypeAPIError, "failed to delete job", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log or handle close error if needed
			_ = closeErr
		}
	}()

	if resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewSDKError(ErrorTypeAPIError, fmt.Sprintf("failed to delete job: %s (failed to read response body)", resp.Status), err)
		}
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return NewSDKError(ErrorTypeAPIError, fmt.Sprintf("failed to delete job: %s", message), nil)
	}

	return nil
}
