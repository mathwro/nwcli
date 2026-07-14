# nwcli

`nwcli` is a Go network toolbox for subnet planning and connectivity checks. It can be used as a normal Cobra CLI or as an interactive Bubble Tea TUI.

## Features

- IPv4 subnet carving with a visual allocation map.
- ICMP ping checks with summary output.
- HTTP/S endpoint checks by host, port, path, and timeout.
- Interactive compact-panel TUI when run without arguments.
- Optional `nw` command alias in addition to `nwcli`.

## Install

Install the main command:

```bash
go install .
```

Install the shorter alias:

```bash
go install ./cmd/nw
```

After install, run either:

```bash
nwcli --help
nw --help
```

## Usage

### Interactive mode

Run without arguments to open the TUI:

```bash
nwcli
```

Keyboard shortcuts:

- `↑`/`↓` or `k`/`j`: move in the menu.
- `enter`: select or submit.
- `esc`: go back from forms and results.
- `q` or `ctrl+c`: quit.

### Subnet carving

Carve a parent IPv4 CIDR into child prefixes in the order provided:

```bash
nwcli subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28
```

Example output:

```text
Parent: 10.0.0.0/24  addresses: 256

#  CIDR           FIRST        LAST          ADDRS  MAP
1  10.0.0.0/27    10.0.0.0     10.0.0.31    32     ####............................
2  10.0.0.32/27   10.0.0.32    10.0.0.63    32     ....####........................
3  10.0.0.64/28   10.0.0.64    10.0.0.79    16     ........##......................
4  10.0.0.80/28   10.0.0.80    10.0.0.95    16     ..........##....................

Unused: 10.0.0.96 - 10.0.0.255  addresses: 160
```

The `nw` alias works the same way:

```bash
nw subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28
```

### HTTP/S check

Check an HTTP/S endpoint:

```bash
nwcli http example.com --port 443 --tls --path / --timeout 2s
```

A completed HTTP response is reachable. Status codes `200` through `399` are healthy; other status codes are reachable but unhealthy.

### Ping check

Run an ICMP ping check:

```bash
nwcli ping 127.0.0.1 --count 1 --timeout 1s
```

ICMP may require elevated permissions or OS capability configuration. `nwcli` does not silently fall back to TCP because that would make `ping` mean two different things.

## Development

Build:

```bash
go build ./...
```

Test:

```bash
go test ./...
```

Project layout:

- `cmd/`: Cobra command wiring and interactive action wiring.
- `cmd/nw/`: short `nw` alias entry point.
- `internal/subnet/`: deterministic IPv4 CIDR allocation and rendering.
- `internal/checks/`: network checks and result aggregation.
- `internal/tui/`: Bubble Tea models, compact-panel TUI, and styling.

See `AGENTS.md` for agent/development rules.
