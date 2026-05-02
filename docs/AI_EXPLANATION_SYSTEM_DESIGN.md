# BTSG AI Explanation System Design

## Overview

The AI Explanation System converts raw vulnerability data into human-friendly explanations with simple language, risk impact analysis, and real-world examples.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Vulnerability Input                       │
│  {id, tool, severity, file, line, description, code}        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  Explanation Engine                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Prompt     │  │     AI       │  │    Cache     │     │
│  │  Generator   │  │   Provider   │  │   Manager    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  AI Explanation Output                       │
│  {                                                           │
│    simple_explanation,                                       │
│    risk_impact,                                              │
│    real_world_example,                                       │
│    remediation_steps,                                        │
│    references                                                │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
```

## 1. Input Schema

### Vulnerability Input

```go
type VulnerabilityInput struct {
    // Core Information
    ID          string `json:"id"`
    Tool        string `json:"tool"`
    Severity    string `json:"severity"`
    Type        string `json:"type"`
    
    // Location
    File        string `json:"file"`
    Line        int    `json:"line"`
    
    // Details
    Description string `json:"description"`
    Code        string `json:"code,omitempty"`
    CWE         string `json:"cwe,omitempty"`
    CVE         string `json:"cve,omitempty"`
    
    // Context
    Language    string `json:"language,omitempty"`
    Framework   string `json:"framework,omitempty"`
}
```

### Example Input

```json
{
  "id": "BTSG-001",
  "tool": "bandit",
  "severity": "HIGH",
  "type": "code",
  "file": "app/views.py",
  "line": 42,
  "description": "[B201] Use of insecure pickle module",
  "code": "import pickle\ndata = pickle.loads(user_input)",
  "cwe": "CWE-502",
  "language": "python"
}
```

## 2. Output Schema

### AI Explanation Output

```go
type AIExplanation struct {
    // Simple Explanation
    SimpleExplanation string `json:"simple_explanation"`
    
    // Risk Impact
    RiskImpact RiskImpact `json:"risk_impact"`
    
    // Real-World Example
    RealWorldExample RealWorldExample `json:"real_world_example"`
    
    // Remediation
    RemediationSteps []string `json:"remediation_steps"`
    FixCode          string   `json:"fix_code,omitempty"`
    
    // References
    References []Reference `json:"references"`
    
    // Metadata
    Confidence float64   `json:"confidence"`
    GeneratedAt time.Time `json:"generated_at"`
    Model       string    `json:"model"`
}

type RiskImpact struct {
    Likelihood  string   `json:"likelihood"`  // LOW, MEDIUM, HIGH
    Impact      string   `json:"impact"`      // LOW, MEDIUM, HIGH, CRITICAL
    Scenarios   []string `json:"scenarios"`
    AffectedData []string `json:"affected_data"`
}

type RealWorldExample struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Incident    string `json:"incident,omitempty"`
    Year        int    `json:"year,omitempty"`
    Impact      string `json:"impact"`
}

type Reference struct {
    Type  string `json:"type"`  // CWE, CVE, OWASP, Blog, Documentation
    Title string `json:"title"`
    URL   string `json:"url"`
}
```

### Example Output

```json
{
  "simple_explanation": "Your code uses Python's pickle module to deserialize data from user input. This is dangerous because pickle can execute arbitrary code during deserialization, allowing attackers to run malicious commands on your server.",
  
  "risk_impact": {
    "likelihood": "HIGH",
    "impact": "CRITICAL",
    "scenarios": [
      "Attacker sends malicious pickle data",
      "Server deserializes and executes attacker's code",
      "Attacker gains full control of the server"
    ],
    "affected_data": [
      "Server files and databases",
      "User data and credentials",
      "Internal network access"
    ]
  },
  
  "real_world_example": {
    "title": "Remote Code Execution via Pickle",
    "description": "In 2019, a popular Python web application was compromised when attackers exploited pickle deserialization in an API endpoint. They gained shell access and exfiltrated customer data.",
    "incident": "CVE-2019-XXXXX",
    "year": 2019,
    "impact": "Data breach affecting 50,000 users"
  },
  
  "remediation_steps": [
    "Replace pickle with JSON for data serialization",
    "If pickle is required, validate and sanitize all input",
    "Use signing/encryption to verify pickle data integrity",
    "Implement strict input validation before deserialization",
    "Run deserialization in sandboxed environment"
  ],
  
  "fix_code": "import json\n\n# Instead of:\n# data = pickle.loads(user_input)\n\n# Use:\ndata = json.loads(user_input)",
  
  "references": [
    {
      "type": "CWE",
      "title": "CWE-502: Deserialization of Untrusted Data",
      "url": "https://cwe.mitre.org/data/definitions/502.html"
    },
    {
      "type": "OWASP",
      "title": "Deserialization Cheat Sheet",
      "url": "https://cheatsheetseries.owasp.org/cheatsheets/Deserialization_Cheat_Sheet.html"
    }
  ],
  
  "confidence": 0.95,
  "generated_at": "2026-05-02T05:42:00Z",
  "model": "gpt-4"
}
```

## 3. Prompt Structure

### System Prompt

```
You are a security expert explaining vulnerabilities to developers. Your goal is to:
1. Explain security issues in simple, clear language
2. Describe realistic attack scenarios and impacts
3. Provide actionable remediation steps
4. Include real-world examples when possible

