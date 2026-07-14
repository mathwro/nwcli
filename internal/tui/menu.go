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
	items    []MenuItem
	selected int
	chosen   bool
	quitting bool
}

func NewMenuModel(items []MenuItem) MenuModel {
	return MenuModel{items: items}
}

func (m MenuModel) Selected() int { return m.selected }

func (m MenuModel) Chosen() bool { return m.chosen }

func (m MenuModel) Quitting() bool { return m.quitting }

func (m MenuModel) Init() tea.Cmd { return nil }

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			m.chosen = true
			m.quitting = true
			return m, quitCmd()
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
	b.WriteString("\nq: quit\n")
	return tea.NewView(b.String())
}

func RunMenu(ctx context.Context, items []MenuItem) error {
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		finalModel, err := tea.NewProgram(NewMenuModel(items), tea.WithContext(ctx)).Run()
		if err != nil {
			return err
		}
		menu, ok := finalModel.(MenuModel)
		if !ok || !menu.chosen || menu.selected < 0 || menu.selected >= len(items) || items[menu.selected].Run == nil {
			return nil
		}
		out, err := items[menu.selected].Run(ctx)
		if out != "" {
			fmt.Print(out)
			if !strings.HasSuffix(out, "\n") {
				fmt.Println()
			}
		}
		if err != nil {
			return err
		}
	}
}

func quitCmd() tea.Cmd {
	return func() tea.Msg { return tea.Quit() }
}
