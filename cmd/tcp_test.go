package cmd

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
)

func TestTCPCommandRendersConnection(t *testing.T) {
	original := tcpCheck
	t.Cleanup(func() { tcpCheck = original })
	tcpCheck = func(_ context.Context, _ checks.ContextDialer, opts checks.TCPOptions) checks.TCPResult {
		return checks.TCPResult{Address: opts.Address, LocalAddress: "192.0.2.10:50000", RemoteAddress: "192.0.2.20:443", Latency: 8 * time.Millisecond, Reachable: true}
	}
	stdout, _, err := executeForTest("tcp", "example.com:443")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "Address: example.com:443\nRemote: 192.0.2.20:443\nLocal: 192.0.2.10:50000\nLatency: 8ms\nReachable: true\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestTCPCommandWritesJSONEnvelope(t *testing.T) {
	original := tcpCheck
	t.Cleanup(func() { tcpCheck = original })
	tcpCheck = func(_ context.Context, _ checks.ContextDialer, opts checks.TCPOptions) checks.TCPResult {
		return checks.TCPResult{Address: opts.Address, Reachable: true}
	}
	stdout, _, err := executeForTest("--output", "json", "tcp", "example.com:443")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope struct {
		Command string `json:"command"`
		Result  struct {
			Address string `json:"address"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if envelope.Command != "tcp" || envelope.Result.Address != "example.com:443" {
		t.Fatalf("envelope = %#v", envelope)
	}
}
