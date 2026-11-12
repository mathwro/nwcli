package checks

import (
	"context"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckTLSInspectsVerifiedConnection(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	address := strings.TrimPrefix(server.URL, "https://")

	result := CheckTLS(context.Background(), TLSOptions{
		Address:    address,
		ServerName: "example.com",
		Timeout:    time.Second,
		WarnBefore: 100 * 365 * 24 * time.Hour,
		RootCAs:    roots,
	})
	if result.Err != nil {
		t.Fatalf("CheckTLS returned error: %v", result.Err)
	}
	if !result.Reachable || !result.Verified || len(result.Certificates) == 0 {
		t.Fatalf("result = %#v", result)
	}
	if result.Version == "" || result.CipherSuite == "" {
		t.Fatalf("missing negotiated TLS details: %#v", result)
	}
	if !result.ExpiringSoon {
		t.Fatal("ExpiringSoon = false with a 100-year warning window")
	}
}

func TestCheckTLSReportsCertificateVerificationFailureWithUnverifiedChain(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	address := strings.TrimPrefix(server.URL, "https://")

	result := CheckTLS(context.Background(), TLSOptions{Address: address, Timeout: time.Second})
	if result.Err == nil {
		t.Fatal("Err = nil, want verification error")
	}
	if !result.Reachable || result.Verified {
		t.Fatalf("result = %#v, want Reachable=true Verified=false", result)
	}
	if len(result.Certificates) == 0 {
		t.Fatalf("Certificates empty on verification failure: %#v", result)
	}
}

func TestCheckTLSMarksConnectionFailureUnreachable(t *testing.T) {
	result := CheckTLS(context.Background(), TLSOptions{Address: "127.0.0.1:1", Timeout: 50 * time.Millisecond})
	if result.Err == nil || result.Reachable || result.Verified {
		t.Fatalf("result = %#v, want unreachable", result)
	}
}

func TestCheckTLSRejectsAddressWithoutPort(t *testing.T) {
	result := CheckTLS(context.Background(), TLSOptions{Address: "example.com"})
	if result.Err == nil {
		t.Fatal("Err = nil")
	}
}