Guidelines:
- Use plain language, avoid jargon
- Be specific about risks and impacts
- Provide concrete examples
- Focus on practical solutions
- Be encouraging, not alarming
```

### User Prompt Template

```
Analyze this security vulnerability and provide a comprehensive explanation:

VULNERABILITY DETAILS:
- ID: {{.ID}}
- Tool: {{.Tool}}
- Severity: {{.Severity}}
- Type: {{.Type}}
- File: {{.File}}:{{.Line}}
- Description: {{.Description}}
{{if .CWE}}- CWE: {{.CWE}}{{end}}
{{if .CVE}}- CVE: {{.CVE}}{{end}}

{{if .Code}}
VULNERABLE CODE:
```{{.Language}}
{{.Code}}
```
{{end}}

Please provide:

1. SIMPLE EXPLANATION (2-3 sentences)
   - What is this vulnerability?
   - Why is it dangerous?
   - How can it be exploited?

2. RISK IMPACT
   - Likelihood: [LOW/MEDIUM/HIGH]
   - Impact: [LOW/MEDIUM/HIGH/CRITICAL]
   - Attack scenarios (3-5 bullet points)
   - What data/systems are at risk?

3. REAL-WORLD EXAMPLE
   - Title of a similar incident
   - Brief description
   - Year and impact
   - Lessons learned

4. REMEDIATION STEPS
   - 5-7 specific, actionable steps
   - Include code example if applicable
   - Prioritize by importance

5. REFERENCES
   - Relevant CWE/CVE links
   - OWASP resources
   - Security best practices

Format your response as JSON matching this schema:
{
  "simple_explanation": "...",
  "risk_impact": {
    "likelihood": "...",
    "impact": "...",
    "scenarios": [...],
    "affected_data": [...]
  },
  "real_world_example": {
    "title": "...",
    "description": "...",
    "year": 2019,
    "impact": "..."
  },
  "remediation_steps": [...],
  "fix_code": "...",
  "references": [...]
}
```

### Prompt Variations by Vulnerability Type

#### For Code Vulnerabilities (Bandit)

```
Focus on:
- Code-level security issues
- Common programming mistakes
- Secure coding alternatives
- Language-specific best practices
```

#### For Dependency Vulnerabilities (pip-audit)

```
Focus on:
- CVE details and CVSS scores
- Affected versions
- Available patches/updates
- Upgrade path and compatibility
```

#### For Secret Leaks (detect-secrets)

```
Focus on:
- Exposure risks
- Credential rotation
- Secret management solutions
- Prevention strategies
```

## 4. API Flow

### Flow Diagram

```
User Request
     │
     ▼
┌─────────────────┐
│ Check Cache     │
│ (Redis/Memory)  │
└────┬────────────┘
     │
     ├─ Cache Hit ──────────────────┐
     │                              │
     └─ Cache Miss                  │
          │                         │
          ▼                         │
     ┌─────────────────┐           │
     │ Generate Prompt │           │
     └────┬────────────┘           │
          │                         │
          ▼                         │
     ┌─────────────────┐           │
     │ Call AI API     │           │
     │ (OpenAI/Claude) │           │
     └────┬────────────┘           │
          │                         │
          ▼                         │
     ┌─────────────────┐           │
     │ Parse Response  │           │
     └────┬────────────┘           │
          │                         │
          ▼                         │
     ┌─────────────────┐           │
     │ Validate Output │           │
     └────┬────────────┘           │
          │                         │
          ▼                         │
     ┌─────────────────┐           │
     │ Store in Cache  │           │
     └────┬────────────┘           │
          │                         │
          └─────────────────────────┤
                                    │
                                    ▼
                            ┌─────────────────┐
                            │ Return to User  │
                            └─────────────────┘
