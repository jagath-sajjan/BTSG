package explainer

import (
	"context"
	"time"

	"btsg/internal/scanner"
)

// ExplanationRequest represents a request for vulnerability explanation
type ExplanationRequest struct {
	Finding     *scanner.Finding
	Language    string // Programming language context
	Framework   string // Framework context (optional)
	IncludeCode bool   // Whether to include code snippet in explanation
}

// Explanation represents the AI-generated explanation of a vulnerability
type Explanation struct {
	// Core explanation
	SimpleExplanation string `json:"simple_explanation"`
	TechnicalDetails  string `json:"technical_details"`

	// Risk assessment
	RiskImpact RiskImpact `json:"risk_impact"`

	// Real-world context
	RealWorldExample *RealWorldExample `json:"real_world_example,omitempty"`

	// Remediation
	RemediationSteps []RemediationStep `json:"remediation_steps"`
	CodeExample      *CodeExample      `json:"code_example,omitempty"`

	// Metadata
	Confidence   float64   `json:"confidence"` // 0.0-1.0
	Source       string    `json:"source"`     // "ai", "template", "cache"
	GeneratedAt  time.Time `json:"generated_at"`
	CacheKey     string    `json:"cache_key,omitempty"`
	TokensUsed   int       `json:"tokens_used,omitempty"`
	ResponseTime int64     `json:"response_time_ms,omitempty"`
}

// RiskImpact describes the potential impact of the vulnerability
type RiskImpact struct {
	Likelihood   string   `json:"likelihood"`    // "low", "medium", "high", "critical"
	Impact       string   `json:"impact"`        // "low", "medium", "high", "critical"
	Scenarios    []string `json:"scenarios"`     // Potential attack scenarios
	AffectedData []string `json:"affected_data"` // Types of data at risk
	CVSS         *CVSS    `json:"cvss,omitempty"`
}

// CVSS represents Common Vulnerability Scoring System data
type CVSS struct {
	Score  float64 `json:"score"`
	Vector string  `json:"vector"`
	Rating string  `json:"rating"` // "low", "medium", "high", "critical"
}

// RealWorldExample provides context from actual security incidents
type RealWorldExample struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Year        int    `json:"year,omitempty"`
	Company     string `json:"company,omitempty"`
	Impact      string `json:"impact"`
	Lesson      string `json:"lesson"`
	Reference   string `json:"reference,omitempty"`
}

