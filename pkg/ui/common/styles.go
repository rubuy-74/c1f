package common

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			MarginLeft(2).
			MarginTop(1).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#F25D94")).
			Bold(true)

	StatusGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	StatusRed   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)
