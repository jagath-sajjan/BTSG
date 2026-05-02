package fixer

import (
	"time"

	"btsg/internal/scanner"
)

// FixRequest represents a request to fix a vulnerability
type FixRequest struct {
	Finding      *scanner.Finding
	Language     string
	AutoApply    bool // If true, apply without confirmation
	CreateBackup bool // If true, create backup before applying
}

// Fix represents a proposed fix for a vulnerability
type Fix struct {
	// Identification
	FindingID string `json:"finding_id"`
	File      string `json:"file"`
	Line      int    `json:"line"`

	// Fix details
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"` // 0.0-1.0

	// Code changes
	OriginalCode string `json:"original_code"`
	FixedCode    string `json:"fixed_code"`
	Diff         string `json:"diff"`

	// Context
	Explanation string   `json:"explanation"`
	References  []string `json:"references,omitempty"`

	// Metadata
	GeneratedAt time.Time  `json:"generated_at"`
	Source      string     `json:"source"` // "ai", "template", "manual"
	Applied     bool       `json:"applied"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
	BackupPath  string     `json:"backup_path,omitempty"`
}

// FixResult represents the result of applying a fix
type FixResult struct {
	Success    bool      `json:"success"`
	Fix        *Fix      `json:"fix"`
	Error      error     `json:"error,omitempty"`
	BackupPath string    `json:"backup_path,omitempty"`
	AppliedAt  time.Time `json:"applied_at"`
}

// Fixer is the main interface for fixing vulnerabilities
type Fixer interface {
	// GenerateFix generates a fix for a vulnerability
	GenerateFix(req *FixRequest) (*Fix, error)

	// PreviewFix shows the diff without applying
	PreviewFix(fix *Fix) string

	// ApplyFix applies the fix to the file
	ApplyFix(fix *Fix) (*FixResult, error)

	// RollbackFix reverts a fix using backup
	RollbackFix(backupPath string) error

	// ValidateFix checks if a fix is safe to apply
	ValidateFix(fix *Fix) error
}

// FixerConfig holds configuration for the fixer
type FixerConfig struct {
	// AI settings
	UseAI      bool
	AIProvider string
	AIModel    string

	// Safety settings
	RequireConfirmation bool
	CreateBackup        bool
	BackupDir           string
	DryRun              bool

	// Validation
	MaxLineChanges int     // Maximum lines that can be changed
	MinConfidence  float64 // Minimum confidence to suggest fix

	// Features
	EnableTemplates bool
	TemplateDir     string
}

// DefaultFixerConfig returns default configuration
func DefaultFixerConfig() *FixerConfig {
	return &FixerConfig{
		UseAI:               true,
		RequireConfirmation: true,
		CreateBackup:        true,
		BackupDir:           ".btsg-backups",
		DryRun:              false,
		MaxLineChanges:      50,
		MinConfidence:       0.7,
		EnableTemplates:     true,
		TemplateDir:         "templates/fixes",
	}
}

// FixError represents an error during fix generation or application
type FixError struct {
	Code    string
	Message string
	Cause   error
}

func (e *FixError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Error codes
const (
	ErrCodeInvalidRequest   = "INVALID_REQUEST"
	ErrCodeGenerationFailed = "GENERATION_FAILED"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeApplyFailed      = "APPLY_FAILED"
	ErrCodeBackupFailed     = "BACKUP_FAILED"
	ErrCodeRollbackFailed   = "ROLLBACK_FAILED"
	ErrCodeFileNotFound     = "FILE_NOT_FOUND"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
)

// Made with Bob
