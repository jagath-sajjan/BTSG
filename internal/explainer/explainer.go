package explainer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// explainer implements the Explainer interface
type explainer struct {
	config           *ExplainerConfig
	provider         AIProvider
	fallbackProvider AIProvider
	cache            Cache
	promptGenerator  PromptGenerator
	templateEngine   TemplateEngine
	mu               sync.RWMutex
}

// NewExplainer creates a new explainer instance
func NewExplainer(config *ExplainerConfig) (Explainer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create AI provider
	var provider AIProvider
	switch config.Provider {
	case "openai":
		provider = NewOpenAIProvider(config)
	case "hackclub":
		provider = NewHackClubProvider(config)
	default:
		return nil, &ExplanationError{
			Code:    ErrCodeInvalidRequest,
			Message: fmt.Sprintf("unsupported provider: %s", config.Provider),
		}
	}

	// Wrap with retry logic
	provider = NewRetryableProvider(provider, config.RetryAttempts, config.RetryDelay)

	// Create fallback provider if enabled
	var fallbackProvider AIProvider
	if config.EnableFallback && config.FallbackProvider != "" && config.FallbackProvider != config.Provider {
		// In production, create actual fallback provider
		// For now, we'll use template fallback
	}

	// Create cache
	var cache Cache
	if config.EnableCache {
		switch config.CacheType {
		case "memory":
			cache = NewMemoryCache(config.CacheTTL)
		default:
			cache = NewMemoryCache(config.CacheTTL)
		}
	}

	// Create prompt generator
	promptGen := NewPromptGenerator(config)

	// Create template engine
	templateEngine := NewTemplateEngine(config.TemplateDir)

	exp := &explainer{
		config:           config,
		provider:         provider,
		fallbackProvider: fallbackProvider,
		cache:            cache,
		promptGenerator:  promptGen,
		templateEngine:   templateEngine,
	}

	// Wrap with cache middleware if enabled
	if config.EnableCache && cache != nil {
		return NewCacheMiddleware(exp, cache), nil
	}

	return exp, nil
}

// Explain generates an explanation for a single vulnerability
func (e *explainer) Explain(ctx context.Context, req *ExplanationRequest) (*Explanation, error) {
	if req == nil || req.Finding == nil {
		return nil, &ExplanationError{
			Code:    ErrCodeInvalidRequest,
			Message: "invalid request: finding is nil",
		}
	}

	startTime := time.Now()

	// Generate prompt
	prompt, err := e.promptGenerator.GeneratePrompt(req)
	if err != nil {
		return nil, err
	}

	// Try primary provider
	explanation, err := e.generateWithProvider(ctx, prompt, e.provider)
	if err == nil {
		explanation.GeneratedAt = time.Now()
		explanation.ResponseTime = time.Since(startTime).Milliseconds()
		explanation.Source = "ai"
		explanation.Confidence = 0.9
		return explanation, nil
	}

	// Try fallback provider if available
	if e.fallbackProvider != nil && e.fallbackProvider.IsAvailable() {
		explanation, err = e.generateWithProvider(ctx, prompt, e.fallbackProvider)
		if err == nil {
			explanation.GeneratedAt = time.Now()
			explanation.ResponseTime = time.Since(startTime).Milliseconds()
			explanation.Source = "ai-fallback"
			explanation.Confidence = 0.85
			return explanation, nil
		}
	}

	// Use template fallback if enabled
	if e.config.EnableFallback && e.templateEngine != nil {
		explanation, err := e.templateEngine.Generate(req)
		if err == nil {
			explanation.GeneratedAt = time.Now()
			explanation.ResponseTime = time.Since(startTime).Milliseconds()
			explanation.Source = "template"
			explanation.Confidence = 0.6
			return explanation, nil
		}
	}

	return nil, &ExplanationError{
		Code:    ErrCodeNoProviderAvail,
		Message: "all providers failed",
		Cause:   err,
	}
}

