package ui

import (
	"fmt"
	
	"github.com/charmbracelet/lipgloss"
)

// Design System Colors
var (
	// Primary brand colors
	ColorPrimary    = lipgloss.Color("205") // Bright magenta/pink
	ColorSecondary  = lipgloss.Color("33")  // Bright cyan/blue
	ColorAccent     = lipgloss.Color("214") // Bright orange/yellow
	
	// Semantic colors
	ColorSuccess    = lipgloss.Color("10")  // Bright green
	ColorWarning    = lipgloss.Color("11")  // Bright yellow
	ColorError      = lipgloss.Color("9")   // Bright red
	ColorInfo       = lipgloss.Color("12")  // Bright blue
	
	// Neutral colors (refined for better contrast)
	ColorText       = lipgloss.Color("252") // Near white
	ColorTextMuted  = lipgloss.Color("244") // Light gray
	ColorTextDim    = lipgloss.Color("240") // Medium gray
	ColorBorder     = lipgloss.Color("238") // Dark gray
	ColorBackground = lipgloss.Color("235") // Very dark gray
	ColorSurface    = lipgloss.Color("236") // Slightly lighter dark gray
	ColorOverlay    = lipgloss.Color("234") // Darkest gray
)

// Typography Scale
type FontSize struct {
	Size   int
	Height int
}

var (
	FontDisplay = FontSize{36, 40} // Hero headlines
	FontH1      = FontSize{30, 36} // Page titles
	FontH2      = FontSize{24, 32} // Section headers
	FontH3      = FontSize{20, 28} // Card titles
	FontBody    = FontSize{16, 24} // Default text
	FontSmall   = FontSize{14, 20} // Secondary text
	FontTiny    = FontSize{12, 16} // Captions
)

// Spacing System (4px base unit)
var (
	SpacingXS = 1  // 4px
	SpacingSM = 2  // 8px
	SpacingMD = 4  // 16px
	SpacingLG = 6  // 24px
	SpacingXL = 8  // 32px
	SpacingXXL = 12 // 48px
)

// Component Styles
var (
	// Base text styles
	StyleTitle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Padding(0, 1)
	
	StyleSubtitle = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Bold(true).
		Padding(0, 1)
	
	StyleText = lipgloss.NewStyle().
		Foreground(ColorText)
	
	StyleTextMuted = lipgloss.NewStyle().
		Foreground(ColorTextMuted)
	
	StyleTextDim = lipgloss.NewStyle().
		Foreground(ColorTextDim)
	
	// Interactive states
	StyleFocused = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // Pure white
		Background(ColorSecondary).
		Bold(true).
		Padding(0, 1)
	
	StyleSelected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(ColorAccent).
		Bold(true).
		Padding(0, 1)
	
	StyleUnselected = lipgloss.NewStyle().
		Foreground(ColorTextMuted).
		Padding(0, 1)
	
	// Button styles
	StyleButtonPrimary = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(ColorPrimary).
		Bold(true).
		Padding(0, 2).
		MarginRight(1)
	
	StyleButtonSecondary = lipgloss.NewStyle().
		Foreground(ColorText).
		Background(ColorSurface).
		Padding(0, 2).
		MarginRight(1)
	
	StyleBackButton = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Background(ColorSurface).
		Padding(0, 1).
		MarginRight(2)
	
	// Status and feedback
	StyleSuccess = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true).
		Padding(0, 1)
	
	StyleWarning = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true).
		Padding(0, 1)
	
	StyleError = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true).
		Padding(0, 1)
	
	StyleInfo = lipgloss.NewStyle().
		Foreground(ColorInfo).
		Bold(true).
		Padding(0, 1)
	
	// Layout styles
	StyleModal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(2, 3).
		Background(ColorBackground).
		MarginTop(1).
		MarginBottom(1)
	
	StyleCard = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2).
		Background(ColorSurface).
		MarginBottom(1)
	
	StyleContainer = lipgloss.NewStyle().
		Padding(1, 2)
	
	// Form styles
	StyleFormLabel = lipgloss.NewStyle().
		Foreground(ColorText).
		Bold(true).
		MarginBottom(0)
	
	StyleFormHelp = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Italic(true).
		Padding(0, 3)
	
	// Special indicators
	StyleLoading = lipgloss.NewStyle().
		Foreground(ColorInfo).
		Italic(true).
		Padding(0, 1)
	
	StyleSearchIndicator = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Background(ColorSurface).
		Bold(true).
		Padding(0, 1)
	
	StyleMetadata = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Padding(0, 1)
	
	StyleCode = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Background(ColorOverlay).
		Padding(0, 1)
)

