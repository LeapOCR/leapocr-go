package ocr

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
)

// WaitUntilDone waits for a job to complete with exponential backoff
func (s *SDK) WaitUntilDone(ctx context.Context, jobID string) (*OCRResult, error) {
	return s.WaitUntilDoneWithOptions(ctx, jobID, WaitOptions{})
}

// WaitOptions configures the waiting behavior
type WaitOptions struct {
	// InitialDelay is the initial delay before the first poll (default: 1 second)
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between polls (default: 30 seconds)
	MaxDelay time.Duration
	// Multiplier for exponential backoff (default: 1.5)
	Multiplier float64
	// MaxJitter adds randomness to delays (default: 1 second)
	MaxJitter time.Duration
	// MaxAttempts is the maximum number of polling attempts (default: unlimited)
	MaxAttempts int
}

// DefaultWaitOptions returns sensible defaults for waiting
func DefaultWaitOptions() WaitOptions {
	return WaitOptions{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   1.5,
		MaxJitter:    1 * time.Second,
		MaxAttempts:  0, // unlimited
	}
}

// WaitUntilDoneWithOptions waits for job completion with custom options
func (s *SDK) WaitUntilDoneWithOptions(ctx context.Context, jobID string, opts WaitOptions) (*OCRResult, error) {
	opts = applyWaitDefaults(opts)

	currentDelay := opts.InitialDelay
	attempts := 0

	for {
		if err := checkMaxAttempts(attempts, opts.MaxAttempts); err != nil {
			return nil, err
		}

		if err := checkContext(ctx); err != nil {
			return nil, err
		}

		attempts++

		result, shouldContinue, err := s.pollJobStatus(ctx, jobID)
		if err != nil {
			return nil, err
		}
		if !shouldContinue {
			return result, nil
		}

		if err := s.waitWithBackoff(ctx, currentDelay, opts.MaxJitter); err != nil {
			return nil, err
		}

		currentDelay = calculateNextDelay(currentDelay, opts.Multiplier, opts.MaxDelay)
	}
}

func applyWaitDefaults(opts WaitOptions) WaitOptions {
	if opts.InitialDelay == 0 {
		opts.InitialDelay = 1 * time.Second
	}
	if opts.MaxDelay == 0 {
		opts.MaxDelay = 30 * time.Second
	}
	if opts.Multiplier == 0 {
		opts.Multiplier = 1.5
	}
	if opts.MaxJitter == 0 {
		opts.MaxJitter = 1 * time.Second
	}
	return opts
}

func checkMaxAttempts(attempts, maxAttempts int) error {
	if maxAttempts > 0 && attempts >= maxAttempts {
		return NewSDKError(ErrorTypeTimeout, "maximum polling attempts exceeded", nil)
	}
	return nil
}

func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return NewSDKError(ErrorTypeTimeout, "context canceled while waiting for completion", ctx.Err())
	default:
		return nil
	}
}

func (s *SDK) pollJobStatus(ctx context.Context, jobID string) (*OCRResult, bool, error) {
	status, err := s.getJobStatus(ctx, jobID)
	if err != nil {
		return nil, false, err
	}

	switch status.Status {
	case "completed":
		result, err := s.getJobResult(ctx, jobID)
		return result, false, err
	case "failed", "error":
		return nil, false, NewSDKError(ErrorTypeJobError, "job failed", nil)
	case "canceled":
		return nil, false, NewSDKError(ErrorTypeJobError, "job was canceled", nil)
	}

	return nil, true, nil
}

func (s *SDK) waitWithBackoff(ctx context.Context, delay, maxJitter time.Duration) error {
	jitterMax := big.NewInt(int64(maxJitter))
	jitterRand, _ := rand.Int(rand.Reader, jitterMax) //nolint:errcheck
	jitter := time.Duration(jitterRand.Int64())
	sleepDuration := delay + jitter

	select {
	case <-ctx.Done():
		return NewSDKError(ErrorTypeTimeout, "context canceled while waiting", ctx.Err())
	case <-time.After(sleepDuration):
		return nil
	}
}

