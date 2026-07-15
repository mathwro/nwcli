package cmd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var httpCheck = checks.CheckHTTP

func newHTTPCommand() *cobra.Command {
	var port int
	var useTLS bool
	var path string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "http HOST",
		Short: "Check an HTTP/S endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result := httpCheck(cmd.Context(), http.DefaultClient, checks.HTTPOptions{Host: args[0], Port: port, TLS: useTLS, Path: path, Timeout: timeout})
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

type httpResultJSON struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Latency    string `json:"latency"`
	Reachable  bool   `json:"reachable"`
	Healthy    bool   `json:"healthy"`
	Error      string `json:"error,omitempty"`
}

func httpOutput(result checks.HTTPResult) httpResultJSON {
	output := httpResultJSON{
		URL: result.URL, StatusCode: result.StatusCode, Latency: result.Latency.String(),
		Reachable: result.Reachable, Healthy: result.Healthy,
	}
	if result.Err != nil {
		output.Error = result.Err.Error()
	}
	return output
}
