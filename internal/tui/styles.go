package tui

import "charm.land/lipgloss/v2"

var (
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("33")).Padding(1, 2)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	signalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	bodyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("84")).Bold(true)
)
