package checks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type PingOptions struct {
	Host       string
	Count      int
	Timeout    time.Duration
	OnProgress func(completed, total int)
}

type PingAttempt struct {
	Seq     int
	Latency time.Duration
	Err     error
}

type PingSummary struct {
	Sent        int
	Received    int
	LossPercent float64
	Min         time.Duration
	Avg         time.Duration
	Max         time.Duration
}

type ErrICMPPermission struct {
	Cause error
}

func (e ErrICMPPermission) Error() string {
	if e.Cause == nil {
		return "ICMP permission error: ping may require elevated permissions or OS capability configuration"
	}
	return fmt.Sprintf("ICMP permission error: ping may require elevated permissions or OS capability configuration: %v", e.Cause)
}

func (e ErrICMPPermission) Unwrap() error { return e.Cause }

func IsICMPPermissionError(err error) bool {
	var permission ErrICMPPermission
	if errors.As(err, &permission) {
		return true
	}
	return looksLikePermissionError(err)
}

type pingSocket struct {
	conn     *icmp.PacketConn
	datagram bool
}

func listenPingSocket(listen func(network, address string) (*icmp.PacketConn, error)) (pingSocket, error) {
	conn, datagramErr := listen("udp4", "0.0.0.0")
	if datagramErr == nil {
		return pingSocket{conn: conn, datagram: true}, nil
	}

	conn, err := listen("ip4:icmp", "0.0.0.0")
	if err == nil {
		return pingSocket{conn: conn}, nil
	}
	if looksLikePermissionError(err) {
		return pingSocket{}, ErrICMPPermission{Cause: err}
	}
	return pingSocket{}, fmt.Errorf("listen ICMP (unprivileged: %v; raw: %w)", datagramErr, err)
}

func SummarizePing(attempts []PingAttempt) PingSummary {
	summary := PingSummary{Sent: len(attempts)}
	var total time.Duration
	for _, attempt := range attempts {
		if attempt.Err != nil {
			continue
		}
		summary.Received++
		if summary.Min == 0 || attempt.Latency < summary.Min {
			summary.Min = attempt.Latency
		}
		if attempt.Latency > summary.Max {
			summary.Max = attempt.Latency
		}
		total += attempt.Latency
	}
	if summary.Sent > 0 {
		summary.LossPercent = float64(summary.Sent-summary.Received) * 100 / float64(summary.Sent)
	}
	if summary.Received > 0 {
		summary.Avg = total / time.Duration(summary.Received)
	}
	return summary
}

func pingSummaryError(host string, summary PingSummary) error {
	if summary.Sent > 0 && summary.Received == 0 {
		return fmt.Errorf("no ping replies from %s", host)
	}
	return nil
}

func CheckPing(ctx context.Context, opts PingOptions) ([]PingAttempt, PingSummary, error) {
	if opts.Count <= 0 {
		opts.Count = 4
	}
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second
	}
	addr, err := net.ResolveIPAddr("ip4", opts.Host)
	if err != nil {
		return nil, PingSummary{}, fmt.Errorf("resolve %s: %w", opts.Host, err)
	}
	socket, err := listenPingSocket(icmp.ListenPacket)
	if err != nil {
		return nil, PingSummary{}, err
	}
	defer socket.conn.Close()

	var destination net.Addr = addr
	if socket.datagram {
		destination = &net.UDPAddr{IP: addr.IP, Zone: addr.Zone}
	}

	attempts := make([]PingAttempt, 0, opts.Count)
	id := os.Getpid() & 0xffff
	for seq := 1; seq <= opts.Count; seq++ {
		select {
		case <-ctx.Done():
			attempts = append(attempts, PingAttempt{Seq: seq, Err: ctx.Err()})
			if opts.OnProgress != nil {
				opts.OnProgress(len(attempts), opts.Count)
			}
			return attempts, SummarizePing(attempts), ctx.Err()
		default:
		}
		attempt, err := pingOnce(socket.conn, destination, id, seq, opts.Timeout, socket.datagram)
		attempts = append(attempts, attempt)
		if opts.OnProgress != nil {
			opts.OnProgress(len(attempts), opts.Count)
		}
		if err != nil && IsICMPPermissionError(err) {
			return attempts, SummarizePing(attempts), err
		}
	}
	summary := SummarizePing(attempts)
	return attempts, summary, pingSummaryError(opts.Host, summary)
}

func pingOnce(conn *icmp.PacketConn, addr net.Addr, id, seq int, timeout time.Duration, ignoreReplyID bool) (PingAttempt, error) {
	attempt := PingAttempt{Seq: seq}
	data := []byte("nwcli")
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{ID: id, Seq: seq, Data: data},
	}
	payload, err := message.Marshal(nil)
	if err != nil {
		attempt.Err = err
		return attempt, err
	}
	start := time.Now()
	if _, err := conn.WriteTo(payload, addr); err != nil {
		attempt.Err = err
		if looksLikePermissionError(err) {
			attempt.Err = ErrICMPPermission{Cause: err}
		}
		return attempt, attempt.Err
	}
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		attempt.Err = err
		return attempt, err
	}
	buf := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			attempt.Err = err
			if looksLikePermissionError(err) {
				attempt.Err = ErrICMPPermission{Cause: err}
			}
			return attempt, attempt.Err
		}
		reply, err := icmp.ParseMessage(1, buf[:n])
		if err != nil {
			continue
		}
		body, ok := reply.Body.(*icmp.Echo)
		if reply.Type == ipv4.ICMPTypeEchoReply && ok && body.Seq == seq && bytes.Equal(body.Data, data) && (ignoreReplyID || body.ID == id) {
			attempt.Latency = time.Since(start)
			return attempt, nil
		}
	}
}

func looksLikePermissionError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "operation not permitted") || strings.Contains(text, "permission denied") || strings.Contains(text, "raw socket")
}
