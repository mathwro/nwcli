package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

type ProgressEvent struct {
	Label   string
	Current int
	Total   int
	Tone    ResultTone
	Fields  []ResultField
	Done    bool
	Err     error
}

type ProgressModel struct {
	title   string
	bar     progress.Model
	spinner spinner.Model
	label   string
	current int
	total   int
	tone    ResultTone
	fields  []ResultField
	done    bool
	err     error
}

func NewProgressModel(title string) ProgressModel {
	bar := progress.New(progress.WithWidth(32))
	spin := spinner.New()
	return ProgressModel{title: title, bar: bar, spinner: spin}
}

func (m ProgressModel) Done() bool { return m.done }

func (m ProgressModel) LastLabel() string { return m.label }

func (m ProgressModel) Init() tea.Cmd { return m.spinner.Tick }

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressEvent:
		m.label = msg.Label
		m.current = msg.Current
		m.total = msg.Total
		m.tone = msg.Tone
		m.fields = msg.Fields
		m.err = msg.Err
		m.done = msg.Done
		if m.total > 0 {
			_ = m.bar.SetPercent(float64(m.current) / float64(m.total))
		}
		if m.done {
			return m, quitCmd()
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m ProgressModel) View() tea.View {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", m.title)
	if !m.done {
		fmt.Fprintf(&b, "%s %s\n", m.spinner.View(), resultToneStyle(m.tone).Render(m.label))
	} else if m.err != nil {
		fmt.Fprintf(&b, "failed: %v\n", m.err)
	} else {
		fmt.Fprintf(&b, "done: %s\n", m.label)
	}
	if len(m.fields) > 0 {
		b.WriteByte('\n')
		b.WriteString(renderResultFields(m.fields))
	}
	if m.total > 0 {
		percent := float64(m.current) / float64(m.total)
		b.WriteByte('\n')
		b.WriteString(m.bar.ViewAs(percent))
		b.WriteByte('\n')
	}
	return tea.NewView(b.String())
}

type progressRunnerModel struct {
	ProgressModel
	events <-chan ProgressEvent
}

func RunProgress(ctx context.Context, title string, events <-chan ProgressEvent) error {
	if events == nil {
		return nil
	}
	model := progressRunnerModel{ProgressModel: NewProgressModel(title), events: events}
	_, err := tea.NewProgram(model, tea.WithContext(ctx)).Run()
	return err
}

func (m progressRunnerModel) Init() tea.Cmd {
	return waitProgressEvent(m.events)
}

func (m progressRunnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.ProgressModel.Update(msg)
	m.ProgressModel = updated.(ProgressModel)
	if m.done {
		return m, cmd
	}
	if _, ok := msg.(ProgressEvent); ok {
		return m, waitProgressEvent(m.events)
	}
	return m, cmd
}

func (m progressRunnerModel) View() tea.View { return m.ProgressModel.View() }

func waitProgressEvent(events <-chan ProgressEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return ProgressEvent{Label: "complete", Current: 1, Total: 1, Done: true}
		}
		return event
	}
}
