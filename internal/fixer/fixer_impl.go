package fixer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"btsg/internal/explainer"
)

// fixerImpl implements the Fixer interface
type fixerImpl struct {
	config    *FixerConfig
	explainer explainer.Explainer
}

// NewFixer creates a new fixer instance
func NewFixer(config *FixerConfig, exp explainer.Explainer) (*fixerImpl, error) {
	if config == nil {
		config = DefaultFixerConfig()
	}

	// Create backup directory if it doesn't exist
	if config.CreateBackup {
		if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
			return nil, &FixError{
				Code:    ErrCodeBackupFailed,
				Message: "failed to create backup directory",
				Cause:   err,
			}
		}
	}

	return &fixerImpl{
		config:    config,
		explainer: exp,
	}, nil
}

// GenerateFix generates a fix for a vulnerability
func (f *fixerImpl) GenerateFix(req *FixRequest) (*Fix, error) {
	if req == nil || req.Finding == nil {
		return nil, &FixError{
			Code:    ErrCodeInvalidRequest,
			Message: "invalid request: finding is nil",
		}
	}

	// Read the file content
	content, err := os.ReadFile(req.Finding.File)
	if err != nil {
		return nil, &FixError{
			Code:    ErrCodeFileNotFound,
			Message: fmt.Sprintf("failed to read file %s", req.Finding.File),
			Cause:   err,
		}
	}

	lines := strings.Split(string(content), "\n")
	if req.Finding.Line < 1 || req.Finding.Line > len(lines) {
		return nil, &FixError{
			Code:    ErrCodeInvalidRequest,
			Message: fmt.Sprintf("invalid line number %d", req.Finding.Line),
		}
	}

	// Get the vulnerable code
	originalCode := lines[req.Finding.Line-1]

	// Generate fix using AI if enabled
	var fixedCode string
	var explanation string
	var confidence float64
	var source string

	if f.config.UseAI && f.explainer != nil {
		fixedCode, explanation, confidence, source, err = f.generateAIFix(req)
		if err != nil {
			// Fall back to template-based fix
			fixedCode, explanation, confidence = f.generateTemplateFix(req)
			source = "template"
		}
	} else {
		fixedCode, explanation, confidence = f.generateTemplateFix(req)
		source = "template"
	}

	// Generate diff
	diff := f.generateDiff(originalCode, fixedCode, req.Finding.Line)

	fix := &Fix{
		FindingID:    req.Finding.ID,
		File:         req.Finding.File,
		Line:         req.Finding.Line,
		Description:  fmt.Sprintf("Fix for %s", req.Finding.Description),
		Confidence:   confidence,
		OriginalCode: originalCode,
		FixedCode:    fixedCode,
		Diff:         diff,
		Explanation:  explanation,
		GeneratedAt:  time.Now(),
		Source:       source,
		Applied:      false,
	}

	return fix, nil
}

// generateAIFix generates a fix using AI
func (f *fixerImpl) generateAIFix(req *FixRequest) (fixedCode, explanation string, confidence float64, source string, err error) {
	// Create explanation request to get fix suggestion
	explainReq := &explainer.ExplanationRequest{
		Finding:     req.Finding,
		Language:    req.Language,
		IncludeCode: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	expl, err := f.explainer.Explain(ctx, explainReq)
	if err != nil {
		return "", "", 0, "", err
	}

	// Extract fixed code from code example
	if expl.CodeExample != nil && expl.CodeExample.After != "" {
		return expl.CodeExample.After,
			expl.CodeExample.Explanation,
			expl.Confidence,
			expl.Source,
			nil
	}

	return "", "", 0, "", fmt.Errorf("no code example in explanation")
}

// generateTemplateFix generates a fix using templates
func (f *fixerImpl) generateTemplateFix(req *FixRequest) (fixedCode, explanation string, confidence float64) {
	// Simple template-based fixes for common issues
	originalCode := req.Finding.Code
	if originalCode == "" {
		// Read from file if not provided
		content, err := os.ReadFile(req.Finding.File)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			if req.Finding.Line > 0 && req.Finding.Line <= len(lines) {
				originalCode = lines[req.Finding.Line-1]
			}
		}
	}

	// Apply template fixes based on vulnerability type
	switch {
	case strings.Contains(strings.ToLower(req.Finding.Description), "md5"):
		fixedCode = strings.ReplaceAll(originalCode, "md5", "sha256")
		explanation = "Replaced insecure MD5 hash with SHA-256"
		confidence = 0.8

	case strings.Contains(strings.ToLower(req.Finding.Description), "hardcoded"):
		fixedCode = "# TODO: Move to environment variable"
		explanation = "Hardcoded secret should be moved to environment variable"
		confidence = 0.6

	case strings.Contains(strings.ToLower(req.Finding.Description), "sql"):
		fixedCode = "# TODO: Use parameterized query"
		explanation = "SQL query should use parameterized statements"
		confidence = 0.7

	default:
		fixedCode = "# TODO: Fix required"
		explanation = "Manual fix required"
		confidence = 0.5
	}

	return fixedCode, explanation, confidence
}

