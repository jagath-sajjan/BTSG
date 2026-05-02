package cmd

import (
	"btsg/internal/reporter"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	reportFormat string
	reportOutput string
	reportInput  string
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate structured security reports",
	Long: `Generate comprehensive security reports in various formats:
  • JSON - Machine-readable format for CI/CD integration
  • HTML - Interactive web-based report with charts
  • PDF - Professional report for stakeholders
  • Markdown - Documentation-friendly format
  • SARIF - Standard format for security tools

Examples:
  btsg report --format json
  btsg report --format html --output report.html
  btsg report --format pdf --output security-report.pdf
  btsg report --input scan-results.json --format html`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Printf("Generating report in %s format\n", reportFormat)
			if reportInput != "" {
				fmt.Printf("Using input file: %s\n", reportInput)
			}
		}

		// Initialize reporter
		r := reporter.New(reporter.Config{
			Format:  reportFormat,
			Output:  reportOutput,
			Input:   reportInput,
			Verbose: verbose,
		})

		// Generate report
		result, err := r.Generate()
		if err != nil {
			exitWithError(err)
		}

		// Display result
		displayReportResult(result)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "json", "Report format (json,html,pdf,markdown,sarif)")
	reportCmd.Flags().StringVar(&reportOutput, "output", "", "Output file path (default: stdout for json/markdown)")
	reportCmd.Flags().StringVar(&reportInput, "input", "", "Input scan results file (if not provided, runs new scan)")
}

func displayReportResult(result *reporter.ReportResult) {
	fmt.Printf("\n=== BTSG Security Report ===\n\n")
	
	fmt.Printf("Report format: %s\n", result.Format)
	fmt.Printf("Generated at: %s\n", result.Timestamp)
	
	if result.OutputPath != "" {
		fmt.Printf("Report saved to: %s\n", result.OutputPath)
		fmt.Printf("File size: %s\n", result.FileSize)
	}
	
	fmt.Printf("\nReport Summary:\n")
	fmt.Printf("  Total vulnerabilities: %d\n", result.Summary.TotalVulns)
	fmt.Printf("  Critical: %d\n", result.Summary.Critical)
	fmt.Printf("  High: %d\n", result.Summary.High)
	fmt.Printf("  Medium: %d\n", result.Summary.Medium)
	fmt.Printf("  Low: %d\n", result.Summary.Low)
	fmt.Printf("  Info: %d\n\n", result.Summary.Info)
	
	if result.Format == "json" && result.OutputPath == "" {
		fmt.Println("JSON Report:")
		fmt.Println(result.Content)
	}
	
	if result.Format == "html" {
		fmt.Printf("\n💡 Open the report in your browser:\n")
		fmt.Printf("   file://%s\n", result.OutputPath)
	}
}

// Made with Bob
