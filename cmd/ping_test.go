package cmd

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
)

func TestPingContinuousStreamsLivePacketLoss(t *testing.T) {
	original := continuousPingCheck
	t.Cleanup(func() { continuousPingCheck = original })
	continuousPingCheck = func(_ context.Context, opts checks.PingOptions) (checks.PingSummary, error) {
		first := checks.PingSummary{Sent: 1, Received: 1, Min: 8 * time.Millisecond, Avg: 8 * time.Millisecond, Max: 8 * time.Millisecond}
		opts.OnUpdate(checks.PingUpdate{Attempt: checks.PingAttempt{Seq: 1, Latency: 8 * time.Millisecond}, Summary: first})
		second := checks.PingSummary{Sent: 2, Received: 1, LossPercent: 50, Min: 8 * time.Millisecond, Avg: 8 * time.Millisecond, Max: 8 * time.Millisecond}
		opts.OnUpdate(checks.PingUpdate{Attempt: checks.PingAttempt{Seq: 2, Err: context.DeadlineExceeded}, Summary: second})
		return second, context.Canceled
	}

	stdout, _, err := executeForTest("ping", "127.0.0.1", "--continuous")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, fragment := range []string{"seq=1", "latency=8ms", "seq=2 DROP", "sent=2", "received=1", "dropped=1", "loss=50.0%"} {
		if !strings.Contains(stdout, fragment) {
			t.Fatalf("continuous output missing %q:\n%s", fragment, stdout)
		}
	}
	if !strings.Contains(stdout, "summary: sent=2 received=1 loss=50.0%") {
		t.Fatalf("continuous output missing final cancellation summary:\n%s", stdout)
	}
}

func TestPingContinuousRejectsJSONOutput(t *testing.T) {
	_, _, err := executeForTest("--output", "json", "ping", "127.0.0.1", "--continuous")
	if err == nil {
		t.Fatal("Execute returned nil error")
	}
	if !strings.Contains(err.Error(), "continuous ping supports text output only") {
		t.Fatalf("error = %q", err.Error())
	}
}
