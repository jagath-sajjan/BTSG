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

// DetectSecretsScanner implements secret detection using detect-secrets
type DetectSecretsScanner struct {
	path    string
	version string
}

// NewDetectSecretsScanner creates a new detect-secrets scanner instance
func NewDetectSecretsScanner() *DetectSecretsScanner {
	return &DetectSecretsScanner{
		path:    "detect-secrets",
		version: detectSecretsVersion(),
	}
}

// Name returns the scanner name
func (s *DetectSecretsScanner) Name() string {
	return "detect-secrets"
}

// Version returns the scanner version
func (s *DetectSecretsScanner) Version() string {
	return s.version
}

// Type returns the vulnerability type
func (s *DetectSecretsScanner) Type() string {
	return "secrets"
}

// IsAvailable checks if detect-secrets is installed
func (s *DetectSecretsScanner) IsAvailable() bool {
	cmd := exec.Command(s.path, "--version")
	return cmd.Run() == nil
}

// GetCommand returns the command to execute
func (s *DetectSecretsScanner) GetCommand(config *ScanConfig) []string {
	args := []string{
		s.path,
		"scan",
		"--all-files",
	}

	if config.Verbose {
		args = append(args, "-v")
	}

	args = append(args, config.Path)
	return args
}

// Scan executes detect-secrets and captures output
func (s *DetectSecretsScanner) Scan(ctx context.Context, config *ScanConfig) (*RawScanResult, error) {
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

// Normalize converts detect-secrets JSON output to unified format
func (s *DetectSecretsScanner) Normalize(raw *RawScanResult) ([]*Finding, error) {
	if raw.Stdout == "" {
		return []*Finding{}, nil
	}

	var secretsOutput struct {
		Results map[string][]struct {
			Type         string `json:"type"`
			LineNumber   int    `json:"line_number"`
			HashedSecret string `json:"hashed_secret"`
			IsVerified   bool   `json:"is_verified"`
		} `json:"results"`
	}

	if err := json.Unmarshal([]byte(raw.Stdout), &secretsOutput); err != nil {
		return nil, fmt.Errorf("failed to parse detect-secrets output: %w", err)
	}

	var findings []*Finding
	for filename, secrets := range secretsOutput.Results {
		for _, secret := range secrets {
			severity := "HIGH"
			if secret.IsVerified {
				severity = "CRITICAL"
			}

			finding := &Finding{
				ID:          generateID(s.Name(), filename, secret.LineNumber),
				Tool:        s.Name(),
				Severity:    severity,
				File:        filename,
				Line:        secret.LineNumber,
				Description: fmt.Sprintf("Potential %s detected in source code", secret.Type),
			}

			findings = append(findings, finding)
		}
	}

	return findings, nil
}

// detectSecretsVersion detects the installed detect-secrets version
func detectSecretsVersion() string {
	cmd := exec.Command("detect-secrets", "--version")
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
