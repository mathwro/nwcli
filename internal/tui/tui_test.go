package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func testActions(run func(context.Context, map[string]string) (string, error)) []Action {
	var resultRun func(context.Context, map[string]string) (Result, error)
	if run != nil {
		resultRun = func(ctx context.Context, values map[string]string) (Result, error) {
			output, err := run(ctx, values)
			return Result{Text: output}, err
		}
	}
	return []Action{{
		Title: "Subnet carve",
		Fields: []Field{
			{Name: "cidr", Label: "Parent CIDR", Default: "10.0.0.0/24"},
			{Name: "sizes", Label: "Child prefixes", Default: "27,27,28,28"},
		},
		Run: resultRun,
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

func TestAppFormQEditsInputInsteadOfQuitting(t *testing.T) {
	model := NewAppModel(testActions(nil))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	updated, _ = app.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	app = updated.(AppModel)
	if app.Quitting() {
		t.Fatal("Quitting = true, want false while editing")
	}
	if view := app.View().Content; !strings.Contains(view, "10.0.0.0/24q") {
		t.Fatalf("form view does not contain typed q:\n%s", view)
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
	for _, fragment := range []string{"COMPLETE", "Subnet carve", "Parent: 10.0.0.0/24", "╭", "╰"} {
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

func TestAppResultCRequestsClipboardCopyAndShowsStatus(t *testing.T) {
	model := NewAppModel(testActions(func(context.Context, map[string]string) (string, error) {
		return "line one\nline two\n", nil
	}))
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, cmd := updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, _ = updated.(AppModel).Update(cmd())
	app := updated.(AppModel)

	updated, copyCmd := app.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	app = updated.(AppModel)
	if copyCmd == nil {
		t.Fatal("copyCmd = nil")
	}
	if got := fmt.Sprint(copyCmd()); got != "line one\nline two" {
		t.Fatalf("clipboard content = %q", got)
	}
	if app.Screen() != "result" || app.Quitting() {
		t.Fatalf("app state after copy = screen %q, quitting %t", app.Screen(), app.Quitting())
	}
	if view := app.View().Content; !strings.Contains(view, "Copied to clipboard") {
		t.Fatalf("result view missing copy status:\n%s", view)
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
		RunProgress: func(context.Context, map[string]string, func(ProgressEvent)) (Result, error) {
			return Result{Text: "done"}, nil
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

func TestAppRunningViewHighlightsLivePacketLoss(t *testing.T) {
	actions := []Action{{
		Title:  "Continuous ping",
		Fields: []Field{{Name: "host", Label: "Host", Default: "127.0.0.1"}},
		RunProgress: func(context.Context, map[string]string, func(ProgressEvent)) (Result, error) {
			return Result{}, nil
		},
	}}
	model := NewAppModel(actions)
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, _ = updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, _ = updated.(AppModel).Update(ProgressEvent{
		Label:   "DROP #4 · request timed out",
		Current: 4,
		Tone:    ToneDanger,
		Fields: []ResultField{
			{Label: "Sent", Value: "4"},
			{Label: "Received", Value: "3", Tone: ToneSuccess},
			{Label: "Dropped", Value: "1", Tone: ToneDanger},
			{Label: "Loss", Value: "25.0%", Tone: ToneDanger},
		},
	})
	view := updated.(AppModel).View().Content
	for _, fragment := range []string{"DROP #4", "Sent", "4", "Received", "3", "Dropped", "1", "Loss", "25.0%"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("live ping view missing %q:\n%s", fragment, view)
		}
	}
}

func TestAppRunningStopTransitionsToCopyableResult(t *testing.T) {
	stopped := false
	model := NewAppModel([]Action{{Title: "Continuous ping"}})
	model.screen = screenRunning
	model.actionEvents = make(chan tea.Msg)
	model.actionCancel = func() { stopped = true }
	model.progress = NewProgressModel("Continuous ping")

	updated, _ := model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	app := updated.(AppModel)
	if !stopped || !app.stopping || app.Screen() != "running" {
		t.Fatalf("stop state = stopped %t, stopping %t, screen %q", stopped, app.stopping, app.Screen())
	}
	if view := app.View().Content; !strings.Contains(view, "Stopping") {
		t.Fatalf("running view missing stopping status:\n%s", view)
	}

	updated, _ = app.Update(actionResultMsg{
		title:  "Continuous ping",
		result: Result{Text: "sent=4 received=3 dropped=1 loss=25.0%"},
	})
	app = updated.(AppModel)
	if app.Screen() != "result" {
		t.Fatalf("Screen = %q, want result", app.Screen())
	}
	_, copyCmd := app.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if copyCmd == nil || fmt.Sprint(copyCmd()) != "sent=4 received=3 dropped=1 loss=25.0%" {
		t.Fatalf("clipboard command = %v", copyCmd)
	}
}

func TestAppBooleanFieldUsesToggleControl(t *testing.T) {
	var runValues map[string]string
	actions := []Action{{
		Title: "Ping test",
		Fields: []Field{{
			Name: "continuous", Label: "Continuous", Default: "false", Toggle: true,
		}},
		Run: func(_ context.Context, values map[string]string) (Result, error) {
			runValues = values
			return Result{Text: "done"}, nil
		},
	}}
	model := NewAppModel(actions)
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	if view := app.View().Content; !strings.Contains(view, "OFF") {
		t.Fatalf("toggle view missing OFF state:\n%s", view)
	}
	updated, _ = app.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	app = updated.(AppModel)
	if view := app.View().Content; !strings.Contains(view, "ON") {
		t.Fatalf("toggle view missing ON state:\n%s", view)
	}
	updated, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("cmd = nil, want action command")
	}
	updated, _ = updated.(AppModel).Update(cmd())
	if runValues["continuous"] != "true" {
		t.Fatalf("continuous = %q, want true", runValues["continuous"])
	}
}

func TestAppTextFieldShowsPlaceholderAndAcceptsDirectInput(t *testing.T) {
	actions := []Action{{
		Title: "CIDR inspect",
		Fields: []Field{{
			Name: "cidr", Label: "CIDR", Placeholder: "10.0.0.0/24",
		}},
	}}
	model := NewAppModel(actions)
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	if app.input.Value() != "" {
		t.Fatalf("input value = %q, want empty", app.input.Value())
	}
	if app.input.Placeholder != "10.0.0.0/24" || !strings.Contains(app.View().Content, "0.0.0.0/24") {
		t.Fatalf("placeholder = %q, view:\n%s", app.input.Placeholder, app.View().Content)
	}
	updated, _ = app.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	app = updated.(AppModel)
	if app.input.Value() != "1" {
		t.Fatalf("input value = %q, want direct input 1", app.input.Value())
	}
}

func TestAppMenuUsesActionDescriptions(t *testing.T) {
	actions := testActions(nil)
	actions[0].Description = "Allocate child networks and see remaining space"
	model := NewAppModel(actions)
	view := model.View().Content
	if !strings.Contains(view, actions[0].Description) {
		t.Fatalf("menu view missing action description:\n%s", view)
	}
}

func TestAppFormValidationKeepsFocusAndShowsGuidance(t *testing.T) {
	actions := testActions(nil)
	actions[0].Fields[0].Validate = func(value string) error {
		if value == "10.0.0.0/24" {
			return fmt.Errorf("choose a non-default CIDR")
		}
		return nil
	}
	model := NewAppModel(actions)
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, cmd := updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app := updated.(AppModel)
	if cmd != nil {
		t.Fatal("cmd is non-nil for invalid form")
	}
	if app.Screen() != "form" {
		t.Fatalf("Screen = %q, want form", app.Screen())
	}
	if view := app.View().Content; !strings.Contains(view, "choose a non-default CIDR") {
		t.Fatalf("form view missing validation guidance:\n%s", view)
	}
}

func TestAppResultRendersStructuredVisualHierarchy(t *testing.T) {
	actions := []Action{{
		Title:  "HTTP/S test",
		Fields: []Field{{Name: "host", Label: "Host", Default: "example.com"}},
		Run: func(context.Context, map[string]string) (Result, error) {
			return Result{
				Headline: "https://example.com:443/",
				Summary:  "Endpoint responded with HTTP 200",
				Fields: []ResultField{
					{Label: "Status", Value: "200", Tone: ToneSuccess},
					{Label: "Latency", Value: "18ms"},
				},
				Text: "URL: https://example.com:443/\nStatus: 200",
			}, nil
		},
	}}
	model := NewAppModel(actions)
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, cmd := updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, _ = updated.(AppModel).Update(cmd())
	view := updated.(AppModel).View().Content

	for _, fragment := range []string{"✓", "HTTP/S test", "https://example.com:443/", "Endpoint responded with HTTP 200", "OVERVIEW", "Status", "200", "Latency", "18ms"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("structured result view missing %q:\n%s", fragment, view)
		}
	}
}

func TestAppResultRendersSubnetAddressRailAndAllocations(t *testing.T) {
	actions := []Action{{
		Title:  "Subnet carve",
		Fields: []Field{{Name: "cidr", Label: "Parent CIDR", Default: "10.0.0.0/24"}},
		Run: func(context.Context, map[string]string) (Result, error) {
			return Result{
				Headline: "10.0.0.0/24",
				Summary:  "64 of 256 addresses allocated",
				Text:     "Parent: 10.0.0.0/24",
				Allocation: &AllocationMap{
					Start: "10.0.0.0",
					End:   "10.0.0.255",
					Total: 256,
					Segments: []AllocationSegment{
						{Index: 1, Label: "10.0.0.0/27", Start: "10.0.0.0", End: "10.0.0.31", Offset: 0, Size: 32},
						{Index: 2, Label: "10.0.0.32/27", Start: "10.0.0.32", End: "10.0.0.63", Offset: 32, Size: 32},
					},
					Available: &AllocationSegment{Start: "10.0.0.64", End: "10.0.0.255", Offset: 64, Size: 192},
				},
			}, nil
		},
	}}
	model := NewAppModel(actions)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 96, Height: 30})
	updated, _ = updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, cmd := updated.(AppModel).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, _ = updated.(AppModel).Update(cmd())
	view := updated.(AppModel).View().Content

	for _, fragment := range []string{"ADDRESS SPACE", "10.0.0.0", "10.0.0.255", "█", "░", "ALLOCATIONS", "10.0.0.0/27", "10.0.0.32/27", "AVAILABLE", "192 addresses"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("allocation result view missing %q:\n%s", fragment, view)
		}
	}
}
