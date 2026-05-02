package cmd

import (
	"btsg/internal/analyzer"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	explainDetailed bool
	explainCVE      string
)

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain [vulnerability-id]",
	Short: "Get AI-powered explanations for security vulnerabilities",
	Long: `Use AI to explain security vulnerabilities in plain language:
  • What the vulnerability is
  • Why it's dangerous
  • How it can be exploited
  • Recommended fixes

Examples:
  btsg explain CVE-2024-1234
  btsg explain BTSG-001 --detailed
  btsg explain --cve CVE-2024-1234`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var vulnID string

		if len(args) > 0 {
			vulnID = args[0]
		} else if explainCVE != "" {
			vulnID = explainCVE
		} else {
			fmt.Println("Error: Please provide a vulnerability ID or CVE")
			cmd.Usage()
			return
		}

		if verbose {
			fmt.Printf("Explaining vulnerability: %s\n", vulnID)
			fmt.Printf("Detailed mode: %v\n", explainDetailed)
		}

		// Initialize analyzer
		a := analyzer.New(analyzer.Config{
			Verbose:  verbose,
			Detailed: explainDetailed,
		})

		// Get explanation
		explanation, err := a.Explain(vulnID)
		if err != nil {
			exitWithError(err)
		}

		// Display explanation
		displayExplanation(vulnID, explanation)
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)

	explainCmd.Flags().BoolVarP(&explainDetailed, "detailed", "d", false, "Show detailed technical explanation")
	explainCmd.Flags().StringVar(&explainCVE, "cve", "", "CVE identifier to explain")
}

func displayExplanation(vulnID string, explanation *analyzer.Explanation) {
	fmt.Printf("\n=== Vulnerability Explanation: %s ===\n\n", vulnID)
	
	fmt.Printf("📋 Summary:\n%s\n\n", explanation.Summary)
	
	fmt.Printf("⚠️  Severity: %s\n", explanation.Severity)
	fmt.Printf("🔍 Type: %s\n\n", explanation.Type)
	
	fmt.Printf("💡 What is it?\n%s\n\n", explanation.Description)
	
	fmt.Printf("🎯 Impact:\n%s\n\n", explanation.Impact)
	
	if explanation.Exploitation != "" {
		fmt.Printf("🔓 How it can be exploited:\n%s\n\n", explanation.Exploitation)
	}
	
	fmt.Printf("✅ Recommended Fix:\n%s\n\n", explanation.Fix)
	
	if len(explanation.References) > 0 {
		fmt.Printf("📚 References:\n")
		for _, ref := range explanation.References {
			fmt.Printf("  • %s\n", ref)
		}
		fmt.Println()
	}
}

// Made with Bob
