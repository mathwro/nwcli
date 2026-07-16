# nwcli

`nwcli` is a Go network toolbox for subnet planning and connectivity checks. It can be used as a normal Cobra CLI or as an interactive Bubble Tea TUI.

## Features

- IPv4 subnet carving with a visual allocation map.
- IPv4 and IPv6 CIDR inspection, containment, overlap, splitting, and summarization.
- DNS record diagnostics with optional resolver selection.
- ICMP ping with finite summaries or continuous live packet-loss statistics, plus TCP, TLS certificate, and HTTP/S endpoint checks.
- Bounded concurrent DNS, TCP, and TLS batch checks.
- Versioned JSON output and stable automation exit codes.
- Searchable Bubble Tea TUI with structured result summaries and scrollable details.
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
- `/`: filter the action list.
- `enter`: select an action or run the current form.
- `←`/`→` or `space`: switch boolean controls between `ON` and `OFF`.
- `s`: stop a running check and open its result screen.
- `c`: copy the complete test output from a result screen.
- `esc`: go back from forms and results; while a check runs, stop it and open its results.
- `q`: quit outside a form; inside a text field it is normal input.
- `ctrl+c`: quit from any screen.

Forms show examples as placeholders rather than pre-populated text, so typing starts immediately without deleting an example first. Press `enter` without editing a form to run it with the displayed examples. Once a field has been edited, its current content—including an explicitly cleared blank—takes precedence over the placeholder. Result pages keep status and the primary target visible above a scrollable detail view. Subnet carving adds a proportional address-space rail, color-keyed allocations, endpoint ranges, and the remaining available range.

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

### CIDR calculator

Inspect or transform IPv4 and IPv6 CIDRs:

```bash
nwcli subnet inspect 10.0.0.37/24
nwcli subnet contains 10.0.0.0/24 10.0.0.42
nwcli subnet overlap 10.0.0.0/24 10.0.0.128/25
nwcli subnet split 10.0.0.0/24 --prefix 26
nwcli subnet summarize 10.0.0.0/25 10.0.0.128/25
```

### DNS diagnostics

Query common DNS record types through the system resolver or a selected resolver:

```bash
nwcli dns example.com --type AAAA --server 1.1.1.1 --timeout 2s
```

Supported record types are `A`, `AAAA`, `CNAME`, `MX`, `NS`, `TXT`, and `PTR`.

### TCP connection check

Measure transport-level reachability without changing the meaning of ICMP ping:

```bash
nwcli tcp example.com:443 --timeout 2s
```

### TLS inspection

Verify TLS, inspect the certificate chain, and return unhealthy when the leaf certificate expires inside the warning window:

```bash
nwcli tls example.com:443 --server-name example.com --warn-before 30d
```

### Bounded batch checks

Run DNS, TCP, or TLS checks concurrently while preserving input order:

```bash
nwcli batch dns one.example two.example --type A --concurrency 4
nwcli batch tcp one.example:80 two.example:443 --concurrency 4
nwcli batch tls one.example:443 two.example:443 --concurrency 4
```

### Structured output and exit codes

Every command supports `--output text` (default) or a versioned JSON envelope:

```bash
nwcli --output json tcp example.com:443
```

Continuous ping is a streaming text mode and rejects `--output json`.

Exit codes are stable for automation:

- `0`: success and healthy.
- `2`: invalid command usage or input.
- `3`: reachable but unhealthy.
- `4`: unreachable or runtime network failure.
- `5`: ICMP permission failure.

### HTTP/S check

Check an HTTP/S endpoint:

```bash
nwcli http example.com --port 443 --tls --path / --timeout 2s
```

A completed HTTP response is reachable. Status codes `200` through `399` are healthy; other status codes are reachable but unhealthy.

### Ping check

Run a finite ICMP ping check:

```bash
nwcli ping 127.0.0.1 --count 4 --timeout 1s
```

Stream replies, drops, cumulative counts, and packet-loss percentage until interrupted:

```bash
nwcli ping 127.0.0.1 --continuous --timeout 1s
# Short form:
nwcli ping 127.0.0.1 -c
```

Press `ctrl+c` to stop the CLI cleanly and print its final summary. In the interactive TUI, switch the Ping test's `Continuous` toggle to `ON`; the finite `Count` field disappears because it no longer applies, while `Host` and per-request `Timeout` remain available. The running view updates Sent, Received, Dropped, and Loss after every attempt. Press `s` or `esc` to stop, open the final result screen, and then press `c` to copy the results.

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
- `internal/subnet/`: deterministic subnet allocation and dual-stack CIDR calculations.
- `internal/checks/`: network checks and bounded result aggregation.
- `internal/tui/`: Bubble Tea models, compact-panel TUI, and styling.

See `AGENTS.md` for agent/development rules.
