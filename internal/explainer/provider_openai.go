package explainer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIProvider implements the AIProvider interface for OpenAI
type OpenAIProvider struct {
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
	timeout     time.Duration
	baseURL     string
	client      *http.Client
}

// OpenAIRequest represents the request to OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

// OpenAIMessage represents a message in the conversation
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *ExplainerConfig) AIProvider {
	return &OpenAIProvider{
		apiKey:      config.APIKey,
		model:       config.Model,
		temperature: config.Temperature,
		maxTokens:   config.MaxTokens,
		timeout:     config.Timeout,
		baseURL:     "https://api.openai.com/v1",
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// GenerateExplanation generates an explanation using OpenAI
func (p *OpenAIProvider) GenerateExplanation(ctx context.Context, prompt string) (string, error) {
	// Prepare request
	reqBody := OpenAIRequest{
		Model:       p.model,
		Temperature: p.temperature,
		MaxTokens:   p.maxTokens,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: getSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", &ExplanationError{
			Code:    ErrCodeAPIError,
			Message: "failed to marshal request",
			Cause:   err,
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", &ExplanationError{
			Code:    ErrCodeAPIError,
			Message: "failed to create request",
			Cause:   err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", p.handleHTTPError(err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &ExplanationError{
			Code:    ErrCodeAPIError,
			Message: "failed to read response",
			Cause:   err,
		}
	}

	// Parse response
	var apiResp OpenAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", &ExplanationError{
			Code:    ErrCodeParseError,
			Message: "failed to parse response",
			Cause:   err,
		}
	}

	// Check for API errors
	if apiResp.Error != nil {
		return "", &ExplanationError{
			Code:    p.mapErrorCode(apiResp.Error.Type),
			Message: apiResp.Error.Message,
		}
	}

	// Extract content
	if len(apiResp.Choices) == 0 {
		return "", &ExplanationError{
			Code:    ErrCodeAPIError,
			Message: "no choices in response",
		}
	}

	return apiResp.Choices[0].Message.Content, nil
}

// IsAvailable checks if the provider is available
func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetTokenCount estimates token count for a prompt
func (p *OpenAIProvider) GetTokenCount(text string) int {
	// Rough estimation: ~4 characters per token
	// For production, use tiktoken library
	return len(text) / 4
}

// handleHTTPError converts HTTP errors to ExplanationError
func (p *OpenAIProvider) handleHTTPError(err error) error {
	if err == context.DeadlineExceeded {
		return &ExplanationError{
			Code:    ErrCodeTimeout,
			Message: "request timeout",
			Cause:   err,
		}
	}

	return &ExplanationError{
		Code:    ErrCodeAPIError,
		Message: "HTTP request failed",
		Cause:   err,
	}
}

// mapErrorCode maps OpenAI error types to our error codes
func (p *OpenAIProvider) mapErrorCode(errorType string) string {
	switch errorType {
	case "invalid_request_error":
		return ErrCodeInvalidRequest
	case "authentication_error":
		return ErrCodeAuthError
	case "rate_limit_error":
		return ErrCodeRateLimit
	case "server_error", "service_unavailable":
		return ErrCodeAPIError
	default:
		return ErrCodeAPIError
	}
}

// getSystemPrompt returns the system prompt
func getSystemPrompt() string {
	return `You are a security expert assistant helping developers understand and fix vulnerabilities.

Your role:
- Explain security issues in simple, clear language
- Provide actionable remediation steps
- Include real-world context and examples
- Be concise but thorough
- Focus on practical solutions

Output format:
- Always respond with valid JSON
- Use the exact schema provided
- Be specific and actionable
- Avoid jargon unless necessary

Tone:
- Professional but approachable
- Educational, not condescending
- Focus on helping, not blaming`
}

// RetryableProvider wraps a provider with retry logic
type RetryableProvider struct {
	provider      AIProvider
	maxRetries    int
	retryDelay    time.Duration
	backoffFactor float64
}

// NewRetryableProvider creates a provider with retry logic
func NewRetryableProvider(provider AIProvider, maxRetries int, retryDelay time.Duration) AIProvider {
	return &RetryableProvider{
		provider:      provider,
		maxRetries:    maxRetries,
		retryDelay:    retryDelay,
		backoffFactor: 2.0,
	}
}

// Name returns the provider name
func (r *RetryableProvider) Name() string {
	return r.provider.Name()
}

// GenerateExplanation generates an explanation with retry logic
func (r *RetryableProvider) GenerateExplanation(ctx context.Context, prompt string) (string, error) {
	var lastErr error
	delay := r.retryDelay

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return "", ctx.Err()
			}
			delay = time.Duration(float64(delay) * r.backoffFactor)
		}

		result, err := r.provider.GenerateExplanation(ctx, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if explErr, ok := err.(*ExplanationError); ok {
			switch explErr.Code {
			case ErrCodeAuthError, ErrCodeInvalidRequest:
				// Don't retry these errors
				return "", err
			case ErrCodeRateLimit, ErrCodeTimeout:
				// Retry these errors
				continue
			}
		}

		// For other errors, retry
		continue
	}

	return "", &ExplanationError{
		Code:    ErrCodeAPIError,
		Message: fmt.Sprintf("failed after %d retries", r.maxRetries),
		Cause:   lastErr,
	}
}

// IsAvailable checks if the provider is available
func (r *RetryableProvider) IsAvailable() bool {
	return r.provider.IsAvailable()
}

// GetTokenCount estimates token count
func (r *RetryableProvider) GetTokenCount(text string) int {
	return r.provider.GetTokenCount(text)
}

// Made with Bob
