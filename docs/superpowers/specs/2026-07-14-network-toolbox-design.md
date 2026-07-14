# Network Toolbox Design

## Goal

Build `nwcli` as an expandable Go network toolbox with a full first slice: subnet carving, ICMP ping checks, HTTP/S port checks, Bubble Tea progress/menu UI, and project-local agent guidance.

## Current Project State

- Module: `github.com/mathwro/nwcli`
- Existing entry point: `main.go` calls `cmd.Execute()`.
- Existing CLI scaffold: `cmd/root.go` from Cobra with placeholder descriptions and a placeholder `--toggle` flag.
- Existing dependencies already include Cobra, Bubble Tea, Bubbles, and Lip Gloss through `go.mod`.
- The working directory is not currently inside a Git repository, so the design document cannot be committed from this checkout.

## Scope

### In Scope

- Replace the placeholder Cobra root command with a real `nwcli` root command.
- Add command-based flows:
  - `nwcli subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28`
  - `nwcli ping example.com --count 4 --timeout 1s`
  - `nwcli http example.com --port 443 --tls --path / --timeout 2s`
- Add no-argument interactive Bubble Tea menu with actions for subnet carving, ping test, HTTP/S test, and quit.
- Add Bubble Tea progress/spinner UI for ping and HTTP/S tests.
- Add visual subnet carve output that is useful in terminal logs.
- Add `AGENTS.md` documenting project commands, architecture, and contribution rules.
- Add focused tests for subnet allocation, HTTP checks, Cobra command output, and stable TUI model behavior.

### Out of Scope for First Slice

- IPv6 subnet carving.
- JSON/YAML output.
- Named allocation blocks.
- Persistent config files.
- Concurrent scan ranges.
- DNS diagnostics beyond host resolution required by ping and HTTP/S.
- Fully custom TUI forms beyond minimal prompts matching CLI flags.

## Architecture

Use small Cobra commands backed by pure internal packages. Cobra owns argument parsing and output wiring. Internal packages own behavior. Bubble Tea owns only presentation and user interaction.

```text
main.go
  -> cmd.Execute()
     -> cmd/root.go        root command and no-arg menu launcher
     -> cmd/subnet.go      subnet command group and carve subcommand
     -> cmd/ping.go        ping command
     -> cmd/http.go        http command

internal/subnet
  -> parse requested sizes
  -> allocate child CIDRs inside a parent IPv4 CIDR
  -> render allocation tables and bars

internal/checks
  -> ICMP ping checks
  -> HTTP/S checks
  -> shared result types

internal/tui
  -> no-arg menu model
  -> progress model helpers
  -> simple input prompts for interactive flows
```

This keeps the expandable boundary boring: future commands get a Cobra file, an internal package when behavior is non-trivial, and optional TUI wrappers when interactive presentation is useful.

## Command Design

### Root Command

`nwcli` with no arguments starts the Bubble Tea menu.

`nwcli --help` prints Cobra help and does not start the menu.

Root metadata:

- Use: `nwcli`
- Short: `Network toolbox for subnet planning and connectivity checks`
- Long: one short paragraph describing subnet carving, ping checks, and HTTP/S checks.

Remove the generated placeholder text and placeholder `--toggle` flag.

### Subnet Carving

Command:

```bash
nwcli subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28
```

Behavior:

- First build supports IPv4 only.
- `--cidr` is required and must parse as an IPv4 CIDR.
- `--sizes` is required and contains comma-separated child prefix lengths.
- Requested child prefixes are allocated in the exact order provided.
- Each child prefix must be greater than or equal to the parent prefix and less than or equal to `/32`.
- If any requested child subnet does not fit, return an error and print no partial allocation table.
- Allocation starts at the first address of the parent range and advances by each child block size.
- Child ranges must not overlap and must stay within the parent range.

Example output shape:

```text
Parent: 10.0.0.0/24  addresses: 256

#  CIDR           FIRST        LAST          ADDRS  MAP
1  10.0.0.0/27    10.0.0.0    10.0.0.31    32     ########........................
2  10.0.0.32/27   10.0.0.32   10.0.0.63    32     ........########................
3  10.0.0.64/28   10.0.0.64   10.0.0.79    16     ................####............
4  10.0.0.80/28   10.0.0.80   10.0.0.95    16     ....................####........

Unused: 10.0.0.96 - 10.0.0.255  addresses: 160
```

The bar is proportional to the parent range using a fixed width. For small terminal output, approximate proportions are acceptable, but every allocation must occupy at least one visible character.

