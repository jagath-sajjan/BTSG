package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetAbsolutePath returns the absolute path
func GetAbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

// EnsureDir ensures a directory exists, creates it if not
func EnsureDir(path string) error {
	if !FileExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// FormatFileSize formats bytes into human-readable size
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Contains checks if a string slice contains a value
func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// GetFileExtension returns the file extension
func GetFileExtension(path string) string {
	return filepath.Ext(path)
}

// IsTextFile checks if a file is likely a text file based on extension
func IsTextFile(path string) bool {
	textExtensions := []string{
		".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h",
		".rb", ".php", ".sh", ".bash", ".yml", ".yaml", ".json",
		".xml", ".html", ".css", ".md", ".txt", ".sql", ".env",
		".tf", ".hcl", ".dockerfile", ".gitignore",
	}
	
	ext := GetFileExtension(path)
	return Contains(textExtensions, ext)
}

// ShouldExclude checks if a path should be excluded from scanning
func ShouldExclude(path string, excludePatterns []string) bool {
	// Default exclusions
	defaultExclusions := []string{
		"node_modules", "vendor", ".git", ".svn", ".hg",
		"dist", "build", "target", "bin", "obj",
		".idea", ".vscode", ".DS_Store",
	}
	
	base := filepath.Base(path)
	
	// Check default exclusions
	for _, pattern := range defaultExclusions {
		if base == pattern {
			return true
		}
	}
	
	// Check custom exclusions
	for _, pattern := range excludePatterns {
		matched, err := filepath.Match(pattern, base)
		if err == nil && matched {
			return true
		}
	}
	
	return false
}

// Made with Bob
