package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var dnsCheck = checks.CheckDNS

func newDNSCommand() *cobra.Command {
	var recordType string
	var server string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "dns HOST",
		Short: "Query DNS records",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			recordType = strings.ToUpper(strings.TrimSpace(recordType))
			if !supportedDNSType(recordType) {
				return newCommandError(exitUsage, fmt.Errorf("unsupported DNS record type %q: use A, AAAA, CNAME, MX, NS, TXT, or PTR", recordType))
			}
			result := dnsCheck(cmd.Context(), nil, checks.DNSOptions{Host: args[0], Type: recordType, Server: server, Timeout: timeout})
			runErr := dnsResultError(result)
			return emitCommandResult(cmd, "dns", dnsOutput(result), runErr, func(w io.Writer) error {
				renderDNS(w, result)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&recordType, "type", "A", "record type: A, AAAA, CNAME, MX, NS, TXT, or PTR")
	cmd.Flags().StringVar(&server, "server", "", "DNS resolver address, with optional port")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "query timeout")
	return cmd
}

func supportedDNSType(recordType string) bool {
	switch recordType {
	case "A", "AAAA", "CNAME", "MX", "NS", "TXT", "PTR":
		return true
	default:
		return false
	}
}

func dnsResultError(result checks.DNSResult) error {
	if result.Err != nil {
		return newCommandError(exitUnreachable, result.Err)
	}
	if !result.Healthy {
		return newCommandError(exitUnhealthy, fmt.Errorf("DNS query returned no %s records for %s", result.Type, result.Query))
	}
	return nil
}

func renderDNS(w io.Writer, result checks.DNSResult) {
	fmt.Fprintf(w, "Query: %s\nType: %s\n", result.Query, result.Type)
	if result.Server != "" {
		fmt.Fprintf(w, "Server: %s\n", result.Server)
	}
	fmt.Fprintf(w, "Latency: %s\n", result.Latency.Round(time.Millisecond))
	if result.Err != nil {
		fmt.Fprintf(w, "Reachable: no\nError: %v\n", result.Err)
		return
	}
	fmt.Fprintln(w, "Records:")
	for _, record := range result.Records {
		fmt.Fprintf(w, "  %s\n", record)
	}
}

type dnsResultJSON struct {
	Query     string   `json:"query"`
	Type      string   `json:"type"`
	Server    string   `json:"server,omitempty"`
	Records   []string `json:"records"`
	Latency   string   `json:"latency"`
	Reachable bool     `json:"reachable"`
	Healthy   bool     `json:"healthy"`
	Error     string   `json:"error,omitempty"`
}

func dnsOutput(result checks.DNSResult) dnsResultJSON {
	output := dnsResultJSON{
		Query: result.Query, Type: result.Type, Server: result.Server, Records: result.Records,
		Latency: result.Latency.String(), Reachable: result.Reachable, Healthy: result.Healthy,
	}
	if result.Err != nil {
		output.Error = result.Err.Error()
	}
	return output
}
