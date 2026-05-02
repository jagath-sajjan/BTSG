package cmd

import (
	"btsg/internal/fixer"
	"fmt"

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
	Use:   "fix [path]",
	Short: "Automatically fix security vulnerabilities",
	Long: `Auto-fix security vulnerabilities with AI assistance:
  • Interactive mode for reviewing changes before applying
  • Dry-run mode to preview fixes without modifying files
  • Fix specific vulnerabilities or all at once
  • Backup original files before modification

Examples:
  btsg fix . --interactive
  btsg fix . --dry-run
  btsg fix . --all
  btsg fix . --vuln-id BTSG-001`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		if verbose {
			fmt.Printf("Fixing vulnerabilities in: %s\n", path)
			fmt.Printf("Interactive mode: %v\n", fixInteractive)
			fmt.Printf("Dry run: %v\n", fixDryRun)
		}

		// Initialize fixer
		f := fixer.New(fixer.Config{
			Path:        path,
			Interactive: fixInteractive,
			DryRun:      fixDryRun,
			FixAll:      fixAll,
			VulnID:      fixVulnID,
			Verbose:     verbose,
		})

		// Run fixer
		results, err := f.Fix()
		if err != nil {
			exitWithError(err)
		}

		// Display results
		displayFixResults(results)
	},
}

func init() {
	rootCmd.AddCommand(fixCmd)

	fixCmd.Flags().BoolVarP(&fixInteractive, "interactive", "i", false, "Review each fix before applying")
	fixCmd.Flags().BoolVar(&fixDryRun, "dry-run", false, "Preview fixes without modifying files")
	fixCmd.Flags().BoolVarP(&fixAll, "all", "a", false, "Fix all vulnerabilities without prompting")
	fixCmd.Flags().StringVar(&fixVulnID, "vuln-id", "", "Fix specific vulnerability by ID")
}

func displayFixResults(results *fixer.FixResults) {
	fmt.Printf("\n=== BTSG Fix Results ===\n\n")

	if results.DryRun {
		fmt.Println("🔍 DRY RUN MODE - No files were modified\n")
	}

	fmt.Printf("Vulnerabilities found: %d\n", results.TotalVulns)
	fmt.Printf("Fixes applied: %d\n", results.FixesApplied)
	fmt.Printf("Fixes skipped: %d\n", results.FixesSkipped)
	fmt.Printf("Fixes failed: %d\n\n", results.FixesFailed)

	if len(results.FixedVulns) > 0 {
		fmt.Println("✅ Fixed vulnerabilities:")
		for _, fix := range results.FixedVulns {
			fmt.Printf("  • %s in %s\n", fix.VulnID, fix.File)
			if verbose {
				fmt.Printf("    Change: %s\n", fix.Description)
			}
		}
		fmt.Println()
	}

	if len(results.SkippedVulns) > 0 {
		fmt.Println("⏭️  Skipped vulnerabilities:")
		for _, skip := range results.SkippedVulns {
			fmt.Printf("  • %s: %s\n", skip.VulnID, skip.Reason)
		}
		fmt.Println()
	}

	if len(results.FailedVulns) > 0 {
		fmt.Println("❌ Failed to fix:")
		for _, fail := range results.FailedVulns {
			fmt.Printf("  • %s: %s\n", fail.VulnID, fail.Error)
		}
		fmt.Println()
	}

	if results.BackupPath != "" {
		fmt.Printf("💾 Backup created at: %s\n", results.BackupPath)
	}
}

// Made with Bob
