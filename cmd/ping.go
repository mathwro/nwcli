package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var pingCheck = checks.CheckPing

func newPingCommand() *cobra.Command {
	var count int
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "ping HOST",
		Short: "Run an ICMP ping check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			attempts, summary, err := pingCheck(cmd.Context(), checks.PingOptions{Host: args[0], Count: count, Timeout: timeout})
			runErr := pingResultError(err)
			return emitCommandResult(cmd, "ping", pingOutput(args[0], attempts, summary), runErr, func(w io.Writer) error {
				renderPing(w, attempts, summary)
				return nil
			})
		},
	}
	cmd.Flags().IntVar(&count, "count", 4, "number of ICMP echo requests")
	cmd.Flags().DurationVar(&timeout, "timeout", time.Second, "timeout per request")
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
