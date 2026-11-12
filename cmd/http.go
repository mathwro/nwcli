package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var httpCheck = checks.CheckHTTP

func newHTTPCommand() *cobra.Command {
	var port int
	var useTLS bool
	var path string
	var followRedirects bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "http HOST",
		Short: "Check an HTTP/S endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result := httpCheck(cmd.Context(), nil, checks.HTTPOptions{Host: args[0], Port: port, TLS: useTLS, Path: path, FollowRedirects: followRedirects, Timeout: timeout})
			runErr := httpResultError(result)
			return emitCommandResult(cmd, "http", httpOutput(result), runErr, func(w io.Writer) error {
				renderHTTP(w, result)
				return nil
			})
		},
	}
	cmd.Flags().IntVar(&port, "port", 0, "target port (defaults to 80 or 443 with --tls)")
	cmd.Flags().BoolVar(&useTLS, "tls", false, "use HTTPS")
	cmd.Flags().StringVar(&path, "path", "/", "request path")
	cmd.Flags().BoolVarP(&followRedirects, "follow-redirects", "L", false, "follow HTTP redirects")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "request timeout")
	return cmd
}

func renderHTTP(w io.Writer, result checks.HTTPResult) {
	fmt.Fprintf(w, "URL: %s\n", result.URL)
	fmt.Fprintf(w, "Latency: %s\n", result.Latency.Round(time.Millisecond))
	if result.Err != nil {
		fmt.Fprintf(w, "Reachable: no\nError: %v\n", result.Err)
		return
	}
	fmt.Fprintf(w, "Status: %d\n", result.StatusCode)
	if result.Location != "" {
		fmt.Fprintf(w, "Location: %s\n", result.Location)
	}
	if result.Timings.TCPConnect > 0 || result.Timings.DNSLookup > 0 || result.Timings.TLSHandshake > 0 || result.Timings.TTFB > 0 {
		fmt.Fprintf(w, "Timings: dns=%s tcp=%s tls=%s ttfb=%s total=%s\n",
			result.Timings.DNSLookup.Round(time.Microsecond),
			result.Timings.TCPConnect.Round(time.Microsecond),
			result.Timings.TLSHandshake.Round(time.Microsecond),
			result.Timings.TTFB.Round(time.Microsecond),
			result.Timings.Total.Round(time.Microsecond))
	}
	fmt.Fprintf(w, "Reachable: %t\n", result.Reachable)
	fmt.Fprintf(w, "Healthy: %t\n", result.Healthy)
}

func httpResultError(result checks.HTTPResult) error {
	if result.Err != nil {
		return newCommandError(exitUnreachable, result.Err)
	}
	if !result.Healthy {
		return newCommandError(exitUnhealthy, fmt.Errorf("endpoint reachable but unhealthy: status %d", result.StatusCode))
	}
	return nil
}

type httpTimingsJSON struct {
	DNSLookup    string `json:"dns_lookup,omitempty"`
	TCPConnect   string `json:"tcp_connect,omitempty"`
	TLSHandshake string `json:"tls_handshake,omitempty"`
	TTFB         string `json:"ttfb,omitempty"`
	Total        string `json:"total"`
}

type httpResultJSON struct {
	URL        string           `json:"url"`
	StatusCode int              `json:"status_code,omitempty"`
	Location   string           `json:"location,omitempty"`
	Latency    string           `json:"latency"`
	Timings    *httpTimingsJSON `json:"timings,omitempty"`
	Reachable  bool             `json:"reachable"`
	Healthy    bool             `json:"healthy"`
	Error      string           `json:"error,omitempty"`
}

func httpOutput(result checks.HTTPResult) httpResultJSON {
	output := httpResultJSON{
		URL: result.URL, StatusCode: result.StatusCode, Location: result.Location,
		Latency: result.Latency.String(), Reachable: result.Reachable, Healthy: result.Healthy,
	}
	if result.Timings.Total > 0 || result.Timings.TCPConnect > 0 || result.Timings.DNSLookup > 0 {
		timings := &httpTimingsJSON{Total: result.Timings.Total.String()}
		if result.Timings.DNSLookup > 0 {
			timings.DNSLookup = result.Timings.DNSLookup.String()
		}
		if result.Timings.TCPConnect > 0 {
			timings.TCPConnect = result.Timings.TCPConnect.String()
		}
		if result.Timings.TLSHandshake > 0 {
			timings.TLSHandshake = result.Timings.TLSHandshake.String()
		}
		if result.Timings.TTFB > 0 {
			timings.TTFB = result.Timings.TTFB.String()
		}
		output.Timings = timings
	}
	if result.Err != nil {
		output.Error = result.Err.Error()
	}
	return output
}
