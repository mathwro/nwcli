package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/list"
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
	menu           list.Model
	screen         string
	selected       int
	activeField    int
	input          textinput.Model
	values         map[string]string
	width          int
	height         int
	progress       ProgressModel
	actionEvents   <-chan tea.Msg
	actionCancel   context.CancelFunc
	actionAbort    context.CancelFunc
	stopping       bool
	resultTitle    string
	result         Result
	resultErr      error
	resultCopy     string
	resultStatus   string
	formErr        error
	resultViewport viewport.Model
	quitting       bool
}

type actionStartedMsg struct{}

type actionResultMsg struct {
	title  string
	result Result
	err    error
}

func NewAppModel(actions []Action) AppModel {
	resultViewport := viewport.New(viewport.WithWidth(74), viewport.WithHeight(16))
	resultViewport.SoftWrap = true
	items := make([]list.Item, len(actions))
	for i, action := range actions {
		items[i] = actionItem{index: i, action: action}
	}
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(1)
	delegate.Styles.NormalTitle = bodyStyle
	delegate.Styles.NormalDesc = mutedStyle
	delegate.Styles.SelectedTitle = selectedStyle.PaddingLeft(2)
	delegate.Styles.SelectedDesc = signalStyle.PaddingLeft(2)
	menu := list.New(items, delegate, 74, 16)
	menu.SetShowTitle(false)
	menu.SetShowStatusBar(false)
	menu.SetShowFilter(true)
	menu.SetFilteringEnabled(true)
	menu.DisableQuitKeybindings()
	menu.Styles.HelpStyle = mutedStyle
	return AppModel{
		ctx:            context.Background(),
		actions:        actions,
		menu:           menu,
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
		m.resizeMenu()
		m.resizeResultViewport()
		if m.screen == screenResult {
			m.setResultContent()
		}
	case actionStartedMsg:
		return m, waitActionEvent(m.actionEvents)
	case ProgressEvent:
		if m.screen != screenRunning {
			return m, nil
		}
		if m.stopping {
			return m, waitActionEvent(m.actionEvents)
		}
		updated, cmd := m.progress.Update(msg)
		m.progress = updated.(ProgressModel)
		return m, tea.Batch(cmd, waitActionEvent(m.actionEvents))
	case actionResultMsg:
		if m.actionCancel != nil {
			m.actionCancel()
			m.actionCancel = nil
		}
		if m.actionAbort != nil {
			m.actionAbort()
			m.actionAbort = nil
		}
		m.stopping = false
		m.screen = screenResult
		m.actionEvents = nil
		m.resultTitle = msg.title
		m.result = msg.result
		m.resultErr = msg.err
		m.setResultContent()
		return m, nil
	case tea.KeyPressMsg:
		if isCtrlC(msg) {
			if m.actionCancel != nil {
				m.actionCancel()
			}
			if m.actionAbort != nil {
				m.actionAbort()
			}
			m.quitting = true
			return m, quitCmd()
		}
		if keyName(msg) == "q" && m.screen != screenForm && !(m.screen == screenMenu && m.menu.SettingFilter()) {
			if m.actionCancel != nil {
				m.actionCancel()
			}
			if m.actionAbort != nil {
				m.actionAbort()
			}
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
		if key, ok := msg.(tea.KeyPressMsg); ok && (keyName(key) == "s" || keyName(key) == "esc") && m.actionCancel != nil {
			if !m.stopping {
				m.stopping = true
				m.actionCancel()
				m.progress.label = "Stopping... finishing the current request"
				m.progress.tone = ToneWarning
			}
			return m, nil
		}
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
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch keyName(key) {
		case "enter":
			if !m.menu.SettingFilter() {
				item, ok := m.menu.SelectedItem().(actionItem)
				if !ok || (len(item.action.Fields) == 0 && item.action.Run == nil && item.action.RunProgress == nil) {
					m.quitting = true
					return m, quitCmd()
				}
				m.selected = item.index
				m.openForm()
				return m, nil
			}
		case "esc":
			if m.menu.FilterState() == list.Unfiltered {
				m.quitting = true
				return m, quitCmd()
			}
		}
	}
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	if item, ok := m.menu.SelectedItem().(actionItem); ok {
		m.selected = item.index
	}
	return m, cmd
}

func (m *AppModel) openForm() {
	action := m.actions[m.selected]
	m.screen = screenForm
	m.formErr = nil
	m.activeField = -1
	m.values = make(map[string]string, len(action.Fields))
	for _, field := range action.Fields {
		m.values[field.Name] = field.Default
	}
	m.input = textinput.New()
	for i, field := range action.Fields {
		if m.fieldVisible(field) {
			m.activeField = i
			m.configureInput(field)
			break
		}
	}
}

func (m AppModel) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	if m.activeFieldIsToggle() {
		switch keyName(key) {
		case "left":
			m.formErr = nil
			m.values[m.actions[m.selected].Fields[m.activeField].Name] = "false"
			return m, nil
		case "right":
			m.values[m.actions[m.selected].Fields[m.activeField].Name] = "true"
			m.formErr = nil
			return m, nil
		}
		if key.Key().Code == ' ' {
			field := m.actions[m.selected].Fields[m.activeField]
			m.values[field.Name] = strconv.FormatBool(m.values[field.Name] != "true")
			m.formErr = nil
			return m, nil
		}
	}
	switch keyName(key) {
	case "esc":
		m.formErr = nil
		m.screen = screenMenu
		return m, nil
	case "up":
		m.formErr = nil
		m.moveField(-1)
		return m, nil
	case "down", "tab":
		m.formErr = nil
		m.moveField(1)
		return m, nil
	case "enter":
		return m.submitForm()
	}
	if m.activeFieldIsToggle() {
		return m, nil
	}
	m.formErr = nil
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
	for range action.Fields {
		m.activeField += delta
		if m.activeField < 0 {
			m.activeField = len(action.Fields) - 1
		}
		if m.activeField >= len(action.Fields) {
			m.activeField = 0
		}
		if m.fieldVisible(action.Fields[m.activeField]) {
			m.configureInput(action.Fields[m.activeField])
			return
		}
	}
}