func calculateNextDelay(currentDelay time.Duration, multiplier float64, maxDelay time.Duration) time.Duration {
	nextDelay := time.Duration(float64(currentDelay) * multiplier)
	if nextDelay > maxDelay {
		return maxDelay
	}
	return nextDelay
}

// JobStatusInfo represents job status information
type JobStatusInfo struct {
	ID            string  `json:"id"`
	Status        string  `json:"status"`
	Progress      float64 `json:"progress"`
	EstimatedTime int     `json:"estimated_time"`
	Error         string  `json:"error,omitempty"`
}

// getJobStatus gets the current status of a job
func (s *SDK) getJobStatus(ctx context.Context, jobID string) (*JobStatusInfo, error) {
	// Make API call to get job status using generated client
	apiRequest := s.client.SDKAPI.GetJobStatus(ctx, jobID)

	resp, httpResp, err := apiRequest.Execute()
	if err != nil {
		return nil, s.handleAPIError(err, httpResp, "failed to get job status")
	}

	// Convert generated response to our status info
	status := &JobStatusInfo{
		ID: jobID,
	}

	// Add fields if present
	if resp.Status != nil {
		status.Status = *resp.Status
	}
	if resp.ProgressPercentage != nil {
		status.Progress = float64(*resp.ProgressPercentage)
	}
	if resp.ProcessingTime != nil {
		status.EstimatedTime = int(*resp.ProcessingTime)
	}
	if resp.ErrorMessage != nil {
		status.Error = *resp.ErrorMessage
	}

	return status, nil
}

// getJobResult gets the final result of a completed job
func (s *SDK) getJobResult(ctx context.Context, jobID string) (*OCRResult, error) {
	// Make API call to get job result using generated client
	apiRequest := s.client.SDKAPI.GetJobResult(ctx, jobID)

	resp, httpResp, err := apiRequest.Execute()
	if err != nil {
		return nil, s.handleAPIError(err, httpResp, "failed to get job result")
	}

	// Convert generated response to our result type
	result := &OCRResult{
		JobID:  jobID,
		Status: "completed",
	}

	// Extract page results (main content)
	if len(resp.Pages) > 0 {
		result.Pages = make([]PageResult, len(resp.Pages))
		var allText string
		for i, page := range resp.Pages {
			pageResult := PageResult{
				PageNumber: int(i + 1), // Default page number
			}

			// Extract page fields if available
			if page.Text != nil {
				pageResult.Text = *page.Text
				allText += *page.Text + "\n"
			}
			if page.Data != nil {
				pageResult.Data = page.Data
				if result.Data == nil {
					result.Data = make(map[string]interface{})
				}
				// Merge page data into result data
				for k, v := range page.Data {
					result.Data[k] = v
				}
			}
			if page.Confidence != nil {
				pageResult.Confidence = *page.Confidence
			}
			if page.PageNumber != nil {
				pageResult.PageNumber = int(*page.PageNumber)
			}

			result.Pages[i] = pageResult
		}
		result.Text = allText
	}

	// Extract credits and duration if available
	if resp.CreditsUsed != nil {
		result.Credits = int(*resp.CreditsUsed)
	}
	if resp.ProcessingTimeSeconds != nil {
		result.Duration = time.Duration(*resp.ProcessingTimeSeconds) * time.Second
	}

	return result, nil
}

// GetJobStatus returns the current status of a job without waiting
func (s *SDK) GetJobStatus(ctx context.Context, jobID string) (*JobStatusInfo, error) {
	return s.getJobStatus(ctx, jobID)
}

// GetJobResult returns the result of a completed job
func (s *SDK) GetJobResult(ctx context.Context, jobID string) (*OCRResult, error) {
	return s.getJobResult(ctx, jobID)
}
