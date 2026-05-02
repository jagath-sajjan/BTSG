package cmd

import (
	"btsg/internal/scanner"
	"btsg/pkg/results"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	fixInteractive bool
	fixDryRun      bool
	fixAll         bool
	fixVulnID      string
)

// fixCmd represents the fix command
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Automatically fix security vulnerabilities",
	Long: `Auto-fix security vulnerabilities with simple pattern-based fixes:
  • Hardcoded secrets → environment variables
  • debug=True → debug=False
  • Dry-run mode to preview fixes without modifying files
  • Fix specific vulnerabilities by ID

Examples:
  btsg fix --id BTSG-001 --dry-run
  btsg fix --id BTSG-001
  btsg fix --all --dry-run`,
	Run: func(cmd *cobra.Command, args []string) {
		if fixVulnID == "" && !fixAll {
			fmt.Println("Error: Please specify --id or --all")
			cmd.Usage()
			return
		}

		if verbose {
			fmt.Printf("Dry run: %v\n", fixDryRun)
		}

		if fixAll {
			fixAllVulnerabilities()
		} else {
			fixSingleVulnerability(fixVulnID)
		}
	},
}

func init() {
	rootCmd.AddCommand(fixCmd)

	fixCmd.Flags().BoolVar(&fixDryRun, "dry-run", false, "Preview fixes without modifying files")
	fixCmd.Flags().BoolVarP(&fixAll, "all", "a", false, "Fix all vulnerabilities")
	fixCmd.Flags().StringVar(&fixVulnID, "id", "", "Fix specific vulnerability by ID")
}

// fixSingleVulnerability fixes a single vulnerability by ID
func fixSingleVulnerability(id string) {
	finding, err := results.FindByID(id)
	if err != nil {
		exitWithError(err)
	}

	fmt.Printf("\n🔧 Fixing vulnerability: %s\n", id)
	fmt.Printf("File: %s:%d\n", finding.File, finding.Line)
	fmt.Printf("Description: %s\n\n", finding.Description)

	fix := generateFix(finding)
	if fix == nil {
		fmt.Println("❌ No automatic fix available for this vulnerability type")
		return
	}

	displayFix(fix)

	if !fixDryRun {
		if err := applyFix(fix); err != nil {
			exitWithError(fmt.Errorf("failed to apply fix: %w", err))
		}
		fmt.Println("\n✅ Fix applied successfully!")
	} else {
		fmt.Println("\n💡 This is a dry-run. No changes were made.")
		fmt.Println("   Remove --dry-run to apply the fix.")
	}
}

// fixAllVulnerabilities fixes all vulnerabilities
func fixAllVulnerabilities() {
	resultsFile, err := results.Load()
	if err != nil {
		exitWithError(err)
	}

	fmt.Printf("\n🔧 Fixing %d vulnerabilities...\n\n", len(resultsFile.Findings))

	fixed := 0
	skipped := 0

	for _, finding := range resultsFile.Findings {
		fix := generateFix(finding)
		if fix == nil {
			skipped++
			continue
		}

		fmt.Printf("[%s] %s\n", finding.ID, finding.Description)
		displayFix(fix)

		if !fixDryRun {
			if err := applyFix(fix); err != nil {
				fmt.Printf("❌ Failed to apply fix: %v\n\n", err)
				continue
			}
			fmt.Println("✅ Fixed\n")
			fixed++
		} else {
			fmt.Println("💡 Dry-run mode\n")
		}
	}

	fmt.Printf("\nSummary: %d fixed, %d skipped\n", fixed, skipped)
	if fixDryRun {
		fmt.Println("This was a dry-run. Remove --dry-run to apply fixes.")
	}
}

// Fix represents a code fix
type Fix struct {
	Finding *scanner.Finding
	Before  string
	After   string
	Pattern string
}

// generateFix generates a fix for a finding
func generateFix(finding *scanner.Finding) *Fix {
	// Read the file
	content, err := os.ReadFile(finding.File)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	if finding.Line < 1 || finding.Line > len(lines) {
		return nil
	}

	lineContent := lines[finding.Line-1]

	// Pattern 1: Hardcoded secrets → environment variables
	if strings.Contains(finding.Tool, "detect-secrets") ||
		strings.Contains(strings.ToLower(finding.Description), "secret") ||
		strings.Contains(strings.ToLower(finding.Description), "api key") ||
		strings.Contains(strings.ToLower(finding.Description), "password") {

		// Look for common patterns like api_key = "xxx", password = "xxx", etc.
		secretPattern := regexp.MustCompile(`(\w+)\s*=\s*["']([^"']+)["']`)
		if matches := secretPattern.FindStringSubmatch(lineContent); len(matches) > 0 {
			varName := matches[1]
			envVarName := strings.ToUpper(varName)

			// Generate fix
			after := secretPattern.ReplaceAllString(lineContent,
				fmt.Sprintf(`%s = os.getenv("%s")  # TODO: Set %s in environment`,
					varName, envVarName, envVarName))

			return &Fix{
				Finding: finding,
				Before:  lineContent,
				After:   after,
				Pattern: "Hardcoded secret → Environment variable",
			}
		}
	}

	// Pattern 2: debug=True → debug=False
	if strings.Contains(strings.ToLower(finding.Description), "debug") {
		debugPattern := regexp.MustCompile(`(?i)(debug\s*=\s*)True`)
		if debugPattern.MatchString(lineContent) {
			after := debugPattern.ReplaceAllString(lineContent, "${1}False")
			return &Fix{
				Finding: finding,
				Before:  lineContent,
				After:   after,
				Pattern: "debug=True → debug=False",
			}
		}
	}

	return nil
}

// displayFix displays a fix with before/after diff
func displayFix(fix *Fix) {
	fmt.Printf("Pattern: %s\n", fix.Pattern)
	fmt.Printf("─────────────────────────────────────────────────────────────\n")
	fmt.Printf("Before:\n")
	fmt.Printf("  %s\n", strings.TrimSpace(fix.Before))
	fmt.Printf("\nAfter:\n")
	fmt.Printf("  %s\n", strings.TrimSpace(fix.After))
	fmt.Printf("─────────────────────────────────────────────────────────────\n")
}

// applyFix applies a fix to the file
func applyFix(fix *Fix) error {
	// Read the file
	content, err := os.ReadFile(fix.Finding.File)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	if fix.Finding.Line < 1 || fix.Finding.Line > len(lines) {
		return fmt.Errorf("invalid line number: %d", fix.Finding.Line)
	}

	// Replace the line
	lines[fix.Finding.Line-1] = fix.After

	// Write back to file
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(fix.Finding.File, []byte(newContent), 0644)
}

// Made with Bob
