# Network Toolbox Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first complete `nwcli` network toolbox slice: subnet carving, ICMP ping, HTTP/S checks, no-arg Bubble Tea menu, progress UI, and project agent guidance.

**Architecture:** Keep Cobra command files thin and move behavior into focused `internal/*` packages. `internal/subnet` is pure deterministic CIDR math and rendering, `internal/checks` owns network probes and result aggregation, and `internal/tui` owns Bubble Tea models and terminal interaction without duplicating business logic.

**Tech Stack:** Go 1.26.4, Cobra, `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, Go stdlib networking, `golang.org/x/net/icmp`, `golang.org/x/net/ipv4`.

## Global Constraints

- First slice includes subnet carving, ping, HTTP/S port checks, Bubble Tea menu/progress UI, and `AGENTS.md`.
- First slice supports IPv4 subnet carving only.
- Subnet allocations preserve user-provided prefix order.
- Commands return errors; internal packages must not call `os.Exit`.
- No silent TCP fallback for `ping`; ICMP permission failures return clear errors.
- Use stdlib for HTTP/S, URLs, contexts, timers, and formatting.
- Do not add dependencies beyond Cobra, Bubble Tea, Bubbles, and the required `golang.org/x/net` ICMP packages.
- Use TDD for behavior-bearing production code.
- Current baseline: `go test ./...` passes with existing scaffold.

---

## File Structure

- Modify: `cmd/root.go` — root command metadata, `Execute`, injectable menu runner, command registration.
- Create: `cmd/subnet.go` — `subnet carve` command and flag parsing.
- Create: `cmd/ping.go` — `ping` command flags, progress runner, result rendering.
- Create: `cmd/http.go` — `http` command flags, progress runner, result rendering.
- Create: `cmd/command_test.go` — Cobra command tests with buffer output and injected runners.
- Create: `internal/subnet/subnet.go` — IPv4 CIDR parsing, allocation structs, allocation algorithm.
- Create: `internal/subnet/render.go` — deterministic text renderer and allocation bar.
- Create: `internal/subnet/subnet_test.go` — allocation and rendering tests.
- Create: `internal/checks/http.go` — HTTP/S checker and result type.
- Create: `internal/checks/http_test.go` — `httptest` coverage for healthy, unhealthy, and transport failures.
- Create: `internal/checks/ping.go` — ping option/result types, ICMP checker, summary aggregation.
- Create: `internal/checks/ping_test.go` — summary aggregation and permission error text tests.
- Create: `internal/tui/menu.go` — Bubble Tea menu model and menu runner.
- Create: `internal/tui/progress.go` — reusable progress/spinner event model.
- Create: `internal/tui/prompt.go` — minimal text prompts for interactive flows.
- Create: `internal/tui/tui_test.go` — stable model update tests.
- Create: `AGENTS.md` — project-specific agent instructions.

---

### Task 1: Subnet Allocator and Renderer

**Files:**
- Create: `internal/subnet/subnet_test.go`
- Create: `internal/subnet/subnet.go`
- Create: `internal/subnet/render.go`

**Interfaces:**
- Produces: `func ParsePrefixes(raw string) ([]int, error)`
- Produces: `func Plan(parentCIDR string, prefixes []int) (AllocationPlan, error)`
- Produces: `func Render(w io.Writer, plan AllocationPlan) error`
- Produces: `type AllocationPlan struct { Parent *net.IPNet; ParentFirst net.IP; ParentLast net.IP; ParentSize uint32; Blocks []Block; Unused *Range }`
- Produces: `type Block struct { Index int; CIDR *net.IPNet; First net.IP; Last net.IP; Size uint32; StartOffset uint32 }`
- Produces: `type Range struct { First net.IP; Last net.IP; Size uint32; StartOffset uint32 }`
- Consumes: only Go stdlib.

- [ ] **Step 1: Write failing subnet tests**

Create `internal/subnet/subnet_test.go`:

```go
package subnet

import (
	"bytes"
	"strings"
	"testing"
)

