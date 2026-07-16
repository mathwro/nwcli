# AGENTS.md

## Project Purpose

`nwcli` is a Go network toolbox for subnet planning and connectivity checks. It includes IPv4 subnet carving, dual-stack CIDR calculations, DNS diagnostics, ICMP ping, TCP, TLS, HTTP/S checks, bounded batch execution, structured output, and an interactive Bubble Tea terminal UI.

## Commands

- Build: `go build ./...`
- Test: `go test ./...`
- Install `nwcli`: `go install .`
- Update installed `nwcli`: `nwcli update`
- Run subnet carving: `go run . subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28`
- Run HTTP/S check: `go run . http example.com --port 443 --tls --path / --timeout 2s`
- Run ping check: `go run . ping 127.0.0.1 --count 1 --timeout 1s`
- Show help: `go run . --help`

## Package Boundaries

- `cmd/` contains Cobra command wiring, flags, command output, and interactive action wiring.
- `internal/subnet/` contains deterministic subnet allocation and dual-stack CIDR calculations.
- `internal/checks/` contains network checks and bounded result aggregation.
- `internal/selfupdate/` contains the Go-install self-update operation.
- `internal/tui/` contains Bubble Tea models, prompts, menus, progress UI, and structured result rendering.

## Terminal UI Guidelines

- Treat the TUI as a compact network console: quiet framing, strong information hierarchy, and color that communicates state rather than decorating the screen.
- Reuse the shared Lip Gloss styles in `internal/tui/styles.go`. Keep the established cyan/blue signal palette, green success, amber warning, coral failure, bright data text, and muted secondary text.
- Build interactive action output with `tui.Result`: a stable status header, primary target headline, concise summary, overview fields, and optional detail sections. Keep the original plain-text output in `Result.Text` for clipboard use.
- Every completed or user-stopped action must reach the shared result screen with complete, copyable output. Preserve partial measurements when a check is stopped, and keep `c` wired to copy the displayed result rather than requiring users to select terminal text.
- Put result details in the shared viewport and preserve responsive resizing. Do not create action-specific screen models when the structured result renderer can express the content.
- Reserve bespoke visuals for data with spatial meaning. Subnet carving uses the proportional address-space rail, color-keyed numbered allocations, explicit endpoints, address counts, and a visibly distinct available range.
- Never rely on color alone. Pair state colors with text or glyphs, and pair allocation colors with stable indices, CIDRs, ranges, and counts.
- Show examples as placeholders instead of pre-populated editable text. Untouched text fields submit their placeholders as values; edited fields submit their current content, including an explicitly cleared blank. Render boolean values as keyboard-operable toggles rather than asking users to type `true` or `false`. Conditionally hide and skip validation for fields that do not apply to the selected mode.
- Use existing Bubbles components for lists, text input, viewports, and progress. Keep the established keys: arrows or `j`/`k` navigate, `/` filters, `enter` selects or runs, arrows or `space` operate toggles, `s` stops a running check into its result screen, `c` copies results, `esc` goes back or stops a running check, and `ctrl+c` quits.
- Keep errors actionable and in context. Preserve any partial structured result, label the screen `NEEDS ATTENTION`, and explain the concrete failure instead of showing a generic error state.
- Test observable hierarchy and interaction rather than ANSI escape sequences. Visual changes also require a PTY smoke check at a practical terminal size.

## Development Rules

- Use TDD for behavior-bearing code: write the test, watch it fail, then implement the smallest passing code.
- Keep Cobra files thin; business logic belongs in `internal/*` packages.
- Use stdlib before adding dependencies.
- Do not add a TCP fallback to `ping`; ICMP permission failures must stay explicit.
- Keep live network tests out of the default suite unless they are deterministic and do not require external network access.