// generateDiff generates a unified diff
func (f *fixerImpl) generateDiff(original, fixed string, lineNum int) string {
	var diff strings.Builder

	diff.WriteString(fmt.Sprintf("@@ -%d +%d @@\n", lineNum, lineNum))
	diff.WriteString(fmt.Sprintf("-%s\n", original))
	diff.WriteString(fmt.Sprintf("+%s\n", fixed))

	return diff.String()
}

// PreviewFix shows the diff without applying
func (f *fixerImpl) PreviewFix(fix *Fix) string {
	var preview strings.Builder

	preview.WriteString(fmt.Sprintf("\n╔═══════════════════════════════════════════════════════════════╗\n"))
	preview.WriteString(fmt.Sprintf("║  🔧 Proposed Fix for %s\n", fix.FindingID))
	preview.WriteString(fmt.Sprintf("╚═══════════════════════════════════════════════════════════════╝\n\n"))

	preview.WriteString(fmt.Sprintf("📁 File: %s:%d\n", fix.File, fix.Line))
	preview.WriteString(fmt.Sprintf("📝 Description: %s\n", fix.Description))
	preview.WriteString(fmt.Sprintf("🎯 Confidence: %.0f%%\n", fix.Confidence*100))
	preview.WriteString(fmt.Sprintf("🤖 Source: %s\n\n", fix.Source))

	preview.WriteString("💡 Explanation:\n")
	preview.WriteString(fmt.Sprintf("%s\n\n", fix.Explanation))

	preview.WriteString("📊 Diff:\n")
	preview.WriteString("─────────────────────────────────────────────────────────────\n")
	preview.WriteString(fix.Diff)
	preview.WriteString("─────────────────────────────────────────────────────────────\n\n")

	preview.WriteString("Before:\n")
	preview.WriteString(fmt.Sprintf("  %s\n\n", fix.OriginalCode))

	preview.WriteString("After:\n")
	preview.WriteString(fmt.Sprintf("  %s\n\n", fix.FixedCode))

	return preview.String()
}

// ApplyFix applies the fix to the file with automatic backup and rollback on failure
func (f *fixerImpl) ApplyFix(fix *Fix) (*FixResult, error) {
	result := &FixResult{
		Fix:       fix,
		AppliedAt: time.Now(),
	}

	// Validate fix before applying
	if err := f.ValidateFix(fix); err != nil {
		result.Success = false
		result.Error = err
		return result, err
	}

	// ===== STEP 1: CREATE BACKUP =====
	var backupPath string
	if f.config.CreateBackup {
		var err error
		backupPath, err = f.createBackup(fix.File)
		if err != nil {
			result.Success = false
			result.Error = &FixError{
				Code:    ErrCodeBackupFailed,
				Message: "failed to create backup",
				Cause:   err,
			}
			return result, result.Error
		}
		result.BackupPath = backupPath
		fix.BackupPath = backupPath
	}

	// ===== STEP 2: READ FILE =====
	content, err := os.ReadFile(fix.File)
	if err != nil {
		result.Success = false
		result.Error = &FixError{
			Code:    ErrCodeFileNotFound,
			Message: "failed to read file",
			Cause:   err,
		}
		// Backup exists but file unchanged
		return result, result.Error
	}

	// ===== STEP 3: APPLY FIX IN MEMORY =====
	lines := strings.Split(string(content), "\n")
	if fix.Line < 1 || fix.Line > len(lines) {
		result.Success = false
		result.Error = &FixError{
			Code:    ErrCodeInvalidRequest,
			Message: "invalid line number",
		}
		// Backup exists but file unchanged
		return result, result.Error
	}

	lines[fix.Line-1] = fix.FixedCode
	newContent := strings.Join(lines, "\n")

	// ===== STEP 4: WRITE FILE =====
	if err := os.WriteFile(fix.File, []byte(newContent), 0644); err != nil {
		result.Success = false
		result.Error = &FixError{
			Code:    ErrCodeApplyFailed,
			Message: "failed to write file",
			Cause:   err,
		}

		// ===== AUTOMATIC ROLLBACK ON WRITE FAILURE =====
		if backupPath != "" {
			fmt.Printf("⚠️  Write failed, attempting automatic rollback...\n")
			if rollbackErr := f.RollbackFix(backupPath); rollbackErr != nil {
				fmt.Printf("❌ Rollback failed: %v\n", rollbackErr)
				fmt.Printf("💾 Backup preserved at: %s\n", backupPath)
			} else {
				fmt.Printf("✅ File restored from backup\n")
			}
		}

		return result, result.Error
	}

	// ===== STEP 5: SUCCESS - UPDATE STATUS =====
	fix.Applied = true
	now := time.Now()
	fix.AppliedAt = &now

	result.Success = true
	return result, nil
}

