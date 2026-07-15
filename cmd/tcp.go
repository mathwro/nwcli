package cmd

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var tcpCheck = checks.CheckTCP

func newTCPCommand() *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "tcp HOST:PORT",
		Short: "Check a TCP connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if host, port, err := net.SplitHostPort(args[0]); err != nil || host == "" || port == "" {
				return newCommandError(exitUsage, fmt.Errorf("invalid TCP address %q: expected HOST:PORT", args[0]))
			}
			result := tcpCheck(cmd.Context(), nil, checks.TCPOptions{Address: args[0], Timeout: timeout})
			runErr := tcpResultError(result)
			format, err := cmd.Root().PersistentFlags().GetString("output")
			if err != nil {
				return err
			}
			if format == "json" {
				if err := writeJSONOutput(cmd.OutOrStdout(), "tcp", tcpOutput(result), runErr); err != nil {
					return err
				}
			} else {
				renderTCP(cmd.OutOrStdout(), result)
			}
			return runErr
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "connection timeout")
	return cmd
}

func tcpResultError(result checks.TCPResult) error {
	if result.Err != nil {
		return newCommandError(exitUnreachable, result.Err)
	}
	return nil
}

func renderTCP(w io.Writer, result checks.TCPResult) {
	fmt.Fprintf(w, "Address: %s\n", result.Address)
	if result.Err != nil {
		fmt.Fprintf(w, "Latency: %s\nReachable: false\nError: %v\n", result.Latency.Round(time.Millisecond), result.Err)
		return
	}
	fmt.Fprintf(w, "Remote: %s\nLocal: %s\nLatency: %s\nReachable: true\n", result.RemoteAddress, result.LocalAddress, result.Latency.Round(time.Millisecond))
}

type tcpResultJSON struct {
	Address       string `json:"address"`
	LocalAddress  string `json:"local_address,omitempty"`
	RemoteAddress string `json:"remote_address,omitempty"`
	Latency       string `json:"latency"`
	Reachable     bool   `json:"reachable"`
	Error         string `json:"error,omitempty"`
}

func tcpOutput(result checks.TCPResult) tcpResultJSON {
	output := tcpResultJSON{
		Address: result.Address, LocalAddress: result.LocalAddress, RemoteAddress: result.RemoteAddress,
		Latency: result.Latency.String(), Reachable: result.Reachable,
	}
	if result.Err != nil {
		output.Error = result.Err.Error()
	}
	return output
}
