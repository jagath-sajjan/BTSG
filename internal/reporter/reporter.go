package reporter

import (
	"btsg/internal/scanner"
	"btsg/pkg/results"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds reporter configuration
type Config struct {
	Format  string
	Output  string
	Input   string
	Verbose bool
}

// Reporter generates security reports in various formats
type Reporter struct {
	config Config
}

// New creates a new reporter instance
func New(config Config) *Reporter {
	return &Reporter{
		config: config,
	}
}

// ReportResult holds the generated report information
type ReportResult struct {
	Format     string
	OutputPath string
	FileSize   string
	Timestamp  string
	Content    string
	Summary    ReportSummary
}

// ReportSummary provides a summary of vulnerabilities
type ReportSummary struct {
	TotalVulns int
	Critical   int
	High       int
	Medium     int
	Low        int
	Info       int
}

// ReportData holds the complete report data
type ReportData struct {
	Metadata        ReportMetadata        `json:"metadata"`
	Summary         ReportSummary         `json:"summary"`
	Vulnerabilities []VulnerabilityReport `json:"vulnerabilities"`
}

// ReportMetadata contains report metadata
type ReportMetadata struct {
	Tool        string    `json:"tool"`
	Version     string    `json:"version"`
	ScanPath    string    `json:"scan_path"`
	GeneratedAt time.Time `json:"generated_at"`
	Duration    string    `json:"duration"`
}

// VulnerabilityReport represents a vulnerability in the report
type VulnerabilityReport struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Type        string   `json:"type"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	CVE         string   `json:"cve,omitempty"`
	CWE         string   `json:"cwe,omitempty"`
	CVSS        float64  `json:"cvss,omitempty"`
	References  []string `json:"references,omitempty"`
	Remediation string   `json:"remediation"`
}

// Generate creates a security report in the specified format
func (r *Reporter) Generate() (*ReportResult, error) {
	if r.config.Verbose {
		fmt.Printf("Generating %s report...\n", r.config.Format)
	}

	// Load scan results from .btsg/results.json
	reportData, err := r.loadReportData()
	if err != nil {
		return nil, fmt.Errorf("failed to load scan results: %w", err)
	}

	var content string

	switch r.config.Format {
	case "json":
		content, err = r.generateJSON(reportData)
	case "html":
		content, err = r.generateHTML(reportData)
	case "pdf":
		content, err = r.generatePDF(reportData)
	case "markdown":
		content, err = r.generateMarkdown(reportData)
	case "sarif":
		content, err = r.generateSARIF(reportData)
	default:
		return nil, fmt.Errorf("unsupported format: %s", r.config.Format)
	}

	if err != nil {
		return nil, err
	}

	result := &ReportResult{
		Format:    r.config.Format,
		Timestamp: time.Now().Format(time.RFC3339),
		Content:   content,
		Summary:   reportData.Summary,
	}

	// Save to file if output path is specified
	if r.config.Output != "" {
		if err := os.WriteFile(r.config.Output, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write report file: %w", err)
		}
		result.OutputPath = r.config.Output

		// Get file size
		if info, err := os.Stat(r.config.Output); err == nil {
			result.FileSize = fmt.Sprintf("%.2f KB", float64(info.Size())/1024)
		}
	}

	return result, nil
}

// generateJSON creates a JSON report
func (r *Reporter) generateJSON(data *ReportData) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// generateHTML creates an HTML report
func (r *Reporter) generateHTML(data *ReportData) (string, error) {
	// TODO: Implement HTML template generation
	html := `<!DOCTYPE html>
<html>
<head>
    <title>BTSG Security Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #2c3e50; color: white; padding: 20px; }
        .summary { display: flex; gap: 20px; margin: 20px 0; }
        .card { border: 1px solid #ddd; padding: 15px; border-radius: 5px; }
        .critical { color: #e74c3c; }
        .high { color: #e67e22; }
        .medium { color: #f39c12; }
        .low { color: #3498db; }
    </style>
</head>
<body>
    <div class="header">
        <h1>BTSG Security Report</h1>
        <p>Generated: ` + time.Now().Format(time.RFC3339) + `</p>
    </div>
    <div class="summary">
        <div class="card"><h3>Total</h3><p>` + fmt.Sprintf("%d", data.Summary.TotalVulns) + `</p></div>
        <div class="card critical"><h3>Critical</h3><p>` + fmt.Sprintf("%d", data.Summary.Critical) + `</p></div>
        <div class="card high"><h3>High</h3><p>` + fmt.Sprintf("%d", data.Summary.High) + `</p></div>
        <div class="card medium"><h3>Medium</h3><p>` + fmt.Sprintf("%d", data.Summary.Medium) + `</p></div>
        <div class="card low"><h3>Low</h3><p>` + fmt.Sprintf("%d", data.Summary.Low) + `</p></div>
    </div>
</body>
</html>`
	return html, nil
}

// generatePDF creates a PDF report
func (r *Reporter) generatePDF(data *ReportData) (string, error) {
	// TODO: Implement PDF generation using a library like gofpdf
	return "PDF generation not yet implemented", nil
}

// generateMarkdown creates a Markdown report
func (r *Reporter) generateMarkdown(data *ReportData) (string, error) {
	md := fmt.Sprintf(`# BTSG Security Report

**Generated:** %s
**Scan Path:** %s

## Summary

| Severity | Count |
|----------|-------|
| Critical | %d |
| High     | %d |
| Medium   | %d |
| Low      | %d |
| Info     | %d |
| **Total** | **%d** |

## Vulnerabilities

`,
		data.Metadata.GeneratedAt.Format(time.RFC3339),
		data.Metadata.ScanPath,
		data.Summary.Critical,
		data.Summary.High,
		data.Summary.Medium,
		data.Summary.Low,
		data.Summary.Info,
		data.Summary.TotalVulns,
	)

	for i, vuln := range data.Vulnerabilities {
		md += fmt.Sprintf(`### %d. [%s] %s

- **File:** %s:%d
- **Type:** %s
- **Description:** %s
- **Remediation:** %s

`,
			i+1,
			vuln.Severity,
			vuln.Title,
			vuln.File,
			vuln.Line,
			vuln.Type,
			vuln.Description,
			vuln.Remediation,
		)
	}

	return md, nil
}

// generateSARIF creates a SARIF format report
func (r *Reporter) generateSARIF(data *ReportData) (string, error) {
	// TODO: Implement SARIF format generation
	return "SARIF generation not yet implemented", nil
}

// loadReportData loads scan results and converts to report format
func (r *Reporter) loadReportData() (*ReportData, error) {
	// Load results from .btsg/results.json
	resultsFile, err := results.Load()
	if err != nil {
		return nil, err
	}

	// Calculate summary
	summary := ReportSummary{
		TotalVulns: len(resultsFile.Findings),
	}

	for _, finding := range resultsFile.Findings {
		switch finding.Severity {
		case "CRITICAL":
			summary.Critical++
		case "HIGH":
			summary.High++
		case "MEDIUM":
			summary.Medium++
		case "LOW":
			summary.Low++
		case "INFO":
			summary.Info++
		}
	}

	// Convert findings to vulnerability reports
	vulnerabilities := make([]VulnerabilityReport, 0, len(resultsFile.Findings))
	for _, finding := range resultsFile.Findings {
		vulnerabilities = append(vulnerabilities, VulnerabilityReport{
			ID:          finding.ID,
			Title:       finding.Description,
			Description: finding.Description,
			Severity:    finding.Severity,
			Type:        finding.Tool,
			File:        finding.File,
			Line:        finding.Line,
			CWE:         finding.CWE,
			Remediation: getRemediationForFinding(finding),
		})
	}

	return &ReportData{
		Metadata: ReportMetadata{
			Tool:        resultsFile.Metadata.Scanner,
			Version:     resultsFile.Metadata.Version,
			ScanPath:    resultsFile.Metadata.Path,
			GeneratedAt: time.Now(),
			Duration:    resultsFile.Metadata.Duration,
		},
		Summary:         summary,
		Vulnerabilities: vulnerabilities,
	}, nil
}

// getRemediationForFinding provides basic remediation advice
func getRemediationForFinding(finding *scanner.Finding) string {
	// Basic remediation suggestions based on tool/type
	switch finding.Tool {
	case "detect-secrets":
		return "Remove hardcoded secrets and use environment variables or a secrets management system"
	case "bandit":
		return "Review and fix the security issue identified by Bandit"
	case "pip-audit":
		return "Update the vulnerable dependency to a patched version"
	default:
		return "Review and remediate the security finding"
	}
}

// Made with Bob
