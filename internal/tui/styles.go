package tui

import "github.com/charmbracelet/lipgloss"

var (
	stNormal  = lipgloss.NewStyle()
	stTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	stDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	stOK      = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	stFail    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	stWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	stSel     = lipgloss.NewStyle().Bold(true).Reverse(true)
	stBanner  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	stBorder  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
	stKeyHelp = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)