// ExplainBatch generates explanations for multiple vulnerabilities
func (e *explainer) ExplainBatch(ctx context.Context, reqs []*ExplanationRequest) ([]*Explanation, error) {
	if len(reqs) == 0 {
		return []*Explanation{}, nil
	}

	// Use worker pool for concurrent processing
	results := make([]*Explanation, len(reqs))
	errors := make([]error, len(reqs))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, e.config.WorkerPoolSize)

	for i, req := range reqs {
		wg.Add(1)
		go func(index int, request *ExplanationRequest) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Generate explanation
			explanation, err := e.Explain(ctx, request)
			if err != nil {
				errors[index] = err
				return
			}
			results[index] = explanation
		}(i, req)
	}

	wg.Wait()

	// Check if any succeeded
	successCount := 0
	for i, result := range results {
		if result != nil {
			successCount++
		} else if errors[i] != nil {
			// Create a minimal error explanation
			results[i] = &Explanation{
				SimpleExplanation: fmt.Sprintf("Failed to generate explanation: %v", errors[i]),
				Source:            "error",
				Confidence:        0.0,
				GeneratedAt:       time.Now(),
			}
		}
	}

	if successCount == 0 {
		return nil, &ExplanationError{
			Code:    ErrCodeAPIError,
			Message: "all explanations failed",
		}
	}

	return results, nil
}

// generateWithProvider generates an explanation using a specific provider
func (e *explainer) generateWithProvider(ctx context.Context, prompt string, provider AIProvider) (*Explanation, error) {
	if !provider.IsAvailable() {
		return nil, &ExplanationError{
			Code:    ErrCodeNoProviderAvail,
			Message: "provider not available",
		}
	}

	// Call AI provider
	response, err := provider.GenerateExplanation(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response
	explanation, err := e.parseResponse(response)
	if err != nil {
		return nil, err
	}

	// Estimate tokens used
	explanation.TokensUsed = provider.GetTokenCount(prompt) + provider.GetTokenCount(response)

	return explanation, nil
}

// parseResponse parses the AI response into an Explanation
func (e *explainer) parseResponse(response string) (*Explanation, error) {
	var explanation Explanation

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(response), &explanation); err != nil {
		return nil, &ExplanationError{
			Code:    ErrCodeParseError,
			Message: "failed to parse AI response",
			Cause:   err,
		}
	}

	// Validate response
	if err := ValidateResponse(response); err != nil {
		return nil, err
	}

	return &explanation, nil
}

// GetCacheStats returns cache statistics
func (e *explainer) GetCacheStats() *CacheStats {
	if e.cache != nil {
		return e.cache.Stats()
	}
	return &CacheStats{}
}

// ClearCache clears the explanation cache
func (e *explainer) ClearCache() error {
	if e.cache != nil {
		return e.cache.Clear(context.Background())
	}
	return nil
}

// Close closes the explainer and releases resources
func (e *explainer) Close() error {
	if e.cache != nil {
		return e.cache.Close()
	}
	return nil
}

// SimpleTemplateEngine provides basic template-based explanations
type SimpleTemplateEngine struct {
	templates map[string]*ExplanationTemplate
}

