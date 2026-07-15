package checks

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type contextDialFunc func(context.Context, string, string) (net.Conn, error)

func (f contextDialFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

func TestCheckTCPReportsReachableConnection(t *testing.T) {
	called := ""
	dialer := contextDialFunc(func(_ context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" {
			t.Fatalf("network = %q", network)
		}
		called = address
		client, server := net.Pipe()
		_ = server.Close()
		return client, nil
	})
	result := CheckTCP(context.Background(), dialer, TCPOptions{Address: "example.com:443", Timeout: time.Second})
	if result.Err != nil {
		t.Fatalf("CheckTCP returned error: %v", result.Err)
	}
	if called != "example.com:443" || !result.Reachable {
		t.Fatalf("result = %#v, called = %q", result, called)
	}
}

func TestCheckTCPReportsDialFailure(t *testing.T) {
	wantErr := errors.New("connection refused")
	dialer := contextDialFunc(func(context.Context, string, string) (net.Conn, error) {
		return nil, wantErr
	})
	result := CheckTCP(context.Background(), dialer, TCPOptions{Address: "example.com:22"})
	if !errors.Is(result.Err, wantErr) || result.Reachable {
		t.Fatalf("result = %#v", result)
	}
}

func TestCheckTCPRejectsAddressWithoutPort(t *testing.T) {
	result := CheckTCP(context.Background(), nil, TCPOptions{Address: "example.com"})
	if result.Err == nil {
		t.Fatal("Err = nil")
	}
}
