package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose bool
	output  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "btsg",
	Short: "Bob The Security Guy - Your AI-powered security scanner",
	Long: `BTSG (Bob The Security Guy) is a production-ready CLI security tool that:
  • Scans local repositories for vulnerabilities
  • Explains security issues using AI
  • Auto-fixes issues with your approval
  • Generates structured security reports

Example usage:
  btsg scan ./myproject
  btsg explain CVE-2024-1234
  btsg fix --interactive
  btsg report --format json`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output file path (default: stdout)")
}

// exitWithError prints error and exits
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

// Made with Bob