// ExplanationTemplate represents a template for explanations
type ExplanationTemplate struct {
	VulnType          string
	SimpleExplanation string
	TechnicalDetails  string
	RiskImpact        RiskImpact
	RemediationSteps  []RemediationStep
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(templateDir string) TemplateEngine {
	engine := &SimpleTemplateEngine{
		templates: make(map[string]*ExplanationTemplate),
	}

	// Load default templates
	engine.loadDefaultTemplates()

	return engine
}

// Generate creates a template-based explanation
func (t *SimpleTemplateEngine) Generate(req *ExplanationRequest) (*Explanation, error) {
	if req == nil || req.Finding == nil {
		return nil, &ExplanationError{
			Code:    ErrCodeInvalidRequest,
			Message: "invalid request",
		}
	}

	vulnType := inferVulnerabilityType(req.Finding)
	template, exists := t.templates[vulnType]
	if !exists {
		template = t.templates["default"]
	}

	explanation := &Explanation{
		SimpleExplanation: template.SimpleExplanation,
		TechnicalDetails:  template.TechnicalDetails,
		RiskImpact:        template.RiskImpact,
		RemediationSteps:  template.RemediationSteps,
		Confidence:        0.6,
		Source:            "template",
		GeneratedAt:       time.Now(),
	}

	return explanation, nil
}

// LoadTemplates loads templates from directory
func (t *SimpleTemplateEngine) LoadTemplates(dir string) error {
	// In production, load from files
	return nil
}

// HasTemplate checks if a template exists
func (t *SimpleTemplateEngine) HasTemplate(vulnType string) bool {
	_, exists := t.templates[vulnType]
	return exists
}

// loadDefaultTemplates loads built-in templates
func (t *SimpleTemplateEngine) loadDefaultTemplates() {
	// Default template
	t.templates["default"] = &ExplanationTemplate{
		VulnType:          "default",
		SimpleExplanation: "A security vulnerability was detected in your code that could potentially be exploited by attackers.",
		TechnicalDetails:  "This vulnerability requires manual review to determine the exact nature and severity of the issue.",
		RiskImpact: RiskImpact{
			Likelihood:   "medium",
			Impact:       "medium",
			Scenarios:    []string{"Potential security breach", "Data exposure risk"},
			AffectedData: []string{"Application data", "User information"},
		},
		RemediationSteps: []RemediationStep{
			{
				Order:       1,
				Action:      "Review the code",
				Description: "Carefully examine the flagged code section to understand the vulnerability",
				Priority:    "high",
				Effort:      "low",
			},
			{
				Order:       2,
				Action:      "Apply security best practices",
				Description: "Implement recommended security measures for this type of vulnerability",
				Priority:    "high",
				Effort:      "medium",
			},
		},
	}

	// Secret leak template
	t.templates["secret"] = &ExplanationTemplate{
		VulnType:          "secret",
		SimpleExplanation: "A secret or credential has been detected in your code. This could allow unauthorized access to your systems or services.",
		TechnicalDetails:  "Hardcoded secrets in source code are a critical security risk. If this code is committed to version control, the secret is permanently exposed.",
		RiskImpact: RiskImpact{
			Likelihood:   "high",
			Impact:       "critical",
			Scenarios:    []string{"Unauthorized API access", "Account takeover", "Data breach"},
			AffectedData: []string{"API credentials", "Service access", "User data"},
		},
		RemediationSteps: []RemediationStep{
			{
				Order:       1,
				Action:      "Rotate the exposed secret immediately",
				Description: "Generate a new secret/credential and update all systems using it",
				Priority:    "immediate",
				Effort:      "low",
			},
			{
				Order:       2,
				Action:      "Remove secret from code",
				Description: "Delete the hardcoded secret from your source code",
				Priority:    "immediate",
				Effort:      "low",
			},
			{
				Order:       3,
				Action:      "Use environment variables or secret management",
				Description: "Store secrets in environment variables or use a secret management service",
				Priority:    "high",
				Effort:      "medium",
			},
		},
	}

	// Dependency vulnerability template
	t.templates["dependency"] = &ExplanationTemplate{
		VulnType:          "dependency",
		SimpleExplanation: "A known vulnerability exists in one of your project dependencies. This could be exploited if not updated.",
		TechnicalDetails:  "Third-party dependencies may contain security vulnerabilities that are discovered over time. Keeping dependencies updated is crucial for security.",
		RiskImpact: RiskImpact{
			Likelihood:   "medium",
			Impact:       "high",
			Scenarios:    []string{"Remote code execution", "Denial of service", "Data exposure"},
			AffectedData: []string{"Application data", "System resources"},
		},
		RemediationSteps: []RemediationStep{
			{
				Order:       1,
				Action:      "Update the vulnerable dependency",
				Description: "Upgrade to the latest secure version of the package",
				Priority:    "high",
				Effort:      "low",
			},
			{
				Order:       2,
				Action:      "Test your application",
				Description: "Ensure the update doesn't break existing functionality",
				Priority:    "high",
				Effort:      "medium",
			},
			{
				Order:       3,
				Action:      "Set up automated dependency scanning",
				Description: "Use tools to automatically detect vulnerable dependencies",
				Priority:    "medium",
				Effort:      "low",
			},
		},
	}
}

// Made with Bob
