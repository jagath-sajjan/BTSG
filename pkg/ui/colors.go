package ui

import (
	"fmt"

	"github.com/fatih/color"
)

// Color definitions for consistent styling
var (
	// Severity colors
	ColorCritical = color.New(color.FgRed, color.Bold)
	ColorHigh     = color.New(color.FgRed)
	ColorMedium   = color.New(color.FgYellow)
	ColorLow      = color.New(color.FgCyan)
	ColorInfo     = color.New(color.FgBlue)

	// Status colors
	ColorSuccess = color.New(color.FgGreen, color.Bold)
	ColorError   = color.New(color.FgRed, color.Bold)
	ColorWarning = color.New(color.FgYellow, color.Bold)
	ColorInfo2   = color.New(color.FgCyan, color.Bold)

	// UI element colors
	ColorHeader  = color.New(color.FgWhite, color.Bold, color.Underline)
	ColorBold    = color.New(color.Bold)
	ColorDim     = color.New(color.Faint)
	ColorCode    = color.New(color.FgMagenta)
	ColorPath    = color.New(color.FgCyan)
	ColorNumber  = color.New(color.FgYellow)
	ColorCommand = color.New(color.FgGreen)
)

// Severity returns the appropriate color for a severity level
func Severity(level string) *color.Color {
	switch level {
	case "CRITICAL", "Critical":
		return ColorCritical
	case "HIGH", "High":
		return ColorHigh
	case "MEDIUM", "Medium", "MODERATE", "Moderate":
		return ColorMedium
	case "LOW", "Low":
		return ColorLow
	default:
		return ColorInfo
	}
}

// SeverityIcon returns an icon for a severity level
func SeverityIcon(level string) string {
	switch level {
	case "CRITICAL", "Critical":
		return "🔴"
	case "HIGH", "High":
		return "🟠"
	case "MEDIUM", "Medium", "MODERATE", "Moderate":
		return "🟡"
	case "LOW", "Low":
		return "🟢"
	default:
		return "ℹ️"
	}
}

// PrintSuccess prints a success message
func PrintSuccess(format string, args ...interface{}) {
	fmt.Print("✅ ")
	ColorSuccess.Printf(format+"\n", args...)
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	fmt.Print("❌ ")
	ColorError.Printf(format+"\n", args...)
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	fmt.Print("⚠️  ")
	ColorWarning.Printf(format+"\n", args...)
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	fmt.Print("ℹ️  ")
	ColorInfo2.Printf(format+"\n", args...)
}

// PrintHeader prints a section header
func PrintHeader(text string) {
	fmt.Println()
	ColorHeader.Println(text)
	fmt.Println()
}

// PrintSeparator prints a visual separator
func PrintSeparator() {
	ColorDim.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// PrintBanner prints a styled banner
func PrintBanner(title, subtitle string) {
	fmt.Println()
	ColorBold.Printf("╔═══════════════════════════════════════════════════════════════════════════╗\n")
	ColorBold.Printf("║                                                                           ║\n")
	ColorBold.Printf("║  %-71s  ║\n", title)
	if subtitle != "" {
		ColorDim.Printf("║  %-71s  ║\n", subtitle)
	}
	ColorBold.Printf("║                                                                           ║\n")
	ColorBold.Printf("╚═══════════════════════════════════════════════════════════════════════════╝\n")
	fmt.Println()
}

// PrintSeverityBadge prints a colored severity badge
func PrintSeverityBadge(severity string) {
	icon := SeverityIcon(severity)
	color := Severity(severity)
	fmt.Printf("%s ", icon)
	color.Printf("[%s]", severity)
}

// FormatPath formats a file path with color
func FormatPath(path string) string {
	return ColorPath.Sprint(path)
}

// FormatCode formats code with color
func FormatCode(code string) string {
	return ColorCode.Sprint(code)
}

// FormatNumber formats a number with color
func FormatNumber(n interface{}) string {
	return ColorNumber.Sprintf("%v", n)
}

// FormatCommand formats a command with color
func FormatCommand(cmd string) string {
	return ColorCommand.Sprint(cmd)
}

// Made with Bob
