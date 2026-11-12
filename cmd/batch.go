package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

func newBatchCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "batch", Short: "Run bounded concurrent network checks"}
	cmd.AddCommand(newBatchTCPCommand(), newBatchDNSCommand(), newBatchTLSCommand())
	return cmd
}

func newBatchTCPCommand() *cobra.Command {
	var concurrency int
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "tcp HOST:PORT...",
		Short: "Check multiple TCP endpoints",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateHostPorts("TCP", args); err != nil {
				return newCommandError(exitUsage, err)
			}
			batch := checks.RunBatch(cmd.Context(), args, concurrency, func(ctx context.Context, target string) checks.TCPResult {
				return tcpCheck(ctx, nil, checks.TCPOptions{Address: target, Timeout: timeout})
			})
			results := make([]tcpResultJSON, len(batch))
			unreachable := 0
			for i, item := range batch {
				results[i] = tcpOutput(item.Result)
				if item.Result.Err != nil {
					unreachable++
				}
			}
			runErr := batchFailure("TCP", unreachable, 0)
			return emitCommandResult(cmd, "batch tcp", results, runErr, func(w io.Writer) error {
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
	addBatchFlags(cmd, &concurrency)
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "connection timeout per target")
	return cmd
}

func newBatchDNSCommand() *cobra.Command {
	var concurrency int
	var recordType string
	var server string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "dns HOST...",
		Short: "Query DNS for multiple hosts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			recordType = strings.ToUpper(strings.TrimSpace(recordType))
			if !supportedDNSType(recordType) {
				return newCommandError(exitUsage, fmt.Errorf("unsupported DNS record type %q", recordType))
			}
			batch := checks.RunBatch(cmd.Context(), args, concurrency, func(ctx context.Context, target string) checks.DNSResult {
				return dnsCheck(ctx, nil, checks.DNSOptions{Host: target, Type: recordType, Server: server, Timeout: timeout})
			})
			results := make([]dnsResultJSON, len(batch))
			unreachable, unhealthy := 0, 0
			for i, item := range batch {
				results[i] = dnsOutput(item.Result)
				if item.Result.Err != nil {
					unreachable++
				} else if !item.Result.Healthy {
					unhealthy++
				}
			}
			runErr := batchFailure("DNS", unreachable, unhealthy)
			return emitCommandResult(cmd, "batch dns", results, runErr, func(w io.Writer) error {
				for _, item := range batch {
					if item.Result.Err != nil {
						if _, err := fmt.Fprintf(w, "%s reachable=false latency=%s error=%v\n", item.Item, item.Result.Latency.Round(time.Millisecond), item.Result.Err); err != nil {
							return err
						}
						continue
					}
					if _, err := fmt.Fprintf(w, "%s records=%d latency=%s\n", item.Item, len(item.Result.Records), item.Result.Latency.Round(time.Millisecond)); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	addBatchFlags(cmd, &concurrency)
	cmd.Flags().StringVar(&recordType, "type", "A", "record type")
	cmd.Flags().StringVar(&server, "server", "", "DNS resolver address, with optional port")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "query timeout per target")
	return cmd
}

func newBatchTLSCommand() *cobra.Command {
	var concurrency int
	var timeout time.Duration
	var warnBeforeRaw string
	cmd := &cobra.Command{
		Use:   "tls HOST:PORT...",
		Short: "Inspect multiple TLS endpoints",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateHostPorts("TLS", args); err != nil {
				return newCommandError(exitUsage, err)
			}
			warnBefore, err := parseExtendedDuration(warnBeforeRaw)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			batch := checks.RunBatch(cmd.Context(), args, concurrency, func(ctx context.Context, target string) checks.TLSResult {
				return tlsCheck(ctx, checks.TLSOptions{Address: target, Timeout: timeout, WarnBefore: warnBefore})
			})
			results := make([]tlsResultJSON, len(batch))
			unreachable, unhealthy := 0, 0
			for i, item := range batch {
				results[i] = tlsOutput(item.Result)
				if item.Result.Err != nil {
					if item.Result.Reachable {
						unhealthy++
					} else {
						unreachable++
					}
				} else if item.Result.ExpiringSoon {
					unhealthy++
				}
			}
			runErr := batchFailure("TLS", unreachable, unhealthy)
			return emitCommandResult(cmd, "batch tls", results, runErr, func(w io.Writer) error {
				for _, item := range batch {
					if item.Result.Err != nil {
						if _, err := fmt.Fprintf(w, "%s verified=false latency=%s error=%v\n", item.Item, item.Result.Latency.Round(time.Millisecond), item.Result.Err); err != nil {
							return err
						}
						continue
					}
					if _, err := fmt.Fprintf(w, "%s verified=%t expiring_soon=%t version=%s latency=%s\n", item.Item, item.Result.Verified, item.Result.ExpiringSoon, item.Result.Version, item.Result.Latency.Round(time.Millisecond)); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	addBatchFlags(cmd, &concurrency)
	cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "connection timeout per target")
	cmd.Flags().StringVar(&warnBeforeRaw, "warn-before", "30d", "certificate expiry warning window")
	return cmd
}

func addBatchFlags(cmd *cobra.Command, concurrency *int) {
	cmd.Flags().IntVar(concurrency, "concurrency", 4, "maximum concurrent checks")
}

func validateHostPorts(kind string, targets []string) error {
	for _, target := range targets {
		host, port, err := net.SplitHostPort(target)
		if err != nil || host == "" || port == "" {
			return fmt.Errorf("invalid %s address %q: expected HOST:PORT", kind, target)
		}
	}
	return nil
}

func batchFailure(kind string, unreachable, unhealthy int) error {
	if unreachable > 0 {
		return newCommandError(exitUnreachable, fmt.Errorf("%s batch: %d unreachable target(s), %d unhealthy target(s)", kind, unreachable, unhealthy))
	}
	if unhealthy > 0 {
		return newCommandError(exitUnhealthy, fmt.Errorf("%s batch: %d unhealthy target(s)", kind, unhealthy))
	}
	return nil
}
