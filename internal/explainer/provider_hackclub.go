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

// HackClubProvider implements the AIProvider interface for Hack Club AI
type HackClubProvider struct {
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
	timeout     time.Duration
	baseURL     string
	client      *http.Client
}

// HackClubRequest represents the request to Hack Club AI API
type HackClubRequest struct {
	Model       string            `json:"model"`
	Messages    []HackClubMessage `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
}

// HackClubMessage represents a message in the conversation
type HackClubMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// HackClubResponse represents the response from Hack Club AI API
type HackClubResponse struct {
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

// NewHackClubProvider creates a new Hack Club AI provider
func NewHackClubProvider(config *ExplainerConfig) AIProvider {
	return &HackClubProvider{
		apiKey:      config.APIKey,
		model:       config.Model,
		temperature: config.Temperature,
		maxTokens:   config.MaxTokens,
		timeout:     config.Timeout,
		baseURL:     "https://ai.hackclub.com/proxy/v1",
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name returns the provider name
func (p *HackClubProvider) Name() string {
	return "hackclub"
}

// GenerateExplanation generates an explanation using Hack Club AI
func (p *HackClubProvider) GenerateExplanation(ctx context.Context, prompt string) (string, error) {
	// Prepare request
	reqBody := HackClubRequest{
		Model:       p.model,
		Temperature: p.temperature,
		MaxTokens:   p.maxTokens,
		Messages: []HackClubMessage{
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

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return "", &ExplanationError{
			Code:    ErrCodeAPIError,
			Message: fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Parse response
	var apiResp HackClubResponse
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
func (p *HackClubProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetTokenCount estimates token count for a prompt
func (p *HackClubProvider) GetTokenCount(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// handleHTTPError converts HTTP errors to ExplanationError
func (p *HackClubProvider) handleHTTPError(err error) error {
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

// mapErrorCode maps Hack Club AI error types to our error codes
func (p *HackClubProvider) mapErrorCode(errorType string) string {
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

// Made with Bob
