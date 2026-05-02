package scanner

import (
	"fmt"
	"time"
)

// Config holds scanner configuration
type Config struct {
	Path      string
	Recursive bool
	Types     []string
	Verbose   bool
}

// Scanner performs security scans
type Scanner struct {
	config Config
}

// New creates a new scanner instance
func New(config Config) *Scanner {
	return &Scanner{
		config: config,
	}
}

// ScanResults holds the results of a security scan
type ScanResults struct {
	Path            string
	FilesScanned    int
	Duration        time.Duration
	Vulnerabilities []Vulnerability
	Timestamp       time.Time
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string
	Title       string
	Description string
	Severity    string // CRITICAL, HIGH, MEDIUM, LOW, INFO
	Type        string // secrets, dependencies, code, config
	File        string
	Line        int
	Column      int
	CVE         string
	CWE         string
	CVSS        float64
	Remediation string
}

// Scan performs the security scan
func (s *Scanner) Scan() (*ScanResults, error) {
	startTime := time.Now()

	if s.config.Verbose {
		fmt.Printf("Starting scan of %s...\n", s.config.Path)
	}

	// TODO: Implement actual scanning logic
	// This is a placeholder that demonstrates the structure
	results := &ScanResults{
		Path:         s.config.Path,
		FilesScanned: 0,
		Duration:     time.Since(startTime),
		Timestamp:    time.Now(),
		Vulnerabilities: []Vulnerability{
			// Example vulnerability for demonstration
			{
				ID:          "BTSG-001",
				Title:       "Hardcoded API Key Detected",
				Description: "An API key was found hardcoded in the source code",
				Severity:    "HIGH",
				Type:        "secrets",
				File:        "config/api.go",
				Line:        15,
				Column:      20,
				Remediation: "Move API keys to environment variables or a secure vault",
			},
		},
	}

	if s.config.Verbose {
		fmt.Printf("Scan completed in %s\n", results.Duration)
	}

	return results, nil
}

// ScanFile scans a single file for vulnerabilities
func (s *Scanner) ScanFile(path string) ([]Vulnerability, error) {
	// TODO: Implement file scanning logic
	return nil, nil
}

// ValidateScanTypes checks if the provided scan types are valid
func ValidateScanTypes(types []string) error {
	validTypes := map[string]bool{
		"all":          true,
		"secrets":      true,
		"dependencies": true,
		"code":         true,
		"config":       true,
	}

	for _, t := range types {
		if !validTypes[t] {
			return fmt.Errorf("invalid scan type: %s", t)
		}
	}

	return nil
}

// Made with Bob
