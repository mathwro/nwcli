# nwcli

`nwcli` is a Go network toolbox for subnet planning and connectivity checks. It can be used as a normal Cobra CLI or as an interactive Bubble Tea TUI.

## Features

- IPv4 subnet carving with named blocks, optional VLSM largest-first packing (`--sort`), and a visual allocation map.
- IPv4 and IPv6 CIDR inspection (netmask, wildcard mask, usable host range, RFC scope), containment, overlap, splitting, and summarization.
- DNS record diagnostics with optional resolver selection and accurate `CNAME` alias detection.
- Dual-stack IPv4/IPv6 ICMP ping (finite or continuous live packet-loss stats) and hop-by-hop ICMP traceroute (`trace`).
- Single-port, port-list, and port-range TCP checks, TLS certificate chain inspection (including unverified chain diagnostics), and HTTP/S endpoint checks with `httptrace` phase timing breakdowns.
- Local network interface and CIDR snapshot (`interfaces` / `ip`).
- Bounded concurrent DNS, TCP, TLS, HTTP/S, and Ping batch checks with `--file` / stdin (`-`) target lists.
- Versioned JSON output, stable automation exit codes, and shell completions (`bash`, `zsh`, `fish`, `powershell`).
- Searchable Bubble Tea TUI with structured result summaries, quick re-run (`r`), form editing (`e`), and clipboard copy (`c`).
- Self-update through the Go toolchain with Homebrew/Scoop package-manager guards.

## Install

After the first stable release is available, install the prebuilt executable with Scoop on Windows:

```powershell
scoop bucket add mathwro https://github.com/mathwro/homebrew-tools
scoop install mathwro/nwcli
```

Install with Homebrew on macOS or Linux:

```bash
brew tap mathwro/tools
brew install mathwro/tools/nwcli
```

The Go toolchain remains an alternative installation method:

```bash
go install github.com/mathwro/nwcli@latest
```

After installation:

```bash
nwcli --version
nwcli --help
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
- `←`/`→` or `space`: switch boolean toggles (`ON`/`OFF`) or cycle multi-option selectors.
- `s`: stop a running check and open its result screen.
- `r`: immediately re-run the current check from its result screen.
- `e`: reopen the form from a result screen with your previous inputs preserved.
- `c`: copy the complete test output from a result screen.
- `esc`: go back from forms and results; while a check runs, stop it and open its results.
- `q`: quit outside a form; inside a text field it is normal input.
- `ctrl+c`: quit from any screen.

Forms show examples as placeholders rather than pre-populated text, so typing starts immediately without deleting an example first. Press `enter` without editing a form to run it with the displayed examples. Once a field has been edited, its current content—including an explicitly cleared blank—takes precedence over the placeholder. Result pages keep status and the primary target visible above a scrollable detail view. Subnet carving adds a proportional address-space rail, color-keyed allocations, endpoint ranges, and the remaining available range.

### Subnet carving

Carve a parent IPv4 CIDR into child prefixes in the order provided, or pass `--sort` for automatic VLSM largest-first packing and optional `name:prefix` labels:

```bash
nwcli subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28
nwcli subnet carve --cidr 10.0.0.0/24 --sizes db:28,web:27,mgmt:28 --sort
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

### CIDR calculator

Inspect (including IPv4 netmask/wildcard, usable host range, and RFC address scope) or transform IPv4 and IPv6 CIDRs:

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

Measure transport-level reachability to a single port, comma-separated port list, or port range:

```bash
nwcli tcp example.com:443 --timeout 2s
nwcli tcp example.com:80,443,8080-8085 --concurrency 8
```

### TLS inspection

Verify TLS, inspect the certificate chain (even when x509 verification fails, exiting `3` unhealthy vs `4` unreachable), and warn when the leaf certificate expires inside `--warn-before`:

```bash
nwcli tls example.com:443 --server-name example.com --warn-before 30d
```

### Traceroute & local interfaces

