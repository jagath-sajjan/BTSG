package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"btsg/internal/explainer"
	"btsg/internal/scanner"
	"btsg/pkg/results"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	explainDetailed bool
	explainVulnID   string
	explainFromScan bool
)

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain [vulnerability-id]",
	Short: "Get AI-powered explanations for security vulnerabilities",
	Long: `Use AI to explain security vulnerabilities in plain language:
  • What the vulnerability is
  • Why it's dangerous
  • Risk assessment and impact
  • Recommended fixes with code examples

Examples:
  btsg explain BTSG-001
  btsg explain --from-scan
  btsg explain BTSG-001 --detailed`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load environment variables
		if err := godotenv.Load(); err != nil {
			if verbose {
				fmt.Println("Warning: .env file not found, using system environment variables")
			}
		}

		var vulnID string
		if len(args) > 0 {
			vulnID = args[0]
		} else if explainVulnID != "" {
			vulnID = explainVulnID
		}

		if explainFromScan {
			// Explain vulnerabilities from last scan
			explainFromLastScan()
			return
		}

		if vulnID == "" {
			fmt.Println("Error: Please provide a vulnerability ID or use --from-scan")
			cmd.Usage()
			return
		}

		if verbose {
			fmt.Printf("🔍 Explaining vulnerability: %s\n", vulnID)
		}

		// Get finding from scan results
		finding, err := loadFindingByID(vulnID)
		if err != nil {
			exitWithError(fmt.Errorf("failed to load vulnerability: %w", err))
		}

		// Generate explanation
		explanation := generateExplanation(finding)
		if explanation == nil {
			return
		}

		// Display explanation
		displayExplanation(finding, explanation)
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)

	explainCmd.Flags().BoolVarP(&explainDetailed, "detailed", "d", false, "Show detailed technical explanation")
	explainCmd.Flags().StringVar(&explainVulnID, "id", "", "Vulnerability ID to explain")
	explainCmd.Flags().BoolVar(&explainFromScan, "from-scan", false, "Explain all vulnerabilities from last scan")
}

// generateExplanation generates an AI explanation for a finding
func generateExplanation(finding *scanner.Finding) *explainer.Explanation {
	// Get configuration from environment
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ Error: AI_API_KEY not set in .env file")
		fmt.Println("Please add your API key to .env file:")
		fmt.Println("  AI_API_KEY=your-api-key-here")
		return nil
	}

	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		provider = "hackclub"
	}

	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "openai/gpt-5.5-pro"
	}

	baseURL := os.Getenv("AI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://ai.hackclub.com/proxy/v1"
	}

	// Create explainer configuration
	config := &explainer.ExplainerConfig{
		Provider:        provider,
		APIKey:          apiKey,
		Model:           model,
		Temperature:     0.7,
		MaxTokens:       2000,
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      2 * time.Second,
		EnableCache:     true,
		CacheType:       "memory",
		CacheTTL:        24 * time.Hour,
		EnableFallback:  true,
		IncludeExamples: true,
		IncludeCode:     true,
		DetailLevel:     "detailed",
	}

	// Override base URL for Hack Club provider
	if provider == "hackclub" {
		// Store base URL in a way the provider can access it
		// For now, we'll pass it through the config
		config.Model = model // Ensure model is set
	}

	// Create explainer
	exp, err := explainer.NewExplainer(config)
	if err != nil {
		exitWithError(fmt.Errorf("failed to create explainer: %w", err))
	}
	defer exp.Close()

	// Create explanation request
	req := &explainer.ExplanationRequest{
		Finding:     finding,
		Language:    detectLanguage(finding.File),
		IncludeCode: true,
	}

	// Generate explanation
	fmt.Println("🤖 Generating AI explanation...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	explanation, err := exp.Explain(ctx, req)
	if err != nil {
		exitWithError(fmt.Errorf("failed to generate explanation: %w", err))
	}

	if verbose {
		fmt.Printf("✅ Explanation generated (source: %s, confidence: %.2f)\n\n",
			explanation.Source, explanation.Confidence)
	}

	return explanation
}

