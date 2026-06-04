package modelstrategy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// RetryConfig configures the retry behaviour for model calls.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   10 * time.Second,
	}
}

// RetryableFunc is a function that can be retried.
type RetryableFunc func() (string, error)

// RetryEngine wraps retry logic around model calls.
type RetryEngine struct {
	config RetryConfig
}

// NewRetryEngine creates a retry engine with the given config.
func NewRetryEngine(config RetryConfig) *RetryEngine {
	return &RetryEngine{config: config}
}

// Execute calls fn and retries on failure.
func (e *RetryEngine) Execute(fn RetryableFunc) (string, error) {
	var lastErr error
	delay := e.config.BaseDelay

	for attempt := 0; attempt <= e.config.MaxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err

		if attempt < e.config.MaxRetries {
			time.Sleep(delay)
			delay *= 2
			if delay > e.config.MaxDelay {
				delay = e.config.MaxDelay
			}
		}
	}
	return "", fmt.Errorf("all %d retries exhausted: %w", e.config.MaxRetries, lastErr)
}

// ──────────────────────────────────────────────
// Response validation helpers
// ──────────────────────────────────────────────

// ValidateJSON checks that raw is valid JSON.
func ValidateJSON(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("empty response")
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// ValidateFields checks that a JSON response contains the required fields.
func ValidateFields(raw string, requiredFields []string) error {
	if err := ValidateJSON(raw); err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return err
	}
	var missing []string
	for _, field := range requiredFields {
		if v, ok := doc[field]; !ok || v == nil {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// IsRetryableError checks if an error should trigger a retry.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "invalid JSON") ||
		strings.Contains(msg, "missing required fields") ||
		strings.Contains(msg, "empty response") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "internal server error")
}
