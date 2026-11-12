package checks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
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

	netDialer := &net.Dialer{Timeout: opts.Timeout}
	start := time.Now()
	rawConn, err := netDialer.DialContext(ctx, "tcp", opts.Address)
	if err != nil {
		result.Latency = time.Since(start)
		result.Err = fmt.Errorf("connect TLS %s: %w", opts.Address, err)
		return result
	}
	result.Reachable = true

	tlsConn := tls.Client(rawConn, &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    opts.RootCAs,
		ServerName: opts.ServerName,
	})
	defer tlsConn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = tlsConn.SetDeadline(deadline)
	}

	err = tlsConn.HandshakeContext(ctx)
	result.Latency = time.Since(start)
	if err != nil {
		var certErr *tls.CertificateVerificationError
		if errors.As(err, &certErr) {
			result.Certificates = convertCertificates(certErr.UnverifiedCertificates)
			if len(certErr.UnverifiedCertificates) > 0 {
				result.ExpiringSoon = time.Until(certErr.UnverifiedCertificates[0].NotAfter) <= opts.WarnBefore
			}
		}
		result.Err = fmt.Errorf("verify TLS %s: %w", opts.Address, err)
		return result
	}

	state := tlsConn.ConnectionState()
	result.Verified = true
	result.Version = tls.VersionName(state.Version)
	result.CipherSuite = tls.CipherSuiteName(state.CipherSuite)
	result.Certificates = convertCertificates(state.PeerCertificates)
	if len(state.PeerCertificates) > 0 {
		result.ExpiringSoon = time.Until(state.PeerCertificates[0].NotAfter) <= opts.WarnBefore
	}
	return result
}

func convertCertificates(certs []*x509.Certificate) []CertificateInfo {
	out := make([]CertificateInfo, len(certs))
	for i, certificate := range certs {
		out[i] = CertificateInfo{
			Subject:   certificate.Subject.String(),
			Issuer:    certificate.Issuer.String(),
			DNSNames:  append([]string(nil), certificate.DNSNames...),
			NotBefore: certificate.NotBefore,
			NotAfter:  certificate.NotAfter,
			Serial:    certificate.SerialNumber.String(),
		}
	}
	return out
}
