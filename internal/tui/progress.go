package tui

import (
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
	return NewProgressModelWithWidth(title, 36)
}

func NewProgressModelWithWidth(title string, width int) ProgressModel {
	bar := progress.New(progress.WithWidth(max(16, width)))
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
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")
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
