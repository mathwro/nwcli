package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

const (
	screenMenu    = "menu"
	screenForm    = "form"
	screenRunning = "running"
	screenResult  = "result"
)

type AppModel struct {
	ctx          context.Context
	actions      []Action
	screen       string
	selected     int
	activeField  int
	input        textinput.Model
	values       map[string]string
	resultTitle  string
	resultOutput string
	resultErr    error
	quitting     bool
}

type actionResultMsg struct {
	title  string
	output string
	err    error
}

func NewAppModel(actions []Action) AppModel {
	return AppModel{ctx: context.Background(), actions: actions, screen: screenMenu}
}

func RunApp(ctx context.Context, actions []Action) error {
	if ctx == nil {
		ctx = context.Background()
	}
	model := NewAppModel(actions)
	model.ctx = ctx
	_, err := tea.NewProgram(model, tea.WithContext(ctx)).Run()
	return err
}

func (m AppModel) Screen() string { return m.screen }

func (m AppModel) Selected() int { return m.selected }

func (m AppModel) ActiveField() int { return m.activeField }

func (m AppModel) Quitting() bool { return m.quitting }

func (m AppModel) Init() tea.Cmd { return nil }

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case actionResultMsg:
		m.screen = screenResult
		m.resultTitle = msg.title
		m.resultOutput = msg.output
		m.resultErr = msg.err
		return m, nil
	case tea.KeyPressMsg:
		if isCtrlC(msg) {
			m.quitting = true
			return m, quitCmd()
		}
	}

	switch m.screen {
	case screenMenu:
		return m.updateMenu(msg)
	case screenForm:
		return m.updateForm(msg)
	case screenRunning:
		return m, nil
	case screenResult:
		return m.updateResult(msg)
	default:
		m.screen = screenMenu
		return m, nil
	}
}

func (m AppModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch keyName(key) {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.actions)-1 {
			m.selected++
		}
	case "enter":
		if len(m.actions) == 0 || m.selected >= len(m.actions) || (len(m.actions[m.selected].Fields) == 0 && m.actions[m.selected].Run == nil) {
			m.quitting = true
			return m, quitCmd()
		}
		m.openForm()
	case "q", "esc":
		m.quitting = true
		return m, quitCmd()
	}
	return m, nil
}

func (m *AppModel) openForm() {
	action := m.actions[m.selected]
	m.screen = screenForm
	m.activeField = 0
	m.values = make(map[string]string, len(action.Fields))
	for _, field := range action.Fields {
		m.values[field.Name] = field.Default
	}
	m.input = textinput.New()
	if len(action.Fields) > 0 {
		m.input.SetValue(action.Fields[0].Default)
		m.input.Placeholder = action.Fields[0].Default
	}
	m.input.Focus()
}

func (m AppModel) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	switch keyName(key) {
	case "esc":
		m.screen = screenMenu
		return m, nil
	case "up":
		m.moveField(-1)
		return m, nil
	case "down", "tab":
		m.moveField(1)
		return m, nil
	case "enter":
		return m.submitForm()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *AppModel) moveField(delta int) {
	action := m.actions[m.selected]
	if len(action.Fields) == 0 {
		return
	}
	m.storeActiveField()
	m.activeField += delta
	if m.activeField < 0 {
		m.activeField = len(action.Fields) - 1
	}
	if m.activeField >= len(action.Fields) {
		m.activeField = 0
	}
	field := action.Fields[m.activeField]
	m.input.SetValue(m.values[field.Name])
	m.input.Placeholder = field.Default
	m.input.CursorEnd()
}

func (m *AppModel) storeActiveField() {
	action := m.actions[m.selected]
	if m.activeField < 0 || m.activeField >= len(action.Fields) {
		return
	}
	field := action.Fields[m.activeField]
	m.values[field.Name] = strings.TrimSpace(m.input.Value())
	if m.values[field.Name] == "" {
		m.values[field.Name] = field.Default
	}
}

func (m AppModel) submitForm() (tea.Model, tea.Cmd) {
	m.storeActiveField()
	action := m.actions[m.selected]
	values := make(map[string]string, len(m.values))
	for key, value := range m.values {
		values[key] = value
	}
	m.screen = screenRunning
	return m, func() tea.Msg {
		out, err := action.Run(m.ctx, values)
		return actionResultMsg{title: action.Title, output: out, err: err}
	}
}

func (m AppModel) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch keyName(key) {
	case "esc", "enter", "backspace":
		m.screen = screenMenu
		m.resultTitle = ""
		m.resultOutput = ""
		m.resultErr = nil
	case "q":
		m.quitting = true
		return m, quitCmd()
	}
	return m, nil
}

func (m AppModel) View() tea.View {
	switch m.screen {
	case screenForm:
		return appView(m.renderForm())
	case screenRunning:
		return appView(m.renderRunning())
	case screenResult:
		return appView(m.renderResult())
	default:
		return appView(m.renderMenu())
	}
}

func appView(content string) tea.View {
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m AppModel) renderMenu() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("nwcli"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Network toolbox"))
	b.WriteString("\n\n")
	for i, action := range m.actions {
		line := fmt.Sprintf("  %s", action.Title)
		if i == m.selected {
			line = selectedStyle.Render("> " + action.Title)
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("↑/↓ move • enter select • q quit"))
	return panelStyle.Render(b.String())
}

func (m AppModel) renderForm() string {
	action := m.actions[m.selected]
	var b strings.Builder
	b.WriteString(titleStyle.Render(action.Title))
	b.WriteString("\n\n")
	for i, field := range action.Fields {
		value := m.values[field.Name]
		if i == m.activeField {
			value = m.input.View()
		}
		line := fmt.Sprintf("  %-16s %s", mutedStyle.Render(field.Label), value)
		if i == m.activeField {
			line = selectedStyle.Render(">") + line[1:]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("↑/↓ fields • enter submit • esc back • ctrl+c quit"))
	return panelStyle.Render(b.String())
}

func (m AppModel) renderRunning() string {
	title := "Running"
	if m.selected >= 0 && m.selected < len(m.actions) {
		title = m.actions[m.selected].Title
	}
	content := titleStyle.Render(title) + "\n\n" + mutedStyle.Render("Working...")
	return panelStyle.Render(content)
}

func (m AppModel) renderResult() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.resultTitle + " result"))
	b.WriteString("\n\n")
	if m.resultErr != nil {
		b.WriteString(errorStyle.Render("Error: "))
		b.WriteString(m.resultErr.Error())
		b.WriteString("\n\n")
	}
	if m.resultOutput != "" {
		b.WriteString(m.resultOutput)
		if !strings.HasSuffix(m.resultOutput, "\n") {
			b.WriteByte('\n')
		}
	} else if m.resultErr == nil {
		b.WriteString(successStyle.Render("Done"))
		b.WriteByte('\n')
	}
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("esc back • q quit"))
	return panelStyle.Render(b.String())
}

func keyName(key tea.KeyPressMsg) string {
	switch key.Key().Code {
	case tea.KeyUp:
		return "up"
	case tea.KeyDown:
		return "down"
	case tea.KeyEnter:
		return "enter"
	case tea.KeyEscape:
		return "esc"
	case tea.KeyBackspace:
		return "backspace"
	case tea.KeyTab:
		return "tab"
	}
	return key.String()
}

func isCtrlC(key tea.KeyPressMsg) bool {
	return key.String() == "ctrl+c" || (key.Key().Mod == tea.ModCtrl && key.Key().Code == 'c')
}
