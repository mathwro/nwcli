package cmd

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
)

func TestDNSCommandRendersRecords(t *testing.T) {
	original := dnsCheck
	t.Cleanup(func() { dnsCheck = original })
	dnsCheck = func(_ context.Context, _ checks.DNSResolver, opts checks.DNSOptions) checks.DNSResult {
		if opts.Host != "example.com" || opts.Type != "AAAA" || opts.Server != "1.1.1.1" {
			t.Fatalf("options = %#v", opts)
		}
		return checks.DNSResult{Query: opts.Host, Type: opts.Type, Server: opts.Server, Records: []string{"2001:db8::1"}, Latency: 12 * time.Millisecond, Reachable: true, Healthy: true}
	}
	stdout, _, err := executeForTest("dns", "example.com", "--type", "AAAA", "--server", "1.1.1.1")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "Query: example.com\nType: AAAA\nServer: 1.1.1.1\nLatency: 12ms\nRecords:\n  2001:db8::1\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestDNSCommandWritesJSONEnvelope(t *testing.T) {
	original := dnsCheck
	t.Cleanup(func() { dnsCheck = original })
	dnsCheck = func(_ context.Context, _ checks.DNSResolver, opts checks.DNSOptions) checks.DNSResult {
		return checks.DNSResult{Query: opts.Host, Type: opts.Type, Records: []string{"192.0.2.1"}, Reachable: true, Healthy: true}
	}
	stdout, _, err := executeForTest("--output", "json", "dns", "example.com")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope struct {
		Command string `json:"command"`
		Result  struct {
			Records []string `json:"records"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if envelope.Command != "dns" || !reflect.DeepEqual(envelope.Result.Records, []string{"192.0.2.1"}) {
		t.Fatalf("envelope = %#v", envelope)
	}
}
