package checks

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestRunTraceStopsAtDestinationAndStreamsHops(t *testing.T) {
	var streamed []TraceHop
	result, err := runTrace(context.Background(), "198.51.100.10", 10, func(_ context.Context, ttl int) (TraceHop, error) {
		switch ttl {
		case 1:
			return TraceHop{TTL: 1, Address: "192.0.2.1", Latency: 2 * time.Millisecond}, nil
		case 2:
			return TraceHop{TTL: 2, Timeout: true}, nil
		case 3:
			return TraceHop{TTL: 3, Address: "198.51.100.10", Latency: 11 * time.Millisecond, Reached: true}, nil
		default:
			t.Fatalf("unexpected probe at ttl %d", ttl)
			return TraceHop{}, nil
		}
	}, func(hop TraceHop) {
		streamed = append(streamed, hop)
	})
	if err != nil {
		t.Fatalf("runTrace returned error: %v", err)
	}
	if !result.Reached || len(result.Hops) != 3 || len(streamed) != 3 {
		t.Fatalf("result = %+v, streamed = %d", result, len(streamed))
	}
	if !result.Hops[1].Timeout || !result.Hops[2].Reached {
		t.Fatalf("hops = %+v", result.Hops)
	}
}

func TestRunTracePropagatesPermissionError(t *testing.T) {
	_, err := runTrace(context.Background(), "198.51.100.10", 5, func(context.Context, int) (TraceHop, error) {
		return TraceHop{TTL: 1}, ErrICMPPermission{Cause: errors.New("operation not permitted")}
	}, nil)
	if !IsICMPPermissionError(err) {
		t.Fatalf("err = %v, want ICMP permission error", err)
	}
}

func TestCollectInterfacesFormatsFlagsAndCIDRs(t *testing.T) {
	mac, _ := net.ParseMAC("02:42:ac:11:00:02")
	items, err := collectInterfaces(false, func() ([]net.Interface, error) {
		return []net.Interface{
			{Index: 1, Name: "lo", MTU: 65536, Flags: net.FlagUp | net.FlagLoopback},
			{Index: 2, Name: "eth0", MTU: 1500, HardwareAddr: mac, Flags: net.FlagUp | net.FlagBroadcast},
			{Index: 3, Name: "eth1", MTU: 1500, Flags: 0},
		}, nil
	}, func(iface net.Interface) ([]net.Addr, error) {
		if iface.Name == "eth0" {
			_, n1, _ := net.ParseCIDR("10.0.0.15/24")
			n1.IP = net.ParseIP("10.0.0.15")
			return []net.Addr{n1}, nil
		}
		if iface.Name == "lo" {
			_, n1, _ := net.ParseCIDR("127.0.0.1/8")
			n1.IP = net.ParseIP("127.0.0.1")
			return []net.Addr{n1}, nil
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("collectInterfaces error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2 active interfaces (got %+v)", len(items), items)
	}
	if items[1].Name != "eth0" || items[1].HardwareAddr != "02:42:ac:11:00:02" || !reflect.DeepEqual(items[1].Addresses, []string{"10.0.0.15/24"}) {
		t.Fatalf("eth0 = %+v", items[1])
	}
}
