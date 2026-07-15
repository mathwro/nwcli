package checks

import (
	"context"
	"fmt"
	"net"
	"time"
)

type TCPOptions struct {
	Address string
	Timeout time.Duration
}

type TCPResult struct {
	Address       string
	LocalAddress  string
	RemoteAddress string
	Latency       time.Duration
	Reachable     bool
	Err           error
}

type ContextDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

func CheckTCP(ctx context.Context, dialer ContextDialer, opts TCPOptions) TCPResult {
	result := TCPResult{Address: opts.Address}
	host, port, err := net.SplitHostPort(opts.Address)
	if err != nil || host == "" || port == "" {
		result.Err = fmt.Errorf("invalid TCP address %q: expected HOST:PORT", opts.Address)
		return result
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 2 * time.Second
	}
	if dialer == nil {
		dialer = &net.Dialer{Timeout: opts.Timeout}
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", opts.Address)
	result.Latency = time.Since(start)
	if err != nil {
		result.Err = fmt.Errorf("connect TCP %s: %w", opts.Address, err)
		return result
	}
	defer conn.Close()
	result.Reachable = true
	result.LocalAddress = conn.LocalAddr().String()
	result.RemoteAddress = conn.RemoteAddr().String()
	return result
}