func (m *AppModel) storeActiveField() {
	action := m.actions[m.selected]
	if m.activeField < 0 || m.activeField >= len(action.Fields) {
		return
	}
	field := action.Fields[m.activeField]
	if field.Toggle {
		return
	}
	m.values[field.Name] = strings.TrimSpace(m.input.Value())
}

func (m *AppModel) activeFieldIsToggle() bool {
	action := m.actions[m.selected]
	return m.activeField >= 0 && m.activeField < len(action.Fields) && action.Fields[m.activeField].Toggle
}

func (m AppModel) fieldVisible(field Field) bool {
	return field.Visible == nil || field.Visible(m.values)
}

func (m *AppModel) configureInput(field Field) {
	if field.Toggle {
		m.input.Blur()
		return
	}
	m.input.SetWidth(max(12, m.width-26))
	m.input.SetValue(m.values[field.Name])
	m.input.Placeholder = field.Placeholder
	m.input.CursorEnd()
	m.input.Focus()
}

func (m AppModel) submitForm() (tea.Model, tea.Cmd) {
	m.storeActiveField()
	action := m.actions[m.selected]
	for i, field := range action.Fields {
		if !m.fieldVisible(field) {
			continue
		}
		if field.Validate == nil {
			continue
		}
		if err := field.Validate(m.values[field.Name]); err != nil {
			m.formErr = fmt.Errorf("%s: %w", field.Label, err)
			m.activeField = i
			m.configureInput(field)
			return m, nil
		}
	}
	values := make(map[string]string, len(m.values))
	for _, field := range action.Fields {
		if m.fieldVisible(field) {
			values[field.Name] = m.values[field.Name]
		}
	}
	m.formErr = nil
	m.screen = screenRunning
	if action.RunProgress != nil {
		events := make(chan tea.Msg, 1)
		deliveryCtx, abort := context.WithCancel(m.ctx)
		runCtx, cancel := context.WithCancel(deliveryCtx)
		m.actionEvents = events
		m.actionCancel = cancel
		m.actionAbort = abort
		m.stopping = false
		m.progress = NewProgressModel(action.Title)
		m.progress.label = "Starting..."
		return m, tea.Batch(m.progress.Init(), startProgressAction(runCtx, deliveryCtx, action, values, events))
	}
	return m, func() tea.Msg {
		result, err := action.Run(m.ctx, values)
		return actionResultMsg{title: action.Title, result: result, err: err}
	}
}