func TestParsePrefixesParsesCommaSeparatedCIDRLengths(t *testing.T) {
	got, err := ParsePrefixes("27,27,28,28")
	if err != nil {
		t.Fatalf("ParsePrefixes returned error: %v", err)
	}
	want := []int{27, 27, 28, 28}
	if len(got) != len(want) {
		t.Fatalf("len(ParsePrefixes) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParsePrefixes[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestPlanAllocatesIPv4BlocksInRequestedOrder(t *testing.T) {
	plan, err := Plan("10.0.0.0/24", []int{27, 27, 28, 28})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	wantCIDRs := []string{"10.0.0.0/27", "10.0.0.32/27", "10.0.0.64/28", "10.0.0.80/28"}
	wantFirst := []string{"10.0.0.0", "10.0.0.32", "10.0.0.64", "10.0.0.80"}
	wantLast := []string{"10.0.0.31", "10.0.0.63", "10.0.0.79", "10.0.0.95"}
	wantSize := []uint32{32, 32, 16, 16}

	if plan.ParentSize != 256 {
		t.Fatalf("ParentSize = %d, want 256", plan.ParentSize)
	}
	if len(plan.Blocks) != len(wantCIDRs) {
		t.Fatalf("len(Blocks) = %d, want %d", len(plan.Blocks), len(wantCIDRs))
	}
	for i, block := range plan.Blocks {
		if block.CIDR.String() != wantCIDRs[i] || block.First.String() != wantFirst[i] || block.Last.String() != wantLast[i] || block.Size != wantSize[i] {
			t.Fatalf("block %d = %s %s-%s size %d, want %s %s-%s size %d", i, block.CIDR, block.First, block.Last, block.Size, wantCIDRs[i], wantFirst[i], wantLast[i], wantSize[i])
		}
	}
	if plan.Unused == nil {
		t.Fatal("Unused is nil, want remaining parent range")
	}
	if plan.Unused.First.String() != "10.0.0.96" || plan.Unused.Last.String() != "10.0.0.255" || plan.Unused.Size != 160 {
		t.Fatalf("Unused = %s-%s size %d, want 10.0.0.96-10.0.0.255 size 160", plan.Unused.First, plan.Unused.Last, plan.Unused.Size)
	}
}

func TestPlanRejectsChildPrefixShorterThanParent(t *testing.T) {
	_, err := Plan("10.0.0.0/24", []int{23})
	if err == nil {
		t.Fatal("Plan returned nil error, want child prefix validation error")
	}
	if !strings.Contains(err.Error(), "child prefix /23 is shorter than parent /24") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestPlanRejectsAllocationsThatExceedParent(t *testing.T) {
	_, err := Plan("10.0.0.0/30", []int{31, 31, 31})
	if err == nil {
		t.Fatal("Plan returned nil error, want capacity error")
	}
	if !strings.Contains(err.Error(), "does not fit") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestRenderPrintsRowsAndUnusedRange(t *testing.T) {
	plan, err := Plan("10.0.0.0/24", []int{27, 27, 28, 28})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	var out bytes.Buffer
	if err := Render(&out, plan); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	text := out.String()
	for _, fragment := range []string{
		"Parent: 10.0.0.0/24  addresses: 256",
		"1  10.0.0.0/27",
		"2  10.0.0.32/27",
		"3  10.0.0.64/28",
		"4  10.0.0.80/28",
		"Unused: 10.0.0.96 - 10.0.0.255  addresses: 160",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("render output missing %q:\n%s", fragment, text)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify RED**

Run: `go test ./internal/subnet`

Expected: FAIL because `ParsePrefixes`, `Plan`, `AllocationPlan`, and `Render` are undefined.

- [ ] **Step 3: Implement IPv4 subnet allocation**

Create `internal/subnet/subnet.go` with these concrete rules:

```go
package subnet

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type AllocationPlan struct {
	Parent      *net.IPNet
	ParentFirst net.IP
	ParentLast  net.IP
	ParentSize  uint32
	Blocks      []Block
	Unused      *Range
}

type Block struct {
	Index       int
	CIDR        *net.IPNet
	First       net.IP
	Last        net.IP
	Size        uint32
	StartOffset uint32
}

type Range struct {
	First       net.IP
	Last        net.IP
	Size        uint32
	StartOffset uint32
}

func ParsePrefixes(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	prefixes := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty prefix in %q", raw)
		}
		prefix, err := strconv.Atoi(strings.TrimPrefix(part, "/"))
		if err != nil {
			return nil, fmt.Errorf("invalid prefix %q: %w", part, err)
		}
		if prefix < 0 || prefix > 32 {
			return nil, fmt.Errorf("prefix /%d outside IPv4 range /0-/32", prefix)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func Plan(parentCIDR string, prefixes []int) (AllocationPlan, error) {
	ip, network, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return AllocationPlan{}, fmt.Errorf("invalid parent CIDR %q: %w", parentCIDR, err)
	}
	parentIP := ip.To4()
	if parentIP == nil {
		return AllocationPlan{}, fmt.Errorf("parent CIDR %q is not IPv4", parentCIDR)
	}
	ones, bits := network.Mask.Size()
	if bits != 32 {
		return AllocationPlan{}, fmt.Errorf("parent CIDR %q is not IPv4", parentCIDR)
	}
	base := ipToUint32(parentIP.Mask(network.Mask))
	parentSize := uint32(1) << uint32(32-ones)
	plan := AllocationPlan{
		Parent:      &net.IPNet{IP: uint32ToIP(base), Mask: network.Mask},
		ParentFirst: uint32ToIP(base),
		ParentLast:  uint32ToIP(base + parentSize - 1),
		ParentSize:  parentSize,
		Blocks:      make([]Block, 0, len(prefixes)),
	}

	var offset uint32
	for i, prefix := range prefixes {
		if prefix < ones {
			return AllocationPlan{}, fmt.Errorf("child prefix /%d is shorter than parent /%d", prefix, ones)
		}
		if prefix > 32 {
			return AllocationPlan{}, fmt.Errorf("child prefix /%d outside IPv4 range /0-/32", prefix)
		}
		size := uint32(1) << uint32(32-prefix)
		if offset > parentSize || size > parentSize-offset {
			return AllocationPlan{}, fmt.Errorf("child prefix /%d at position %d does not fit in %s", prefix, i+1, plan.Parent.String())
		}
		first := base + offset
		last := first + size - 1
		plan.Blocks = append(plan.Blocks, Block{
			Index:       i + 1,
			CIDR:        &net.IPNet{IP: uint32ToIP(first), Mask: net.CIDRMask(prefix, 32)},
			First:       uint32ToIP(first),
			Last:        uint32ToIP(last),
			Size:        size,
			StartOffset: offset,
		})
		offset += size
	}
	if offset < parentSize {
		plan.Unused = &Range{First: uint32ToIP(base + offset), Last: uint32ToIP(base + parentSize - 1), Size: parentSize - offset, StartOffset: offset}
	}
	return plan, nil
}

func ipToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func uint32ToIP(v uint32) net.IP {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	return net.IPv4(b[0], b[1], b[2], b[3])
}
```

- [ ] **Step 4: Implement deterministic renderer**

Create `internal/subnet/render.go`:

```go
package subnet

import (
	"fmt"
	"io"
	"strings"
)

const barWidth = 32

func Render(w io.Writer, plan AllocationPlan) error {
	if _, err := fmt.Fprintf(w, "Parent: %s  addresses: %d\n\n", plan.Parent.String(), plan.ParentSize); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "#  CIDR           FIRST        LAST          ADDRS  MAP"); err != nil {
		return err
	}
	for _, block := range plan.Blocks {
		if _, err := fmt.Fprintf(w, "%-2d %-14s %-12s %-12s %-6d %s\n", block.Index, block.CIDR.String(), block.First.String(), block.Last.String(), block.Size, allocationBar(plan.ParentSize, block.StartOffset, block.Size)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if plan.Unused != nil {
		_, err := fmt.Fprintf(w, "Unused: %s - %s  addresses: %d\n", plan.Unused.First.String(), plan.Unused.Last.String(), plan.Unused.Size)
		return err
	}
	_, err := fmt.Fprintln(w, "Unused: none")
	return err
}

func allocationBar(parentSize, start, size uint32) string {
	if parentSize == 0 || size == 0 {
		return strings.Repeat(".", barWidth)
	}
	startCell := int((uint64(start) * barWidth) / uint64(parentSize))
	endCell := int((uint64(start+size) * barWidth) / uint64(parentSize))
	if endCell <= startCell {
		endCell = startCell + 1
	}
	if endCell > barWidth {
		endCell = barWidth
	}
	var b strings.Builder
	b.Grow(barWidth)
	for i := 0; i < barWidth; i++ {
		if i >= startCell && i < endCell {
			b.WriteByte('#')
		} else {
			b.WriteByte('.')
		}
	}
	return b.String()
}
```

- [ ] **Step 5: Run tests to verify GREEN**

Run: `go test ./internal/subnet`

Expected: PASS.

- [ ] **Step 6: Commit subnet package**

Run:

```bash
git add internal/subnet
git commit -m "feat: add subnet carving planner"
```

Expected: commit succeeds.

---

### Task 2: HTTP/S Checker

**Files:**
- Create: `internal/checks/http_test.go`
- Create: `internal/checks/http.go`

**Interfaces:**
- Produces: `type HTTPOptions struct { Host string; Port int; TLS bool; Path string; Timeout time.Duration }`
- Produces: `type HTTPResult struct { URL string; StatusCode int; Latency time.Duration; Reachable bool; Healthy bool; Err error }`
- Produces: `func CheckHTTP(ctx context.Context, client *http.Client, opts HTTPOptions) HTTPResult`
- Consumes: Go stdlib only.

- [ ] **Step 1: Write failing HTTP tests**

Create `internal/checks/http_test.go`:

```go
package checks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCheckHTTPMarks2xxReachableAndHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	host, port := hostPortFromURL(t, server.URL)
	result := CheckHTTP(context.Background(), server.Client(), HTTPOptions{Host: host, Port: port, Path: "/", Timeout: time.Second})

	if result.Err != nil {
		t.Fatalf("Err = %v", result.Err)
	}
	if !result.Reachable || !result.Healthy || result.StatusCode != http.StatusNoContent {
		t.Fatalf("result = %+v, want reachable healthy 204", result)
	}
	if result.Latency <= 0 {
		t.Fatalf("Latency = %v, want positive", result.Latency)
	}
}

func TestCheckHTTPMarks404ReachableButUnhealthy(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	host, port := hostPortFromURL(t, server.URL)
	result := CheckHTTP(context.Background(), server.Client(), HTTPOptions{Host: host, Port: port, Path: "/missing", Timeout: time.Second})

	if result.Err != nil {
		t.Fatalf("Err = %v", result.Err)
	}
	if !result.Reachable || result.Healthy || result.StatusCode != http.StatusNotFound {
		t.Fatalf("result = %+v, want reachable unhealthy 404", result)
	}
}

func TestCheckHTTPReportsTransportError(t *testing.T) {
	result := CheckHTTP(context.Background(), http.DefaultClient, HTTPOptions{Host: "127.0.0.1", Port: 1, Path: "/", Timeout: 20 * time.Millisecond})
	if result.Err == nil {
		t.Fatal("Err = nil, want transport error")
	}
	if result.Reachable || result.Healthy {
		t.Fatalf("result = %+v, want unreachable unhealthy", result)
	}
}

func TestCheckHTTPBuildsHTTPSURL(t *testing.T) {
	result := CheckHTTP(context.Background(), roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://example.com:443/health" {
			t.Fatalf("URL = %q", req.URL.String())
		}
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	}), HTTPOptions{Host: "example.com", Port: 443, TLS: true, Path: "health", Timeout: time.Second})
	if result.Err != nil || !strings.HasPrefix(result.URL, "https://example.com:443/health") {
		t.Fatalf("result = %+v", result)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func hostPortFromURL(t *testing.T, raw string) (string, int) {
	t.Helper()
	trimmed := strings.TrimPrefix(raw, "http://")
	parts := strings.Split(trimmed, ":")
	if len(parts) != 2 {
		t.Fatalf("unexpected server URL %q", raw)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		t.Fatalf("invalid port in %q: %v", raw, err)
	}
	return parts[0], port
}
```

Run this test file with the import list exactly as shown above.

- [ ] **Step 2: Run tests to verify RED**

Run: `go test ./internal/checks -run TestCheckHTTP`

Expected: FAIL because `CheckHTTP`, `HTTPOptions`, and `HTTPResult` are undefined.

- [ ] **Step 3: Implement HTTP checker**

Create `internal/checks/http.go`:

```go
package checks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type HTTPOptions struct {
	Host    string
	Port    int
	TLS     bool
	Path    string
	Timeout time.Duration
}

type HTTPResult struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Reachable  bool
	Healthy    bool
	Err        error
}

func CheckHTTP(ctx context.Context, client *http.Client, opts HTTPOptions) HTTPResult {
	if client == nil {
		client = http.DefaultClient
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 2 * time.Second
	}
	if opts.Port == 0 {
		if opts.TLS {
			opts.Port = 443
		} else {
			opts.Port = 80
		}
	}
	path := opts.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	scheme := "http"
	if opts.TLS {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s:%d%s", scheme, opts.Host, opts.Port, path)
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return HTTPResult{URL: url, Err: err}
	}
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return HTTPResult{URL: url, Latency: latency, Err: err}
	}
	defer resp.Body.Close()
	healthy := resp.StatusCode >= 200 && resp.StatusCode <= 399
	return HTTPResult{URL: url, StatusCode: resp.StatusCode, Latency: latency, Reachable: true, Healthy: healthy}
}
```

- [ ] **Step 4: Fix the test imports and run GREEN**

Ensure `internal/checks/http_test.go` imports `strconv`.

Run: `go test ./internal/checks -run TestCheckHTTP`

Expected: PASS.

- [ ] **Step 5: Commit HTTP checker**

Run:

```bash
git add internal/checks/http.go internal/checks/http_test.go
git commit -m "feat: add http reachability checker"
```

Expected: commit succeeds.

---

### Task 3: Ping Summary and ICMP Checker

**Files:**
- Create: `internal/checks/ping_test.go`
- Create: `internal/checks/ping.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Produces: `type PingOptions struct { Host string; Count int; Timeout time.Duration }`
- Produces: `type PingAttempt struct { Seq int; Latency time.Duration; Err error }`
- Produces: `type PingSummary struct { Sent int; Received int; LossPercent float64; Min time.Duration; Avg time.Duration; Max time.Duration }`
- Produces: `func SummarizePing(attempts []PingAttempt) PingSummary`
- Produces: `func CheckPing(ctx context.Context, opts PingOptions) ([]PingAttempt, PingSummary, error)`
- Produces: `func IsICMPPermissionError(err error) bool`
- Consumes: `golang.org/x/net/icmp`, `golang.org/x/net/ipv4`.

- [ ] **Step 1: Add required ICMP modules**

Run:

```bash
go get golang.org/x/net/icmp golang.org/x/net/ipv4
```

Expected: `go.mod` and `go.sum` update with `golang.org/x/net` entries.

- [ ] **Step 2: Write failing ping tests**

Create `internal/checks/ping_test.go`:

```go
package checks

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSummarizePingCalculatesLossAndLatency(t *testing.T) {
	attempts := []PingAttempt{
		{Seq: 1, Latency: 10 * time.Millisecond},
		{Seq: 2, Err: errors.New("timeout")},
		{Seq: 3, Latency: 30 * time.Millisecond},
	}
	summary := SummarizePing(attempts)
	if summary.Sent != 3 || summary.Received != 2 {
		t.Fatalf("summary counts = sent %d received %d, want sent 3 received 2", summary.Sent, summary.Received)
	}
	if summary.LossPercent < 33.3 || summary.LossPercent > 33.4 {
		t.Fatalf("LossPercent = %.2f, want about 33.33", summary.LossPercent)
	}
	if summary.Min != 10*time.Millisecond || summary.Avg != 20*time.Millisecond || summary.Max != 30*time.Millisecond {
		t.Fatalf("latency = min %v avg %v max %v", summary.Min, summary.Avg, summary.Max)
	}
}

func TestSummarizePingHandlesAllFailedAttempts(t *testing.T) {
	summary := SummarizePing([]PingAttempt{{Seq: 1, Err: errors.New("timeout")}})
	if summary.Sent != 1 || summary.Received != 0 || summary.LossPercent != 100 {
		t.Fatalf("summary = %+v, want one sent, zero received, 100%% loss", summary)
	}
	if summary.Min != 0 || summary.Avg != 0 || summary.Max != 0 {
		t.Fatalf("latency = min %v avg %v max %v, want zeroes", summary.Min, summary.Avg, summary.Max)
	}
}

func TestICMPPermissionErrorTextIsActionable(t *testing.T) {
	err := ErrICMPPermission{Cause: errors.New("operation not permitted")}
	text := err.Error()
	for _, fragment := range []string{"ICMP", "permission", "elevated", "capability"} {
		if !strings.Contains(strings.ToLower(text), strings.ToLower(fragment)) {
			t.Fatalf("permission error %q missing %q", text, fragment)
		}
	}
}
```

- [ ] **Step 3: Run tests to verify RED**

Run: `go test ./internal/checks -run 'TestSummarizePing|TestICMPPermission'`

Expected: FAIL because ping types and functions are undefined.

- [ ] **Step 4: Implement ping summary and ICMP checker**

Create `internal/checks/ping.go` with:

- `ErrICMPPermission` error type whose message says ICMP may require elevated permissions or OS capability configuration.
- `SummarizePing` that only includes successful attempts in min/avg/max.
- `CheckPing` that resolves the host, opens an ICMP packet connection, sends `Count` echo requests, and records latency or per-attempt error.
- `Count <= 0` defaults to `4`.
- `Timeout <= 0` defaults to `1s`.
- Use `icmp.ListenPacket("ip4:icmp", "0.0.0.0")` and `icmp.Echo` messages.
- Return `ErrICMPPermission` when listen/send/read fails with permission-like text: `operation not permitted`, `permission denied`, or `raw socket`.

The implementation should keep packet parsing small and direct; no goroutine pool and no fallback protocol.

- [ ] **Step 5: Run tests to verify GREEN**

Run: `go test ./internal/checks -run 'TestSummarizePing|TestICMPPermission|TestCheckHTTP'`

Expected: PASS.

- [ ] **Step 6: Commit ping checker**

Run:

```bash
git add go.mod go.sum internal/checks/ping.go internal/checks/ping_test.go
git commit -m "feat: add icmp ping checker"
```

Expected: commit succeeds.

---

### Task 4: Bubble Tea Menu, Prompts, and Progress Models

**Files:**
- Create: `internal/tui/tui_test.go`
- Create: `internal/tui/menu.go`
- Create: `internal/tui/progress.go`
- Create: `internal/tui/prompt.go`

**Interfaces:**
- Produces: `type MenuItem struct { Title string; Run func(context.Context) (string, error) }`
- Produces: `func NewMenuModel(items []MenuItem) MenuModel`
- Produces: `func RunMenu(ctx context.Context, items []MenuItem) error`
- Produces: `type Field struct { Name string; Prompt string; Default string }`
- Produces: `func Ask(ctx context.Context, title string, fields []Field) (map[string]string, error)`
- Produces: `type ProgressEvent struct { Label string; Current int; Total int; Done bool; Err error }`
- Produces: `func RunProgress(ctx context.Context, title string, events <-chan ProgressEvent) error`
- Consumes: Bubble Tea and Bubbles packages only.

- [ ] **Step 1: Write failing TUI model tests**

Create `internal/tui/tui_test.go`:

```go
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
		t.Fatal("cmd = nil, want tea.Quit command")
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
```

- [ ] **Step 2: Run tests to verify RED**

Run: `go test ./internal/tui`

Expected: FAIL because TUI types and functions are undefined.

- [ ] **Step 3: Implement menu model**

Create `internal/tui/menu.go`:

- `MenuModel` stores `items []MenuItem`, `selected int`, and `quitting bool`.
- `Selected() int` and `Quitting() bool` expose stable test state.
- `Update` handles `up/k`, `down/j`, `enter`, `q`, `esc`, and `ctrl+c`.
- `View` renders a compact list with `>` before the selected item.
- `RunMenu` starts `tea.NewProgram(NewMenuModel(items)).Run()` and returns the error.

Use Bubble Tea v2 message types from `go doc charm.land/bubbletea/v2` as needed.

- [ ] **Step 4: Implement progress model**

Create `internal/tui/progress.go`:

- `ProgressModel` stores title, Bubbles `progress.Model`, Bubbles `spinner.Model`, last label, current, total, done, and error.
- `NewProgressModel(title string) ProgressModel` initializes width to `32`.
- `Done() bool` and `LastLabel() string` expose stable test state.
- `Update` accepts `ProgressEvent`, updates percentage, stores labels/errors, and quits when `Done` is true.
- `RunProgress` reads progress events from a channel through a Bubble Tea command until done.

- [ ] **Step 5: Implement minimal prompts**

Create `internal/tui/prompt.go`:

- `Ask` returns an error when `fields` is empty.
- For real terminal use, implement one text input at a time with Bubbles `textinput.Model`.
- Enter accepts the current field and advances.
- Defaults are used when the user submits an empty value.
- `q`, `esc`, and `ctrl+c` cancel with an error containing `cancelled`.

- [ ] **Step 6: Run tests to verify GREEN**

Run: `go test ./internal/tui`

Expected: PASS.

- [ ] **Step 7: Commit TUI package**

Run:

```bash
git add internal/tui
git commit -m "feat: add interactive tui models"
```

Expected: commit succeeds.

---

### Task 5: Cobra Command Wiring

**Files:**
- Create: `cmd/command_test.go`
- Modify: `cmd/root.go`
- Create: `cmd/subnet.go`
- Create: `cmd/http.go`
- Create: `cmd/ping.go`

**Interfaces:**
- Consumes: `internal/subnet.ParsePrefixes`, `internal/subnet.Plan`, `internal/subnet.Render`.
- Consumes: `internal/checks.CheckHTTP`, `internal/checks.CheckPing`, `internal/checks.SummarizePing`.
- Consumes: `internal/tui.RunMenu`, `internal/tui.Ask`, `internal/tui.RunProgress`.
- Produces: `func NewRootCommand(menuRunner func(context.Context) error) *cobra.Command`.
- Produces: `func Execute()`.

- [ ] **Step 1: Write failing Cobra tests**

Create `cmd/command_test.go`:

```go
package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func executeForTest(args ...string) (string, string, error) {
	root := NewRootCommand(func(context.Context) error { return nil })
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func TestSubnetCarveCommandPrintsAllocation(t *testing.T) {
	stdout, _, err := executeForTest("subnet", "carve", "--cidr", "10.0.0.0/24", "--sizes", "27,27,28,28")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, fragment := range []string{"10.0.0.0/27", "10.0.0.32/27", "10.0.0.64/28", "10.0.0.80/28", "Unused: 10.0.0.96 - 10.0.0.255"} {
		if !strings.Contains(stdout, fragment) {
			t.Fatalf("stdout missing %q:\n%s", fragment, stdout)
		}
	}
}

func TestSubnetCarveRequiresCIDR(t *testing.T) {
	_, _, err := executeForTest("subnet", "carve", "--sizes", "27")
	if err == nil {
		t.Fatal("Execute returned nil error, want required cidr error")
	}
	if !strings.Contains(err.Error(), "required flag") || !strings.Contains(err.Error(), "cidr") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestRootNoArgsRunsInjectedMenu(t *testing.T) {
	run := false
	root := NewRootCommand(func(context.Context) error {
		run = true
		return nil
	})
	root.SetArgs([]string{})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !run {
		t.Fatal("menu runner was not called")
	}
}

func TestHTTPCommandRejectsMissingHost(t *testing.T) {
	_, _, err := executeForTest("http")
	if err == nil {
		t.Fatal("Execute returned nil error, want host arg error")
	}
}
```

- [ ] **Step 2: Run tests to verify RED**

Run: `go test ./cmd`

Expected: FAIL because `NewRootCommand` and subcommands are not implemented.

- [ ] **Step 3: Replace root command scaffold**

Modify `cmd/root.go`:

- Remove generated comments and placeholder `--toggle` flag.
- Add `NewRootCommand(menuRunner func(context.Context) error) *cobra.Command`.
- Root `RunE` calls `menuRunner(cmd.Context())` only when no subcommand is selected.
- Register `newSubnetCommand()`, `newPingCommand()`, and `newHTTPCommand()`.
- `Execute()` calls `NewRootCommand(runInteractiveMenu).Execute()` and exits with code `1` only at the top level.

- [ ] **Step 4: Add subnet command**

Create `cmd/subnet.go`:

- `newSubnetCommand()` returns `subnet` parent command with `carve` subcommand.
- `carve` has required `--cidr` and `--sizes` flags.
- It calls `subnet.ParsePrefixes`, `subnet.Plan`, then `subnet.Render(cmd.OutOrStdout(), plan)`.
- It returns errors directly.

- [ ] **Step 5: Add HTTP command**

Create `cmd/http.go`:

- `newHTTPCommand()` accepts one required host arg.
- Flags: `--port int`, `--tls bool`, `--path string`, `--timeout duration`.
- It calls `checks.CheckHTTP(cmd.Context(), http.DefaultClient, opts)`.
- It renders URL, status, latency, reachable, and healthy to `cmd.OutOrStdout()`.
- It returns an error when transport fails or when `Healthy` is false.
- It can call `tui.RunProgress` around the request, but command output must remain deterministic enough for tests.

- [ ] **Step 6: Add ping command**

Create `cmd/ping.go`:

- `newPingCommand()` accepts one required host arg.
- Flags: `--count int` default `4`, `--timeout duration` default `1s`.
- It calls `checks.CheckPing`.
- It prints each attempt and summary.
- It returns `ErrICMPPermission` text directly for permission failures.

- [ ] **Step 7: Add interactive menu action wiring**

In `cmd/root.go` or a small helper in the same package, implement `runInteractiveMenu(ctx context.Context) error`:

- Build `[]tui.MenuItem` with titles `Subnet carve`, `Ping test`, `HTTP/S test`, `Quit`.
- Each non-quit action uses `tui.Ask` for the same fields as command flags.
- Each action calls the same internal package function used by its command.
- Each action returns rendered output as a string.
- Quit action returns nil.

- [ ] **Step 8: Run command tests to verify GREEN**

Run: `go test ./cmd`

Expected: PASS.

- [ ] **Step 9: Commit command wiring**

Run:

```bash
git add cmd
git commit -m "feat: wire network toolbox commands"
```

Expected: commit succeeds.

---

### Task 6: Agent Guide and End-to-End Verification

**Files:**
- Create: `AGENTS.md`
- Modify: any file found by verification only when the failure is caused by this implementation.

**Interfaces:**
- Consumes: all commands from earlier tasks.
- Produces: repository agent guidance.

- [ ] **Step 1: Write project agent guide**

Create `AGENTS.md`:

```markdown
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
```

- [ ] **Step 2: Run full automated verification**

Run: `go test ./...`

Expected: PASS.

Run: `go build ./...`

Expected: PASS.

- [ ] **Step 3: Run subnet smoke test**

Run:

```bash
go run . subnet carve --cidr 10.0.0.0/24 --sizes 27,27,28,28
```

Expected output contains:

```text
10.0.0.0/27
10.0.0.32/27
10.0.0.64/28
10.0.0.80/28
Unused: 10.0.0.96 - 10.0.0.255  addresses: 160
```

- [ ] **Step 4: Run HTTP/S smoke test**

Run:

```bash
go run . http example.com --port 443 --tls --path / --timeout 2s
```

Expected: command prints URL, status code, and latency. A non-2xx/3xx status is acceptable only if the output marks it reachable but unhealthy and exits non-zero.

- [ ] **Step 5: Run ping smoke test**

Run:

```bash
go run . ping 127.0.0.1 --count 1 --timeout 1s
```

Expected: command either prints one successful ICMP attempt and summary, or exits with the documented ICMP permission/capability error. Panic or unrelated error text fails verification.

- [ ] **Step 6: Run help smoke test**

Run:

```bash
go run . --help
```

Expected output contains:

```text
Network toolbox for subnet planning and connectivity checks
subnet
ping
http
```

- [ ] **Step 7: Manually check interactive menu launches**

Run in a terminal:

```bash
go run .
```

Expected: Bubble Tea menu opens with `Subnet carve`, `Ping test`, `HTTP/S test`, and `Quit`. Press `q` and confirm clean exit.

- [ ] **Step 8: Commit guide and verification fixes**

Run:

```bash
git add AGENTS.md go.mod go.sum cmd internal
git commit -m "docs: add project agent guide"
```

Expected: commit succeeds if files changed since the previous task. If only `AGENTS.md` changed, the commit contains only `AGENTS.md`.
