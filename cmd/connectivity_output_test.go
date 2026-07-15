package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
)

func TestHTTPCommandWritesJSONEnvelope(t *testing.T) {
	original := httpCheck
	t.Cleanup(func() { httpCheck = original })
	httpCheck = func(_ context.Context, _ *http.Client, opts checks.HTTPOptions) checks.HTTPResult {
		return checks.HTTPResult{URL: "https://" + opts.Host + ":443/", StatusCode: 204, Latency: 9 * time.Millisecond, Reachable: true, Healthy: true}
	}
	stdout, _, err := executeForTest("--output", "json", "http", "example.com", "--tls")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope struct {
		Command string `json:"command"`
		Result  struct {
			StatusCode int `json:"status_code"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if envelope.Command != "http" || envelope.Result.StatusCode != 204 {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestPingCommandWritesJSONEnvelope(t *testing.T) {
	original := pingCheck
	t.Cleanup(func() { pingCheck = original })
	pingCheck = func(_ context.Context, opts checks.PingOptions) ([]checks.PingAttempt, checks.PingSummary, error) {
		return []checks.PingAttempt{{Seq: 1, Latency: time.Millisecond}}, checks.PingSummary{Sent: 1, Received: 1, Min: time.Millisecond, Avg: time.Millisecond, Max: time.Millisecond}, nil
	}
	stdout, _, err := executeForTest("--output", "json", "ping", "127.0.0.1", "--count", "1")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope struct {
		Command string `json:"command"`
		Result  struct {
			Summary struct {
				Received int `json:"received"`
			} `json:"summary"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if envelope.Command != "ping" || envelope.Result.Summary.Received != 1 {
		t.Fatalf("envelope = %#v", envelope)
	}
}
