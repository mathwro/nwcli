package cmd

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var tlsCheck = checks.CheckTLS

func newTLSCommand() *cobra.Command {
	var serverName string
	var timeout time.Duration
	var warnBeforeRaw string
	cmd := &cobra.Command{
		Use:   "tls HOST:PORT",
		Short: "Inspect a TLS endpoint and certificate chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if host, port, err := net.SplitHostPort(args[0]); err != nil || host == "" || port == "" {
				return newCommandError(exitUsage, fmt.Errorf("invalid TLS address %q: expected HOST:PORT", args[0]))
			}
			warnBefore, err := parseExtendedDuration(warnBeforeRaw)
			if err != nil {
				return newCommandError(exitUsage, fmt.Errorf("invalid warn-before %q: %w", warnBeforeRaw, err))
			}
			result := tlsCheck(cmd.Context(), checks.TLSOptions{Address: args[0], ServerName: serverName, Timeout: timeout, WarnBefore: warnBefore})
			runErr := tlsResultError(result)
			format, err := cmd.Root().PersistentFlags().GetString("output")
			if err != nil {
				return err
			}
			if format == "json" {
				if err := writeJSONOutput(cmd.OutOrStdout(), "tls", tlsOutput(result), runErr); err != nil {
					return err
				}
			} else {
				renderTLS(cmd.OutOrStdout(), result)
			}
			return runErr
		},
	}
	cmd.Flags().StringVar(&serverName, "server-name", "", "TLS server name for SNI and certificate verification")
	cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "connection and handshake timeout")
	cmd.Flags().StringVar(&warnBeforeRaw, "warn-before", "30d", "return unhealthy when the leaf certificate expires within this duration")
	return cmd
}

func parseExtendedDuration(raw string) (time.Duration, error) {
	if strings.HasSuffix(raw, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(raw, "d"))
		if err != nil || days <= 0 {
			return 0, fmt.Errorf("days must be a positive integer")
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(raw)
}

func tlsResultError(result checks.TLSResult) error {
	if result.Err != nil {
		return newCommandError(exitUnreachable, result.Err)
	}
	if result.ExpiringSoon {
		return newCommandError(exitUnhealthy, fmt.Errorf("TLS certificate for %s expires within the warning window", result.ServerName))
	}
	return nil
}

func renderTLS(w io.Writer, result checks.TLSResult) {
	fmt.Fprintf(w, "Address: %s\nServer name: %s\nLatency: %s\n", result.Address, result.ServerName, result.Latency.Round(time.Millisecond))
	if result.Err != nil {
		fmt.Fprintf(w, "Reachable: false\nVerified: false\nError: %v\n", result.Err)
		return
	}
	fmt.Fprintf(w, "Reachable: %t\nVerified: %t\nVersion: %s\nCipher: %s\nExpiring soon: %t\n", result.Reachable, result.Verified, result.Version, result.CipherSuite, result.ExpiringSoon)
	for i, certificate := range result.Certificates {
		fmt.Fprintf(w, "\nCertificate %d:\nSubject: %s\nIssuer: %s\n", i+1, certificate.Subject, certificate.Issuer)
		if len(certificate.DNSNames) > 0 {
			fmt.Fprintf(w, "DNS names: %s\n", strings.Join(certificate.DNSNames, ", "))
		}
		fmt.Fprintf(w, "Serial: %s\nExpires: %s\n", certificate.Serial, certificate.NotAfter.UTC().Format(time.RFC3339))
	}
}

type certificateJSON struct {
	Subject   string   `json:"subject"`
	Issuer    string   `json:"issuer"`
	DNSNames  []string `json:"dns_names"`
	NotBefore string   `json:"not_before"`
	NotAfter  string   `json:"not_after"`
	Serial    string   `json:"serial"`
}

type tlsResultJSON struct {
	Address      string            `json:"address"`
	ServerName   string            `json:"server_name"`
	Latency      string            `json:"latency"`
	Reachable    bool              `json:"reachable"`
	Verified     bool              `json:"verified"`
	Version      string            `json:"version,omitempty"`
	CipherSuite  string            `json:"cipher_suite,omitempty"`
	ExpiringSoon bool              `json:"expiring_soon"`
	Certificates []certificateJSON `json:"certificates"`
	Error        string            `json:"error,omitempty"`
}

func tlsOutput(result checks.TLSResult) tlsResultJSON {
	certificates := make([]certificateJSON, len(result.Certificates))
	for i, certificate := range result.Certificates {
		certificates[i] = certificateJSON{
			Subject: certificate.Subject, Issuer: certificate.Issuer, DNSNames: certificate.DNSNames,
			NotBefore: certificate.NotBefore.UTC().Format(time.RFC3339), NotAfter: certificate.NotAfter.UTC().Format(time.RFC3339), Serial: certificate.Serial,
		}
	}
	output := tlsResultJSON{
		Address: result.Address, ServerName: result.ServerName, Latency: result.Latency.String(),
		Reachable: result.Reachable, Verified: result.Verified, Version: result.Version,
		CipherSuite: result.CipherSuite, ExpiringSoon: result.ExpiringSoon, Certificates: certificates,
	}
	if result.Err != nil {
		output.Error = result.Err.Error()
	}
	return output
}
