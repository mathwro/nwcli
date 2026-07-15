package checks

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"
)

type stubDNSResolver struct {
	ips   []net.IPAddr
	mx    []*net.MX
	ptr   []string
	texts []string
}

func (s stubDNSResolver) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	return s.ips, nil
}
func (s stubDNSResolver) LookupCNAME(context.Context, string) (string, error) {
	return "alias.example.com.", nil
}
func (s stubDNSResolver) LookupMX(context.Context, string) ([]*net.MX, error) { return s.mx, nil }
func (s stubDNSResolver) LookupNS(context.Context, string) ([]*net.NS, error) {
	return []*net.NS{{Host: "ns1.example.com."}}, nil
}
func (s stubDNSResolver) LookupTXT(context.Context, string) ([]string, error)  { return s.texts, nil }
func (s stubDNSResolver) LookupAddr(context.Context, string) ([]string, error) { return s.ptr, nil }

func TestCheckDNSFiltersAndSortsAddressRecords(t *testing.T) {
	resolver := stubDNSResolver{ips: []net.IPAddr{
		{IP: net.ParseIP("2001:db8::1")},
		{IP: net.ParseIP("192.0.2.20")},
		{IP: net.ParseIP("192.0.2.10")},
	}}
	result := CheckDNS(context.Background(), resolver, DNSOptions{Host: "example.com", Type: "a", Timeout: time.Second})
	if result.Err != nil {
		t.Fatalf("CheckDNS returned error: %v", result.Err)
	}
	want := []string{"192.0.2.10", "192.0.2.20"}
	if !reflect.DeepEqual(result.Records, want) {
		t.Fatalf("Records = %#v, want %#v", result.Records, want)
	}
	if result.Type != "A" || !result.Reachable {
		t.Fatalf("result = %#v", result)
	}
}

func TestCheckDNSFormatsMXAndPTRRecords(t *testing.T) {
	resolver := stubDNSResolver{
		mx:  []*net.MX{{Host: "mail.example.com.", Pref: 10}},
		ptr: []string{"host.example.com."},
	}
	mx := CheckDNS(context.Background(), resolver, DNSOptions{Host: "example.com", Type: "MX"})
	if !reflect.DeepEqual(mx.Records, []string{"10 mail.example.com."}) {
		t.Fatalf("MX records = %#v", mx.Records)
	}
	ptr := CheckDNS(context.Background(), resolver, DNSOptions{Host: "192.0.2.10", Type: "PTR"})
	if !reflect.DeepEqual(ptr.Records, []string{"host.example.com."}) {
		t.Fatalf("PTR records = %#v", ptr.Records)
	}
}

func TestCheckDNSRejectsUnsupportedRecordType(t *testing.T) {
	result := CheckDNS(context.Background(), stubDNSResolver{}, DNSOptions{Host: "example.com", Type: "SRV"})
	if result.Err == nil {
		t.Fatal("Err = nil")
	}
}