### Ping Check

Command:

```bash
nwcli ping example.com --count 4 --timeout 1s
```

Behavior:

- Sends ICMP echo requests with `golang.org/x/net/icmp`.
- Default count: `4`.
- Default timeout per request: `1s`.
- Shows Bubble Tea spinner/progress while attempts run.
- Prints one line per attempt with success/failure and latency.
- Prints a summary with sent, received, loss percentage, min/avg/max latency for successful replies.
- If raw socket or ICMP permissions are denied, fail with a clear error explaining that ICMP may require elevated permissions or OS capability configuration.
- Do not silently fall back to TCP. A TCP fallback would make `ping` mean two different things.

### HTTP/S Check

Command:

```bash
nwcli http example.com --port 443 --tls --path / --timeout 2s
```

Behavior:

- Uses Go stdlib `net/http`.
- Builds `http://host:port/path` unless `--tls` is set, then `https://host:port/path`.
- Default port: `80` without `--tls`, `443` with `--tls`.
- Default path: `/`.
- Default timeout: `2s`.
- Shows Bubble Tea spinner/progress while the request runs.
- Prints status code, latency, and final success/failure.
- Treat any completed HTTP response as reachable; mark status `200` through `399` as healthy and other status codes as reachable but unhealthy.
- Return non-zero for transport errors and unhealthy HTTP status codes.

## Interactive Menu

When `nwcli` is run without subcommands or flags:

```text
nwcli

> Subnet carve
  Ping test
  HTTP/S test
  Quit
```

Rules:

- Arrow keys and `j/k` move selection.
- Enter chooses an item.
- `q`, `esc`, and `ctrl+c` quit.
- Each action prompts for the same fields accepted by flags.
- After an action completes, show the same result renderer used by the command and allow returning to the menu.
- Interactive flows call the same internal functions as command flows. No duplicate networking or subnet logic lives in TUI code.

## Error Handling

- Parse and validation errors return actionable messages without stack traces.
- Cobra command errors should be returned, not `os.Exit` from internal packages.
- Internal packages return typed or wrapped errors where callers need to distinguish validation from runtime failures.
- Network checks accept `context.Context` so Cobra and TUI flows can cancel cleanly.
- TUI cancellation exits without printing partial success claims.

## Testing Strategy

Use TDD for behavior-bearing production code.

Subnet tests:

- Allocates `/27,/27,/28,/28` inside `/24` in exact order.
- Rejects child prefix shorter than parent prefix.
- Rejects allocations that exceed the parent range.
- Renders the expected CIDR rows and unused range.

HTTP tests:

- Uses `httptest.Server` to verify healthy `2xx` response.
- Verifies unhealthy non-`2xx-3xx` response returns reachable but unhealthy.
- Verifies transport errors return non-zero command errors.

Cobra tests:

- `subnet carve` writes deterministic output to a buffer.
- Missing required flags return useful errors.
- Root with no args launches the menu through an injectable runner so tests do not require a real terminal.

TUI tests:

- Menu model moves selection and quits on expected keys.
- Progress model reaches complete state after all check events.
- Avoid full terminal snapshots; test stable model state and key output fragments.

Ping tests:

- Unit-test result aggregation without raw sockets.
- Keep live ICMP tests out of the default suite because they depend on OS permissions and network availability.

## Dependencies

Keep the requested stack:

- `github.com/spf13/cobra`
- `charm.land/bubbletea/v2`
- `charm.land/bubbles/v2`

Add only what is needed for true ICMP ping:

- `golang.org/x/net/icmp`
- `golang.org/x/net/ipv4`

Use stdlib for HTTP/S, URL construction, timing, context, and table formatting.

## AGENTS.md Content

Create `AGENTS.md` at the repository root with:

- Project purpose.
- Build and test commands.
- Package boundaries.
- TDD requirement for behavior-bearing code.
- Guidance to keep Cobra command files thin and business logic in `internal/*` packages.
- Guidance not to add dependencies when stdlib covers the need.

## Acceptance Criteria

- `go test ./...` passes.
- `go run . subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28` prints the expected CIDR rows and unused range.
- `go run . http example.com --port 443 --tls --path / --timeout 2s` performs one HTTP/S check and prints status plus latency.
- `go run . ping 127.0.0.1 --count 1 --timeout 1s` either succeeds or returns the documented ICMP permission error; it must not panic.
- `go run . --help` prints Cobra help.
- `go run .` opens the interactive menu in a terminal.
- `AGENTS.md` exists and documents the project workflow.
