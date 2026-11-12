package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var tcpCheck = checks.CheckTCP

func newTCPCommand() *cobra.Command {
	var timeout time.Duration
	var concurrency int
	cmd := &cobra.Command{
		Use:   "tcp HOST:PORT[,PORT2,START-END]",
		Short: "Check one or more TCP ports on a host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := checks.ExpandTCPTarget(args[0])
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			if len(targets) == 1 {
				result := tcpCheck(cmd.Context(), nil, checks.TCPOptions{Address: targets[0], Timeout: timeout})
				runErr := tcpResultError(result)
				return emitCommandResult(cmd, "tcp", tcpOutput(result), runErr, func(w io.Writer) error {
					renderTCP(w, result)
					return nil
				})
			}
			batch := checks.RunBatch(cmd.Context(), targets, concurrency, func(ctx context.Context, target string) checks.TCPResult {
				return tcpCheck(ctx, nil, checks.TCPOptions{Address: target, Timeout: timeout})
			})
			results := make([]tcpResultJSON, len(batch))
			closed := 0
			for i, item := range batch {
				results[i] = tcpOutput(item.Result)
				if item.Result.Err != nil {
					closed++
				}
			}
			var runErr error
			if closed == len(batch) {
				runErr = newCommandError(exitUnreachable, fmt.Errorf("all %d TCP ports unreachable on %s", len(batch), args[0]))
			} else if closed > 0 {
				runErr = newCommandError(exitUnhealthy, fmt.Errorf("%d of %d TCP ports unreachable on %s", closed, len(batch), args[0]))
			}
			return emitCommandResult(cmd, "tcp", results, runErr, func(w io.Writer) error {
				for _, item := range batch {
					if item.Result.Err != nil {
						if _, err := fmt.Fprintf(w, "%s reachable=false latency=%s error=%v\n", item.Item, item.Result.Latency.Round(time.Millisecond), item.Result.Err); err != nil {
							return err
						}
						continue
					}
					if _, err := fmt.Fprintf(w, "%s reachable=true latency=%s\n", item.Item, item.Result.Latency.Round(time.Millisecond)); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "connection timeout")
	cmd.Flags().IntVar(&concurrency, "concurrency", 8, "maximum concurrent port checks when scanning multiple ports")
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
