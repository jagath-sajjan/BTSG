package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BanditScanner implements Python security scanning using Bandit
type BanditScanner struct {
	path    string
	version string
}

// NewBanditScanner creates a new Bandit scanner instance
func NewBanditScanner() *BanditScanner {
	return &BanditScanner{
		path:    "bandit",
		version: detectBanditVersion(),
	}
}

// Name returns the scanner name
func (s *BanditScanner) Name() string {
	return "bandit"
}

// Version returns the scanner version
func (s *BanditScanner) Version() string {
	return s.version
}

// Type returns the vulnerability type
func (s *BanditScanner) Type() string {
	return "code"
}

// IsAvailable checks if Bandit is installed
func (s *BanditScanner) IsAvailable() bool {
	cmd := exec.Command(s.path, "--version")
	return cmd.Run() == nil
}

// GetCommand returns the command to execute
func (s *BanditScanner) GetCommand(config *ScanConfig) []string {
	args := []string{
		s.path,
		"-f", "json",
		"-r",
	}

	if config.Verbose {
		args = append(args, "-v")
	}

	args = append(args, config.Path)
	return args
}

// Scan executes Bandit and captures output
func (s *BanditScanner) Scan(ctx context.Context, config *ScanConfig) (*RawScanResult, error) {
	startTime := time.Now()

	cmdArgs := s.GetCommand(config)
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &RawScanResult{
		Scanner:  s.Name(),
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(startTime),
		Error:    err,
	}, nil
}

// Normalize converts Bandit JSON output to unified format
func (s *BanditScanner) Normalize(raw *RawScanResult) ([]*Finding, error) {
	if raw.Stdout == "" {
		return []*Finding{}, nil
	}

	var banditOutput struct {
		Results []struct {
			TestID          string `json:"test_id"`
			TestName        string `json:"test_name"`
			IssueText       string `json:"issue_text"`
			IssueSeverity   string `json:"issue_severity"`
			IssueConfidence string `json:"issue_confidence"`
			Filename        string `json:"filename"`
			LineNumber      int    `json:"line_number"`
			LineRange       []int  `json:"line_range"`
			Code            string `json:"code"`
			CWE             struct {
				ID   int    `json:"id"`
				Link string `json:"link"`
			} `json:"cwe"`
		} `json:"results"`
	}

	if err := json.Unmarshal([]byte(raw.Stdout), &banditOutput); err != nil {
		return nil, fmt.Errorf("failed to parse bandit output: %w", err)
	}

	var findings []*Finding
	for _, result := range banditOutput.Results {
		finding := &Finding{
			ID:          generateID(s.Name(), result.Filename, result.LineNumber),
			Tool:        s.Name(),
			Severity:    mapBanditSeverity(result.IssueSeverity),
			File:        result.Filename,
			Line:        result.LineNumber,
			Description: fmt.Sprintf("[%s] %s: %s", result.TestID, result.TestName, result.IssueText),
			Code:        result.Code,
			Confidence:  result.IssueConfidence,
		}

		if result.CWE.ID > 0 {
			finding.CWE = fmt.Sprintf("CWE-%d", result.CWE.ID)
		}

		findings = append(findings, finding)
	}

	return findings, nil
}

// detectBanditVersion detects the installed Bandit version
func detectBanditVersion() string {
	cmd := exec.Command("bandit", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse version from output like "bandit 1.7.5"
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		return parts[1]
	}

	return "unknown"
}

// mapBanditSeverity maps Bandit severity to unified format
func mapBanditSeverity(severity string) string {
	switch strings.ToUpper(severity) {
	case "HIGH":
		return "HIGH"
	case "MEDIUM":
		return "MEDIUM"
	case "LOW":
		return "LOW"
	default:
		return "INFO"
	}
}

// Made with Bob
