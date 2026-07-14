package tui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestMenuModelMovesSelection(t *testing.T) {
	model := NewMenuModel([]MenuItem{{Title: "Subnet carve"}, {Title: "Ping test"}, {Title: "Quit"}})
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	menu := updated.(MenuModel)
	if menu.Selected() != 1 {
		t.Fatalf("Selected = %d, want 1", menu.Selected())
	}
	updated, _ = menu.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	menu = updated.(MenuModel)
	if menu.Selected() != 0 {
		t.Fatalf("Selected = %d, want 0", menu.Selected())
	}
}

func TestMenuModelQuitsOnQ(t *testing.T) {
	model := NewMenuModel([]MenuItem{{Title: "Quit"}})
	updated, cmd := model.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	menu := updated.(MenuModel)
	if !menu.Quitting() {
		t.Fatal("Quitting = false, want true")
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want quit command")
	}
}

func TestMenuModelEnterChoosesItemWithoutRunningNestedProgram(t *testing.T) {
	run := false
	model := NewMenuModel([]MenuItem{{Title: "Subnet carve", Run: func(context.Context) (string, error) {
		run = true
		return "", nil
	}}})
	updated, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	menu := updated.(MenuModel)
	if run {
		t.Fatal("menu action ran inside Update, want RunMenu to launch it after the menu exits")
	}
	if !menu.Chosen() || menu.Selected() != 0 {
		t.Fatalf("Chosen=%t Selected=%d, want chosen selected 0", menu.Chosen(), menu.Selected())
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want quit command")
	}
}

func TestProgressModelCompletes(t *testing.T) {
	model := NewProgressModel("HTTP/S test")
	updated, _ := model.Update(ProgressEvent{Label: "attempt 1", Current: 1, Total: 1, Done: true})
	progress := updated.(ProgressModel)
	if !progress.Done() {
		t.Fatal("Done = false, want true")
	}
	if progress.LastLabel() != "attempt 1" {
		t.Fatalf("LastLabel = %q", progress.LastLabel())
	}
}

func TestAskRejectsNoFields(t *testing.T) {
	_, err := Ask(context.Background(), "empty", nil)
	if err == nil {
		t.Fatal("Ask returned nil error, want no fields error")
	}
}
