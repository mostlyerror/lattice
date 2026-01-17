package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// BaseURL is the Claude API base URL
	BaseURL = "https://api.anthropic.com/v1"

	// MessagesEndpoint is the endpoint for sending messages
	MessagesEndpoint = "/messages"

	// AnthropicVersion is the API version header value
	AnthropicVersion = "2023-06-01"

	// DefaultModel is the default Claude model to use
	DefaultModel = "claude-sonnet-4-5-20250929"

	// DefaultMaxTokens is the default max tokens for responses
	DefaultMaxTokens = 4096

	// DefaultTimeout is the default request timeout
	DefaultTimeout = 60 * time.Second
)

// Client handles Claude API interactions
type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // The message text
}

// MessageRequest represents a request to the Claude API
type MessageRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"` // Optional system prompt
	Temperature float64   `json:"temperature,omitempty"`
}

// MessageResponse represents a response from the Claude API
type MessageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ErrorResponse represents an error from the Claude API
type ErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a new Claude API client
func NewClient() (*Client, error) {
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		return nil, ErrAPIKeyMissing
	}

	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = DefaultModel
	}

	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}, nil
}

// SendMessage sends a message to Claude and returns the response
func (c *Client) SendMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	// Set default model if not specified
	if req.Model == "" {
		req.Model = c.model
	}

	// Set default max tokens if not specified
	if req.MaxTokens == 0 {
		req.MaxTokens = DefaultMaxTokens
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+MessagesEndpoint,
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", AnthropicVersion)

	// Send request with retry logic for rate limits
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = c.httpClient.Do(httpReq)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
			}
			// Wait before retry with exponential backoff
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue
		}

		// If rate limited, retry with backoff
		if resp.StatusCode == 429 {
			resp.Body.Close()
			if attempt == maxRetries-1 {
				return nil, ErrRateLimitExceeded
			}
			// Wait longer for rate limits
			time.Sleep(time.Duration(attempt+1) * 10 * time.Second)
			continue
		}

		// Success or non-retryable error
		break
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error responses
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("%w: status %d, body: %s", ErrAPIError, resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("%w: %s", ErrAPIError, errResp.Error.Message)
	}

	// Parse success response
	var msgResp MessageResponse
	if err := json.Unmarshal(body, &msgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if response is empty
	if len(msgResp.Content) == 0 || msgResp.Content[0].Text == "" {
		return nil, ErrEmptyResponse
	}

	return &msgResp, nil
}

// SendSimpleMessage is a helper to send a simple user message and get back Claude's response text
func (c *Client) SendSimpleMessage(ctx context.Context, userMessage string) (string, error) {
	req := MessageRequest{
		Model:     c.model,
		MaxTokens: DefaultMaxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	resp, err := c.SendMessage(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Content) == 0 {
		return "", ErrEmptyResponse
	}

	return resp.Content[0].Text, nil
}

// SendMessageWithSystem sends a message with a system prompt
func (c *Client) SendMessageWithSystem(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	req := MessageRequest{
		Model:     c.model,
		MaxTokens: DefaultMaxTokens,
		System:    systemPrompt,
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	resp, err := c.SendMessage(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Content) == 0 {
		return "", ErrEmptyResponse
	}

	return resp.Content[0].Text, nil
}

// ParseJSONResponse is a helper to parse JSON from Claude's response
func ParseJSONResponse(responseText string, target interface{}) error {
	// Claude might wrap JSON in markdown code blocks, so let's handle that
	text := responseText

	// Try to extract JSON from markdown code blocks if present
	if bytes.Contains([]byte(text), []byte("```json")) {
		// Find JSON block
		start := bytes.Index([]byte(text), []byte("```json"))
		if start != -1 {
			start += len("```json")
			end := bytes.Index([]byte(text[start:]), []byte("```"))
			if end != -1 {
				text = text[start : start+end]
			}
		}
	} else if bytes.Contains([]byte(text), []byte("```")) {
		// Generic code block
		start := bytes.Index([]byte(text), []byte("```"))
		if start != -1 {
			start += len("```")
			// Skip language identifier if present
			newlineIdx := bytes.Index([]byte(text[start:]), []byte("\n"))
			if newlineIdx != -1 {
				start += newlineIdx + 1
			}
			end := bytes.Index([]byte(text[start:]), []byte("```"))
			if end != -1 {
				text = text[start : start+end]
			}
		}
	}

	// Parse JSON
	if err := json.Unmarshal([]byte(text), target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}
