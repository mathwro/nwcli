package tui

import tea "charm.land/bubbletea/v2"

func quitCmd() tea.Cmd {
	return func() tea.Msg { return tea.Quit() }
}
