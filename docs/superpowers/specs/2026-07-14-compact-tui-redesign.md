# Compact TUI Redesign

## Goal

Improve `nwcli` interactive mode so it behaves like one cohesive Bubble Tea TUI instead of multiple terminal programs stitched together with stdout output.

## Root Cause

Two implementation choices cause the reported behavior:

- `tui.Ask` returns a `cancelled` error when Escape is pressed. `RunMenu` currently treats any selected action error as a fatal application error, so Escape inside a submenu exits the whole CLI.
- `RunMenu` exits the menu program, runs action-specific prompt programs, prints action output with `fmt.Print`, then starts a new menu program. That creates newline output between screens instead of one persistent TUI.

## Scope

### In Scope

- Replace interactive mode internals with a single Bubble Tea app model.
- Keep all non-interactive Cobra commands unchanged.
- Add compact panel styling with Lip Gloss colors and borders.
- Make Escape return from action forms/results to the main menu.
- Keep `q` and Ctrl+C as application quit controls.
- Render subnet carve results, ping results, HTTP/S results, and validation errors inside TUI panels.
- Add automated tests for state transitions and stable rendered fragments.

### Out of Scope

- Mouse support.
- Large responsive layout framework.
- New commands or new network behavior.
- JSON/export support.
- Replacing Cobra command output formatting.

## Interaction Design

Interactive mode has four screens:

1. **Menu** — list actions: Subnet carve, Ping test, HTTP/S test, Quit.
2. **Form** — show the selected action fields in one bordered panel.
3. **Running** — show spinner/progress text for network checks or immediate execution for subnet carve.
4. **Result** — show command result or error in a bordered panel.

Keyboard behavior:

- Menu:
  - Up/Down or `k`/`j`: move selection.
  - Enter: open selected action.
  - `q`, Escape, Ctrl+C: quit app.
- Form:
  - Up/Down: move between fields.
  - Enter: submit the form.
  - Escape: return to Menu without error.
  - Ctrl+C: quit app.
- Result:
  - Escape, Enter, Backspace: return to Menu.
  - `q`, Ctrl+C: quit app.

Escape from a submenu is navigation, not failure.

## Visual Design

Use Lip Gloss only; do not add new styling dependencies.

Style rules:

- App frame: rounded border around current screen.
- Title: colored and bold, e.g. cyan or blue.
- Selected menu item: colored foreground and `>` marker.
- Field labels: muted color.
- Active field: highlighted border or marker.
- Help/footer: muted text.
- Success: green label.
- Error: red label, but still inside the app shell.

Example shape:

```text
╭─ nwcli ─────────────────────────────────────────╮
│ Network toolbox                                 │
│                                                │
│ > Subnet carve                                  │
│   Ping test                                     │
│   HTTP/S test                                   │
│   Quit                                          │
│                                                │
│ ↑/↓ move • enter select • q quit                │
╰────────────────────────────────────────────────╯
```

Form shape:

```text
╭─ Subnet carve ─────────────────────────────────╮
│ Parent CIDR     10.0.0.0/24                    │
│ > Child prefixes 27,27,28,28                   │
│                                                │
│ enter submit • esc back • ctrl+c quit           │
╰────────────────────────────────────────────────╯
```

Result shape:

```text
╭─ Subnet carve result ───────────────────────────╮
│ Parent: 10.0.0.0/24  addresses: 256            │
│ 1  10.0.0.0/27 ...                             │
│                                                │
│ esc back • q quit                               │
╰────────────────────────────────────────────────╯
```

Exact colors may use terminal-safe ANSI colors from Lip Gloss. Tests should assert meaningful text fragments and state, not full ANSI snapshots.

## Architecture

Replace `RunMenu`/`Ask` composition with one app-level model:

```text
cmd/root.go
  builds interactive actions from existing command logic
  calls tui.RunApp(ctx, actions)

internal/tui
  app.go      screen state machine and RunApp
  action.go   Action, Field, Result types
  styles.go   Lip Gloss styles
  menu.go     menu rendering/update helpers
  form.go     form rendering/update helpers
```

`cmd/root.go` should still own business action execution because it already calls `internal/subnet` and `internal/checks`. TUI owns input state, screen transitions, and rendering only.

Proposed interface:

```go
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

func RunApp(ctx context.Context, actions []Action) error
```

The old `Ask` function can be removed if no callers remain. Keeping unused prompt code would preserve the broken mental model.

## Error Handling

- Form Escape returns to Menu and produces no error.
- Result Escape returns to Menu.
- Action validation/runtime errors render in Result screen with red error styling.
- Ctrl+C quits immediately.
- `RunApp` returns only terminal/program-level errors.

## Testing Strategy

Use TDD.

Tests should cover:

- Menu selection opens Form.
- Escape from Form returns to Menu and does not mark app quitting.
- Ctrl+C from Form quits app.
- Submitting a form stores defaults for blank fields and runs the action.
- Action output is stored in Result screen state and rendered by `View`.
- Result Escape returns to Menu.
- Rendered views contain border/title/help fragments without asserting complete ANSI output.

Manual smoke test after implementation:

1. Run `go run .`.
2. Select Subnet carve.
3. Press Escape.
4. Confirm main menu remains open.
5. Select Subnet carve again.
6. Submit defaults.
7. Confirm result appears inside bordered TUI panel.
8. Press Escape to return to menu, then `q` to quit.

## Acceptance Criteria

- `go test ./...` passes.
- `go build ./...` passes.
- `go run .` opens a styled compact-panel TUI.
- Escape inside Subnet/Ping/HTTP forms returns to the main menu, not the shell.
- Action results render inside the TUI panel, not as plain newline output between menu sessions.
- Existing command smoke tests still work:
  - `go run . subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28`
  - `go run ./cmd/nw subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28`
  - `go run . --help`
