package checks

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
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

const maxExpandedPorts = 1024

func ExpandTCPTarget(raw string) ([]string, error) {
	host, portSpec, err := net.SplitHostPort(strings.TrimSpace(raw))
	if err != nil || host == "" || portSpec == "" {
		return nil, fmt.Errorf("invalid TCP address %q: expected HOST:PORT or HOST:PORT1,PORT2,START-END", raw)
	}
	parts := strings.Split(portSpec, ",")
	var addresses []string
	seen := make(map[int]bool)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("invalid TCP port list %q", raw)
		}
		if loStr, hiStr, ok := strings.Cut(part, "-"); ok {
			lo, err1 := parsePortNumber(loStr)
			hi, err2 := parsePortNumber(hiStr)
			if err1 != nil || err2 != nil || lo > hi {
				return nil, fmt.Errorf("invalid TCP port range %q in %q", part, raw)
			}
			if hi-lo+1 > maxExpandedPorts || len(addresses)+(hi-lo+1) > maxExpandedPorts {
				return nil, fmt.Errorf("TCP port range %q exceeds maximum %d ports", raw, maxExpandedPorts)
			}
			for p := lo; p <= hi; p++ {
				if !seen[p] {
					seen[p] = true
					addresses = append(addresses, net.JoinHostPort(host, strconv.Itoa(p)))
				}
			}
			continue
		}
		p, err := parsePortNumber(part)
		if err != nil {
			return nil, fmt.Errorf("invalid TCP port %q in %q", part, raw)
		}
		if !seen[p] {
			seen[p] = true
			addresses = append(addresses, net.JoinHostPort(host, strconv.Itoa(p)))
		}
		if len(addresses) > maxExpandedPorts {
			return nil, fmt.Errorf("TCP port list %q exceeds maximum %d ports", raw, maxExpandedPorts)
		}
	}
	return addresses, nil
}

func parsePortNumber(raw string) (int, error) {
	p, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || p < 1 || p > 65535 {
		return 0, fmt.Errorf("port must be 1-65535")
	}
	return p, nil
}
