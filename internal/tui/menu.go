package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type MenuItem struct {
	Title string
	Run   func(context.Context) (string, error)
}

type MenuModel struct {
	ctx      context.Context
	items    []MenuItem
	selected int
	quitting bool
	output   string
	err      error
}

type actionDoneMsg struct {
	output string
	err    error
}

func NewMenuModel(items []MenuItem) MenuModel {
	return MenuModel{ctx: context.Background(), items: items}
}

func (m MenuModel) Selected() int { return m.selected }

func (m MenuModel) Quitting() bool { return m.quitting }

func (m MenuModel) Init() tea.Cmd { return nil }

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case actionDoneMsg:
		m.output = msg.output
		m.err = msg.err
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case "enter":
			if len(m.items) == 0 || m.items[m.selected].Run == nil {
				m.quitting = true
				return m, quitCmd()
			}
			item := m.items[m.selected]
			ctx := m.ctx
			return m, func() tea.Msg {
				out, err := item.Run(ctx)
				if err != nil {
					return actionDoneMsg{err: err}
				}
				return actionDoneMsg{output: out}
			}
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, quitCmd()
		}
	}
	return m, nil
}

func (m MenuModel) View() tea.View {
	var b strings.Builder
	b.WriteString("nwcli\n\n")
	for i, item := range m.items {
		marker := " "
		if i == m.selected {
			marker = ">"
		}
		fmt.Fprintf(&b, "%s %s\n", marker, item.Title)
	}
	if m.output != "" {
		fmt.Fprintf(&b, "\n%s\n", m.output)
	}
	if m.err != nil {
		fmt.Fprintf(&b, "\nerror: %v\n", m.err)
	}
	b.WriteString("\nq: quit\n")
	return tea.NewView(b.String())
}

func RunMenu(ctx context.Context, items []MenuItem) error {
	model := NewMenuModel(items)
	if ctx != nil {
		model.ctx = ctx
	}
	_, err := tea.NewProgram(model, tea.WithContext(ctx)).Run()
	return err
}

func quitCmd() tea.Cmd {
	return func() tea.Msg { return tea.Quit() }
}
