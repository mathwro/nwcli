package checks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCheckHTTPMarks2xxReachableAndHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	host, port := hostPortFromURL(t, server.URL)
	result := CheckHTTP(context.Background(), server.Client(), HTTPOptions{Host: host, Port: port, Path: "/", Timeout: time.Second})

	if result.Err != nil {
		t.Fatalf("Err = %v", result.Err)
	}
	if !result.Reachable || !result.Healthy || result.StatusCode != http.StatusNoContent {
		t.Fatalf("result = %+v, want reachable healthy 204", result)
	}
	if result.Latency <= 0 {
		t.Fatalf("Latency = %v, want positive", result.Latency)
	}
}

func TestCheckHTTPMarks404ReachableButUnhealthy(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	host, port := hostPortFromURL(t, server.URL)
	result := CheckHTTP(context.Background(), server.Client(), HTTPOptions{Host: host, Port: port, Path: "/missing", Timeout: time.Second})

	if result.Err != nil {
		t.Fatalf("Err = %v", result.Err)
	}
	if !result.Reachable || result.Healthy || result.StatusCode != http.StatusNotFound {
		t.Fatalf("result = %+v, want reachable unhealthy 404", result)
	}
}

func TestCheckHTTPReportsTransportError(t *testing.T) {
	result := CheckHTTP(context.Background(), http.DefaultClient, HTTPOptions{Host: "127.0.0.1", Port: 1, Path: "/", Timeout: 20 * time.Millisecond})
	if result.Err == nil {
		t.Fatal("Err = nil, want transport error")
	}
	if result.Reachable || result.Healthy {
		t.Fatalf("result = %+v, want unreachable unhealthy", result)
	}
}

func TestCheckHTTPBuildsHTTPSURL(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://example.com:443/health" {
			t.Fatalf("URL = %q", req.URL.String())
		}
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})}
	result := CheckHTTP(context.Background(), client, HTTPOptions{Host: "example.com", Port: 443, TLS: true, Path: "health", Timeout: time.Second})
	if result.Err != nil || !strings.HasPrefix(result.URL, "https://example.com:443/health") {
		t.Fatalf("result = %+v", result)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func hostPortFromURL(t *testing.T, raw string) (string, int) {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse server URL %q: %v", raw, err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("invalid port in %q: %v", raw, err)
	}
	return parsed.Hostname(), port
}
