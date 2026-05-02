package ui

import (
	"fmt"
	"strings"

	"btsg/internal/scanner"
)

// FindingsTable creates a formatted table for security findings
type FindingsTable struct {
	findings []*scanner.Finding
}

// NewFindingsTable creates a new findings table
func NewFindingsTable(findings []*scanner.Finding) *FindingsTable {
	return &FindingsTable{
		findings: findings,
	}
}

// Render renders the findings table with colors
func (ft *FindingsTable) Render() {
	if len(ft.findings) == 0 {
		PrintSuccess("No vulnerabilities found!")
		return
	}

	PrintHeader(fmt.Sprintf("Found %d Vulnerabilities", len(ft.findings)))

	// Print table header
	fmt.Printf("%-4s %-12s %-40s %-6s %-50s %-15s\n",
		"#", "Severity", "File", "Line", "Description", "Tool")
	PrintSeparator()

	for i, finding := range ft.findings {
		// Truncate description if too long
		desc := finding.Description
		if len(desc) > 47 {
			desc = desc[:44] + "..."
		}

		// Truncate file path if too long
		file := finding.File
		if len(file) > 37 {
			parts := strings.Split(file, "/")
			if len(parts) > 2 {
				file = ".../" + strings.Join(parts[len(parts)-2:], "/")
			}
			if len(file) > 37 {
				file = file[:34] + "..."
			}
		}

		// Format severity with icon
		sevIcon := SeverityIcon(finding.Severity)
		sevColor := Severity(finding.Severity)

		fmt.Printf("%-4d %s ", i+1, sevIcon)
		sevColor.Printf("%-10s", finding.Severity)
		fmt.Printf(" ")
		ColorPath.Printf("%-40s", file)
		fmt.Printf(" ")
		ColorNumber.Printf("%-6d", finding.Line)
		fmt.Printf(" %-50s ", desc)
		ColorInfo.Printf("%-15s\n", finding.Tool)
	}

	fmt.Println()
}

// RenderCompact renders a compact version of the findings table
func (ft *FindingsTable) RenderCompact() {
	if len(ft.findings) == 0 {
		PrintSuccess("No vulnerabilities found!")
		return
	}

	// Group by severity
	bySeverity := make(map[string]int)
	for _, f := range ft.findings {
		bySeverity[f.Severity]++
	}

	fmt.Println()
	ColorBold.Println("Vulnerability Summary:")
	fmt.Println()

	severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}
	for _, sev := range severities {
		if count, ok := bySeverity[sev]; ok && count > 0 {
			icon := SeverityIcon(sev)
			color := Severity(sev)
			fmt.Printf("  %s ", icon)
			color.Printf("%-8s", sev)
			fmt.Printf(": %s\n", FormatNumber(count))
		}
	}
	fmt.Println()
}

// RenderList renders findings as a simple list
func (ft *FindingsTable) RenderList() {
	if len(ft.findings) == 0 {
		PrintSuccess("No vulnerabilities found!")
		return
	}

	PrintHeader(fmt.Sprintf("Found %d Vulnerabilities", len(ft.findings)))

	for i, finding := range ft.findings {
		fmt.Printf("\n%s. ", FormatNumber(i+1))
		PrintSeverityBadge(finding.Severity)
		fmt.Printf(" %s\n", ColorBold.Sprint(finding.Description))
		fmt.Printf("   %s %s:%s\n",
			ColorDim.Sprint("Location:"),
			FormatPath(finding.File),
			FormatNumber(finding.Line))
		fmt.Printf("   %s %s\n",
			ColorDim.Sprint("Tool:"),
			finding.Tool)
		if finding.Code != "" {
			fmt.Printf("   %s\n", ColorDim.Sprint("Code:"))
			fmt.Printf("   %s\n", FormatCode(finding.Code))
		}
	}
	fmt.Println()
}

// SummaryTable creates a summary statistics table
type SummaryTable struct {
	data map[string]string
}

// NewSummaryTable creates a new summary table
func NewSummaryTable() *SummaryTable {
	return &SummaryTable{
		data: make(map[string]string),
	}
}

// Add adds a metric to the summary
func (st *SummaryTable) Add(metric, value string) {
	st.data[metric] = value
}

// Render renders the summary table
func (st *SummaryTable) Render() {
	PrintHeader("Scan Summary")

	fmt.Printf("%-30s %s\n", "Metric", "Value")
	PrintSeparator()

	for metric, value := range st.data {
		fmt.Printf("%-30s ", metric)
		ColorBold.Printf("%s\n", value)
	}
	fmt.Println()
}

// ScannerInfoTable displays information about available scanners
type ScannerInfoTable struct {
	scanners []map[string]string
}

// NewScannerInfoTable creates a new scanner info table
func NewScannerInfoTable(scanners []map[string]string) *ScannerInfoTable {
	return &ScannerInfoTable{
		scanners: scanners,
	}
}

// Render renders the scanner info table
func (sit *ScannerInfoTable) Render() {
	PrintHeader("Available Security Scanners")

	fmt.Printf("%-20s %-15s %-20s %s\n", "Scanner", "Version", "Type", "Status")
	PrintSeparator()

	for _, scanner := range sit.scanners {
		ColorBold.Printf("%-20s", scanner["name"])
		fmt.Printf(" ")
		ColorNumber.Printf("%-15s", scanner["version"])
		fmt.Printf(" ")
		ColorInfo.Printf("%-20s", scanner["type"])
		fmt.Printf(" ")
		ColorSuccess.Printf("✓ Available\n")
	}

	fmt.Println()
}

// Made with Bob
