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

// PipAuditScanner implements Python dependency scanning using pip-audit
type PipAuditScanner struct {
	path    string
	version string
}

// NewPipAuditScanner creates a new pip-audit scanner instance
func NewPipAuditScanner() *PipAuditScanner {
	return &PipAuditScanner{
		path:    "pip-audit",
		version: detectPipAuditVersion(),
	}
}

// Name returns the scanner name
func (s *PipAuditScanner) Name() string {
	return "pip-audit"
}

// Version returns the scanner version
func (s *PipAuditScanner) Version() string {
	return s.version
}

// Type returns the vulnerability type
func (s *PipAuditScanner) Type() string {
	return "dependencies"
}

// IsAvailable checks if pip-audit is installed
func (s *PipAuditScanner) IsAvailable() bool {
	cmd := exec.Command(s.path, "--version")
	return cmd.Run() == nil
}

// GetCommand returns the command to execute
func (s *PipAuditScanner) GetCommand(config *ScanConfig) []string {
	args := []string{
		s.path,
		"--format", "json",
		"--desc", "on",
	}

	if config.Verbose {
		args = append(args, "-v")
	}

	return args
}

// Scan executes pip-audit and captures output
func (s *PipAuditScanner) Scan(ctx context.Context, config *ScanConfig) (*RawScanResult, error) {
	startTime := time.Now()

	cmdArgs := s.GetCommand(config)
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = config.Path

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

// Normalize converts pip-audit JSON output to unified format
func (s *PipAuditScanner) Normalize(raw *RawScanResult) ([]*Finding, error) {
	if raw.Stdout == "" {
		return []*Finding{}, nil
	}

	var pipOutput struct {
		Dependencies []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Vulns   []struct {
				ID          string   `json:"id"`
				Description string   `json:"description"`
				FixVersions []string `json:"fix_versions"`
				Aliases     []string `json:"aliases"`
			} `json:"vulns"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal([]byte(raw.Stdout), &pipOutput); err != nil {
		return nil, fmt.Errorf("failed to parse pip-audit output: %w", err)
	}

	var findings []*Finding
	for _, dep := range pipOutput.Dependencies {
		for _, vuln := range dep.Vulns {
			severity := "MEDIUM"
			if strings.Contains(strings.ToLower(vuln.Description), "critical") {
				severity = "CRITICAL"
			} else if strings.Contains(strings.ToLower(vuln.Description), "high") {
				severity = "HIGH"
			}

			fixVersion := "N/A"
			if len(vuln.FixVersions) > 0 {
				fixVersion = vuln.FixVersions[0]
			}

			finding := &Finding{
				ID:       generateID(s.Name(), dep.Name, 0),
				Tool:     s.Name(),
				Severity: severity,
				File:     "requirements.txt",
				Line:     0,
				Description: fmt.Sprintf("[%s] Vulnerable dependency: %s@%s - %s (Fix: %s)",
					vuln.ID, dep.Name, dep.Version, vuln.Description, fixVersion),
			}

			findings = append(findings, finding)
		}
	}

	return findings, nil
}

// detectPipAuditVersion detects the installed pip-audit version
func detectPipAuditVersion() string {
	cmd := exec.Command("pip-audit", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse version from output
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		return parts[1]
	}

	return "unknown"
}

// Made with Bob
