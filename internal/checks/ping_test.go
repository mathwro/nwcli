package checks

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/icmp"
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

func TestListenPingSocketPrefersUnprivilegedICMP(t *testing.T) {
	want := new(icmp.PacketConn)
	var networks []string
	socket, err := listenPingSocket(func(network, _ string) (*icmp.PacketConn, error) {
		networks = append(networks, network)
		return want, nil
	})
	if err != nil {
		t.Fatalf("listenPingSocket returned error: %v", err)
	}
	if socket.conn != want || !socket.datagram {
		t.Fatalf("socket = %+v, want unprivileged datagram socket", socket)
	}
	if len(networks) != 1 || networks[0] != "udp4" {
		t.Fatalf("listen networks = %v, want [udp4]", networks)
	}
}

func TestListenPingSocketFallsBackToRawICMP(t *testing.T) {
	want := new(icmp.PacketConn)
	var networks []string
	socket, err := listenPingSocket(func(network, _ string) (*icmp.PacketConn, error) {
		networks = append(networks, network)
		if network == "udp4" {
			return nil, errors.New("datagram ICMP unavailable")
		}
		return want, nil
	})
	if err != nil {
		t.Fatalf("listenPingSocket returned error: %v", err)
	}
	if socket.conn != want || socket.datagram {
		t.Fatalf("socket = %+v, want raw socket", socket)
	}
	if len(networks) != 2 || networks[0] != "udp4" || networks[1] != "ip4:icmp" {
		t.Fatalf("listen networks = %v, want [udp4 ip4:icmp]", networks)
	}
}

func TestListenPingSocketReportsRawPermissionError(t *testing.T) {
	_, err := listenPingSocket(func(network, _ string) (*icmp.PacketConn, error) {
		if network == "udp4" {
			return nil, errors.New("datagram ICMP unavailable")
		}
		return nil, errors.New("operation not permitted")
	})
	if !IsICMPPermissionError(err) {
		t.Fatalf("listenPingSocket error = %v, want ICMP permission error", err)
	}
}

func TestPingTrackerUpdatesLiveDropCountAndPercentage(t *testing.T) {
	var tracker pingTracker
	first := tracker.Add(PingAttempt{Seq: 1, Latency: 10 * time.Millisecond})
	dropped := tracker.Add(PingAttempt{Seq: 2, Err: errors.New("timeout")})
	third := tracker.Add(PingAttempt{Seq: 3, Latency: 20 * time.Millisecond})

	if first.Sent != 1 || first.Received != 1 || first.LossPercent != 0 {
		t.Fatalf("first summary = %+v", first)
	}
	if dropped.Sent != 2 || dropped.Received != 1 || dropped.LossPercent != 50 {
		t.Fatalf("dropped summary = %+v, want one of two dropped", dropped)
	}
	if third.Sent != 3 || third.Received != 2 || third.LossPercent < 33.3 || third.LossPercent > 33.4 {
		t.Fatalf("third summary = %+v, want one of three dropped", third)
	}
	if third.Min != 10*time.Millisecond || third.Avg != 15*time.Millisecond || third.Max != 20*time.Millisecond {
		t.Fatalf("third latency summary = %+v", third)
	}
}

func TestRunContinuousPingStreamsEveryAttemptUntilCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var updates []PingUpdate
	summary, err := runContinuousPing(ctx, 0, func(seq int) (PingAttempt, error) {
		if seq == 2 {
			err := errors.New("timeout")
			return PingAttempt{Seq: seq, Err: err}, err
		}
		return PingAttempt{Seq: seq, Latency: time.Duration(seq) * time.Millisecond}, nil
	}, func(update PingUpdate) {
		updates = append(updates, update)
		if update.Summary.Sent == 3 {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runContinuousPing error = %v, want context canceled", err)
	}
	if len(updates) != 3 {
		t.Fatalf("len(updates) = %d, want 3", len(updates))
	}
	if updates[1].Attempt.Err == nil || updates[1].Summary.Sent-updates[1].Summary.Received != 1 || updates[1].Summary.LossPercent != 50 {
		t.Fatalf("drop update = %+v", updates[1])
	}
	if summary.Sent != 3 || summary.Received != 2 {
		t.Fatalf("final summary = %+v", summary)
	}
}
