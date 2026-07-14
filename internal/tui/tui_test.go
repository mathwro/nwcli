package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func testActions(run func(context.Context, map[string]string) (string, error)) []Action {
	return []Action{{
		Title: "Subnet carve",
		Fields: []Field{
			{Name: "cidr", Label: "Parent CIDR", Default: "10.0.0.0/24"},
			{Name: "sizes", Label: "Child prefixes", Default: "27,27,28,28"},
		},
		Run: run,
	}, {Title: "Quit"}}
}

func TestAppMenuEnterOpensForm(t *testing.T) {
	model := NewAppModel(testActions(nil))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	if app.Screen() != "form" {
		t.Fatalf("Screen = %q, want form", app.Screen())
	}
	if app.ActiveField() != 0 {
		t.Fatalf("ActiveField = %d, want 0", app.ActiveField())
	}
}

func TestAppFormEscapeReturnsToMenuWithoutQuitting(t *testing.T) {
	model := NewAppModel(testActions(nil))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	updated, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	app = updated.(AppModel)
	if app.Screen() != "menu" {
		t.Fatalf("Screen = %q, want menu", app.Screen())
	}
	if app.Quitting() {
		t.Fatal("Quitting = true, want false")
	}
}

func TestAppFormCtrlCQuits(t *testing.T) {
	model := NewAppModel(testActions(nil))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	updated, cmd := app.Update(tea.KeyPressMsg{Code: 'c', Text: "c", Mod: tea.ModCtrl})
	app = updated.(AppModel)
	if !app.Quitting() {
		t.Fatal("Quitting = false, want true")
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want quit command")
	}
}

func TestAppFormQQuits(t *testing.T) {
	model := NewAppModel(testActions(nil))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	updated, cmd := app.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	app = updated.(AppModel)
	if !app.Quitting() {
		t.Fatal("Quitting = false, want true")
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want quit command")
	}
}

func TestAppRunningQQuits(t *testing.T) {
	model := NewAppModel(testActions(func(context.Context, map[string]string) (string, error) { return "ok", nil }))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	updated, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app = updated.(AppModel)
	if app.Screen() != "running" {
		t.Fatalf("Screen = %q, want running", app.Screen())
	}
	updated, cmd := app.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	app = updated.(AppModel)
	if !app.Quitting() {
		t.Fatal("Quitting = false, want true")
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want quit command")
	}
}

func TestAppSubmitRunsActionAndShowsResultInView(t *testing.T) {
	runValues := map[string]string{}
	model := NewAppModel(testActions(func(_ context.Context, values map[string]string) (string, error) {
		runValues = values
		return "Parent: 10.0.0.0/24", nil
	}))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	updated, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("cmd = nil, want action command")
	}
	msg := cmd()
	updated, _ = updated.(AppModel).Update(msg)
	app = updated.(AppModel)
	if app.Screen() != "result" {
		t.Fatalf("Screen = %q, want result", app.Screen())
	}
	if runValues["cidr"] != "10.0.0.0/24" || runValues["sizes"] != "27,27,28,28" {
		t.Fatalf("values = %#v, want defaults", runValues)
	}
	view := app.View().Content
	for _, fragment := range []string{"Subnet carve result", "Parent: 10.0.0.0/24", "╭", "╰"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("view missing %q:\n%s", fragment, view)
		}
	}
}

func TestAppResultEscapeReturnsToMenu(t *testing.T) {
	model := NewAppModel(testActions(func(context.Context, map[string]string) (string, error) { return "ok", nil }))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	updated, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, _ = updated.(AppModel).Update(cmd())
	app = updated.(AppModel)
	updated, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	app = updated.(AppModel)
	if app.Screen() != "menu" {
		t.Fatalf("Screen = %q, want menu", app.Screen())
	}
}

func TestAppViewUsesAltScreen(t *testing.T) {
	model := NewAppModel(testActions(nil))
	view := model.View()
	if !view.AltScreen {
		t.Fatal("AltScreen = false, want true")
	}
}

func TestAppResultViewportScrollsLongOutput(t *testing.T) {
	output := "first\n" + strings.Repeat("middle\n", 48) + "last"
	model := NewAppModel(testActions(func(context.Context, map[string]string) (string, error) {
		return output, nil
	}))
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	updated, _ = updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, cmd := updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, _ = updated.(AppModel).Update(cmd())
	app := updated.(AppModel)

	view := app.View().Content
	if !strings.Contains(view, "first") {
		t.Fatalf("initial result view missing first line:\n%s", view)
	}
	if strings.Contains(view, "last") {
		t.Fatalf("initial result view contains last line without scrolling:\n%s", view)
	}

	updated, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	view = updated.(AppModel).View().Content
	if !strings.Contains(view, "last") {
		t.Fatalf("scrolled result view missing last line:\n%s", view)
	}
}

func TestAppRunningViewShowsActionProgress(t *testing.T) {
	actions := []Action{{
		Title:  "Ping test",
		Fields: []Field{{Name: "count", Label: "Count", Default: "50"}},
		RunProgress: func(context.Context, map[string]string, func(ProgressEvent)) (string, error) {
			return "done", nil
		},
	}}
	model := NewAppModel(actions)
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, cmd := updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("cmd = nil, want progress action command")
	}
	updated, _ = updated.(AppModel).Update(ProgressEvent{Label: "Ping 2 of 50", Current: 2, Total: 50})
	app := updated.(AppModel)
	if app.Screen() != "running" {
		t.Fatalf("Screen = %q, want running", app.Screen())
	}
	if view := app.View().Content; !strings.Contains(view, "Ping 2 of 50") {
		t.Fatalf("running view missing progress label:\n%s", view)
	}
}
