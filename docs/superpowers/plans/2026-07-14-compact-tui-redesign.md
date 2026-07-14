# Compact TUI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current multi-program interactive flow with one compact, styled Bubble Tea TUI where Escape navigates back from submenus instead of quitting the app.

**Architecture:** Keep Cobra commands unchanged for non-interactive usage. Move interactive state into one `internal/tui` app model with Menu, Form, Running, and Result screens. `cmd/root.go` builds TUI actions that call existing subnet/check logic; `internal/tui` owns input state, styling, transitions, and rendering.

**Tech Stack:** Go 1.26.4, Cobra, `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`.

## Global Constraints

- Use one Bubble Tea app for the interactive mode.
- Escape in Form or Result returns to Menu, not the shell.
- `q` and Ctrl+C quit the app.
- Action results render inside a bordered TUI panel, not with `fmt.Print` between Bubble Tea sessions.
- Use Lip Gloss for borders, color, title, selected item, muted help, success, and error styling.
- Do not change existing non-interactive Cobra command behavior.
- Do not add new dependencies.
- Use TDD for behavior-bearing code.

---

## File Structure

- Modify: `internal/tui/tui_test.go` — replace/extend tests for the new app model transitions and rendered fragments.
- Create: `internal/tui/action.go` — shared `Action` and `Field` types.
- Create: `internal/tui/app.go` — app model, screen state, transitions, `RunApp`.
- Create: `internal/tui/styles.go` — Lip Gloss styles and small render helpers.
- Modify: `internal/tui/menu.go` — keep only menu rendering/update helpers or fold into app model.
- Modify: `internal/tui/prompt.go` — remove `Ask` if unused after migration.
- Modify: `cmd/root.go` — build `[]tui.Action`, call `tui.RunApp`, and adapt interactive action signatures.
- Modify: `cmd/command_test.go` — root no-args test should inject the menu runner unchanged; no terminal needed.

---

### Task 1: TUI App State and Navigation

**Files:**
- Modify: `internal/tui/tui_test.go`
- Create: `internal/tui/action.go`
- Create: `internal/tui/app.go`
- Create: `internal/tui/styles.go`
- Modify: `internal/tui/menu.go`
- Modify: `internal/tui/prompt.go`

**Interfaces:**
- Produces: `type Action struct { Title string; Fields []Field; Run func(context.Context, map[string]string) (string, error) }`
- Produces: `type Field struct { Name string; Label string; Default string }`
- Produces: `func NewAppModel(actions []Action) AppModel`
- Produces: `func RunApp(ctx context.Context, actions []Action) error`
- Produces test helpers/state accessors: `Screen() string`, `Selected() int`, `ActiveField() int`, `Quitting() bool`
- Consumes: Bubble Tea, Bubbles textinput, Lip Gloss.

- [ ] **Step 1: Write failing tests for form Escape and result rendering**

Replace `internal/tui/tui_test.go` with tests for the new app model:

```go
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
```

- [ ] **Step 2: Run tests to verify RED**

Run: `go test ./internal/tui`

Expected: FAIL because `Action`, `NewAppModel`, `AppModel.Screen`, and the new form/result behavior are not implemented.

- [ ] **Step 3: Add action types and styles**

Create `internal/tui/action.go`:

```go
package tui

import "context"

type Action struct {
	Title  string
	Fields []Field
	Run    func(context.Context, map[string]string) (string, error)
}

type Field struct {
	Name    string
	Label   string
	Default string
}
```

Create `internal/tui/styles.go` with Lip Gloss styles:

```go
package tui

import "charm.land/lipgloss/v2"

var (
	panelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 2)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)
```

- [ ] **Step 4: Implement one app model**

Create `internal/tui/app.go`:

