package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"btsg/internal/scanner"
)

func main() {
	fmt.Println("=== BTSG Scanner Example ===\n")

	// Example 1: Basic scan
	fmt.Println("Example 1: Basic Scan")
	basicScan()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 2: Verbose scan with statistics
	fmt.Println("Example 2: Verbose Scan with Statistics")
	verboseScan()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 3: JSON output
	fmt.Println("Example 3: JSON Output")
	jsonOutput()
}

// basicScan demonstrates a simple scan
func basicScan() {
	config := &scanner.ScanConfig{
		Path:      ".",
		Recursive: true,
		Verbose:   false,
		Timeout:   5 * time.Minute,
	}

	s := scanner.New(config)

	ctx := context.Background()
	results, err := s.Scan(ctx)
	if err != nil {
		log.Printf("Scan error: %v", err)
		return
	}

	fmt.Printf("Found %d findings in %s\n", results.TotalScanned, results.Duration)

	if len(results.Findings) > 0 {
		fmt.Println("\nTop 3 findings:")
		for i, finding := range results.Findings {
			if i >= 3 {
				break
			}
			fmt.Printf("  %d. [%s] %s\n", i+1, finding.Severity, finding.Description)
		}
	}
}

// verboseScan demonstrates a scan with detailed statistics
func verboseScan() {
	config := &scanner.ScanConfig{
		Path:      ".",
		Recursive: true,
		Verbose:   true,
		Timeout:   5 * time.Minute,
	}

	s := scanner.New(config)

	// List available scanners
	scanners := s.ListAvailableScanners()
	fmt.Printf("Available scanners: %v\n\n", scanners)

	ctx := context.Background()
	results, err := s.Scan(ctx)
	if err != nil {
		log.Printf("Scan error: %v", err)
		return
	}

	// Sort by severity
	scanner.SortFindingsBySeverity(results.Findings)

	// Get statistics
	severityCounts := scanner.CountBySeverity(results.Findings)
	toolCounts := scanner.CountByTool(results.Findings)

	fmt.Printf("\nScan Statistics:\n")
	fmt.Printf("  Duration: %s\n", results.Duration)
	fmt.Printf("  Total Findings: %d\n", results.TotalScanned)

	fmt.Printf("\nBy Severity:\n")
	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"} {
		if count, ok := severityCounts[sev]; ok && count > 0 {
			fmt.Printf("  %s: %d\n", sev, count)
		}
	}

	fmt.Printf("\nBy Tool:\n")
	for tool, count := range toolCounts {
		fmt.Printf("  %s: %d\n", tool, count)
	}

	if len(results.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, err := range results.Errors {
			fmt.Printf("  • %s\n", err)
		}
	}
}

// jsonOutput demonstrates JSON output format
func jsonOutput() {
	config := &scanner.ScanConfig{
		Path:      ".",
		Recursive: true,
		Verbose:   false,
		Timeout:   5 * time.Minute,
	}

	s := scanner.New(config)

	ctx := context.Background()
	results, err := s.Scan(ctx)
	if err != nil {
		log.Printf("Scan error: %v", err)
		return
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	fmt.Println("JSON Output:")
	fmt.Println(string(jsonData))
}

// Made with Bob