Trace the hop-by-hop ICMP path to a target or inspect local interfaces and assigned CIDRs:

```bash
nwcli trace 1.1.1.1 --max-hops 20 --timeout 1s
nwcli interfaces
nwcli interfaces --all
```

### Bounded batch checks

Run DNS, TCP, TLS, HTTP/S, or Ping checks concurrently from positional arguments or `--file` (`-` for stdin, `#` comments ignored) while preserving input order:

```bash
nwcli batch dns one.example two.example --type A --concurrency 4
nwcli batch tcp one.example:80 two.example:443 --concurrency 4
nwcli batch tls one.example:443 two.example:443 --concurrency 4
nwcli batch http https://one.example/health https://two.example/ready --file targets.txt
nwcli batch ping 1.1.1.1 8.8.8.8 --count 2
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
nwcli http example.com --tls --follow-redirects
```

A completed HTTP response is reachable and reports DNS, TCP, TLS, TTFB, and total phase timings. Status codes `200` through `399` are healthy; redirects are reported directly unless `--follow-redirects` (`-L`) is passed.

### Ping check

Run a finite IPv4 or IPv6 ICMP ping check:

```bash
nwcli ping 127.0.0.1 --count 4 --interval 1s --timeout 1s
```


Stream replies, drops, cumulative counts, and packet-loss percentage until interrupted:
```bash
nwcli ping 127.0.0.1 --continuous --timeout 1s
# Short form:
nwcli ping 127.0.0.1 -c
```

Press `ctrl+c` to stop the CLI cleanly and print its final summary. In the interactive TUI, switch the Ping test's `Continuous` toggle to `ON`; the finite `Count` field disappears because it no longer applies, while `Host` and per-request `Timeout` remain available. The running view updates Sent, Received, Dropped, and Loss after every attempt. Press `s` or `esc` to stop, open the final result screen, and then press `c` to copy the results.

ICMP may require elevated permissions or OS capability configuration. `nwcli` does not silently fall back to TCP because that would make `ping` mean two different things.

### Update nwcli

Update `nwcli` to the latest version available from its Go module:

```bash
nwcli update
```

This direct CLI command runs `go install github.com/mathwro/nwcli@latest` when `go` is on `PATH`. If `nwcli` was installed via Homebrew or Scoop, `nwcli update` detects the managed path and prints the matching `brew upgrade` or `scoop update` command instead of creating a duplicate binary.

### Shell completions

Generate tab-completion scripts for your shell:

```bash
nwcli completion bash
nwcli completion zsh
nwcli completion fish
nwcli completion powershell
```

## Release process

Releases are built by the centralized typed workflow in [`mathwro/homebrew-tools`](https://github.com/mathwro/homebrew-tools). This repository keeps only the declarative `release.json` and thin reusable-workflow callers; the central Go adapter owns tests, six native builds, version and help smoke checks, deterministic archives, checksums, publication, and Scoop/Homebrew notification.

Maintainers push an annotated SemVer tag from protected `main`. Prerelease tags such as `v1.0.0-rc.1` run the complete six-platform build and staging matrix as a dry run without publishing. Stable tags such as `v1.0.0` wait for approval in the protected `release` environment before publishing an immutable release containing six archives and `checksums.txt`.

The repository must define `DISTRIBUTION_DISPATCH_TOKEN` as a fine-grained token restricted to `mathwro/homebrew-tools` with permission to create repository dispatch events. After a stable release publishes, the canonical `cli-release` dispatch opens the package-metadata update in `mathwro/homebrew-tools`; generated Scoop and Homebrew files are never maintained here.

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
- `internal/subnet/`: deterministic subnet allocation and dual-stack CIDR calculations.
- `internal/checks/`: network checks and bounded result aggregation.
- `internal/tui/`: Bubble Tea models, compact-panel TUI, and styling.

See `AGENTS.md` for agent/development rules.

## License

MIT — see [`LICENSE`](LICENSE).
