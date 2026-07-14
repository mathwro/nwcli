package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Field struct {
	Name    string
	Prompt  string
	Default string
}

func Ask(ctx context.Context, title string, fields []Field) (map[string]string, error) {
	if len(fields) == 0 {
		return nil, errors.New("prompt requires at least one field")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	model := newPromptModel(title, fields)
	finalModel, err := tea.NewProgram(model, tea.WithContext(ctx)).Run()
	if err != nil {
		return nil, err
	}
	prompt, ok := finalModel.(promptModel)
	if !ok {
		return nil, errors.New("unexpected prompt model")
	}
	if prompt.cancelled {
		return nil, errors.New("cancelled")
	}
	return prompt.values, nil
}

type promptModel struct {
	title     string
	fields    []Field
	index     int
	input     textinput.Model
	values    map[string]string
	cancelled bool
}

func newPromptModel(title string, fields []Field) promptModel {
	input := textinput.New()
	input.Prompt = fields[0].Prompt + ": "
	input.Placeholder = fields[0].Default
	return promptModel{title: title, fields: fields, input: input, values: make(map[string]string, len(fields))}
}

func (m promptModel) Init() tea.Cmd {
	return m.input.Focus()
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.cancelled = true
			return m, quitCmd()
		case "enter":
			field := m.fields[m.index]
			value := strings.TrimSpace(m.input.Value())
			if value == "" {
				value = field.Default
			}
			m.values[field.Name] = value
			m.index++
			if m.index >= len(m.fields) {
				return m, quitCmd()
			}
			next := m.fields[m.index]
			m.input.Reset()
			m.input.Prompt = next.Prompt + ": "
			m.input.Placeholder = next.Default
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m promptModel) View() tea.View {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", m.title)
	fmt.Fprintf(&b, "Field %d/%d\n", m.index+1, len(m.fields))
	b.WriteString(m.input.View())
	b.WriteString("\n\nesc: cancel\n")
	return tea.NewView(b.String())
}
