package checks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"
)

type TLSOptions struct {
	Address    string
	ServerName string
	Timeout    time.Duration
	WarnBefore time.Duration
	RootCAs    *x509.CertPool
}

type CertificateInfo struct {
	Subject   string
	Issuer    string
	DNSNames  []string
	NotBefore time.Time
	NotAfter  time.Time
	Serial    string
}

type TLSResult struct {
	Address      string
	ServerName   string
	Latency      time.Duration
	Reachable    bool
	Verified     bool
	Version      string
	CipherSuite  string
	Certificates []CertificateInfo
	ExpiringSoon bool
	Err          error
}

func CheckTLS(ctx context.Context, opts TLSOptions) TLSResult {
	result := TLSResult{Address: opts.Address, ServerName: opts.ServerName}
	host, port, err := net.SplitHostPort(opts.Address)
	if err != nil || host == "" || port == "" {
		result.Err = fmt.Errorf("invalid TLS address %q: expected HOST:PORT", opts.Address)
		return result
	}
	if opts.ServerName == "" {
		opts.ServerName = strings.Trim(host, "[]")
		result.ServerName = opts.ServerName
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 3 * time.Second
	}
	if opts.WarnBefore <= 0 {
		opts.WarnBefore = 30 * 24 * time.Hour
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	dialer := tls.Dialer{
		NetDialer: &net.Dialer{Timeout: opts.Timeout},
		Config: &tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    opts.RootCAs,
			ServerName: opts.ServerName,
		},
	}
	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", opts.Address)
	result.Latency = time.Since(start)
	if err != nil {
		result.Err = fmt.Errorf("connect TLS %s: %w", opts.Address, err)
		return result
	}
	defer conn.Close()
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		result.Err = fmt.Errorf("connect TLS %s: unexpected connection type %T", opts.Address, conn)
		return result
	}
	state := tlsConn.ConnectionState()
	result.Reachable = true
	result.Verified = true
	result.Version = tls.VersionName(state.Version)
	result.CipherSuite = tls.CipherSuiteName(state.CipherSuite)
	result.Certificates = make([]CertificateInfo, len(state.PeerCertificates))
	for i, certificate := range state.PeerCertificates {
		result.Certificates[i] = CertificateInfo{
			Subject: certificate.Subject.String(), Issuer: certificate.Issuer.String(),
			DNSNames: append([]string(nil), certificate.DNSNames...), NotBefore: certificate.NotBefore,
			NotAfter: certificate.NotAfter, Serial: certificate.SerialNumber.String(),
		}
	}
	if len(state.PeerCertificates) > 0 {
		result.ExpiringSoon = time.Until(state.PeerCertificates[0].NotAfter) <= opts.WarnBefore
	}
	return result
}
