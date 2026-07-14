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
