package cmd

import (
	"btsg/internal/scanner"
	"btsg/pkg/results"
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	scanPath      string
	scanRecursive bool
	scanTimeout   time.Duration
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a repository for security vulnerabilities",
	Long: `Scan local repositories for security vulnerabilities including:
  • Dependency vulnerabilities (CVEs)
  • Secret leaks (API keys, tokens, passwords)
  • Code security issues (SQL injection, XSS, etc.)
  • Misconfigurations (Docker, K8s, Terraform)

Examples:
  btsg scan .
  btsg scan /path/to/project
  btsg scan . --recursive
  btsg scan . --timeout 10m`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Default to current directory if no path provided
		if len(args) == 0 {
			scanPath = "."
		} else {
			scanPath = args[0]
		}

		if verbose {
			fmt.Printf("Scanning path: %s\n", scanPath)
			fmt.Printf("Recursive: %v\n", scanRecursive)
			fmt.Printf("Timeout: %v\n", scanTimeout)
		}

		// Initialize scanner
		s := scanner.New(&scanner.ScanConfig{
			Path:      scanPath,
			Recursive: scanRecursive,
			Verbose:   verbose,
			Timeout:   scanTimeout,
		})

		// List available scanners
		if verbose {
			scanners := s.ListAvailableScanners()
			fmt.Printf("Available scanners: %v\n\n", scanners)
		}

		// Run scan
		ctx := context.Background()
		scanResults, err := s.Scan(ctx)
		if err != nil {
			exitWithError(err)
		}

		// Save results to .btsg/results.json
		if err := results.Save(scanResults, scanPath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to save results: %v\n", err)
		} else {
			fmt.Printf("✓ Results saved to %s\n\n", results.GetResultsPath())
		}

		// Display results
		if err := displayScanResults(scanResults); err != nil {
			exitWithError(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVarP(&scanRecursive, "recursive", "r", true, "Scan directories recursively")
	scanCmd.Flags().DurationVar(&scanTimeout, "timeout", 5*time.Minute, "Timeout for each scanner")
}

func displayScanResults(results *scanner.ScanResults) error {
	fmt.Printf("\n=== BTSG Security Scan Results ===\n\n")
	fmt.Printf("Duration: %s\n", results.Duration)
	fmt.Printf("Total findings: %d\n\n", results.TotalScanned)

	// Display errors if any
	if len(results.Errors) > 0 {
		fmt.Printf("⚠️  Errors encountered:\n")
		for _, err := range results.Errors {
			fmt.Printf("   • %s\n", err)
		}
		fmt.Println()
	}

	if len(results.Findings) == 0 {
		fmt.Println("✓ No vulnerabilities found!")
		return nil
	}

	// Sort findings by severity
	scanner.SortFindingsBySeverity(results.Findings)

	// Display findings grouped by severity
	severityCounts := scanner.CountBySeverity(results.Findings)
	fmt.Printf("Findings by severity:\n")
	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"} {
		if count, ok := severityCounts[sev]; ok && count > 0 {
			fmt.Printf("  %s: %d\n", sev, count)
		}
	}
	fmt.Println()

	// Display findings by tool
	toolCounts := scanner.CountByTool(results.Findings)
	fmt.Printf("Findings by tool:\n")
	for tool, count := range toolCounts {
		fmt.Printf("  %s: %d\n", tool, count)
	}
	fmt.Println()

	// Display detailed findings with IDs
	fmt.Printf("Detailed findings:\n\n")
	for _, finding := range results.Findings {
		fmt.Printf("[%s] [%s] %s in %s", finding.ID, finding.Severity, finding.Description, finding.File)
		if finding.Line > 0 {
			fmt.Printf(":%d", finding.Line)
		}
		fmt.Println()
		if finding.CWE != "" {
			fmt.Printf("   CWE: %s\n", finding.CWE)
		}
		if finding.Confidence != "" {
			fmt.Printf("   Confidence: %s\n", finding.Confidence)
		}
		if finding.Code != "" && verbose {
			fmt.Printf("   Code:\n%s\n", finding.Code)
		}
		fmt.Println()
	}

	// Display actionable next steps
	displayNextSteps(results.Findings)

	return nil
}

// displayNextSteps shows actionable next steps after scan
func displayNextSteps(findings []*scanner.Finding) {
	if len(findings) == 0 {
		return
	}

	// Count high severity issues
	highCount := 0
	criticalCount := 0
	for _, f := range findings {
		if f.Severity == "HIGH" {
			highCount++
		} else if f.Severity == "CRITICAL" {
			criticalCount++
		}
	}

	fmt.Println("─────────────────────────────────────────────────────────────")

	if criticalCount > 0 {
		fmt.Printf("🚨 %d CRITICAL severity issues detected\n", criticalCount)
	}
	if highCount > 0 {
		fmt.Printf("⚠️  %d HIGH severity issues detected\n", highCount)
	}

	fmt.Println("\n👉 Next steps:")
	fmt.Println("   btsg explain")
	fmt.Println("   btsg fix --dry-run")
	fmt.Println("   btsg report --format markdown")
	fmt.Println()
}

// Made with Bob
