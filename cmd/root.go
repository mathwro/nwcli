package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/mathwro/nwcli/internal/subnet"
	"github.com/mathwro/nwcli/internal/tui"
	"github.com/spf13/cobra"
)

func NewRootCommand(menuRunner func(context.Context) error) *cobra.Command {
	if menuRunner == nil {
		menuRunner = runInteractiveMenu
	}
	var outputFormat string
	root := &cobra.Command{
		Use:   "nwcli",
		Short: "Network toolbox for subnet planning and connectivity checks",
		Long:  "nwcli is a network toolbox for subnet planning, DNS diagnostics, and connectivity checks.",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return validateOutputFormat(outputFormat)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return menuRunner(cmd.Context())
		},
	}
	root.PersistentFlags().StringVar(&outputFormat, "output", "text", "output format: text or json")
	root.AddCommand(newSubnetCommand(), newPingCommand(), newHTTPCommand(), newDNSCommand(), newTCPCommand(), newTLSCommand(), newBatchCommand())
	return root
}

func Execute() {
	if err := NewRootCommand(runInteractiveMenu).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitCode(err))
	}
}

func runInteractiveMenu(ctx context.Context) error {
	return tui.RunApp(ctx, interactiveActions())
}

func interactiveActions() []tui.Action {
	return []tui.Action{
		{
			Title:       "Subnet carve",
			Description: "Allocate ordered IPv4 child networks and see remaining space",
			Fields: []tui.Field{
				{Name: "cidr", Label: "Parent CIDR", Default: "10.0.0.0/24", Validate: validateCIDR},
				{Name: "sizes", Label: "Child prefixes", Default: "27,27,28,28", Validate: validatePrefixes},
			},
			Run: runInteractiveSubnet,
		},
		{
			Title:       "CIDR inspect",
			Description: "Show normalized range, family, prefix, and address count",
			Fields: []tui.Field{
				{Name: "cidr", Label: "CIDR", Default: "10.0.0.0/24", Validate: validateCIDR},
			},
			Run: runInteractiveCIDR,
		},
		{
			Title:       "Ping test",
			Description: "Measure ICMP reachability, loss, and latency",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Default: "127.0.0.1", Validate: validateRequired},
				{Name: "count", Label: "Count", Default: "4", Validate: validatePositiveInteger},
				{Name: "timeout", Label: "Timeout", Default: "1s", Validate: validateDuration},
			},
			RunProgress: runInteractivePing,
		},
		{
			Title:       "DNS lookup",
			Description: "Query A, AAAA, CNAME, MX, NS, TXT, or PTR records",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Default: "example.com", Validate: validateRequired},
				{Name: "type", Label: "Record type", Default: "A", Validate: validateDNSRecordType},
				{Name: "server", Label: "Resolver", Default: ""},
				{Name: "timeout", Label: "Timeout", Default: "2s", Validate: validateDuration},
			},
			RunProgress: runInteractiveDNS,
		},
		{
			Title:       "TCP connect",
			Description: "Measure transport-level reachability to a host and port",
			Fields: []tui.Field{
				{Name: "address", Label: "Address", Default: "example.com:443", Validate: validateHostPort},
				{Name: "timeout", Label: "Timeout", Default: "2s", Validate: validateDuration},
			},
			RunProgress: runInteractiveTCP,
		},
		{
			Title:       "TLS inspect",
			Description: "Verify TLS and inspect certificate validity and negotiation",
			Fields: []tui.Field{
				{Name: "address", Label: "Address", Default: "example.com:443", Validate: validateHostPort},
				{Name: "server_name", Label: "Server name", Default: ""},
				{Name: "timeout", Label: "Timeout", Default: "3s", Validate: validateDuration},
				{Name: "warn_before", Label: "Expiry warning", Default: "30d", Validate: validateExtendedDuration},
			},
			RunProgress: runInteractiveTLS,
		},
		{
			Title:       "HTTP/S test",
			Description: "Check endpoint status, reachability, health, and latency",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Default: "example.com", Validate: validateRequired},
				{Name: "port", Label: "Port", Default: "443", Validate: validatePort},
				{Name: "tls", Label: "TLS true/false", Default: "true", Validate: validateBoolean},
				{Name: "path", Label: "Path", Default: "/"},
				{Name: "timeout", Label: "Timeout", Default: "2s", Validate: validateDuration},
			},
			RunProgress: runInteractiveHTTP,
		},
		{Title: "Quit", Description: "Exit nwcli"},
	}
}

func runInteractiveSubnet(_ context.Context, values map[string]string) (string, error) {
	prefixes, err := subnet.ParsePrefixes(values["sizes"])
	if err != nil {
		return "", err
	}
	plan, err := subnet.Plan(values["cidr"], prefixes)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := subnet.Render(&out, plan); err != nil {
		return "", err
	}
	return out.String(), nil
}