// Helper functions for consistent styling
func CreateHeader(backText, titleText string) string {
	backButton := StyleBackButton.Render("← " + backText)
	title := StyleTitle.Render(titleText)
	return lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)
}

func CreateMetadata(text string) string {
	return StyleMetadata.Render(text)
}

func CreateHelp(text string) string {
	return StyleTextDim.Render(text)
}

// Context-aware help creation
func CreateContextualHelp(essential []string, additional []string) string {
	// Show essential keybinds
	essentialText := lipgloss.JoinHorizontal(lipgloss.Left, essential...)
	
	// If there are additional keybinds, show expansion hint
	if len(additional) > 0 {
		hint := lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Italic(true).
			Render(" • ? for more")
		essentialText = lipgloss.JoinHorizontal(lipgloss.Left, essentialText, hint)
	}
	
	return StyleTextDim.Render(essentialText)
}

// Compact help for the most common actions
func CreateCompactHelp(primary, secondary, exit string) string {
	parts := []string{}
	
	if primary != "" {
		parts = append(parts, primary)
	}
	if secondary != "" {
		parts = append(parts, secondary)
	}
	if exit != "" {
		parts = append(parts, exit)
	}
	
	text := lipgloss.JoinHorizontal(lipgloss.Left, parts...)
	return StyleTextDim.Render(text)
}

// Guaranteed help text that ensures visibility regardless of terminal size
func CreateGuaranteedHelp(helpText string, width int) string {
	helpStyle := lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Width(width).
		Align(lipgloss.Left).
		Padding(0, 1)
	
	// Truncate help text if it's too long for the terminal width
	if width > 0 && len(helpText) > width-2 {
		helpText = helpText[:width-5] + "..."
	}
	
	return helpStyle.Render(helpText)
}

func CreateStatus(text string, statusType string) string {
	switch statusType {
	case "success":
		return StyleSuccess.Render(text)
	case "warning":
		return StyleWarning.Render(text)
	case "error":
		return StyleError.Render(text)
	case "info":
		return StyleInfo.Render(text)
	default:
		return StyleText.Render(text)
	}
}

// Option rendering with consistent styling
func CreateOption(label, description string, isSelected bool) []string {
	var style lipgloss.Style
	if isSelected {
		style = StyleFocused
	} else {
		style = StyleUnselected
	}
	
	lines := []string{style.Render("▶ " + label)}
	
	if description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Italic(true).
			Padding(0, 3)
		lines = append(lines, descStyle.Render(description))
	}
	
	lines = append(lines, "") // Add spacing
	return lines
}

// Git status styling
func CreateGitStatus(status string) string {
	return StyleMetadata.Render("Git: " + status)
}

// Search indicator styling
func CreateSearchIndicator(expression string, count int) string {
	text := lipgloss.JoinHorizontal(
		lipgloss.Left,
		"Boolean: ",
		expression,
		lipgloss.NewStyle().Foreground(ColorTextMuted).Render(" ("),
		lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render(fmt.Sprintf("%d", count)),
		lipgloss.NewStyle().Foreground(ColorTextMuted).Render(" results)"),
	)
	return StyleSearchIndicator.Render(text)
}

// Modal centering helper
func CenterModal(content string, width, height int) string {
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}