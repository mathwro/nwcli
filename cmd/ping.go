package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var pingCheck = checks.CheckPing
var continuousPingCheck = checks.CheckPingContinuous

func newPingCommand() *cobra.Command {
	var count int
	var interval time.Duration
	var timeout time.Duration
	var continuous bool
	cmd := &cobra.Command{
		Use:   "ping HOST",
		Short: "Run an ICMP ping check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if continuous {
				format, flagErr := cmd.Root().PersistentFlags().GetString("output")
				if flagErr != nil {
					return flagErr
				}
				if format == "json" {
					return newCommandError(exitUsage, fmt.Errorf("continuous ping supports text output only"))
				}
				summary, err := continuousPingCheck(cmd.Context(), checks.PingOptions{
					Host:     args[0],
					Interval: interval,
					Timeout:  timeout,
					OnUpdate: func(update checks.PingUpdate) {
						renderPingUpdate(cmd.OutOrStdout(), update)
					},
				})
				fmt.Fprint(cmd.OutOrStdout(), "summary: ")
				renderPing(cmd.OutOrStdout(), nil, summary)
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return pingResultError(err)
			}
			attempts, summary, err := pingCheck(cmd.Context(), checks.PingOptions{Host: args[0], Count: count, Interval: interval, Timeout: timeout})
			runErr := pingResultError(err)
			return emitCommandResult(cmd, "ping", pingOutput(args[0], attempts, summary), runErr, func(w io.Writer) error {
				renderPing(w, attempts, summary)
				return nil
			})
		},
	}
	cmd.Flags().IntVar(&count, "count", 4, "number of ICMP echo requests")
	cmd.Flags().DurationVarP(&interval, "interval", "i", time.Second, "interval between ICMP echo requests")
	cmd.Flags().DurationVar(&timeout, "timeout", time.Second, "timeout per request")
	cmd.Flags().BoolVarP(&continuous, "continuous", "c", false, "ping until canceled and stream live packet-loss statistics")
	return cmd
}

func renderPing(w io.Writer, attempts []checks.PingAttempt, summary checks.PingSummary) {
	for _, attempt := range attempts {
		if attempt.Err != nil {
			fmt.Fprintf(w, "seq=%d error=%v\n", attempt.Seq, attempt.Err)
			continue
		}
		fmt.Fprintf(w, "seq=%d latency=%s\n", attempt.Seq, attempt.Latency.Round(time.Millisecond))
	}
	fmt.Fprintf(w, "sent=%d received=%d loss=%.1f%% min=%s avg=%s max=%s\n", summary.Sent, summary.Received, summary.LossPercent, summary.Min.Round(time.Millisecond), summary.Avg.Round(time.Millisecond), summary.Max.Round(time.Millisecond))
}

func renderPingUpdate(w io.Writer, update checks.PingUpdate) {
	dropped := update.Summary.Sent - update.Summary.Received
	if update.Attempt.Err != nil {
		fmt.Fprintf(w, "seq=%d DROP error=%v sent=%d received=%d dropped=%d loss=%.1f%%\n",
			update.Attempt.Seq, update.Attempt.Err, update.Summary.Sent, update.Summary.Received, dropped, update.Summary.LossPercent)
		return
	}
	fmt.Fprintf(w, "seq=%d latency=%s sent=%d received=%d dropped=%d loss=%.1f%%\n",
		update.Attempt.Seq, update.Attempt.Latency.Round(time.Millisecond), update.Summary.Sent, update.Summary.Received, dropped, update.Summary.LossPercent)
}

func pingResultError(err error) error {
	if err == nil {
		return nil
	}
	if checks.IsICMPPermissionError(err) {
		return newCommandError(exitPermission, err)
	}
	return newCommandError(exitUnreachable, err)
}

type pingAttemptJSON struct {
	Sequence int    `json:"sequence"`
	Latency  string `json:"latency,omitempty"`
	Error    string `json:"error,omitempty"`
}

type pingSummaryJSON struct {
	Sent        int     `json:"sent"`
	Received    int     `json:"received"`
	LossPercent float64 `json:"loss_percent"`
	Min         string  `json:"min"`
	Average     string  `json:"average"`
	Max         string  `json:"max"`
}

type pingResultJSON struct {
	Host     string            `json:"host"`
	Attempts []pingAttemptJSON `json:"attempts"`
	Summary  pingSummaryJSON   `json:"summary"`
}

func pingOutput(host string, attempts []checks.PingAttempt, summary checks.PingSummary) pingResultJSON {
	outputAttempts := make([]pingAttemptJSON, len(attempts))
	for i, attempt := range attempts {
		outputAttempts[i] = pingAttemptJSON{Sequence: attempt.Seq, Latency: attempt.Latency.String()}
		if attempt.Err != nil {
			outputAttempts[i].Error = attempt.Err.Error()
			outputAttempts[i].Latency = ""
		}
	}
	return pingResultJSON{
		Host:     host,
		Attempts: outputAttempts,
		Summary: pingSummaryJSON{
			Sent: summary.Sent, Received: summary.Received, LossPercent: summary.LossPercent,
			Min: summary.Min.String(), Average: summary.Avg.String(), Max: summary.Max.String(),
		},
	}
}