// RemediationStep describes a specific action to fix the vulnerability
type RemediationStep struct {
	Order       int    `json:"order"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // "immediate", "high", "medium", "low"
	Effort      string `json:"effort"`   // "low", "medium", "high"
}

// CodeExample shows secure code implementation
type CodeExample struct {
	Language    string `json:"language"`
	Before      string `json:"before,omitempty"`
	After       string `json:"after"`
	Explanation string `json:"explanation"`
}

// ExplainerConfig holds configuration for the explainer
type ExplainerConfig struct {
	// AI Provider settings
	Provider      string        // "openai", "claude", "local"
	APIKey        string        // API key for the provider
	Model         string        // Model name (e.g., "gpt-4", "claude-3-opus")
	Temperature   float64       // 0.0-1.0, controls randomness
	MaxTokens     int           // Maximum tokens in response
	Timeout       time.Duration // Request timeout
	RetryAttempts int           // Number of retry attempts
	RetryDelay    time.Duration // Delay between retries

	// Cache settings
	EnableCache   bool          // Enable caching
	CacheType     string        // "memory", "redis"
	CacheTTL      time.Duration // Cache time-to-live
	RedisAddr     string        // Redis address (if using Redis)
	RedisPassword string        // Redis password
	RedisDB       int           // Redis database number

	// Fallback settings
	EnableFallback   bool   // Enable template fallback
	FallbackProvider string // Secondary AI provider
	TemplateDir      string // Directory for template files

	// Batch processing
	BatchSize      int // Number of concurrent explanations
	WorkerPoolSize int // Number of worker goroutines

	// Features
	IncludeExamples bool   // Include real-world examples
	IncludeCode     bool   // Include code examples
	DetailLevel     string // "basic", "detailed", "expert"
}

// DefaultConfig returns a default configuration
func DefaultConfig() *ExplainerConfig {
	return &ExplainerConfig{
		Provider:        "openai",
		Model:           "gpt-4",
		Temperature:     0.7,
		MaxTokens:       2000,
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      2 * time.Second,
		EnableCache:     true,
		CacheType:       "memory",
		CacheTTL:        24 * time.Hour,
		EnableFallback:  true,
		TemplateDir:     "templates",
		BatchSize:       10,
		WorkerPoolSize:  5,
		IncludeExamples: true,
		IncludeCode:     true,
		DetailLevel:     "detailed",
	}
}

// Explainer is the main interface for generating vulnerability explanations
type Explainer interface {
	// Explain generates an explanation for a single vulnerability
	Explain(ctx context.Context, req *ExplanationRequest) (*Explanation, error)

	// ExplainBatch generates explanations for multiple vulnerabilities
	ExplainBatch(ctx context.Context, reqs []*ExplanationRequest) ([]*Explanation, error)

	// GetCacheStats returns cache statistics
	GetCacheStats() *CacheStats

	// ClearCache clears the explanation cache
	ClearCache() error

	// Close closes the explainer and releases resources
	Close() error
}

// AIProvider is the interface for AI service providers
type AIProvider interface {
	// Name returns the provider name
	Name() string

	// GenerateExplanation generates an explanation using the AI model
	GenerateExplanation(ctx context.Context, prompt string) (string, error)

	// IsAvailable checks if the provider is available
	IsAvailable() bool

	// GetTokenCount estimates token count for a prompt
	GetTokenCount(text string) int
}

// Cache is the interface for caching explanations
type Cache interface {
	// Get retrieves an explanation from cache
	Get(ctx context.Context, key string) (*Explanation, error)

	// Set stores an explanation in cache
	Set(ctx context.Context, key string, explanation *Explanation, ttl time.Duration) error

	// Delete removes an explanation from cache
	Delete(ctx context.Context, key string) error

	// Clear clears all cached explanations
	Clear(ctx context.Context) error

	// Stats returns cache statistics
	Stats() *CacheStats

	// Close closes the cache connection
	Close() error
}

// CacheStats holds cache performance statistics
type CacheStats struct {
	Hits       int64   `json:"hits"`
	Misses     int64   `json:"misses"`
	HitRate    float64 `json:"hit_rate"`
	Size       int64   `json:"size"`
	Evictions  int64   `json:"evictions"`
	AvgGetTime int64   `json:"avg_get_time_ms"`
	TotalSaved float64 `json:"total_saved_usd"` // Estimated cost savings
}

// PromptGenerator generates prompts for AI models
type PromptGenerator interface {
	// GeneratePrompt creates a prompt for the given vulnerability
	GeneratePrompt(req *ExplanationRequest) (string, error)

	// GetSystemPrompt returns the system prompt for the AI model
	GetSystemPrompt() string
}

// TemplateEngine provides template-based explanations as fallback
type TemplateEngine interface {
	// Generate creates a template-based explanation
	Generate(req *ExplanationRequest) (*Explanation, error)

	// LoadTemplates loads templates from directory
	LoadTemplates(dir string) error

	// HasTemplate checks if a template exists for the vulnerability type
	HasTemplate(vulnType string) bool
}

// ExplanationError represents an error during explanation generation
type ExplanationError struct {
	Code    string // Error code
	Message string // Error message
	Cause   error  // Underlying error
}

func (e *ExplanationError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Error codes
const (
	ErrCodeInvalidRequest  = "INVALID_REQUEST"
	ErrCodeAPIError        = "API_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeRateLimit       = "RATE_LIMIT"
	ErrCodeAuthError       = "AUTH_ERROR"
	ErrCodeParseError      = "PARSE_ERROR"
	ErrCodeCacheError      = "CACHE_ERROR"
	ErrCodeTemplateError   = "TEMPLATE_ERROR"
	ErrCodeNoProviderAvail = "NO_PROVIDER_AVAILABLE"
)

// Made with Bob