```

### Implementation Flow

```go
func (e *ExplanationEngine) Explain(vuln *Vulnerability) (*AIExplanation, error) {
    // 1. Check cache
    cacheKey := generateCacheKey(vuln)
    if cached, found := e.cache.Get(cacheKey); found {
        return cached.(*AIExplanation), nil
    }
    
    // 2. Generate prompt
    prompt, err := e.promptGenerator.Generate(vuln)
    if err != nil {
        return nil, err
    }
    
    // 3. Call AI API
    response, err := e.aiProvider.Complete(prompt)
    if err != nil {
        return nil, err
    }
    
    // 4. Parse response
    explanation, err := e.parser.Parse(response)
    if err != nil {
        return nil, err
    }
    
    // 5. Validate
    if err := e.validator.Validate(explanation); err != nil {
        return nil, err
    }
    
    // 6. Cache result
    e.cache.Set(cacheKey, explanation, 24*time.Hour)
    
    return explanation, nil
}
```

## 5. AI Provider Integration

### Provider Interface

```go
type AIProvider interface {
    // Complete generates a completion for the given prompt
    Complete(prompt string) (string, error)
    
    // CompleteWithOptions generates with custom options
    CompleteWithOptions(prompt string, opts CompletionOptions) (string, error)
    
    // Name returns the provider name
    Name() string
    
    // Model returns the model being used
    Model() string
}

type CompletionOptions struct {
    Temperature   float64
    MaxTokens     int
    TopP          float64
    FrequencyPenalty float64
    PresencePenalty  float64
}
```

### OpenAI Provider

```go
type OpenAIProvider struct {
    apiKey string
    model  string
    client *openai.Client
}

func (p *OpenAIProvider) Complete(prompt string) (string, error) {
    resp, err := p.client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: p.model,
            Messages: []openai.ChatCompletionMessage{
                {
                    Role:    openai.ChatMessageRoleSystem,
                    Content: systemPrompt,
                },
                {
                    Role:    openai.ChatMessageRoleUser,
                    Content: prompt,
                },
            },
            Temperature: 0.7,
            MaxTokens:   2000,
        },
    )
    
    if err != nil {
        return "", err
    }
    
    return resp.Choices[0].Message.Content, nil
}
```

### Claude Provider

```go
type ClaudeProvider struct {
    apiKey string
    model  string
    client *anthropic.Client
}

func (p *ClaudeProvider) Complete(prompt string) (string, error) {
    resp, err := p.client.Complete(
        context.Background(),
        anthropic.CompletionRequest{
            Model:       p.model,
            Prompt:      fmt.Sprintf("%s\n\nHuman: %s\n\nAssistant:", systemPrompt, prompt),
            MaxTokens:   2000,
            Temperature: 0.7,
        },
    )
    
    if err != nil {
        return "", err
    }
    
    return resp.Completion, nil
}
```

## 6. Caching Strategy

### Cache Key Generation

```go
func generateCacheKey(vuln *Vulnerability) string {
    // Create deterministic key from vulnerability attributes
    data := fmt.Sprintf("%s:%s:%s:%s:%d",
        vuln.Tool,
        vuln.Type,
        vuln.Description,
        vuln.CWE,
        vuln.Line,
    )
    
    hash := sha256.Sum256([]byte(data))
    return fmt.Sprintf("btsg:explain:%x", hash[:16])
}
```

### Cache Implementation

```go
type ExplanationCache interface {
    Get(key string) (*AIExplanation, bool)
    Set(key string, explanation *AIExplanation, ttl time.Duration) error
    Delete(key string) error
    Clear() error
}

// In-memory cache
type MemoryCache struct {
    data map[string]*cacheEntry
    mu   sync.RWMutex
}

type cacheEntry struct {
    explanation *AIExplanation
    expiresAt   time.Time
}

// Redis cache
type RedisCache struct {
    client *redis.Client
}
```

### Cache Strategy

1. **TTL**: 24 hours for explanations
2. **Invalidation**: On vulnerability definition updates
3. **Size Limit**: LRU eviction for memory cache
4. **Persistence**: Redis for distributed caching

## 7. Error Handling

### Error Types

```go
type ExplanationError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType string

const (
    ErrorTypeAPIFailure      ErrorType = "api_failure"
    ErrorTypeRateLimit       ErrorType = "rate_limit"
    ErrorTypeInvalidResponse ErrorType = "invalid_response"
    ErrorTypeTimeout         ErrorType = "timeout"
    ErrorTypeCacheMiss       ErrorType = "cache_miss"
)
```

### Fallback Strategy

```go
func (e *ExplanationEngine) ExplainWithFallback(vuln *Vulnerability) (*AIExplanation, error) {
    // Try primary provider
    explanation, err := e.primaryProvider.Explain(vuln)
    if err == nil {
        return explanation, nil
    }
    
    // Try fallback provider
    if e.fallbackProvider != nil {
        explanation, err = e.fallbackProvider.Explain(vuln)
        if err == nil {
            return explanation, nil
        }
    }
    
    // Use template-based explanation
    return e.templateGenerator.Generate(vuln), nil
}
```

### Template-Based Fallback

```go
type TemplateGenerator struct {
    templates map[string]*ExplanationTemplate
}

