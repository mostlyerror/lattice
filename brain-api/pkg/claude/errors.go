package claude

import "errors"

var (
	// ErrAPIKeyMissing is returned when CLAUDE_API_KEY is not set
	ErrAPIKeyMissing = errors.New("CLAUDE_API_KEY environment variable is not set")

	// ErrInvalidRequest is returned when the request is malformed
	ErrInvalidRequest = errors.New("invalid request to Claude API")

	// ErrRateLimitExceeded is returned when rate limit is hit
	ErrRateLimitExceeded = errors.New("Claude API rate limit exceeded")

	// ErrAPIError is returned for general API errors
	ErrAPIError = errors.New("Claude API error")

	// ErrTimeout is returned when request times out
	ErrTimeout = errors.New("Claude API request timeout")

	// ErrEmptyResponse is returned when Claude returns empty content
	ErrEmptyResponse = errors.New("Claude returned empty response")
)
