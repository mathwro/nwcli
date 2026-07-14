package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

func newPingCommand() *cobra.Command {
	var count int
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "ping HOST",
		Short: "Run an ICMP ping check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			attempts, summary, err := checks.CheckPing(cmd.Context(), checks.PingOptions{Host: args[0], Count: count, Timeout: timeout})
			renderPing(cmd.OutOrStdout(), attempts, summary)
			return err
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