func runInteractiveCIDR(_ context.Context, values map[string]string) (string, error) {
	info, err := subnet.InspectCIDR(values["cidr"])
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "CIDR: %s\nFamily: %s\nPrefix: /%d\nFirst: %s\nLast: %s\nAddresses: %s\n", info.CIDR, info.Family, info.Prefix, info.First, info.Last, info.Addresses)
	return out.String(), nil
}

func runInteractivePing(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (string, error) {
	count, err := strconv.Atoi(values["count"])
	if err != nil {
		return "", fmt.Errorf("invalid count %q: %w", values["count"], err)
	}
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return "", fmt.Errorf("invalid timeout %q: %w", values["timeout"], err)
	}
	options := checks.PingOptions{
		Host:    values["host"],
		Count:   count,
		Timeout: timeout,
		OnProgress: func(current, total int) {
			if report != nil {
				report(tui.ProgressEvent{
					Label:   fmt.Sprintf("Ping %d of %d", current, total),
					Current: current,
					Total:   total,
				})
			}
		},
	}
	attempts, summary, err := pingCheck(ctx, options)
	var out bytes.Buffer
	renderPing(&out, attempts, summary)
	return out.String(), pingResultError(err)
}

func runInteractiveHTTP(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (string, error) {
	port, err := strconv.Atoi(values["port"])
	if err != nil {
		return "", fmt.Errorf("invalid port %q: %w", values["port"], err)
	}
	useTLS, err := strconv.ParseBool(values["tls"])
	if err != nil {
		return "", fmt.Errorf("invalid tls value %q: %w", values["tls"], err)
	}
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return "", fmt.Errorf("invalid timeout %q: %w", values["timeout"], err)
	}
	reportSingle(report, "Requesting endpoint", false)
	result := httpCheck(ctx, nil, checks.HTTPOptions{Host: values["host"], Port: port, TLS: useTLS, Path: values["path"], Timeout: timeout})
	reportSingle(report, "HTTP check complete", true)
	var out bytes.Buffer
	renderHTTP(&out, result)
	return out.String(), httpResultError(result)
}

func runInteractiveDNS(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (string, error) {
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return "", err
	}
	reportSingle(report, "Querying DNS", false)
	result := dnsCheck(ctx, nil, checks.DNSOptions{Host: values["host"], Type: strings.ToUpper(values["type"]), Server: values["server"], Timeout: timeout})
	reportSingle(report, "DNS query complete", true)
	var out bytes.Buffer
	renderDNS(&out, result)
	return out.String(), dnsResultError(result)
}

func runInteractiveTCP(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (string, error) {
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return "", err
	}
	reportSingle(report, "Opening TCP connection", false)
	result := tcpCheck(ctx, nil, checks.TCPOptions{Address: values["address"], Timeout: timeout})
	reportSingle(report, "TCP check complete", true)
	var out bytes.Buffer
	renderTCP(&out, result)
	return out.String(), tcpResultError(result)
}

func runInteractiveTLS(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (string, error) {
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return "", err
	}
	warnBefore, err := parseExtendedDuration(values["warn_before"])
	if err != nil {
		return "", err
	}
	reportSingle(report, "Negotiating TLS", false)
	result := tlsCheck(ctx, checks.TLSOptions{Address: values["address"], ServerName: values["server_name"], Timeout: timeout, WarnBefore: warnBefore})
	reportSingle(report, "TLS inspection complete", true)
	var out bytes.Buffer
	renderTLS(&out, result)
	return out.String(), tlsResultError(result)
}

func reportSingle(report func(tui.ProgressEvent), label string, complete bool) {
	if report == nil {
		return
	}
	current := 0
	if complete {
		current = 1
	}
	report(tui.ProgressEvent{Label: label, Current: current, Total: 1})
}

func validateRequired(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func validateCIDR(value string) error {
	_, err := subnet.InspectCIDR(value)
	return err
}

func validatePrefixes(value string) error {
	_, err := subnet.ParsePrefixes(value)
	return err
}

func validatePositiveInteger(value string) error {
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

func validateDuration(value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fmt.Errorf("must be a positive duration")
	}
	return nil
}

func validateExtendedDuration(value string) error {
	_, err := parseExtendedDuration(value)
	return err
}

func validateDNSRecordType(value string) error {
	if !supportedDNSType(strings.ToUpper(strings.TrimSpace(value))) {
		return fmt.Errorf("use A, AAAA, CNAME, MX, NS, TXT, or PTR")
	}
	return nil
}

func validateHostPort(value string) error {
	host, port, err := net.SplitHostPort(value)
	if err != nil || host == "" || port == "" {
		return fmt.Errorf("expected HOST:PORT")
	}
	return nil
}

func validatePort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("must be between 1 and 65535")
	}
	return nil
}

func validateBoolean(value string) error {
	_, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("use true or false")
	}
	return nil
}
