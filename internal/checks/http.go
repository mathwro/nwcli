package checks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type HTTPOptions struct {
	Host    string
	Port    int
	TLS     bool
	Path    string
	Timeout time.Duration
}

type HTTPResult struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Reachable  bool
	Healthy    bool
	Err        error
}

func CheckHTTP(ctx context.Context, client *http.Client, opts HTTPOptions) HTTPResult {
	if client == nil {
		client = http.DefaultClient
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 2 * time.Second
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
	url := fmt.Sprintf("%s://%s:%d%s", scheme, opts.Host, opts.Port, path)
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return HTTPResult{URL: url, Err: err}
	}
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return HTTPResult{URL: url, Latency: latency, Err: err}
	}
	defer resp.Body.Close()
	healthy := resp.StatusCode >= 200 && resp.StatusCode <= 399
	return HTTPResult{URL: url, StatusCode: resp.StatusCode, Latency: latency, Reachable: true, Healthy: healthy}
}
