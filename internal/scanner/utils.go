package scanner

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// generateID creates a unique ID for a finding
func generateID(tool, file string, line int) string {
	// Create a hash based on tool, file, and line
	data := fmt.Sprintf("%s:%s:%d:%d", tool, file, line, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("BTSG-%x", hash[:8])
}

// mapSeverityToInt converts severity string to numeric value for sorting
func mapSeverityToInt(severity string) int {
	switch severity {
	case "CRITICAL":
		return 4
	case "HIGH":
		return 3
	case "MEDIUM":
		return 2
	case "LOW":
		return 1
	default:
		return 0
	}
}

// SortFindingsBySeverity sorts findings by severity (highest first)
func SortFindingsBySeverity(findings []*Finding) {
	// Simple bubble sort for demonstration
	n := len(findings)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if mapSeverityToInt(findings[j].Severity) < mapSeverityToInt(findings[j+1].Severity) {
				findings[j], findings[j+1] = findings[j+1], findings[j]
			}
		}
	}
}

// DeduplicateFindings removes duplicate findings based on file and line
func DeduplicateFindings(findings []*Finding) []*Finding {
	seen := make(map[string]bool)
	var unique []*Finding

	for _, finding := range findings {
		key := fmt.Sprintf("%s:%d:%s", finding.File, finding.Line, finding.Description)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, finding)
		}
	}

	return unique
}

// CountBySeverity returns a map of severity counts
func CountBySeverity(findings []*Finding) map[string]int {
	counts := make(map[string]int)
	for _, finding := range findings {
		counts[finding.Severity]++
	}
	return counts
}

// CountByTool returns a map of tool counts
func CountByTool(findings []*Finding) map[string]int {
	counts := make(map[string]int)
	for _, finding := range findings {
		counts[finding.Tool]++
	}
	return counts
}

// Made with Bob
