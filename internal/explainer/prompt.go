package explainer

import (
	"btsg/internal/scanner"
	"encoding/json"
	"fmt"
	"strings"
)

// promptGenerator implements the PromptGenerator interface
type promptGenerator struct {
	config *ExplainerConfig
}

// NewPromptGenerator creates a new prompt generator
func NewPromptGenerator(config *ExplainerConfig) PromptGenerator {
	return &promptGenerator{
		config: config,
	}
}

// GetSystemPrompt returns the system prompt for the AI model
func (p *promptGenerator) GetSystemPrompt() string {
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

// GeneratePrompt creates a prompt for the given vulnerability
func (p *promptGenerator) GeneratePrompt(req *ExplanationRequest) (string, error) {
	if req == nil || req.Finding == nil {
		return "", &ExplanationError{
			Code:    ErrCodeInvalidRequest,
			Message: "request or finding is nil",
		}
	}

	// Select template based on vulnerability type
	// Infer type from tool name and description
	vulnType := inferVulnerabilityType(req.Finding)

	var template string
	switch vulnType {
	case "secret":
		template = p.generateSecretPrompt(req)
	case "dependency":
		template = p.generateDependencyPrompt(req)
	default:
		template = p.generateCodePrompt(req)
	}

	return template, nil
}

// generateCodePrompt creates a prompt for code vulnerabilities
func (p *promptGenerator) generateCodePrompt(req *ExplanationRequest) string {
	finding := req.Finding
	vulnType := inferVulnerabilityType(finding)

	prompt := fmt.Sprintf(`Analyze this security vulnerability and provide a comprehensive explanation.

## Vulnerability Details
- **Type**: %s
- **Severity**: %s
- **Tool**: %s
- **File**: %s
- **Line**: %d
- **Description**: %s
`, vulnType, finding.Severity, finding.Tool, finding.File, finding.Line, finding.Description)

	if finding.CWE != "" {
		prompt += fmt.Sprintf("- **CWE**: %s\n", finding.CWE)
	}

	if req.IncludeCode && finding.Code != "" {
		prompt += fmt.Sprintf("\n## Vulnerable Code\n```%s\n%s\n```\n", req.Language, finding.Code)
	}

	if req.Language != "" {
		prompt += fmt.Sprintf("\n**Language**: %s\n", req.Language)
	}

	if req.Framework != "" {
		prompt += fmt.Sprintf("**Framework**: %s\n", req.Framework)
	}

	prompt += p.getOutputSchema()
	prompt += p.getInstructions()

	return prompt
}

// generateDependencyPrompt creates a prompt for dependency vulnerabilities
func (p *promptGenerator) generateDependencyPrompt(req *ExplanationRequest) string {
	finding := req.Finding
	packageName := extractPackageName(finding)

	prompt := fmt.Sprintf(`Analyze this dependency vulnerability and provide a comprehensive explanation.

## Vulnerability Details
- **Package**: %s
- **Severity**: %s
- **Description**: %s
`, packageName, finding.Severity, finding.Description)

	if finding.CWE != "" {
		prompt += fmt.Sprintf("- **CVE/CWE**: %s\n", finding.CWE)
	}

	if finding.File != "" {
		prompt += fmt.Sprintf("- **Dependency File**: %s\n", finding.File)
	}

	if req.Language != "" {
		prompt += fmt.Sprintf("\n**Ecosystem**: %s\n", req.Language)
	}

	prompt += p.getOutputSchema()
	prompt += p.getInstructions()

	return prompt
}

// generateSecretPrompt creates a prompt for secret/credential leaks
func (p *promptGenerator) generateSecretPrompt(req *ExplanationRequest) string {
	finding := req.Finding
	secretType := extractSecretType(finding)

	prompt := fmt.Sprintf(`Analyze this secret/credential leak and provide a comprehensive explanation.

## Secret Leak Details
- **Type**: %s
- **Severity**: %s
- **File**: %s
- **Line**: %d
- **Description**: %s
`, secretType, finding.Severity, finding.File, finding.Line, finding.Description)

	if req.IncludeCode && finding.Code != "" {
		// Redact the actual secret in the prompt
		redactedCode := redactSecret(finding.Code)
		prompt += fmt.Sprintf("\n## Location (redacted)\n```\n%s\n```\n", redactedCode)
	}

	prompt += p.getOutputSchema()
	prompt += p.getInstructions()

	return prompt
}

// getOutputSchema returns the JSON schema for the expected output
func (p *promptGenerator) getOutputSchema() string {
	schema := `

## Required Output Format

Respond with a JSON object matching this exact schema:

{
  "simple_explanation": "string - Plain language explanation (2-3 sentences)",
  "technical_details": "string - Technical explanation for developers",
  "risk_impact": {
    "likelihood": "string - low|medium|high|critical",
    "impact": "string - low|medium|high|critical",
    "scenarios": ["string - potential attack scenario 1", "..."],
    "affected_data": ["string - type of data at risk 1", "..."]
  },`

	if p.config.IncludeExamples {
		schema += `
  "real_world_example": {
    "title": "string - Brief title of real incident",
    "description": "string - What happened",
    "year": number - Year of incident (optional),
    "company": "string - Affected organization (optional)",
    "impact": "string - Consequences",
    "lesson": "string - Key takeaway",
    "reference": "string - URL or source (optional)"
  },`
	}

	schema += `
  "remediation_steps": [
    {
      "order": number - Step number,
      "action": "string - Brief action title",
      "description": "string - Detailed step description",
      "priority": "string - immediate|high|medium|low",
      "effort": "string - low|medium|high"
    }
  ]`

	if p.config.IncludeCode {
		schema += `,
  "code_example": {
    "language": "string - Programming language",
    "before": "string - Vulnerable code (optional)",
    "after": "string - Secure code",
    "explanation": "string - What changed and why"
  }`
	}

	schema += `
}
`
	return schema
}

// getInstructions returns specific instructions for the AI
func (p *promptGenerator) getInstructions() string {
	instructions := `

## Instructions

1. **Simple Explanation**: Write for developers who may not be security experts. Use analogies if helpful.

2. **Risk Assessment**: 
   - Be realistic about likelihood and impact
   - Provide 2-4 concrete attack scenarios
   - List specific types of data that could be compromised

3. **Remediation Steps**:
   - Provide 3-5 actionable steps
   - Order by priority (most critical first)
   - Be specific about what to do
   - Include effort estimates

`

	if p.config.IncludeExamples {
		instructions += `4. **Real-World Example**: 
   - Use actual security incidents when possible
   - Focus on lessons learned
   - Keep it relevant to this vulnerability type

`
	}

	if p.config.IncludeCode {
		instructions += `5. **Code Example**:
   - Show secure implementation
   - Explain the key security improvements
   - Use best practices for the language/framework

`
	}

	instructions += `
**Important**: 
- Output ONLY valid JSON, no additional text
- Be concise but complete
- Focus on actionable advice
- Use professional but friendly tone
`

	return instructions
}

// redactSecret replaces potential secrets with placeholders
func redactSecret(code string) string {
	// Simple redaction - replace anything that looks like a secret
	redacted := code

	// Redact common secret patterns
	if strings.Contains(redacted, "sk-") {
		redacted = strings.ReplaceAll(redacted, "sk-", "[REDACTED_API_KEY]")
	}
	if strings.Contains(redacted, "ghp_") {
		redacted = strings.ReplaceAll(redacted, "ghp_", "[REDACTED_GITHUB_TOKEN]")
	}
	if strings.Contains(redacted, "xox") {
		redacted = strings.ReplaceAll(redacted, "xox", "[REDACTED_SLACK_TOKEN]")
	}

	return redacted
}

// inferVulnerabilityType infers the vulnerability type from the finding
func inferVulnerabilityType(finding *scanner.Finding) string {
	tool := strings.ToLower(finding.Tool)
	desc := strings.ToLower(finding.Description)

	// Check for secret detection
	if tool == "detect-secrets" || strings.Contains(desc, "secret") ||
		strings.Contains(desc, "credential") || strings.Contains(desc, "api key") ||
		strings.Contains(desc, "token") || strings.Contains(desc, "password") {
		return "secret"
	}

	// Check for dependency vulnerabilities
	if tool == "pip-audit" || strings.Contains(desc, "cve-") ||
		strings.Contains(desc, "dependency") || strings.Contains(desc, "package") {
		return "dependency"
	}

	// Default to code vulnerability
	return "code"
}

// extractPackageName extracts package name from finding
func extractPackageName(finding *scanner.Finding) string {
	// Try to extract from description
	desc := finding.Description
	if strings.Contains(desc, "package") {
		// Simple extraction - in production use regex
		parts := strings.Split(desc, " ")
		for i, part := range parts {
			if part == "package" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return "Unknown Package"
}

// extractSecretType extracts the type of secret from finding
func extractSecretType(finding *scanner.Finding) string {
	desc := strings.ToLower(finding.Description)

	if strings.Contains(desc, "api key") {
		return "API Key"
	}
	if strings.Contains(desc, "password") {
		return "Password"
	}
	if strings.Contains(desc, "token") {
		return "Access Token"
	}
	if strings.Contains(desc, "private key") {
		return "Private Key"
	}
	if strings.Contains(desc, "secret") {
		return "Secret"
	}

	return "Credential"
}

// GenerateBatchPrompt creates a prompt for batch processing
func (p *promptGenerator) GenerateBatchPrompt(reqs []*ExplanationRequest) (string, error) {
	if len(reqs) == 0 {
		return "", &ExplanationError{
			Code:    ErrCodeInvalidRequest,
			Message: "no requests provided",
		}
	}

	prompt := `Analyze these security vulnerabilities and provide explanations for each.

## Vulnerabilities

`

	for i, req := range reqs {
		if req.Finding == nil {
			continue
		}

		vulnType := inferVulnerabilityType(req.Finding)
		prompt += fmt.Sprintf(`### Vulnerability %d
- Type: %s
- Severity: %s
- File: %s:%d
- Description: %s

`, i+1, vulnType, req.Finding.Severity,
			req.Finding.File, req.Finding.Line, req.Finding.Description)
	}

	prompt += `
## Output Format

Respond with a JSON array of explanation objects, one for each vulnerability in order.
Each object should follow the schema provided earlier.

Example:
[
  { /* explanation for vulnerability 1 */ },
  { /* explanation for vulnerability 2 */ },
  ...
]
`

	return prompt, nil
}

// EstimateTokens estimates the number of tokens in a prompt
func (p *promptGenerator) EstimateTokens(prompt string) int {
	// Rough estimation: ~4 characters per token
	// This is a simplified version; production should use tiktoken or similar
	return len(prompt) / 4
}

// ValidateResponse validates the AI response structure
func ValidateResponse(response string) error {
	var explanation Explanation
	if err := json.Unmarshal([]byte(response), &explanation); err != nil {
		return &ExplanationError{
			Code:    ErrCodeParseError,
			Message: "failed to parse AI response as JSON",
			Cause:   err,
		}
	}

	// Validate required fields
	if explanation.SimpleExplanation == "" {
		return &ExplanationError{
			Code:    ErrCodeParseError,
			Message: "missing required field: simple_explanation",
		}
	}

	if len(explanation.RemediationSteps) == 0 {
		return &ExplanationError{
			Code:    ErrCodeParseError,
			Message: "missing required field: remediation_steps",
		}
	}

	if explanation.RiskImpact.Likelihood == "" || explanation.RiskImpact.Impact == "" {
		return &ExplanationError{
			Code:    ErrCodeParseError,
			Message: "missing required risk_impact fields",
		}
	}

	return nil
}

// Made with Bob