// createBackup creates a backup of the file before modification
func (f *fixerImpl) createBackup(filePath string) (string, error) {
	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Base(filePath)
	backupPath := filepath.Join(f.config.BackupDir, fmt.Sprintf("%s.%s.backup", filename, timestamp))

	// Copy original file to backup
	source, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return backupPath, nil
}

// RollbackFix reverts a fix by restoring from backup
func (f *fixerImpl) RollbackFix(backupPath string) error {
	if backupPath == "" {
		return &FixError{
			Code:    ErrCodeInvalidRequest,
			Message: "backup path is empty",
		}
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return &FixError{
			Code:    ErrCodeFileNotFound,
			Message: fmt.Sprintf("backup file not found: %s", backupPath),
		}
	}

	// Extract original filename from backup path
	// Format: filename.YYYYMMDD-HHMMSS.backup
	filename := filepath.Base(backupPath)
	parts := strings.Split(filename, ".")
	if len(parts) < 3 {
		return &FixError{
			Code:    ErrCodeInvalidRequest,
			Message: "invalid backup filename format",
		}
	}

	// Reconstruct original filename (remove timestamp and .backup)
	originalFile := strings.Join(parts[:len(parts)-2], ".")
	originalPath := filepath.Join(filepath.Dir(filepath.Dir(backupPath)), originalFile)

	// Copy backup back to original location
	source, err := os.Open(backupPath)
	if err != nil {
		return &FixError{
			Code:    ErrCodeRollbackFailed,
			Message: "failed to open backup file",
			Cause:   err,
		}
	}
	defer source.Close()

	dest, err := os.Create(originalPath)
	if err != nil {
		return &FixError{
			Code:    ErrCodeRollbackFailed,
			Message: "failed to create original file",
			Cause:   err,
		}
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return &FixError{
			Code:    ErrCodeRollbackFailed,
			Message: "failed to copy backup",
			Cause:   err,
		}
	}

	return nil
}

// ValidateFix checks if a fix is safe to apply
func (f *fixerImpl) ValidateFix(fix *Fix) error {
	if fix == nil {
		return &FixError{
			Code:    ErrCodeInvalidRequest,
			Message: "fix is nil",
		}
	}

	// Check confidence threshold
	if fix.Confidence < f.config.MinConfidence {
		return &FixError{
			Code:    ErrCodeValidationFailed,
			Message: fmt.Sprintf("confidence %.2f below threshold %.2f", fix.Confidence, f.config.MinConfidence),
		}
	}

	// Check file exists
	if _, err := os.Stat(fix.File); os.IsNotExist(err) {
		return &FixError{
			Code:    ErrCodeFileNotFound,
			Message: fmt.Sprintf("file not found: %s", fix.File),
		}
	}

	// Check line count changes
	originalLines := strings.Count(fix.OriginalCode, "\n") + 1
	fixedLines := strings.Count(fix.FixedCode, "\n") + 1
	lineChanges := abs(fixedLines - originalLines)

	if lineChanges > f.config.MaxLineChanges {
		return &FixError{
			Code:    ErrCodeValidationFailed,
			Message: fmt.Sprintf("line changes %d exceed maximum %d", lineChanges, f.config.MaxLineChanges),
		}
	}

	return nil
}

// PromptUser prompts the user for confirmation
func PromptUser(message string) (bool, error) {
	fmt.Printf("%s (y/n): ", message)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// abs returns the absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Made with Bob
