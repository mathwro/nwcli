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
	"golang.org/x/net/ipv6"
)

type PingOptions struct {
	Host     string
	Count    int
	Interval time.Duration
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
	ipv6     bool
}

func listenPingSocket(listen func(network, address string) (*icmp.PacketConn, error)) (pingSocket, error) {
	return listenPingSocketForFamily(false, listen)
}

func listenPingSocketForFamily(useIPv6 bool, listen func(network, address string) (*icmp.PacketConn, error)) (pingSocket, error) {
	udpNet, rawNet, bindAddr := "udp4", "ip4:icmp", "0.0.0.0"
	if useIPv6 {
		udpNet, rawNet, bindAddr = "udp6", "ip6:ipv6-icmp", "::"
	}
	conn, datagramErr := listen(udpNet, bindAddr)
	if datagramErr == nil {
		return pingSocket{conn: conn, datagram: true, ipv6: useIPv6}, nil
	}

	conn, err := listen(rawNet, bindAddr)
	if err == nil {
		return pingSocket{conn: conn, ipv6: useIPv6}, nil
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
	addr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return pingSession{}, fmt.Errorf("resolve %s: %w", host, err)
	}
	useIPv6 := addr.IP.To4() == nil
	socket, err := listenPingSocketForFamily(useIPv6, icmp.ListenPacket)
	if err != nil {
		return pingSession{}, err
	}
	var destination net.Addr = addr
	if socket.datagram {
		destination = &net.UDPAddr{IP: addr.IP, Zone: addr.Zone}
	}
	return pingSession{socket: socket, destination: destination, id: os.Getpid() & 0xffff}, nil
}

func (s pingSession) ping(ctx context.Context, seq int, timeout time.Duration) (PingAttempt, error) {
	return pingOnce(ctx, s.socket.conn, s.destination, s.id, seq, timeout, s.socket.datagram, s.socket.ipv6)
}

func CheckPing(ctx context.Context, opts PingOptions) ([]PingAttempt, PingSummary, error) {
	if opts.Count <= 0 {
		opts.Count = 4
	}
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second
	}
	interval := opts.Interval
	if interval <= 0 {
		interval = time.Second
	}
	session, err := newPingSession(opts.Host)
	if err != nil {
		return nil, PingSummary{}, err
	}
	defer session.socket.conn.Close()

	return runFinitePing(ctx, opts.Host, opts.Count, interval, func(seq int) (PingAttempt, error) {
		return session.ping(ctx, seq, opts.Timeout)
	}, opts.OnUpdate)
}

func runFinitePing(ctx context.Context, host string, count int, interval time.Duration, ping func(int) (PingAttempt, error), onUpdate func(PingUpdate)) ([]PingAttempt, PingSummary, error) {
	attempts := make([]PingAttempt, 0, count)
	var tracker pingTracker
	for seq := 1; seq <= count; seq++ {
		select {
		case <-ctx.Done():
			return attempts, tracker.summary, ctx.Err()
		default:
		}
		started := time.Now()
		attempt, attemptErr := ping(seq)
		attempts = append(attempts, attempt)
		summary := tracker.Add(attempt)
		if onUpdate != nil {
			onUpdate(PingUpdate{Attempt: attempt, Summary: summary})
		}
		if attemptErr != nil && IsICMPPermissionError(attemptErr) {
			return attempts, summary, attemptErr
		}
		if errors.Is(attemptErr, context.Canceled) || errors.Is(attemptErr, context.DeadlineExceeded) && ctx.Err() != nil {
			return attempts, summary, ctx.Err()
		}
		if seq == count {
			break
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
			return attempts, summary, ctx.Err()
		case <-timer.C:
		}
	}
	summary := tracker.summary
	return attempts, summary, pingSummaryError(host, summary)
}

func CheckPingContinuous(ctx context.Context, opts PingOptions) (PingSummary, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second
	}
	interval := opts.Interval
	if interval <= 0 {
		interval = time.Second
	}
	session, err := newPingSession(opts.Host)
	if err != nil {
		return PingSummary{}, err
	}
	defer session.socket.conn.Close()
	return runContinuousPing(ctx, interval, func(seq int) (PingAttempt, error) {
		return session.ping(ctx, seq, opts.Timeout)
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
		if ctx.Err() != nil {
			return summary, ctx.Err()
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

func pingOnce(ctx context.Context, conn *icmp.PacketConn, addr net.Addr, id, seq int, timeout time.Duration, ignoreReplyID, useIPv6 bool) (PingAttempt, error) {
	attempt := PingAttempt{Seq: seq}
	data := []byte("nwcli")
	var echoType icmp.Type = ipv4.ICMPTypeEcho
	var replyType icmp.Type = ipv4.ICMPTypeEchoReply
	proto := 1
	if useIPv6 {
		echoType = ipv6.ICMPTypeEchoRequest
		replyType = ipv6.ICMPTypeEchoReply
		proto = 58
	}
	message := icmp.Message{
		Type: echoType,
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
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			attempt.Err = fmt.Errorf("request timed out")
			return attempt, attempt.Err
		}
		packet, err := readWithContextDeadline(ctx, conn, remaining)
		if err != nil {
			attempt.Err = err
			if looksLikePermissionError(err) {
				attempt.Err = ErrICMPPermission{Cause: err}
			}
			return attempt, attempt.Err
		}
		reply, err := icmp.ParseMessage(proto, packet)
		if err != nil {
			continue
		}
		body, ok := reply.Body.(*icmp.Echo)
		if reply.Type == replyType && ok && body.Seq == seq && bytes.Equal(body.Data, data) && (ignoreReplyID || body.ID == id) {
			attempt.Latency = time.Since(start)
			return attempt, nil
		}
	}
}

func readWithContextDeadline(ctx context.Context, conn net.PacketConn, timeout time.Duration) ([]byte, error) {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.SetReadDeadline(time.Now())
		case <-done:
		}
	}()
	defer close(done)

	buf := make([]byte, 1500)
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	return buf[:n], nil
}

func looksLikePermissionError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "operation not permitted") || strings.Contains(text, "permission denied") || strings.Contains(text, "raw socket")
}
