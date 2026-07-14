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

func TestPingSummaryErrorWhenNoReplies(t *testing.T) {
	err := pingSummaryError("127.0.0.1", PingSummary{Sent: 2, Received: 0})
	if err == nil {
		t.Fatal("pingSummaryError returned nil, want no replies error")
	}
	if !strings.Contains(err.Error(), "no ping replies") {
		t.Fatalf("error = %q", err.Error())
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

func TestIsICMPPermissionErrorDetectsWrappedPermissionError(t *testing.T) {
	err := ErrICMPPermission{Cause: errors.New("permission denied")}
	if !IsICMPPermissionError(err) {
		t.Fatal("IsICMPPermissionError = false, want true")
	}
}
