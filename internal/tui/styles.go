package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary = lipgloss.Color("#7D56F4")
	ColorGreen   = lipgloss.Color("#04B575")
	ColorGray    = lipgloss.Color("#626262")
	ColorRed     = lipgloss.Color("#FF4672")

	StyleTitle     = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).MarginBottom(1)
	StyleActive    = lipgloss.NewStyle().Bold(true).Foreground(ColorGreen)
	StyleInstalled = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	StyleAvailable = lipgloss.NewStyle().Foreground(ColorGray)
	StyleError     = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
	StyleHint      = lipgloss.NewStyle().Foreground(ColorGray).Italic(true)
)

func RenderTitle(s string) string { return StyleTitle.Render(s) }
