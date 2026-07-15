package tui

import "charm.land/lipgloss/v2"

var (
	panelStyle          = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("33")).Padding(1, 2)
	titleStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
	selectedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	signalStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	bodyStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	mutedStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	successStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("84")).Bold(true)
	warningStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	sectionStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Bold(true)
	resultHeadlineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	resultValueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	fieldLabelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorDetailStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("210"))
	unusedRailStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	allocationSegmentStyles = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("45")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("69")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("75")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("81")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("117")),
	}
)
