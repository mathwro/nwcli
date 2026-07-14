# AGENTS.md

## Project Purpose

`nwcli` is a Go network toolbox for subnet planning and connectivity checks. The first feature set includes IPv4 subnet carving, ICMP ping checks, HTTP/S port checks, and Bubble Tea terminal UI helpers.

## Commands

- Build: `go build ./...`
- Test: `go test ./...`
- Run subnet carving: `go run . subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28`
- Run HTTP/S check: `go run . http example.com --port 443 --tls --path / --timeout 2s`
- Run ping check: `go run . ping 127.0.0.1 --count 1 --timeout 1s`
- Show help: `go run . --help`

## Package Boundaries

- `cmd/` contains Cobra command wiring, flags, command output, and interactive action wiring.
- `internal/subnet/` contains deterministic IPv4 CIDR allocation and rendering.
- `internal/checks/` contains network checks and result aggregation.
- `internal/tui/` contains Bubble Tea models, prompts, menus, and progress UI.

## Development Rules

- Use TDD for behavior-bearing code: write the test, watch it fail, then implement the smallest passing code.
- Keep Cobra files thin; business logic belongs in `internal/*` packages.
- Use stdlib before adding dependencies.
- Do not add a TCP fallback to `ping`; ICMP permission failures must stay explicit.
- Keep live network tests out of the default suite unless they are deterministic and do not require external network access.
