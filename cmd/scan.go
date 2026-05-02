package cmd

import (
	"btsg/internal/scanner"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	scanPath      string
	scanRecursive bool
	scanTypes     []string
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
  btsg scan . --types secrets,dependencies`,
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
			fmt.Printf("Scan types: %v\n", scanTypes)
		}

		// Initialize scanner
		s := scanner.New(scanner.Config{
			Path:      scanPath,
			Recursive: scanRecursive,
			Types:     scanTypes,
			Verbose:   verbose,
		})

		// Run scan
		results, err := s.Scan()
		if err != nil {
			exitWithError(err)
		}

		// Display results
		if err := displayScanResults(results); err != nil {
			exitWithError(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVarP(&scanRecursive, "recursive", "r", true, "Scan directories recursively")
	scanCmd.Flags().StringSliceVarP(&scanTypes, "types", "t", []string{"all"}, "Vulnerability types to scan (secrets,dependencies,code,config,all)")
}

func displayScanResults(results *scanner.ScanResults) error {
	fmt.Printf("\n=== BTSG Security Scan Results ===\n\n")
	fmt.Printf("Scanned: %s\n", results.Path)
	fmt.Printf("Files scanned: %d\n", results.FilesScanned)
	fmt.Printf("Duration: %s\n\n", results.Duration)

	if len(results.Vulnerabilities) == 0 {
		fmt.Println("✓ No vulnerabilities found!")
		return nil
	}

	fmt.Printf("Found %d vulnerabilities:\n\n", len(results.Vulnerabilities))

	for i, vuln := range results.Vulnerabilities {
		fmt.Printf("%d. [%s] %s\n", i+1, vuln.Severity, vuln.Title)
		fmt.Printf("   File: %s:%d\n", vuln.File, vuln.Line)
		fmt.Printf("   Type: %s\n", vuln.Type)
		if vuln.CVE != "" {
			fmt.Printf("   CVE: %s\n", vuln.CVE)
		}
		fmt.Printf("   Description: %s\n\n", vuln.Description)
	}

	return nil
}

// Made with Bob
