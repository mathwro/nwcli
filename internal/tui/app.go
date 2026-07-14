package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

const (
	screenMenu    = "menu"
	screenForm    = "form"
	screenRunning = "running"
	screenResult  = "result"
)

type AppModel struct {
	ctx            context.Context
	actions        []Action
	screen         string
	selected       int
	activeField    int
	input          textinput.Model
	values         map[string]string
	width          int
	height         int
	progress       ProgressModel
	actionEvents   <-chan tea.Msg
	resultTitle    string
	resultOutput   string
	resultErr      error
	resultViewport viewport.Model
	quitting       bool
}

type actionStartedMsg struct{}

type actionResultMsg struct {
	title  string
	output string
	err    error
}

func NewAppModel(actions []Action) AppModel {
	resultViewport := viewport.New(viewport.WithWidth(74), viewport.WithHeight(16))
	resultViewport.SoftWrap = true
	return AppModel{
		ctx:            context.Background(),
		actions:        actions,
		screen:         screenMenu,
		width:          80,
		height:         24,
		resultViewport: resultViewport,
	}
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeResultViewport()
	case actionStartedMsg:
		return m, waitActionEvent(m.actionEvents)
	case ProgressEvent:
		if m.screen != screenRunning {
			return m, nil
		}
		updated, cmd := m.progress.Update(msg)
		m.progress = updated.(ProgressModel)
		return m, tea.Batch(cmd, waitActionEvent(m.actionEvents))
	case actionResultMsg:
		m.screen = screenResult
		m.actionEvents = nil
		m.resultTitle = msg.title
		m.resultOutput = msg.output
		m.resultErr = msg.err
		m.setResultContent()
		return m, nil
	case tea.KeyPressMsg:
		if isCtrlC(msg) || keyName(msg) == "q" {
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
		if m.actionEvents != nil {
			updated, cmd := m.progress.Update(msg)
			m.progress = updated.(ProgressModel)
			return m, cmd
		}
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
	if action.RunProgress != nil {
		events := make(chan tea.Msg, 1)
		m.actionEvents = events
		m.progress = NewProgressModel(action.Title)
		m.progress.label = "Starting..."
		return m, tea.Batch(m.progress.Init(), startProgressAction(m.ctx, action, values, events))
	}
	return m, func() tea.Msg {
		out, err := action.Run(m.ctx, values)
		return actionResultMsg{title: action.Title, output: out, err: err}
	}
}

func startProgressAction(ctx context.Context, action Action, values map[string]string, events chan<- tea.Msg) tea.Cmd {
	return func() tea.Msg {
		go func() {
			report := func(event ProgressEvent) {
				select {
				case events <- event:
				case <-ctx.Done():
				}
			}
			out, err := action.RunProgress(ctx, values, report)
			select {
			case events <- actionResultMsg{title: action.Title, output: out, err: err}:
			case <-ctx.Done():
			}
		}()
		return actionStartedMsg{}
	}
}

func waitActionEvent(events <-chan tea.Msg) tea.Cmd {
	if events == nil {
		return nil
	}
	return func() tea.Msg { return <-events }
}

func (m AppModel) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch keyName(key) {
		case "esc", "enter", "backspace":
			m.screen = screenMenu
			m.resultTitle = ""
			m.resultOutput = ""
			m.resultErr = nil
			m.resultViewport.SetContent("")
			m.resultViewport.GotoTop()
			return m, nil
		case "home":
			m.resultViewport.GotoTop()
			return m, nil
		case "end":
			m.resultViewport.GotoBottom()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.resultViewport, cmd = m.resultViewport.Update(msg)
	return m, cmd
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

func (m *AppModel) resizeResultViewport() {
	m.resultViewport.SetWidth(max(1, m.width-panelStyle.GetHorizontalFrameSize()))
	m.resultViewport.SetHeight(max(1, m.height-panelStyle.GetVerticalFrameSize()-4))
}

func (m *AppModel) setResultContent() {
	var b strings.Builder
	if m.resultErr != nil {
		b.WriteString(errorStyle.Render("Error: "))
		b.WriteString(m.resultErr.Error())
		b.WriteString("\n\n")
	}
	if m.resultOutput != "" {
		b.WriteString(strings.TrimSuffix(m.resultOutput, "\n"))
	} else if m.resultErr == nil {
		b.WriteString(successStyle.Render("Done"))
	}
	m.resultViewport.SetContent(b.String())
	m.resultViewport.GotoTop()
}

func (m AppModel) renderRunning() string {
	if m.actionEvents != nil {
		content := m.progress.View().Content + "\n" + mutedStyle.Render("q quit")
		return panelStyle.Render(content)
	}
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
	b.WriteString(m.resultViewport.View())
	b.WriteString("\n\n")
	if m.resultViewport.TotalLineCount() > m.resultViewport.VisibleLineCount() {
		b.WriteString(mutedStyle.Render("↑/↓ scroll • pgup/pgdn page • esc back • q quit"))
	} else {
		b.WriteString(mutedStyle.Render("esc back • q quit"))
	}
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
