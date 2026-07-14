package checks

import (
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