type ExplanationTemplate struct {
    SimpleExplanation string
    RiskScenarios     []string
    RemediationSteps  []string
}

func (g *TemplateGenerator) Generate(vuln *Vulnerability) *AIExplanation {
    template := g.getTemplate(vuln.Type, vuln.CWE)
    
    return &AIExplanation{
        SimpleExplanation: template.SimpleExplanation,
        RiskImpact: RiskImpact{
            Likelihood: mapSeverityToLikelihood(vuln.Severity),
            Impact:     vuln.Severity,
            Scenarios:  template.RiskScenarios,
        },
        RemediationSteps: template.RemediationSteps,
        Confidence:       0.6, // Lower confidence for template
    }
}
```

## 8. Configuration

### Configuration Schema

```go
type ExplanationConfig struct {
    // AI Provider
    Provider      string `yaml:"provider"`       // openai, claude, local
    APIKey        string `yaml:"api_key"`
    Model         string `yaml:"model"`
    
    // Caching
    CacheEnabled  bool          `yaml:"cache_enabled"`
    CacheType     string        `yaml:"cache_type"`     // memory, redis
    CacheTTL      time.Duration `yaml:"cache_ttl"`
    RedisURL      string        `yaml:"redis_url"`
    
    // Behavior
    Temperature   float64 `yaml:"temperature"`
    MaxTokens     int     `yaml:"max_tokens"`
    Timeout       time.Duration `yaml:"timeout"`
    
    // Fallback
    UseFallback   bool   `yaml:"use_fallback"`
    FallbackProvider string `yaml:"fallback_provider"`
}
```

### Example Configuration

```yaml
# .btsg.yml
explanation:
  provider: openai
  api_key: ${OPENAI_API_KEY}
  model: gpt-4
  
  cache_enabled: true
  cache_type: redis
  cache_ttl: 24h
  redis_url: redis://localhost:6379
  
  temperature: 0.7
  max_tokens: 2000
  timeout: 30s
  
  use_fallback: true
  fallback_provider: template
```

## 9. Usage Examples

### Basic Usage

```go
// Initialize
config := &ExplanationConfig{
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Model:    "gpt-4",
}

engine := NewExplanationEngine(config)

// Explain vulnerability
vuln := &Vulnerability{
    ID:          "BTSG-001",
    Tool:        "bandit",
    Severity:    "HIGH",
    Description: "Use of insecure pickle module",
    Code:        "data = pickle.loads(user_input)",
}

explanation, err := engine.Explain(vuln)
if err != nil {
    log.Fatal(err)
}

// Display
fmt.Println("Simple Explanation:")
fmt.Println(explanation.SimpleExplanation)
fmt.Println("\nRisk Impact:")
fmt.Printf("Likelihood: %s, Impact: %s\n", 
    explanation.RiskImpact.Likelihood,
    explanation.RiskImpact.Impact)
```

### Batch Explanation

```go
func ExplainBatch(engine *ExplanationEngine, vulns []*Vulnerability) ([]*AIExplanation, error) {
    var wg sync.WaitGroup
    results := make([]*AIExplanation, len(vulns))
    errors := make([]error, len(vulns))
    
    for i, vuln := range vulns {
        wg.Add(1)
        go func(idx int, v *Vulnerability) {
            defer wg.Done()
            explanation, err := engine.Explain(v)
            if err != nil {
                errors[idx] = err
                return
            }
            results[idx] = explanation
        }(i, vuln)
    }
    
    wg.Wait()
    
    // Check for errors
    for _, err := range errors {
        if err != nil {
            return results, err
        }
    }
    
    return results, nil
}
```

## 10. Performance Considerations

### Optimization Strategies

1. **Caching**: Cache explanations for 24 hours
2. **Batching**: Group similar vulnerabilities
3. **Async Processing**: Generate explanations in background
4. **Rate Limiting**: Respect API rate limits
5. **Compression**: Compress cached data

### Performance Metrics

```go
type ExplanationMetrics struct {
    TotalRequests    int64
    CacheHits        int64
    CacheMisses      int64
    APICallsSuccess  int64
    APICallsFailure  int64
    AverageLatency   time.Duration
    TotalTokensUsed  int64
}
```

## Conclusion

This AI Explanation System provides:

✅ **Simple Explanations** - Plain language for developers
✅ **Risk Analysis** - Likelihood and impact assessment
✅ **Real Examples** - Concrete incidents and lessons
✅ **Actionable Steps** - Clear remediation guidance
✅ **Flexible Integration** - Multiple AI providers
✅ **Performance** - Caching and optimization
✅ **Reliability** - Fallback mechanisms

The system is designed to be extensible, performant, and production-ready.