- `screen` constants: `menu`, `form`, `running`, `result`.
- `AppModel` fields: context, actions, selected, activeField, values map, resultTitle, resultOutput, resultErr, quitting.
- `NewAppModel(actions []Action) AppModel` initializes context to `context.Background()`.
- `RunApp(ctx, actions)` runs `tea.NewProgram(model, tea.WithContext(ctx)).Run()`.
- `Update` delegates to menu/form/result handlers.
- Form fields are editable: the active field uses `textinput.Model`, starts with the field default as its value, Up/Down moves between fields while preserving edits, and Enter submits the current values.
- Action command returns a private message containing output/error.
- Result screen stores output/error and renders it.
- Escape in form/result returns to menu without error.
- `ctrl+c` always sets `quitting` and returns quit command.

Keep the form implementation small: one active `textinput.Model`, a `values map[string]string`, and no validation beyond what the action already returns.

- [ ] **Step 5: Remove old prompt/menu multi-program code**

Modify `internal/tui/menu.go` and `internal/tui/prompt.go`:

- Delete old `MenuItem`, `RunMenu`, `Ask`, and old `Field` definitions.
- Keep only helpers used by `app.go`, or remove these files if empty.
- Ensure there is only one `Field` type in `internal/tui`.

- [ ] **Step 6: Run tests to verify GREEN**

Run: `go test ./internal/tui`

Expected: PASS.

---

### Task 2: Command Integration and Smoke Tests

**Files:**
- Modify: `cmd/root.go`
- Modify: `cmd/command_test.go`
- Modify: `internal/tui/tui_test.go` only if integration exposes a missing state test.

**Interfaces:**
- Consumes: `tui.Action`, `tui.Field`, `tui.RunApp`.
- Preserves: `NewRootCommand(menuRunner func(context.Context) error) *cobra.Command` for test injection.

- [ ] **Step 1: Write failing command/TUI integration test**

Add to `cmd/command_test.go`:

```go
func TestInteractiveActionsUseTUIAppFields(t *testing.T) {
	actions := interactiveActions()
	if len(actions) < 3 {
		t.Fatalf("len(actions) = %d, want at least 3", len(actions))
	}
	if actions[0].Title != "Subnet carve" {
		t.Fatalf("first action = %q, want Subnet carve", actions[0].Title)
	}
	if len(actions[0].Fields) != 2 || actions[0].Fields[0].Label != "Parent CIDR" || actions[0].Fields[1].Label != "Child prefixes" {
		t.Fatalf("subnet fields = %#v", actions[0].Fields)
	}
}
```

- [ ] **Step 2: Run command test to verify RED**

Run: `go test ./cmd -run TestInteractiveActionsUseTUIAppFields`

Expected: FAIL because `interactiveActions` does not exist or returns old `MenuItem` values.

- [ ] **Step 3: Migrate root interactive wiring**

Modify `cmd/root.go`:

- `runInteractiveMenu(ctx)` calls `tui.RunApp(ctx, interactiveActions())`.
- Add `interactiveActions() []tui.Action` with Subnet carve, Ping test, HTTP/S test, Quit.
- Change `runInteractiveSubnet`, `runInteractivePing`, and `runInteractiveHTTP` to accept `values map[string]string` instead of calling `tui.Ask`.
- Keep output rendering logic exactly as command mode uses today.
- Quit action has no fields and nil Run.

- [ ] **Step 4: Run targeted tests**

Run: `go test ./cmd ./internal/tui`

Expected: PASS.

- [ ] **Step 5: Run full verification**

Run: `go test ./...`

Expected: PASS.

Run: `go build ./...`

Expected: PASS.

- [ ] **Step 6: Smoke test interactive behavior manually**

Run: `go run .`

Verify:

- Main menu has bordered panel and styled selected item.
- Select `Subnet carve`, press Escape, and confirm the menu remains open.
- Select `Subnet carve`, press Enter to submit defaults, and confirm result is displayed inside a bordered panel.
- Press Escape from result and confirm menu returns.
- Press `q` and confirm clean exit.

- [ ] **Step 7: Smoke test existing commands**

Run:

```bash
go run . subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28
go run ./cmd/nw subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28
go run . --help
```

Expected: existing command behavior remains intact.

- [ ] **Step 8: Commit implementation**

Run:

```bash
git add cmd internal/tui
git commit -m "fix: keep interactive flows inside tui"
```

Expected: commit succeeds.