func startProgressAction(runCtx, deliveryCtx context.Context, action Action, values map[string]string, events chan<- tea.Msg) tea.Cmd {
	return func() tea.Msg {
		go func() {
			report := func(event ProgressEvent) {
				select {
				case events <- event:
				case <-deliveryCtx.Done():
				}
			}
			result, err := action.RunProgress(runCtx, values, report)
			select {
			case events <- actionResultMsg{title: action.Title, result: result, err: err}:
			case <-deliveryCtx.Done():
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
			m.result = Result{}
			m.resultErr = nil
			m.resultCopy = ""
			m.resultStatus = ""
			m.resultViewport.SetContent("")
			m.resultViewport.GotoTop()
			return m, nil
		case "c":
			if m.resultCopy == "" {
				return m, nil
			}
			m.resultStatus = "Copied to clipboard"
			return m, tea.SetClipboard(m.resultCopy)
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
	b.WriteString(signalStyle.Render("local ── probe ── target"))
	b.WriteString("\n\n")
	b.WriteString(m.menu.View())
	return panelStyle.Render(b.String())
}

func (m AppModel) renderForm() string {
	action := m.actions[m.selected]
	var b strings.Builder
	b.WriteString(titleStyle.Render(action.Title))
	b.WriteString("\n\n")
	for i, field := range action.Fields {
		if !m.fieldVisible(field) {
			continue
		}
		value := m.values[field.Name]
		if field.Toggle {
			value = renderToggle(value == "true")
		} else if i == m.activeField {
			value = m.input.View()
		} else if value == "" {
			value = mutedStyle.Render(field.Placeholder)
		}
		line := fmt.Sprintf("  %-16s %s", mutedStyle.Render(field.Label), value)
		if i == m.activeField {
			line = selectedStyle.Render(">") + line[1:]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if m.formErr != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.formErr.Error()))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("↑/↓ fields • ←/→ or space toggle • enter run • esc back • ctrl+c quit"))
	return panelStyle.Render(b.String())
}

func (m *AppModel) resizeMenu() {
	m.menu.SetSize(
		max(24, m.width-panelStyle.GetHorizontalFrameSize()),
		max(8, m.height-panelStyle.GetVerticalFrameSize()-4),
	)
}

func (m *AppModel) resizeResultViewport() {
	m.resultViewport.SetWidth(max(1, m.width-panelStyle.GetHorizontalFrameSize()))
	m.resultViewport.SetHeight(max(1, m.height-panelStyle.GetVerticalFrameSize()-7))
}

func (m *AppModel) setResultContent() {
	m.resultStatus = ""
	contentWidth := max(1, m.width-panelStyle.GetHorizontalFrameSize())
	m.resultCopy = resultClipboardText(m.result, m.resultErr)
	m.resultViewport.SetContent(renderResultBody(m.result, m.resultErr, contentWidth))
	m.resultViewport.GotoTop()
}

func (m AppModel) renderRunning() string {
	if m.actionEvents != nil {
		help := "s stop → results • q quit"
		if m.stopping {
			help = "stopping current check..."
		}
		content := m.progress.View().Content + "\n" + mutedStyle.Render(help)
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
	b.WriteString(renderResultHeader(m.resultTitle, m.result, m.resultErr))
	b.WriteString("\n\n")
	b.WriteString(m.resultViewport.View())
	b.WriteString("\n\n")
	if m.resultStatus != "" {
		b.WriteString(successStyle.Render(m.resultStatus))
		b.WriteString("\n")
	}
	if m.resultViewport.TotalLineCount() > m.resultViewport.VisibleLineCount() {
		b.WriteString(mutedStyle.Render("↑/↓ scroll • pgup/pgdn page • c copy • esc back • q quit"))
	} else {
		b.WriteString(mutedStyle.Render("c copy • esc back • q quit"))
	}
	return panelStyle.Render(b.String())
}

func renderToggle(on bool) string {
	if on {
		return successStyle.Render("● ON")
	}
	return mutedStyle.Render("○ OFF")
}

func keyName(key tea.KeyPressMsg) string {
	switch key.Key().Code {
	case tea.KeyUp:
		return "up"
	case tea.KeyDown:
		return "down"
	case tea.KeyEnter:
		return "enter"
	case tea.KeyLeft:
		return "left"
	case tea.KeyRight:
		return "right"
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

type actionItem struct {
	index  int
	action Action
}

func (item actionItem) Title() string       { return item.action.Title }
func (item actionItem) Description() string { return item.action.Description }
func (item actionItem) FilterValue() string {
	return item.action.Title + " " + item.action.Description
}
