package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
)

func TestBatchTCPCommandPreservesTargetOrder(t *testing.T) {
	original := tcpCheck
	t.Cleanup(func() { tcpCheck = original })
	tcpCheck = func(_ context.Context, _ checks.ContextDialer, opts checks.TCPOptions) checks.TCPResult {
		return checks.TCPResult{Address: opts.Address, Reachable: true}
	}
	stdout, _, err := executeForTest("batch", "tcp", "one.example:80", "two.example:443", "--concurrency", "1")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "one.example:80 ") || !strings.HasPrefix(lines[1], "two.example:443 ") {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestBatchDNSCommandWritesJSONResults(t *testing.T) {
	original := dnsCheck
	t.Cleanup(func() { dnsCheck = original })
	dnsCheck = func(_ context.Context, _ checks.DNSResolver, opts checks.DNSOptions) checks.DNSResult {
		return checks.DNSResult{Query: opts.Host, Type: opts.Type, Records: []string{"192.0.2.1"}, Reachable: true, Healthy: true}
	}
	stdout, _, err := executeForTest("--output", "json", "batch", "dns", "one.example", "two.example")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope struct {
		Command string          `json:"command"`
		Result  []dnsResultJSON `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if envelope.Command != "batch dns" || len(envelope.Result) != 2 || envelope.Result[1].Query != "two.example" {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestBatchTLSCommandReturnsUnhealthyForExpiringCertificate(t *testing.T) {
	original := tlsCheck
	t.Cleanup(func() { tlsCheck = original })
	tlsCheck = func(_ context.Context, opts checks.TLSOptions) checks.TLSResult {
		return checks.TLSResult{Address: opts.Address, ServerName: "example.com", Reachable: true, Verified: true, ExpiringSoon: true}
	}
	_, _, err := executeForTest("batch", "tls", "example.com:443")
	if ExitCode(err) != exitUnhealthy {
		t.Fatalf("ExitCode = %d, err = %v", ExitCode(err), err)
	}
}

func TestBatchHTTPAndPingAndFileInput(t *testing.T) {
	origHTTP := httpCheck
	origPing := pingCheck
	t.Cleanup(func() {
		httpCheck = origHTTP
		pingCheck = origPing
	})
	httpCheck = func(_ context.Context, _ *http.Client, opts checks.HTTPOptions) checks.HTTPResult {
		return checks.HTTPResult{URL: "https://" + opts.Host + ":443/", StatusCode: 200, Latency: 5 * time.Millisecond, Reachable: true, Healthy: true}
	}
	pingCheck = func(_ context.Context, opts checks.PingOptions) ([]checks.PingAttempt, checks.PingSummary, error) {
		return []checks.PingAttempt{{Seq: 1, Latency: 2 * time.Millisecond}}, checks.PingSummary{Sent: 1, Received: 1, Avg: 2 * time.Millisecond}, nil
	}

	targetsFile := filepath.Join(t.TempDir(), "targets.txt")
	if err := os.WriteFile(targetsFile, []byte("# web endpoints\nhttps://one.example/health\ntwo.example\n\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	stdout, _, err := executeForTest("batch", "http", "--file", targetsFile)
	if err != nil {
		t.Fatalf("batch http --file returned error: %v", err)
	}
	if !strings.Contains(stdout, "one.example") || !strings.Contains(stdout, "two.example") {
		t.Fatalf("batch http stdout = %q", stdout)
	}

	pingOut, _, err := executeForTest("--output", "json", "batch", "ping", "192.0.2.1", "192.0.2.2", "--count", "1")
	if err != nil {
		t.Fatalf("batch ping returned error: %v", err)
	}
	var envelope struct {
		Command string           `json:"command"`
		Result  []pingResultJSON `json:"result"`
	}
	if err := json.Unmarshal([]byte(pingOut), &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if envelope.Command != "batch ping" || len(envelope.Result) != 2 {
		t.Fatalf("batch ping envelope = %#v", envelope)
	}
}
