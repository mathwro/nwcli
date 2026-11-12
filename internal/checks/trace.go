package checks

import (
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

type TraceOptions struct {
	Host     string
	MaxHops  int
	Timeout  time.Duration
	OnUpdate func(TraceHop)
}

type TraceHop struct {
	TTL     int
	Address string
	Latency time.Duration
	Timeout bool
	Reached bool
	Err     error
}

type TraceResult struct {
	Host    string
	Target  string
	Hops    []TraceHop
	Reached bool
}

func CheckTrace(ctx context.Context, opts TraceOptions) (TraceResult, error) {
	if strings.TrimSpace(opts.Host) == "" {
		return TraceResult{}, fmt.Errorf("trace host is required")
	}
	if opts.MaxHops <= 0 {
		opts.MaxHops = 20
	}
	if opts.Timeout <= 0 {
		opts.Timeout = time.Second
	}
	addr, err := net.ResolveIPAddr("ip", opts.Host)
	if err != nil {
		return TraceResult{Host: opts.Host}, fmt.Errorf("resolve %s: %w", opts.Host, err)
	}
	useIPv6 := addr.IP.To4() == nil
	socket, err := listenPingSocketForFamily(useIPv6, icmp.ListenPacket)
	if err != nil {
		return TraceResult{Host: opts.Host, Target: addr.IP.String()}, err
	}
	defer socket.conn.Close()

	var destination net.Addr = addr
	if socket.datagram {
		destination = &net.UDPAddr{IP: addr.IP, Zone: addr.Zone}
	}
	id := os.Getpid() & 0xffff

	res, err := runTrace(ctx, addr.IP.String(), opts.MaxHops, func(probeCtx context.Context, ttl int) (TraceHop, error) {
		return traceOnce(probeCtx, socket, destination, addr.IP, id, ttl, opts.Timeout)
	}, opts.OnUpdate)
	res.Host = opts.Host
	return res, err
}

func runTrace(ctx context.Context, target string, maxHops int, probe func(context.Context, int) (TraceHop, error), onUpdate func(TraceHop)) (TraceResult, error) {
	result := TraceResult{Target: target, Hops: make([]TraceHop, 0, maxHops)}
	for ttl := 1; ttl <= maxHops; ttl++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		hop, err := probe(ctx, ttl)
		if err != nil {
			if IsICMPPermissionError(err) {
				return result, err
			}
			if errors.Is(err, context.Canceled) || (errors.Is(err, context.DeadlineExceeded) && ctx.Err() != nil) {
				return result, ctx.Err()
			}
			hop.Err = err
		}
		result.Hops = append(result.Hops, hop)
		if onUpdate != nil {
			onUpdate(hop)
		}
		if hop.Reached {
			result.Reached = true
			return result, nil
		}
	}
	return result, fmt.Errorf("destination %s not reached within %d hops", target, maxHops)
}

func traceOnce(ctx context.Context, socket pingSocket, dst net.Addr, targetIP net.IP, id, ttl int, timeout time.Duration) (TraceHop, error) {
	hop := TraceHop{TTL: ttl}
	var setErr error
	var echoType icmp.Type = ipv4.ICMPTypeEcho
	proto := 1
	if socket.ipv6 {
		echoType = ipv6.ICMPTypeEchoRequest
		proto = 58
		if p6 := socket.conn.IPv6PacketConn(); p6 != nil {
			setErr = p6.SetHopLimit(ttl)
		}
	} else if p4 := socket.conn.IPv4PacketConn(); p4 != nil {
		setErr = p4.SetTTL(ttl)
	}
	if setErr != nil {
		if looksLikePermissionError(setErr) {
			return hop, ErrICMPPermission{Cause: setErr}
		}
		return hop, fmt.Errorf("set TTL %d: %w", ttl, setErr)
	}

	msg := icmp.Message{
		Type: echoType,
		Code: 0,
		Body: &icmp.Echo{ID: id, Seq: ttl, Data: []byte("nwcli-trace")},
	}
	payload, err := msg.Marshal(nil)
	if err != nil {
		return hop, err
	}
	start := time.Now()
	if _, err := socket.conn.WriteTo(payload, dst); err != nil {
		if looksLikePermissionError(err) {
			return hop, ErrICMPPermission{Cause: err}
		}
		return hop, err
	}

	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			hop.Timeout = true
			return hop, nil
		}
		if err := socket.conn.SetReadDeadline(time.Now().Add(remaining)); err != nil {
			return hop, err
		}
		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				_ = socket.conn.SetReadDeadline(time.Now())
			case <-done:
			}
		}()
		buf := make([]byte, 1500)
		n, peer, readErr := socket.conn.ReadFrom(buf)
		close(done)
		if readErr != nil {
			if ctx.Err() != nil {
				return hop, ctx.Err()
			}
			if looksLikePermissionError(readErr) {
				return hop, ErrICMPPermission{Cause: readErr}
			}
			var netErr net.Error
			if errors.As(readErr, &netErr) && netErr.Timeout() {
				hop.Timeout = true
				return hop, nil
			}
			return hop, readErr
		}
		reply, parseErr := icmp.ParseMessage(proto, buf[:n])
		if parseErr != nil {
			continue
		}
		peerHost := extractPeerIP(peer)
		switch {
		case (!socket.ipv6 && reply.Type == ipv4.ICMPTypeEchoReply) || (socket.ipv6 && reply.Type == ipv6.ICMPTypeEchoReply):
			if body, ok := reply.Body.(*icmp.Echo); ok && body.Seq == ttl && (socket.datagram || body.ID == id) {
				hop.Address = peerHost
				hop.Latency = time.Since(start)
				hop.Reached = true
				return hop, nil
			}
		case (!socket.ipv6 && (reply.Type == ipv4.ICMPTypeTimeExceeded || reply.Type == ipv4.ICMPTypeDestinationUnreachable)) ||
			(socket.ipv6 && (reply.Type == ipv6.ICMPTypeTimeExceeded || reply.Type == ipv6.ICMPTypeDestinationUnreachable)):
			hop.Address = peerHost
			hop.Latency = time.Since(start)
			if net.ParseIP(peerHost) != nil && net.ParseIP(peerHost).Equal(targetIP) {
				hop.Reached = true
			}
			return hop, nil
		}
	}
}

func extractPeerIP(peer net.Addr) string {
	if peer == nil {
		return ""
	}
	switch a := peer.(type) {
	case *net.IPAddr:
		return a.IP.String()
	case *net.UDPAddr:
		return a.IP.String()
	default:
		if host, _, err := net.SplitHostPort(peer.String()); err == nil {
			return host
		}
		return peer.String()
	}
}
