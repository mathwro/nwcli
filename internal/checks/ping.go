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
	Host     string
	Count    int
	Timeout  time.Duration
	OnUpdate func(PingUpdate)
}

type PingAttempt struct {
	Seq     int
	Latency time.Duration
	Err     error
}

type PingUpdate struct {
	Attempt PingAttempt
	Summary PingSummary
}

type PingSummary struct {
	Sent        int
	Received    int
	LossPercent float64
	Min         time.Duration
	Avg         time.Duration
	Max         time.Duration
}

type pingTracker struct {
	summary      PingSummary
	totalLatency time.Duration
}

func (t *pingTracker) Add(attempt PingAttempt) PingSummary {
	t.summary.Sent++
	if attempt.Err == nil {
		t.summary.Received++
		if t.summary.Received == 1 || attempt.Latency < t.summary.Min {
			t.summary.Min = attempt.Latency
		}
		if attempt.Latency > t.summary.Max {
			t.summary.Max = attempt.Latency
		}
		t.totalLatency += attempt.Latency
		t.summary.Avg = t.totalLatency / time.Duration(t.summary.Received)
	}
	t.summary.LossPercent = float64(t.summary.Sent-t.summary.Received) * 100 / float64(t.summary.Sent)
	return t.summary
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
	var tracker pingTracker
	for _, attempt := range attempts {
		tracker.Add(attempt)
	}
	return tracker.summary
}

func pingSummaryError(host string, summary PingSummary) error {
	if summary.Sent > 0 && summary.Received == 0 {
		return fmt.Errorf("no ping replies from %s", host)
	}
	return nil
}

type pingSession struct {
	socket      pingSocket
	destination net.Addr
	id          int
}

func newPingSession(host string) (pingSession, error) {
	addr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return pingSession{}, fmt.Errorf("resolve %s: %w", host, err)
	}
	socket, err := listenPingSocket(icmp.ListenPacket)
	if err != nil {
		return pingSession{}, err
	}
	var destination net.Addr = addr
	if socket.datagram {
		destination = &net.UDPAddr{IP: addr.IP, Zone: addr.Zone}
	}
	return pingSession{socket: socket, destination: destination, id: os.Getpid() & 0xffff}, nil
}

func (s pingSession) ping(seq int, timeout time.Duration) (PingAttempt, error) {
	return pingOnce(s.socket.conn, s.destination, s.id, seq, timeout, s.socket.datagram)
}

func CheckPing(ctx context.Context, opts PingOptions) ([]PingAttempt, PingSummary, error) {
	if opts.Count <= 0 {
		opts.Count = 4
	}
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second
	}
	session, err := newPingSession(opts.Host)
	if err != nil {
		return nil, PingSummary{}, err
	}
	defer session.socket.conn.Close()

	attempts := make([]PingAttempt, 0, opts.Count)
	var tracker pingTracker
	for seq := 1; seq <= opts.Count; seq++ {
		select {
		case <-ctx.Done():
			attempt := PingAttempt{Seq: seq, Err: ctx.Err()}
			attempts = append(attempts, attempt)
			summary := tracker.Add(attempt)
			if opts.OnUpdate != nil {
				opts.OnUpdate(PingUpdate{Attempt: attempt, Summary: summary})
			}
			return attempts, summary, ctx.Err()
		default:
		}
		attempt, attemptErr := session.ping(seq, opts.Timeout)
		attempts = append(attempts, attempt)
		summary := tracker.Add(attempt)
		if opts.OnUpdate != nil {
			opts.OnUpdate(PingUpdate{Attempt: attempt, Summary: summary})
		}
		if attemptErr != nil && IsICMPPermissionError(attemptErr) {
			return attempts, summary, attemptErr
		}
	}
	summary := tracker.summary
	return attempts, summary, pingSummaryError(opts.Host, summary)
}

func CheckPingContinuous(ctx context.Context, opts PingOptions) (PingSummary, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second
	}
	session, err := newPingSession(opts.Host)
	if err != nil {
		return PingSummary{}, err
	}
	defer session.socket.conn.Close()
	return runContinuousPing(ctx, time.Second, func(seq int) (PingAttempt, error) {
		return session.ping(seq, opts.Timeout)
	}, opts.OnUpdate)
}

func runContinuousPing(ctx context.Context, interval time.Duration, ping func(int) (PingAttempt, error), onUpdate func(PingUpdate)) (PingSummary, error) {
	var tracker pingTracker
	for seq := 1; ; seq++ {
		select {
		case <-ctx.Done():
			return tracker.summary, ctx.Err()
		default:
		}
		started := time.Now()
		attempt, err := ping(seq)
		summary := tracker.Add(attempt)
		if onUpdate != nil {
			onUpdate(PingUpdate{Attempt: attempt, Summary: summary})
		}
		if err != nil && IsICMPPermissionError(err) {
			return summary, err
		}
		wait := interval - time.Since(started)
		if wait <= 0 {
			continue
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return summary, ctx.Err()
		case <-timer.C:
		}
	}
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
