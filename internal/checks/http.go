package checks

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"
)

type HTTPOptions struct {
	Host            string
	Port            int
	TLS             bool
	Path            string
	FollowRedirects bool
	Timeout         time.Duration
}

type HTTPTimings struct {
	DNSLookup    time.Duration
	TCPConnect   time.Duration
	TLSHandshake time.Duration
	TTFB         time.Duration
	Total        time.Duration
}

type HTTPResult struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Timings    HTTPTimings
	Location   string
	Reachable  bool
	Healthy    bool
	Err        error
}

func CheckHTTP(ctx context.Context, client *http.Client, opts HTTPOptions) HTTPResult {
	if opts.Timeout <= 0 {
		opts.Timeout = 2 * time.Second
	}
	if client == nil {
		transport := &http.Transport{
			DisableKeepAlives: true,
			DialContext:       (&net.Dialer{Timeout: opts.Timeout}).DialContext,
		}
		client = &http.Client{Transport: transport}
		if !opts.FollowRedirects {
			client.CheckRedirect = func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
		defer transport.CloseIdleConnections()
	}
	if opts.Port == 0 {
		if opts.TLS {
			opts.Port = 443
		} else {
			opts.Port = 80
		}
	}
	path := opts.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	scheme := "http"
	if opts.TLS {
		scheme = "https"
	}
	hostPort := net.JoinHostPort(strings.Trim(opts.Host, "[]"), strconv.Itoa(opts.Port))
	url := fmt.Sprintf("%s://%s%s", scheme, hostPort, path)
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	var dnsStart, connectStart, tlsStart, wroteReq time.Time
	var timings HTTPTimings
	trace := &httptrace.ClientTrace{
		DNSStart:     func(httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:      func(httptrace.DNSDoneInfo) { timings.DNSLookup += time.Since(dnsStart) },
		ConnectStart: func(string, string) { connectStart = time.Now() },
		ConnectDone: func(string, string, error) {
			if !connectStart.IsZero() {
				timings.TCPConnect += time.Since(connectStart)
			}
		},
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			if !tlsStart.IsZero() {
				timings.TLSHandshake += time.Since(tlsStart)
			}
		},
		WroteRequest: func(httptrace.WroteRequestInfo) { wroteReq = time.Now() },
		GotFirstResponseByte: func() {
			if !wroteReq.IsZero() {
				timings.TTFB += time.Since(wroteReq)
			}
		},
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, url, nil)
	if err != nil {
		return HTTPResult{URL: url, Err: err}
	}
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)
	timings.Total = latency
	if err != nil {
		return HTTPResult{URL: url, Latency: latency, Timings: timings, Err: err}
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
	}()
	healthy := resp.StatusCode >= 200 && resp.StatusCode <= 399
	return HTTPResult{
		URL:        url,
		StatusCode: resp.StatusCode,
		Latency:    latency,
		Timings:    timings,
		Location:   resp.Header.Get("Location"),
		Reachable:  true,
		Healthy:    healthy,
	}
}
