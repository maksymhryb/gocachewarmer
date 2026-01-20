package ui

import "github.com/charmbracelet/lipgloss"

var styles = map[string]lipgloss.Style{
	"success":   lipgloss.NewStyle().Foreground(lipgloss.Color(COLOR_STATUS_SUCCESS)).Bold(true),
	"redirect":  lipgloss.NewStyle().Foreground(lipgloss.Color(COLOR_STATUS_REDIRECT)).Bold(true),
	"error":     lipgloss.NewStyle().Foreground(lipgloss.Color(COLOR_STATUS_ERROR)).Bold(true),
	"footer":    lipgloss.NewStyle().Background(lipgloss.Color(COLOR_FOOTER_BACKGROUND)).Foreground(lipgloss.Color(COLOR_FOOTER_FONT)).Padding(0, 1),
	"separator": lipgloss.NewStyle().Foreground(lipgloss.Color(COLOR_SEPARATOR)),
}
