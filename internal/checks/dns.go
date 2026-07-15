package checks

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DNSOptions struct {
	Host    string
	Type    string
	Server  string
	Timeout time.Duration
}

type DNSResult struct {
	Query     string
	Type      string
	Server    string
	Records   []string
	Latency   time.Duration
	Reachable bool
	Healthy   bool
	Err       error
}

type DNSResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
	LookupCNAME(context.Context, string) (string, error)
	LookupMX(context.Context, string) ([]*net.MX, error)
	LookupNS(context.Context, string) ([]*net.NS, error)
	LookupTXT(context.Context, string) ([]string, error)
	LookupAddr(context.Context, string) ([]string, error)
}

func CheckDNS(ctx context.Context, resolver DNSResolver, opts DNSOptions) DNSResult {
	recordType := strings.ToUpper(strings.TrimSpace(opts.Type))
	if recordType == "" {
		recordType = "A"
	}
	result := DNSResult{Query: opts.Host, Type: recordType, Server: opts.Server}
	if opts.Host == "" {
		result.Err = fmt.Errorf("DNS query host is required")
		return result
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 2 * time.Second
	}
	if resolver == nil {
		resolver = dnsResolver(opts.Server)
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	start := time.Now()
	records, err := lookupDNSRecords(ctx, resolver, opts.Host, recordType)
	result.Latency = time.Since(start)
	if err != nil {
		result.Err = fmt.Errorf("lookup %s %s: %w", recordType, opts.Host, err)
		return result
	}
	sort.Strings(records)
	result.Records = records
	result.Reachable = true
	result.Healthy = len(records) > 0
	return result
}

func lookupDNSRecords(ctx context.Context, resolver DNSResolver, host, recordType string) ([]string, error) {
	switch recordType {
	case "A", "AAAA":
		addresses, err := resolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		records := make([]string, 0, len(addresses))
		for _, address := range addresses {
			isIPv4 := address.IP.To4() != nil
			if (recordType == "A" && isIPv4) || (recordType == "AAAA" && !isIPv4) {
				records = append(records, address.IP.String())
			}
		}
		return records, nil
	case "CNAME":
		name, err := resolver.LookupCNAME(ctx, host)
		if err != nil {
			return nil, err
		}
		return []string{name}, nil
	case "MX":
		records, err := resolver.LookupMX(ctx, host)
		if err != nil {
			return nil, err
		}
		values := make([]string, len(records))
		for i, record := range records {
			values[i] = strconv.Itoa(int(record.Pref)) + " " + record.Host
		}
		return values, nil
	case "NS":
		records, err := resolver.LookupNS(ctx, host)
		if err != nil {
			return nil, err
		}
		values := make([]string, len(records))
		for i, record := range records {
			values[i] = record.Host
		}
		return values, nil
	case "TXT":
		return resolver.LookupTXT(ctx, host)
	case "PTR":
		return resolver.LookupAddr(ctx, host)
	default:
		return nil, fmt.Errorf("unsupported DNS record type %q: use A, AAAA, CNAME, MX, NS, TXT, or PTR", recordType)
	}
}

func dnsResolver(server string) DNSResolver {
	if strings.TrimSpace(server) == "" {
		return net.DefaultResolver
	}
	address := strings.TrimSpace(server)
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(strings.Trim(address, "[]"), "53")
	}
	dialer := net.Dialer{}
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, address)
		},
	}
}