// displayExplanation displays the explanation in a formatted way
func displayExplanation(finding *scanner.Finding, explanation *explainer.Explanation) {
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  🔒 Security Vulnerability Explanation                        ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════╝\n\n")

	// Basic info
	fmt.Printf("📋 ID: %s\n", finding.ID)
	fmt.Printf("📁 File: %s:%d\n", finding.File, finding.Line)
	fmt.Printf("🔧 Tool: %s\n", finding.Tool)
	fmt.Printf("⚠️  Severity: %s\n\n", finding.Severity)

	// Simple explanation
	fmt.Printf("💡 What is it?\n")
	fmt.Printf("─────────────────────────────────────────────────────────────\n")
	fmt.Printf("%s\n\n", explanation.SimpleExplanation)

	// Technical details
	if explainDetailed && explanation.TechnicalDetails != "" {
		fmt.Printf("🔬 Technical Details\n")
		fmt.Printf("─────────────────────────────────────────────────────────────\n")
		fmt.Printf("%s\n\n", explanation.TechnicalDetails)
	}

	// Risk impact
	fmt.Printf("🎯 Risk Assessment\n")
	fmt.Printf("─────────────────────────────────────────────────────────────\n")
	fmt.Printf("Likelihood: %s | Impact: %s\n\n",
		explanation.RiskImpact.Likelihood,
		explanation.RiskImpact.Impact)

	if len(explanation.RiskImpact.Scenarios) > 0 {
		fmt.Printf("Attack Scenarios:\n")
		for i, scenario := range explanation.RiskImpact.Scenarios {
			fmt.Printf("  %d. %s\n", i+1, scenario)
		}
		fmt.Println()
	}

	if len(explanation.RiskImpact.AffectedData) > 0 {
		fmt.Printf("Data at Risk:\n")
		for _, data := range explanation.RiskImpact.AffectedData {
			fmt.Printf("  • %s\n", data)
		}
		fmt.Println()
	}

	// Real-world example
	if explanation.RealWorldExample != nil {
		fmt.Printf("🌍 Real-World Example\n")
		fmt.Printf("─────────────────────────────────────────────────────────────\n")
		fmt.Printf("Title: %s\n", explanation.RealWorldExample.Title)
		fmt.Printf("%s\n\n", explanation.RealWorldExample.Description)
		if explanation.RealWorldExample.Impact != "" {
			fmt.Printf("Impact: %s\n", explanation.RealWorldExample.Impact)
		}
		if explanation.RealWorldExample.Lesson != "" {
			fmt.Printf("Lesson: %s\n", explanation.RealWorldExample.Lesson)
		}
		fmt.Println()
	}

	// Remediation steps
	fmt.Printf("✅ How to Fix\n")
	fmt.Printf("─────────────────────────────────────────────────────────────\n")
	for _, step := range explanation.RemediationSteps {
		fmt.Printf("%d. %s [Priority: %s, Effort: %s]\n",
			step.Order, step.Action, step.Priority, step.Effort)
		fmt.Printf("   %s\n\n", step.Description)
	}

	// Code example
	if explanation.CodeExample != nil {
		fmt.Printf("💻 Code Example\n")
		fmt.Printf("─────────────────────────────────────────────────────────────\n")
		if explanation.CodeExample.Before != "" {
			fmt.Printf("Before (Vulnerable):\n```%s\n%s\n```\n\n",
				explanation.CodeExample.Language,
				explanation.CodeExample.Before)
		}
		fmt.Printf("After (Secure):\n```%s\n%s\n```\n\n",
			explanation.CodeExample.Language,
			explanation.CodeExample.After)
		fmt.Printf("Explanation: %s\n\n", explanation.CodeExample.Explanation)
	}

	// Metadata
	if verbose {
		fmt.Printf("─────────────────────────────────────────────────────────────\n")
		fmt.Printf("Source: %s | Confidence: %.0f%% | Tokens: %d | Time: %dms\n",
			explanation.Source,
			explanation.Confidence*100,
			explanation.TokensUsed,
			explanation.ResponseTime)
	}

	fmt.Println()
}

// explainFromLastScan explains all vulnerabilities from the last scan
func explainFromLastScan() {
	findings, err := loadAllFindings()
	if err != nil {
		exitWithError(fmt.Errorf("failed to load scan results: %w", err))
	}

	if len(findings) == 0 {
		fmt.Println("No vulnerabilities found in last scan")
		return
	}

	fmt.Printf("Found %d vulnerabilities. Generating explanations...\n\n", len(findings))

	for i, finding := range findings {
		fmt.Printf("[%d/%d] Explaining %s...\n", i+1, len(findings), finding.ID)
		explanation := generateExplanation(finding)
		if explanation != nil {
			displayExplanation(finding, explanation)
			fmt.Println("\n" + strings.Repeat("═", 65) + "\n")
		}
	}
}

// loadFindingByID loads a specific finding by ID from scan results
func loadFindingByID(id string) (*scanner.Finding, error) {
	return results.FindByID(id)
}

// loadAllFindings loads all findings from the last scan
func loadAllFindings() ([]*scanner.Finding, error) {
	resultsFile, err := results.Load()
	if err != nil {
		return nil, err
	}

	return resultsFile.Findings, nil
}

// detectLanguage detects the programming language from file extension
func detectLanguage(filename string) string {
	ext := filename[len(filename)-3:]
	switch ext {
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".go":
		return "go"
	case "ava":
		return "java"
	case ".rb":
		return "ruby"
	case "php":
		return "php"
	default:
		return "unknown"
	}
}

// Made with Bob
