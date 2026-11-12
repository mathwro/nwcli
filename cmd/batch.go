package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

func newBatchCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "batch", Short: "Run bounded concurrent network checks"}
	cmd.AddCommand(newBatchTCPCommand(), newBatchDNSCommand(), newBatchTLSCommand(), newBatchHTTPCommand(), newBatchPingCommand())
	return cmd
}

func newBatchTCPCommand() *cobra.Command {
	var concurrency int
	var targetFile string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "tcp [HOST:PORT...]",
		Short: "Check multiple TCP endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := resolveBatchTargets(cmd, args, targetFile)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			if err := validateHostPorts("TCP", targets); err != nil {
				return newCommandError(exitUsage, err)
			}
			batch := checks.RunBatch(cmd.Context(), targets, concurrency, func(ctx context.Context, target string) checks.TCPResult {
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
	addBatchFlags(cmd, &concurrency, &targetFile)
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "connection timeout per target")
	return cmd
}

func newBatchDNSCommand() *cobra.Command {
	var concurrency int
	var targetFile string
	var recordType string
	var server string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "dns [HOST...]",
		Short: "Query DNS for multiple hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := resolveBatchTargets(cmd, args, targetFile)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			recordType = strings.ToUpper(strings.TrimSpace(recordType))
			if !supportedDNSType(recordType) {
				return newCommandError(exitUsage, fmt.Errorf("unsupported DNS record type %q", recordType))
			}
			batch := checks.RunBatch(cmd.Context(), targets, concurrency, func(ctx context.Context, target string) checks.DNSResult {
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
	addBatchFlags(cmd, &concurrency, &targetFile)
	cmd.Flags().StringVar(&recordType, "type", "A", "record type")
	cmd.Flags().StringVar(&server, "server", "", "DNS resolver address, with optional port")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "query timeout per target")
	return cmd
}

func newBatchTLSCommand() *cobra.Command {
	var concurrency int
	var targetFile string
	var timeout time.Duration
	var warnBeforeRaw string
	cmd := &cobra.Command{
		Use:   "tls [HOST:PORT...]",
		Short: "Inspect multiple TLS endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := resolveBatchTargets(cmd, args, targetFile)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			if err := validateHostPorts("TLS", targets); err != nil {
				return newCommandError(exitUsage, err)
			}
			warnBefore, err := parseExtendedDuration(warnBeforeRaw)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			batch := checks.RunBatch(cmd.Context(), targets, concurrency, func(ctx context.Context, target string) checks.TLSResult {
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
	addBatchFlags(cmd, &concurrency, &targetFile)
	cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "connection timeout per target")
	cmd.Flags().StringVar(&warnBeforeRaw, "warn-before", "30d", "certificate expiry warning window")
	return cmd
}

func newBatchHTTPCommand() *cobra.Command {
	var concurrency int
	var targetFile string
	var useTLS bool
	var followRedirects bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "http [URL_OR_HOST...]",
		Short: "Check multiple HTTP/S endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := resolveBatchTargets(cmd, args, targetFile)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			optsList := make([]checks.HTTPOptions, len(targets))
			for i, target := range targets {
				parsedOpts, err := parseBatchHTTPTarget(target, useTLS, followRedirects, timeout)
				if err != nil {
					return newCommandError(exitUsage, err)
				}
				optsList[i] = parsedOpts
			}
			indices := make([]string, len(targets))
			for i := range targets {
				indices[i] = strconv.Itoa(i)
			}
			batch := checks.RunBatch(cmd.Context(), indices, concurrency, func(ctx context.Context, idxStr string) checks.HTTPResult {
				idx, _ := strconv.Atoi(idxStr)
				return httpCheck(ctx, nil, optsList[idx])
			})
			results := make([]httpResultJSON, len(batch))
			unreachable, unhealthy := 0, 0
			for i, item := range batch {
				results[i] = httpOutput(item.Result)
				if item.Result.Err != nil {
					unreachable++
				} else if !item.Result.Healthy {
					unhealthy++
				}
			}
			runErr := batchFailure("HTTP", unreachable, unhealthy)
			return emitCommandResult(cmd, "batch http", results, runErr, func(w io.Writer) error {
				for _, item := range batch {
					if item.Result.Err != nil {
						if _, err := fmt.Fprintf(w, "%s reachable=false latency=%s error=%v\n", item.Result.URL, item.Result.Latency.Round(time.Millisecond), item.Result.Err); err != nil {
							return err
						}
						continue
					}
					if _, err := fmt.Fprintf(w, "%s status=%d healthy=%t latency=%s\n", item.Result.URL, item.Result.StatusCode, item.Result.Healthy, item.Result.Latency.Round(time.Millisecond)); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	addBatchFlags(cmd, &concurrency, &targetFile)
	cmd.Flags().BoolVar(&useTLS, "tls", true, "default to HTTPS when target omits scheme")
	cmd.Flags().BoolVarP(&followRedirects, "follow-redirects", "L", false, "follow HTTP redirects")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "request timeout per target")
	return cmd
}

func newBatchPingCommand() *cobra.Command {
	var concurrency int
	var targetFile string
	var count int
	var interval time.Duration
	var timeout time.Duration
	type pingBatchOutcome struct {
		attempts []checks.PingAttempt
		summary  checks.PingSummary
		err      error
	}
	cmd := &cobra.Command{
		Use:   "ping [HOST...]",
		Short: "Ping multiple hosts concurrently",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := resolveBatchTargets(cmd, args, targetFile)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			batch := checks.RunBatch(cmd.Context(), targets, concurrency, func(ctx context.Context, target string) pingBatchOutcome {
				attempts, summary, err := pingCheck(ctx, checks.PingOptions{Host: target, Count: count, Interval: interval, Timeout: timeout})
				return pingBatchOutcome{attempts: attempts, summary: summary, err: err}
			})
			results := make([]pingResultJSON, len(batch))
			unreachable, unhealthy, permissionFailed := 0, 0, false
			for i, item := range batch {
				results[i] = pingOutput(item.Item, item.Result.attempts, item.Result.summary)
				if item.Result.err != nil {
					if checks.IsICMPPermissionError(item.Result.err) {
						permissionFailed = true
					}
					unreachable++
				} else if item.Result.summary.LossPercent > 0 {
					unhealthy++
				}
			}
			var runErr error
			if permissionFailed {
				runErr = newCommandError(exitPermission, fmt.Errorf("ICMP permission error during ping batch"))
			} else {
				runErr = batchFailure("Ping", unreachable, unhealthy)
			}
			return emitCommandResult(cmd, "batch ping", results, runErr, func(w io.Writer) error {
				for _, item := range batch {
					s := item.Result.summary
					if item.Result.err != nil {
						if _, err := fmt.Fprintf(w, "%s reachable=false sent=%d received=%d loss=%.1f%% error=%v\n", item.Item, s.Sent, s.Received, s.LossPercent, item.Result.err); err != nil {
							return err
						}
						continue
					}
					if _, err := fmt.Fprintf(w, "%s sent=%d received=%d loss=%.1f%% avg=%s\n", item.Item, s.Sent, s.Received, s.LossPercent, s.Avg.Round(time.Millisecond)); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	addBatchFlags(cmd, &concurrency, &targetFile)
	cmd.Flags().IntVar(&count, "count", 2, "number of ICMP echo requests per host")
	cmd.Flags().DurationVar(&interval, "interval", 250*time.Millisecond, "interval between probes per host")
	cmd.Flags().DurationVar(&timeout, "timeout", time.Second, "timeout per request")
	return cmd
}

func addBatchFlags(cmd *cobra.Command, concurrency *int, targetFile *string) {
	cmd.Flags().IntVar(concurrency, "concurrency", 4, "maximum concurrent checks")
	cmd.Flags().StringVarP(targetFile, "file", "f", "", "read newline-delimited targets from file ('-' for stdin)")
}

func resolveBatchTargets(cmd *cobra.Command, args []string, targetFile string) ([]string, error) {
	targets := make([]string, 0, len(args))
	for _, arg := range args {
		if trimmed := strings.TrimSpace(arg); trimmed != "" {
			targets = append(targets, trimmed)
		}
	}
	if targetFile != "" {
		var reader io.Reader
		if targetFile == "-" {
			reader = cmd.InOrStdin()
		} else {
			file, err := os.Open(targetFile)
			if err != nil {
				return nil, fmt.Errorf("open targets file %q: %w", targetFile, err)
			}
			defer file.Close()
			reader = file
		}
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if idx := strings.IndexByte(line, '#'); idx >= 0 {
				line = strings.TrimSpace(line[:idx])
			}
			if line != "" {
				targets = append(targets, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read targets %q: %w", targetFile, err)
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one target is required (pass arguments or --file)")
	}
	return targets, nil
}

func parseBatchHTTPTarget(raw string, defaultTLS, followRedirects bool, timeout time.Duration) (checks.HTTPOptions, error) {
	target := strings.TrimSpace(raw)
	if !strings.Contains(target, "://") {
		scheme := "http://"
		if defaultTLS {
			scheme = "https://"
		}
		target = scheme + target
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.Hostname() == "" {
		return checks.HTTPOptions{}, fmt.Errorf("invalid HTTP target %q", raw)
	}
	useTLS := strings.EqualFold(parsed.Scheme, "https")
	if !useTLS && !strings.EqualFold(parsed.Scheme, "http") {
		return checks.HTTPOptions{}, fmt.Errorf("unsupported HTTP scheme %q in %q", parsed.Scheme, raw)
	}
	port := 0
	if parsed.Port() != "" {
		port, err = strconv.Atoi(parsed.Port())
		if err != nil || port < 1 || port > 65535 {
			return checks.HTTPOptions{}, fmt.Errorf("invalid port in HTTP target %q", raw)
		}
	}
	path := parsed.RequestURI()
	if path == "" {
		path = "/"
	}
	return checks.HTTPOptions{
		Host:            parsed.Hostname(),
		Port:            port,
		TLS:             useTLS,
		Path:            path,
		FollowRedirects: followRedirects,
		Timeout:         timeout,
	}, nil
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
