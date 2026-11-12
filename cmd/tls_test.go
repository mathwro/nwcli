package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
)

func TestTLSCommandRendersCertificateDetails(t *testing.T) {
	original := tlsCheck
	t.Cleanup(func() { tlsCheck = original })
	tlsCheck = func(_ context.Context, opts checks.TLSOptions) checks.TLSResult {
		return checks.TLSResult{
			Address: opts.Address, ServerName: opts.ServerName, Latency: 15 * time.Millisecond,
			Reachable: true, Verified: true, Version: "TLS 1.3", CipherSuite: "TLS_AES_128_GCM_SHA256",
			Certificates: []checks.CertificateInfo{{Subject: "CN=example.com", Issuer: "CN=Test CA", DNSNames: []string{"example.com"}, NotAfter: time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC), Serial: "42"}},
		}
	}
	stdout, _, err := executeForTest("tls", "example.com:443", "--server-name", "example.com")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, want := range []string{"Address: example.com:443", "Version: TLS 1.3", "Cipher: TLS_AES_128_GCM_SHA256", "Subject: CN=example.com", "Expires: 2030-01-02T03:04:05Z"} {
		if !containsLine(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestTLSCommandWritesJSONEnvelope(t *testing.T) {
	original := tlsCheck
	t.Cleanup(func() { tlsCheck = original })
	tlsCheck = func(_ context.Context, opts checks.TLSOptions) checks.TLSResult {
		return checks.TLSResult{Address: opts.Address, ServerName: "example.com", Reachable: true, Verified: true}
	}
	stdout, _, err := executeForTest("--output", "json", "tls", "example.com:443")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope struct {
		Command string `json:"command"`
		Result  struct {
			Verified bool `json:"verified"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if envelope.Command != "tls" || !envelope.Result.Verified {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestTLSCommandExitsUnhealthyAndRendersChainWhenVerificationFails(t *testing.T) {
	original := tlsCheck
	t.Cleanup(func() { tlsCheck = original })
	tlsCheck = func(_ context.Context, opts checks.TLSOptions) checks.TLSResult {
		return checks.TLSResult{
			Address: opts.Address, ServerName: "example.com", Reachable: true, Verified: false,
			Certificates: []checks.CertificateInfo{{Subject: "CN=untrusted.local", Issuer: "CN=untrusted.local", NotAfter: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), Serial: "99"}},
			Err:          context.DeadlineExceeded,
		}
	}
	stdout, _, err := executeForTest("tls", "example.com:443")
	if ExitCode(err) != exitUnhealthy {
		t.Fatalf("ExitCode = %d, want %d (err=%v)", ExitCode(err), exitUnhealthy, err)
	}
	if !strings.Contains(stdout, "Subject: CN=untrusted.local") {
		t.Fatalf("stdout missing unverified certificate details:\n%s", stdout)
	}
}

func containsLine(output, line string) bool {
	for _, candidate := range strings.Split(output, "\n") {
		if candidate == line {
			return true
		}
	}
	return false
}
