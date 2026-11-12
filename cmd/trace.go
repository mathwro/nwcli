package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var traceCheck = checks.CheckTrace

func newTraceCommand() *cobra.Command {
	var maxHops int
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "trace HOST",
		Short: "Trace ICMP network path hop-by-hop to a host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if maxHops < 1 || maxHops > 64 {
				return newCommandError(exitUsage, fmt.Errorf("max-hops must be between 1 and 64"))
			}
			format, flagErr := cmd.Root().PersistentFlags().GetString("output")
			if flagErr != nil {
				return flagErr
			}
			var streamed bool
			result, err := traceCheck(cmd.Context(), checks.TraceOptions{
				Host:    args[0],
				MaxHops: maxHops,
				Timeout: timeout,
				OnUpdate: func(hop checks.TraceHop) {
					if format != "json" {
						if !streamed {
							streamed = true
						}
						renderTraceHop(cmd.OutOrStdout(), hop)
					}
				},
			})
			runErr := pingResultError(err)
			return emitCommandResult(cmd, "trace", traceOutput(result), runErr, func(w io.Writer) error {
				if !streamed {
					renderTrace(w, result)
				}
				return nil
			})
		},
	}
	cmd.Flags().IntVarP(&maxHops, "max-hops", "m", 20, "maximum TTL hops (1-64)")
	cmd.Flags().DurationVar(&timeout, "timeout", time.Second, "probe timeout per hop")
	return cmd
}

func renderTrace(w io.Writer, result checks.TraceResult) {
	if result.Target != "" && result.Target != result.Host {
		fmt.Fprintf(w, "Target: %s (%s)\n", result.Host, result.Target)
	} else if result.Host != "" {
		fmt.Fprintf(w, "Target: %s\n", result.Host)
	}
	for _, hop := range result.Hops {
		renderTraceHop(w, hop)
	}
}

func renderTraceHop(w io.Writer, hop checks.TraceHop) {
	if hop.Timeout {
		fmt.Fprintf(w, "%2d  *\n", hop.TTL)
		return
	}
	if hop.Err != nil {
		fmt.Fprintf(w, "%2d  error=%v\n", hop.TTL, hop.Err)
		return
	}
	fmt.Fprintf(w, "%2d  %-39s %s\n", hop.TTL, hop.Address, hop.Latency.Round(time.Microsecond))
}

type traceHopJSON struct {
	TTL     int    `json:"ttl"`
	Address string `json:"address,omitempty"`
	Latency string `json:"latency,omitempty"`
	Timeout bool   `json:"timeout,omitempty"`
	Reached bool   `json:"reached,omitempty"`
	Error   string `json:"error,omitempty"`
}

type traceResultJSON struct {
	Host    string         `json:"host"`
	Target  string         `json:"target,omitempty"`
	Reached bool           `json:"reached"`
	Hops    []traceHopJSON `json:"hops"`
}

func traceOutput(result checks.TraceResult) traceResultJSON {
	hops := make([]traceHopJSON, len(result.Hops))
	for i, hop := range result.Hops {
		item := traceHopJSON{
			TTL:     hop.TTL,
			Address: hop.Address,
			Timeout: hop.Timeout,
			Reached: hop.Reached,
		}
		if hop.Latency > 0 {
			item.Latency = hop.Latency.String()
		}
		if hop.Err != nil {
			item.Error = hop.Err.Error()
		}
		hops[i] = item
	}
	return traceResultJSON{
		Host:    result.Host,
		Target:  result.Target,
		Reached: result.Reached,
		Hops:    hops,
	}
}
