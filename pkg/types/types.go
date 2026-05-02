package types

import "time"

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    Severity  `json:"severity"`
	Type        VulnType  `json:"type"`
	File        string    `json:"file"`
	Line        int       `json:"line"`
	Column      int       `json:"column"`
	CVE         string    `json:"cve,omitempty"`
	CWE         string    `json:"cwe,omitempty"`
	CVSS        float64   `json:"cvss,omitempty"`
	Remediation string    `json:"remediation"`
	References  []string  `json:"references,omitempty"`
	DetectedAt  time.Time `json:"detected_at"`
}

// Severity represents vulnerability severity levels
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// VulnType represents types of vulnerabilities
type VulnType string

const (
	VulnTypeSecrets      VulnType = "secrets"
	VulnTypeDependencies VulnType = "dependencies"
	VulnTypeCode         VulnType = "code"
	VulnTypeConfig       VulnType = "config"
)

// ScanConfig holds configuration for scanning
type ScanConfig struct {
	Path      string
	Recursive bool
	Types     []VulnType
	Exclude   []string
	Verbose   bool
}

// FixConfig holds configuration for fixing vulnerabilities
type FixConfig struct {
	Interactive bool
	DryRun      bool
	BackupDir   string
	AutoApprove bool
}

// ReportConfig holds configuration for report generation
type ReportConfig struct {
	Format      ReportFormat
	OutputPath  string
	IncludeCode bool
	Template    string
}

// ReportFormat represents supported report formats
type ReportFormat string

const (
	ReportFormatJSON     ReportFormat = "json"
	ReportFormatHTML     ReportFormat = "html"
	ReportFormatPDF      ReportFormat = "pdf"
	ReportFormatMarkdown ReportFormat = "markdown"
	ReportFormatSARIF    ReportFormat = "sarif"
)

// Made with Bob
